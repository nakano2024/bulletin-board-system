package postgres

import "context"

// PostgresThreadFileChecker implements domainthread.IThreadFileChecker against the files and threads tables.
type PostgresThreadFileChecker struct {
	db querier
}

func NewPostgresThreadFileChecker(db querier) *PostgresThreadFileChecker {
	return &PostgresThreadFileChecker{db: db}
}

func (c *PostgresThreadFileChecker) ExistsFile(ctx context.Context, fileID string) (bool, error) {
	var exists bool
	err := querierFrom(ctx, c.db).QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM files WHERE id = $1)", fileID).Scan(&exists)
	return exists, err
}

// IsFileAttachedToThread counts threads regardless of is_alive, matching the UNIQUE constraint on threads.file_id.
func (c *PostgresThreadFileChecker) IsFileAttachedToThread(ctx context.Context, fileID string) (bool, error) {
	var attached bool
	err := querierFrom(ctx, c.db).QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM threads WHERE file_id = $1)", fileID).Scan(&attached)
	return attached, err
}
