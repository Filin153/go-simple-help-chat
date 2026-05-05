package usecase

import (
	"context"
	"errors"
	"strconv"
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
	saveFn func(ctx context.Context, folder string, file []byte) (string, error)
}

func (m *chatS3Mock) Save(ctx context.Context, folder string, file []byte) (string, error) {
	return m.saveFn(ctx, folder, file)
}

type chatMsgRepoMock struct {
	createFn       func(ctx context.Context, msg domain.CreateMsg, fromType domain.MsgFromType, draft bool, tx pgx.Tx) (*domain.Msg, error)
	createFileFn   func(ctx context.Context, msgID int, path string, tx pgx.Tx) (domain.MsgFileContent, error)
	getUnreadFn    func(ctx context.Context, userUUID int, tx pgx.Tx) ([]*domain.Msg, error)
	markReadByIDFn func(ctx context.Context, userUUID string, id int, tx pgx.Tx) error
	getHistoryFn   func(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]*domain.Msg, error)
}

func (m *chatMsgRepoMock) Create(ctx context.Context, msg domain.CreateMsg, fromType domain.MsgFromType, draft bool, tx pgx.Tx) (*domain.Msg, error) {
	return m.createFn(ctx, msg, fromType, draft, tx)
}

func (m *chatMsgRepoMock) CreateFile(ctx context.Context, msgID int, path string, tx pgx.Tx) (domain.MsgFileContent, error) {
	return m.createFileFn(ctx, msgID, path, tx)
}

