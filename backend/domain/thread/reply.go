package thread

import (
	"errors"
	"time"
	"unicode/utf8"

	"github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
)

// MaxReplyBodyLength is the maximum number of characters (runes) in a reply's body.
const MaxReplyBodyLength = 1000

var (
	ErrReplyIDEmpty       = errors.New("reply id is empty")
	ErrReplyUserIDEmpty   = errors.New("reply user id is empty")
	ErrReplyBodyEmpty     = errors.New("reply body is empty")
	ErrReplyBodyTooLong   = errors.New("reply body is too long")
	ErrReplyCreatedAtZero = errors.New("reply created at is zero")
)

// Reply is a reply posted to a thread. It belongs to the ThreadDetail aggregate, so it does not carry its thread's id.
type Reply struct {
	id        string
	userID    string
	body      string
	filePath  *image_file.FilePath
	createdAt time.Time
}

// NewReply reconstructs a Reply, validating its invariants: body must be 1 to MaxReplyBodyLength characters. Unlike a thread, attaching an image is optional, so filePath may be nil.
func NewReply(id, userID, body string, filePath *image_file.FilePath, createdAt time.Time) (*Reply, error) {
	if id == "" {
		return nil, ErrReplyIDEmpty
	}
	if userID == "" {
		return nil, ErrReplyUserIDEmpty
	}
	if body == "" {
		return nil, ErrReplyBodyEmpty
	}
	if utf8.RuneCountInString(body) > MaxReplyBodyLength {
		return nil, ErrReplyBodyTooLong
	}
	if createdAt.IsZero() {
		return nil, ErrReplyCreatedAtZero
	}

	return &Reply{id: id, userID: userID, body: body, filePath: filePath, createdAt: createdAt}, nil
}

func (r *Reply) ID() string {
	return r.id
}

func (r *Reply) UserID() string {
	return r.userID
}

func (r *Reply) Body() string {
	return r.body
}

// FilePath returns nil when no image is attached.
func (r *Reply) FilePath() *image_file.FilePath {
	return r.filePath
}

func (r *Reply) CreatedAt() time.Time {
	return r.createdAt
}
