package zipper

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/mock"
	"io"
)

type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) Download(ctx context.Context, w io.WriterAt, input *s3.GetObjectInput, options ...func(*manager.Downloader)) (n int64, err error) {
	args := m.Called(w, input, options)
	return args.Get(0).(int64), args.Error(1)
}
