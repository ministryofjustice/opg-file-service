package zipper

import (
	"archive/zip"
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go/types"
	"net/http"
	"net/url"
	"opg-file-service/session"
	"opg-file-service/storage"
	"os"
	"strconv"
	"strings"
	"github.com/thoas/go-funk"
)

type ZipperInterface interface {
	Open(rw http.ResponseWriter)
	Close() error
	AddFile(f *storage.File) error
}

type Zipper struct {
	rw http.ResponseWriter
	zw ZipWriter
	s3 Downloader
	filesAdded []fileCounter
}

type fileCounter struct {
	filename string
	count int
}

func NewZipper(sess session.Session) *Zipper {
	endpoint := os.Getenv("AWS_S3_ENDPOINT")
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

func (z *Zipper) getFileNameAndExtension(f *storage.File) (string, string) {
	bits := strings.Split(f.OutputFilePath, ".")
	extension := bits[len(bits)-1]
	theRest := bits[0:len(bits)-1]
	fileNameWithouExt := strings.Join(theRest, ".")
	return fileNameWithouExt, extension
}

func (z *Zipper) indexOf(arr []fileCounter, filename string) int {
	for i, fc := range arr {
		if fc.filename == filename {
			return i
		}
	}
	return -1
}

// Applies a file count if a file with the same filename has already been added to the zip stream
func (z *Zipper) applyFilenameCount(f *storage.File) {
	key := z.indexOf(z.filesAdded, f.OutputFilePath)

	if (key == -1) {
		fc := new(fileCounter)
		fc.count = 0
		fc.filename = f.OutputFilePath
		z.filesAdded[key] = *fc
		return
	}

	z.filesAdded[key].count++
	fileNameWithouExt, extension := z.getFileNameAndExtension(f)
	f.OutputFilePath = fileNameWithouExt + "(" + strconv.Itoa(z.filesAdded[key].count) + ")." + extension
	return
}

func (z *Zipper) AddFile(f *storage.File) error {
	if f.S3path == "" {
		return errors.New("missing S3 path")
	}

	if f.OutputFilePath == "" {
		return errors.New("missing OutputFilePath")
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

	z.applyFilenameCount(f)

	input := s3.GetObjectInput{
		Bucket: aws.String(u.Host),
		Key:    aws.String(strings.Trim(u.Path, "/")),
		ResponseContentDisposition: aws.String("attachment; filename =" + f.OutputFilePath),
	}

	_, err = z.s3.Download(fw, &input)
	if err != nil {
		return err
	}

	return nil
}
