package user

// PendingUser represents a poster identity being resolved, before an id has been assigned.
type PendingUser struct {
	ip string
}

// NewPendingUser validates the invariant required to resolve a user: ip must not be empty.
func NewPendingUser(ip string) (*PendingUser, error) {
	if ip == "" {
		return nil, ErrUserIPEmpty
	}

	return &PendingUser{ip: ip}, nil
}

func (p *PendingUser) IP() string {
	return p.ip
}
