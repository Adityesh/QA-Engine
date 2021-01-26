package middlewares

import (
	"errors"
	"net/http"

	"example.org/model"
	"github.com/dgrijalva/jwt-go"
)

func VerifyRequest(response http.ResponseWriter, request *http.Request) error {
	c, err := request.Cookie("token")

	if err != nil {
		if err == http.ErrNoCookie {
			// Cookie not present in the request
			response.WriteHeader(http.StatusUnauthorized)
			return errors.New(err.Error())
		}
		response.WriteHeader(http.StatusBadRequest)
		return errors.New(err.Error())
		
		
	} else {
		tokenString := c.Value

		claims := &model.Claims{}


		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return token, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				response.WriteHeader(http.StatusUnauthorized)
				return errors.New("Unauthorized Access")
			}
			response.WriteHeader(http.StatusBadRequest)
			return errors.New("Bad Request")
		}
		if !token.Valid {
			response.WriteHeader(http.StatusUnauthorized)
			return errors.New("Unauthorized Access")
		}

		return nil
		
	}
}