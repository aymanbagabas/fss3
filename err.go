package fss3

import (
	"errors"
	"fmt"
)

var ErrNoFileInfo = errors.New("fileInfo not found")

type ErrInvalidHeader struct {
	name  string
	value string
	err   error
}

func (e ErrInvalidHeader) Error() string {
	return fmt.Sprintf("invalid header key: %s, value: %s: %s", e.name, e.value, e.err)
}

type ErrNotDirectory struct {
	name string
}

func (e ErrNotDirectory) Error() string {
	return fmt.Sprintf("'%s' not a directory", e.name)
}

type ErrNotEmpty struct {
	name string
}

func (e ErrNotEmpty) Error() string {
	return fmt.Sprintf("'%s' not empty", e.name)
}
