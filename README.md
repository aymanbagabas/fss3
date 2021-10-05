# FSS3

[![Build Status](https://github.com/aymanbagabas/fss3/actions/workflows/ci.yml/badge.svg)](https://github.com/aymanbagabas/fss3/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/github/go-mod/go-version/aymanbagabas/fss3.svg)](https://github.com/aymanbagabas/fss3)
[![Go Reference](https://pkg.go.dev/badge/github.com/aymanbagabas/fss3.svg)](https://pkg.go.dev/github.com/aymanbagabas/fss3)
[![Go Report Card](https://goreportcard.com/badge/github.com/aymanbagabas/fss3)](https://goreportcard.com/report/github.com/aymanbagabas/fss3)
[![GitHub release](https://img.shields.io/github/release/aymanbagabas/fss3.svg)](https://GitHub.com/aymanbagabas/fss3/releases/)


FSS3 is an S3 filesystem abstraction layer for Golang that implements most of [fs](https://pkg.go.dev/io/fs), and [io](https://pkg.go.dev/io) interfaces, and [os](https://pkg.go.dev/os) functions. It is based on [minio-go](https://github.com/minio/minio-go) which makes it compatible with any S3 compliant service.

## Download

```
go get github.com/aymanbagabas/fss3
```

## Quick Start

```go
package main

import "github.com/aymanbagabas/fss3"

func main() {
	cfg := fss3.Config{
		AccessKeyID:     "AWS_ACCESS_KEY_ID",
		SecretAccessKey: "AWS_SECRET_ACCESS_KEY",
		Endpoint:        "ENDPOINT",
		UseSSL:          true,
		BucketName:      "MY_BUCKET_NAME",
		Region:          "REGION",
		DirFileName:     "_", // special directory file name that stores directory metadata
		Umask:           0, // Don't set umask
	}
	s3, err := fss3.New(cfg)
	if err != nil {
		panic(err)
	}

	err = s3.Mkdir("Newfolder", 0777)
	if err != nil {
		panic(err)
	}

	data := []byte{"hello world"}
	err = s3.WriteFile("Newfolder/myfile.txt", data, 0644)
	if err != nil {
		panic(err)
	}

	err = s3.RemoveAll("Newfolder")
	if err != nil {
		panic(err)
	}
}
```

## License

This library is distributed under the [MIT License](https://opensource.org/licenses/MIT), see [LICENSE](https://github.com/aymanbagabas/fss3/blob/master/LICENSE) for more information.
