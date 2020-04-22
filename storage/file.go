package storage

import (
	"regexp"
	"strings"
)

type File struct {
	S3path   string
	FileName string
	Folder   string
}

func (f *File) getSafeFileName() string {
	regex := regexp.MustCompile(`[#<>:"/|?*\\]`)
	safe := regex.ReplaceAllString(f.FileName, "")
	if safe == "" {
		return "file" // fallback
	}
	return safe
}

func (f *File) GetZipPath() string {
	path := ""

	if f.Folder != "" {
		path += f.Folder
		if !strings.HasSuffix(path, "/") {
			path += "/"
		}
	}

	return path + f.getSafeFileName()
}
