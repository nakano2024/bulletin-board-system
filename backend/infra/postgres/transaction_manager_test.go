package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

// savepointBeginner lets PostgresTransactionManager begin on the test transaction: pgx.Tx has no BeginTx, so BeginTx
// opens a savepoint via Begin. The test transaction is rolled back at cleanup either way.
type savepointBeginner struct{ pgx.Tx }

func (b savepointBeginner) BeginTx(ctx context.Context, _ pgx.TxOptions) (pgx.Tx, error) {
	return b.Begin(ctx)
}

// runInTestTx runs fn through RunInTx on testTx, so the ctx fn receives carries a transaction nested in testTx.
func runInTestTx(t *testing.T, testTx pgx.Tx, fn func(txCtx context.Context)) {
	t.Helper()
	sut := postgres.NewPostgresTransactionManager(savepointBeginner{testTx})
	require.NoError(t, sut.RunInTx(context.Background(), func(txCtx context.Context) error {
		fn(txCtx)
		return nil
	}))
}

// The tests below build each repository on the pool but call it with the ctx that RunInTx passes to fn. Data seeded
// only inside the (uncommitted) test transaction is invisible to the pool, so reading it — or writing so that the pool
// cannot see the row — shows the repository used the transaction from ctx.

func TestPostgresUserRepository_FindByIPAndDate_ctxのトランザクション(t *testing.T) {
	pool := newTestPool(t)
	date := mustUserCreateDate(t, time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name   string
		seedID string
		ip     string
		want   *domainuser.User
	}{
		{
			name:   "RunInTxのfnに渡されたctxで呼んだとき、そのトランザクション内でだけ投入したusersの行が、Userとして返ること",
			seedID: "user-tx",
			ip:     "203.0.113.20",
			want:   mustNewUser(t, "user-tx", "203.0.113.20"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)
			_, err := testTx.Exec(ctx, "INSERT INTO users (id, ip, create_date) VALUES ($1, $2, $3)", tt.seedID, tt.ip, date.Time())
			require.NoError(t, err)

			sut := postgres.NewPostgresUserRepository(pool)
			var got *domainuser.User
			runInTestTx(t, testTx, func(txCtx context.Context) {
				got, err = sut.FindByIPAndDate(txCtx, tt.ip, date)
			})

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgresUserRepository_CreateUser_ctxのトランザクション(t *testing.T) {
	pool := newTestPool(t)
	date := mustUserCreateDate(t, time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name              string
		ip                string
		wantCountFromPool int
	}{
		{
			name:              "RunInTxのfnに渡されたctxで呼んだとき、書き込んだusersの行が、そのトランザクションの外(プール)からは見えないこと",
			ip:                "203.0.113.21",
			wantCountFromPool: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)

			sut := postgres.NewPostgresUserRepository(pool)
			var err error
			runInTestTx(t, testTx, func(txCtx context.Context) {
				_, err = sut.CreateUser(txCtx, mustNewPendingUser(t, tt.ip, date))
			})
			require.NoError(t, err)

			var countFromPool int
			require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM users WHERE ip = $1", tt.ip).Scan(&countFromPool))
			assert.Equal(t, tt.wantCountFromPool, countFromPool)
		})
	}
}

func TestPostgresThreadFileChecker_ExistsFile_ctxのトランザクション(t *testing.T) {
	pool := newTestPool(t)

	tests := []struct {
		name      string
		seedFiles []seedFile
		fileID    string
		want      bool
	}{
		{
			name:      "RunInTxのfnに渡されたctxで、そのトランザクション内でだけ投入したfilesのfile_idを渡したとき、trueが返ること",
			seedFiles: []seedFile{{id: "file-tx", name: "sample.png"}},
			fileID:    "file-tx",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTx := beginTx(t, pool)
			seedFilesAndThreads(t, testTx, tt.seedFiles, nil)

			sut := postgres.NewPostgresThreadFileChecker(pool)
			var got bool
			var err error
			runInTestTx(t, testTx, func(txCtx context.Context) {
				got, err = sut.ExistsFile(txCtx, tt.fileID)
			})

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgresThreadFileChecker_IsFileAttachedToThread_ctxのトランザクション(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		seedFiles   []seedFile
		seedThreads []seedThread
		fileID      string
		want        bool
	}{
		{
			name:        "RunInTxのfnに渡されたctxで、そのトランザクション内でだけ投入したthreadsが参照するfile_idを渡したとき、trueが返ること",
			seedFiles:   []seedFile{{id: "file-tx", name: "sample.png"}},
			seedThreads: []seedThread{{id: "thread-tx", body: "hello", fileID: "file-tx", isAlive: true, createdAt: fixedTime}},
			fileID:      "file-tx",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTx := beginTx(t, pool)
			seedFilesAndThreads(t, testTx, tt.seedFiles, tt.seedThreads)

			sut := postgres.NewPostgresThreadFileChecker(pool)
			var got bool
			var err error
			runInTestTx(t, testTx, func(txCtx context.Context) {
				got, err = sut.IsFileAttachedToThread(txCtx, tt.fileID)
			})

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgresThreadRepository_CreateThread_ctxのトランザクション(t *testing.T) {
	pool := newTestPool(t)

	tests := []struct {
		name          string
		seedFiles     []seedFile
		pendingThread *domainthread.PendingThread
	}{
		{
			name:          "RunInTxのfnに渡されたctxで、そのトランザクション内でだけ投入したfilesを参照するPendingThreadを渡したとき、エラーなくThreadが作成されること",
			seedFiles:     []seedFile{{id: "file-tx", name: "sample.png"}},
			pendingThread: mustNewPendingThread(t, "user-1", "hello", "file-tx"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTx := beginTx(t, pool)
			seedFilesAndThreads(t, testTx, tt.seedFiles, nil)

			sut := postgres.NewPostgresThreadRepository(pool, "thread_images/")
			var err error
			runInTestTx(t, testTx, func(txCtx context.Context) {
				_, err = sut.CreateThread(txCtx, tt.pendingThread)
			})

			require.NoError(t, err)
		})
	}
}

func TestPostgresThreadRepository_FetchActiveThreadListNewestFirst_ctxのトランザクション(t *testing.T) {
	pool := newTestPool(t)
	fixedTime := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)
	threadInTx, _ := domainthread.NewThread("thread-tx", "hello", mustFilePath(t, "sample.png"), fixedTime)

	tests := []struct {
		name        string
		seedFiles   []seedFile
		seedThreads []seedThread
		want        []*domainthread.Thread
	}{
		{
			name:        "RunInTxのfnに渡されたctxで呼んだとき、そのトランザクション内でだけ投入したスレッドが、ThreadListに含まれること",
			seedFiles:   []seedFile{{id: "file-tx", name: "sample.png"}},
			seedThreads: []seedThread{{id: "thread-tx", body: "hello", fileID: "file-tx", isAlive: true, createdAt: fixedTime}},
			want:        []*domainthread.Thread{threadInTx},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTx := beginTx(t, pool)
			seedFilesAndThreads(t, testTx, tt.seedFiles, tt.seedThreads)

			sut := postgres.NewPostgresThreadRepository(pool, "thread_images/")
			var got *domainthread.ThreadList
			var err error
			runInTestTx(t, testTx, func(txCtx context.Context) {
				got, err = sut.FetchActiveThreadListNewestFirst(txCtx)
			})

			require.NoError(t, err)
			assert.Equal(t, normalizeThreads(tt.want), normalizeThreads(got.Threads()))
		})
	}
}
