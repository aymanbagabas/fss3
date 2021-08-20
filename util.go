package fss3

import (
	"io/fs"
	"mime"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
)

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, dirFileName, ".")
	name = strings.Trim(name, "/")
	name = filepath.Clean(name)
	if name == "." {
		return dirFileName
	}
	return name
}

func keyBaseName(key string) string {
	name := sanitizeName(key)
	return filepath.Base(name)
}

func errToRspErr(err error) minio.ErrorResponse {
	return minio.ToErrorResponse(err)
}

func minioErrToPathErr(err error) *fs.PathError {
	rspErr := errToRspErr(err)
	return &fs.PathError{
		Op:   "open",
		Path: rspErr.Key,
		Err:  err,
	}
}

func guessContentType(name string) string {
	contentType := mime.TypeByExtension(filepath.Ext(name))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return contentType
}

func umask(mask int, mode fs.FileMode) fs.FileMode {
	return mode - fs.FileMode(mask)
}
