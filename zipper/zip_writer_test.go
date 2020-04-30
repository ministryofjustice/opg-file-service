package zipper

import (
	"archive/zip"
	"github.com/stretchr/testify/mock"
	"io"
)

type MockZipWriter struct {
	mock.Mock
}

func (m *MockZipWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockZipWriter) CreateHeader(fh *zip.FileHeader) (io.Writer, error) {
	args := m.Called(fh)
	return args.Get(0).(io.Writer), args.Error(1)
}
