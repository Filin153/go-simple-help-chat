package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"shc/domain"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

type MsgCacheInterface interface {
	Set(ctx context.Context, userUUID string, val *domain.Msg) error
	Get(ctx context.Context, userUUID string) ([]*domain.Msg, bool, error)
	DeleteByMsgID(ctx context.Context, userUUID string, id int) error
}

type S3Interface interface {
	Save(ctx context.Context, folder string, file []byte) (path string, err error)
	Delete(ctx context.Context, path string) error
}

type MsgRepo interface {
	CreateForClient(ctx context.Context, msg domain.CreateMsg, tx pgx.Tx) (savedMsg *domain.Msg, err error)
	CreateForManager(ctx context.Context, msg domain.CreateMsg, tx pgx.Tx) (savedMsg *domain.Msg, err error)
	CreateFile(ctx context.Context, msgID int, fileName, path string, tx pgx.Tx) (file domain.MsgFileContent, err error)
	GetUnread(ctx context.Context, userUUID string, tx pgx.Tx) (msg []*domain.Msg, err error)
	MarkReadByID(ctx context.Context, userUUID string, id int, tx pgx.Tx) error
	GetHistory(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]*domain.Msg, error)
}

type TicketRepo interface {
	GetByID(ctx context.Context, id int) (domain.Ticket, error)
}

type EncryptionInterface interface {
	Encrypt(plaintext []byte, aad []byte) ([]byte, error)
	Decrypt(data []byte, aad []byte) ([]byte, error)
}

type ChatUseCase struct {
	msgCache            MsgCacheInterface
	s3                  S3Interface
	msgRepo             MsgRepo
	ticketRepo          TicketRepo
	mainRepo            MainRepo
	encryption          EncryptionInterface
	longPullReadTimeOut time.Duration
	pollInterval        time.Duration
}

func NewChatUseCase(msgCache MsgCacheInterface, s3 S3Interface, msgRepo MsgRepo, ticketRepo TicketRepo, mainRepo MainRepo, encryption EncryptionInterface, longPullReadTimeOut, pollInterval time.Duration) *ChatUseCase {
	return &ChatUseCase{
		msgCache:            msgCache,
		s3:                  s3,
		msgRepo:             msgRepo,
		ticketRepo:          ticketRepo,
		mainRepo:            mainRepo,
		encryption:          encryption,
		longPullReadTimeOut: longPullReadTimeOut,
		pollInterval:        pollInterval,
	}
}