func (m *chatMsgRepoMock) GetUnread(ctx context.Context, userUUID int, tx pgx.Tx) ([]*domain.Msg, error) {
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

type chatFixture struct {
	useCase    *ChatUseCase
	tx         *fakeTx
	mainRepo   *mainRepoMock
	msgCache   *chatMsgCacheMock
	s3         *chatS3Mock
	msgRepo    *chatMsgRepoMock
	ticketRepo *chatTicketRepoMock
	msg        domain.CreateMsg
	savedMsg   *domain.Msg
	ticket     domain.Ticket
	history    []*domain.Msg
}

func newChatFixture() *chatFixture {
	txObj := &fakeTx{}
	msg := domain.CreateMsg{
		TicketID: 99,
		Text:     "hello",
	}
	savedMsg := &domain.Msg{
		ID:       15,
		FromType: domain.MsgFromTypeManager,
		Text:     msg.Text,
		TicketID: msg.TicketID,
		Status:   domain.MsgStatusSent,
	}
	ticket := domain.Ticket{
		ID:              msg.TicketID,
		ManagerUserUUID: "manager-uuid",
		ClientUserUUID:  "client-uuid",
	}
	history := []*domain.Msg{savedMsg}

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
		deleteByMsgIDFn: func(_ context.Context, _ string, _ int) error { return nil },
	}
	s3 := &chatS3Mock{
		saveFn: func(_ context.Context, folder string, file []byte) (string, error) {
			return folder + "/" + string(file), nil
		},
	}
	msgRepo := &chatMsgRepoMock{
		createFn: func(_ context.Context, createMsg domain.CreateMsg, fromType domain.MsgFromType, draft bool, tx pgx.Tx) (*domain.Msg, error) {
			if tx != txObj {
				return nil, errors.New("unexpected tx")
			}
			if createMsg.TicketID != msg.TicketID {
				return nil, errors.New("unexpected ticket id")
			}
			if draft {
				return nil, errors.New("unexpected draft flag")
			}
			savedCopy := *savedMsg
			savedCopy.FromType = fromType
			return &savedCopy, nil
		},
		createFileFn: func(_ context.Context, msgID int, path string, tx pgx.Tx) (domain.MsgFileContent, error) {
			if msgID != savedMsg.ID {
				return domain.MsgFileContent{}, errors.New("unexpected msg id")
			}
			if tx != txObj {
				return domain.MsgFileContent{}, errors.New("unexpected tx")
			}
			return domain.MsgFileContent{ID: 1, MsgID: msgID, Path: path}, nil
		},
		getUnreadFn: func(_ context.Context, _ int, _ pgx.Tx) ([]*domain.Msg, error) {
			return nil, nil
		},
		markReadByIDFn: func(_ context.Context, _ string, _ int, _ pgx.Tx) error {
			return nil
		},
		getHistoryFn: func(_ context.Context, _ string, _ int, _, _ time.Time) ([]*domain.Msg, error) {
			return history, nil
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

	useCase := NewChatUseCase(msgCache, s3, msgRepo, ticketRepo, mainRepo, 20*time.Millisecond)

	return &chatFixture{
		useCase:    useCase,
		tx:         txObj,
		mainRepo:   mainRepo,
		msgCache:   msgCache,
		s3:         s3,
		msgRepo:    msgRepo,
		ticketRepo: ticketRepo,
		msg:        msg,
		savedMsg:   savedMsg,
		ticket:     ticket,
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
	if f.useCase.msgRepo != MsgRepo(f.msgRepo) {
		t.Fatal("unexpected msgRepo")
	}
	if f.useCase.msgCache != MsgCacheInterface(f.msgCache) {
		t.Fatal("unexpected msgCache")
	}
	if f.useCase.s3 != f.s3 {
		t.Fatal("unexpected s3")
	}
	if f.useCase.ticketRepo != f.ticketRepo {
		t.Fatal("unexpected ticketRepo")
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

func Test_ChatUseCase_Send_CreateError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create error")
	f.msgRepo.createFn = func(_ context.Context, _ domain.CreateMsg, _ domain.MsgFromType, _ bool, tx pgx.Tx) (*domain.Msg, error) {
		if tx != f.tx {
			t.Fatal("expected tx in Create")
		}
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_TooManyFiles(t *testing.T) {
	f := newChatFixture()
	s3Called := false
	f.s3.saveFn = func(_ context.Context, _ string, _ []byte) (string, error) {
		s3Called = true
		return "", nil
	}
	f.msg.Files = make([][]byte, 11)

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, domain.ErrFilesLoad) {
		t.Fatalf("expected ErrFilesLoad, got=%v", err)
	}
	if s3Called {
		t.Fatal("s3.Save must not be called when too many files are provided")
	}
}

func Test_ChatUseCase_Send_FileTooLarge(t *testing.T) {
	f := newChatFixture()
	s3Called := false
	f.s3.saveFn = func(_ context.Context, _ string, _ []byte) (string, error) {
		s3Called = true
		return "", nil
	}
	f.msg.Files = [][]byte{make([]byte, 16*1024*1024)}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, domain.ErrFilesLoad) {
		t.Fatalf("expected ErrFilesLoad, got=%v", err)
	}
	if s3Called {
		t.Fatal("s3.Save must not be called when file is too large")
	}
}

func Test_ChatUseCase_Send_S3Error(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("s3 error")
	f.msg.Files = [][]byte{[]byte("a")}
	f.s3.saveFn = func(_ context.Context, folder string, file []byte) (string, error) {
		if folder != "ticket/file/"+strconv.Itoa(f.savedMsg.ID) {
			t.Fatalf("unexpected folder: got=%q", folder)
		}
		if string(file) != "a" {
			t.Fatalf("unexpected file payload: got=%q", string(file))
		}
		return "", wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected s3 error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_CreateFileError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("create file error")
	f.msg.Files = [][]byte{[]byte("a")}
	f.msgRepo.createFileFn = func(_ context.Context, msgID int, path string, tx pgx.Tx) (domain.MsgFileContent, error) {
		if msgID != f.savedMsg.ID {
			t.Fatalf("unexpected msg id: got=%d want=%d", msgID, f.savedMsg.ID)
		}
		if path == "" {
			t.Fatal("expected non-empty path")
		}
		if tx != f.tx {
			t.Fatal("expected tx in CreateFile")
		}
		return domain.MsgFileContent{}, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create file error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_TicketError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("ticket error")
	f.ticketRepo.getByIDFn = func(_ context.Context, id int) (domain.Ticket, error) {
		if id != f.savedMsg.TicketID {
			t.Fatalf("unexpected ticket id: got=%d want=%d", id, f.savedMsg.TicketID)
		}
		return domain.Ticket{}, wantErr
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected ticket error, got=%v", err)
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
	f.msg.Files = [][]byte{[]byte("a"), []byte("b")}
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
			t.Fatalf("unexpected files len in cached msg: got=%d want=2", len(cachedMsg.Files))
		}
		if cachedMsg.Files[0].MsgID != f.savedMsg.ID || cachedMsg.Files[1].MsgID != f.savedMsg.ID {
			t.Fatalf("unexpected msg ids in cached files: got=%+v", cachedMsg.Files)
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

func Test_ChatUseCase_Send_CacheSetError_Manager(t *testing.T) {
	f := newChatFixture()
	done := make(chan string, 1)
	f.msgCache.setFn = func(_ context.Context, userUUID string, _ *domain.Msg) error {
		done <- userUUID
		return errors.New("cache set error")
	}

	err := f.useCase.Send(context.Background(), f.msg, domain.MsgFromTypeManager)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	select {
	case userUUID := <-done:
		if userUUID != f.ticket.ClientUserUUID {
			t.Fatalf("unexpected cache key: got=%q want=%q", userUUID, f.ticket.ClientUserUUID)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for cache.Set")
	}
}

func Test_ChatUseCase_Send_CacheSetError_Client(t *testing.T) {
	f := newChatFixture()
	done := make(chan string, 1)
	f.msgCache.setFn = func(_ context.Context, userUUID string, _ *domain.Msg) error {
		done <- userUUID
		return errors.New("cache set error")
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

func Test_ChatUseCase_Send_CacheSetError_System(t *testing.T) {
	f := newChatFixture()
	done := make(chan string, 2)
	f.msgCache.setFn = func(_ context.Context, userUUID string, _ *domain.Msg) error {
		done <- userUUID
		return errors.New("cache set error")
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

	res, ok, err := f.useCase.GetAllNew(ctx, 1)
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
	origTickerDuration := getAllNewTickerDuration
	getAllNewTickerDuration = time.Millisecond
	t.Cleanup(func() {
		getAllNewTickerDuration = origTickerDuration
	})
	want := []*domain.Msg{{ID: 1}}
	f.msgCache.getFn = func(_ context.Context, userUUID string) ([]*domain.Msg, bool, error) {
		if userUUID != "42" {
			t.Fatalf("unexpected cache key: got=%q want=%q", userUUID, "42")
		}
		return want, true, nil
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), 42)
	if err != nil {
		t.Fatalf("GetAllNew returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if len(res) != 1 || res[0].ID != want[0].ID {
		t.Fatalf("unexpected result: got=%v want=%v", res, want)
	}
}

func Test_ChatUseCase_GetAllNew_CacheErrorUnreadError(t *testing.T) {
	f := newChatFixture()
	origTickerDuration := getAllNewTickerDuration
	getAllNewTickerDuration = time.Millisecond
	t.Cleanup(func() {
		getAllNewTickerDuration = origTickerDuration
	})
	wantErr := errors.New("unread error")
	f.msgCache.getFn = func(_ context.Context, _ string) ([]*domain.Msg, bool, error) {
		return nil, false, errors.New("cache error")
	}
	f.msgRepo.getUnreadFn = func(_ context.Context, userUUID int, tx pgx.Tx) ([]*domain.Msg, error) {
		if userUUID != 7 {
			t.Fatalf("unexpected user uuid: got=%d want=7", userUUID)
		}
		if tx != nil {
			t.Fatal("expected nil tx in GetUnread")
		}
		return nil, wantErr
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), 7)
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
	origTickerDuration := getAllNewTickerDuration
	getAllNewTickerDuration = time.Millisecond
	t.Cleanup(func() {
		getAllNewTickerDuration = origTickerDuration
	})
	want := []*domain.Msg{{ID: 5}}
	f.msgCache.getFn = func(_ context.Context, _ string) ([]*domain.Msg, bool, error) {
		return nil, false, errors.New("cache error")
	}
	f.msgRepo.getUnreadFn = func(_ context.Context, _ int, _ pgx.Tx) ([]*domain.Msg, error) {
		return want, nil
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), 7)
	if err != nil {
		t.Fatalf("GetAllNew returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected ok to be true")
	}
	if len(res) != 1 || res[0].ID != want[0].ID {
		t.Fatalf("unexpected result: got=%v want=%v", res, want)
	}
}

func Test_ChatUseCase_GetAllNew_CacheMissTimeout(t *testing.T) {
	f := newChatFixture()
	origTickerDuration := getAllNewTickerDuration
	getAllNewTickerDuration = time.Millisecond
	t.Cleanup(func() {
		getAllNewTickerDuration = origTickerDuration
	})
	f.useCase.longPullReadTimeOut = 5 * time.Millisecond
	cacheReads := 0
	f.msgCache.getFn = func(_ context.Context, _ string) ([]*domain.Msg, bool, error) {
		cacheReads++
		return nil, false, nil
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), 9)
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

func Test_ChatUseCase_GetHistory(t *testing.T) {
	f := newChatFixture()
	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.April, 2, 0, 0, 0, 0, time.UTC)

	history, err := f.useCase.GetHistory(context.Background(), "user-1", 99, from, to)
	if err != nil {
		t.Fatalf("GetHistory returned error: %v", err)
	}
	if len(history) != 1 || history[0].ID != f.history[0].ID {
		t.Fatalf("unexpected history: got=%v want=%v", history, f.history)
	}
}
