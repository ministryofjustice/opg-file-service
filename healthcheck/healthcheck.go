package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	_, err := http.Get(fmt.Sprintf("http://127.0.0.1:8000%s/health-check", os.Getenv("PATH_PREFIX")))
	if err != nil {
		os.Exit(1)
	}
}
