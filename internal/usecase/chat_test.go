package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"shc/domain"

	"github.com/jackc/pgx/v5"
)

type chatMsgCacheMock struct {
	setMsgFn        func(ctx context.Context, msg *domain.Msg) error
	getByUserUUIDFn func(ctx context.Context, userUUID string) ([]domain.Msg, bool, error)
	getByTicketIDFn func(ctx context.Context, ticketID int) ([]domain.Msg, bool, error)
	deleteByMsgIDFn func(ctx context.Context, userUUID string, msgID int) error
}

func (m *chatMsgCacheMock) SetMsg(ctx context.Context, msg *domain.Msg) error {
	return m.setMsgFn(ctx, msg)
}

func (m *chatMsgCacheMock) GetMsgByUserUUID(ctx context.Context, userUUID string) ([]domain.Msg, bool, error) {
	return m.getByUserUUIDFn(ctx, userUUID)
}

func (m *chatMsgCacheMock) GetMsgByTiketID(ctx context.Context, ticketID int) ([]domain.Msg, bool, error) {
	return m.getByTicketIDFn(ctx, ticketID)
}

func (m *chatMsgCacheMock) DeleteByMsgID(ctx context.Context, userUUID string, msgID int) error {
	return m.deleteByMsgIDFn(ctx, userUUID, msgID)
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
	createFn       func(ctx context.Context, msg domain.CreateMsg, userUUID string, tx pgx.Tx) (*domain.Msg, error)
	createFileFn   func(ctx context.Context, msgID int, fileName, path string, tx pgx.Tx) (*domain.MsgFileContent, error)
	getUnreadFn    func(ctx context.Context, userUUID string, tx pgx.Tx) ([]domain.Msg, error)
	markReadByIDFn func(ctx context.Context, userUUID string, id int, tx pgx.Tx) error
	getHistoryFn   func(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]domain.Msg, error)
}

func (m *chatMsgRepoMock) Create(ctx context.Context, msg domain.CreateMsg, userUUID string, tx pgx.Tx) (*domain.Msg, error) {
	return m.createFn(ctx, msg, userUUID, tx)
}

func (m *chatMsgRepoMock) CreateFile(ctx context.Context, msgID int, fileName, path string, tx pgx.Tx) (*domain.MsgFileContent, error) {
	return m.createFileFn(ctx, msgID, fileName, path, tx)
}

func (m *chatMsgRepoMock) GetUnread(ctx context.Context, userUUID string, tx pgx.Tx) ([]domain.Msg, error) {
	return m.getUnreadFn(ctx, userUUID, tx)
}

func (m *chatMsgRepoMock) MarkReadByID(ctx context.Context, userUUID string, id int, tx pgx.Tx) error {
	return m.markReadByIDFn(ctx, userUUID, id, tx)
}

func (m *chatMsgRepoMock) GetHistory(ctx context.Context, userUUID string, ticketUUID int, from, to time.Time) ([]domain.Msg, error) {
	return m.getHistoryFn(ctx, userUUID, ticketUUID, from, to)
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
	encryption *chatEncryptionMock
	user       domain.UserSystemInfo
	msg        domain.CreateMsg
}

