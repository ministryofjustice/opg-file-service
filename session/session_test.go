package session

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

func TestNewSession(t *testing.T) {
	tests := map[string]struct {
		region     string
		role       string
		wantRegion string
	}{
		"no_role":  {"us-west-1", "", "us-west-1"},
		"has_role": {"us-east-2", "some-iam-role", "us-east-2"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := NewSession(test.region, test.role)
			assert.Nil(t, err)
			assert.IsType(t, new(session.Session), got.AwsSession)
			assert.Equal(t, test.wantRegion, *got.AwsSession.Config.Region)
		})
	}
}
