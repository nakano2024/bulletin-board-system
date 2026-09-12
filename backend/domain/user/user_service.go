package user

import (
	"context"
	"time"
)

// UserService resolves the poster identity for an IP address on a given day:
// reuse the existing User if one was already assigned, otherwise create one.
type UserService struct {
	userRepository IUserRepository
}

func NewUserService(userRepository IUserRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (s *UserService) CreateOrFetch(ctx context.Context, ip string, date time.Time) (*User, error) {
	existing, err := s.userRepository.FindByIPAndDate(ctx, ip, date)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	pendingUser, err := NewPendingUser(ip)
	if err != nil {
		return nil, err
	}

	return s.userRepository.CreateUser(ctx, pendingUser, date)
}
