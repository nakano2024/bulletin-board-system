package image_file

import (
	"errors"
	"regexp"
)

// basePathPattern accepts either a bare directory-style prefix ("thread_images/") or a full
// http(s) URL prefix ("http://localhost:9000/images/"), always ending in "/".
var basePathPattern = regexp.MustCompile(`^[A-Za-z0-9_]+/$|^https?://[A-Za-z0-9.\-]+(?::[0-9]+)?(?:/[A-Za-z0-9_\-]+)*/$`)

var ErrFilePathBasePathInvalid = errors.New("file path base path is invalid")

// FilePath is a value object representing the storage path of an attached file,
// composed of a base path (deployment-level, e.g. from an env var) and a file name (per-thread).
type FilePath struct {
	basePath string
	fileName *FileName
}

// NewFilePath validates basePath (either "thread_images/"-style or a full URL prefix like
// "http://localhost:9000/images/", always ending in "/"), validates fileName via FileName's rule,
// then holds them to be combined by Value().
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
