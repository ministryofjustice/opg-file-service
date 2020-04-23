package middleware

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJwtVerify(t *testing.T) {
	token := "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6OTk5OTk5OTk5OSwic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.8HtN6aTAnE2YFI9rJD8drzqgrXPkyUbwRRJymcPSmHk"
	req, err := http.NewRequest("GET", "/jwt", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	rw := httptest.NewRecorder()
	handler := JwtVerify(testHandler)
	handler.ServeHTTP(rw, req)

	assert.NotEqual(t, 403,  rw.Result().StatusCode, "Status Code should be 200")
}

func TestJwtVerifyInvalidToken(t *testing.T) {
	token := "Bearer NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6MTU4NzA1MjkxNywic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.f0oM4fSH_b1Xi5zEF0VK-t5uhpVidk5HY1O0EGR4SQQ"

	req, err := http.NewRequest("GET", "/jwt", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	rw := httptest.NewRecorder()
	handler := JwtVerify(testHandler)
	handler.ServeHTTP(rw, req)
	fmt.Println(rw.Result())

	assert.NotEqual(t, 200, rw.Result().StatusCode, "Status Code should be 403")
}

func TestJwtVerifyNoJwtToken(t *testing.T) {
	req, err := http.NewRequest("GET", "/jwt", nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", "")

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	rw := httptest.NewRecorder()
	handler := JwtVerify(testHandler)
	handler.ServeHTTP(rw, req)
	fmt.Println(rw.Result())

	assert.NotEqual(t, 200, rw.Result().StatusCode, "Status Code should be 403")
}

func TestJwtVerifyWrongSigningMethod(t *testing.T) {
	token := "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.POstGetfAytaZS82wHcjoTyoqhMyxXiWdR7Nn7A29DNSl0EiXLdwJ6xC6AfgZWF1bOsS_TuYI3OG85AmiExREkrS6tDfTQ2B3WXlrr-wp5AokiRbz3_oB4OxG-W9KcEEbDRcZc0nH3L7LzYptiy1PtAylQGxHTWZXtGz4ht0bAecBgmpdgXMguEIcoqPJ1n3pIWk_dUZegpqx0Lka21H6XxUTxiy8OcaarA8zdnPUnV6AmNP3ecFawIFYdvJB_cm-GvpCSbr8G8y_Mllj8f4x9nBH8pQux89_6gUY618iYv7tuPWBFfEbLxtF2pZS6YC1aSfLQxeNe8djT9YjpvRZA"

	req, err := http.NewRequest("GET", "/jwt", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	rw := httptest.NewRecorder()
	handler := JwtVerify(testHandler)
	handler.ServeHTTP(rw, req)
	fmt.Println(rw.Result())

	assert.NotEqual(t, 200, rw.Result().StatusCode, "Status Code should be 403")
}

func TestJwtVerifyExpiredToken(t *testing.T) {
	token := "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1ODcwNTIzMTcsImV4cCI6MTU4NzA1MjMxNywic2Vzc2lvbi1kYXRhIjoiVGVzdC5NY1Rlc3RGYWNlQG1haWwuY29tIn0.OuafGwOMHkXrFiQFrog8-zR14hxRwFkq5SeWXgvKi2o"
	req, err := http.NewRequest("GET", "/jwt", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", token)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	rw := httptest.NewRecorder()
	handler := JwtVerify(testHandler)
	handler.ServeHTTP(rw, req)

	assert.NotEqual(t, 200, rw.Result().StatusCode, "Status Code should be 403")
}