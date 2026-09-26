package user

import (
	"context"
)

type UserService struct {
	userRepository IUserRepository
}

func NewUserService(userRepository IUserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) FetchOrCreate(ctx context.Context, ip string, date *UserCreateDate) (*User, error) {
	existing, err := s.userRepository.FindByIPAndDate(ctx, ip, date)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	pendingUser, err := NewPendingUser(ip, date)
	if err != nil {
		return nil, err
	}

	return s.userRepository.CreateUser(ctx, pendingUser)
}
