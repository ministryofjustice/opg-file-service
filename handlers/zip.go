package handlers

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"log/slog"
	"net/http"
	"opg-file-service/dynamo"
	"opg-file-service/internal"
	"opg-file-service/middleware"
	"opg-file-service/zipper"
	"time"
)

type ZipHandler struct {
	repo   dynamo.RepositoryInterface
	zipper zipper.ZipperInterface
	logger *slog.Logger
}

func NewZipHandler(logger *slog.Logger, cfg *aws.Config, repo dynamo.RepositoryInterface) *ZipHandler {
	return &ZipHandler{
		repo,
		zipper.NewZipper(cfg),
		logger,
	}
}

func (zh *ZipHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	reference := r.PathValue("reference")
	zh.logger.Info("Zip files for reference: " + reference)

	// fetch entry from DynamoDB
	entry, err := zh.repo.Get(r.Context(), reference)
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

	err = zh.repo.Delete(r.Context(), entry)
	if err != nil {
		zh.logger.Error("Unable to delete entry for reference", slog.Any("err", err.Error()), slog.Any("ref", entry.Ref))
	}

	zh.logger.Info("Request took: " + time.Since(start).String())
}
