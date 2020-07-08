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

type envConfig struct {
	AWSRegion         string
	IAMRole           string
	DynamoDBEndpoint  string
	DynamoDBTableName string
	S3Endpoint        string
	JWTSecret         string
	UserHashSalt      string
}

func readEnvConfig() envConfig {
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "eu-west-1"
	}

	table := os.Getenv("AWS_DYNAMODB_TABLE_NAME")
	if table == "" {
		table = "zip-requests"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "MyTestSecret"
	}

	salt := os.Getenv("USER_HASH_SALT")
	if salt == "" {
		salt = "ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0"
	}

	return envConfig{
		AWSRegion:         awsRegion,
		IAMRole:           os.Getenv("AWS_IAM_ROLE"),
		DynamoDBEndpoint:  os.Getenv("AWS_DYNAMODB_ENDPOINT"),
		DynamoDBTableName: table,
		S3Endpoint:        os.Getenv("AWS_S3_ENDPOINT"),
		JWTSecret:         jwtSecret,
		UserHashSalt:      salt,
	}
}

func newServer(logger *log.Logger) (*http.Server, error) {
	config := readEnvConfig()

	sess, err := session.NewSession(config.AWSRegion, config.IAMRole)
	if err != nil {
		logger.Println(err.Error())
		return nil, errors.New("unable to create a new session")
	}

	repo := dynamo.NewRepository(*sess, logger, config.DynamoDBEndpoint, config.DynamoDBTableName)

	zh := handlers.NewZipHandler(logger, zipper.NewZipper(*sess, config.S3Endpoint), repo)

	router := mux.NewRouter().PathPrefix(os.Getenv("PATH_PREFIX")).Subrouter()
	router.
		HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {})
	router.
		Handle("/zip/{reference}", handlers.JwtVerify(config.JWTSecret, config.UserHashSalt, zh)).
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
