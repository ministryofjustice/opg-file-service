package session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
)

type Session struct {
	AwsSession *session.Session
}

func NewSession() (*Session, error) {
	awsRegion := os.Getenv("AWS_REGION")

	if awsRegion == "" {
		awsRegion = "eu-west-1" // default region
	}

	sess, err := session.NewSession(&aws.Config{Region: &awsRegion})
	if err != nil {
		return nil, err
	}

	iamRole, isIamRoleSet := os.LookupEnv("AWS_IAM_ROLE")
	if isIamRoleSet {
		c := stscreds.NewCredentials(sess, iamRole)
		*sess.Config.Credentials = *c
	}

	return &Session{sess}, nil
}
