package encode

type TokenStandard struct {
	Identity     Identity
	TokenHost    TokenHost
	TokenEncoded TokenEncoded
	OriginalURL  OriginalURL
}

func NewTokenStandard(codec CodecProvider, identity Identity, tokenHost TokenHost, originalUrl OriginalURL) (*TokenStandard, error) {
	tokenEncoded, err := NewTokenEncoded(codec.Encode(identity.Value()))
	if err != nil {
		return nil, err
	}

	return &TokenStandard{identity, tokenHost, *tokenEncoded, originalUrl}, nil
}
