package core

import (
	"fmt"
)

type Link struct {
	Key            LinkKey
	Slug           LinkSlug
	Host           LinkHost
	DestinationURL DestinationURL
}

func NewLink(linkKey LinkKey, linkHost LinkHost, destinationURL DestinationURL) (*Link, error) {
	if destinationURL.Hostname() == linkHost.Hostname() {
		return nil, fmt.Errorf("%w NewLink: destination url cannot have same host as link", errValidation)
	}

	linkSlug, err := linkKey.IntoLinkSlug()
	if err != nil {
		return nil, err
	}
	return &Link{Key: linkKey, Slug: *linkSlug, Host: linkHost, DestinationURL: destinationURL}, nil
}

type LinkDTO struct {
	Key            LinkKeyDto
	Slug           LinkSlugDto
	Host           LinkHostDto
	DestinationURL URLDto
}

func (l *Link) IntoDto() LinkDTO {
	return LinkDTO{
		Key:            l.Key.IntoDto(),
		Slug:           l.Slug.IntoDto(),
		Host:           l.Host.IntoDto(),
		DestinationURL: l.DestinationURL.IntoDto(),
	}
}

func (dto LinkDTO) IntoDomain() (*Link, error) {
	key, err := dto.Key.IntoDomain()
	if err != nil {
		return nil, err
	}

	slug, err := dto.Slug.IntoDomain()
	if err != nil {
		return nil, err
	}

	host, err := dto.Host.IntoDomain()
	if err != nil {
		return nil, err
	}

	destinationURL, err := dto.DestinationURL.IntoDomain()
	if err != nil {
		return nil, err
	}

	return &Link{
		Key:            *key,
		Slug:           *slug,
		Host:           *host,
		DestinationURL: *destinationURL,
	}, nil
}
