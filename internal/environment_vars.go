package internal

import "os"

func GetEnvVar(e string, d string) string {
	if envVar, ok := os.LookupEnv(e); !ok {
		return d
	} else {
		return envVar
	}
}