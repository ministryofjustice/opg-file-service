package cache

import (
	"opg-file-service/session"
	"os"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

type Cacheable interface {
	GetSecretString(key string) (string, error)
}

type SecretsCache struct {
	env   string
	cache *secretcache.Cache
}

func applyAwsConfig(session *session.Session) func(c *secretcache.Cache) {
	return func(c *secretcache.Cache) {
		c.Client = secretsmanager.New(session.AwsSession)
	}
}

func New() *SecretsCache {
	session, _ := session.NewSession()
	endpoint := os.Getenv("SECRETS_MANAGER_ENDPOINT")
	session.AwsSession.Config.Endpoint = &endpoint
	cache, _ := secretcache.New(applyAwsConfig(session))
	env := os.Getenv("ENVIRONMENT")
	return &SecretsCache{env, cache}
}

func (c *SecretsCache) GetSecretString(key string) (string, error) {
	return c.cache.GetSecretString(c.env + "/" + key)
}
