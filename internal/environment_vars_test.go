package internal

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func str(s string) *string {
	return &s
}

func TestGetEnvVar(t *testing.T) {
	tests := []struct {
		env  string
		val  *string
		def  string
		want string
	}{
		{"TEST", str("test"), "test2", "test"},
		{"TEST", str(""), "test2", ""},
		{"TEST", nil, "test2", "test2"},
		{"", nil, "test2", "test2"},
	}

	for _, test := range tests {
		os.Unsetenv(test.env)
		if test.val != nil {
			os.Setenv(test.env, *test.val)
		}
		actual := GetEnvVar(test.env, test.def)
		assert.Equal(t, test.want, actual)
	}
}
