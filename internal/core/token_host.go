package core

import "fmt"

type TokenHost interface {
	Hostname() string
}

func TokenHostFromString(host *string) (TokenHost, error) {
	if host == nil || *host == "" || *host == standardTokenHost {
		return &tokenHostStandard{}, nil
	}

	return nil, fmt.Errorf("token host %v is not supported", *host)
}

const standardTokenHost = "shortl.org"

type tokenHostStandard struct{}

func (t *tokenHostStandard) Hostname() string {
	return standardTokenHost
}
