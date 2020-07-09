package handlers

import (
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteJSONError(t *testing.T) {
	tests := map[string]struct {
		error string
		desc  string
		code  int
		want  string
	}{
		"error": {
			error: "test",
			desc:  "error msg",
			code:  500,
			want:  `{"error":"test","error_description":"error msg"}` + "\n",
		},
		"success": {
			code: 200,
			want: `{"error":"","error_description":""}` + "\n",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSONError(w, test.error, test.desc, test.code)

			resp := w.Result()
			b, _ := ioutil.ReadAll(resp.Body)

			assert.Equal(t, test.want, string(b))
			assert.Equal(t, test.code, resp.StatusCode)
		})
	}
}
