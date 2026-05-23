package usecase

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type chatMsgCacheMock struct {
	setFn           func(ctx context.Context, userUUID string, val *domain.Msg) error
	getFn           func(ctx context.Context, userUUID string) ([]*domain.Msg, bool, error)
	deleteByMsgIDFn func(ctx context.Context, userUUID string, id int) error
}

func (m *chatMsgCacheMock) Set(ctx context.Context, userUUID string, val *domain.Msg) error {
	return m.setFn(ctx, userUUID, val)
}

func (m *chatMsgCacheMock) Get(ctx context.Context, userUUID string) ([]*domain.Msg, bool, error) {
	return m.getFn(ctx, userUUID)
}

func (m *chatMsgCacheMock) DeleteByMsgID(ctx context.Context, userUUID string, id int) error {
	return m.deleteByMsgIDFn(ctx, userUUID, id)
}

type chatS3Mock struct {
	saveFn   func(ctx context.Context, folder string, file []byte) (string, error)
	deleteFn func(ctx context.Context, path string) error
}

func (m *chatS3Mock) Save(ctx context.Context, folder string, file []byte) (string, error) {
	return m.saveFn(ctx, folder, file)
}

func (m *chatS3Mock) Delete(ctx context.Context, path string) error {
	return m.deleteFn(ctx, path)
}

type chatMsgRepoMock struct {
	createForClientFn  func(ctx context.Context, msg domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error)
	createForManagerFn func(ctx context.Context, msg domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error)
	createFileFn       func(ctx context.Context, msgID int, fileName, path string, tx pgx.Tx) (domain.MsgFileContent, error)
	getUnreadFn        func(ctx context.Context, userUUID string, tx pgx.Tx) ([]*domain.Msg, error)
	markReadByIDFn     func(ctx context.Context, userUUID string, id int, tx pgx.Tx) error
	getHistoryFn       func(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]*domain.Msg, error)
}

func (m *chatMsgRepoMock) CreateForClient(ctx context.Context, msg domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
	return m.createForClientFn(ctx, msg, tx)
}

func (m *chatMsgRepoMock) CreateForManager(ctx context.Context, msg domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
	return m.createForManagerFn(ctx, msg, tx)
}

func (m *chatMsgRepoMock) CreateFile(ctx context.Context, msgID int, fileName, path string, tx pgx.Tx) (domain.MsgFileContent, error) {
	return m.createFileFn(ctx, msgID, fileName, path, tx)
}

func (m *chatMsgRepoMock) GetUnread(ctx context.Context, userUUID string, tx pgx.Tx) ([]*domain.Msg, error) {
	return m.getUnreadFn(ctx, userUUID, tx)
}

func (m *chatMsgRepoMock) MarkReadByID(ctx context.Context, userUUID string, id int, tx pgx.Tx) error {
	return m.markReadByIDFn(ctx, userUUID, id, tx)
}

func (m *chatMsgRepoMock) GetHistory(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]*domain.Msg, error) {
	return m.getHistoryFn(ctx, userUUID, ticketUUID, from, to)
}

type chatTicketRepoMock struct {
	getByIDFn func(ctx context.Context, id int) (domain.Ticket, error)
}

func (m *chatTicketRepoMock) GetByID(ctx context.Context, id int) (domain.Ticket, error) {
	return m.getByIDFn(ctx, id)
}

type chatEncryptionMock struct {
	encryptFn func(plaintext []byte, aad []byte) ([]byte, error)
	decryptFn func(data []byte, aad []byte) ([]byte, error)
}

func (m *chatEncryptionMock) Encrypt(plaintext []byte, aad []byte) ([]byte, error) {
	return m.encryptFn(plaintext, aad)
}

func (m *chatEncryptionMock) Decrypt(data []byte, aad []byte) ([]byte, error) {
	return m.decryptFn(data, aad)
}

type chatFixture struct {
	useCase    *ChatUseCase
	tx         *fakeTx
	mainRepo   *mainRepoMock
	msgCache   *chatMsgCacheMock
	s3         *chatS3Mock
	msgRepo    *chatMsgRepoMock
	ticketRepo *chatTicketRepoMock
	encryption *chatEncryptionMock
	msg        domain.CreateMsg
	ticket     domain.Ticket
	clientMsg  *domain.Msg
	managerMsg *domain.Msg
	history    []*domain.Msg
}

