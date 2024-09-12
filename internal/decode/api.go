package decode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type APIRequest struct {
	URL string `json:"short_url"`
}

func (r APIRequest) Url() string {
	return r.URL
}

type APIResponse struct {
	OriginalURL string `json:"url"`
	ShortURL    string `json:"short_url"`
}

type APIErrResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func HttpHandler(
	logger *zap.SugaredLogger,
	decodeFunc func(context.Context, DecodingRequest) (*UrlWasDecoded, *OriginalUrlWasNotFound, error),
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var apiRequest APIRequest
			err := json.NewDecoder(r.Body).Decode(&apiRequest)
			if err != nil {
				handleError(w, fmt.Errorf("invalid request body"))
				return
			}

			urlWasDecoded, originalWasNotFound, err := decodeFunc(r.Context(), apiRequest)

			if err != nil {
				handleError(w, err)
				return
			}
			if originalWasNotFound != nil {
				var apiErr APIErrResponse
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				apiErr = APIErrResponse{Code: "originalWasNotFound", Message: ""}
				if err := json.NewEncoder(w).Encode(apiErr); err != nil {
					http.Error(w, "Failed to write error response", http.StatusInternalServerError)
				}
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := APIResponse{
				OriginalURL: urlWasDecoded.Token.OriginalURL.String(),
				ShortURL:    fmt.Sprintf("https://%s/%s", urlWasDecoded.Token.Host.Hostname(), urlWasDecoded.Token.KeyEncoded.Value()),
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		},
	)
}

func handleError(w http.ResponseWriter, err error) {
	var apiErr APIErrResponse

	w.WriteHeader(http.StatusInternalServerError)
	apiErr = APIErrResponse{Code: "UnknownError", Message: err.Error()}
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		http.Error(w, "Failed to write error response", http.StatusInternalServerError)
	}
}
