package util

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println(err)
	}
}

func HashValue(str string) string {
	h := sha1.New()
	io.WriteString(h, str)
	return string(h.Sum(nil))
}

func HashValueByte(str string) []byte {
	h := sha1.New()
	io.WriteString(h, str)
	return h.Sum(nil)
}
