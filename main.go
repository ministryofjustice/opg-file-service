package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"opg-file-service/dynamo"
	"opg-file-service/handlers"
	"opg-file-service/session"
	"opg-file-service/zipper"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/ministryofjustice/opg-go-healthcheck/healthcheck"
)

func newServer(logger *log.Logger) (*http.Server, error) {
	sess, err := session.NewSession()
	if err != nil {
		logger.Println(err.Error())
		return nil, errors.New("unable to create a new session")
	}

	repo := dynamo.NewRepository(*sess, logger)

	zh := handlers.NewZipHandler(logger, zipper.NewZipper(*sess), repo)

	router := mux.NewRouter().PathPrefix(os.Getenv("PATH_PREFIX")).Subrouter()
	router.
		HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {})
	router.
		Handle("/zip/{reference}", handlers.JwtVerify(zh)).
		Methods(http.MethodGet)

	return &http.Server{
		Addr:         ":8000",           // configure the bind address
		Handler:      router,            // set the default handler
		ErrorLog:     logger,            // Set the logger for the server
		IdleTimeout:  120 * time.Second, // max time fro connections using TCP Keep-Alive
		ReadTimeout:  1 * time.Second,   // max time to read request from the client
		WriteTimeout: 15 * time.Minute,  // max time to write response to the client
	}, nil
}

func main() {
	healthcheck.Register("http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/health-check")
	logger := log.New(os.Stdout, "opg-file-service ", log.LstdFlags)

	server, err := newServer(logger)
	if err != nil {
		logger.Fatal(err)
	}

	// start the server
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Fatal(err)
		}
	}()

	// Gracefully shutdown when signal received
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	sig := <-c
	logger.Println("Received terminate, graceful shutdown", sig)

	tc, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(tc); err != nil {
		logger.Println(err)
	}
}
