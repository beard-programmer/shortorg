package decode

import (
	"context"

	"go.uber.org/zap"
)

type DecodingRequest interface {
	Url() string
}

type UrlWasDecoded struct {
}

func NewDecodeFunc(

	logger *zap.SugaredLogger,
	// urlWasEncodedChan chan<- UrlWasEncoded,
) func(context.Context, DecodingRequest) (*UrlWasDecoded, error) {
	return func(ctx context.Context, r DecodingRequest) (*UrlWasDecoded, error) {
		return decode(
			ctx, logger,
			//urlWasEncodedChan,
			r,
		)
	}
}

func decode(
	ctx context.Context,
	logger *zap.SugaredLogger,
	//urlWasEncodedChan chan<- UrlWasEncoded,
	request DecodingRequest,
) (*UrlWasDecoded, error) {
	return &UrlWasDecoded{}, nil
}
