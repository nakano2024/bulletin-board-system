package thread

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

var fileNameBasePattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

var allowedFileExtensions = map[string]struct{}{
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".gif":  {},
	".webp": {},
	".svg":  {},
	".bmp":  {},
	".heif": {},
	".heic": {},
}

var ErrFileNameInvalid = errors.New("file name is invalid")

// FileName is a value object representing a stored file's bare name (no deployment-level path prefix).
type FileName struct {
	value string
}

// NewFileName validates name: the part before the extension must be half-width alphanumeric/underscore/hyphen
// (snake_case and kebab-case allowed), and the extension (case-insensitive) must be one of the allowed
// image formats (png, jpg, jpeg, gif, webp, svg, bmp, heif, heic).
func NewFileName(name string) (*FileName, error) {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	if base == "" || !fileNameBasePattern.MatchString(base) {
		return nil, ErrFileNameInvalid
	}
	if _, ok := allowedFileExtensions[strings.ToLower(ext)]; !ok {
		return nil, ErrFileNameInvalid
	}

	return &FileName{value: name}, nil
}

func (f *FileName) Value() string {
	return f.value
}
