package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"opg-file-service/dynamo"
	"opg-file-service/internal"
	"opg-file-service/middleware"
	"opg-file-service/storage"
	"time"

	"github.com/rs/xid"
)

type ZipRequestResponseBody struct {
	Link string
}

type ZipRequestHandler struct {
	repo   dynamo.RepositoryInterface
	logger *slog.Logger
}

func NewZipRequestHandler(logger *slog.Logger, repo dynamo.RepositoryInterface) *ZipRequestHandler {
	return &ZipRequestHandler{
		repo,
		logger,
	}
}

func (zrh *ZipRequestHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	entry := new(storage.Entry)

	err := d.Decode(entry)
	if err != nil {
		zrh.logger.Info(err.Error())
		internal.WriteJSONError(rw, "request", "Invalid JSON request.", http.StatusBadRequest)
		return
	}

	if entry.Files == nil {
		internal.WriteJSONError(rw, "request", "Missing field 'files'", http.StatusBadRequest)
		return
	}

	entry.Ref = xid.New().String()
	entry.Ttl = time.Now().Add(5 * time.Minute).Unix()
	entry.Hash = r.Context().Value(middleware.HashedEmail{}).(string)

	if ok, err := entry.Validate(); !ok {
		zrh.logger.Error(err.Error())
		rw.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(rw).Encode(err); err != nil {
			zrh.logger.Info("Unable to write JSON error to response:", "err", err)
		}
		return
	}

	err = zrh.repo.Add(r.Context(), entry)
	if err != nil {
		zrh.logger.Error(err.Error())
		internal.WriteJSONError(rw, "request", "Unable to save the zip request.", http.StatusInternalServerError)
		return
	}

	jsonResp, err := json.Marshal(ZipRequestResponseBody{Link: "/zip/" + entry.Ref})
	if err != nil {
		zrh.logger.Error(err.Error())
		internal.WriteJSONError(rw, "request", "Unable to encode response object to JSON.", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	_, err = rw.Write(jsonResp)
	if err != nil {
		zrh.logger.Error(err.Error())
	}

	zrh.logger.Info("Request took: " + time.Since(start).String())
}
