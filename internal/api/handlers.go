package api

import "net/http"

func Ping() http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
	}
}
