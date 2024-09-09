package encode

const StandardTokenHost = "short.est"

// TokenHostStandard represents the default host for the shortened URL
type TokenHostStandard struct{}

func (t *TokenHostStandard) Host() string {
	return StandardTokenHost
}

func (t *TokenHostStandard) Domain() string {
	return StandardTokenHost
}
