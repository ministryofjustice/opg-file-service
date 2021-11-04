package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"opg-file-service/handlers"
	"opg-file-service/session"
	"opg-file-service/storage"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/suite"
)

type EndToEndTestSuite struct {
	suite.Suite
	bucket     *string
	sess       *session.Session
	s3         *s3.S3
	s3uploader *s3manager.Uploader
	testEntry  *storage.Entry
	authHeader string
}

func (suite *EndToEndTestSuite) SetupSuite() {
	os.Setenv("ENVIRONMENT", "local")
	suite.sess, _ = session.NewSession()
	suite.bucket = aws.String("files")
	s3sess := *suite.sess.AwsSession
	s3sess.Config.Endpoint = aws.String(os.Getenv("AWS_S3_ENDPOINT"))
	s3sess.Config.S3ForcePathStyle = aws.Bool(true)
	suite.s3 = s3.New(&s3sess)
	suite.s3uploader = s3manager.NewUploader(&s3sess)

	// create an S3 bucket
	suite.s3.CreateBucket(&s3.CreateBucketInput{
		Bucket: suite.bucket,
	})
	suite.s3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: suite.bucket,
	})

	// define fixtures
	suite.testEntry = &storage.Entry{
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

	suite.authHeader = "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6MzAwMDAwMDAwMCwic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.T1ufbp8mDZBGp84BqLnC2Vb366aVJfrZl_XaGeX4SH8"

	// add files to bucket
	for _, file := range suite.testEntry.Files {
		suite.s3uploader.Upload(&s3manager.UploadInput{
			Bucket: suite.bucket,
			Key:    &file.FileName,
			Body:   strings.NewReader("contents of " + file.FileName),
		})
	}

	// start the app
	go main()

	// wait up to 5 seconds for the server to start
	retries := 5
	for i := 1; i <= retries; i++ {
		conn, err := net.DialTimeout("tcp", "localhost:8000", time.Second)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		conn.Close()
		return
	}

	suite.Fail(fmt.Sprintf("Unable to start file service server after %d attempts", retries))
}

func (suite *EndToEndTestSuite) TearDownSuite() {
	// empty the bucket
	iter := s3manager.NewDeleteListIterator(suite.s3, &s3.ListObjectsInput{
		Bucket: suite.bucket,
	})
	s3manager.NewBatchDeleteWithClient(suite.s3).Delete(aws.BackgroundContext(), iter)
}

func (suite *EndToEndTestSuite) GetUrl(path string) string {
	return "http://localhost:8000" + os.Getenv("PATH_PREFIX") + path
}

func (suite *EndToEndTestSuite) TestHealthCheck() {
	resp, err := http.Get(suite.GetUrl("/health-check"))
	suite.Nil(err)
	suite.Equal(http.StatusOK, resp.StatusCode)
}

func (suite *EndToEndTestSuite) TestZip() {
	client := new(http.Client)

	// create a new zip request
	jsonBody, _ := json.Marshal(suite.testEntry)

	fmt.Println(string(jsonBody))

	reqBody := bytes.NewReader(jsonBody)
	req, _ := http.NewRequest(http.MethodPost, suite.GetUrl("/zip/request"), reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", suite.authHeader)

	resp, err := client.Do(req)
	if err != nil {
		suite.Fail("", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusCreated, resp.StatusCode)

	var respBody handlers.ZipRequestResponseBody
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		suite.Fail("", err)
	}

	// download zip file
	req, _ = http.NewRequest("GET", suite.GetUrl(respBody.Link), nil)
	req.Header.Set("Authorization", suite.authHeader)
	resp, err = client.Do(req)
	if err != nil {
		suite.Fail("", err)
	}
	defer resp.Body.Close()

	suite.Equal(http.StatusOK, resp.StatusCode)

	suite.Contains(respBody.Link, "/zip/")
	suite.Greater(len(respBody.Link), len("/zip/"))

	// store file on disk because zip.Reader expects an io.ReaderAt0
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
