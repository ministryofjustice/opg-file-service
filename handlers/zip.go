package handlers

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"opg-s3-zipper-service/dynamo"
	"opg-s3-zipper-service/internal"
	"opg-s3-zipper-service/middleware"
	"opg-s3-zipper-service/session"
	"opg-s3-zipper-service/zipper"
	"time"
)

type ZipHandler struct {
	logger *log.Logger
}

func NewZipHandler(logger *log.Logger) *ZipHandler {
	return &ZipHandler{
		logger,
	}
}

func (zh *ZipHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get the reference from the request
	vars := mux.Vars(r)
	reference := vars["reference"]
	zh.logger.Println("Zip files for reference:", reference)

	sess, err := session.NewSession()
	if err != nil {
		zh.logger.Println(err.Error())
		internal.WriteJSONError(rw, "session", "Unable to start a new session", http.StatusInternalServerError)
		return
	}

	repo := dynamo.NewRepository(*sess, zh.logger)

	entry, err := repo.Get(reference)
	if err != nil {
		zh.logger.Println(err.Error())
		internal.WriteJSONError(rw, "ref", "Reference token not found.", http.StatusNotFound)
		return
	}

	if entry.IsExpired() {
		zh.logger.Println("Reference token has expired.")
		internal.WriteJSONError(rw, "ref", "Reference token has expired.", http.StatusNotFound)
		return
	}

	userHash := r.Context().Value(middleware.HashedEmail{})
	if entry.Hash != userHash {
		zh.logger.Println("Access denied for user ", userHash)
		internal.WriteJSONError(rw, "auth", "Access denied.", http.StatusForbidden)
		return
	}

	zipService := zipper.NewZipper(*sess, rw)

	zipService.Open()

	for _, file := range entry.Files {
		err := zipService.AddFile(&file)
		if err != nil {
			zh.logger.Println(err.Error())
			internal.WriteJSONError(rw, "zip", "Unable to zip requested file.", http.StatusInternalServerError)
			return
		}
	}

	err = zipService.Close()
	if err != nil {
		zh.logger.Println(err.Error())
	}

	err = repo.Delete(entry)
	if err != nil {
		zh.logger.Println(err.Error())
		zh.logger.Println("Unable to delete reference ", entry.Ref)
	}

	zh.logger.Println("Request took: ", time.Since(start))
}
