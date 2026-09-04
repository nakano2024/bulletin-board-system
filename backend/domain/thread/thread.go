package thread

import (
	"errors"
	"time"
)

var (
	ErrThreadIDEmpty       = errors.New("thread id is empty")
	ErrThreadBodyEmpty     = errors.New("thread body is empty")
	ErrThreadCreatedAtZero = errors.New("thread created at is zero")
)

type Thread struct {
	id        string
	body      string
	filePath  *FilePath
	createdAt time.Time
}

// NewThread reconstructs a Thread, validating its invariants. filePath may be nil (no image attached).
func NewThread(id, body string, filePath *FilePath, createdAt time.Time) (*Thread, error) {
	if id == "" {
		return nil, ErrThreadIDEmpty
	}
	if body == "" {
		return nil, ErrThreadBodyEmpty
	}
	if createdAt.IsZero() {
		return nil, ErrThreadCreatedAtZero
	}

	return &Thread{id: id, body: body, filePath: filePath, createdAt: createdAt}, nil
}

func (t *Thread) ID() string {
	return t.id
}

func (t *Thread) Body() string {
	return t.body
}

func (t *Thread) FilePath() *FilePath {
	return t.filePath
}

// FilePathValue returns the file path as a plain string, or "" when no image is attached.
func (t *Thread) FilePathValue() string {
	if t.filePath == nil {
		return ""
	}
	return t.filePath.Value()
}

func (t *Thread) CreatedAt() time.Time {
	return t.createdAt
}
