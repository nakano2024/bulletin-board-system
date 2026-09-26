package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainuser "github.com/nakanokota/bulletin-board-system/backend/domain/user"
	"github.com/nakanokota/bulletin-board-system/backend/infra/postgres"
)

func TestPostgresUserRepository_FindByIPAndDate_正常系(t *testing.T) {
	pool := newTestPool(t)
	ip := "203.0.113.1"
	date := mustUserCreateDate(t, time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))
	existingUser := mustNewUser(t, "user-1", ip)

	tests := []struct {
		name   string
		seedIP string
		seedID string
		want   *domainuser.User
	}{
		{
			name:   "指定したip・dateに一致するUserが存在するとき、そのUserが返ること",
			seedIP: ip,
			seedID: "user-1",
			want:   existingUser,
		},
		{
			name:   "指定したip・dateに一致するUserが存在しないとき、nilが返ること",
			seedIP: "203.0.113.9",
			seedID: "user-2",
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { tx.Rollback(ctx) })

			_, err = tx.Exec(ctx, "INSERT INTO users (id, ip, create_date) VALUES ($1, $2, $3)", tt.seedID, tt.seedIP, date.Time())
			require.NoError(t, err)

			sut := postgres.NewPostgresUserRepository(tx)
			got, err := sut.FindByIPAndDate(ctx, ip, date)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPostgresUserRepository_CreateUser_正常系(t *testing.T) {
	pool := newTestPool(t)
	ip := "203.0.113.1"
	date := mustUserCreateDate(t, time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name        string
		pendingUser *domainuser.PendingUser
	}{
		{
			name:        "有効なPendingUserを渡したとき、idが採番されたUserが返り、DBにも保存されること",
			pendingUser: mustNewPendingUser(t, ip, date),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			tx, err := pool.Begin(ctx)
			require.NoError(t, err)
			t.Cleanup(func() { tx.Rollback(ctx) })

			sut := postgres.NewPostgresUserRepository(tx)
			got, err := sut.CreateUser(ctx, tt.pendingUser)

			require.NoError(t, err)
			assert.NotEmpty(t, got.ID())
			assert.Equal(t, ip, got.IP())

			var storedIP string
			var storedDate time.Time
			err = tx.QueryRow(ctx, "SELECT ip, create_date FROM users WHERE id = $1", got.ID()).Scan(&storedIP, &storedDate)
			require.NoError(t, err)
			assert.Equal(t, ip, storedIP)
			assert.True(t, date.Time().Equal(storedDate))
		})
	}
}

func mustUserCreateDate(t *testing.T, tm time.Time) *domainuser.UserCreateDate {
	t.Helper()
	d, err := domainuser.NewUserCreateDate(tm)
	require.NoError(t, err)
	return d
}

func mustNewUser(t *testing.T, id, ip string) *domainuser.User {
	t.Helper()
	u, err := domainuser.NewUser(id, ip)
	require.NoError(t, err)
	return u
}

func mustNewPendingUser(t *testing.T, ip string, date *domainuser.UserCreateDate) *domainuser.PendingUser {
	t.Helper()
	pu, err := domainuser.NewPendingUser(ip, date)
	require.NoError(t, err)
	return pu
}
