package user_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nakanokota/bulletin-board-system/backend/domain/user"
)

func TestNewUserCreateDate_正常系(t *testing.T) {
	tests := []struct {
		name     string
		t        time.Time
		wantTime time.Time
	}{
		{
			name:     "時刻情報を含むtime.Timeを渡したとき、日付境界(0時0分0秒)に丸められること",
			t:        time.Date(2026, 9, 13, 15, 30, 45, 0, time.UTC),
			wantTime: time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
		},
		{
			name:     "すでに日付境界のtime.Timeを渡したとき、そのままの値になること",
			t:        time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
			wantTime: time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := user.NewUserCreateDate(tt.t)

			require.NoError(t, err)
			assert.Equal(t, tt.wantTime, got.Time())
		})
	}
}

func TestNewUserCreateDate_異常系(t *testing.T) {
	tests := []struct {
		name    string
		t       time.Time
		wantErr error
	}{
		{
			name:    "tがゼロ値のとき、ErrUserCreateDateZeroが返ること",
			t:       time.Time{},
			wantErr: user.ErrUserCreateDateZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := user.NewUserCreateDate(tt.t)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
