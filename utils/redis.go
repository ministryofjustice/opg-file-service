package utils

import (
	"github.com/go-redis/redis/v7"
	"log"
)


type Redis struct {
	l *log.Logger
}

func initRedis()  {

	client := redis.NewClient(&redis.Options{
		Addr:               "localhost:6379",
		Password:           "",
		DB:                 0,
	})

	ping, err := client.Ping().Result()
	if err != nil {
		d.l.Println(err)
	}
	d.l.Println(ping)
}
