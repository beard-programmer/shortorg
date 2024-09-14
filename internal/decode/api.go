package decode

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/beard-programmer/shortorg/internal/httpApi"
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
	httpStatusCode int
	Code           string `json:"code"`
	Message        string `json:"message"`
}

func HttpHandlerFunc(
	logger *zap.Logger,
	decodeFunc Fn,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiRequest, err := httpApi.DecodeRequest[APIRequest](r)
		if err != nil {
			handleError(w, r, fmt.Errorf("%w: invalid request body: %v", ValidationError, err))
			return
		}

		urlWasDecoded, found, err := decodeFunc(r.Context(), apiRequest)

		if err != nil {
			handleError(w, r, err)
			return
		}
		if !found {
			handleError(w, r, fmt.Errorf("%w: original was not found: %v", ApplicationError, err))
			return
		}

		response := APIResponse{
			OriginalURL: urlWasDecoded.Token.OriginalURL.String(),
			ShortURL:    fmt.Sprintf("https://%s/%s", urlWasDecoded.Token.Host.Hostname(), urlWasDecoded.Token.KeyEncoded.Value()),
		}
		httpApi.EncodeResponse(w, r, http.StatusOK, response)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr APIErrResponse

	switch {
	case errors.Is(err, ValidationError):
		apiErr = APIErrResponse{Code: "ValidationError", Message: err.Error(), httpStatusCode: http.StatusBadRequest}
	case errors.Is(err, ApplicationError):
		apiErr = APIErrResponse{Code: "ApplicationError", Message: err.Error(), httpStatusCode: http.StatusUnprocessableEntity}
	case errors.Is(err, InfrastructureError):
		apiErr = APIErrResponse{Code: "InfrastructureError", Message: err.Error(), httpStatusCode: http.StatusServiceUnavailable}
	default:
		apiErr = APIErrResponse{Code: "UnknownError", Message: err.Error(), httpStatusCode: http.StatusInternalServerError}
	}

	httpApi.EncodeResponse(w, r, apiErr.httpStatusCode, apiErr)
}
