package internal

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
)

func RunHealthcheck(path string) {
	f := flag.Bool("hc", false, "perform a health check")
	flag.Parse()

	if !*f {
		return
	}

	// Only allow slash, alphanum, dash & underscore
	if !regexp.MustCompile(`^/[a-zA-Z0-9/_-]*$`).MatchString(path) {
		fmt.Print("FAIL: INVALID PATH")
		os.Exit(1)
	}

	u := url.URL{
		Scheme: "http",
		Host:   "localhost:8000",
		Path:   path,
	}

	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Print("FAIL: ERROR")
		os.Exit(1)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Print("FAIL: INVALID STATUS CODE")
		os.Exit(1)
	}

	fmt.Print("OK")
	os.Exit(0)
}