func newChatFixture() *chatFixture {
	txObj := &fakeTx{}

	msg := domain.CreateMsg{
		TicketID: 99,
		Text:     "hello",
	}
	ticket := domain.Ticket{
		ID:              msg.TicketID,
		ClientUserUUID:  "client-uuid",
		ManagerUserUUID: "manager-uuid",
	}
	clientMsg := &domain.Msg{
		ID:            10,
		FromType:      domain.MsgFromTypeManager,
		EncryptedText: []byte("enc:hello"),
		TicketID:      msg.TicketID,
		Status:        domain.MsgStatusSent,
	}
	managerMsg := &domain.Msg{
		ID:            11,
		FromType:      domain.MsgFromTypeClient,
		EncryptedText: []byte("enc:hello"),
		TicketID:      msg.TicketID,
		Status:        domain.MsgStatusSent,
	}
	history := []*domain.Msg{
		{
			ID:            20,
			TicketID:      msg.TicketID,
			EncryptedText: []byte("enc:history"),
			Status:        domain.MsgStatusRead,
		},
	}

	mainRepo := &mainRepoMock{
		createSessionFn: func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
			return txObj, nil
		},
	}
	msgCache := &chatMsgCacheMock{
		setFn: func(_ context.Context, _ string, _ *domain.Msg) error {
			return nil
		},
		getFn: func(_ context.Context, _ string) ([]*domain.Msg, bool, error) {
			return nil, false, nil
		},
		deleteByMsgIDFn: func(_ context.Context, _ string, _ int) error {
			return nil
		},
	}
	s3 := &chatS3Mock{
		saveFn: func(_ context.Context, folder string, file []byte) (string, error) {
			return folder + "/" + string(file), nil
		},
		deleteFn: func(_ context.Context, _ string) error {
			return nil
		},
	}
	msgRepo := &chatMsgRepoMock{
		createForClientFn: func(_ context.Context, createMsg domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
			if tx != txObj {
				return nil, errors.New("unexpected tx")
			}
			if createMsg.TicketID != msg.TicketID {
				return nil, errors.New("unexpected ticket id")
			}
			if createMsg.Text != "" {
				return nil, errors.New("expected empty text before db save")
			}
			if !bytes.Equal(createMsg.EncryptedText, []byte("enc:hello")) {
				return nil, errors.New("unexpected encrypted text")
			}
			saved := *clientMsg
			saved.EncryptedText = createMsg.EncryptedText
			return &saved, nil
		},
		createForManagerFn: func(_ context.Context, createMsg domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
			if tx != txObj {
				return nil, errors.New("unexpected tx")
			}
			if createMsg.TicketID != msg.TicketID {
				return nil, errors.New("unexpected ticket id")
			}
			if createMsg.Text != "" {
				return nil, errors.New("expected empty text before db save")
			}
			if !bytes.Equal(createMsg.EncryptedText, []byte("enc:hello")) {
				return nil, errors.New("unexpected encrypted text")
			}
			saved := *managerMsg
			saved.EncryptedText = createMsg.EncryptedText
			return &saved, nil
		},
		createFileFn: func(_ context.Context, msgID int, fileName, path string, tx pgx.Tx) (domain.MsgFileContent, error) {
			if tx != txObj {
				return domain.MsgFileContent{}, errors.New("unexpected tx")
			}
			return domain.MsgFileContent{
				ID:       msgID*100 + len(fileName),
				MsgID:    msgID,
				FileName: fileName,
				Path:     path,
			}, nil
		},
		getUnreadFn: func(_ context.Context, _ string, _ pgx.Tx) ([]*domain.Msg, error) {
			return nil, nil
		},
		markReadByIDFn: func(_ context.Context, _ string, _ int, _ pgx.Tx) error {
			return nil
		},
		getHistoryFn: func(_ context.Context, _ string, _ int, _, _ time.Time) ([]*domain.Msg, error) {
			res := make([]*domain.Msg, 0, len(history))
			for _, item := range history {
				copyItem := *item
				res = append(res, &copyItem)
			}
			return res, nil
		},
	}
	ticketRepo := &chatTicketRepoMock{
		getByIDFn: func(_ context.Context, id int) (domain.Ticket, error) {
			if id != ticket.ID {
				return domain.Ticket{}, errors.New("unexpected ticket id")
			}
			return ticket, nil
		},
	}
	encryption := &chatEncryptionMock{
		encryptFn: func(plaintext []byte, aad []byte) ([]byte, error) {
			if !bytes.Equal(aad, []byte("99")) {
				return nil, errors.New("unexpected aad")
			}
			return append([]byte("enc:"), plaintext...), nil
		},
		decryptFn: func(data []byte, aad []byte) ([]byte, error) {
			if !bytes.Equal(aad, []byte("99")) {
				return nil, errors.New("unexpected aad")
			}
			return bytes.TrimPrefix(data, []byte("enc:")), nil
		},
	}

	useCase := NewChatUseCase(msgCache, s3, msgRepo, ticketRepo, mainRepo, encryption, 20*time.Millisecond, time.Millisecond)

	return &chatFixture{
		useCase:    useCase,
		tx:         txObj,
		mainRepo:   mainRepo,
		msgCache:   msgCache,
		s3:         s3,
		msgRepo:    msgRepo,
		ticketRepo: ticketRepo,
		encryption: encryption,
		msg:        msg,
		ticket:     ticket,
		clientMsg:  clientMsg,
		managerMsg: managerMsg,
		history:    history,
	}
}

