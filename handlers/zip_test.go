package handlers

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func getMockLogger() *log.Logger {
	logOutput := ""
	return log.New(bytes.NewBufferString(logOutput), "test", log.LstdFlags)
}

func TestNewZipHandler(t *testing.T) {
	zh := NewZipHandler(getMockLogger())
	assert.IsType(t, ZipHandler{}, *zh)
}
