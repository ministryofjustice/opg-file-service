package middleware

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"opg-s3-zipper-service/utils"
	"os"
	"strings"
)

type hashedEmail struct {}

func JwtVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		//Get the token from the header
		header := r.Header.Get("Authorization")

		//If Authorization is empty, return a 403
		if header == "" {
			rw.WriteHeader(http.StatusForbidden)
			utils.WriteJSONError(rw, "missing_token", "Missing Authentication Token", http.StatusForbidden)
			return
		}

		header = strings.Split(header, "Bearer ")[1]

		token, err := jwt.Parse(header, func(token *jwt.Token) (i interface{}, err error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("MyTestSecret"), nil
		})

		// Return the error
		if err != nil {
			rw.WriteHeader(http.StatusForbidden)
			utils.WriteJSONError(rw, "error_with_token", err.Error(), http.StatusForbidden)
			return
		}

		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)
			e := claims["session-data"].(string)
			he := hashEmail(e)
			log.Println("JWT Token is valid for user", he)

			ctx := context.WithValue(r.Context(), hashedEmail{}, he)
			r.WithContext(ctx)
			next.ServeHTTP(rw, r)
		}
	})
}

// Create a hash of the users email
func hashEmail(e string) string {
	salt := os.Getenv("USER_HASH_SALT")
	h := md5.New()
	h.Write([]byte(salt + e))
	return hex.EncodeToString(h.Sum(nil))
}
