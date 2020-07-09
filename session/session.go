package session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

type Session struct {
	AwsSession *session.Session
}

func NewSession(region string, iamRole string) (*Session, error) {
	sess, err := session.NewSession(&aws.Config{Region: aws.String(region)})
	if err != nil {
		return nil, err
	}

	if iamRole != "" {
		sess.Config.Credentials = stscreds.NewCredentials(sess, iamRole)
	}

	return &Session{sess}, nil
}
