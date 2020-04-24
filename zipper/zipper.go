package zipper

import (
	"archive/zip"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"net/http"
	"net/url"
	"opg-file-service/session"
	"opg-file-service/storage"
	"os"
	"strings"
)

type Zipper struct {
	rw http.ResponseWriter
	zw *zip.Writer
	s3 *s3manager.Downloader
}

func NewZipper(sess session.Session, rw http.ResponseWriter) *Zipper {
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
	sess.AwsSession.Config.Endpoint = &endpoint
	sess.AwsSession.Config.S3ForcePathStyle = aws.Bool(true)

	downloader := s3manager.NewDownloader(sess.AwsSession)
	downloader.Concurrency = 1

	return &Zipper{
		rw: rw,
		zw: zip.NewWriter(rw),
		s3: downloader,
	}
}

func (z *Zipper) Open() {
	z.rw.Header().Add("Content-Disposition", "attachment; filename=\"download.zip\"")
	z.rw.Header().Add("Content-Type", "application/zip")
}

func (z *Zipper) Close() error {
	return z.zw.Close()
}

func (z *Zipper) AddFile(f *storage.File) error {
	if f.S3path == "" {
		return errors.New("missing S3 path for file")
	}

	// We have to set a special flag so zip files recognize utf file names
	// See http://stackoverflow.com/questions/30026083/creating-a-zip-archive-with-unicode-filenames-using-gos-archive-zip
	fh := &zip.FileHeader{
		Name:   f.GetPathInZip(),
		Method: zip.Deflate,
		Flags:  0x800,
	}

	w, _ := z.zw.CreateHeader(fh)
	fw := FakeWriterAt{w} // wrap our io.Writer in a fake io.WriterAt, as S3 requires a io.WriterAt

	u, err := url.Parse(f.S3path)
	if err != nil {
		return err
	}

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
