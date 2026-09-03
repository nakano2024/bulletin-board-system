package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	applicationthread "github.com/nakanokota/bulletin-board-system/backend/application/thread"
)

// querier is satisfied by both *pgxpool.Pool and pgx.Tx, so tests can run
// FetchThreadList inside a transaction for isolation.
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type PostgresThreadQueryService struct {
	db querier
}

func NewPostgresThreadQueryService(db querier) *PostgresThreadQueryService {
	return &PostgresThreadQueryService{db: db}
}

func (s *PostgresThreadQueryService) FetchThreadList(ctx context.Context) ([]applicationthread.ThreadListItem, error) {
	rows, err := s.db.Query(ctx, "SELECT id, body, image_path, created_at FROM threads ORDER BY created_at ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]applicationthread.ThreadListItem, 0)
	for rows.Next() {
		var item applicationthread.ThreadListItem
		if err := rows.Scan(&item.ID, &item.Body, &item.ImagePath, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
