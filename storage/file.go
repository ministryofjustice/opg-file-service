package storage

import (
	"archive/zip"
	"regexp"
	"strings"
	"time"
)

type File struct {
	S3path   string `json:"s3path"`
	FileName string `json:"filename"`
	Folder   string `json:"folder"`
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

func (f *File) Validate() (bool, *ErrValidation) {
	var errs []ErrFieldValidation

	if f.S3path == "" {
		errs = append(errs, ErrFieldValidation{
			Field:   "S3Path",
			Message: "S3Path cannot be blank",
		})
	}

	if f.FileName == "" {
		errs = append(errs, ErrFieldValidation{
			Field:   "FileName",
			Message: "FileName cannot be blank",
		})
	}

	var err *ErrValidation
	if len(errs) > 0 {
		err = &ErrValidation{Errors: errs}
	}

	return len(errs) == 0, err
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
