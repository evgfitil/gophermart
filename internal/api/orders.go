package api

import (
	"context"
	"fmt"
	"github.com/ShiraazMoollatjie/goluhn"
	"github.com/go-chi/jwtauth"
	"io"
	"net/http"
)

func UploadOrderHandler(s Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		requestContext, cancel := context.WithTimeout(req.Context(), requestTimeout)
		defer cancel()

		_, claims, err := jwtauth.FromContext(requestContext)
		if err != nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		if claims == nil {
			http.Error(res, err.Error(), http.StatusUnauthorized)
			return
		}

		username, ok := claims["user_id"].(string)
		if !ok {
			http.Error(res, err.Error(), http.StatusUnauthorized)
		}

		body, err := io.ReadAll(req.Body)
		defer req.Body.Close()

		if err != nil {
			http.Error(res, "failed to read the request body", http.StatusInternalServerError)
			return
		}

		orderNumber := string(body)
		if err = goluhn.Validate(orderNumber); err != nil {
			http.Error(res, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		fmt.Println(orderNumber, username)

		res.WriteHeader(http.StatusOK)
		res.Write([]byte("upload in successfully"))
	}
}
