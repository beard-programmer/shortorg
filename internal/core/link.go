package core

import (
	"fmt"
)

type NonBrandedLink struct {
	Key            LinkKey
	Slug           LinkSlug
	Host           LinkHost
	DestinationURL DestinationURL
}

func NewNonBrandedLink(linkKey LinkKey, linkHost LinkHost, destinationURL DestinationURL) (*NonBrandedLink, error) {
	switch linkHost.(type) {
	case *linkHostStandard:
		return NonBrandedLink{}.new(linkKey, linkHost, destinationURL)
	default:
		return nil, fmt.Errorf("NewNonBrandedLink: only standard hosts are supported, given: %v", linkHost)
	}
}

func (NonBrandedLink) new(
	linkKey LinkKey,
	linkHost LinkHost,
	destinationURL DestinationURL,
) (*NonBrandedLink, error) {
	linkSlug, err := NewLinkSlugFromLinkKey(linkKey)
	if err != nil {
		return nil, err
	}

	return &NonBrandedLink{linkKey, *linkSlug, linkHost, destinationURL}, nil
}
