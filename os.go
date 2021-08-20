package fss3

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

// Create creates or truncates the named object.
// The object is created with mode 0666 (before umask).
func (fss3 *FSS3) Create(name string) (*File, error) {
	buf := bytes.NewBuffer([]byte{})
	key := sanitizeName(name)
	parent := sanitizeName(filepath.Dir(key))

	err := fss3.MkdirAll(parent, umask(fss3.cfg.Umask, fs.ModeDir|fs.ModePerm))
	if err != nil {
		return nil, err
	}
	opts := putObjectOptions{
		UserMetadata: map[string]string{
			"mode": fmt.Sprintf("%o", umask(fss3.cfg.Umask, 0666)),
		},
		ContentType: guessContentType(name),
	}
	_, err = fss3.putObject(key, buf, int64(buf.Len()), &opts)
	if err != nil {
		return nil, minioErrToPathErr(err)
	}

	f, err := fss3.Open(name)
	if err != nil {
		return nil, err
	}

	file := f.(*File)
	return file, nil
}

// Mkdir creates a new directory with the specified name and permission bits.
func (fss3 *FSS3) Mkdir(name string, mode fs.FileMode) error {
	name = sanitizeName(name)
	key := name
	buf := bytes.NewBuffer([]byte{})
	if name != fss3.cfg.DirFileName {
		key = name + "/" + fss3.cfg.DirFileName
	}

	opts := putObjectOptions{
		UserMetadata: map[string]string{
			"mode": fmt.Sprintf("%o", umask(fss3.cfg.Umask, mode|fs.ModeDir)),
		},
	}
	_, err := fss3.putObject(key, buf, int64(buf.Len()), &opts)
	if err != nil {
		return minioErrToPathErr(err)
	}

	return nil
}

// MkdirAll creates a directory named path, along with any necessary parents,
func (fss3 *FSS3) MkdirAll(name string, mode fs.FileMode) error {
	name = sanitizeName(name)
	tokens := append([]string{fss3.cfg.DirFileName}, strings.Split(name, "/")...)
	for i := range tokens {
		err := fss3.Mkdir(strings.Join(tokens[:i+1], "/"), mode)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove removes the named file or directory.
// If directory is not empty, it returns an error.
func (fss3 *FSS3) Remove(name string) error {
	name = sanitizeName(name)
	dirName := name
	if name != fss3.cfg.DirFileName {
		dirName = name + "/" + fss3.cfg.DirFileName
	}

	_, err := fss3.statObject(name, nil)
	if err != nil {
		rspErr := errToRspErr(err)
		if rspErr.Code == "NoSuchKey" {
			_, dirErr := fss3.statObject(dirName, nil)
			if dirErr != nil {
				return minioErrToPathErr(err)
			}
			opts := listObjectsOptions{
				Recursive: false,
				Prefix:    name,
			}
			objs := make([]objectInfo, 0)
			for obj := range fss3.listObjects(&opts) {
				objs = append(objs, obj)
			}
			if len(objs) > 1 || (len(objs) != 1 && objs[0].Key != dirName) {
				return &fs.PathError{
					Op:   "remove",
					Path: name,
					Err:  ErrNotEmpty{name: name},
				}
			}
			dirErr = fss3.removeObject(dirName, nil)
			if dirErr != nil {
				return minioErrToPathErr(dirErr)
			}
			return nil
		}
		return minioErrToPathErr(err)
	}

	err = fss3.removeObject(name, nil)
	if err != nil {
		return minioErrToPathErr(err)
	}

	return nil
}

// RemoveAll removes path and any children it contains.
func (fss3 *FSS3) RemoveAll(path string) error {
	name := sanitizeName(path)
	prefix := name
	if name == fss3.cfg.DirFileName {
		prefix = ""
	} else {
		prefix = name[:len(name)-len(fss3.cfg.DirFileName)]
	}
	objsCh := make(chan objectInfo)

	go func() {
		defer close(objsCh)
		opts := listObjectsOptions{
			Recursive: true,
			Prefix:    prefix,
		}
		for obj := range fss3.listObjects(&opts) {
			if obj.Err != nil {
				log.Printf("warning: %s", obj.Err)
			}
			objsCh <- obj
		}
	}()

	errCh := fss3.removeObjects(objsCh, nil)
	for err := range errCh {
		return minioErrToPathErr(err.Err)
	}

	return nil
}

func (fss3 *FSS3) writeFrom(name string, r io.Reader, size int64, perm fs.FileMode) error {
	name = sanitizeName(name)
	parent := sanitizeName(filepath.Dir(name))
	err := fss3.MkdirAll(parent, fs.ModePerm)
	if err != nil {
		return err
	}

	opts := putObjectOptions{
		UserMetadata: map[string]string{
			"mode": fmt.Sprintf("%o", umask(fss3.cfg.Umask, perm)),
		},
		ContentType: guessContentType(name),
	}
	_, err = fss3.putObject(name, r, size, &opts)
	if err != nil {
		return minioErrToPathErr(err)
	}

	return nil
}

// WriteFile writes data to an object and creates any necessary parent.
// It creates the file if it doesn't exist.
func (fss3 *FSS3) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return fss3.writeFrom(name, bytes.NewReader(data), int64(len(data)), perm)
}

// WriteFrom writes the contents of reader to an object.
func (fss3 *FSS3) WriteFrom(name string, r io.Reader, perm fs.FileMode) error {
	return fss3.writeFrom(name, r, -1, perm)
}
