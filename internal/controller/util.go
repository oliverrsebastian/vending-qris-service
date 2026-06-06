package controller

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data   any    `json:"data"`
	Code   int    `json:"code"`
	Status string `json:"status"`
}

type AppHandler func(r *http.Request) (Response, error)

func Handle(handler AppHandler) (handlerFunc http.HandlerFunc) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response, err := handler(r)

		w.WriteHeader(response.Code)

		if response.Code == 0 {
			w.WriteHeader(http.StatusInternalServerError)
		}

		if err != nil {
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}
