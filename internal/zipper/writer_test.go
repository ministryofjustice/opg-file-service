package zipper

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFakeWriterAt_WriteAt(t *testing.T) {
	b := new(bytes.Buffer)
	w := FakeWriterAt{b}

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
