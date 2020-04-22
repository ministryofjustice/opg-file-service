package zipper

import "io"

type FakeWriterAt struct {
	w io.Writer
}

func (fw FakeWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we force sequential downloads
	return fw.w.Write(p)
}
