package user

import "errors"

var (
	ErrUserIDEmpty = errors.New("user id is empty")
	ErrUserIPEmpty = errors.New("user ip is empty")
)

// User represents a poster identity, scoped to an IP address and a calendar day.
type User struct {
	id string
	ip string
}

// NewUser reconstructs a User, validating its invariants.
func NewUser(id, ip string) (*User, error) {
	if id == "" {
		return nil, ErrUserIDEmpty
	}
	if ip == "" {
		return nil, ErrUserIPEmpty
	}

	return &User{id: id, ip: ip}, nil
}

func (u *User) ID() string {
	return u.id
}

func (u *User) IP() string {
	return u.ip
}
