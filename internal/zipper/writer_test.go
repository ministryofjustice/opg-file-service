package zipper

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFakeWriterAt_WriteAt(t *testing.T) {
	b := new(bytes.Buffer)
	w := sequentialWriterAt{b}

	_, err := w.WriteAt([]byte("Hello"), 0)
	assert.Nil(t, err)
	assert.Equal(t, "Hello", b.String())

	_, err = w.WriteAt([]byte(" "), 100)
	assert.Nil(t, err)
	assert.Equal(t, "Hello ", b.String())

	_, err = w.WriteAt([]byte("World"), 0)
	assert.Nil(t, err)
	assert.Equal(t, "Hello World", b.String())
}
