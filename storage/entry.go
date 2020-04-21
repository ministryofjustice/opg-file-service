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
	Ttl   time.Time
	Files []File
}

func (entry Entry) isHashValid(user *string) bool {
	salt := os.Getenv("USER_HASH_SALT")
	hash := md5.Sum([]byte(salt + *user))
	return hex.EncodeToString(hash[:]) == entry.Hash
}

func (entry Entry) isExpired() bool {
	return entry.Ttl.After(time.Now())
}
