package internal

import (
	"flag"
	"net/http"
	"os"
)

func RunHealthcheck(addr string) {
	f := flag.Bool("hc", false, "perform a health check")
	flag.Parse()

	if !*f {
		return
	}

	resp, err := http.Get(addr)
	if err != nil || resp.StatusCode != 200 {
		os.Stdout.Write([]byte("FAIL"))
		os.Exit(1)
	}

	os.Stdout.Write([]byte("OK"))
	os.Exit(0)
}
