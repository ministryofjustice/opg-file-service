package cache

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
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

	sc := New()
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
		msc.On("GetSecretString", test.env + "/" + test.secretKey).Return(test.returnedSecret, test.returnedErr).Times(1)

		sc := SecretsCache{
			env:   test.env,
			cache: msc,
		}
		secret, err := sc.GetSecretString(test.secretKey)

		assert.Equal(t, test.returnedSecret, secret, test.scenario)
		assert.Equal(t, test.returnedErr, err, test.scenario)
	}
}

func TestApplyAwsConfig(t *testing.T) {
	tests := []struct {
		scenario   string
		endpoint   string
		region     string
		wantRegion string
		role       string
	}{
		{
			scenario:   "Blank AWS region",
			endpoint:   "test-endpoint",
			region:     "",
			wantRegion: "eu-west-1",
			role:       "test-role",
		},
		{
			scenario:   "Custom AWS region",
			endpoint:   "test-endpoint",
			region:     "test-region",
			wantRegion: "test-region",
			role:       "test-role",
		},
		{
			scenario:   "Blank Iam Role",
			endpoint:   "test-endpoint",
			region:     "test-region",
			wantRegion: "test-region",
			role:       "",
		},
	}
	for _, test := range tests {
		oldEndpoint := os.Getenv("SECRETS_MANAGER_ENDPOINT")
		oldRegion := os.Getenv("AWS_REGION")
		oldIamRole := os.Getenv("AWS_IAM_ROLE")

		_ = os.Setenv("SECRETS_MANAGER_ENDPOINT", test.endpoint)
		if test.region == "" {
			_ = os.Unsetenv("AWS_REGION")
		} else {
			_ = os.Setenv("AWS_REGION", test.region)
		}
		if test.role == "" {
			_ = os.Unsetenv("AWS_IAM_ROLE")
		} else {
			_ = os.Setenv("AWS_IAM_ROLE", test.role)
		}

		c := new(secretcache.Cache)

		applyAwsConfig(c)

		cl := c.Client.(*secretsmanager.SecretsManager)

		assert.Equal(t, "https://"+test.endpoint, cl.Endpoint, test.scenario)
		assert.Equal(t, test.wantRegion, *cl.Config.Region, test.scenario)

		_ = os.Setenv("AWS_SECRETS_MANAGER_ENDPOINT", oldEndpoint)
		_ = os.Setenv("AWS_REGION", oldRegion)

		if oldIamRole == "" {
			_ = os.Unsetenv("AWS_IAM_ROLE")
		} else {
			_ = os.Setenv("AWS_IAM_ROLE", oldIamRole)
		}
	}
}