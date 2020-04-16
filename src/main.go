package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/nicholasjackson/env"
	"log"
	"net/http"
	"opg-s3-zipper-service/handlers"
	"opg-s3-zipper-service/utils"
	"os"
	"os/signal"
	"time"
)

func main() {
	// Create a Logger
	l := log.New(os.Stdout, "aws-s3-zipper ", log.LstdFlags)

	//Init Redis connection
	utils.InitRedisPool(l)

	// create the handlers
	dh := handlers.NewDocuments(l)

	//Create new serveMux and register the handlers
	sm := mux.NewRouter()

	getRouter := sm.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/zip-documents/{reference}", dh.GetDocuments)
	getRouter.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "I'm working")
		l.Println("I'm Logging!")
	})

	s := &http.Server{
		Addr:         ":8000",           // configure the bind address
		Handler:      sm,                // set the default handler
		ErrorLog:     l,                 // Set the logger for the server
		IdleTimeout:  120 * time.Second, // max time fro connections using TCP Keep-Alive
		ReadTimeout:  1 * time.Second,   // max time to read request from the client
		WriteTimeout: 1 * time.Second,   // max time to write response to the client
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
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	l.Println("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)
}

