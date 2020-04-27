package handlers

import (
	"github.com/stretchr/testify/assert"
	"log"
	"opg-file-service/dynamo"
	"opg-file-service/zipper"
	"testing"
)

func TestNewZipHandler(t *testing.T) {
	l := new(log.Logger)
	zh, err := NewZipHandler(l)
	assert.Nil(t, err)
	assert.IsType(t, ZipHandler{}, *zh)
	assert.Equal(t, l, zh.logger)
	assert.IsType(t, new(zipper.Zipper), zh.zipper)
	assert.IsType(t, new(dynamo.Repository), zh.repo)
}

func TestZipHandler_ServeHTTP(t *testing.T) {
}
