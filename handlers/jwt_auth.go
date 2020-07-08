package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

type hashedEmail struct{}

func JwtVerify(secret, salt string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		//Get the token from the header
		header := r.Header.Get("Authorization")

		//If Authorization is empty, return a 401
		if header == "" {
			writeJSONError(rw, "missing_token", "Missing Authentication Token", http.StatusUnauthorized)
			return
		}

		header = strings.Split(header, "Bearer ")[1]

		token, err := jwt.Parse(header, func(token *jwt.Token) (i interface{}, err error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		// Return the error
		if err != nil {
			writeJSONError(rw, "error_with_token", err.Error(), http.StatusUnauthorized)
			return
		}

		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)
			e := claims["session-data"].(string)
			he := hashEmail(salt, e)
			log.Println("JWT Token is valid for user ", he)

			ctx := context.WithValue(r.Context(), hashedEmail{}, he)
			next.ServeHTTP(rw, r.WithContext(ctx))
		}
	})
}

// Create a hash of the users email
func hashEmail(salt, e string) string {
	h := sha256.New()
	h.Write([]byte(salt + e))
	return hex.EncodeToString(h.Sum(nil))
}
