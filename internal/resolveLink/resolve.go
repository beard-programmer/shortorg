package resolveLink

import (
	"context"
	"errors"
	"fmt"

	appLogger "github.com/beard-programmer/shortorg/internal/app/logger"
	"github.com/beard-programmer/shortorg/internal/core"
)

type linkWasResolvedEvent struct {
	NonBrandedLink core.Link
}

var (
	errValidation     = errors.New("validation")
	errInfrastructure = errors.New("infrastructure")
	errApplication    = errors.New("application")
)

type ResolveLinkFn = func(context.Context, resolveLinkRequest) (*linkWasResolvedEvent, bool, error)

func NewResolveLinkFn(logger *appLogger.AppLogger, encodedUrlsProvider LinksStore) ResolveLinkFn {
	return func(ctx context.Context, r resolveLinkRequest) (*linkWasResolvedEvent, bool, error) {
		return resolveLink(ctx, logger, encodedUrlsProvider, r)
	}
}

func resolveLink(
	ctx context.Context,
	l *appLogger.AppLogger,
	linksStore LinksStore,
	request resolveLinkRequest,
) (*linkWasResolvedEvent, bool, error) {
	validatedRequest, err := newValidatedRequest(request)
	if err != nil {
		return nil, false, err // validation
	}

	shortURL := validatedRequest.ShortURL

	tokenKey, err := shortURL.linkSlug.IntoLinkKey()
	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to validate request: %v", errValidation, err)
	}

	dto, isFound, err := linksStore.FindOneNonBrandedLink(
		ctx,
		shortURL.linkSlug.IntoDto(),
		tokenKey.IntoDto(),
		shortURL.linkHost.IntoDto(),
	)

	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to resolve link: %v", errInfrastructure, err)
	}

	if !isFound {
		return nil, false, nil
	}

	link, err := dto.IntoDomain()

	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to resolve link: %v", errApplication, err)
	}

	//url, isFound, err := linksStore.FindOne(ctx, *tokenKey)
	//if err != nil {
	//	return nil, isFound, fmt.Errorf("%w: failed to generate unclaimedKey %v", errInfrastructure, err)
	//}
	//if !isFound {
	//	return nil, isFound, nil
	//}
	//
	//originalURL, err := core.NewURL(url)
	//
	//if err != nil {
	//	return nil, false, fmt.Errorf("%w: failed to parse original url from storage %v", errApplication, err)
	//}
	//
	//token, err := core.NewLink(*tokenKey, shortURL.linkHost, *originalURL)
	//if err != nil {
	//	return nil, false, fmt.Errorf("%w: failed to build token %v", errApplication, err)
	//}

	return &linkWasResolvedEvent{*link}, true, nil
}
