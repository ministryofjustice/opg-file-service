package storage

import (
	"strconv"
	"time"
)

type Entry struct {
	Ref   string
	Hash  string
	Ttl   int64  // Unix timestamp
	Files []File `json:"files"`
}

func (entry Entry) IsExpired() bool {
	ttlTime := time.Unix(entry.Ttl, 0)
	return ttlTime.Before(time.Now())
}

func (entry Entry) Validate() (bool, *ErrValidation) {
	var errs []ErrFieldValidation

	if entry.Ref == "" {
		errs = append(errs, ErrFieldValidation{
			Field:   "Ref",
			Message: "entry Ref cannot be blank",
		})
	}

	if entry.Hash == "" {
		errs = append(errs, ErrFieldValidation{
			Field:   "Hash",
			Message: "user Hash cannot be blank",
		})
	}

	if entry.Ttl == 0 {
		errs = append(errs, ErrFieldValidation{
			Field:   "Ttl",
			Message: "entry Ttl cannot be blank",
		})
	} else if entry.IsExpired() {
		errs = append(errs, ErrFieldValidation{
			Field:   "IsExpired",
			Message: "entry has expired",
		})
	}

	if len(entry.Files) == 0 {
		errs = append(errs, ErrFieldValidation{
			Field:   "Files",
			Message: "entry does not contain any Files",
		})
	}

	for _, file := range entry.Files {
		if ok, validationErr := file.Validate(); !ok {
			errs = append(errs, validationErr.Errors...)
		}
	}

	var err *ErrValidation
	if len(errs) > 0 {
		err = &ErrValidation{Errors: errs}
	}

	return len(errs) == 0, err
}

func (entry Entry) DeDupe() {
	type fileCounter struct {
		filename string
		count    int
	}

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
