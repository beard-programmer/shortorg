package decode

import (
	"context"
	"errors"
	"fmt"

	"github.com/beard-programmer/shortorg/internal/base58"
	"github.com/beard-programmer/shortorg/internal/core"
	"go.uber.org/zap"
)

type UrlWasDecoded struct {
	Token core.TokenStandard
}

type OriginalUrlWasNotFound struct {
}

var (
	ValidationError     = errors.New("validation")
	InfrastructureError = errors.New("infrastructure")
	ApplicationError    = errors.New("application")
)

type Fn = func(context.Context, DecodingRequest) (*UrlWasDecoded, bool, error)

func NewDecodeFn(
	logger *zap.Logger,
	urlParser UrlParser,
	codec Codec,
	encodedUrlsProvider EncodedUrlsProvider,
	// urlWasEncodedChan chan<- UrlWasDecoded,
) Fn {
	return func(ctx context.Context, r DecodingRequest) (*UrlWasDecoded, bool, error) {
		return decode(
			ctx,
			logger,
			urlParser,
			codec,
			encodedUrlsProvider,
			//urlWasEncodedChan,
			r,
		)
	}
}

type UnclaimedKey = base58.IntegerExp5To6

func decode(
	ctx context.Context,
	_ *zap.Logger,
	urlParser UrlParser,
	codec Codec,
	encodedUrlsProvider EncodedUrlsProvider,
	//_ chan<- UrlWasDecoded,
	request DecodingRequest,
) (*UrlWasDecoded, bool, error) {
	validatedRequest, err := ValidatedRequest{}.New(urlParser, request)
	if err != nil {
		return nil, false, err // validation
	}

	shortUrl := validatedRequest.ShortUrl

	tokenKey, err := shortUrl.KeyEncoded.Decode(codec)
	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to validate request: %v", ValidationError, err)
	}

	url, isFound, err := encodedUrlsProvider.FindOne(ctx, *tokenKey)
	if err != nil {
		return nil, isFound, fmt.Errorf("%w: failed to generate unclaimedKey %v", InfrastructureError, err)
	}
	if !isFound {
		return nil, isFound, nil
	}

	originalUrl, err := core.OriginalURLFromString(UrlParserAdapter{urlParser}, url)

	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to parse original url from storage %v", ApplicationError, err)
	}

	token, err := core.NewToken(codec, *tokenKey, shortUrl.Host, *originalUrl)
	if err != nil {
		return nil, false, fmt.Errorf("%w: failed to build tokene %v", ApplicationError, err)
	}

	return &UrlWasDecoded{*token}, true, nil
}
