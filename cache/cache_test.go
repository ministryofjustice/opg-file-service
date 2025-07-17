package cache

import (
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-secretsmanager-caching-go/v2/secretcache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
)

type MockAwsSecretsCache struct {
	mock.Mock
}

func (m *MockAwsSecretsCache) GetSecretString(secretId string) (string, error) {
	args := m.Called(secretId)
	return args.Get(0).(string), args.Error(1)
}

func TestNew(t *testing.T) {
	oldEnv := os.Getenv("ENVIRONMENT")
	_ = os.Setenv("ENVIRONMENT", "test_env")

	sc := New(aws.NewConfig())
	assert.IsType(t, new(SecretsCache), sc)
	assert.Equal(t, "test_env", sc.env)
	assert.IsType(t, new(secretcache.Cache), sc.cache)

	_ = os.Setenv("ENVIRONMENT", oldEnv)
}

func TestSecretsCache_GetSecretString(t *testing.T) {
	tests := []struct {
		scenario       string
		env            string
		secretKey      string
		returnedSecret string
		returnedErr    error
	}{
		{
			scenario:       "Secret retrieved successfully",
			env:            "test_env",
			secretKey:      "test_key",
			returnedSecret: "test_secret",
			returnedErr:    nil,
		},
		{
			scenario:       "AwsSecretsCache returns an error",
			env:            "test_env",
			secretKey:      "test_key",
			returnedSecret: "",
			returnedErr:    errors.New("test error"),
		},
		{
			scenario:       "No ENVIRONMENT defined",
			env:            "",
			secretKey:      "test_key",
			returnedSecret: "",
			returnedErr:    errors.New("test error"),
		},
	}
	for _, test := range tests {
		msc := new(MockAwsSecretsCache)
		msc.On("GetSecretString", test.env+"/"+test.secretKey).Return(test.returnedSecret, test.returnedErr).Times(1)

		sc := SecretsCache{
			env:   test.env,
			cache: msc,
		}
		secret, err := sc.GetSecretString(test.secretKey)

		assert.Equal(t, test.returnedSecret, secret, test.scenario)
		assert.Equal(t, test.returnedErr, err, test.scenario)
	}
}
