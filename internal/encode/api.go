package encode

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

type APIRequest struct {
	URL          string  `json:"url"`
	EncodeAtHost *string `json:"encode_at_host"`
}

type APIResponse struct {
	URL      string `json:"url"`
	ShortURL string `json:"short_url"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func ApiHandler(identityProvider IdentifierProvider, parseUrl func(string) (URL, error), logger *zap.SugaredLogger, encodedUrlChan chan<- UrlWasEncoded) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req APIRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		urlWasEncoded, err := Encode(r.Context(), identityProvider, parseUrl, logger, Request{
			URL:          req.URL,
			EncodeAtHost: req.EncodeAtHost,
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
			if err != nil {
				panic(err)
			}
			return
		}

		encodedUrlChan <- *urlWasEncoded

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(APIResponse{
			URL:      urlWasEncoded.URL,
			ShortURL: "https://" + urlWasEncoded.Token.TokenHost.Host() + "/" + urlWasEncoded.Token.TokenEncoded.Value(),
		})
		if err != nil {
			panic(err)
		}
	}
}
