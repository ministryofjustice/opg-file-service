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
	"log"
	"log/slog"
	"net/http"
	"opg-file-service/cache"
	"opg-file-service/handlers"
	"opg-file-service/middleware"
	"os"
	"os/signal"
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

func run(ctx context.Context, l *slog.Logger) error {
	pathPrefix := os.Getenv("PATH_PREFIX")
	exportTraces := env.Get("TRACING_ENABLED", "0") == "1"

	shutdown, err := telemetry.StartTracerProvider(ctx, l, exportTraces)
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

	secretsCache := cache.New()
	jwt := middleware.JwtVerify(l, secretsCache)

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
	zrh, err := handlers.NewZipRequestHandler(l)
	if err != nil {
		return err
	}
	mux.Handle("POST /zip/request", jwt(zrh))

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
	zh, err := handlers.NewZipHandler(l)
	if err != nil {
		return err
	}
	mux.Handle("GET /zip/{reference}", jwt(zh))

	stdLogger := log.New(os.Stdout, "opg-file-service", log.LstdFlags)

	telemetryMiddleware := telemetry.Middleware(l)

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
			l.Error("listen and serve error", slog.Any("err", err.Error()))
			os.Exit(1)
		}
	}()

	// Gracefully shutdown when signal received
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	sig := <-c
	l.Info("Received terminate, graceful shutdown", "sig", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return s.Shutdown(tc)
}
