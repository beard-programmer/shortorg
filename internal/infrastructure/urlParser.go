package infrastructure

import (
	"fmt"
	"net/url"

	"github.com/beard-programmer/shortorg/internal/core"
)

type ParsingError struct {
	Message string
}

func (e ParsingError) Error() string {
	return e.Message
}

type parsedUrl struct {
	uri *url.URL
}

func (u parsedUrl) Scheme() string {
	return u.uri.Scheme
}

func (u parsedUrl) Hostname() string {
	return u.uri.Hostname()
}

func (u parsedUrl) Path() string {
	return u.uri.Path
}

func (u parsedUrl) String() string {
	return u.uri.String()

}

type UrlParser struct{}

func (UrlParser) Parse(urlString string) (core.URL, error) {
	if urlString == "" {
		return nil, ParsingError{Message: "parsedUrl must be a non-empty string."}
	}

	if len(urlString) > 2048 {
		return nil, ParsingError{Message: "parsedUrl too long."}
	}

	parsed, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, ParsingError{Message: fmt.Sprintf("Failed to parse parsedUrl: %v", err)}
	}

	return &parsedUrl{uri: parsed}, nil
}
