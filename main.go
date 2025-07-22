// File service API
//
// Documentation for creating a file request and downloading files using the API
//
//	  Schemes: http, https
//	  BasePath: /
//	  Version: 1.0.0
//	  securityDefinitions:
//	    Bearer:
//	      type: apiKey
//	      name: Authorization
//	      in: header
//
//	  Consumes:
//		  - application/json
//
//	  Produces:
//	   - application/json
//
// swagger:meta
package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"log"
	"log/slog"
	"net/http"
	"opg-file-service/cache"
	"opg-file-service/dynamo"
	"opg-file-service/handlers"
	"opg-file-service/internal"
	"opg-file-service/middleware"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ministryofjustice/opg-go-common/env"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-go-healthcheck/healthcheck"
)

func main() {
	ctx := context.Background()
	logger := telemetry.NewLogger("opg-file-service")

	if err := run(ctx, logger); err != nil {
		logger.Error("fatal startup error", slog.Any("err", err.Error()))
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	pathPrefix := os.Getenv("PATH_PREFIX")
	exportTraces := env.Get("TRACING_ENABLED", "0") == "1"

	shutdown, err := telemetry.StartTracerProvider(ctx, logger, exportTraces)
	defer shutdown()
	if err != nil {
		return err
	}

	healthcheck.Register("http://localhost:8000" + pathPrefix + "/health-check")

	// Create new serveMux
	mux := http.NewServeMux()

	// swagger:operation GET /health-check check health-check
	// Register the health check handler
	// ---
	// responses:
	//   '200':
	//     description: File service is up and running
	//   '404':
	//     description: Not found
	mux.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	cfg, err := awsConfig(ctx)
	if err != nil {
		return err
	}

	repository := dynamo.NewRepository(cfg, logger)

	secretsCache := cache.New(cfg)
	jwt := middleware.JwtVerify(logger, secretsCache)

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
	//   '201':
	//     description: Zip request created
	//     schema:
	//       type: object
	//       properties:
	//         link:
	//           type: string
	//           description: Link to download the zip file
	//   '403':
	//     description: Access denied
	//   '401':
	//     description: Missing, invalid or expired JWT token
	//   '400':
	//     description: Invalid JSON request
	//   '500':
	//     description: Unexpected error occurred
	mux.Handle("POST /zip/request", jwt(handlers.NewZipRequestHandler(logger, repository)))

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
	//     description: Missing, invalid or expired JWT token
	//   '500':
	//     description: Unexpected error occurred
	mux.Handle("GET /zip/{reference}", jwt(handlers.NewZipHandler(logger, cfg, repository)))

	stdLogger := log.New(os.Stdout, "opg-file-service", log.LstdFlags)

	telemetryMiddleware := telemetry.Middleware(logger)

	handler := http.StripPrefix(pathPrefix, telemetryMiddleware(mux))

	s := &http.Server{
		Addr:         ":8000",           // configure the bind address
		Handler:      handler,           // set the default handler
		ErrorLog:     stdLogger,         // Set the logger for the server
		IdleTimeout:  120 * time.Second, // max time fro connections using TCP Keep-Alive
		ReadTimeout:  1 * time.Second,   // max time to read request from the client
		WriteTimeout: 15 * time.Minute,  // max time to write response to the client
	}

	// start the server
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			logger.Error("listen and serve error", slog.Any("err", err.Error()))
			os.Exit(1)
		}
	}()

	// Gracefully shutdown when signal received
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	sig := <-c
	logger.Info("signal received: ", "sig", sig)

	tc, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return s.Shutdown(tc)
}

func awsConfig(ctx context.Context) (*aws.Config, error) {
	awsRegion := internal.GetEnvVar("AWS_REGION", "eu-west-1")

	creds := credentials.NewStaticCredentialsProvider(
		os.Getenv("AWS_ACCESS_KEY_ID"),
		os.Getenv("AWS_SECRET_ACCESS_KEY"),
		os.Getenv("AWS_SESSION_TOKEN"), // optional
	)

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(creds),
		config.WithAssumeRoleCredentialOptions(func(o *stscreds.AssumeRoleOptions) {
			o.RoleARN = os.Getenv("AWS_IAM_ROLE")
		}),

	)
	if err != nil {
		return nil, err
	}

	if endpoint, ok := os.LookupEnv("AWS_ENDPOINT"); ok {
		cfg.BaseEndpoint = &endpoint
	}

	return &cfg, nil
}
