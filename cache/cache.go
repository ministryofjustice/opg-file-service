package cache

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

type SecretsCache struct {
	env   string
	cache *secretcache.Cache
}

func applyAwsConfig(c *secretcache.Cache) {
	config := aws.Config{
		Credentials: credentials.NewEnvCredentials(),
	}

	endpoint := os.Getenv("SECRETS_MANAGER_ENDPOINT")

	if endpoint != "" {
		config.Endpoint = aws.String(endpoint)
	}

	sess, err := session.NewSession(&config)
	if err != nil {
		log.Fatal(err.Error())
	}

	c.Client = secretsmanager.New(sess)
}

func New() *SecretsCache {
	cache, _ := secretcache.New(applyAwsConfig)
	env := os.Getenv("ACCOUNT_NAME")
	return &SecretsCache{env, cache}
}

func (c *SecretsCache) GetSecretString(key string) (string, error) {
	secret, err := c.cache.GetSecretString(c.env + "/" + key)

	if err != nil {

	}

	return secret, err
}
