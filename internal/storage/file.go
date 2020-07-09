package storage

import (
	"archive/zip"
	"errors"
	"regexp"
	"strings"
	"time"
)

var safePathRegex = regexp.MustCompile(`[#\[\]<>:"/|?*\\]`)

type File struct {
	S3path   string
	FileName string
	Folder   string
}

func (f *File) GetZipFileHeader() *zip.FileHeader {
	loc, _ := time.LoadLocation("Europe/London")

	// We have to set a special flag so zip files recognize utf file names
	// See http://stackoverflow.com/questions/30026083/creating-a-zip-archive-with-unicode-filenames-using-gos-archive-zip
	return &zip.FileHeader{
		Name:     f.GetRelativePath(),
		Method:   zip.Deflate,
		Flags:    0x800,
		Modified: time.Now().In(loc),
	}
}

func (f *File) GetRelativePath() string {
	path := ""

	if f.Folder != "" {
		folder := safePathRegex.ReplaceAllString(f.Folder, "")
		if folder != "" {
			path += folder
			if !strings.HasSuffix(path, "/") {
				path += "/"
			}
		}
	}

	file := safePathRegex.ReplaceAllString(f.FileName, "")
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

func (f *File) GetFileNameAndExtension() (string, string) {
	bits := strings.Split(f.FileName, ".")
	extension := ""
	fileNameWithoutExt := f.FileName
	if len(bits) > 1 {
		extension = bits[len(bits)-1]
		theRest := bits[0 : len(bits)-1]
		fileNameWithoutExt = strings.Join(theRest, ".")
	}
	return fileNameWithoutExt, extension
}
