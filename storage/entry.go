package storage

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"time"
)

type Entry struct {
	Ref   string
	Hash  string
	Ttl   int64 // Unix timestamp
	Files []File
}

func (entry Entry) IsHashValid(user *string) bool {
	salt := os.Getenv("USER_HASH_SALT")
	hash := md5.Sum([]byte(salt + *user))
	return hex.EncodeToString(hash[:]) == entry.Hash
}

func (entry Entry) IsExpired() bool {
	ttlTime := time.Unix(entry.Ttl, 0)
	return ttlTime.Before(time.Now())
}
