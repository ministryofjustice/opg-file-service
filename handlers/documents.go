package handlers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"net/http"
)

// Documents is a http.Handler
type Documents struct {
	l *log.Logger
}

func NewDocuments(l *log.Logger) *Documents {
	return &Documents{l}
}

func (d *Documents) GetDocuments(rw http.ResponseWriter, r *http.Request) {
	d.l.Println("Handle GET zip-documents")

	sess, err := session.NewSession()
	if err != nil {
		d.l.Println(err)
	}
	creds := stscreds.NewCredentials(sess, "arn:aws:iam::288342028542:role/operator")
	awsConfig := aws.Config{Credentials: creds, Region: aws.String("eu-west-1")}

	s3Svc := s3.New(sess, &awsConfig)

}

