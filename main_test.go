package main

import (
	"archive/zip"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ministryofjustice/opg-file-service/dynamo"
	"github.com/ministryofjustice/opg-file-service/session"
	"github.com/ministryofjustice/opg-file-service/storage"
	"github.com/stretchr/testify/suite"
)

type EndToEndTestSuite struct {
	suite.Suite
	bucket     *string
	sess       *session.Session
	s3         *s3.S3
	s3uploader *s3manager.Uploader
	repo       *dynamo.Repository
	testEntry  *storage.Entry
}

func (suite *EndToEndTestSuite) SetupSuite() {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "eu-west-1"
	}
	endpoint := os.Getenv("AWS_DYNAMODB_ENDPOINT")
	table := os.Getenv("AWS_DYNAMODB_TABLE_NAME")
	if table == "" {
		table = "zip-requests"
	}

	suite.sess, _ = session.NewSession(region, os.Getenv("AWS_IAM_ROLE"))
	suite.bucket = aws.String("files")
	s3sess := *suite.sess.AwsSession
	s3sess.Config.Endpoint = aws.String(os.Getenv("AWS_S3_ENDPOINT"))
	s3sess.Config.S3ForcePathStyle = aws.Bool(true)
	suite.s3 = s3.New(&s3sess)
	suite.s3uploader = s3manager.NewUploader(&s3sess)
	suite.repo = dynamo.NewRepository(*suite.sess, new(log.Logger), endpoint, table)

	// create an S3 bucket
	suite.s3.CreateBucket(&s3.CreateBucketInput{
		Bucket: suite.bucket,
	})
	suite.s3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: suite.bucket,
	})

	// define fixtures
	suite.testEntry = &storage.Entry{
		Ref:  "test",
		Hash: "d1a046e6300ea9a75cc4f9eda85e8442c3e9913b8eeb4ed0895896571e479a99", // hash is for Test.McTestFace@mail.com
		Ttl:  9999999999,
		Files: []storage.File{
			{
				S3path:   "s3://files/file1",
				FileName: "file1",
			},
			{
				S3path:   "s3://files/file2",
				FileName: "file2",
			},
			{
				S3path:   "s3://files/file3",
				FileName: "file3",
				Folder:   "folder",
			},
		},
	}

	// start the app
	go main()
}

func (suite *EndToEndTestSuite) TearDownSuite() {
	suite.ClearFixtures()
}

func (suite *EndToEndTestSuite) SetupTest() {
	suite.ClearFixtures()

	// add files to bucket
	for _, file := range suite.testEntry.Files {
		suite.s3uploader.Upload(&s3manager.UploadInput{
			Bucket: suite.bucket,
			Key:    &file.FileName,
			Body:   strings.NewReader("contents of " + file.FileName),
		})
	}

	// add a "zip request" to DynamoDB
	if ok, errs := suite.testEntry.Validate(); !ok {
		suite.Failf("Invalid entry: %e", "", errs)
	}
	suite.repo.Add(suite.testEntry)
}

func (suite *EndToEndTestSuite) ClearFixtures() {
	// empty the bucket
	iter := s3manager.NewDeleteListIterator(suite.s3, &s3.ListObjectsInput{
		Bucket: suite.bucket,
	})
	s3manager.NewBatchDeleteWithClient(suite.s3).Delete(aws.BackgroundContext(), iter)

	// delete entry from DynamoDB
	suite.repo.Delete(suite.testEntry)
}

func (suite *EndToEndTestSuite) GetUrl(path string) string {
	return "http://localhost:8000/" + os.Getenv("PATH_PREFIX") + path
}

func (suite *EndToEndTestSuite) TestHealthCheck() {
	resp, err := http.Get(suite.GetUrl("/health-check"))
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *EndToEndTestSuite) TestZip() {
	// download zip file
	req, _ := http.NewRequest("GET", suite.GetUrl("/zip/test"), nil)
	req.Header.Set("Authorization", "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6OTk5OTk5OTk5OSwic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.8HtN6aTAnE2YFI9rJD8drzqgrXPkyUbwRRJymcPSmHk")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		suite.Fail("", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	// store file on disk because zip.Reader expects an io.ReaderAt
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		suite.Fail("", err)
	}
	err = ioutil.WriteFile("test.zip", bodyBytes, 0644)
	if err != nil {
		suite.Fail("", err)
	}

	// extract archive and make assertions
	rc, err := zip.OpenReader("test.zip")
	if err != nil {
		suite.Fail("", err)
	}
	defer rc.Close()
	defer os.Remove("test.zip")

	want := make(map[string]string)
	got := make(map[string]string)

	// map filename to file contents
	for _, file := range suite.testEntry.Files {
		fn := file.FileName
		if file.Folder != "" {
			fn = file.Folder + "/" + fn
		}
		want[fn] = "contents of " + file.FileName
	}

	// loop through files in zip and do the same mapping
	for _, file := range rc.File {
		if file.FileInfo().IsDir() {
			continue
		}
		fo, _ := file.Open()
		fb, _ := ioutil.ReadAll(fo)
		got[file.Name] = string(fb)
		fo.Close()

		// assert that file's modified date is within 5 seconds from now
		suite.InDelta(time.Now().Unix(), file.Modified.Unix(), 5)
	}

	suite.Equal(want, got)
}

func TestEndToEnd(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}
