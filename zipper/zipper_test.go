package zipper

import (
	"archive/zip"
	"bytes"
	"errors"
	"net/http/httptest"
	"opg-file-service/session"
	"opg-file-service/storage"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	tests := []struct {
		s3path            string
		createHeaderError error
		downloadError     error
		expectedS3Bucket  string
		expectedS3Key     string
		expectedError     error
	}{
		{
			"",
			nil,
			nil,
			"",
			"",
			errors.New("missing S3 path"),
		},
		{
			"http://some/path",
			nil,
			nil,
			"",
			"",
			errors.New("invalid S3 path: http://some/path"),
		},
		{
			"s3://file",
			nil,
			nil,
			"",
			"",
			errors.New("invalid S3 path: s3://file"),
		},
		{
			"s3://bucket/file",
			nil,
			nil,
			"bucket",
			"file",
			nil,
		},
		{
			"s3://bucket/file",
			errors.New("some problem with zip header"),
			nil,
			"bucket",
			"file",
			errors.New("some problem with zip header"),
		},
		{
			"s3://bucket/file",
			nil,
			errors.New("some problem downloading from S3"),
			"bucket",
			"file",
			errors.New("some problem downloading from S3"),
		},
		{
			":file",
			nil,
			nil,
			"bucket",
			"file",
			errors.New("unable to parse S3 path: :file"),
		},
	}

	for _, test := range tests {
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
	}
}
