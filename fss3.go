package fss3

import (
	"context"
	"io"
	"io/fs"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type object = minio.Object
type objectInfo = minio.ObjectInfo
type uploadInfo = minio.UploadInfo
type removeObjectError = minio.RemoveObjectError
type putObjectOptions = minio.PutObjectOptions
type getObjectOptions = minio.GetObjectOptions
type statObjectOptions = minio.StatObjectOptions
type listObjectsOptions = minio.ListObjectsOptions
type removeObjectOptions = minio.RemoveObjectOptions
type removeObjectsOptions = minio.RemoveObjectsOptions

var dirFileName = "."

// FSS3 represents an opened bucket.
type FSS3 struct {
	client *minio.Client
	cfg    *Config
}

// New creates a new FSS3 object
func New(cfg Config) (*FSS3, error) {
	creds := credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  creds,
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if cfg.DirFileName == "" {
		cfg.DirFileName = "."
	}
	dirFileName = cfg.DirFileName
	fss3 := FSS3{
		client: client,
		cfg:    &cfg,
	}
	return &fss3, err
}

// FS returns a fs.FS from the FSS3 object
func (fss3 *FSS3) FS() fs.FS {
	return &FS{fss3}
}

// listObjects lists all objects at the given prefix
func (fss3 *FSS3) listObjects(opts *listObjectsOptions) <-chan objectInfo {
	if opts == nil {
		opts = &listObjectsOptions{}
	}
	return fss3.client.ListObjects(context.Background(), fss3.cfg.BucketName, *opts)
}

// getObject returns an Object for the given key
func (fss3 *FSS3) getObject(key string, opts *getObjectOptions) (*object, error) {
	if opts == nil {
		opts = &getObjectOptions{}
	}
	obj, err := fss3.client.GetObject(context.Background(), fss3.cfg.BucketName, key, *opts)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// statObject gets info about the object at the given key
func (fss3 *FSS3) statObject(key string, opts *statObjectOptions) (*objectInfo, error) {
	if opts == nil {
		opts = &statObjectOptions{}
	}
	stat, err := fss3.client.StatObject(context.Background(), fss3.cfg.BucketName, key, *opts)
	if err != nil {
		return nil, err
	}
	return &stat, nil
}

// PutObject uploads a file to the given key
func (fss3 *FSS3) putObject(key string, r io.Reader, size int64, opts *putObjectOptions) (*uploadInfo, error) {
	if opts == nil {
		opts = &putObjectOptions{}
	}
	ui, err := fss3.client.PutObject(context.Background(), fss3.cfg.BucketName, key, r, size, *opts)
	if err != nil {
		return nil, err
	}
	return &ui, nil
}

// RemoveObject removes a file for the given key
func (fss3 *FSS3) removeObject(key string, opts *removeObjectOptions) error {
	if opts == nil {
		opts = &removeObjectOptions{}
	}
	return fss3.client.RemoveObject(context.Background(), fss3.cfg.BucketName, key, *opts)
}

// RemoveObjects removes multiple files for the given object infos
func (fss3 *FSS3) removeObjects(objsCh <-chan objectInfo, opts *removeObjectsOptions) <-chan removeObjectError {
	if opts == nil {
		opts = &removeObjectsOptions{}
	}
	return fss3.client.RemoveObjects(context.Background(), fss3.cfg.BucketName, objsCh, *opts)
}
