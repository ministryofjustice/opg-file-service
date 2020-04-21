package handlers

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"opg-s3-zipper-service/dynamo"
	"opg-s3-zipper-service/session"
)

type Zip struct {
	logger *log.Logger
}

func NewZip(logger *log.Logger) *Zip {
	return &Zip{
		logger,
	}
}

func (zip *Zip) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// Get the reference from the request
	vars := mux.Vars(r)
	reference := vars["reference"]
	zip.logger.Println("Zip files for reference:", reference)

	sess, err := session.NewSession()
	if err != nil {
		zip.logger.Println(err.Error())
		http.Error(rw, "", 500) // TODO: convert to JSON
	}

	repo := dynamo.NewRepository(sess, zip.logger)

	repo.ListTables() // TODO: remove debug

	entry, err := repo.GetEntry(reference)
	if err != nil {
		zip.logger.Println(err.Error())
		http.Error(rw, err.Error(), 404) // TODO: convert to JSON
		return
	}

	for _, file := range entry.Files {
		// TODO: fetch files from S3 and create zip
		zip.logger.Println(file.S3path)
	}
}
