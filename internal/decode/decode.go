package decode

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/base58"
	"github.com/beard-programmer/shortorg/internal/core"
	"go.uber.org/zap"
)

type UrlWasDecoded struct {
	Token core.TokenStandard
}

type OriginalUrlWasNotFound struct {
}

type Fn = func(context.Context, DecodingRequest) (*UrlWasDecoded, *OriginalUrlWasNotFound, error)

func NewDecodeFn(
	logger *zap.Logger,
	urlParser UrlParser,
	codec Codec,
	encodedUrlsProvider EncodedUrlsProvider,
	// urlWasEncodedChan chan<- UrlWasDecoded,
) Fn {
	return func(ctx context.Context, r DecodingRequest) (*UrlWasDecoded, *OriginalUrlWasNotFound, error) {
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
) (*UrlWasDecoded, *OriginalUrlWasNotFound, error) {
	validatedRequest, err := ValidatedRequest{}.New(urlParser, request)
	if err != nil {
		return nil, nil, err // validation
	}

	shortUrl := validatedRequest.ShortUrl

	tokenKey, err := shortUrl.KeyEncoded.Decode(codec)
	if err != nil {
		return nil, nil, err // vdaliton
	}

	url, err := encodedUrlsProvider.FindOne(ctx, *tokenKey)
	if err != nil {
		return nil, nil, err // inra
	}
	if url == "" {
		return nil, &OriginalUrlWasNotFound{}, nil
	}

	originalUrl, err := core.OriginalURLFromString(UrlParserAdapter{urlParser}, url)

	if err != nil {
		return nil, nil, err
	}

	token, err := core.NewToken(codec, *tokenKey, shortUrl.Host, *originalUrl)
	if err != nil {
		return nil, nil, err
	}

	return &UrlWasDecoded{*token}, nil, nil
}
