package thread

import (
	"errors"
	"regexp"
)

var (
	basePathPattern = regexp.MustCompile(`^[A-Za-z0-9_]+/$`)
	fileNamePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+(\.[A-Za-z0-9_-]+)?$`)
)

var (
	ErrFilePathBasePathInvalid = errors.New("file path base path is invalid")
	ErrFilePathFileNameInvalid = errors.New("file path file name is invalid")
)

// FilePath is a value object representing the storage path of an attached file,
// composed of a base path (deployment-level, e.g. from an env var) and a file name (per-thread).
type FilePath struct {
	value string
}

// NewFilePath validates basePath (half-width alphanumeric/underscore + trailing "/", e.g. "thread_images/")
// and fileName (half-width alphanumeric/underscore/hyphen — snake_case and kebab-case allowed —
// optionally followed by a "." + the same character set as an extension, e.g. "sample", "sample_file",
// "sample-file", or "sample.png"), then combines them into a single value.
func NewFilePath(basePath, fileName string) (*FilePath, error) {
	if !basePathPattern.MatchString(basePath) {
		return nil, ErrFilePathBasePathInvalid
	}
	if !fileNamePattern.MatchString(fileName) {
		return nil, ErrFilePathFileNameInvalid
	}

	return &FilePath{value: basePath + fileName}, nil
}

func (f *FilePath) Value() string {
	return f.value
}
