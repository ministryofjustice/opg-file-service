package handlers

import (
	"github.com/stretchr/testify/mock"
	"net/http"
	"opg-file-service/storage"
)

// Uses Testify

type MockRepository struct {
	mock.Mock
}

func (m MockRepository) Get(ref string) (*storage.Entry, error) {
	args := m.Called(ref)
	return args.Get(0).(*storage.Entry), args.Error(1)
}

func (m MockRepository) Delete(entry *storage.Entry) error {
	args := m.Called(entry)
	return args.Error(0)
}

func (m MockRepository) Add(entry *storage.Entry) error {
	args := m.Called(entry)
	return args.Error(0)
}

type MockZipper struct {
	mock.Mock
}

func (m MockZipper) Open(rw http.ResponseWriter) {
	m.Called(rw)
}

func (m MockZipper) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m MockZipper) AddFile(f *storage.File) error {
	args := m.Called(f)
	return args.Error(0)
}
