package decode

import (
	"github.com/beard-programmer/shortorg/internal/base58"
)

type DecodingRequest interface {
	Url() string
}

type TokenKey = base58.IntegerExp5To6

type ValidatedRequest struct {
	TokenKey TokenKey
}

type ShortUrl struct {
	Path  string
	Token string
	Host  string
}

//      def self.from_uri(uri)
//        path = uri.path
//        to_token = SimpleTypes::StringExp5To6.from_string(uri.path[1..])
//
//        case [uri.scheme, uri.host, uri.domain, to_token.ok?]
//        in ['http', STANDARD_HOST, String, true] | ['https', STANDARD_HOST, String, true] if 6 < path.size
//          Result.ok new(path:, token: to_token.unwrap!.value)
//        else Result.err 'Not a standard short url.'
//        end
//      end
