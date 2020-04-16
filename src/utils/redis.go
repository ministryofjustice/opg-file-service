package utils

import (
	"encoding/json"
	"errors"
	"github.com/gomodule/redigo/redis"
	"github.com/nicholasjackson/env"
	"log"
	"time"
)

var redisUrl = env.String("REDIS_URL", true, "127.0.0.1:6379", "Redis URL to redis connection")
var redisPool *redis.Pool

type RedisFile struct {
	FileName string
	Folder   string
	S3Path   string
}

func InitRedisPool(l *log.Logger) {
	redisPool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 1 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", *redisUrl)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) (err error) {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err = c.Do("PING")
			if err != nil {
				l.Fatal("Error connecting to redis")
			}
			return
		},
	}
}

// TODO: Ideally we need a unit test for this
func GetFilesFromRedis(ref string) (files []*RedisFile, err error) {

	// Testing - enable to test. Remove later.
	if 1 == 0 && ref == "test" {
		files = append(files, &RedisFile{FileName: "test.zip", Folder: "", S3Path: "test/test.zip"}) // Edit and dplicate line to test
		return
	}

	redis := redisPool.Get()
	defer redis.Close()

	// Get the value from Redis
	result, err := redis.Do("GET", "zip:"+ref)
	if err != nil || result == nil {
		err = errors.New("Access Denied (sorry your link has timed out)")
		return
	}

	// Convert to bytes
	var resultByte []byte
	var ok bool
	if resultByte, ok = result.([]byte); !ok {
		err = errors.New("Error converting data stream to bytes")
		return
	}

	// Decode JSON
	err = json.Unmarshal(resultByte, &files)
	if err != nil {
		err = errors.New("Error decoding json: " + string(resultByte))
	}

	return
}

