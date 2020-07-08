package handlers

import (
	"log"
	"net/http"
	"opg-file-service/internal"
	"opg-file-service/middleware"
	"opg-file-service/session"
	"opg-file-service/storage"
	"opg-file-service/zipper"
	"time"

	"github.com/gorilla/mux"
)

type Repository interface {
	Get(ref string) (*storage.Entry, error)
	Delete(entry *storage.Entry) error
	Add(entry *storage.Entry) error
}

type ZipHandler struct {
	repo   Repository
	zipper zipper.ZipperInterface
	logger *log.Logger
}

func NewZipHandler(logger *log.Logger, sess *session.Session, repo Repository) (*ZipHandler, error) {
	return &ZipHandler{
		repo:   repo,
		zipper: zipper.NewZipper(*sess),
		logger: logger,
	}, nil
}

func (zh *ZipHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get the reference from the request
	vars := mux.Vars(r)
	reference := vars["reference"]

	zh.logger.Println("Zip files for reference:", reference)

	// fetch entry from DynamoDB
	entry, err := zh.repo.Get(reference)
	if err != nil {
		zh.logger.Println(err.Error())
		internal.WriteJSONError(rw, "ref", "Reference token not found.", http.StatusNotFound)
		return
	}

	if entry.IsExpired() {
		zh.logger.Println("Reference token '" + reference + "' has expired.")
		internal.WriteJSONError(rw, "ref", "Reference token has expired.", http.StatusNotFound)
		return
	}

	userHash := r.Context().Value(middleware.HashedEmail{})
	if entry.Hash != userHash {
		zh.logger.Println("Access denied for user:", userHash)
		internal.WriteJSONError(rw, "auth", "Access denied.", http.StatusForbidden)
		return
	}

	entry.DeDupe()

	zh.zipper.Open(rw)

	for _, file := range entry.Files {
		err := zh.zipper.AddFile(&file)
		if err != nil {
			zh.logger.Println(err.Error())
			internal.WriteJSONError(rw, "zip", "Unable to zip requested file.", http.StatusInternalServerError)
			return
		}
	}

	err = zh.zipper.Close()
	if err != nil {
		zh.logger.Println(err.Error())
	}

	err = zh.repo.Delete(entry)
	if err != nil {
		zh.logger.Println(err.Error())
		zh.logger.Println("Unable to delete entry for reference", entry.Ref)
	}

	zh.logger.Println("Request took: ", time.Since(start))
}
