package resolveLink

import (
	"errors"
	"fmt"
	"net/http"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/httpEncoder"
)

type requestHTTP struct {
	URL string `json:"shortUrl"`
}

func (r requestHTTP) Url() string {
	return r.URL
}

type responseHTTP struct {
	OriginalURL string `json:"url"`
	ShortURL    string `json:"shortUrl"`
}

type responseErrHTTP struct {
	httpStatusCode int
	Code           string `json:"code"`
	Message        string `json:"message"`
}

func HTTPHandlerFunc(
	logger *appLogger.AppLogger,
	resolveLinkFn ResolveLinkFn,
) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		apiRequest, err := httpEncoder.DecodeRequest[requestHTTP](request)
		if err != nil {
			handleError(writer, request, fmt.Errorf("%writer: invalid request body: %v", errValidation, err))

			return
		}

		urlWasDecoded, found, err := resolveLinkFn(request.Context(), apiRequest)

		if err != nil {
			handleError(writer, request, err)

			return
		}
		if !found {
			handleError(writer, request, fmt.Errorf("%writer: original was not found: %v", errApplication, err))

			return
		}

		response := responseHTTP{
			OriginalURL: urlWasDecoded.NonBrandedLink.DestinationURL.String(),
			ShortURL: fmt.Sprintf(
				"https://%s/%s",
				urlWasDecoded.NonBrandedLink.Host.Hostname(),
				urlWasDecoded.NonBrandedLink.Slug.Value(),
			),
		}
		httpEncoder.EncodeResponse(writer, request, http.StatusOK, response)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr responseErrHTTP

	switch {
	case errors.Is(err, errValidation):
		apiErr = responseErrHTTP{Code: "ErrorValidation", Message: err.Error(), httpStatusCode: http.StatusBadRequest}
	case errors.Is(err, errApplication):
		apiErr = responseErrHTTP{
			Code:           "ErrorApplication",
			Message:        err.Error(),
			httpStatusCode: http.StatusUnprocessableEntity,
		}
	case errors.Is(err, errInfrastructure):
		apiErr = responseErrHTTP{
			Code:           "ErrorInfrastructure",
			Message:        err.Error(),
			httpStatusCode: http.StatusServiceUnavailable,
		}
	default:
		apiErr = responseErrHTTP{
			Code:           "UnknownError",
			Message:        err.Error(),
			httpStatusCode: http.StatusInternalServerError,
		}
	}

	httpEncoder.EncodeResponse(w, r, apiErr.httpStatusCode, apiErr)
}
