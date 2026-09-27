package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

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
// used to reconstruct each Thread's FilePath from files.name.
func NewPostgresThreadRepository(db querier, baseThreadImagePath string) *PostgresThreadRepository {
	return &PostgresThreadRepository{db: db, baseThreadImagePath: baseThreadImagePath}
}

func (r *PostgresThreadRepository) FetchActiveThreadListNewestFirst(ctx context.Context) (*domainthread.ThreadList, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.body, f.name, t.created_at
		FROM threads t
		JOIN files f ON f.id = t.file_id
		WHERE t.is_alive = true
		ORDER BY t.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	threads := make([]*domainthread.Thread, 0)
	for rows.Next() {
		var id, body, fileName string
		var createdAt time.Time
		if err := rows.Scan(&id, &body, &fileName, &createdAt); err != nil {
			return nil, err
		}

		filePath, err := domainthread.NewFilePath(r.baseThreadImagePath, fileName)
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

// CreateThread translates the threads.file_id FK / UNIQUE violations into ErrFileNotFound / ErrFileAlreadyUsed
// for requests that slip past ThreadCreationService's checks concurrently.
func (r *PostgresThreadRepository) CreateThread(ctx context.Context, pendingThread *domainthread.PendingThread) (*domainthread.Thread, error) {
	var id, fileName string
	var createdAt time.Time

	err := r.db.QueryRow(ctx, `
		WITH inserted AS (
			INSERT INTO threads (user_id, body, file_id) VALUES ($1, $2, $3)
			RETURNING id, file_id, created_at
		)
		SELECT i.id, f.name, i.created_at
		FROM inserted i
		JOIN files f ON f.id = i.file_id`,
		pendingThread.UserID(), pendingThread.Body(), pendingThread.FileID(),
	).Scan(&id, &fileName, &createdAt)
	if err != nil {
		return nil, toFileConstraintError(err)
	}

	filePath, err := domainthread.NewFilePath(r.baseThreadImagePath, fileName)
	if err != nil {
		return nil, err
	}

	return domainthread.NewThread(id, pendingThread.Body(), filePath, createdAt)
}

// toFileConstraintError maps violations of the threads.file_id constraints to the domain's sentinel errors;
// any other error is returned unchanged.
func toFileConstraintError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.ConstraintName {
	case "threads_file_id_fkey":
		return errors.Join(domainthread.ErrFileNotFound, err)
	case "threads_file_id_key":
		return errors.Join(domainthread.ErrFileAlreadyUsed, err)
	default:
		return err
	}
}