func newChatFixture() *chatFixture {
	tx := &fakeTx{}
	user := domain.UserSystemInfo{UUID: "user-uuid-1", UserRole: domain.UserRoleManager}
	msg := domain.CreateMsg{TicketID: 99, Text: "hello"}

	mainRepo := &mainRepoMock{
		createSessionFn: func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
			return tx, nil
		},
	}
	msgCache := &chatMsgCacheMock{
		setMsgFn: func(_ context.Context, _ *domain.Msg) error {
			return nil
		},
		getByUserUUIDFn: func(_ context.Context, _ string) ([]domain.Msg, bool, error) {
			return nil, false, nil
		},
		getByTicketIDFn: func(_ context.Context, _ int) ([]domain.Msg, bool, error) {
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
		createFn: func(_ context.Context, createMsg domain.CreateMsg, userUUID string, gotTx pgx.Tx) (*domain.Msg, error) {
			if gotTx != tx {
				return nil, errors.New("unexpected tx")
			}
			if userUUID != user.UUID {
				return nil, errors.New("unexpected user uuid")
			}
			if createMsg.Text != "" {
				return nil, errors.New("expected plaintext to be cleared")
			}
			if !bytes.Equal(createMsg.EncryptedText, []byte("enc:hello")) {
				return nil, errors.New("unexpected encrypted text")
			}
			return &domain.Msg{
				ID:            10,
				UserUUID:      userUUID,
				TicketID:      createMsg.TicketID,
				EncryptedText: createMsg.EncryptedText,
				Status:        domain.MsgStatusSent,
			}, nil
		},
		createFileFn: func(_ context.Context, msgID int, fileName, path string, gotTx pgx.Tx) (*domain.MsgFileContent, error) {
			if gotTx != tx {
				return nil, errors.New("unexpected tx")
			}
			return &domain.MsgFileContent{
				ID:       msgID + len(fileName),
				MsgID:    msgID,
				FileName: fileName,
				Path:     path,
			}, nil
		},
		getUnreadFn: func(_ context.Context, _ string, _ pgx.Tx) ([]domain.Msg, error) {
			return nil, nil
		},
		markReadByIDFn: func(_ context.Context, _ string, _ int, _ pgx.Tx) error {
			return nil
		},
		getHistoryFn: func(_ context.Context, _ string, _ int, _, _ time.Time) ([]domain.Msg, error) {
			return []domain.Msg{{ID: 1, TicketID: msg.TicketID, EncryptedText: []byte("enc:history")}}, nil
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

	return &chatFixture{
		useCase:    NewChatUseCase(msgCache, s3, msgRepo, mainRepo, encryption, 20*time.Millisecond, time.Millisecond),
		tx:         tx,
		mainRepo:   mainRepo,
		msgCache:   msgCache,
		s3:         s3,
		msgRepo:    msgRepo,
		encryption: encryption,
		user:       user,
		msg:        msg,
	}
}

func waitChatValue[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case val := <-ch:
		return val
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for async call")
	}
	var zero T
	return zero
}

func Test_NewChatUseCase(t *testing.T) {
	f := newChatFixture()
	if f.useCase.msgCache != f.msgCache {
		t.Fatal("unexpected msgCache")
	}
	if f.useCase.s3 != f.s3 {
		t.Fatal("unexpected s3")
	}
	if f.useCase.msgRepo != f.msgRepo {
		t.Fatal("unexpected msgRepo")
	}
	if f.useCase.mainRepo != f.mainRepo {
		t.Fatal("unexpected mainRepo")
	}
	if f.useCase.encryption != f.encryption {
		t.Fatal("unexpected encryption")
	}
}

func Test_ChatUseCase_Send_OK_WithFiles(t *testing.T) {
	f := newChatFixture()
	f.msg.Files = []domain.CreateFile{{Name: "a.txt", Data: []byte("a")}}
	cached := make(chan *domain.Msg, 1)
	f.msgCache.setMsgFn = func(_ context.Context, msg *domain.Msg) error {
		cached <- msg
		return nil
	}

	err := f.useCase.Send(context.Background(), f.user, f.msg)
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit once, got=%d", f.tx.commitCalls)
	}

	msg := waitChatValue(t, cached)
	if msg.ID != 10 || len(msg.Files) != 1 {
		t.Fatalf("unexpected cached msg: %+v", msg)
	}
	if msg.Files[0].FileName != "a.txt" || msg.Files[0].Path != "ticket/file/99/a" {
		t.Fatalf("unexpected file content: %+v", msg.Files[0])
	}
}

func Test_ChatUseCase_Send_EncryptError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("encrypt error")
	f.encryption.encryptFn = func(_ []byte, _ []byte) ([]byte, error) {
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.user, f.msg)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected encrypt error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_CreateSessionError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("session error")
	f.mainRepo.createSessionFn = func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.user, f.msg)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected session error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_FilesLimitError(t *testing.T) {
	f := newChatFixture()
	f.msg.Files = make([]domain.CreateFile, 11)

	err := f.useCase.Send(context.Background(), f.user, f.msg)
	if !errors.Is(err, domain.ErrFilesLoad) {
		t.Fatalf("expected files limit error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_FileTooLargeError(t *testing.T) {
	f := newChatFixture()
	files := []domain.CreateFile{{Name: "big.bin", Data: make([]byte, 15*1024*1024+1)}}

	_, paths, err := f.useCase.fileToS3(context.Background(), files, 99)
	if !errors.Is(err, domain.ErrFilesLoad) {
		t.Fatalf("expected files load error, got=%v", err)
	}
	if len(paths) != 0 {
		t.Fatalf("expected no paths, got=%v", paths)
	}
}

func Test_ChatUseCase_Send_S3Error_CleansSavedFiles(t *testing.T) {
	f := newChatFixture()
	f.msg.Files = []domain.CreateFile{
		{Name: "a.txt", Data: []byte("a")},
		{Name: "b.txt", Data: []byte("b")},
	}
	wantErr := errors.New("save error")
	f.s3.saveFn = func(_ context.Context, folder string, file []byte) (string, error) {
		if string(file) == "b" {
			return "", wantErr
		}
		return folder + "/" + string(file), nil
	}
	deleted := make(chan string, 1)
	f.s3.deleteFn = func(_ context.Context, path string) error {
		deleted <- path
		return nil
	}

	err := f.useCase.Send(context.Background(), f.user, f.msg)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected save error, got=%v", err)
	}
	if got := waitChatValue(t, deleted); got != "ticket/file/99/a" {
		t.Fatalf("unexpected deleted path: %q", got)
	}
}

func Test_ChatUseCase_Send_CreateMsgError_CleansFiles(t *testing.T) {
	f := newChatFixture()
	f.msg.Files = []domain.CreateFile{{Name: "a.txt", Data: []byte("a")}}
	wantErr := errors.New("create message error")
	f.msgRepo.createFn = func(_ context.Context, _ domain.CreateMsg, _ string, _ pgx.Tx) (*domain.Msg, error) {
		return nil, wantErr
	}
	deleted := make(chan string, 1)
	f.s3.deleteFn = func(_ context.Context, path string) error {
		deleted <- path
		return nil
	}

	err := f.useCase.Send(context.Background(), f.user, f.msg)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create message error, got=%v", err)
	}
	if got := waitChatValue(t, deleted); got != "ticket/file/99/a" {
		t.Fatalf("unexpected deleted path: %q", got)
	}
}

