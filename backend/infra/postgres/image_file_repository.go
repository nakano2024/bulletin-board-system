package postgres

import (
	"context"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"

	domainimagefile "github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
)

type PostgresImageFileRepository struct {
	db querier
}

func NewPostgresImageFileRepository(db querier) *PostgresImageFileRepository {
	return &PostgresImageFileRepository{db: db}
}

func (r *PostgresImageFileRepository) CreateImageFile(ctx context.Context, imageFile *domainimagefile.ImageFile) error {
	_, err := trmpgx.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db).Exec(ctx,
		"INSERT INTO files (id, name, content_type, size_bytes) VALUES ($1, $2, $3, $4)",
		imageFile.ID(), imageFile.Name().Value(), imageFile.ContentType(), imageFile.SizeBytes(),
	)
	return err
}
