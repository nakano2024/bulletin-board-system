package user_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/user"
)

func TestNewUser_正常系(t *testing.T) {
	tests := []struct {
		name string
		id   string
		ip   string
	}{
		{
			name: "idとipがともに正常な値のとき、Userが生成されること",
			id:   "user-1",
			ip:   "203.0.113.1",
		},
		{
			name: "idとipがともに1文字のとき、Userが生成されること",
			id:   "a",
			ip:   "b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := user.NewUser(tt.id, tt.ip)

			require.NoError(t, err)
			assert.Equal(t, tt.id, got.ID())
			assert.Equal(t, tt.ip, got.IP())
		})
	}
}

func TestNewUser_異常系(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		ip      string
		wantErr error
	}{
		{
			name:    "idが空文字のとき、ErrUserIDEmptyが返ること",
			id:      "",
			ip:      "203.0.113.1",
			wantErr: user.ErrUserIDEmpty,
		},
		{
			name:    "ipが空文字のとき、ErrUserIPEmptyが返ること",
			id:      "user-1",
			ip:      "",
			wantErr: user.ErrUserIPEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := user.NewUser(tt.id, tt.ip)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