func Test_ChatUseCase_Send_CreateFileError(t *testing.T) {
	f := newChatFixture()
	f.msg.Files = []domain.CreateFile{{Name: "a.txt", Data: []byte("a")}}
	wantErr := errors.New("create file error")
	f.msgRepo.createFileFn = func(_ context.Context, _ int, _ string, _ string, _ pgx.Tx) (*domain.MsgFileContent, error) {
		return nil, wantErr
	}

	err := f.useCase.Send(context.Background(), f.user, f.msg)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected create file error, got=%v", err)
	}
}

func Test_ChatUseCase_Send_CommitError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("commit error")
	f.tx.commitErr = wantErr

	err := f.useCase.Send(context.Background(), f.user, f.msg)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected commit error, got=%v", err)
	}
}

func Test_ChatUseCase_GetAllNew_ContextDone(t *testing.T) {
	f := newChatFixture()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, ok, err := f.useCase.GetAllNew(ctx, f.user)
	if err != nil {
		t.Fatalf("GetAllNew returned error: %v", err)
	}
	if ok || res != nil {
		t.Fatalf("expected empty timeout result, got res=%v ok=%v", res, ok)
	}
}

func Test_ChatUseCase_GetAllNew_CacheHit(t *testing.T) {
	f := newChatFixture()
	decrypted := false
	f.msgCache.getByUserUUIDFn = func(_ context.Context, userUUID string) ([]domain.Msg, bool, error) {
		if userUUID != f.user.UUID {
			t.Fatalf("unexpected user uuid: %q", userUUID)
		}
		return []domain.Msg{{ID: 1, TicketID: f.msg.TicketID, EncryptedText: []byte("enc:cache")}}, true, nil
	}
	f.encryption.decryptFn = func(data []byte, aad []byte) ([]byte, error) {
		decrypted = true
		if !bytes.Equal(data, []byte("enc:cache")) || !bytes.Equal(aad, []byte("99")) {
			t.Fatalf("unexpected decrypt args: data=%q aad=%q", data, aad)
		}
		return []byte("cache"), nil
	}

	res, ok, err := f.useCase.GetAllNew(context.Background(), f.user)
	if err != nil {
		t.Fatalf("GetAllNew returned error: %v", err)
	}
	if !ok || len(res) != 1 || res[0].Text != "cache" || !decrypted {
		t.Fatalf("unexpected result: res=%v ok=%v decrypted=%v", res, ok, decrypted)
	}
}

func Test_ChatUseCase_GetAllNew_CacheErrorUnreadError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("unread error")
	f.msgCache.getByUserUUIDFn = func(_ context.Context, _ string) ([]domain.Msg, bool, error) {
		return nil, false, errors.New("cache error")
	}
	f.msgRepo.getUnreadFn = func(_ context.Context, userUUID string, tx pgx.Tx) ([]domain.Msg, error) {
		if userUUID != f.user.UUID || tx != nil {
			t.Fatalf("unexpected unread args: userUUID=%q tx=%v", userUUID, tx)
		}
		return nil, wantErr
	}

	_, _, err := f.useCase.GetAllNew(context.Background(), f.user)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected unread error, got=%v", err)
	}
}

