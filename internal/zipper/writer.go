package zipper

import "io"

type sequentialWriterAt struct {
	w io.Writer
}

func (fw sequentialWriterAt) WriteAt(p []byte, offset int64) (n int, err error) {
	// ignore 'offset' because we force sequential downloads
	return fw.w.Write(p)
}
