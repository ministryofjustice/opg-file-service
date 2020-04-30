package zipper

import (
	"archive/zip"
	"io"
)

// allows us to mock zip.Writer in our tests
type ZipWriter interface {
	Close() error
	CreateHeader(fh *zip.FileHeader) (io.Writer, error)
}
