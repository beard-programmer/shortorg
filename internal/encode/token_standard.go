package encode

type TokenStandard struct {
	TokenIdentifier TokenIdentifier
	TokenHost       TokenHost
	TokenEncoded    TokenEncoded
}

func (_ TokenStandard) New(codec Codec, tokenIdentifier TokenIdentifier, tokenHost TokenHost) (*TokenStandard, error) {
	tokenEncoded, err := new(TokenEncoded).FromString(codec.Encode(tokenIdentifier.Value()))
	if err != nil {
		return nil, err
	}

	return &TokenStandard{tokenIdentifier, tokenHost, *tokenEncoded}, nil
}
