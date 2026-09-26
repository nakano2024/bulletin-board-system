package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/user"
)

func TestNewPendingUser_正常系(t *testing.T) {
	createDate, _ := user.NewUserCreateDate(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name       string
		ip         string
		createDate *user.UserCreateDate
	}{
		{
			name:       "ipとcreateDateがともに正常な値のとき、PendingUserが生成されること",
			ip:         "203.0.113.1",
			createDate: createDate,
		},
		{
			name:       "ipが1文字のとき、PendingUserが生成されること",
			ip:         "a",
			createDate: createDate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := user.NewPendingUser(tt.ip, tt.createDate)

			require.NoError(t, err)
			assert.Equal(t, tt.ip, got.IP())
			assert.Equal(t, tt.createDate, got.CreateDate())
		})
	}
}

func TestNewPendingUser_異常系(t *testing.T) {
	createDate, _ := user.NewUserCreateDate(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name       string
		ip         string
		createDate *user.UserCreateDate
		wantErr    error
	}{
		{
			name:       "ipが空文字のとき、ErrUserIPEmptyが返ること",
			ip:         "",
			createDate: createDate,
			wantErr:    user.ErrUserIPEmpty,
		},
		{
			name:       "createDateがnilのとき、ErrPendingUserCreateDateMissingが返ること",
			ip:         "203.0.113.1",
			createDate: nil,
			wantErr:    user.ErrPendingUserCreateDateMissing,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := user.NewPendingUser(tt.ip, tt.createDate)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
