package handlers

import (
	"bytes"
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"opg-file-service/dynamo"
	"opg-file-service/middleware"
	"opg-file-service/storage"
	"opg-file-service/zipper"
	"testing"
)

func TestNewZipHandler(t *testing.T) {
	l := new(log.Logger)
	zh, err := NewZipHandler(l)
	assert.Nil(t, err)
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
		filesInZip	 *[]storage.File
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
			"No ref token passed in URL",
			"",
			"",
			0,
			nil,
			nil,
			nil,
			0,
			nil,
			0,
			nil,
			0,
			0,
			nil,
			404,
			[]string{},
		},
		{
			"Ref token not found",
			"test",
			"",
			1,
			nil,
			nil,
			storage.NotFoundError{Ref: "test"},
			0,
			nil,
			0,
			nil,
			0,
			0,
			nil,
			404,
			[]string{
				"Zip files for reference: test",
				storage.NotFoundError{Ref: "test"}.Error(),
			},
		},
		{
			"Ref token has expired",
			"test",
			"",
			1,
			&storage.Entry{
				Ref: "test",
				Ttl: 0,
			},
			nil,
			nil,
			0,
			nil,
			0,
			nil,
			0,
			0,
			nil,
			404,
			[]string{
				"Reference token 'test' has expired",
			},
		},
		{
			"Ref token does not belong to authenticated user",
			"test",
			"user",
			1,
			&storage.Entry{
				Ref:  "test",
				Hash: "otherUser",
				Ttl:  9999999999,
			},
			nil,
			nil,
			0,
			nil,
			0,
			nil,
			0,
			0,
			nil,
			403,
			[]string{
				"Access denied for user: user",
			},
		},
		{
			"Unable to zip one of the files",
			"test",
			"",
			1,
			&storage.Entry{
				Ref:   "test",
				Ttl:   9999999999,
				Files: []storage.File{{}},
			},
			&[]storage.File{{}},
			nil,
			0,
			nil,
			1,
			errors.New("error adding file to zip"),
			1,
			0,
			nil,
			500,
			[]string{
				"error adding file to zip",
			},
		},
		{
			"Error when closing zip",
			"test",
			"",
			1,
			&storage.Entry{
				Ref: "test",
				Ttl: 9999999999,
			},
			nil,
			nil,
			1,
			nil,
			0,
			nil,
			1,
			1,
			errors.New("some error when closing zip"),
			200,
			[]string{
				"some error when closing zip",
			},
		},
		{
			"Error when deleting entry from DB after it has been processed",
			"test",
			"",
			1,
			&storage.Entry{
				Ref: "test",
				Ttl: 9999999999,
			},
			nil,
			nil,
			1,
			errors.New("some error deleting entry"),
			0,
			nil,
			1,
			1,
			nil,
			200,
			[]string{
				"some error deleting entry",
			},
		},
		{
			"Successfully zip multiple files",
			"test",
			"user",
			1,
			&storage.Entry{
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
			&[]storage.File{
				{
					S3path:   "s3://files/file",
					FileName: "file",
					Folder:   "",
				},
				{
					S3path:   "s3://files/file",
					FileName: "file (1)",
					Folder:   "",
				},
			},
			nil,
			1,
			nil,
			1,
			nil,
			1,
			1,
			nil,
			200,
			[]string{
				"Request took:",
			},
		},
	}

	for _, test := range tests {
		mr := new(MockRepository)
		mz := new(MockZipper)
		buf := new(bytes.Buffer)
		l := log.New(buf, "test", log.LstdFlags)

		zh := ZipHandler{
			repo:   mr,
			zipper: mz,
			logger: l,
		}

		sm := mux.NewRouter()
		sm.Handle("/zip/{reference}", &zh)

		req, err := http.NewRequest("GET", "/zip/"+test.ref, nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		ctx := context.WithValue(req.Context(), middleware.HashedEmail{}, test.userHash)

		mr.On("Get", test.ref).Return(test.repoGetOut, test.repoGetErr).Times(test.repoGetCalls)
		mr.On("Delete", test.repoGetOut).Return(test.repoDelErr).Times(test.repoDelCalls)

		mz.On("Open", rr).Return().Times(test.openCalls)
		mz.On("Close").Return(test.closeErr).Times(test.closeCalls)
		if test.repoGetOut != nil && len(test.repoGetOut.Files) > 0 {
			for _, file := range *test.filesInZip {
				mz.On("AddFile", &file).Return(test.addFileErr).Times(test.addFileCalls)
			}
		}

		sm.ServeHTTP(rr, req.WithContext(ctx))
		res := rr.Result()

		for _, ls := range test.wantInLog {
			assert.Contains(t, buf.String(), ls, test.scenario)
		}

		assert.Equal(t, test.wantCode, res.StatusCode, test.scenario)
	}
}
