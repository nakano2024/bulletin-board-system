package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	domainthread "github.com/nakanokota/bulletin-board-system/backend/domain/thread"
)

// querier is satisfied by both *pgxpool.Pool and pgx.Tx, so tests can run
// FetchActiveThreadListNewestFirst/CreateThread inside a transaction for isolation.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresThreadRepository struct {
	db                  querier
	baseThreadImagePath string
}

// NewPostgresThreadRepository takes baseThreadImagePath (the deployment-level storage prefix for thread images, e.g. from an env var)
// used to reconstruct each Thread's FilePath from the DB's file_name column.
func NewPostgresThreadRepository(db querier, baseThreadImagePath string) *PostgresThreadRepository {
	return &PostgresThreadRepository{db: db, baseThreadImagePath: baseThreadImagePath}
}

func (r *PostgresThreadRepository) FetchActiveThreadListNewestFirst(ctx context.Context) (*domainthread.ThreadList, error) {
	rows, err := r.db.Query(ctx, "SELECT id, body, file_name, created_at FROM threads WHERE is_alive = true ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	threads := make([]*domainthread.Thread, 0)
	for rows.Next() {
		var id, body string
		var fileName *string
		var createdAt time.Time
		if err := rows.Scan(&id, &body, &fileName, &createdAt); err != nil {
			return nil, err
		}

		filePath, err := r.filePathFromFileName(fileName)
		if err != nil {
			return nil, err
		}

		th, err := domainthread.NewThread(id, body, filePath, createdAt)
		if err != nil {
			return nil, err
		}

		threads = append(threads, th)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return domainthread.NewThreadList(threads), nil
}

func (r *PostgresThreadRepository) CreateThread(ctx context.Context, pendingThread *domainthread.PendingThread) (*domainthread.Thread, error) {
	var id string
	var createdAt time.Time

	err := r.db.QueryRow(
		ctx,
		"INSERT INTO threads (user_id, body, file_name) VALUES ($1, $2, $3) RETURNING id, created_at",
		pendingThread.UserID(), pendingThread.Body(), pendingThread.FileName().Value(),
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	filePath, err := domainthread.NewFilePath(r.baseThreadImagePath, pendingThread.FileName().Value())
	if err != nil {
		return nil, err
	}

	return domainthread.NewThread(id, pendingThread.Body(), filePath, createdAt)
}

func (r *PostgresThreadRepository) filePathFromFileName(fileName *string) (*domainthread.FilePath, error) {
	if fileName == nil {
		return nil, nil
	}
	return domainthread.NewFilePath(r.baseThreadImagePath, *fileName)
}
