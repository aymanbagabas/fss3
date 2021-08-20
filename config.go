package fss3

type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Endpoint        string
	Region          string
	UseSSL          bool
	BucketName      string
	Umask           int
	DirFileName     string
}
