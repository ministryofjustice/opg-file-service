package main

import (
	"context"
	"log"
	"net/http"
	"opg-file-service/dynamo"
	"opg-file-service/handlers"
	"opg-file-service/middleware"
	"opg-file-service/session"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/ministryofjustice/opg-go-healthcheck/healthcheck"
)

func main() {
	healthcheck.Register("http://localhost:8000" + os.Getenv("PATH_PREFIX") + "/health-check")

	// Create a Logger
	l := log.New(os.Stdout, "opg-file-service ", log.LstdFlags)

	// Create new serveMux
	sm := mux.NewRouter().PathPrefix(os.Getenv("PATH_PREFIX")).Subrouter()

	// Register the health check handler
	sm.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a sub-router for protected handlers
	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.Use(middleware.JwtVerify)

	// create a new AWS session
	sess, err := session.NewSession()
	if err != nil {
		l.Println(err.Error())
		l.Fatal("unable to create a new session")
	}

	repo := dynamo.NewRepository(*sess, l)

	// Register protected handlers
	zh, err := handlers.NewZipHandler(l, sess, repo)
	if err != nil {
		l.Fatal(err)
	}
	getRouter.Handle("/zip/{reference}", zh)

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
