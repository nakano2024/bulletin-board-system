package thread

import (
	"errors"
	"regexp"
)

var basePathPattern = regexp.MustCompile(`^[A-Za-z0-9_]+/$`)

var ErrFilePathBasePathInvalid = errors.New("file path base path is invalid")

// FilePath is a value object representing the storage path of an attached file,
// composed of a base path (deployment-level, e.g. from an env var) and a file name (per-thread).
type FilePath struct {
	basePath string
	fileName *FileName
}

// NewFilePath validates basePath (half-width alphanumeric/underscore + trailing "/", e.g. "thread_images/"),
// validates fileName via FileName's rule, then holds them to be combined by Value().
func NewFilePath(basePath, fileName string) (*FilePath, error) {
	if !basePathPattern.MatchString(basePath) {
		return nil, ErrFilePathBasePathInvalid
	}
	fn, err := NewFileName(fileName)
	if err != nil {
		return nil, err
	}

	return &FilePath{basePath: basePath, fileName: fn}, nil
}

func (f *FilePath) Value() string {
	return f.basePath + f.fileName.Value()
}
