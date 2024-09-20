package encode

import (
	"context"
	"errors"
	"fmt"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/core"
)

type URLWasEncoded struct {
	NonBrandedLink core.NonBrandedLink
}

var (
	errValidation     = errors.New("validation")
	errInfrastructure = errors.New("infrastructure")
	errApplication    = errors.New("application")
)

type Fn = func(context.Context, EncodingRequest) (*URLWasEncoded, error)

func NewEncodeFn(
	tokenKeyStore LinkKeyStore,
	logger *appLogger.AppLogger,
	urlWasEncodedChan chan<- URLWasEncoded,
) Fn {
	return func(ctx context.Context, r EncodingRequest) (*URLWasEncoded, error) {
		return encode(ctx, tokenKeyStore, logger, urlWasEncodedChan, r)
	}
}

func encode(
	ctx context.Context,
	linkKeyStore LinkKeyStore,
	logger *appLogger.AppLogger,
	urlWasEncodedChan chan<- URLWasEncoded,
	request EncodingRequest,
) (*URLWasEncoded, error) {
	validatedRequest, err := NewValidatedRequest(
		request,
	)

	if err != nil {
		return nil, fmt.Errorf("%w: encode: %v", errValidation, err)
	}

	unclaimedKey, err := linkKeyStore.Issue(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: encode: failed to generate unclaimedKey: %v", errInfrastructure, err)
	}

	token, err := core.NewNonBrandedLink(*unclaimedKey, validatedRequest.TokenHost, validatedRequest.OriginalURL)

	if err != nil {
		return nil, fmt.Errorf("%w: encode: failed to build non branded link: %v", errApplication, err)
	}

	event := URLWasEncoded{*token}
	go func() {
		urlWasEncodedChan <- event
	}()

	return &event, nil
}
