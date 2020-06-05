package storage

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestEntry_IsExpired(t *testing.T) {
	tests := []struct {
		ttl  int64
		want bool
	}{
		{time.Now().Add(time.Minute).Unix(), false},
		{time.Now().Add(-time.Minute).Unix(), true},
	}

	entry := Entry{}
	for _, test := range tests {
		entry.Ttl = test.ttl
		assert.Equal(t, test.want, entry.IsExpired())
	}
}

func TestEntry_Validate(t *testing.T) {
	tests := []struct {
		scenario  string
		entry     *Entry
		wantValid bool
		wantErrs  []error
	}{
		{
			"No errors returned for valid entry",
			&Entry{
				Ref:  "test",
				Hash: "user",
				Ttl:  9999999999,
				Files: []File{
					{
						S3path:   "s3://files/file",
						FileName: "file",
					},
				},
			},
			true,
			nil,
		},
		{
			"Validate blank fields",
			&Entry{},
			false,
			[]error{
				errors.New("entry Ref cannot be blank"),
				errors.New("user Hash cannot be blank"),
				errors.New("entry Ttl cannot be blank"),
				errors.New("entry does not contain any Files"),
			},
		},
		{
			"Validate expired entry",
			&Entry{
				Ref:  "test",
				Hash: "test",
				Ttl:  1,
				Files: []File{
					{
						S3path:   "s3://files/file",
						FileName: "file",
					},
				},
			},
			false,
			[]error{
				errors.New("entry has expired"),
			},
		},
		{
			"Errors include File validations",
			&Entry{
				Ref:  "test",
				Hash: "user",
				Ttl:  9999999999,
				Files: []File{
					{},
				},
			},
			false,
			nil,
		},
	}

	for _, test := range tests {
		if len(test.entry.Files) > 0 {
			// append file validation errors to expected errors
			_, fileErrs := test.entry.Files[0].Validate()
			test.wantErrs = append(test.wantErrs, fileErrs...)
		}
		valid, errs := test.entry.Validate()
		assert.Equal(t, test.wantValid, valid, test.scenario)
		assert.Equal(t, test.wantErrs, errs, test.scenario)
	}
}

func TestEntry_DeDupe(t *testing.T) {
	tests := []struct {
		scenario string
		filesBefore []File
		filesAfter []File
	}{
		{
			"Files with no duplicates",
			[]File{
				{FileName: "file1.txt",Folder: ""},
				{FileName: "file2.csv",Folder: ""},
				{FileName: "file3.pdf",Folder: ""},
				{FileName: "file4.xls",Folder: ""},
			},
			[]File{
				{FileName: "file1.txt",Folder: ""},
				{FileName: "file2.csv",Folder: ""},
				{FileName: "file3.pdf",Folder: ""},
				{FileName: "file4.xls",Folder: ""},
			},
		},
		{
			"Files with duplicates",
			[]File{
				{FileName: "file1.txt",Folder: ""},
				{FileName: "file1.txt",Folder: ""},
				{FileName: "file2.pdf",Folder: ""},
				{FileName: "file2.pdf",Folder: "some-directory"},
				{FileName: "file3.pdf",Folder: "some-directory"},
				{FileName: "file3.pdf",Folder: "some-directory"},
				{FileName: "file3.pdf",Folder: "some-directory"},
				{FileName: "file4.pdf",Folder: "some-directory"},
				{FileName: "file4.pdf",Folder: "some-directory"},
				{FileName: "file4.pdf",Folder: "some-directory/another-directory"},
				{FileName: "file5",Folder: "some-directory/another-directory/more-directory"},
				{FileName: "file5",Folder: "some-directory/another-directory/more-directory"},
			},
			[]File{
				{FileName: "file1.txt",Folder: ""},
				{FileName: "file1 (1).txt",Folder: ""},
				{FileName: "file2.pdf",Folder: ""},
				{FileName: "file2.pdf",Folder: "some-directory"},
				{FileName: "file3.pdf",Folder: "some-directory"},
				{FileName: "file3 (1).pdf",Folder: "some-directory"},
				{FileName: "file3 (2).pdf",Folder: "some-directory"},
				{FileName: "file4.pdf",Folder: "some-directory"},
				{FileName: "file4 (1).pdf",Folder: "some-directory"},
				{FileName: "file4.pdf",Folder: "some-directory/another-directory"},
				{FileName: "file5",Folder: "some-directory/another-directory/more-directory"},
				{FileName: "file5 (1)",Folder: "some-directory/another-directory/more-directory"},
			},
		},
	}

	entry := Entry{}

	for _, test := range tests {
		entry.Files = test.filesBefore
		entry.DeDupe()
		assert.Equal(t, test.filesAfter, entry.Files, test.scenario)
	}
}
