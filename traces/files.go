package traces

import (
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type readSeekCloser interface {
	io.ReadCloser
	io.Seeker
}

type gzipFile struct {
	r *gzip.Reader
	f *os.File
}

func newGzipFile(f *os.File) *gzipFile {
	r, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	return &gzipFile{
		r: r,
		f: f,
	}
}

func (f *gzipFile) Read(p []byte) (int, error) {
	return f.r.Read(p)
}

func (f *gzipFile) Seek(offset int64, whence int) (int64, error) {
	n, err := f.f.Seek(offset, whence)
	if err != nil {
		return n, err
	}
	f.r.Reset(f.f)
	return n, nil
}

func (f *gzipFile) Close() error {
	err1 := f.r.Close()
	err2 := f.f.Close()
	if err2 != nil {
		return err2
	}
	return err1
}

type bzip2File struct {
	r io.Reader
	f *os.File
}

func newBzip2File(f *os.File) *bzip2File {
	return &bzip2File{
		r: bzip2.NewReader(f),
		f: f,
	}
}

func (f *bzip2File) Read(p []byte) (int, error) {
	return f.r.Read(p)
}

func (f *bzip2File) Seek(offset int64, whence int) (int64, error) {
	n, err := f.f.Seek(offset, whence)
	if err != nil {
		return n, err
	}
	f.r = bzip2.NewReader(f.f)
	return n, nil
}

func (f *bzip2File) Close() error {
	return f.f.Close()
}

type filesReader struct {
	io.Reader
	files []readSeekCloser
}

func openFilesGlob(pattern string) (*filesReader, error) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%s not found", pattern)
	}
	return openFiles(files...)
}

func openFiles(files ...string) (*filesReader, error) {
	r := &filesReader{}
	r.files = make([]readSeekCloser, 0, len(files))
	readers := make([]io.Reader, 0, len(files))
	for _, name := range files {
		f, err := os.Open(name)
		if err != nil {
			r.Close()
			return nil, err
		}
		var rs readSeekCloser
		if strings.HasSuffix(name, ".gz") {
			rs = newGzipFile(f)
		} else if strings.HasSuffix(name, ".bz2") {
			rs = newBzip2File(f)
		} else {
			rs = f
		}
		r.files = append(r.files, rs)
		readers = append(readers, rs)
	}
	r.Reader = io.MultiReader(readers...)
	return r, nil
}

func (r *filesReader) Close() error {
	var err error
	for _, f := range r.files {
		e := f.Close()
		if err != nil && e != nil {
			err = e
		}
	}
	return err
}

func (r *filesReader) Reset() error {
	readers := make([]io.Reader, 0, len(r.files))
	for _, f := range r.files {
		_, err := f.Seek(0, 0)
		if err != nil {
			return err
		}
		readers = append(readers, f)
	}
	r.Reader = io.MultiReader(readers...)
	return nil
}
