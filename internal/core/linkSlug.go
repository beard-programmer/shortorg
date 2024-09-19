package core

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/itchyny/base58-go"
)

type LinkSlug struct {
	value string
}

var pattern = regexp.MustCompile(`^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`)

func NewLinkSlug(s string) (*LinkSlug, error) {
	if len(s) != 6 {
		return nil, fmt.Errorf("value length must 6 characters, got %d", len(s))
	}

	if !pattern.MatchString(s) {
		return nil, errors.New("value contains invalid characters: must only contain Base58 characters")
	}

	return &LinkSlug{value: s}, nil
}

func NewLinkSlugFromLinkKey(k LinkKey) (*LinkSlug, error) {
	value := string(base58.BitcoinEncoding.EncodeUint64(uint64(k.Value())))
	return NewLinkSlug(value)
}

func (s LinkSlug) Value() string {
	return s.value
}
