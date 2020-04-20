package session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"os"
)

type Session struct {
	AwsSession *session.Session
}

func NewSession(l *log.Logger) *Session {
	iamRole := os.Getenv("AWS_IAM_ROLE")
	awsRegion := os.Getenv("AWS_REGION")

	if awsRegion == "" {
		awsRegion = "eu-west-1" // default region
	}

	sess, err := session.NewSession(&aws.Config{Region: &awsRegion})
	if err != nil {
		l.Println(err)
	}

	c := stscreds.NewCredentials(sess, iamRole)
	*sess.Config.Credentials = *c

	return &Session{sess}
}
