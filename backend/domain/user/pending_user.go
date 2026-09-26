package user

import "errors"

var ErrPendingUserCreateDateMissing = errors.New("pending user create date is missing")

// PendingUser represents a poster identity being resolved, before an id has been assigned.
type PendingUser struct {
	ip         string
	createDate *UserCreateDate
}

// NewPendingUser validates the invariants required to resolve a user: ip must not be empty,
// createDate must not be nil. The result is complete except for the id, which is assigned on persistence.
func NewPendingUser(ip string, createDate *UserCreateDate) (*PendingUser, error) {
	if ip == "" {
		return nil, ErrUserIPEmpty
	}
	if createDate == nil {
		return nil, ErrPendingUserCreateDateMissing
	}

	return &PendingUser{ip: ip, createDate: createDate}, nil
}

func (p *PendingUser) IP() string {
	return p.ip
}

func (p *PendingUser) CreateDate() *UserCreateDate {
	return p.createDate
}
