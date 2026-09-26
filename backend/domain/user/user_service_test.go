package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/nakanokota/bulletin-board-system/backend/domain/user"
	"github.com/nakanokota/bulletin-board-system/backend/domain/user/mock_user"
)

func TestUserService_FetchOrCreate_正常系(t *testing.T) {
	ip := "203.0.113.1"
	date, _ := user.NewUserCreateDate(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))
	pendingUser, _ := user.NewPendingUser(ip, date)
	existingUser, _ := user.NewUser("user-1", ip)
	createdUser, _ := user.NewUser("user-2", ip)

	tests := []struct {
		name      string
		setupMock func(*mock_user.MockIUserRepository)
		want      *user.User
	}{
		{
			name: "同一IP・同一日のUserが既に存在するとき、そのUserがそのまま返り、CreateUserは呼ばれないこと",
			setupMock: func(m *mock_user.MockIUserRepository) {
				m.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(existingUser, nil)
				m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			want: existingUser,
		},
		{
			name: "同一IP・同一日のUserが存在しないとき、CreateUserで作成されたUserが返ること",
			setupMock: func(m *mock_user.MockIUserRepository) {
				m.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(nil, nil)
				m.EXPECT().CreateUser(gomock.Any(), pendingUser).Return(createdUser, nil)
			},
			want: createdUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userRepository := mock_user.NewMockIUserRepository(ctrl)
			tt.setupMock(userRepository)

			sut := user.NewUserService(userRepository)
			got, err := sut.FetchOrCreate(context.Background(), ip, date)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserService_FetchOrCreate_異常系(t *testing.T) {
	ip := "203.0.113.1"
	date, _ := user.NewUserCreateDate(time.Date(2026, 9, 13, 0, 0, 0, 0, time.UTC))
	pendingUser, _ := user.NewPendingUser(ip, date)
	errRepositoryFailed := errors.New("repository failed")

	tests := []struct {
		name      string
		ip        string
		setupMock func(*mock_user.MockIUserRepository)
		wantErr   error
	}{
		{
			name: "FindByIPAndDateがエラーを返すとき、FetchOrCreateはエラーを返し、CreateUserは呼ばれないこと",
			ip:   ip,
			setupMock: func(m *mock_user.MockIUserRepository) {
				m.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(nil, errRepositoryFailed)
				m.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: errRepositoryFailed,
		},
		{
			name: "Userが存在せずCreateUserがエラーを返すとき、FetchOrCreateはそのエラーを返すこと",
			ip:   ip,
			setupMock: func(m *mock_user.MockIUserRepository) {
				m.EXPECT().FindByIPAndDate(gomock.Any(), ip, date).Return(nil, nil)
				m.EXPECT().CreateUser(gomock.Any(), pendingUser).Return(nil, errRepositoryFailed)
			},
			wantErr: errRepositoryFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			userRepository := mock_user.NewMockIUserRepository(ctrl)
			tt.setupMock(userRepository)

			sut := user.NewUserService(userRepository)
			_, err := sut.FetchOrCreate(context.Background(), tt.ip, date)

			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
