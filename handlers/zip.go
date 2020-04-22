package handlers

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"opg-s3-zipper-service/dynamo"
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
		http.Error(rw, "", 500) // TODO: convert to JSON
		return
	}

	repo := dynamo.NewRepository(*sess, zh.logger)

	entry, err := repo.GetEntry(reference)
	if err != nil {
		zh.logger.Println(err.Error())
		http.Error(rw, err.Error(), 404) // TODO: convert to JSON
		return
	}

	if entry.IsExpired() {
		zh.logger.Println("Zip reference has expired")
		http.Error(rw, "Zip reference has expired", 404) // TODO: convert to JSON
		return
	}

	// TODO: return a 403 if the user hash doesn't match

	zipService := zipper.NewZipper(*sess, rw)

	zipService.Open()

	for _, file := range entry.Files {
		err := zipService.AddFile(&file)
		if err != nil {
			zh.logger.Println(err.Error())
			http.Error(rw, "Could not add file to zip", 500) // TODO: convert to JSON
			return
		}
	}

	err = zipService.Close()
	if err != nil {
		zh.logger.Println(err.Error())
	}

	zh.logger.Println("Request took: ", time.Since(start))
}
