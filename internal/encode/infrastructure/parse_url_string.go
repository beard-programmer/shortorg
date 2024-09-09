package infrastructure

import (
	"fmt"
	"net/url"
)

type ParsingError struct {
	Message string
}

func (e ParsingError) Error() string {
	return e.Message
}

type ParsedUrl struct {
	uri *url.URL
}

func (u ParsedUrl) Scheme() string {
	return u.uri.Scheme
}

func (u ParsedUrl) Hostname() string {
	return u.uri.Hostname()
}

func (u ParsedUrl) String() string {
	return u.uri.String()

}

func ParseURLString(urlString string) (*ParsedUrl, error) {
	if urlString == "" {
		return nil, ParsingError{Message: "ParsedUrl must be a non-empty string."}
	}

	if len(urlString) > 2048 {
		return nil, ParsingError{Message: "ParsedUrl too long."}
	}

	parsed, err := url.ParseRequestURI(urlString)
	if err != nil {
		return nil, ParsingError{Message: fmt.Sprintf("Failed to parse ParsedUrl: %v", err)}
	}

	return &ParsedUrl{uri: parsed}, nil
}