func Test_NewChatUseCase(t *testing.T) {
	f := newChatFixture()

	if f.useCase == nil {
		t.Fatal("NewChatUseCase returned nil")
	}
	if f.useCase.mainRepo != f.mainRepo {
		t.Fatal("unexpected mainRepo")
	}
	if f.useCase.ticketRepo != f.ticketRepo {
		t.Fatal("unexpected ticketRepo")
	}
	if f.useCase.s3 != S3Interface(f.s3) {
		t.Fatal("unexpected s3")
	}
	if f.useCase.encryption != f.encryption {
		t.Fatal("unexpected encryption")
	}
}

func Test_ChatUseCase_Send_TicketError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("ticket error")
	f.ticketRepo.getByIDFn = func(_ context.Context, id int) (domain.Ticket, error) {
		if id != f.msg.TicketID {
			t.Fatalf("unexpected ticket id: got=%d want=%d", id, f.msg.TicketID)
		}
		return domain.Ticket{}, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected ticket error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_EncryptError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("encrypt error")
	f.encryption.encryptFn = func(_ []byte, _ []byte) ([]byte, error) {
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected encrypt error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_CreateSessionError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create session error")
	f.mainRepo.createSessionFn = func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create session error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_FileToS3Error(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("save error")
	f.msg.Files = []domain.CreateFile{{Name: "a.txt", Data: []byte("a")}}
	f.s3.saveFn = func(_ context.Context, _ string, _ []byte) (string, error) {
		return "", wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected save error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_ManagerDBError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create for client error")
	f.msgRepo.createForClientFn = func(_ context.Context, _ domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
		if tx != f.tx {
			t.Fatal("expected tx in CreateForClient")
		}
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create for client error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_ClientDBError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create for manager error")
	f.msgRepo.createForManagerFn = func(_ context.Context, _ domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
		if tx != f.tx {
			t.Fatal("expected tx in CreateForManager")
		}
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeClient)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create for manager error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_SystemDBError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("system db error")
	f.msgRepo.createForClientFn = func(_ context.Context, _ domain.CreateMsg, _ pgx.Tx) (*domain.Msg, error) {
		return nil, wantErr
	}
	f.msgRepo.createForManagerFn = func(_ context.Context, _ domain.CreateMsg, _ pgx.Tx) (*domain.Msg, error) {
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeSystem)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected system db error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_SystemSecondDBError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("system second db error")
	f.msgRepo.createForManagerFn = func(_ context.Context, _ domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
		if tx != f.tx {
			t.Fatal("expected tx in CreateForManager")
		}
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeSystem)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected system second db error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_CleanupFilesOnDBError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create for client error")
	deleted := make(chan string, 1)
	f.msg.Files = []domain.CreateFile{{Name: "a.txt", Data: []byte("a")}}
	f.msgRepo.createForClientFn = func(_ context.Context, _ domain.CreateMsg, tx pgx.Tx) (*domain.Msg, error) {
		if tx != f.tx {
			t.Fatal("expected tx in CreateForClient")
		}
		return nil, wantErr
	}
	f.s3.deleteFn = func(_ context.Context, path string) error {
		deleted <- path
		return nil
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create for client error, got=%v", err)
	}

	select {
	case path := <-deleted:
		if path != "ticket/file/99/a" {
			t.Fatalf("unexpected deleted path: got=%q", path)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for s3 delete")
	}
}

func Test_ChatUseCase_Send_UnknownFromType(t *testing.T) {
	f := newChatFixture()

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromType("unknown"))
	if err == nil {
		t.Fatal("expected error for unknown MsgFromType")
	}
	if err.Error() != "Unknown MsgFromType" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_ChatUseCase_Send_CommitError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("commit error")
	f.tx.commitErr = wantErr

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected commit error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_OK_Manager_WithFiles(t *testing.T) {
	f := newChatFixture()
	done := make(chan *domain.Msg, 1)
	f.msg.Files = []domain.CreateFile{
		{Name: "a.txt", Data: []byte("a")},
		{Name: "b.txt", Data: []byte("b")},
	}
	f.msgCache.setFn = func(_ context.Context, userUUID string, val *domain.Msg) error {
		if userUUID != f.ticket.ClientUserUUID {
			t.Fatalf("unexpected cache key: got=%q want=%q", userUUID, f.ticket.ClientUserUUID)
		}
		done <- val
		return nil
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit once, got=%d", f.tx.commitCalls)
	}

	select {
	case cachedMsg := <-done:
		if len(cachedMsg.Files) != 2 {
			t.Fatalf("unexpected files len: got=%d want=2", len(cachedMsg.Files))
		}
		if cachedMsg.Files[0].FileName != "a.txt" || cachedMsg.Files[1].FileName != "b.txt" {
			t.Fatalf("unexpected file names: got=%+v", cachedMsg.Files)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for cache.Set")
	}
}

func Test_ChatUseCase_Send_OK_Client(t *testing.T) {
	f := newChatFixture()
	done := make(chan string, 1)
	f.msgCache.setFn = func(_ context.Context, userUUID string, _ *domain.Msg) error {
		done <- userUUID
		return nil
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeClient)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	select {
	case userUUID := <-done:
		if userUUID != f.ticket.ManagerUserUUID {
			t.Fatalf("unexpected cache key: got=%q want=%q", userUUID, f.ticket.ManagerUserUUID)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for cache.Set")
	}
}

func Test_ChatUseCase_Send_OK_System(t *testing.T) {
	f := newChatFixture()
	done := make(chan string, 2)
	f.msgCache.setFn = func(_ context.Context, userUUID string, _ *domain.Msg) error {
		done <- userUUID
		return nil
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeSystem)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	got := map[string]int{}
	for i := 0; i < 2; i++ {
		select {
		case userUUID := <-done:
			got[userUUID]++
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timed out waiting for cache.Set calls")
		}
	}
	if got[f.ticket.ClientUserUUID] != 1 || got[f.ticket.ManagerUserUUID] != 1 {
		t.Fatalf("unexpected cache recipients: got=%v", got)
	}
}

func Test_ChatUseCase_GetAllNew_ContextDone(t *testing.T) {
	f := newChatFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, ok, err := f.useCase.GetAllNew(ctx, "user-1")
	if err != nil {
		t.Fatalf("GetAllNew returned error: %v", err)
	}
	if ok {
		t.Fatal("expected ok to be false")
	}
	if res != nil {
		t.Fatalf("expected nil result, got=%v", res)
	}
}

func Test_ChatUseCase_GetAllNew_CacheHit(t *testing.T) {
	f := newChatFixture()
	f.useCase.pollInterval = time.Millisecond
	want := []*domain.Msg{
		{ID: 1, TicketID: f.msg.TicketID, EncryptedText: []byte("enc:cache")},
	}
	f.msgCache.getFn = func(_ context.Context, userUUID string) ([]*domain.Msg, bool, error) {
		if userUUID != "user-42" {
			t.Fatalf("unexpected cache key: got=%q want=%q", userUUID, "user-42")
		}
		return want, true, nil
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), "user-42")
	if err != nil {
		t.Fatalf("GetAllNew returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if len(res) != 1 || res[0].Text != "cache" {
		t.Fatalf("unexpected result: got=%v", res)
	}
}

func Test_ChatUseCase_GetAllNew_CacheErrorUnreadError(t *testing.T) {
	f := newChatFixture()
	f.useCase.pollInterval = time.Millisecond
	wantErr := errors.New("unread error")
	f.msgCache.getFn = func(_ context.Context, _ string) ([]*domain.Msg, bool, error) {
		return nil, false, errors.New("cache error")
	}
	f.msgRepo.getUnreadFn = func(_ context.Context, userUUID string, tx pgx.Tx) ([]*domain.Msg, error) {
		if userUUID != "user-7" {
			t.Fatalf("unexpected user uuid: got=%q want=%q", userUUID, "user-7")
		}
		if tx != nil {
			t.Fatal("expected nil tx in GetUnread")
		}
		return nil, wantErr
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), "user-7")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected unread error, got=%v", err)
	}
	if ok {
		t.Fatal("expected ok to be false")
	}
	if res != nil {
		t.Fatalf("expected nil result, got=%v", res)
	}
}

func Test_ChatUseCase_GetAllNew_CacheErrorUnreadFound(t *testing.T) {
	f := newChatFixture()
	f.useCase.pollInterval = time.Millisecond
	want := []*domain.Msg{
		{ID: 5, TicketID: f.msg.TicketID, EncryptedText: []byte("enc:fallback")},
	}
	f.msgCache.getFn = func(_ context.Context, _ string) ([]*domain.Msg, bool, error) {
		return nil, false, errors.New("cache error")
	}
	f.msgRepo.getUnreadFn = func(_ context.Context, _ string, _ pgx.Tx) ([]*domain.Msg, error) {
		return want, nil
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), "user-7")
	if err != nil {
		t.Fatalf("GetAllNew returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if len(res) != 1 || res[0].Text != "fallback" {
		t.Fatalf("unexpected result: got=%v", res)
	}
}

func Test_ChatUseCase_GetAllNew_CacheMissTimeout(t *testing.T) {
	f := newChatFixture()
	f.useCase.pollInterval = time.Millisecond
	f.useCase.longPullReadTimeOut = 5 * time.Millisecond
	cacheReads := 0
	f.msgCache.getFn = func(_ context.Context, _ string) ([]*domain.Msg, bool, error) {
		cacheReads++
		return nil, false, nil
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), "user-9")
	if err != nil {
		t.Fatalf("GetAllNew returned error: %v", err)
	}
	if ok {
		t.Fatal("expected ok to be false")
	}
	if res != nil {
		t.Fatalf("expected nil result, got=%v", res)
	}
	if cacheReads == 0 {
		t.Fatal("expected cache.Get to be called at least once")
	}
}

func Test_ChatUseCase_GetAllNew_DecryptError(t *testing.T) {
	f := newChatFixture()
	f.useCase.pollInterval = time.Millisecond
	wantErr := errors.New("decrypt error")
	f.msgCache.getFn = func(_ context.Context, _ string) ([]*domain.Msg, bool, error) {
		return []*domain.Msg{{ID: 1, TicketID: f.msg.TicketID, EncryptedText: []byte("boom")}}, true, nil
	}
	f.encryption.decryptFn = func(_ []byte, _ []byte) ([]byte, error) {
		return nil, wantErr
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), "user-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected decrypt error, got=%v", err)
	}
	if ok != true {
		t.Fatal("expected ok to stay true")
	}
	if len(res) != 1 {
		t.Fatalf("expected one result, got=%v", res)
	}
}

func Test_ChatUseCase_MarkAsRead_CreateSessionError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create session error")
	f.mainRepo.createSessionFn = func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
		return nil, wantErr
	}

	err := f.useCase.MarkAsRead(context.Background(), "user-1", []int{1})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create session error, got=%v", err)
	}
}

