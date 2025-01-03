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
	EncodeAtHost *string `json:"encodeAt_host"`
}

func (r APIRequest) OriginalUrl() string {
	return r.URL
}

func (r APIRequest) Host() *string {
	return r.EncodeAtHost
}

type APIResponse struct {
	URL      string `json:"url"`
	ShortURL string `json:"shortUrl"`
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
			handleError(w, r, fmt.Errorf("%w encode: invalid request body", errValidation))
			return
		}

		urlWasEncoded, err := encodeFunc(r.Context(), apiRequest)

		if err != nil {
			handleError(w, r, err)
			return
		}

		response := APIResponse{
			URL: urlWasEncoded.NonBrandedLink.DestinationURL.String(),
			ShortURL: fmt.Sprintf(
				"https://%s/%s",
				urlWasEncoded.NonBrandedLink.Host.Hostname(),
				urlWasEncoded.NonBrandedLink.Slug.Value(),
			),
		}
		httpEncoder.EncodeResponse(w, r, http.StatusOK, response)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr APIErrResponse

	switch {
	case errors.Is(err, errValidation):
		apiErr = APIErrResponse{
			Code:           "ValidationError",
			Message:        err.Error(),
			httpStatusCode: http.StatusBadRequest,
		}
	case errors.Is(err, errApplication):
		apiErr = APIErrResponse{
			Code:           "ApplicationError",
			Message:        err.Error(),
			httpStatusCode: http.StatusUnprocessableEntity,
		}
	case errors.Is(err, errInfrastructure):
		apiErr = APIErrResponse{
			Code:           "InfrastructureError",
			Message:        err.Error(),
			httpStatusCode: http.StatusServiceUnavailable,
		}
	default:
		apiErr = APIErrResponse{
			Code:           "UnknownError",
			Message:        err.Error(),
			httpStatusCode: http.StatusInternalServerError,
		}
	}

	httpEncoder.EncodeResponse(w, r, apiErr.httpStatusCode, apiErr)
}
