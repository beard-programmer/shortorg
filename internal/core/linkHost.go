package core

import "fmt"

type LinkHost interface {
	Hostname() string
}

func LinkHostFromString(host *string) (LinkHost, error) {
	if host == nil || *host == "" || *host == standardTokenHost {
		return &linkHostStandard{}, nil
	}

	return nil, fmt.Errorf("token host %v is not supported", *host)
}

const standardTokenHost = "shortl.org"

type linkHostStandard struct{}

func (t *linkHostStandard) Hostname() string {
	return standardTokenHost
}
