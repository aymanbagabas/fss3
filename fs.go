package fss3

import (
	"io/fs"
)

// FS implements the fs.FS interface.
type FS struct {
	fss3 *FSS3
}

// Open opens the named object for reading.
func (f *FS) Open(name string) (fs.File, error) {
	return f.fss3.Open(name)
}

// Stat returns a FileInfo for the given name.
func (f *FS) Stat(name string) (fs.FileInfo, error) {
	return f.fss3.Stat(name)
}

// ReadFile reads the whole object into a byte slice.
func (f *FS) ReadFile(name string) ([]byte, error) {
	return f.fss3.ReadFile(name)
}

// ReadDir reads the directory and returns a list of DirEntry.
func (f *FS) ReadDir(name string) ([]fs.DirEntry, error) {
	return f.fss3.ReadDir(name)
}
