package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"opg-file-service/internal"
	"strings"
)

type HashedEmail struct{}

func JwtVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		jwtSecret := internal.GetEnvVar("JWT_SECRET", "MyTestSecret")

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
			he := hashEmail(e)
			log.Println("JWT Token is valid for user ", he)

			ctx := context.WithValue(r.Context(), HashedEmail{}, he)
			r.WithContext(ctx)
			next.ServeHTTP(rw, r)
		}
	})
}

// Create a hash of the users email
func hashEmail(e string) string {
	salt := internal.GetEnvVar("USER_HASH_SALT", "ufUvZWyqrCikO1HPcPfrz7qQ6ENV84p0")
	h := sha256.New()
	h.Write([]byte(salt + e))
	return hex.EncodeToString(h.Sum(nil))
}
