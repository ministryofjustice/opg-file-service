package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"io"
	"net"
	"net/http"
	"opg-file-service/handlers"
	"opg-file-service/storage"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/suite"
)

type EndToEndTestSuite struct {
	suite.Suite
	bucket     *string
	s3         *s3.Client
	s3uploader *manager.Uploader
	testEntry  *storage.Entry
	authHeader string
}

func (suite *EndToEndTestSuite) SetupSuite() {
	ctx := context.Background()
	_ = os.Setenv("ENVIRONMENT", "local")
	_ = os.Setenv("AWS_ACCESS_KEY_ID", "test")
	_ = os.Setenv("AWS_SECRET_ACCESS_KEY", "test")
	_ = os.Setenv("AWS_ENDPOINT", "http://localhost:4566")
	suite.bucket = aws.String("files")
	config, _ := awsConfig()
	suite.s3 = s3.NewFromConfig(*config)
	suite.s3uploader = manager.NewUploader(suite.s3)

	_, err := suite.s3.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: suite.bucket,
	})
	if err != nil {
		panic(err)
	}

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
		_, _ = suite.s3uploader.Upload(ctx, &s3.PutObjectInput{
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
		_ = conn.Close()
		return
	}

	suite.Fail(fmt.Sprintf("Unable to start file service server after %d attempts", retries))
}

func (suite *EndToEndTestSuite) TearDownSuite() {
	// empty the bucket
	_, err := suite.s3.DeleteBucket(context.Background(), &s3.DeleteBucketInput{Bucket: suite.bucket})
	if err != nil {
		suite.Fail("", err)
	}
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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	suite.Equal(http.StatusOK, resp.StatusCode)

	suite.Contains(respBody.Link, "/zip/")
	suite.Greater(len(respBody.Link), len("/zip/"))

	// store file on disk because zip.Reader expects an io.ReaderAt0
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.Fail("", err)
	}
	err = os.WriteFile("/tmp/test.zip", bodyBytes, 0644)
	if err != nil {
		suite.Fail("", err)
	}

	// extract archive and make assertions
	rc, err := zip.OpenReader("/tmp/test.zip")
	if err != nil {
		suite.Fail("", err)
	}
	defer func(rc *zip.ReadCloser) {
		_ = rc.Close()
	}(rc)
	defer func() {
		_ = os.Remove("/tmp/test.zip")
	}()

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
		fb, _ := io.ReadAll(fo)
		got[file.Name] = string(fb)
		_ = fo.Close()

		// assert that file's modified date is within 5 seconds from now
		suite.InDelta(time.Now().Unix(), file.Modified.Unix(), 5)
	}

	suite.Equal(want, got)
}

func TestEndToEnd(t *testing.T) {
	suite.Run(t, new(EndToEndTestSuite))
}
