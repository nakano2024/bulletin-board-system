package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

// In the RunInTx tests the manager is given the test transaction, so RunInTx opens a savepoint inside it;
// the test transaction is rolled back at cleanup either way.

func countUsersByIP(t *testing.T, tx pgx.Tx, ip string) int {
	t.Helper()
	var n int
	require.NoError(t, tx.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE ip = $1", ip).Scan(&n))
	return n
}

func TestPostgresTransactionManager_RunInTx_正常系_コミット(t *testing.T) {
	pool := newTestPool(t)
	date := mustUserCreateDate(t, time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name          string
		ip            string
		wantUserCount int
	}{
		{
			name:          "fnがnilを返したとき、fn内でCreateUserが書き込んだusersの行が、RunInTxの後も残っていること",
			ip:            "203.0.113.10",
			wantUserCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)
			userRepo := postgres.NewPostgresUserRepository(pool)

			sut := postgres.NewPostgresTransactionManager(testTx)
			err := sut.RunInTx(ctx, func(txCtx context.Context) error {
				_, err := userRepo.CreateUser(txCtx, mustNewPendingUser(t, tt.ip, date))
				return err
			})
			require.NoError(t, err)

			assert.Equal(t, tt.wantUserCount, countUsersByIP(t, testTx, tt.ip))
		})
	}
}

func TestPostgresTransactionManager_RunInTx_異常系_ロールバック(t *testing.T) {
	pool := newTestPool(t)
	date := mustUserCreateDate(t, time.Date(2026, 9, 27, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name          string
		ip            string
		fnErr         error
		wantUserCount int
	}{
		{
			name:          "fnがエラーを返したとき、fn内でCreateUserが書き込んだusersの行が、RunInTxの後に残っていないこと",
			ip:            "203.0.113.11",
			fnErr:         errors.New("fn failed"),
			wantUserCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)
			userRepo := postgres.NewPostgresUserRepository(pool)

			sut := postgres.NewPostgresTransactionManager(testTx)
			_ = sut.RunInTx(ctx, func(txCtx context.Context) error {
				_, err := userRepo.CreateUser(txCtx, mustNewPendingUser(t, tt.ip, date))
				require.NoError(t, err)
				return tt.fnErr
			})

			assert.Equal(t, tt.wantUserCount, countUsersByIP(t, testTx, tt.ip))
		})
	}
}

func TestPostgresTransactionManager_RunInTx_正常系_戻り値(t *testing.T) {
	pool := newTestPool(t)

	tests := []struct {
		name  string
		fnErr error
	}{
		{
			name:  "fnがnilを返したとき、RunInTxがnilを返すこと",
			fnErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTx := beginTx(t, pool)

			sut := postgres.NewPostgresTransactionManager(testTx)
			err := sut.RunInTx(context.Background(), func(context.Context) error { return tt.fnErr })

			require.NoError(t, err)
		})
	}
}

func TestPostgresTransactionManager_RunInTx_異常系_戻り値(t *testing.T) {
	pool := newTestPool(t)
	errFn := errors.New("fn failed")

	tests := []struct {
		name    string
		fnErr   error
		wantErr error
	}{
		{
			name:    "fnがエラーを返したとき、RunInTxがそのエラーを返すこと",
			fnErr:   errFn,
			wantErr: errFn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testTx := beginTx(t, pool)

			sut := postgres.NewPostgresTransactionManager(testTx)
			err := sut.RunInTx(context.Background(), func(context.Context) error { return tt.fnErr })

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// The tests below build each repository on the pool but pass a ctx carrying the test transaction. Data seeded only
// inside that (uncommitted) transaction is invisible to the pool, so reading it — or writing so that the pool cannot
// see the row — shows the repository used the transaction from ctx.

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
			name:   "ctxのトランザクション内でだけ投入したusersの行が、Userとして返ること",
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
			got, err := sut.FindByIPAndDate(postgres.WithTx(ctx, testTx), tt.ip, date)

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
			name:              "書き込んだusersの行が、ctxのトランザクションの外(プール)からは見えないこと",
			ip:                "203.0.113.21",
			wantCountFromPool: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)

			sut := postgres.NewPostgresUserRepository(pool)
			_, err := sut.CreateUser(postgres.WithTx(ctx, testTx), mustNewPendingUser(t, tt.ip, date))
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
			name:      "ctxのトランザクション内でだけ投入したfilesのfile_idを渡したとき、trueが返ること",
			seedFiles: []seedFile{{id: "file-tx", name: "sample.png"}},
			fileID:    "file-tx",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)
			seedFilesAndThreads(t, testTx, tt.seedFiles, nil)

			sut := postgres.NewPostgresThreadFileChecker(pool)
			got, err := sut.ExistsFile(postgres.WithTx(ctx, testTx), tt.fileID)

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
			name:        "ctxのトランザクション内でだけ投入したthreadsが参照するfile_idを渡したとき、trueが返ること",
			seedFiles:   []seedFile{{id: "file-tx", name: "sample.png"}},
			seedThreads: []seedThread{{id: "thread-tx", body: "hello", fileID: "file-tx", isAlive: true, createdAt: fixedTime}},
			fileID:      "file-tx",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)
			seedFilesAndThreads(t, testTx, tt.seedFiles, tt.seedThreads)

			sut := postgres.NewPostgresThreadFileChecker(pool)
			got, err := sut.IsFileAttachedToThread(postgres.WithTx(ctx, testTx), tt.fileID)

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
			name:          "ctxのトランザクション内でだけ投入したfilesを参照するPendingThreadを渡したとき、エラーなくThreadが作成されること",
			seedFiles:     []seedFile{{id: "file-tx", name: "sample.png"}},
			pendingThread: mustNewPendingThread(t, "user-1", "hello", "file-tx"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)
			seedFilesAndThreads(t, testTx, tt.seedFiles, nil)

			sut := postgres.NewPostgresThreadRepository(pool, "thread_images/")
			_, err := sut.CreateThread(postgres.WithTx(ctx, testTx), tt.pendingThread)

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
			name:        "ctxのトランザクション内でだけ投入したスレッドが、ThreadListに含まれること",
			seedFiles:   []seedFile{{id: "file-tx", name: "sample.png"}},
			seedThreads: []seedThread{{id: "thread-tx", body: "hello", fileID: "file-tx", isAlive: true, createdAt: fixedTime}},
			want:        []*domainthread.Thread{threadInTx},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			testTx := beginTx(t, pool)
			seedFilesAndThreads(t, testTx, tt.seedFiles, tt.seedThreads)

			sut := postgres.NewPostgresThreadRepository(pool, "thread_images/")
			got, err := sut.FetchActiveThreadListNewestFirst(postgres.WithTx(ctx, testTx))

			require.NoError(t, err)
			assert.Equal(t, normalizeThreads(tt.want), normalizeThreads(got.Threads()))
		})
	}
}
