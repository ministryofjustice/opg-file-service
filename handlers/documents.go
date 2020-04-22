package handlers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/nicholasjackson/env"
	"log"
	"net/http"
)

var iamRole = env.String("IAM_ROLE", true, "", "IAM Role for use with AWS-SDK")

// Documents is a http.Handler
type Documents struct {
	l *log.Logger
}

func NewDocuments(l *log.Logger) *Documents {
	return &Documents{l}
}

func (d *Documents) GetDocuments(rw http.ResponseWriter, r *http.Request) {

	//Get the reference from the request
	vars := mux.Vars(r)
	reference := vars["reference"]
	d.l.Println("Zip files for reference:", reference)

	// TODO: Get the files from DynamoDB
}

func getDocumentsFromS3(l *log.Logger) {
	// TODO: Actually pull documents from S3
	sess, err := session.NewSession()
	if err != nil {
		l.Println(err)
	}
	c := stscreds.NewCredentials(sess, *iamRole)
	awsConfig := aws.Config{Credentials: c, Region: aws.String("eu-west-1"), Endpoint: aws.String("http://localstack:8080")}
	svc := s3.New(sess, &awsConfig)

	// List buckets to check if auth is working TODO: Remove this once its working
	result, err := svc.ListBuckets(nil)
	if err != nil {
		l.Fatalln("Failed to list buckets")
	}

	for _, b := range result.Buckets {
		l.Println("Bucket name:", aws.StringValue(b.Name))
	}
}

func zipDocuments() {
	// TODO: Implement the zip functionality from s3zipper
}
