package storage

import (
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
