package storage

import (
	"errors"
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
	assert.Equal(t, f.GetPathInZip(), fh.Name)
}

func TestFile_GetPathInZip(t *testing.T) {
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
		assert.Equal(t, test.want, file.GetPathInZip())
	}
}

func TestFile_Validate(t *testing.T) {
	tests := []struct {
		scenario  string
		file      *File
		wantValid bool
		wantErrs  []error
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
			[]error{
				errors.New("S3Path cannot be blank"),
				errors.New("FileName cannot be blank"),
			},
		},
	}

	for _, test := range tests {
		valid, errs := test.file.Validate()
		assert.Equal(t, test.wantValid, valid, test.scenario)
		assert.Equal(t, test.wantErrs, errs, test.scenario)
	}
}
