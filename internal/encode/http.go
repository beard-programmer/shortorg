package encode

import (
	"errors"
	"fmt"
	"net/http"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/httpEncoder"
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
	httpStatusCode int
	Code           string `json:"code"`
	Message        string `json:"message"`
}

func HttpHandlerFunc(
	logger *appLogger.AppLogger,
	encodeFunc Fn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiRequest, err := httpEncoder.DecodeRequest[APIRequest](r)
		if err != nil {
			handleError(w, r, ValidationError{Err: fmt.Errorf("invalid request body")})
			return
		}

		urlWasEncoded, err := encodeFunc(r.Context(), apiRequest)

		if err != nil {
			handleError(w, r, err)
			return
		}

		response := APIResponse{
			URL:      urlWasEncoded.Token.OriginalURL.String(),
			ShortURL: fmt.Sprintf("https://%s/%s", urlWasEncoded.Token.Host.Hostname(), urlWasEncoded.Token.KeyEncoded.Value()),
		}
		httpEncoder.EncodeResponse(w, r, http.StatusOK, response)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr APIErrResponse

	var validationErr ValidationError
	var applicationErr ApplicationError
	var infrastructureErr InfrastructureError

	switch {
	case errors.As(err, &validationErr):
		apiErr = APIErrResponse{Code: "ValidationError", Message: err.Error(), httpStatusCode: http.StatusBadRequest}
	case errors.As(err, &applicationErr):
		apiErr = APIErrResponse{Code: "ApplicationError", Message: err.Error(), httpStatusCode: http.StatusUnprocessableEntity}
	case errors.As(err, &infrastructureErr):
		apiErr = APIErrResponse{Code: "InfrastructureError", Message: err.Error(), httpStatusCode: http.StatusServiceUnavailable}
	default:
		apiErr = APIErrResponse{Code: "UnknownError", Message: err.Error(), httpStatusCode: http.StatusInternalServerError}
	}

	httpEncoder.EncodeResponse(w, r, apiErr.httpStatusCode, apiErr)
}
