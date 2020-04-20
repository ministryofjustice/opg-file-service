package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strings"
)

type ErrorMsg struct {
	Message string `json:"message"`
}

func JwtVerify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {

		//Get the token from the header
		header := r.Header.Get("Authorization")

		//If Authorization is empty, return a 403
		if header == "" {
			rw.WriteHeader(http.StatusForbidden)
			json.NewEncoder(rw).Encode(ErrorMsg{ Message: "Missing Authentication Token" })
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
			json.NewEncoder(rw).Encode(ErrorMsg{Message: err.Error()})
			return
		}

		if token.Valid {
			log.Println("JWT Token is valid")
			next.ServeHTTP(rw, r)
		}
	})
}
