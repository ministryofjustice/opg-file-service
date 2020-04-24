package storage

import (
	"fmt"
	"regexp"
	"strings"
)

type File struct {
	S3path   string
	FileName string
	Folder   string
}

func (f *File) GetPathInZip() string {
	// regex for getting a safe filename and folder
	regex := regexp.MustCompile(`[#\[\]<>:"/|?*\\]`)

	path := ""

	if f.Folder != "" {
		folder := regex.ReplaceAllString(f.Folder, "")
		if folder != "" {
			path += folder
			if !strings.HasSuffix(path, "/") {
				path += "/"
			}
		}
	}

	file := regex.ReplaceAllString(f.FileName, "")
	if file == "" {
		file = fmt.Sprint("undefined") // default filename
	}

	return path + file
}
