package zipper

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ministryofjustice/opg-file-service/internal/session"
	"github.com/ministryofjustice/opg-file-service/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewZipper(t *testing.T) {
	sess, _ := session.NewSession("test", "test")
	z := NewZipper(*sess, "endpoint")
	assert.Nil(t, z.rw)
	assert.Nil(t, z.zw)
	assert.NotNil(t, z.s3)
}

func TestZipper_Open(t *testing.T) {
	rr := httptest.NewRecorder()
	z := Zipper{}
	z.Open(rr)
	hm := rr.Result().Header

	assert.Equal(t, "application/zip", hm.Get("Content-Type"))
	assert.Equal(t, "attachment; filename=\"download.zip\"", hm.Get("Content-Disposition"))
	assert.Equal(t, rr, z.rw)
	assert.IsType(t, new(zip.Writer), z.zw)
}

func TestZipper_Close(t *testing.T) {
	m := new(MockZipWriter)
	e := errors.New("test")
	m.On("Close").Return(e).Once()

	z := Zipper{zw: m}
	err := z.Close()

	assert.Equal(t, e, err)
	assert.Nil(t, z.rw)
	assert.Nil(t, z.zw)
	m.AssertExpectations(t)
}

func TestZipper_AddFile(t *testing.T) {
	tests := map[string]struct {
		s3path            string
		createHeaderError error
		downloadError     error
		expectedS3Bucket  string
		expectedS3Key     string
		expectedError     error
	}{
		"missing_path": {
			expectedError: errors.New("missing S3 path"),
		},
		"invalid_scheme": {
			s3path:        "http://some/path",
			expectedError: errors.New("invalid S3 path: http://some/path"),
		},
		"invalid_path": {
			s3path:        "s3://file",
			expectedError: errors.New("invalid S3 path: s3://file"),
		},
		"correct": {
			s3path:           "s3://bucket/file",
			expectedS3Bucket: "bucket",
			expectedS3Key:    "file",
		},
		"error_with_zip_header": {
			s3path:            "s3://bucket/file",
			createHeaderError: errors.New("some problem with zip header"),
			expectedS3Bucket:  "bucket",
			expectedS3Key:     "file",
			expectedError:     errors.New("some problem with zip header"),
		},
		"error_downloading": {
			s3path:           "s3://bucket/file",
			downloadError:    errors.New("some problem downloading from S3"),
			expectedS3Bucket: "bucket",
			expectedS3Key:    "file",
			expectedError:    errors.New("some problem downloading from S3"),
		},
		"error_parsing_path": {
			s3path:           ":file",
			expectedS3Bucket: "bucket",
			expectedS3Key:    "file",
			expectedError:    errors.New("unable to parse S3 path: :file"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mz := new(MockZipWriter)
			md := new(MockDownloader)
			rr := httptest.NewRecorder()

			z := Zipper{rr, mz, md}
			f := storage.File{
				S3path:   test.s3path,
				FileName: "file",
				Folder:   "folder",
			}

			buf := new(bytes.Buffer)
			mz.On("CreateHeader", mock.AnythingOfType("*zip.FileHeader")).Return(buf, test.createHeaderError)

			s3input := s3.GetObjectInput{
				Bucket: aws.String(test.expectedS3Bucket),
				Key:    aws.String(test.expectedS3Key),
			}
			var options []func(*s3manager.Downloader)
			md.On("Download", FakeWriterAt{buf}, &s3input, options).Return(int64(0), test.downloadError)

			err := z.AddFile(&f)
			assert.Equal(t, test.expectedError, err)
		})
	}
}

type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error) {
	args := m.Called(w, input, options)
	return args.Get(0).(int64), args.Error(1)
}

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
