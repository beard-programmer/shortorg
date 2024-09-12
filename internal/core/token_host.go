package core

import "fmt"

type TokenHost interface {
	Hostname() string
}

func TokenHostFromString(host *string) (TokenHost, error) {
	if host == nil || *host == "" || *host == StandardTokenHost {
		return &TokenHostStandard{}, nil
	}

	return nil, fmt.Errorf("token host %v is not supported", *host)
}

const StandardTokenHost = "shortl.org"

type TokenHostStandard struct{}

func (t *TokenHostStandard) Hostname() string {
	return StandardTokenHost
}
