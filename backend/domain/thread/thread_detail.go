package thread

import (
	"errors"
	"time"

	"github.com/nakanokota/bulletin-board-system/backend/domain/image_file"
)

var (
	ErrThreadDetailIDEmpty         = errors.New("thread detail id is empty")
	ErrThreadDetailUserIDEmpty     = errors.New("thread detail user id is empty")
	ErrThreadDetailBodyEmpty       = errors.New("thread detail body is empty")
	ErrThreadDetailCreatedAtZero   = errors.New("thread detail created at is zero")
	ErrThreadDetailFilePathMissing = errors.New("thread detail file path is missing")
	ErrThreadDetailRepliesNil      = errors.New("thread detail replies is nil")
)

// ThreadDetail is a thread as shown on its detail page: in addition to Thread's fields, it carries the id of the user who created it,
// whether it is still alive (so that a fallen-off (dat落ち) thread can be told apart), and its replies.
// It is the aggregate root for its replies.
type ThreadDetail struct {
	id        string
	userID    string
	body      string
	filePath  *image_file.FilePath
	createdAt time.Time
	isAlive   bool
	replies   []*Reply
}

// NewThreadDetail reconstructs a ThreadDetail, validating its invariants. filePath must not be nil — image attachment is mandatory.
// replies must not be nil; pass an empty slice when the thread has no replies.
func NewThreadDetail(id, userID, body string, filePath *image_file.FilePath, createdAt time.Time, isAlive bool, replies []*Reply) (*ThreadDetail, error) {
	if id == "" {
		return nil, ErrThreadDetailIDEmpty
	}
	if userID == "" {
		return nil, ErrThreadDetailUserIDEmpty
	}
	if body == "" {
		return nil, ErrThreadDetailBodyEmpty
	}
	if createdAt.IsZero() {
		return nil, ErrThreadDetailCreatedAtZero
	}
	if filePath == nil {
		return nil, ErrThreadDetailFilePathMissing
	}
	if replies == nil {
		return nil, ErrThreadDetailRepliesNil
	}

	return &ThreadDetail{id: id, userID: userID, body: body, filePath: filePath, createdAt: createdAt, isAlive: isAlive, replies: replies}, nil
}

func (t *ThreadDetail) ID() string {
	return t.id
}

func (t *ThreadDetail) UserID() string {
	return t.userID
}

func (t *ThreadDetail) Body() string {
	return t.body
}

func (t *ThreadDetail) FilePath() *image_file.FilePath {
	return t.filePath
}

func (t *ThreadDetail) CreatedAt() time.Time {
	return t.createdAt
}

// IsDead reports whether the thread has fallen off (dat落ち). A dead thread must not be shown.
func (t *ThreadDetail) IsDead() bool {
	return !t.isAlive
}

// Replies returns the thread's replies, newest first.
func (t *ThreadDetail) Replies() []*Reply {
	return t.replies
}
