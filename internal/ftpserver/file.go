package ftpserver

import (
	"io"
	"os"
)

type FileInfo struct {
	fullName string
	os.FileInfo
}

func (f *FileInfo) Name() string {
	return f.fullName
}

type File struct {
	*os.File
	io.ReadWriter
}

func (f *File) Write(b []byte) (n int, err error) {
	return f.ReadWriter.Write(b)
}

func (f *File) Read(b []byte) (n int, err error) {
	return f.ReadWriter.Read(b)
}

func (f *File) ReadFrom(r io.Reader) (n int64, err error) {
	return io.Copy(f.ReadWriter, r)
}
