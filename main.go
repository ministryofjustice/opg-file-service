// File service API
//
// Documentation for creating a file request and downloading files using the API
//
//   Schemes: http, https
//   BasePath: /
//   Version: 1.0.0
//   securityDefinitions:
//     Bearer:
//       type: apiKey
//       name: Authorization
//       in: header
//
//   Consumes:
// 	  - application/json
//
//   Produces:
//    - application/json
//
// swagger:meta
package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/ministryofjustice/opg-go-healthcheck/healthcheck"
	"log"
	"net/http"
	"opg-file-service/handlers"
	"opg-file-service/middleware"
	"os"
	"os/signal"
	"time"
)

func main() {
	healthcheck.Register("http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/health-check")

	// Create a Logger
	l := log.New(os.Stdout, "opg-file-service ", log.LstdFlags)

	// Create new serveMux
	sm := mux.NewRouter().PathPrefix(os.Getenv("PATH_PREFIX")).Subrouter()

	// swagger:operation GET /health-check check health-check
	// Register the health check handler
	// ---
	// responses:
	//   '200':
	//     description: Checks to see if file service is up and running
	//   '404':
	//     description: Not found
	sm.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a sub-router for protected handlers
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.Use(middleware.JwtVerify)

	// Register protected handlers
	zh, err := handlers.NewZipHandler(l)
	if err != nil {
		l.Fatal(err)
	}

    // swagger:operation GET /zip/{reference} zip download
    // Download Zip file from zip request reference
    // ---
    // produces:
	//   - application/zip
	//   - application/json
	// security:
	//  - Bearer: []
    // parameters:
    // - name: reference
    //   in: path
    //   description: reference of the zip file request
    //   required: true
    //
    // responses:
    //   '200':
    //     description: Zip file download
    //   '404':
    //     description: File download request for ref not found
	//   '403':
	//     description: Access denied
	//   '401':
	//     description: Missing JWT token
	//   '500':
	//     description: Unexpected error occurred
	getRouter.Handle("/zip/{reference}", zh)

	// swagger:operation POST /zip/request zip request
	// Makes a request for a set of files to be downloaded from S3
	// ---
	// security:
	//  - Bearer: []
	// parameters:
	// - name: files
	//   in: body
	//   description: s3 file paths alongside the human readable filenames as each file will be displayed in the zip file
	//   required: true
	//   schema:
	//       type: array
	//       items:
	//           type: object
	//           properties:
	//              s3path:
	//                  type: string
	//              filename:
	//                  type: string
	//              folder:
	//                  type: string
	// responses:
	//   '200':
	//     description: Zip file download
	//     schema:
	//       type: object
	//       properties:
	//         link:
	//           type: string
	//           description: Link to download the zip file
	//   '403':
	//     description: Access denied
	//   '401':
	//     description: Missing JWT token
	//   '500':
	//     description: Unexpected error occurred

	// @todo write code for this endpoint (or generate?)

	s := &http.Server{
		Addr:         ":8000",           // configure the bind address
		Handler:      sm,                // set the default handler
		ErrorLog:     l,                 // Set the logger for the server
		IdleTimeout:  120 * time.Second, // max time fro connections using TCP Keep-Alive
		ReadTimeout:  1 * time.Second,   // max time to read request from the client
		WriteTimeout: 15 * time.Minute,  // max time to write response to the client
	}

	// start the server
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal(err)
		}
	}()

	// Gracefully shutdown when signal received
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	sig := <-c
	l.Println("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}
