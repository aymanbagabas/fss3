package fss3

import (
	"io/fs"
	"io/ioutil"
	"log"
	"strings"
)

// Open opens a S3 file using the given name.
func (fss3 *FSS3) Open(name string) (fs.File, error) {
	name = sanitizeName(name)
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{
			Op:   "open",
			Path: name,
			Err:  fs.ErrInvalid,
		}
	}

	// Set the initial isDir to the root directory key
	isDir := name == fss3.cfg.DirFileName
	stat, err := fss3.statObject(name, nil)
	if err != nil {
		// Check if the requested path is a directory
		rspErr := errToRspErr(err)
		if rspErr.Code == "NoSuchKey" {
			isDir = true
			// Check if name is the root path
			if name != fss3.cfg.DirFileName {
				name = name + "/" + fss3.cfg.DirFileName
			}
			dirStat, dirErr := fss3.statObject(name, nil)
			if dirErr != nil {
				return nil, minioErrToPathErr(err)
			}
			stat = dirStat
		} else {
			return nil, minioErrToPathErr(err)
		}
	}

	obj, err := fss3.getObject(name, nil)
	if err != nil {
		return nil, minioErrToPathErr(err)
	}

	fileInfo := FileInfo{
		info:    stat,
		size:    stat.Size,
		modTime: stat.LastModified,
	}
	// If directory, get the last modified time and calculate the size
	if isDir && strings.HasSuffix(name, fss3.cfg.DirFileName) {
		var prefix string
		if name == fss3.cfg.DirFileName {
			prefix = ""
		} else {
			prefix = name[:len(name)-len(fss3.cfg.DirFileName)]
		}
		opts := listObjectsOptions{
			Recursive:    true,
			Prefix:       prefix,
			WithMetadata: true,
		}
		for obj := range fss3.listObjects(&opts) {
			if obj.Err != nil {
				log.Printf("warning: %s", obj.Err)
				continue
			}
			fileInfo.size += obj.Size
			if obj.LastModified.After(fileInfo.modTime) {
				fileInfo.modTime = obj.LastModified
			}
		}
	}

	file := File{
		fs:       fss3,
		obj:      obj,
		fileInfo: &fileInfo,
	}

	return &file, nil
}

// Stat returns a FileInfo for the given name.
func (fss3 *FSS3) Stat(name string) (fs.FileInfo, error) {
	f, err := fss3.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return stat, nil
}

// ReadFile reads the whole object into a byte slice.
func (fss3 *FSS3) ReadFile(name string) ([]byte, error) {
	f, err := fss3.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ReadDir reads the directory and returns a list of DirEntry.
func (fss3 *FSS3) ReadDir(name string) ([]fs.DirEntry, error) {
	f, err := fss3.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	file := f.(*File)
	return file.ReadDir(0)
}