func Test_ChatUseCase_GetAllNew_DecryptError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("decrypt error")
	f.msgCache.getByUserUUIDFn = func(_ context.Context, _ string) ([]domain.Msg, bool, error) {
		return nil, false, errors.New("cache error")
	}
	f.msgRepo.getUnreadFn = func(_ context.Context, _ string, _ pgx.Tx) ([]domain.Msg, error) {
		return []domain.Msg{{ID: 1, TicketID: f.msg.TicketID, EncryptedText: []byte("bad")}}, nil
	}
	f.encryption.decryptFn = func(_ []byte, _ []byte) ([]byte, error) {
		return nil, wantErr
	}

	_, _, err := f.useCase.GetAllNew(context.Background(), f.user)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected decrypt error, got=%v", err)
	}
}

func Test_ChatUseCase_MarkAsRead_CreateSessionError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("session error")
	f.mainRepo.createSessionFn = func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
		return nil, wantErr
	}

	err := f.useCase.MarkAsRead(context.Background(), f.user, []int{1})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected session error, got=%v", err)
	}
}

func Test_ChatUseCase_MarkAsRead_RepoError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("mark read error")
	f.msgRepo.markReadByIDFn = func(_ context.Context, _ string, _ int, _ pgx.Tx) error {
		return wantErr
	}

	err := f.useCase.MarkAsRead(context.Background(), f.user, []int{1})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected mark read error, got=%v", err)
	}
}

func Test_ChatUseCase_MarkAsRead_CommitError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("commit error")
	f.tx.commitErr = wantErr

	err := f.useCase.MarkAsRead(context.Background(), f.user, []int{1})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected commit error, got=%v", err)
	}
}

func Test_ChatUseCase_MarkAsRead_OK(t *testing.T) {
	f := newChatFixture()
	deleted := make(chan int, 2)
	f.msgCache.deleteByMsgIDFn = func(_ context.Context, userUUID string, msgID int) error {
		if userUUID != f.user.UUID {
			t.Fatalf("unexpected user uuid: %q", userUUID)
		}
		deleted <- msgID
		return nil
	}

	err := f.useCase.MarkAsRead(context.Background(), f.user, []int{1, 2})
	if err != nil {
		t.Fatalf("MarkAsRead returned error: %v", err)
	}
	if f.tx.commitCalls != 1 {
		t.Fatalf("expected commit once, got=%d", f.tx.commitCalls)
	}
	got := map[int]bool{
		waitChatValue(t, deleted): true,
		waitChatValue(t, deleted): true,
	}
	if !got[1] || !got[2] {
		t.Fatalf("unexpected deleted ids: %v", got)
	}
}

func Test_ChatUseCase_GetHistory_DurationError(t *testing.T) {
	f := newChatFixture()
	from := time.Now()
	to := from.Add(oneMonthDuration + time.Second)

	res, err := f.useCase.GetHistory(context.Background(), f.user, f.msg.TicketID, from, to)
	if !errors.Is(err, domain.ErrDurationFromTo) {
		t.Fatalf("expected duration error, got=%v", err)
	}
	if res != nil {
		t.Fatalf("expected nil result, got=%v", res)
	}
}

func Test_ChatUseCase_GetHistory_RepoError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("history error")
	f.msgRepo.getHistoryFn = func(_ context.Context, _ string, _ int, _, _ time.Time) ([]domain.Msg, error) {
		return nil, wantErr
	}

	_, err := f.useCase.GetHistory(context.Background(), f.user, f.msg.TicketID, time.Now(), time.Now())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected history error, got=%v", err)
	}
}

func Test_ChatUseCase_GetHistory_DecryptError(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("decrypt error")
	f.encryption.decryptFn = func(_ []byte, _ []byte) ([]byte, error) {
		return nil, wantErr
	}

	_, err := f.useCase.GetHistory(context.Background(), f.user, f.msg.TicketID, time.Now(), time.Now())
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected decrypt error, got=%v", err)
	}
}

func Test_ChatUseCase_GetHistory_OK(t *testing.T) {
	f := newChatFixture()

	res, err := f.useCase.GetHistory(context.Background(), f.user, f.msg.TicketID, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("GetHistory returned error: %v", err)
	}
	if len(res) != 1 || res[0].Text != "history" {
		t.Fatalf("unexpected history: %+v", res)
	}
}

func Test_ChatUseCase_DeleteFromS3_Error(t *testing.T) {
	f := newChatFixture()
	wantErr := errors.New("delete error")
	f.s3.deleteFn = func(_ context.Context, _ string) error {
		return wantErr
	}

	err := f.useCase.deleteFormS3(context.Background(), []string{"path"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected delete error, got=%v", err)
	}
}

func Test_ChatUseCase_DeleteFromS3_OK(t *testing.T) {
	f := newChatFixture()

	err := f.useCase.deleteFormS3(context.Background(), []string{"path"})
	if err != nil {
		t.Fatalf("deleteFormS3 returned error: %v", err)
	}
}
