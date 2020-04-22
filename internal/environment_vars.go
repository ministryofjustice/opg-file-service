package internal

import "os"

func GetEnvVar(e string, d string) string {
	envVar := os.Getenv(e)
	if envVar == "" {
		envVar := d
		return envVar
	}
	return envVar
}