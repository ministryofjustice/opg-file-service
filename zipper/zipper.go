package zipper

import (
	"archive/zip"
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"net/http"
	"net/url"
	"opg-file-service/storage"
	"strings"
)

type ZipperInterface interface {
	Open(rw http.ResponseWriter)
	Close() error
	AddFile(ctx context.Context, f *storage.File) error
}

type Zipper struct {
	rw http.ResponseWriter
	zw ZipWriter
	s3 Downloader
}

func NewZipper(cfg *aws.Config) *Zipper {
	s3Client := s3.NewFromConfig(*cfg, func(u *s3.Options) {
		u.UsePathStyle = true
	})

	downloader := manager.NewDownloader(s3Client)
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

func (z *Zipper) AddFile(ctx context.Context, f *storage.File) error {
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

	_, err = z.s3.Download(ctx, fw, &input)
	if err != nil {
		return err
	}

	return nil
}