func Test_ChatUseCase_MarkAsRead_MarkReadError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("mark read error")
	f.msgRepo.markReadByIDFn = func(_ context.Context, userUUID string, id int, tx pgx.Tx) error {
		if userUUID != "user-1" {
			t.Fatalf("unexpected user uuid: got=%q want=%q", userUUID, "user-1")
		}
		if id != 10 {
			t.Fatalf("unexpected msg id: got=%d want=%d", id, 10)
		}
		if tx != f.tx {
			t.Fatal("expected tx in MarkReadByID")
		}
		return wantErr
	}

	err := f.useCase.MarkAsRead(context.Background(), "user-1", []int{10})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected mark read error, got=%v", err)
	}
}

func Test_ChatUseCase_MarkAsRead_CommitError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("commit error")
	f.tx.commitErr = wantErr

	err := f.useCase.MarkAsRead(context.Background(), "user-1", []int{10})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected commit error, got=%v", err)
	}
}

func Test_ChatUseCase_MarkAsRead_OK(t *testing.T) {
	f := newChatFixture()
	deleted := make(chan int, 2)
	f.msgCache.deleteByMsgIDFn = func(_ context.Context, userUUID string, id int) error {
		if userUUID != "user-1" {
			t.Fatalf("unexpected user uuid: got=%q want=%q", userUUID, "user-1")
		}
		deleted <- id
		return nil
	}
	seen := make([]int, 0, 2)
	f.msgRepo.markReadByIDFn = func(_ context.Context, _ string, id int, tx pgx.Tx) error {
		if tx != f.tx {
			t.Fatal("expected tx in MarkReadByID")
		}
		seen = append(seen, id)
		return nil
	}

	err := f.useCase.MarkAsRead(context.Background(), "user-1", []int{10, 11})
	if err != nil {
		t.Fatalf("MarkAsRead returned error: %v", err)
	}
	if len(seen) != 2 || seen[0] != 10 || seen[1] != 11 {
		t.Fatalf("unexpected marked ids: got=%v", seen)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit once, got=%d", f.tx.commitCalls)
	}

	gotDeleted := map[int]int{}
	for i := 0; i < 2; i++ {
		select {
		case id := <-deleted:
			gotDeleted[id]++
		case <-time.After(200 * time.Millisecond):
			t.Fatal("timed out waiting for delete calls")
		}
	}
	if gotDeleted[10] != 1 || gotDeleted[11] != 1 {
		t.Fatalf("unexpected deleted ids: got=%v", gotDeleted)
	}
}

