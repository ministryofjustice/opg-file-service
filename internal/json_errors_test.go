package internal

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http/httptest"
	"testing"
)

func TestWriteJSONError(t *testing.T) {
	tests := []struct {
		error string
		descr string
		code  int
		want  string
	}{
		{"test", "error msg", 500, `{"error":"test","error_description":"error msg"}` + "\n"},
		{"", "", 200, `{"error":"","error_description":""}` + "\n"},
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()
		WriteJSONError(rr, test.error, test.descr, test.code)

		r := rr.Result()
		b, _ := io.ReadAll(r.Body)

		assert.Equal(t, test.want, string(b))
		assert.Equal(t, test.code, r.StatusCode)
	}
}
