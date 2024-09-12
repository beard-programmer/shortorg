package decode

import (
	"context"

	"github.com/beard-programmer/shortorg/internal/base58"
	"go.uber.org/zap"
)

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

type UnclaimedKey = base58.IntegerExp5To6

func decode(
	ctx context.Context,
	logger *zap.SugaredLogger,
	//urlWasEncodedChan chan<- UrlWasEncoded,
	request DecodingRequest,
) (*UrlWasDecoded, error) {

	//      validate_request = RequestValidated.from_unvalidated_request(Infrastructure.method(:parse_url_string), request:)
	//
	//      to_token_identifier = validate_request.and_then do |validated_request|
	//        UrlManagement::TokenIdentifier.from_string(Infrastructure.codec_base58.method(:decode),
	//                                                   validated_request.short_url.token)
	//      end
	//
	//      find_encoded_url = to_token_identifier.and_then do |token_identifier|
	//        Infrastructure.find_encoded_url(db, token_identifier)
	//      end
	//      case find_encoded_url
	//      in Result::Ok[encoded_url_string]
	//        url = UrlManagement::OriginalUrl.from_string(Infrastructure.method(:parse_url_string),
	//                                                     encoded_url_string).unwrap!.to_s
	//        short_url = validate_request.unwrap!.short_url
	//        Result.ok ShortUrlDecoded.new(url:, short_url_host: short_url.host, short_url_token: short_url.token)
	//      in Result::Ok[] then Result.ok OriginalWasNotFound.new(request.short_url)
	//      in Result::Err[Infrastructure::DatabaseError => e] then Result.err InfrastructureError.new(e)
	//      in Result::Err[e] then Result.err ValidationError.new(e)
	//      else raise "Unexpected response when fetching encoded url."
	//      end
	return &UrlWasDecoded{}, nil
}
