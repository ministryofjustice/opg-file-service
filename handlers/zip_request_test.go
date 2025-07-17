package handlers

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"opg-file-service/middleware"
	"opg-file-service/storage"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestZipRequestHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		scenario       string
		reqBody        string
		repoAddCalls   int
		repoAddErr     error
		wantCode       int
		wantInResponse string
	}{
		{
			scenario:       "Invalid JSON request",
			reqBody:        "not a JSON",
			repoAddCalls:   0,
			repoAddErr:     nil,
			wantCode:       http.StatusBadRequest,
			wantInResponse: "Invalid JSON request",
		},
		{
			scenario:       "Empty JSON request",
			reqBody:        `{}`,
			repoAddCalls:   0,
			repoAddErr:     nil,
			wantCode:       http.StatusBadRequest,
			wantInResponse: "Missing field 'files'",
		},
		{
			scenario:       "Empty Files array in JSON request",
			reqBody:        `{"files":[]}`,
			repoAddCalls:   0,
			repoAddErr:     nil,
			wantCode:       http.StatusBadRequest,
			wantInResponse: "entry does not contain any Files",
		},
		{
			scenario:       "Missing fileName in JSON request",
			reqBody:        `{"files":[{"s3path":"s3://test/path"}]}`,
			repoAddCalls:   0,
			repoAddErr:     nil,
			wantCode:       http.StatusBadRequest,
			wantInResponse: "FileName cannot be blank",
		},
		{
			scenario:       "Missing s3path in JSON request",
			reqBody:        `{"files":[{"fileName":"file.test"}]}`,
			repoAddCalls:   0,
			repoAddErr:     nil,
			wantCode:       http.StatusBadRequest,
			wantInResponse: "S3Path cannot be blank",
		},
		{
			scenario:       "Unable to save new zip request in DynamoDB",
			reqBody:        `{"files":[{"s3path":"s3://test/test","fileName":"test","folder":"test-folder"}]}`,
			repoAddCalls:   1,
			repoAddErr:     errors.New("DynamoDB error"),
			wantCode:       http.StatusInternalServerError,
			wantInResponse: "Unable to save the zip request",
		},
		{
			scenario:       "New zip request created successfully",
			reqBody:        `{"files":[{"s3path":"s3://test/test","fileName":"test","folder":"test-folder"}]}`,
			repoAddCalls:   1,
			repoAddErr:     nil,
			wantCode:       http.StatusCreated,
			wantInResponse: `"Link":"/zip/`,
		},
	}
	for _, test := range tests {
		mr := new(MockRepository)
		_, l := newTestLogger()

		zh := ZipRequestHandler{
			repo:   mr,
			logger: l,
		}

		mux := http.NewServeMux()
		mux.Handle("GET /zip/{reference}", &zh)

		req, err := http.NewRequest("GET", "/zip/request", strings.NewReader(test.reqBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), middleware.HashedEmail{}, "testHash")

		entryRef := "RefPlaceholder"

		if test.repoAddCalls > 0 {
			mockCall := mr.On("Add", mock.AnythingOfType("*storage.Entry"))
			if test.repoAddErr == nil {
				mockCall.RunFn = func(args mock.Arguments) {
					entry := args[0].(*storage.Entry)
					assert.Equal(t, "testHash", entry.Hash, test.scenario)
					assert.NotEmpty(t, entry.Ref, test.scenario)
					entryRef = entry.Ref
				}
			}
			mockCall.Return(test.repoAddErr).Times(test.repoAddCalls)
		}

		mux.ServeHTTP(rr, req.WithContext(ctx))
		res := rr.Result()
		bodyBuf := new(bytes.Buffer)
		_, _ = bodyBuf.ReadFrom(res.Body)
		body := bodyBuf.String()

		assert.Equal(t, test.wantCode, res.StatusCode, test.scenario)
		assert.Contains(t, body, test.wantInResponse, test.scenario)

		if test.wantCode == http.StatusCreated {
			assert.Contains(t, body, entryRef, test.scenario)
		}
	}
}
