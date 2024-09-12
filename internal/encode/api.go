package encode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

type APIRequest struct {
	URL          string  `json:"url"`
	EncodeAtHost *string `json:"encode_at_host"`
}

func (r APIRequest) OriginalUrl() string {
	return r.URL
}

func (r APIRequest) Host() *string {
	return r.EncodeAtHost
}

type APIResponse struct {
	URL      string `json:"url"`
	ShortURL string `json:"short_url"`
}

type APIErrResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func HandleEncode(
	logger *zap.SugaredLogger,
	encodeFunc func(context.Context, EncodingRequest) (*EncodedUrl, error),
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			var apiRequest APIRequest
			err := json.NewDecoder(r.Body).Decode(&apiRequest)
			if err != nil {
				handleError(w, ValidationError{Err: fmt.Errorf("invalid request body")})
				return
			}

			encodedUrl, err := encodeFunc(r.Context(), apiRequest)

			if err != nil {
				handleError(w, err)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := APIResponse{
				URL:      encodedUrl.URL.Value,
				ShortURL: fmt.Sprintf("https://%s/%s", encodedUrl.Token.TokenHost.Host(), encodedUrl.Token.TokenEncoded.Value()),
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				http.Error(w, "Failed to write response", http.StatusInternalServerError)
			}
		},
	)
}

func handleError(w http.ResponseWriter, err error) {
	var apiErr APIErrResponse
	switch e := err.(type) {
	case ValidationError:
		w.WriteHeader(http.StatusBadRequest)
		apiErr = APIErrResponse{Code: "ValidationError", Message: e.Error()}
	case InfrastructureError:
		w.WriteHeader(http.StatusInternalServerError)
		apiErr = APIErrResponse{Code: "InfrastructureError", Message: e.Error()}
	case ApplicationError:
		w.WriteHeader(http.StatusInternalServerError)
		apiErr = APIErrResponse{Code: "ApplicationError", Message: e.Error()}
	default:
		w.WriteHeader(http.StatusInternalServerError)
		apiErr = APIErrResponse{Code: "UnknownError", Message: e.Error()}
	}
	if err := json.NewEncoder(w).Encode(apiErr); err != nil {
		http.Error(w, "Failed to write error response", http.StatusInternalServerError)
	}
}

//
//func ApiHandler(identityProvider Identities, urlProvider UrlProvider, logger *zap.SugaredLogger, encodedUrlChan chan<- EncodedUrl) http.HandlerFunc {
//	return func(w http.ResponseWriter, r *http.EncodingRequest) {
//		var req APIRequest
//		err := json.NewDecoder(r.Body).Decode(&req)
//		if err != nil {
//			http.Error(w, "Invalid request body", http.StatusBadRequest)
//			return
//		}
//
//		urlWasEncoded, err := Encode(r.Context(), identityProvider, urlProvider, logger, EncodingRequest{
//			URL:          req.URL,
//			EncodeAtHost: req.EncodeAtHost,
//		})
//
//		if err != nil {
//			w.WriteHeader(http.StatusInternalServerError)
//			err := json.NewEncoder(w).Encode(ErrorResponse{Message: err.Error()})
//			if err != nil {
//				panic(err)
//			}
//			return
//		}
//
//		encodedUrlChan <- *urlWasEncoded
//
//		w.Header().Set("Content-Type", "application/json")
//		w.WriteHeader(http.StatusOK)
//		err = json.NewEncoder(w).Encode(APIResponse{
//			URL:      urlWasEncoded.URL,
//			ShortURL: "https://" + urlWasEncoded.Token.TokenHost.Host() + "/" + urlWasEncoded.Token.TokenEncoded.Value(),
//		})
//		if err != nil {
//			panic(err)
//		}
//	}
//}
