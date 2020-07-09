package handlers

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/ministryofjustice/opg-file-service/internal/dynamo"
	"github.com/ministryofjustice/opg-file-service/internal/storage"
	"github.com/ministryofjustice/opg-file-service/internal/zipper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewZipHandler(t *testing.T) {
	l := new(log.Logger)
	zh := NewZipHandler(l, &zipper.Zipper{}, &dynamo.Repository{})
	assert.IsType(t, ZipHandler{}, *zh)
	assert.Equal(t, l, zh.logger)
	assert.IsType(t, new(zipper.Zipper), zh.zipper)
	assert.IsType(t, new(dynamo.Repository), zh.repo)
}

func TestZipHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		scenario     string
		ref          string
		userHash     string
		repoGetCalls int
		repoGetOut   *storage.Entry
		repoGetErr   error
		repoDelCalls int
		repoDelErr   error
		addFileCalls int
		addFileErr   error
		openCalls    int
		closeCalls   int
		closeErr     error
		wantCode     int
		wantInLog    []string
	}{
		{
			scenario:  "no_ref",
			wantCode:  404,
			wantInLog: []string{},
		},
		{
			scenario:     "ref_not_found",
			ref:          "test",
			repoGetCalls: 1,
			repoGetErr:   errors.New("could not find entry 'test'"),
			wantCode:     404,
			wantInLog: []string{
				"Zip files for reference: test",
				"could not find entry 'test'",
			},
		},
		{
			scenario:     "ref_expired",
			ref:          "test",
			repoGetCalls: 1,
			repoGetOut: &storage.Entry{
				Ref: "test",
				Ttl: 0,
			},
			wantCode: 404,
			wantInLog: []string{
				"Reference token 'test' has expired",
			},
		},
		{
			scenario:     "ref_does_not_belong_to_user",
			ref:          "test",
			userHash:     "user",
			repoGetCalls: 1,
			repoGetOut: &storage.Entry{
				Ref:  "test",
				Hash: "otherUser",
				Ttl:  9999999999,
			},
			wantCode: 403,
			wantInLog: []string{
				"Access denied for user: user",
			},
		},
		{
			scenario:     "unable_to_zip_file",
			ref:          "test",
			repoGetCalls: 1,
			repoGetOut: &storage.Entry{
				Ref:   "test",
				Ttl:   9999999999,
				Files: []storage.File{{}},
			},
			addFileCalls: 1,
			addFileErr:   errors.New("error adding file to zip"),
			openCalls:    1,
			wantCode:     500,
			wantInLog: []string{
				"error adding file to zip",
			},
		},
		{
			scenario:     "error_on_closing_zip",
			ref:          "test",
			repoGetCalls: 1,
			repoGetOut: &storage.Entry{
				Ref: "test",
				Ttl: 9999999999,
			},
			repoDelCalls: 1,
			openCalls:    1,
			closeCalls:   1,
			closeErr:     errors.New("some error when closing zip"),
			wantCode:     200,
			wantInLog: []string{
				"some error when closing zip",
			},
		},
		{
			scenario:     "error_deleting_entry_from_db",
			ref:          "test",
			repoGetCalls: 1,
			repoGetOut: &storage.Entry{
				Ref: "test",
				Ttl: 9999999999,
			},
			repoDelCalls: 1,
			repoDelErr:   errors.New("some error deleting entry"),
			openCalls:    1,
			closeCalls:   1,
			wantCode:     200,
			wantInLog: []string{
				"some error deleting entry",
			},
		},
		{
			scenario:     "success",
			ref:          "test",
			userHash:     "user",
			repoGetCalls: 1,
			repoGetOut: &storage.Entry{
				Ref:  "test",
				Hash: "user",
				Ttl:  9999999999,
				Files: []storage.File{
					{
						S3path:   "s3://files/file",
						FileName: "file",
						Folder:   "",
					},
					{
						S3path:   "s3://files/file",
						FileName: "file",
						Folder:   "",
					},
				},
			},
			repoDelCalls: 1,
			addFileCalls: 2,
			openCalls:    1,
			closeCalls:   1,
			wantCode:     200,
			wantInLog: []string{
				"Request took:",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			w := httptest.NewRecorder()

			var buf bytes.Buffer
			l := log.New(&buf, "test", log.LstdFlags)

			mr := &MockRepository{}
			mr.On("Get", test.ref).Return(test.repoGetOut, test.repoGetErr).Times(test.repoGetCalls)
			mr.On("Delete", test.repoGetOut).Return(test.repoDelErr).Times(test.repoDelCalls)

			mz := &MockZipper{}
			mz.On("Open", w).Return().Times(test.openCalls)
			mz.On("Close").Return(test.closeErr).Times(test.closeCalls)

			sm := mux.NewRouter()
			sm.Handle("/zip/{reference}", &ZipHandler{
				repo:   mr,
				zipper: mz,
				logger: l,
			})

			req, err := http.NewRequest("GET", "/zip/"+test.ref, nil)
			if err != nil {
				t.Fatal(err)
			}

			ctx := context.WithValue(req.Context(), hashedEmail{}, test.userHash)

			if test.addFileCalls > 0 {
				mz.On("AddFile", mock.AnythingOfType("*storage.File")).Return(test.addFileErr).Times(test.addFileCalls)
			}

			sm.ServeHTTP(w, req.WithContext(ctx))
			resp := w.Result()

			for _, ls := range test.wantInLog {
				assert.Contains(t, buf.String(), ls, test.scenario)
			}

			assert.Equal(t, test.wantCode, resp.StatusCode, test.scenario)
		})
	}
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Get(ref string) (*storage.Entry, error) {
	args := m.Called(ref)
	return args.Get(0).(*storage.Entry), args.Error(1)
}

func (m *MockRepository) Delete(entry *storage.Entry) error {
	args := m.Called(entry)
	return args.Error(0)
}

func (m *MockRepository) Add(entry *storage.Entry) error {
	args := m.Called(entry)
	return args.Error(0)
}

type MockZipper struct {
	mock.Mock
}

func (m *MockZipper) Open(rw http.ResponseWriter) {
	m.Called(rw)
}

func (m *MockZipper) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockZipper) AddFile(f *storage.File) error {
	args := m.Called(f)
	return args.Error(0)
}
