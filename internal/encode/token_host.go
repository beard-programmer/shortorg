package encode

import (
	"fmt"
)

type TokenHost interface {
	Domain() string
	Host() string
}

func TokenHostFromString(host *string) (TokenHost, error) {
	if host == nil || *host == "" || *host == StandardTokenHost {
		return &TokenHostStandard{}, nil
	}

	return nil, fmt.Errorf("not allowed to encode URL using host %s", *host)

}
