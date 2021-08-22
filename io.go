package fss3

import (
	"bytes"
	"io"
	"io/ioutil"
)

// Write writes len(p) bytes from p to the object at key.
func (f *File) Write(p []byte) (int, error) {
	buf := bytes.NewBuffer(p)
	ui, err := f.fs.fss3.putObject(f.fileInfo.info.Key, buf, int64(buf.Len()), nil)
	if err != nil {
		return 0, err
	}
	return int(ui.Size), nil
}

// WriteString writes a string to the object at key.
func (f *File) WriteString(s string) (int, error) {
	return f.Write([]byte(s))
}

// WriteTo writes the object data to w until there's no more data to write
func (f *File) WriteTo(w io.Writer) (int64, error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return 0, err
	}
	n, err := w.Write(data)
	if err != nil {
		return int64(n), err
	}
	return int64(n), nil
}
