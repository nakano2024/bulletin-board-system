package thread

import (
	"errors"
	"time"
)

var (
	ErrThreadIDEmpty         = errors.New("thread id is empty")
	ErrThreadBodyEmpty       = errors.New("thread body is empty")
	ErrThreadCreatedAtZero   = errors.New("thread created at is zero")
	ErrThreadFilePathMissing = errors.New("thread file path is missing")
)

type Thread struct {
	id        string
	body      string
	filePath  *FilePath
	createdAt time.Time
}

// NewThread reconstructs a Thread, validating its invariants. filePath must not be nil — image attachment is mandatory.
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
	if filePath == nil {
		return nil, ErrThreadFilePathMissing
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

// FilePathValue returns the file path as a plain string.
func (t *Thread) FilePathValue() string {
	return t.filePath.Value()
}

func (t *Thread) CreatedAt() time.Time {
	return t.createdAt
}
