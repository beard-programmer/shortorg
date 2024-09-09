package encode

import (
	"encoding/json"
	"net/http"

	"github.com/beard-programmer/shortorg/internal/encode/infrastructure"
	"github.com/jmoiron/sqlx"
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

// LoggingDB wraps sqlx.DB to log all queries with Zap
type LoggingDB struct {
	*sqlx.DB
	Logger *zap.SugaredLogger
}

func ApiHandler(IdentityDB *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req APIRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		identityProvider := infrastructure.PostgresIdentifierProvider{DB: IdentityDB}
		urlWasEncoded, err := Encode(&identityProvider, Request{
			URL:          req.URL,
			EncodeAtHost: req.EncodeAtHost,
		})

		if err != nil {
			// Respond with an error in JSON format
			w.WriteHeader(http.StatusInternalServerError)
			err := json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
			if err != nil {
				panic(err)
			}
			return
		}

		// Respond with the encoded URL in JSON format
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(APIResponse{
			URL:      urlWasEncoded.URL,
			ShortURL: "https://" + urlWasEncoded.ShortHost + "/" + urlWasEncoded.ShortToken,
		})
		if err != nil {
			panic(err)
		}
	}
}
