package storage

import (
	"archive/zip"
	"errors"
	"regexp"
	"strings"
)

type File struct {
	S3path         string
	FileName       string
	Folder         string
	OutputFilePath string
}

func (f *File) GetZipFileHeader() *zip.FileHeader {
	// We have to set a special flag so zip files recognize utf file names
	// See http://stackoverflow.com/questions/30026083/creating-a-zip-archive-with-unicode-filenames-using-gos-archive-zip
	return &zip.FileHeader{
		Name:   f.GetPathInZip(),
		Method: zip.Deflate,
		Flags:  0x800,
	}
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
		file = "undefined" // default filename
	}

	return path + file
}

func (f *File) Validate() (bool, []error) {
	var errs []error

	if f.S3path == "" {
		errs = append(errs, errors.New("S3Path cannot be blank"))
	}

	if f.FileName == "" {
		errs = append(errs, errors.New("FileName cannot be blank"))
	}

	return len(errs) == 0, errs
}
