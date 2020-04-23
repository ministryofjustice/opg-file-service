package session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"opg-s3-zipper-service/internal"
	"os"
)

type Session struct {
	AwsSession *session.Session
}

func NewSession() (*Session, error) {
	awsRegion := internal.GetEnvVar("AWS_REGION", "eu-west-1")

	if awsRegion == "" {
		awsRegion = "eu-west-1" // default region
	}

	sess, err := session.NewSession(&aws.Config{Region: &awsRegion})
	if err != nil {
		return nil, err
	}

	if iamRole, ok := os.LookupEnv("AWS_IAM_ROLE"); ok {
		c := stscreds.NewCredentials(sess, iamRole)
		*sess.Config.Credentials = *c
	}

	return &Session{sess}, nil
}
