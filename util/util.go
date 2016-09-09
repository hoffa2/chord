package util

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func ErrorResponse(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusNotFound)
}

func ErrorNotFound(w http.ResponseWriter, format string, a ...interface{}) {
	errString := fmt.Sprintf(format, a)
	http.Error(w, errString, http.StatusNotFound)
}

func WriteJson(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(v)
}