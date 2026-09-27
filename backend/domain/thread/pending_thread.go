package thread

import "errors"

var (
	ErrPendingThreadUserIDEmpty = errors.New("pending thread user id is empty")
	ErrPendingThreadBodyEmpty   = errors.New("pending thread body is empty")
	ErrPendingThreadFileIDEmpty = errors.New("pending thread file id is empty")
)

// PendingThread is a thread being created. It references its attached file by the id of an already-registered file.
type PendingThread struct {
	userID string
	body   string
	fileID string
}

func NewPendingThread(userID, body, fileID string) (*PendingThread, error) {
	if userID == "" {
		return nil, ErrPendingThreadUserIDEmpty
	}
	if body == "" {
		return nil, ErrPendingThreadBodyEmpty
	}
	if fileID == "" {
		return nil, ErrPendingThreadFileIDEmpty
	}

	return &PendingThread{userID: userID, body: body, fileID: fileID}, nil
}

func (p *PendingThread) UserID() string {
	return p.userID
}

func (p *PendingThread) Body() string {
	return p.body
}

func (p *PendingThread) FileID() string {
	return p.fileID
}