func Test_ChatUseCase_GetHistory_RepoError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("history error")
	f.msgRepo.getHistoryFn = func(_ context.Context, _ string, _ int, _, _ time.Time) ([]*domain.Msg, error) {
		return nil, wantErr
	}

	history, err := f.useCase.GetHistory(context.Background(), "user-1", 99, time.Time{}, time.Time{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected history error, got=%v", err)
	}
	if history != nil {
		t.Fatalf("expected nil history, got=%v", history)
	}
}

func Test_ChatUseCase_GetHistory_DecryptError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("decrypt error")
	f.encryption.decryptFn = func(_ []byte, _ []byte) ([]byte, error) {
		return nil, wantErr
	}

	history, err := f.useCase.GetHistory(context.Background(), "user-1", 99, time.Time{}, time.Time{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected decrypt error, got=%v", err)
	}
	if history != nil {
		t.Fatalf("expected nil history, got=%v", history)
	}
}

func Test_ChatUseCase_GetHistory_OK(t *testing.T) {
	f := newChatFixture()
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)

	history, err := f.useCase.GetHistory(context.Background(), "user-1", 99, from, to)
	if err != nil {
		t.Fatalf("GetHistory returned error: %v", err)
	}
	if len(history) != 1 || history[0].Text != "history" {
		t.Fatalf("unexpected history: got=%v", history)
	}
}

