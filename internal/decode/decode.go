package decode

import (
	"context"
	"errors"
	"fmt"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/core"
)

type urlWasDecoded struct {
	Token core.NonBrandedLink
}

var (
	errValidation     = errors.New("validation")
	errInfrastructure = errors.New("infrastructure")
	errApplication    = errors.New("application")
)

type Fn = func(context.Context, decodingRequest) (*urlWasDecoded, bool, error)

func NewDecodeFn(logger *appLogger.AppLogger, urlParser core.URLParser, encodedUrlsProvider EncodedUrlsProvider) Fn {
	return func(ctx context.Context, r decodingRequest) (*urlWasDecoded, bool, error) {
		return decode(ctx, logger, urlParser, encodedUrlsProvider, r)
	}
}

func decode(
	ctx context.Context,
	l *appLogger.AppLogger,
	urlParser core.URLParser,
	encodedUrlsProvider EncodedUrlsProvider,
	request decodingRequest,
) (*urlWasDecoded, bool, error) {
	validatedRequest, err := newValidatedRequest(urlParser, request)
	if err != nil {
		return nil, false, err // validation
	}

	shortURL := validatedRequest.ShortURL

	tokenKey, err := core.NewLinkKeyFromLinkSlug(shortURL.linkSlug)
	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to validate request: %v", errValidation, err)
	}

	url, isFound, err := encodedUrlsProvider.FindOne(ctx, *tokenKey)
	if err != nil {
		return nil, isFound, fmt.Errorf("%w: failed to generate unclaimedKey %v", errInfrastructure, err)
	}
	if !isFound {
		return nil, isFound, nil
	}

	originalURL, err := core.DestinationURLFromString(urlParser, url)

	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to parse original url from storage %v", errApplication, err)
	}

	token, err := core.NewNonBrandedLink(*tokenKey, shortURL.linkHost, *originalURL)
	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to build token %v", errApplication, err)
	}

	return &urlWasDecoded{*token}, true, nil
}
