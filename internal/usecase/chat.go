package usecase

import (
	"context"
	"log/slog"
	"shc/domain"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

var getAllNewTickerDuration = time.Second

type MsgCacheInterface interface {
	Set(ctx context.Context, userUUID string, val *domain.Msg) error
	Get(ctx context.Context, userUUID string) ([]domain.Msg, bool, error)
	DeleteByMsgID(ctx context.Context, userUUID string, id int)
}

type S3Interface interface {
	Save(ctx context.Context, folder string, file []byte) (path string, err error)
}

type MsgRepo interface {
	Create(ctx context.Context, msg domain.CreateMsg, fromType domain.MsgFromType, draft bool, tx pgx.Tx) (savedMsg *domain.Msg, err error)
	CreateFile(ctx context.Context, msgID int, path string, tx pgx.Tx) (file domain.MsgFileContent, err error)
	GetUnread(ctx context.Context, userUUID int, tx pgx.Tx) (msg []domain.Msg, err error)
	MarkReadByID(ctx context.Context, userUUID string, id int, tx pgx.Tx) error
	GetHistory(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]domain.Msg, error)
}

type TicketRepo interface {
	GetByID(ctx context.Context, id int) (domain.Ticket, error)
}

type ChatUseCase struct {
	msgCache            MsgCacheInterface
	s3                  S3Interface
	msgRepo             MsgRepo
	ticketRepo          TicketRepo
	mainRepo            MainRepo
	longPullReadTimeOut time.Duration
}

func NewChatUseCase(msgCache MsgCacheInterface, s3 S3Interface, msgRepo MsgRepo, ticketRepo TicketRepo, mainRepo MainRepo, longPullReadTimeOut time.Duration) *ChatUseCase {
	return &ChatUseCase{
		msgCache:            msgCache,
		s3:                  s3,
		msgRepo:             msgRepo,
		ticketRepo:          ticketRepo,
		mainRepo:            mainRepo,
		longPullReadTimeOut: longPullReadTimeOut,
	}
}

func (c *ChatUseCase) Send(ctx context.Context, msg domain.CreateMsg, fromType domain.MsgFromType) error {
	tx, err := c.mainRepo.CreateSession(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	savedMsg, err := c.msgRepo.Create(ctx, msg, fromType, false, tx)
	if err != nil {
		return err
	}

	if len(msg.Files) > 0 {
		if len(msg.Files) > 10 {
			return domain.ErrFilesLoad
		}

		for _, file := range msg.Files {
			if float64(len(file))/(1024*1024) > 15 {
				return domain.ErrFilesLoad
			}
		}

		errGroup, gCtx := errgroup.WithContext(ctx)
		pathes := make([]string, len(msg.Files))
		for i, file := range msg.Files {
			i, file := i, file
			errGroup.Go(func() error {
				path, err := c.s3.Save(gCtx, "ticket/file/"+strconv.Itoa(savedMsg.ID), file)
				if err != nil {
					return err
				}
				pathes[i] = path
				return nil
			})
		}

		if err := errGroup.Wait(); err != nil {
			return err
		}

		savedMsg.Files = make([]domain.MsgFileContent, 0, len(msg.Files))
		for _, path := range pathes {
			fileObj, err := c.msgRepo.CreateFile(gCtx, savedMsg.ID, path, tx)
			if err != nil {
				return err
			}
			savedMsg.Files = append(savedMsg.Files, fileObj)
		}
	}

	ticket, err := c.ticketRepo.GetByID(ctx, savedMsg.TicketID)
	if err != nil {
		return err
	}

	go func() {
		if fromType == domain.MsgFromTypeManager {
			if err := c.msgCache.Set(ctx, ticket.ClientUserUUID, savedMsg); err != nil {
				slog.Error("save to cache error")
			}
		} else if fromType == domain.MsgFromTypeClient {
			if err := c.msgCache.Set(ctx, ticket.ManagerUserUUID, savedMsg); err != nil {
				slog.Error("save to cache error")
			}
		} else {
			if err := c.msgCache.Set(ctx, ticket.ClientUserUUID, savedMsg); err != nil {
				slog.Error("save to cache error")
			}
			if err := c.msgCache.Set(ctx, ticket.ManagerUserUUID, savedMsg); err != nil {
				slog.Error("save to cache error")
			}
		}
	}()

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (c *ChatUseCase) GetAllNew(ctx context.Context, userUUID int) (res []domain.Msg, ok bool, err error) {
	timeOutContext, cf := context.WithTimeout(ctx, c.longPullReadTimeOut)
	defer cf()

	ticker := time.NewTicker(getAllNewTickerDuration)
	defer ticker.Stop()

	for {
		select {
		case <-timeOutContext.Done():
			return
		case <-ticker.C:
			res, ok, err = c.msgCache.Get(timeOutContext, strconv.Itoa(userUUID))
			if err != nil {
				slog.Error("read from cache error")
				res, err = c.msgRepo.GetUnread(ctx, userUUID, nil)
				if err != nil {
					return
				} else if len(res) > 0 {
					ok = true
					return
				}
			} else if ok {
				return
			}
		}
	}
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

func (c *ChatUseCase) GetHistory(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]domain.Msg, error) {
	return c.msgRepo.GetHistory(ctx, userUUID, ticketUUID, from, to)
}