func Test_ChatUseCase_fileToS3_TooManyFiles(t *testing.T) {
	f := newChatFixture()

	files, pathes, err := f.useCase.fileToS3(context.Background(), make([]domain.CreateFile, 11), f.msg.TicketID)
	if !errors.Is(err, domain.ErrFilesLoad) {
		t.Fatalf("expected ErrFilesLoad, got=%v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty files result, got=%v", files)
	}
	if len(pathes) != 0 {
		t.Fatalf("expected empty paths result, got=%v", pathes)
	}
}

func Test_ChatUseCase_fileToS3_FileTooLarge(t *testing.T) {
	f := newChatFixture()

	files, pathes, err := f.useCase.fileToS3(context.Background(), []domain.CreateFile{{Name: "big.bin", Data: make([]byte, 16*1024*1024)}}, f.msg.TicketID)
	if !errors.Is(err, domain.ErrFilesLoad) {
		t.Fatalf("expected ErrFilesLoad, got=%v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected empty files result, got=%v", files)
	}
	if len(pathes) != 0 {
		t.Fatalf("expected empty paths result, got=%v", pathes)
	}
}

func Test_ChatUseCase_fileToS3_SaveError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("save error")
	f.s3.saveFn = func(_ context.Context, _ string, _ []byte) (string, error) {
		return "", wantErr
	}

	files, pathes, err := f.useCase.fileToS3(context.Background(), []domain.CreateFile{{Name: "a.txt", Data: []byte("a")}}, f.msg.TicketID)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected save error, got=%v", err)
	}
	if len(files) != 1 || files[0].Path != "" {
		t.Fatalf("unexpected files result: got=%v", files)
	}
	if len(pathes) != 0 {
		t.Fatalf("expected empty paths result, got=%v", pathes)
	}
}

func Test_ChatUseCase_fileToS3_PartialSaveError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("save error")
	f.s3.saveFn = func(_ context.Context, folder string, file []byte) (string, error) {
		if bytes.Equal(file, []byte("bad")) {
			return "", wantErr
		}
		return folder + "/" + string(file), nil
	}

	files, pathes, err := f.useCase.fileToS3(context.Background(), []domain.CreateFile{
		{Name: "ok.txt", Data: []byte("ok")},
		{Name: "bad.txt", Data: []byte("bad")},
	}, f.msg.TicketID)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected save error, got=%v", err)
	}
	if len(files) != 2 {
		t.Fatalf("unexpected files result: got=%v", files)
	}
	if len(pathes) != 1 || pathes[0] != "ticket/file/99/ok" {
		t.Fatalf("unexpected paths result: got=%v", pathes)
	}
}

func Test_ChatUseCase_fileToS3_OK(t *testing.T) {
	f := newChatFixture()

	files, pathes, err := f.useCase.fileToS3(context.Background(), []domain.CreateFile{{Name: "a.txt", Data: []byte("a")}}, f.msg.TicketID)
	if err != nil {
		t.Fatalf("fileToS3 returned error: %v", err)
	}
	if len(files) != 1 || files[0].Path == "" {
		t.Fatalf("unexpected files result: got=%v", files)
	}
	if len(pathes) != 1 || pathes[0] != files[0].Path {
		t.Fatalf("unexpected paths result: got=%v", pathes)
	}
}

func Test_ChatUseCase_encryptMsgText_OK(t *testing.T) {
	f := newChatFixture()
	msg := domain.CreateMsg{TicketID: 77, Text: "secret"}
	f.encryption.encryptFn = func(plaintext []byte, aad []byte) ([]byte, error) {
		if !bytes.Equal(plaintext, []byte("secret")) {
			t.Fatalf("unexpected plaintext: got=%q", string(plaintext))
		}
		if !bytes.Equal(aad, []byte("77")) {
			t.Fatalf("unexpected aad: got=%q", string(aad))
		}
		return []byte("cipher"), nil
	}

	err := f.useCase.encryptMsgText(&msg)
	if err != nil {
		t.Fatalf("encryptMsgText returned error: %v", err)
	}
	if string(msg.EncryptedText) != "cipher" {
		t.Fatalf("unexpected encrypted text: got=%q", string(msg.EncryptedText))
	}
}

