package zipper

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ministryofjustice/opg-file-service/session"
	"github.com/ministryofjustice/opg-file-service/storage"
)

type Downloader interface {
	Download(w io.WriterAt, input *s3.GetObjectInput, options ...func(*s3manager.Downloader)) (n int64, err error)
}

type ZipWriter interface {
	io.Closer
	CreateHeader(fh *zip.FileHeader) (io.Writer, error)
}

type Zipper struct {
	rw http.ResponseWriter
	zw ZipWriter
	s3 Downloader
}

func NewZipper(sess session.Session, endpoint string) *Zipper {
	sess.AwsSession.Config.Endpoint = &endpoint
	sess.AwsSession.Config.S3ForcePathStyle = aws.Bool(true)

	downloader := s3manager.NewDownloader(sess.AwsSession)
	downloader.Concurrency = 1

	return &Zipper{
		s3: downloader,
	}
}

func (z *Zipper) Open(rw http.ResponseWriter) {
	z.rw = rw
	z.zw = zip.NewWriter(rw)
	z.rw.Header().Add("Content-Disposition", "attachment; filename=\"download.zip\"")
	z.rw.Header().Add("Content-Type", "application/zip")
}

func (z *Zipper) Close() error {
	err := z.zw.Close()
	z.rw = nil
	z.zw = nil
	return err
}

func (z *Zipper) AddFile(f *storage.File) error {
	if f.S3path == "" {
		return errors.New("missing S3 path")
	}

	u, err := url.Parse(f.S3path)
	if err != nil {
		return errors.New("unable to parse S3 path: " + f.S3path)
	}

	if u.Scheme != "s3" || u.Host == "" || u.Path == "" {
		return errors.New("invalid S3 path: " + f.S3path)
	}

	w, err := z.zw.CreateHeader(f.GetZipFileHeader())
	if err != nil {
		return err
	}
	fw := FakeWriterAt{w} // wrap our io.Writer in a fake io.WriterAt, as S3 requires a io.WriterAt

	input := s3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(strings.Trim(u.Path, "/")),
	}

	_, err = z.s3.Download(fw, &input)
	if err != nil {
		return err
	}

	return nil
}
