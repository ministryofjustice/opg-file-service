package zipper

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/mock"
	"io"
)

type MockDownloader struct {
	mock.Mock
}

func (m MockDownloader) Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error) {
	args := m.Called(w, input, options)
	return args.Get(0).(int64), args.Error(1)
}
