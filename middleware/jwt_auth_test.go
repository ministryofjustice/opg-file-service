package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Msg struct {
	Message string `json:"message"`
}

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

	if rw.Result().StatusCode == 403 {
		t.Fatalf("Status code should be 200 as a valid JWT token was passed")
	}

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
	if rw.Result().StatusCode == 200 {
		t.Fatalf("Status code should be 403 as a invalid JWT token was passed")
	}
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
	if rw.Result().StatusCode == 200 {
		t.Fatalf("Status code should be 403 as no JWT token was passed")
	}
}
