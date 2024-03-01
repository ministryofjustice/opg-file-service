package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"opg-file-service/dynamo"
	"opg-file-service/internal"
	"opg-file-service/middleware"
	"opg-file-service/session"
	"opg-file-service/zipper"
	"time"

	"github.com/gorilla/mux"
)

type ZipHandler struct {
	repo   dynamo.RepositoryInterface
	zipper zipper.ZipperInterface
	logger *slog.Logger
}

func NewZipHandler(logger *slog.Logger) (*ZipHandler, error) {
	// create a new AWS session
	sess, err := session.NewSession()
	if err != nil {
		logger.Error("unable to create a new session", slog.Any("err", err))
		return nil, errors.New("unable to create a new session")
	}

	return &ZipHandler{
		dynamo.NewRepository(*sess, logger),
		zipper.NewZipper(*sess),
		logger,
	}, nil
}

func (zh *ZipHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Get the reference from the request
	vars := mux.Vars(r)
	reference := vars["reference"]

	zh.logger.Info("Zip files for reference: " + reference)

	// fetch entry from DynamoDB
	entry, err := zh.repo.Get(reference)
	if err != nil {
		zh.logger.Error(err.Error())
		internal.WriteJSONError(rw, "ref", "Reference token not found.", http.StatusNotFound)
		return
	}

	if entry.IsExpired() {
		zh.logger.Info("Reference token '" + reference + "' has expired.")
		internal.WriteJSONError(rw, "ref", "Reference token has expired.", http.StatusNotFound)
		return
	}

	userHash := r.Context().Value(middleware.HashedEmail{})
	if entry.Hash != userHash {
		zh.logger.Info("Access denied for user", slog.Any("user", userHash))
		internal.WriteJSONError(rw, "auth", "Access denied.", http.StatusForbidden)
		return
	}

	entry.DeDupe()

	zh.zipper.Open(rw)

	for _, file := range entry.Files {
		err := zh.zipper.AddFile(&file)
		if err != nil {
			zh.logger.Error(err.Error())
			internal.WriteJSONError(rw, "zip", "Unable to zip requested file.", http.StatusInternalServerError)
			return
		}
	}

	err = zh.zipper.Close()
	if err != nil {
		zh.logger.Error(err.Error())
	}

	err = zh.repo.Delete(entry)
	if err != nil {
		zh.logger.Error("Unable to delete entry for reference", slog.Any("err", err.Error()), slog.Any("ref", entry.Ref))
	}

	zh.logger.Info("Request took: " + time.Since(start).String())
}
