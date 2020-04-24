package storage

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotFoundError_Error(t *testing.T) {
	err := NotFoundError{Ref: "testRef"}
	assert.Equal(t, "Could not find entry with reference: testRef", err.Error())
}
