package fss3

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/minio/minio-go/v7"
)

var (
	cfg = Config{
		AccessKeyID:     os.Getenv("ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("SECRET_ACCESS_KEY"),
		Endpoint:        os.Getenv("ENDPOINT"),
		BucketName:      os.Getenv("BUCKET_NAME"),
		Region:          os.Getenv("REGION"),
		DirFileName:     os.Getenv("DIR_FILE_NAME"),
		UseSSL:          false,
	}
	fss3 *FSS3 = nil
)

func TestMain(m *testing.M) {
	s3, err := New(cfg)
	if err != nil {
		panic(err)
	}
	fss3 = s3
	code := m.Run()
	os.Exit(code)
}

func TestCreate(t *testing.T) {
	f, err := fss3.Create("testfile")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
}

func TestWriteFile(t *testing.T) {
	data := []byte("hello world")
	err := fss3.WriteFile("testfile", data, 0644)
	if err != nil {
		t.Error(err)
	}
}

func TestOpen(t *testing.T) {
	f, err := fss3.Open("testfile")
	if err != nil {
		t.Error(err)
	}
	stats, err := f.Stat()
	if err != nil {
		t.Error(err)
	}
	t.Log(stats)
	var size int64 = 11
	name := "testfile"
	mode := os.FileMode(0644)
	if stats.Size() != size {
		t.Errorf("stats size error, expect %d, but %d", size, stats.Size())
	}
	if stats.Name() != name {
		t.Errorf("stats name error, expect %s, but %s", name, stats.Name())
	}
	if stats.Mode() != mode {
		t.Errorf("stats mode error, expect %d, but %d", mode, stats.Mode())
	}
}

func TestReadFile(t *testing.T) {
	b, err := fss3.ReadFile("testfile")
	if err != nil {
		t.Error(err)
	}
	str := string(b)
	if str != "hello world" {
		t.Error("read file error")
	}
}

func TestOpenNotFound(t *testing.T) {
	_, err := fss3.Open("file/not/found")
	rsp := minio.ToErrorResponse((err.(*fs.PathError)).Err)
	if err != nil && rsp.Code != "NoSuchKey" {
		t.Errorf("expect NoSuchKey error, but got '%+v'", rsp.Code)
	}
}

func TestMkdirAll(t *testing.T) {
	err := fss3.MkdirAll("a/b/c", 0777)
	if err != nil {
		t.Error(err)
	}
	a, err := fss3.Stat("a")
	if err != nil {
		t.Error(err)
	}
	b, err := fss3.Stat("a/b")
	if err != nil {
		t.Error(err)
	}
	c, err := fss3.Stat("a/b/c")
	if err != nil {
		t.Error(err)
	}
	if !a.IsDir() {
		t.Error("a is not a dir")
	}
	if !b.IsDir() {
		t.Error("a/b is not dir")
	}
	if !c.IsDir() {
		t.Error("a/b/c is not dir")
	}
}

func TestFileCreate(t *testing.T) {
	f, err := fss3.Create("a/file")
	if err != nil {
		t.Error(err)
	}
	f.Close()
}

func TestFileOpenWrite(t *testing.T) {
	f, err := fss3.Open("a/file")
	if err != nil {
		t.Errorf("open error: %s", err)
	}
	defer f.Close()
	f.WriteString("hello go")
}

func TestWriteTo(t *testing.T) {
	f, err := fss3.Open("a/file")
	if err != nil {
		t.Errorf("open error: %s", err)
	}
	defer f.Close()
	out, err := os.Create("out")
	if err != nil {
		t.Errorf("create error: %s", err)
	}
	defer out.Close()
	defer os.Remove("out")
	n, err := f.WriteTo(out)
	if err != nil {
		t.Errorf("write to error: %s", err)
	}
	if n != 8 {
		t.Error("write to error")
	}
	fout, err := os.Open("out")
	if err != nil {
		t.Errorf("open error: %s", err)
	}
	defer fout.Close()
	b, err := ioutil.ReadAll(fout)
	if err != nil {
		t.Errorf("read error: %s", err)
	}
	if string(b) != "hello go" {
		t.Error("read file error")
	}
}

func TestReadDir(t *testing.T) {
	err := fss3.Mkdir("one", 0777)
	if err != nil {
		t.Fatal(err)
	}
	defer fss3.RemoveAll("one")
	_, err = fss3.Create("one/file")
	if err != nil {
		t.Fatal(err)
	}
	_, err = fss3.Create("one/b")
	if err != nil {
		t.Fatal(err)
	}
	f, err := fss3.ReadDir("one")
	if err != nil {
		t.Errorf("read dir error, %s", err)
	}
	if len(f) != 2 {
		t.Errorf("read dir error, expect 2, but got %d", len(f))
	}
	for _, v := range f {
		if v.Name() != "file" && v.Name() != "b" {
			t.Errorf("read dir error, expect file or b, but got %s", v.Name())
		}
	}
}

func TestWalkDir(t *testing.T) {
	root := fss3.cfg.DirFileName
	_, err := fss3.Create("testfile")
	if err != nil {
		t.Fatal(err)
	}
	err = fss3.MkdirAll("a/b/c", 0777)
	if err != nil {
		t.Fatal(err)
	}
	_, err = fss3.Create("a/file")
	if err != nil {
		t.Fatal(err)
	}
	defer fss3.RemoveAll("a")
	defer fss3.Remove("testfile")
	expect := []string{
		root,
		path.Join(root, "testfile"),
		path.Join(root, "a"),
		path.Join(root, "a/file"),
		path.Join(root, "a/b"),
		path.Join(root, "a/b/c"),
		path.Join(root, "a/b/c"),
	}
	fs.WalkDir(fss3.FS(), root, func(path string, info fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !contains(expect, path) {
			t.Errorf("walk dir error, expect %s in %v", path, expect)
		}
		return nil
	})
}

func TestRemove(t *testing.T) {
	_, err := fss3.Create("a/file")
	if err != nil {
		t.Fatal(err)
	}
	err = fss3.Remove("a/file")
	if err != nil {
		t.Errorf("remove error: %s", err)
	}
	_, err = fss3.Stat("a/file")
	if err == nil {
		t.Error("remove error, expect not nil, but nil")
	}
}

func TestRemoveAll(t *testing.T) {
	err := fss3.RemoveAll("")
	if err != nil {
		t.Errorf("remove all error: %s", err)
	}
	_, err = fss3.Stat(fss3.cfg.DirFileName)
	if err == nil {
		t.Errorf("remove all error, expect not nil, but nil: %s", err)
	}
}

func TestChmod(t *testing.T) {
	f, err := fss3.Create("chmod")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer fss3.Remove("chmod")
	err = fss3.Chmod("chmod", 0777)
	if err != nil {
		t.Fatalf("chmod error: %s", err)
	}
	info, err := fss3.Stat("chmod")
	if err != nil {
		t.Fatalf("stat error: %s", err)
	}
	if info.Mode() != 0777 {
		t.Fatalf("chmod error, expect 0777, but %o", info.Mode())
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
