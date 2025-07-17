package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"opg-file-service/internal"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type HashedEmail struct{}

type cacheable interface {
	GetSecretString(key string) (string, error)
}

func JwtVerify(logger *slog.Logger, secretsCache cacheable) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			jwtSecret, jwtErr := secretsCache.GetSecretString("jwt-key")

			if jwtErr != nil {
				logger.Error("Error in fetching JWT secret from cache", slog.Any("err", jwtErr.Error()))
				internal.WriteJSONError(rw, "missing_secret_key", jwtErr.Error(), http.StatusInternalServerError)
				return
			}

			//Get the token from the header
			header := r.Header.Get("Authorization")

			//If Authorization is empty, return a 401
			if header == "" {
				internal.WriteJSONError(rw, "missing_token", "Missing Authentication Token", http.StatusUnauthorized)
				return
			}

			header = strings.Split(header, "Bearer ")[1]

			token, parseErr := jwt.Parse(header, func(token *jwt.Token) (i interface{}, err error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			// Return the error
			if parseErr != nil {
				internal.WriteJSONError(rw, "error_with_token", parseErr.Error(), http.StatusUnauthorized)
				return
			}

			if token.Valid {
				claims := token.Claims.(jwt.MapClaims)
				e := claims["session-data"].(string)
				salt, saltErr := secretsCache.GetSecretString("user-hash-salt")
				if saltErr != nil {
					logger.Error("Error in fetching hash salt from cache:", slog.Any("err", saltErr.Error()))
					internal.WriteJSONError(rw, "missing_secret_salt", saltErr.Error(), http.StatusInternalServerError)
					return
				}
				he := hashEmail(e, salt)
				logger.Info("JWT Token is valid for user " + he)

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
