package util

import (
	"fmt"
	"net/http"
)

func ErrorRespose(w http.ResponseWriter, err string) {

	w.WriteHeader(http.StatusNotFound)
}

func ErrorNotFound(w http.ResponseWriter, format string, a ...interface{}) {
	errString := fmt.Sprintf(format, a)
	http.Error(w, errString, http.StatusNotFound)
}
