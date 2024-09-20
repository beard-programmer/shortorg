package core

import (
	"fmt"
	"regexp"

	"github.com/itchyny/base58-go"
)

type LinkSlug struct {
	value string
}

var pattern = regexp.MustCompile(`^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`)

const minSlugSize = 6

func NewLinkSlug(s string) (*LinkSlug, error) {
	if len(s) != minSlugSize {
		return nil, fmt.Errorf("%w NewLinkSlug: value length must 6 characters, got %d", errValidation, len(s))
	}

	if !pattern.MatchString(s) {
		return nil,
			fmt.Errorf(
				"%w NewLinkSlug: %s value contains invalid characters: must only contain Base58 characters",
				errValidation,
				s,
			)
	}

	return &LinkSlug{value: s}, nil
}

func (s LinkSlug) Value() string {
	return s.value
}

// TODO: deprecated
func (s LinkSlug) IntoLinkKey() (*LinkKey, error) {
	value, err := base58.BitcoinEncoding.DecodeUint64([]byte(s.value))
	if err != nil {
		return nil, err
	}
	return NewLinkKey(value)
}
