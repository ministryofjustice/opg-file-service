package storage

import (
	"errors"
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

func (entry Entry) Validate() (bool, []error) {
	var errs []error

	if entry.Ref == "" {
		errs = append(errs, errors.New("entry Ref cannot be blank"))
	}

	if entry.Hash == "" {
		errs = append(errs, errors.New("user Hash cannot be blank"))
	}

	if entry.Ttl == 0 {
		errs = append(errs, errors.New("entry Ttl cannot be blank"))
	} else if entry.IsExpired() {
		errs = append(errs, errors.New("entry has expired"))
	}

	if len(entry.Files) == 0 {
		errs = append(errs, errors.New("entry does not contain any Files"))
	}

	for _, file := range entry.Files {
		if ok, fileErrs := file.Validate(); !ok {
			errs = append(errs, fileErrs...)
		}
	}

	return len(errs) == 0, errs
}
