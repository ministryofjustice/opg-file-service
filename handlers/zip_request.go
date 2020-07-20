package handlers

import (
	"encoding/json"
	"errors"
	"github.com/rs/xid"
	"log"
	"net/http"
	"opg-file-service/dynamo"
	"opg-file-service/internal"
	"opg-file-service/middleware"
	"opg-file-service/session"
	"opg-file-service/storage"
	"time"
)

type ZipRequestResponseBody struct {
	Link string
}

type ZipRequestHandler struct {
	repo   dynamo.RepositoryInterface
	logger *log.Logger
}

func NewZipRequestHandler(logger *log.Logger) (*ZipRequestHandler, error) {
	// create a new AWS session
	sess, err := session.NewSession()
	if err != nil {
		logger.Println(err.Error())
		return nil, errors.New("unable to create a new session")
	}

	return &ZipRequestHandler{
		dynamo.NewRepository(*sess, logger),
		logger,
	}, nil
}

func (zrh *ZipRequestHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	start := time.Now()

	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()

	entry := new(storage.Entry)

	err := d.Decode(entry)
	if err != nil {
		zrh.logger.Println(err.Error())
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
		zrh.logger.Println(err.Error())
		rw.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(rw).Encode(err); err != nil {
			log.Println("Unable to write JSON error to response:", err)
		}
		return
	}

	err = zrh.repo.Add(entry)
	if err != nil {
		zrh.logger.Println(err.Error())
		internal.WriteJSONError(rw, "request", "Unable to save the zip request.", http.StatusInternalServerError)
		return
	}

	jsonResp, err := json.Marshal(ZipRequestResponseBody{Link: "/zip/" + entry.Ref})
	if err != nil {
		zrh.logger.Println(err.Error())
		internal.WriteJSONError(rw, "request", "Unable to encode response object to JSON.", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusCreated)
	_, err = rw.Write(jsonResp)
	if err != nil {
		zrh.logger.Println(err.Error())
	}

	zrh.logger.Println("Request took: ", time.Since(start))
}
