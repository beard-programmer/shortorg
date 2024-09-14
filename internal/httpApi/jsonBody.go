package httpApi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func EncodeResponse[T any](w http.ResponseWriter, r *http.Request, status int, v T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, fmt.Sprintf("DecodeRequest: %s", err), http.StatusInternalServerError)
	}
}

func DecodeRequest[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("DecodeRequest: %w", err)
	}
	return v, nil
}
