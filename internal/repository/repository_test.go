package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repoTxStub struct{}

func (r *repoTxStub) Begin(context.Context) (pgx.Tx, error) { return nil, nil }
func (r *repoTxStub) Commit(context.Context) error          { return nil }
func (r *repoTxStub) Rollback(context.Context) error        { return nil }
func (r *repoTxStub) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (r *repoTxStub) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (r *repoTxStub) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }
func (r *repoTxStub) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (r *repoTxStub) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (r *repoTxStub) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (r *repoTxStub) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (r *repoTxStub) Conn() *pgx.Conn                                         { return nil }

type repoDBMock struct {
	beginTxFn  func(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error)
	execFn     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFn    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	queryRowFn func(ctx context.Context, sql string, args ...any) pgx.Row
}

func (r *repoDBMock) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	return r.beginTxFn(ctx, options)
}

func (r *repoDBMock) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return r.execFn(ctx, sql, args...)
}

func (r *repoDBMock) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return r.queryFn(ctx, sql, args...)
}

func (r *repoDBMock) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return r.queryRowFn(ctx, sql, args...)
}

func Test_DefaultPingPool(t *testing.T) {
	wantErr := errors.New("exec error")
	db := &repoDBMock{
		beginTxFn: func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) { return nil, nil },
		execFn: func(_ context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
			if sql != "SELECT 1;" {
				t.Fatalf("unexpected sql: got=%q want=%q", sql, "SELECT 1;")
			}
			if len(args) != 0 {
				t.Fatalf("unexpected args: got=%d want=0", len(args))
			}
			return pgconn.CommandTag{}, wantErr
		},
		queryFn:    func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) { return nil, nil },
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row { return nil },
	}

	err := defaultPingPool(context.Background(), db)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected exec error, got=%v", err)
	}
}

func Test_NewRepository_ParseConfigError(t *testing.T) {
	origParsePoolConfig := parsePoolConfig
	origNewPoolWithConfig := newPoolWithConfig
	origPingPool := pingPool
	t.Cleanup(func() {
		parsePoolConfig = origParsePoolConfig
		newPoolWithConfig = origNewPoolWithConfig
		pingPool = origPingPool
	})

	wantErr := errors.New("parse config error")
	parsePoolConfig = func(_ string) (*pgxpool.Config, error) {
		return nil, wantErr
	}

	repo, err := NewRepository(context.Background(), "dsn")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected parse config error, got=%v", err)
	}
	if repo != nil {
		t.Fatalf("expected nil repo on error, got=%v", repo)
	}
}

func Test_NewRepository_NewWithConfigError(t *testing.T) {
	origParsePoolConfig := parsePoolConfig
	origNewPoolWithConfig := newPoolWithConfig
	origPingPool := pingPool
	t.Cleanup(func() {
		parsePoolConfig = origParsePoolConfig
		newPoolWithConfig = origNewPoolWithConfig
		pingPool = origPingPool
	})

	wantErr := errors.New("new pool error")
	parsePoolConfig = func(_ string) (*pgxpool.Config, error) {
		return &pgxpool.Config{}, nil
	}
	newPoolWithConfig = func(_ context.Context, _ *pgxpool.Config) (*pgxpool.Pool, error) {
		return nil, wantErr
	}

	repo, err := NewRepository(context.Background(), "dsn")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected new pool error, got=%v", err)
	}
	if repo != nil {
		t.Fatalf("expected nil repo on error, got=%v", repo)
	}
}

func Test_NewRepository_PingError(t *testing.T) {
	origParsePoolConfig := parsePoolConfig
	origNewPoolWithConfig := newPoolWithConfig
	origPingPool := pingPool
	t.Cleanup(func() {
		parsePoolConfig = origParsePoolConfig
		newPoolWithConfig = origNewPoolWithConfig
		pingPool = origPingPool
	})

	pool := &pgxpool.Pool{}
	wantErr := errors.New("ping error")
	parsePoolConfig = func(_ string) (*pgxpool.Config, error) {
		return &pgxpool.Config{}, nil
	}
	newPoolWithConfig = func(_ context.Context, _ *pgxpool.Config) (*pgxpool.Pool, error) {
		return pool, nil
	}
	pingPool = func(_ context.Context, gotPool poolPinger) error {
		if gotPool != pool {
			t.Fatal("unexpected pool passed to ping")
		}
		return wantErr
	}

	repo, err := NewRepository(context.Background(), "dsn")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected ping error, got=%v", err)
	}
	if repo != nil {
		t.Fatalf("expected nil repo on error, got=%v", repo)
	}
}

func Test_NewRepository_OK(t *testing.T) {
	origParsePoolConfig := parsePoolConfig
	origNewPoolWithConfig := newPoolWithConfig
	origPingPool := pingPool
	t.Cleanup(func() {
		parsePoolConfig = origParsePoolConfig
		newPoolWithConfig = origNewPoolWithConfig
		pingPool = origPingPool
	})

	pool := &pgxpool.Pool{}
	parsePoolConfig = func(_ string) (*pgxpool.Config, error) {
		return &pgxpool.Config{}, nil
	}
	newPoolWithConfig = func(_ context.Context, _ *pgxpool.Config) (*pgxpool.Pool, error) {
		return pool, nil
	}
	pingPool = func(_ context.Context, gotPool poolPinger) error {
		if gotPool != pool {
			t.Fatal("unexpected pool passed to ping")
		}
		return nil
	}

	repo, err := NewRepository(context.Background(), "dsn")
	if err != nil {
		t.Fatalf("NewRepository returned error: %v", err)
	}
	if repo == nil {
		t.Fatal("NewRepository returned nil repo")
	}
	if repo.DB != pool {
		t.Fatal("unexpected DB in repository")
	}
	if repo.db != pool {
		t.Fatal("unexpected db execer in repository")
	}
	if repo.beginTx == nil {
		t.Fatal("expected beginTx to be initialized")
	}
}

func Test_Repository_CreateSession(t *testing.T) {
	wantTx := &repoTxStub{}
	repo := &Repository{
		beginTx: func(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
			if ctx == nil {
				t.Fatal("expected non-nil context")
			}
			if options.IsoLevel != pgx.Serializable {
				t.Fatalf("unexpected isolation level: got=%v want=%v", options.IsoLevel, pgx.Serializable)
			}
			return wantTx, nil
		},
	}

	tx, err := repo.CreateSession(context.Background(), pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	if tx != wantTx {
		t.Fatalf("unexpected tx: got=%v want=%v", tx, wantTx)
	}
}

func Test_Repository_GetDB_WithoutTx(t *testing.T) {
	db := &repoDBMock{
		beginTxFn: func(_ context.Context, _ pgx.TxOptions) (pgx.Tx, error) { return nil, nil },
		execFn: func(_ context.Context, _ string, _ ...any) (pgconn.CommandTag, error) {
			return pgconn.CommandTag{}, nil
		},
		queryFn:    func(_ context.Context, _ string, _ ...any) (pgx.Rows, error) { return nil, nil },
		queryRowFn: func(_ context.Context, _ string, _ ...any) pgx.Row { return nil },
	}
	repo := &Repository{db: db}

	got := repo.GetDB(nil)
	if got != db {
		t.Fatalf("unexpected db returned: got=%v want=%v", got, db)
	}
}

func Test_Repository_GetDB_WithTx(t *testing.T) {
	tx := &repoTxStub{}
	repo := &Repository{}

	got := repo.GetDB(tx)
	if got != tx {
		t.Fatalf("unexpected db returned: got=%v want=%v", got, tx)
	}
}
