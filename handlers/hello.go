package handlers

import (
	"log"
	"net/http"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "hello world"}`))
	log.Printf("%s %s", r.Method, r.URL)

}
