// @todo.. I have attempted to generate docs using go swagger but for some reason it is super flaky!
// I tried it using GoLand IDE which I suspect may have been the issue with indentation etc.
// Wasted enough time faffing so decided to write swagger docs manually

// File service API
//
// Documentation for Zip API
//
//  swagger: "2.0"
//	Schemes: http
// 	BasePath: /
// 	Version: 1.0.0
//
// 	Consumes:
// 	- application/json
//
//	Produces:
//	- application/json
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
    // parameters:
    // - name: reference
    //   in: path
    //   description: reference of the zip file request
    //   required: true
    // responses:
    //   '200':
    //     description: Zip file download
    //   '404':
    //     description: Zip file download
	getRouter.Handle("/zip/{reference}", zh)

	// swagger:operation POST /zip/request zip request
	// Makes a request for a set of files to be downloaded from S3
	// ---
	// parameters:
	// - name: ref
	//   in: body
	//   description: Unique reference for the zip file download request
	//   required: true
	// - name: files
	//   in: body
	//   description: reference of the zip file request
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
	// - name: hash
	//   in: body
	//   description: JWT token
	//   required: true
	// - name: ttl
	//   in: body
	//   description: Time before the request expires
	//   required: true
	// responses:
	//   '200':
	//     description: Zip file download

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
