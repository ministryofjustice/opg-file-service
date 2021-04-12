package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"opg-file-service/internal"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type HashedEmail struct{}

type Cacheable interface {
	GetSecretString(key string) (string, error)
}

func JwtVerify(secretsCache Cacheable) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			jwtSecret, jwtErr := secretsCache.GetSecretString("jwt-key")

			if jwtErr != nil {
				log.Fatal(jwtErr.Error())
			}

			//Get the token from the header
			header := r.Header.Get("Authorization")

			//If Authorization is empty, return a 401
			if header == "" {
				internal.WriteJSONError(rw, "missing_token", "Missing Authentication Token", http.StatusUnauthorized)
				return
			}

			header = strings.Split(header, "Bearer ")[1]

			token, err := jwt.Parse(header, func(token *jwt.Token) (i interface{}, err error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			// Return the error
			if err != nil {
				internal.WriteJSONError(rw, "error_with_token", err.Error(), http.StatusUnauthorized)
				return
			}

			if token.Valid {
				claims := token.Claims.(jwt.MapClaims)
				e := claims["session-data"].(string)
				salt, err := secretsCache.GetSecretString("user-hash-salt")
				if err != nil {
					log.Fatalln(err.Error())
				}
				he := hashEmail(e, salt)
				log.Println("JWT Token is valid for user ", he)

				ctx := context.WithValue(r.Context(), HashedEmail{}, he)
				next.ServeHTTP(rw, r.WithContext(ctx))
			}
		})
	}
}

// Create a hash of the users email
func hashEmail(e string, salt string) string {
	h := sha256.New()
	h.Write([]byte(salt + e))
	return hex.EncodeToString(h.Sum(nil))
}
