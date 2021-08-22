package fss3

// Config is the configuration for the FSS3 client.
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
