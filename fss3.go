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
type copySrcOptions = minio.CopySrcOptions
type copyDestOptions = minio.CopyDestOptions

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
	return fss3.client.GetObject(context.Background(), fss3.cfg.BucketName, key, *opts)
}

// statObject gets info about the object at the given key
func (fss3 *FSS3) statObject(key string, opts *statObjectOptions) (objectInfo, error) {
	if opts == nil {
		opts = &statObjectOptions{}
	}
	return fss3.client.StatObject(context.Background(), fss3.cfg.BucketName, key, *opts)
}

// putObject uploads a file to the given key
func (fss3 *FSS3) putObject(key string, r io.Reader, size int64, opts *putObjectOptions) (uploadInfo, error) {
	if opts == nil {
		opts = &putObjectOptions{}
	}
	return fss3.client.PutObject(context.Background(), fss3.cfg.BucketName, key, r, size, *opts)
}

// removeObject removes a file for the given key
func (fss3 *FSS3) removeObject(key string, opts *removeObjectOptions) error {
	if opts == nil {
		opts = &removeObjectOptions{}
	}
	return fss3.client.RemoveObject(context.Background(), fss3.cfg.BucketName, key, *opts)
}

// removeObjects removes multiple files for the given object infos
func (fss3 *FSS3) removeObjects(objsCh <-chan objectInfo, opts *removeObjectsOptions) <-chan removeObjectError {
	if opts == nil {
		opts = &removeObjectsOptions{}
	}
	return fss3.client.RemoveObjects(context.Background(), fss3.cfg.BucketName, objsCh, *opts)
}

// copyObject copies a file from src to dst
func (fss3 *FSS3) copyObject(srcKey, dstKey string, src *copySrcOptions, dst *copyDestOptions) (uploadInfo, error) {
	if src == nil {
		src = &copySrcOptions{
			Bucket: fss3.cfg.BucketName,
			Object: srcKey,
		}
	}
	if src.Bucket == "" {
		src.Bucket = fss3.cfg.BucketName
	}
	if src.Object == "" {
		src.Object = srcKey
	}
	if dst == nil {
		dst = &copyDestOptions{
			Bucket: fss3.cfg.BucketName,
			Object: dstKey,
		}
	}
	if dst.Bucket == "" {
		dst.Bucket = fss3.cfg.BucketName
	}
	if dst.Object == "" {
		dst.Object = dstKey
	}
	return fss3.client.CopyObject(context.Background(), *dst, *src)
}
