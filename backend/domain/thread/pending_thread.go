package thread

import "errors"

var (
	ErrPendingThreadUserIDEmpty     = errors.New("pending thread user id is empty")
	ErrPendingThreadBodyEmpty       = errors.New("pending thread body is empty")
	ErrPendingThreadFileNameMissing = errors.New("pending thread file name is missing")
)

type PendingThread struct {
	userID   string
	body     string
	fileName *FileName
}

func NewPendingThread(userID, body string, fileName *FileName) (*PendingThread, error) {
	if userID == "" {
		return nil, ErrPendingThreadUserIDEmpty
	}
	if body == "" {
		return nil, ErrPendingThreadBodyEmpty
	}
	if fileName == nil {
		return nil, ErrPendingThreadFileNameMissing
	}

	return &PendingThread{userID: userID, body: body, fileName: fileName}, nil
}

func (p *PendingThread) UserID() string {
	return p.userID
}

func (p *PendingThread) Body() string {
	return p.body
}

func (p *PendingThread) FileName() *FileName {
	return p.fileName
}
