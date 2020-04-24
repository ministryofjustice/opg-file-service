package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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