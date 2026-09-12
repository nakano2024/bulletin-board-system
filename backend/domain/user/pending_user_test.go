package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/user"
)

func TestNewPendingUser_正常系(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{
			name: "ipが正常な値のとき、PendingUserが生成されること",
			ip:   "203.0.113.1",
		},
		{
			name: "ipが1文字のとき、PendingUserが生成されること",
			ip:   "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := user.NewPendingUser(tt.ip)

			require.NoError(t, err)
			assert.Equal(t, tt.ip, got.IP())
		})
	}
}

func TestNewPendingUser_異常系(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		wantErr error
	}{
		{
			name:    "ipが空文字のとき、ErrUserIPEmptyが返ること",
			ip:      "",
			wantErr: user.ErrUserIPEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := user.NewPendingUser(tt.ip)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
