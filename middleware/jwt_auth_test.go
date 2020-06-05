package middleware

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJwtVerify(t *testing.T) {
	tests := []struct {
		scenario     string
		token        string
		expectedCode int
	}{
		{
			"Valid token",
			"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6OTk5OTk5OTk5OSwic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.8HtN6aTAnE2YFI9rJD8drzqgrXPkyUbwRRJymcPSmHk",
			200,
		},
		{
			"Invalid token",
			"NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6MTU4NzA1MjkxNywic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.f0oM4fSH_b1Xi5zEF0VK-t5uhpVidk5HY1O0EGR4SQQ",
			401,
		},
		{
			"No token",
			"",
			401,
		},
		{
			"Wrong signing method",
			"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.POstGetfAytaZS82wHcjoTyoqhMyxXiWdR7Nn7A29DNSl0EiXLdwJ6xC6AfgZWF1bOsS_TuYI3OG85AmiExREkrS6tDfTQ2B3WXlrr-wp5AokiRbz3_oB4OxG-W9KcEEbDRcZc0nH3L7LzYptiy1PtAylQGxHTWZXtGz4ht0bAecBgmpdgXMguEIcoqPJ1n3pIWk_dUZegpqx0Lka21H6XxUTxiy8OcaarA8zdnPUnV6AmNP3ecFawIFYdvJB_cm-GvpCSbr8G8y_Mllj8f4x9nBH8pQux89_6gUY618iYv7tuPWBFfEbLxtF2pZS6YC1aSfLQxeNe8djT9YjpvRZA",
			401,
		},
		{
			"Expired token",
			"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6MTU4NzA1MjMxNywic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.OuafGwOMHkXrFiQFrog8-zR14hxRwFkq5SeWXgvKi2o",
			401,
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest("GET", "/jwt", nil)
		if err != nil {
			t.Fatal(err)
		}

		if test.token != "" {
			req.Header.Set("Authorization", "Bearer "+test.token)
		}
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		rw := httptest.NewRecorder()
		handler := JwtVerify(testHandler)
		handler.ServeHTTP(rw, req)

		assert.Equal(t, test.expectedCode, rw.Result().StatusCode, test.scenario)
	}
}

func TestHashEmail(t *testing.T) {
	assert.Equal(
		t,
		"d1a046e6300ea9a75cc4f9eda85e8442c3e9913b8eeb4ed0895896571e479a99",
		hashEmail("Test.McTestFace@mail.com"),
	)
}
