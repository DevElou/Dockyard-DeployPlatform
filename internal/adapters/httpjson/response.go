package httpjson

import (
	"encoding/json"
	"net/http"
)

func Write(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if payload == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(payload)
}

func Error(w http.ResponseWriter, status int, code string, message string) {
	Write(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}
