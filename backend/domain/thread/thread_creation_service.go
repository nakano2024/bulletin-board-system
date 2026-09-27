package thread

import "context"

// ThreadCreationService creates a thread after checking the rule that a thread's image file must exist
// and must not be attached to any other thread. The check spans files and threads, so it lives here rather than on PendingThread.
type ThreadCreationService struct {
	fileChecker             IThreadFileChecker
	pendingThreadRepository IPendingThreadRepository
}

func NewThreadCreationService(fileChecker IThreadFileChecker, pendingThreadRepository IPendingThreadRepository) *ThreadCreationService {
	return &ThreadCreationService{fileChecker: fileChecker, pendingThreadRepository: pendingThreadRepository}
}

// Create returns ErrFileNotFound if the file does not exist, ErrFileAlreadyUsed if it is attached to another thread,
// and otherwise persists pendingThread.
func (s *ThreadCreationService) Create(ctx context.Context, pendingThread *PendingThread) (*Thread, error) {
	fileExists, err := s.fileChecker.ExistsFile(ctx, pendingThread.FileID())
	if err != nil {
		return nil, err
	}
	if !fileExists {
		return nil, ErrFileNotFound
	}

	fileAttached, err := s.fileChecker.IsFileAttachedToThread(ctx, pendingThread.FileID())
	if err != nil {
		return nil, err
	}
	if fileAttached {
		return nil, ErrFileAlreadyUsed
	}

	return s.pendingThreadRepository.CreateThread(ctx, pendingThread)
}
