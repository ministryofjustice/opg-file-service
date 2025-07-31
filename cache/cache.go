package cache

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/v2/secretcache"
	"os"
)

type SecretsCache struct {
	env   string
	cache awsSecretsCache
}

type awsSecretsCache interface {
	GetSecretString(secretId string) (string, error)
}

func applyAwsConfig(cfg *aws.Config) func(c *secretcache.Cache) {
	return func(c *secretcache.Cache) {
		c.Client = secretsmanager.NewFromConfig(*cfg)
	}
}

func New(cfg *aws.Config) *SecretsCache {
	env := os.Getenv("ENVIRONMENT")
	cache, err := secretcache.New(applyAwsConfig(cfg))
	if err != nil {
		panic(err)
	}
	return &SecretsCache{env, cache}
}

func (c *SecretsCache) GetSecretString(key string) (string, error) {
	return c.cache.GetSecretString(c.env + "/" + key)
}
