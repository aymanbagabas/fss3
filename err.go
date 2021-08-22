package fss3

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a file info is not found.
var ErrNoFileInfo = errors.New("fileInfo not found")

// ErrInvalidPath is returned when an invalid path is provided.
type ErrInvalidHeader struct {
	name  string
	value string
	err   error
}

func (e ErrInvalidHeader) Error() string {
	return fmt.Sprintf("invalid header key: %s, value: %s: %s", e.name, e.value, e.err)
}

// ErrNotDirectory is returned when a path is not a directory.
type ErrNotDirectory struct {
	name string
}

func (e ErrNotDirectory) Error() string {
	return fmt.Sprintf("'%s' not a directory", e.name)
}

// ErrNotEmpty is returned when a directory is not empty.
type ErrNotEmpty struct {
	name string
}

func (e ErrNotEmpty) Error() string {
	return fmt.Sprintf("'%s' not empty", e.name)
}