func Test_ChatUseCase_encryptMsgText_Error(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("encrypt error")
	f.encryption.encryptFn = func(_ []byte, _ []byte) ([]byte, error) {
		return nil, wantErr
	}

	err := f.useCase.encryptMsgText(&domain.CreateMsg{TicketID: 1, Text: "x"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected encrypt error, got=%v", err)
	}
}

func Test_ChatUseCase_decryptMsgText_OK(t *testing.T) {
	f := newChatFixture()
	msg := &domain.Msg{TicketID: 77, EncryptedText: []byte("cipher")}
	f.encryption.decryptFn = func(data []byte, aad []byte) ([]byte, error) {
		if !bytes.Equal(data, []byte("cipher")) {
			t.Fatalf("unexpected data: got=%q", string(data))
		}
		if !bytes.Equal(aad, []byte("77")) {
			t.Fatalf("unexpected aad: got=%q", string(aad))
		}
		return []byte("plain"), nil
	}

	err := f.useCase.decryptMsgText(msg)
	if err != nil {
		t.Fatalf("decryptMsgText returned error: %v", err)
	}
	if msg.Text != "plain" {
		t.Fatalf("unexpected decrypted text: got=%q", msg.Text)
	}
}

func Test_ChatUseCase_decryptMsgText_Error(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("decrypt error")
	f.encryption.decryptFn = func(_ []byte, _ []byte) ([]byte, error) {
		return nil, wantErr
	}

	err := f.useCase.decryptMsgText(&domain.Msg{TicketID: 1, EncryptedText: []byte("x")})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected decrypt error, got=%v", err)
	}
}

func Test_ChatUseCase_createAdd(t *testing.T) {
	f := newChatFixture()

	got := f.useCase.createAdd(123)
	if string(got) != "123" {
		t.Fatalf("unexpected aad: got=%q want=%q", string(got), "123")
	}
}

func Test_ChatUseCase_createForClientDB_CreateError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create client msg error")
	f.msgRepo.createForClientFn = func(_ context.Context, _ domain.CreateMsg, _ pgx.Tx) (*domain.Msg, error) {
		return nil, wantErr
	}

	msg, err := f.useCase.createForClientDB(context.Background(), domain.CreateMsg{
		TicketID:      f.msg.TicketID,
		EncryptedText: []byte("enc:hello"),
	}, &f.ticket, nil, f.tx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create error, got=%v", err)
	}
	if msg != nil {
		t.Fatalf("expected nil msg, got=%v", msg)
	}
}

func Test_ChatUseCase_createForClientDB_CreateFileError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create file error")
	f.msgRepo.createFileFn = func(_ context.Context, _ int, _, _ string, _ pgx.Tx) (domain.MsgFileContent, error) {
		return domain.MsgFileContent{}, wantErr
	}

	msg, err := f.useCase.createForClientDB(context.Background(), domain.CreateMsg{
		TicketID:      f.msg.TicketID,
		EncryptedText: []byte("enc:hello"),
		Files:         []domain.CreateFile{{Name: "a.txt"}},
	}, &f.ticket, []domain.CreateFile{{Name: "a.txt", Path: "/tmp/a"}}, f.tx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create file error, got=%v", err)
	}
	if msg != nil {
		t.Fatalf("expected nil msg, got=%v", msg)
	}
}

func Test_ChatUseCase_createForClientDB_OK(t *testing.T) {
	f := newChatFixture()

	msg, err := f.useCase.createForClientDB(context.Background(), domain.CreateMsg{
		TicketID:      f.msg.TicketID,
		EncryptedText: []byte("enc:hello"),
		Files:         []domain.CreateFile{{Name: "a.txt"}},
	}, &f.ticket, []domain.CreateFile{{Name: "a.txt", Path: "/tmp/a"}}, f.tx)
	if err != nil {
		t.Fatalf("createForClientDB returned error: %v", err)
	}
	if len(msg.Files) != 1 || msg.Files[0].FileName != "a.txt" {
		t.Fatalf("unexpected msg files: got=%v", msg.Files)
	}
}

func Test_ChatUseCase_createForClientCache_OK(t *testing.T) {
	f := newChatFixture()
	called := false
	f.msgCache.setFn = func(_ context.Context, userUUID string, val *domain.Msg) error {
		called = true
		if userUUID != f.ticket.ClientUserUUID {
			t.Fatalf("unexpected cache key: got=%q want=%q", userUUID, f.ticket.ClientUserUUID)
		}
		if val != f.clientMsg {
			t.Fatal("unexpected message pointer")
		}
		return nil
	}

	err := f.useCase.createForClientCache(context.Background(), f.clientMsg, &f.ticket)
	if err != nil {
		t.Fatalf("createForClientCache returned error: %v", err)
	}
	if !called {
		t.Fatal("expected cache.Set to be called")
	}
}

