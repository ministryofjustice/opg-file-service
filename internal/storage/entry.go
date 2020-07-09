package storage

import (
	"errors"
	"strconv"
	"time"
)

type Entry struct {
	Ref   string
	Hash  string
	Ttl   int64 // Unix timestamp
	Files []File
}

type fileCounter struct {
	filename string
	count    int
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

func (entry Entry) DeDupe() {
	var filesAdded []fileCounter

	for i, file := range entry.Files {
		foundMatch := false
		for j, fileAdded := range filesAdded {
			if fileAdded.filename != file.GetRelativePath() {
				continue
			}

			fileNameWithoutExt, extension := file.GetFileNameAndExtension()

			if len(extension) > 0 {
				extension = "." + extension
			}

			entry.Files[i].FileName = fileNameWithoutExt + " (" + strconv.Itoa(filesAdded[j].count) + ")" + extension
			foundMatch = true
			filesAdded[j].count++
		}

		if !foundMatch {
			filesAdded = append(filesAdded, fileCounter{file.GetRelativePath(), 1})
		}
	}
}
