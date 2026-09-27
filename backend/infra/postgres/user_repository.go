package postgres

import (
	"context"
	"errors"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5"

	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
)

type PostgresUserRepository struct {
	db querier
}

func NewPostgresUserRepository(db querier) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) FindByIPAndDate(ctx context.Context, ip string, date *domainuser.UserCreateDate) (*domainuser.User, error) {
	var id string

	err := trmpgx.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db).QueryRow(ctx, "SELECT id FROM users WHERE ip = $1 AND create_date = $2", ip, date.Time()).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return domainuser.NewUser(id, ip)
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, pendingUser *domainuser.PendingUser) (*domainuser.User, error) {
	var id string

	err := trmpgx.DefaultCtxGetter.DefaultTrOrDB(ctx, r.db).QueryRow(
		ctx,
		"INSERT INTO users (ip, create_date) VALUES ($1, $2) RETURNING id",
		pendingUser.IP(), pendingUser.CreateDate().Time(),
	).Scan(&id)
	if err != nil {
		return nil, err
	}

	return domainuser.NewUser(id, pendingUser.IP())
}
