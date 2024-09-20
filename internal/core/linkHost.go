package core

import "fmt"

type LinkHost struct {
	hostname  string
	isBranded bool
}

func (h LinkHost) IsBranded() bool {
	return h.isBranded
}

func (h LinkHost) Hostname() string {
	return h.hostname
}

func NewLinkHost(host *string) (*LinkHost, error) {
	if host == nil || *host == "" || *host == DefaultLinkHost {
		return &LinkHost{hostname: DefaultLinkHost, isBranded: false}, nil
	}

	return nil, fmt.Errorf("link host %v is not supported", *host)
}

const DefaultLinkHost = "shortl.org"
