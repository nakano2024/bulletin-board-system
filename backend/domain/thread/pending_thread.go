package thread

import "errors"

var (
	ErrPendingThreadBodyEmpty       = errors.New("pending thread body is empty")
	ErrPendingThreadFileNameMissing = errors.New("pending thread file name is missing")
)

// PendingThread represents a thread being created, before an id and creation time have been assigned.
type PendingThread struct {
	body     string
	fileName *FileName
}

// NewPendingThread validates the invariants required to create a thread: body must not be empty,
// and fileName must not be nil (image attachment is mandatory for thread creation).
// The image format itself is FileName's own rule; NewPendingThread only checks presence.
func NewPendingThread(body string, fileName *FileName) (*PendingThread, error) {
	if body == "" {
		return nil, ErrPendingThreadBodyEmpty
	}
	if fileName == nil {
		return nil, ErrPendingThreadFileNameMissing
	}

	return &PendingThread{body: body, fileName: fileName}, nil
}

func (p *PendingThread) Body() string {
	return p.body
}

func (p *PendingThread) FileName() *FileName {
	return p.fileName
}