func (c *ChatUseCase) Send(ctx context.Context, msg domain.CreateMsg, fromType domain.MsgFromType) (err error) {
	ticket, err := c.ticketRepo.GetByID(ctx, msg.TicketID)
	if err != nil {
		return err
	}

	if err := c.encryptMsgText(&msg); err != nil {
		return err
	}
	msg.Text = ""

	tx, err := c.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	files := make([]domain.CreateFile, 0, len(msg.Files))
	pathes := make([]string, 0, len(msg.Files))

	defer func() {
		if len(pathes) > 0 && err != nil {
			go c.deleteFormS3(ctx, pathes)
		}
	}()

	if len(msg.Files) > 0 {
		files, pathes, err = c.fileToS3(ctx, msg.Files, msg.TicketID)
		if err != nil {
			return err
		}
	}
	switch fromType {
	case domain.MsgFromTypeManager:
		savedMsg, err := c.createForClientDB(ctx, msg, &ticket, files, tx)
		if err != nil {
			return err
		}
		go c.createForClientCache(ctx, savedMsg, &ticket)
	case domain.MsgFromTypeClient:
		savedMsg, err := c.createForManagerDB(ctx, msg, &ticket, files, tx)
		if err != nil {
			return err
		}
		go c.createForManagerCache(ctx, savedMsg, &ticket)
	case domain.MsgFromTypeSystem:
		var savedMsgClient, savedMsgManager *domain.Msg
		savedMsgClient, err = c.createForClientDB(ctx, msg, &ticket, files, tx)
		if err != nil {
			return err
		}

		savedMsgManager, err = c.createForManagerDB(ctx, msg, &ticket, files, tx)
		if err != nil {
			return err
		}

		go c.createForClientCache(ctx, savedMsgClient, &ticket)
		go c.createForManagerCache(ctx, savedMsgManager, &ticket)
	default:
		return fmt.Errorf("Unknown MsgFromType")
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (c *ChatUseCase) GetAllNew(ctx context.Context, userUUID string) (res []*domain.Msg, ok bool, err error) {
	timeOutContext, cf := context.WithTimeout(ctx, c.longPullReadTimeOut)
	defer cf()

	ticker := time.NewTicker(c.pollInterval)
	defer ticker.Stop()

	for {
		if ok {
			break
		}

		select {
		case <-timeOutContext.Done():
			return
		case <-ticker.C:
			res, ok, err = c.msgCache.Get(timeOutContext, userUUID)
			if err != nil {
				slog.Error("read from cache error")
				res, err = c.msgRepo.GetUnread(ctx, userUUID, nil)
				if err != nil {
					return
				} else if len(res) > 0 {
					ok = true
				}
			}
		}
	}

	for _, msg := range res {
		if err = c.decryptMsgText(msg); err != nil {
			return
		}
	}

	return
}

func (c *ChatUseCase) MarkAsRead(ctx context.Context, userUUID string, msgIDs []int) error {
	tx, err := c.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, id := range msgIDs {
		id := id
		go c.msgCache.DeleteByMsgID(ctx, userUUID, id)
		if err := c.msgRepo.MarkReadByID(ctx, userUUID, id, tx); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (c *ChatUseCase) GetHistory(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]*domain.Msg, error) {
	msges, err := c.msgRepo.GetHistory(ctx, userUUID, ticketUUID, from, to)
	if err != nil {
		return nil, err
	}

	for _, msg := range msges {
		if err = c.decryptMsgText(msg); err != nil {
			return nil, err
		}
	}

	return msges, nil
}

func (c *ChatUseCase) fileToS3(ctx context.Context, files []domain.CreateFile, ticketID int) ([]domain.CreateFile, []string, error) {
	if len(files) > 10 {
		return []domain.CreateFile{}, []string{}, domain.ErrFilesLoad
	}

	for _, file := range files {
		if float64(len(file.Data))/(1024*1024) > 15 {
			return []domain.CreateFile{}, []string{}, domain.ErrFilesLoad
		}
	}

	storedPaths := make([]string, len(files))
	errGroup, gCtx := errgroup.WithContext(ctx)
	errGroup.SetLimit(3)
	for i, file := range files {
		i, file := i, file
		errGroup.Go(func() error {
			path, err := c.s3.Save(gCtx, "ticket/file/"+strconv.Itoa(ticketID), file.Data)
			if err != nil {
				return err
			}
			files[i].Path = path
			storedPaths[i] = path
			return nil
		})
	}

	pathes := make([]string, 0, len(files))
	if err := errGroup.Wait(); err != nil {
		for _, path := range storedPaths {
			if path != "" {
				pathes = append(pathes, path)
			}
		}
		return files, pathes, err
	}

	for _, path := range storedPaths {
		if path != "" {
			pathes = append(pathes, path)
		}
	}

	return files, pathes, nil
}

func (c *ChatUseCase) encryptMsgText(msg *domain.CreateMsg) (err error) {
	aad := c.createAdd(msg.TicketID)
	msg.EncryptedText, err = c.encryption.Encrypt([]byte(msg.Text), aad)
	if err != nil {
		return err
	}
	return nil
}

func (c *ChatUseCase) decryptMsgText(msg *domain.Msg) (err error) {
	aad := c.createAdd(msg.TicketID)
	decText, err := c.encryption.Decrypt(msg.EncryptedText, aad)
	if err != nil {
		return err
	}
	msg.Text = string(decText)
	return nil
}

func (c *ChatUseCase) createAdd(ticketID int) []byte {
	return []byte(fmt.Sprintf(
		"%d",
		ticketID,
	))
}

func (c *ChatUseCase) createForClientDB(ctx context.Context, msg domain.CreateMsg, ticket *domain.Ticket, files []domain.CreateFile, tx pgx.Tx) (*domain.Msg, error) {
	savedMsg, err := c.msgRepo.CreateForClient(ctx, msg, tx)
	if err != nil {
		return nil, err
	}

	savedMsg.Files = make([]domain.MsgFileContent, 0, len(msg.Files))
	for _, file := range files {
		fileObj, err := c.msgRepo.CreateFile(ctx, savedMsg.ID, file.Name, file.Path, tx)
		if err != nil {
			return nil, err
		}
		savedMsg.Files = append(savedMsg.Files, fileObj)
	}
	return savedMsg, nil
}

func (c *ChatUseCase) createForClientCache(ctx context.Context, msg *domain.Msg, ticket *domain.Ticket) error {
	if err := c.msgCache.Set(ctx, ticket.ClientUserUUID, msg); err != nil {
		slog.Error("save to cache error")
		return err
	}
	return nil
}

func (c *ChatUseCase) createForManagerDB(ctx context.Context, msg domain.CreateMsg, ticket *domain.Ticket, files []domain.CreateFile, tx pgx.Tx) (*domain.Msg, error) {
	savedMsg, err := c.msgRepo.CreateForManager(ctx, msg, tx)
	if err != nil {
		return nil, err
	}

	savedMsg.Files = make([]domain.MsgFileContent, 0, len(msg.Files))
	for _, file := range files {
		fileObj, err := c.msgRepo.CreateFile(ctx, savedMsg.ID, file.Name, file.Path, tx)
		if err != nil {
			return nil, err
		}
		savedMsg.Files = append(savedMsg.Files, fileObj)
	}
	return savedMsg, nil
}
func (c *ChatUseCase) createForManagerCache(ctx context.Context, msg *domain.Msg, ticket *domain.Ticket) error {
	if err := c.msgCache.Set(ctx, ticket.ManagerUserUUID, msg); err != nil {
		return err
	}
	return nil
}

func (c *ChatUseCase) deleteFormS3(ctx context.Context, pathes []string) error {
	errG, ctxG := errgroup.WithContext(ctx)
	errG.SetLimit(3)
	for _, path := range pathes {
		path := path
		errG.Go(func() error {
			return c.s3.Delete(ctxG, path)
		})
	}

	if err := errG.Wait(); err != nil {
		slog.Error("Can not delete file from S3", "error", err)
		return err
	}

	return nil
}
