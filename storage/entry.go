package storage

import (
	"time"
)

type Entry struct {
	Ref   string
	Hash  string
	Ttl   int64 // Unix timestamp
	Files []File
}

func (entry Entry) IsExpired() bool {
	ttlTime := time.Unix(entry.Ttl, 0)
	return ttlTime.Before(time.Now())
}
