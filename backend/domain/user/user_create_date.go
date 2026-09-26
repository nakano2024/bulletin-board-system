package user

import (
	"errors"
	"time"
)

var ErrUserCreateDateZero = errors.New("user create date is zero")

type UserCreateDate struct {
	value time.Time
}

func NewUserCreateDate(t time.Time) (*UserCreateDate, error) {
	if t.IsZero() {
		return nil, ErrUserCreateDateZero
	}

	return &UserCreateDate{value: time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())}, nil
}

func (d *UserCreateDate) Time() time.Time {
	return d.value
}
