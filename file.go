package fss3

import (
	"io/fs"
	"strconv"
	"strings"
	"time"
)

// FileInfo implements fs.FileInfo.
type FileInfo struct {
	info    *objectInfo
	size    int64
	modTime time.Time
}

// File implements fs.File.
type File struct {
	fs       *FS
	obj      *object
	fileInfo *FileInfo
}

// DirEntry implements fs.DirEntry.
type DirEntry struct {
	info *FileInfo
}

// Name returns the base name of the object extracted from its key.
func (fi *FileInfo) Name() string {
	return keyBaseName(fi.info.Key)
}

// Size returns the file size from the object
func (fi *FileInfo) Size() int64 {
	if fi.size == 0 {
		fi.size = fi.info.Size
	}
	return fi.size
}

// Mode returns the file mode bits from the object metadata.
// Returns 0 on parsing error.
func (fi *FileInfo) Mode() fs.FileMode {
	modeMeta := fi.info.UserMetadata["Mode"]
	mode, err := strconv.ParseUint(modeMeta, 8, 32)
	if err != nil {
		return 0
	}
	return fs.FileMode(mode)
}

// ModTime returns the last modification time of the object.
func (fi *FileInfo) ModTime() time.Time {
	if fi.modTime.IsZero() {
		fi.modTime = fi.info.LastModified
	}
	return fi.modTime
}

// IsDir is an abbreviation for Mode().IsDir()
func (fi *FileInfo) IsDir() bool {
	return fi.Mode().IsDir()
}

// Sys is the underlying data source for the object.
// In this case it is the ObjectInfo.
func (fi *FileInfo) Sys() interface{} {
	return fi.info
}

// Stat returns the FileInfo structure describing this object.
func (f *File) Stat() (fs.FileInfo, error) {
	if f.fileInfo == nil {
		return nil, ErrNoFileInfo
	}
	return f.fileInfo, nil
}

// Read reads up to len(b) bytes from the underlying object.
func (f *File) Read(b []byte) (int, error) {
	return f.obj.Read(b)
}

// Close closes the object.
func (f *File) Close() error {
	return f.obj.Close()
}

// Name returns the base name of the object extracted from its key.
func (de *DirEntry) Name() string {
	return de.info.Name()
}

// Type returns the type bits for the entry.
func (de *DirEntry) Type() fs.FileMode {
	return de.info.Mode()
}

// IsDir reports whether the entry is a directory.
func (de *DirEntry) IsDir() bool {
	return de.info.IsDir()
}

// Info returns the FileInfo structure describing this entry.
func (de *DirEntry) Info() (fs.FileInfo, error) {
	if de.info == nil {
		return nil, ErrNoFileInfo
	}
	return de.info, nil
}

// ReadDir returns the entries from a directory.
func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	var ents []fs.DirEntry
	fStat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	stat := fStat.(*FileInfo)
	dirKey := stat.info.Key
	if stat.IsDir() && (len(dirKey) > len(f.fs.fss3.cfg.DirFileName) && !strings.HasSuffix(dirKey, "/"+f.fs.fss3.cfg.DirFileName)) && dirKey != f.fs.fss3.cfg.DirFileName {
		return nil, ErrNotDirectory{name: fStat.Name()}
	}

	var prefix string
	if dirKey == f.fs.fss3.cfg.DirFileName {
		prefix = ""
	} else {
		prefix = dirKey[:len(dirKey)-len(f.fs.fss3.cfg.DirFileName)]
	}

	opts := listObjectsOptions{
		Prefix:       prefix,
		Recursive:    false,
		WithMetadata: true,
	}

	counter := 0
	for objInfo := range f.fs.fss3.listObjects(&opts) {
		if n > 0 && counter >= n {
			break
		}
		if objInfo.Err != nil {
			return ents, objInfo.Err
		}
		// Skip the current directory
		if strings.HasSuffix(objInfo.Key, f.fs.fss3.cfg.DirFileName) {
			continue
		}

		oi := objInfo
		key := oi.Key
		fi := FileInfo{info: &oi}
		// If directory, set directory key
		if strings.HasSuffix(oi.Key, "/") {
			key += f.fs.fss3.cfg.DirFileName
		}
		// AWS S3 API doesn't return Metadata on listObjects
		// We have to fetch the stats to get the metadata
		// We also fetch the stats when it's a directory
		// Reference: https://github.com/minio/minio-go/issues/1462
		if len(oi.UserMetadata) == 0 || len(oi.Metadata) == 0 || strings.HasSuffix(oi.Key, "/") {
			stat, err := f.fs.Stat(key)
			if err != nil {
				return ents, err
			}
			fi = *stat.(*FileInfo)
		}
		ent := DirEntry{info: &fi}
		ents = append(ents, &ent)

		counter++
	}

	return ents, nil
}