func Test_ChatUseCase_createForClientCache_Error(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("cache set error")
	f.msgCache.setFn = func(_ context.Context, _ string, _ *domain.Msg) error {
		return wantErr
	}

	err := f.useCase.createForClientCache(context.Background(), f.clientMsg, &f.ticket)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected cache set error, got=%v", err)
	}
}

func Test_ChatUseCase_createForManagerDB_CreateError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create manager msg error")
	f.msgRepo.createForManagerFn = func(_ context.Context, _ domain.CreateMsg, _ pgx.Tx) (*domain.Msg, error) {
		return nil, wantErr
	}

	msg, err := f.useCase.createForManagerDB(context.Background(), domain.CreateMsg{
		TicketID:      f.msg.TicketID,
		EncryptedText: []byte("enc:hello"),
	}, &f.ticket, nil, f.tx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create error, got=%v", err)
	}
	if msg != nil {
		t.Fatalf("expected nil msg, got=%v", msg)
	}
}

func Test_ChatUseCase_createForManagerDB_CreateFileError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create file error")
	f.msgRepo.createFileFn = func(_ context.Context, _ int, _, _ string, _ pgx.Tx) (domain.MsgFileContent, error) {
		return domain.MsgFileContent{}, wantErr
	}

	msg, err := f.useCase.createForManagerDB(context.Background(), domain.CreateMsg{
		TicketID:      f.msg.TicketID,
		EncryptedText: []byte("enc:hello"),
		Files:         []domain.CreateFile{{Name: "a.txt"}},
	}, &f.ticket, []domain.CreateFile{{Name: "a.txt", Path: "/tmp/a"}}, f.tx)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create file error, got=%v", err)
	}
	if msg != nil {
		t.Fatalf("expected nil msg, got=%v", msg)
	}
}

func Test_ChatUseCase_createForManagerDB_OK(t *testing.T) {
	f := newChatFixture()

	msg, err := f.useCase.createForManagerDB(context.Background(), domain.CreateMsg{
		TicketID:      f.msg.TicketID,
		EncryptedText: []byte("enc:hello"),
		Files:         []domain.CreateFile{{Name: "a.txt"}},
	}, &f.ticket, []domain.CreateFile{{Name: "a.txt", Path: "/tmp/a"}}, f.tx)
	if err != nil {
		t.Fatalf("createForManagerDB returned error: %v", err)
	}
	if len(msg.Files) != 1 || msg.Files[0].FileName != "a.txt" {
		t.Fatalf("unexpected msg files: got=%v", msg.Files)
	}
}

func Test_ChatUseCase_createForManagerCache_OK(t *testing.T) {
	f := newChatFixture()
	called := false
	f.msgCache.setFn = func(_ context.Context, userUUID string, val *domain.Msg) error {
		called = true
		if userUUID != f.ticket.ManagerUserUUID {
			t.Fatalf("unexpected cache key: got=%q want=%q", userUUID, f.ticket.ManagerUserUUID)
		}
		if val != f.managerMsg {
			t.Fatal("unexpected message pointer")
		}
		return nil
	}

	err := f.useCase.createForManagerCache(context.Background(), f.managerMsg, &f.ticket)
	if err != nil {
		t.Fatalf("createForManagerCache returned error: %v", err)
	}
	if !called {
		t.Fatal("expected cache.Set to be called")
	}
}

func Test_ChatUseCase_createForManagerCache_Error(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("cache set error")
	f.msgCache.setFn = func(_ context.Context, _ string, _ *domain.Msg) error {
		return wantErr
	}

	err := f.useCase.createForManagerCache(context.Background(), f.managerMsg, &f.ticket)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected cache set error, got=%v", err)
	}
}

func Test_ChatUseCase_deleteFormS3_Error(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("delete error")
	f.s3.deleteFn = func(_ context.Context, path string) error {
		if path != "/tmp/a" {
			t.Fatalf("unexpected path: got=%q", path)
		}
		return wantErr
	}

	err := f.useCase.deleteFormS3(context.Background(), []string{"/tmp/a"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected delete error, got=%v", err)
	}
}

func Test_ChatUseCase_deleteFormS3_OK(t *testing.T) {
	f := newChatFixture()
	var mu sync.Mutex
	seen := make(map[string]int, 2)
	f.s3.deleteFn = func(_ context.Context, path string) error {
		mu.Lock()
		seen[path]++
		mu.Unlock()
		return nil
	}

	err := f.useCase.deleteFormS3(context.Background(), []string{"/tmp/a", "/tmp/b"})
	if err != nil {
		t.Fatalf("deleteFormS3 returned error: %v", err)
	}
	if seen["/tmp/a"] != 1 || seen["/tmp/b"] != 1 {
		t.Fatalf("unexpected deleted paths: got=%v", seen)
	}
}
