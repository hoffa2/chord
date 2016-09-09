package util

import (
	"fmt"
	"net/http"
)

func errorRespose(w http.ResponseWriter, err string) {

	w.WriteHeader(http.StatusNotFound)
}

func errorNotFound(w http.ResponseWriter, format string, a ...interface{}) {
	errString := fmt.Sprintf(format, a)
	http.Error(w, errString, http.StatusNotFound)
}
