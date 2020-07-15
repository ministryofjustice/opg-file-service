package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFile_GetZipFileHeader(t *testing.T) {
	f := File{
		S3path:   "s3://bucket/file",
		FileName: "file",
		Folder:   "folder",
	}
	fh := f.GetZipFileHeader()
	assert.Equal(t, f.GetRelativePath(), fh.Name)
}

func TestFile_GetRelativePath(t *testing.T) {
	tests := []struct {
		fileName string
		folder   string
		want     string
	}{
		{"file.test", "", "file.test"},
		{"file.test", "folder", "folder/file.test"},
		{"file.test", "/folder/", "folder/file.test"},
		{"", "", "undefined"},
		{"", "folder", "folder/undefined"},
		{"", "", "undefined"},
		{"", "test/", "test/undefined"},
		{"", "", "undefined"},
		{`[#<>:"/|?*\]`, `[#<>:"/|?*\]`, "undefined"},
	}

	file := File{}
	for _, test := range tests {
		file.FileName = test.fileName
		file.Folder = test.folder
		assert.Equal(t, test.want, file.GetRelativePath())
	}
}

func TestFile_GetFileNameAndExtension(t *testing.T) {
	tests := []struct {
		scenario      string
		filename      string
		wantFilename  string
		wantExtension string
	}{
		{"Filename with one dot", "filename.txt", "filename", "txt"},
		{"Filename with mutiple dots", "filename.something.txt", "filename.something", "txt"},
		{"Filename with no file extension", "filename", "filename", ""},
	}

	file := File{}
	for _, test := range tests {
		file.FileName = test.filename
		filename, extension := file.GetFileNameAndExtension()
		assert.Equal(t, test.wantFilename, filename, test.scenario)
		assert.Equal(t, test.wantExtension, extension, test.scenario)
	}
}

func TestFile_Validate(t *testing.T) {
	tests := []struct {
		scenario  string
		file      *File
		wantValid bool
		wantErr   *ErrValidation
	}{
		{
			"No errors returned for valid file",
			&File{
				S3path:   "s3://files/file",
				FileName: "file",
				Folder:   "",
			},
			true,
			nil,
		},
		{
			"Blank S3Path and FileName",
			&File{},
			false,
			&ErrValidation{
				Errors: []ErrFieldValidation{
					{Field: "S3Path", Message: "S3Path cannot be blank"},
					{Field: "FileName", Message: "FileName cannot be blank"},
				},
			},
		},
	}

	for _, test := range tests {
		valid, err := test.file.Validate()
		assert.Equal(t, test.wantValid, valid, test.scenario)
		assert.Equal(t, test.wantErr, err, test.scenario)
	}
}
