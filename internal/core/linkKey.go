package core

import (
	"fmt"

	"github.com/itchyny/base58-go"
)

const (
	minBase58Exp5 = 656356768   // 58^5
	maxBase58Exp6 = 38068692543 // 58^6
)

type LinkKey struct {
	value uint64
}

func NewLinkKey[T int64 | uint64](value T) (*LinkKey, error) {
	if value < minBase58Exp5 || maxBase58Exp6 <= value {
		return nil, fmt.Errorf(
			"%w NewLinkKey: value %d is out of range: must be included in %d .. %d",
			errValidation,
			value,
			minBase58Exp5,
			maxBase58Exp6-1,
		)
	}

	return &LinkKey{value: uint64(value)}, nil
}

// TODO: deprecated
func (k LinkKey) IntoLinkSlug() (*LinkSlug, error) {
	encoded := string(base58.BitcoinEncoding.EncodeUint64(k.value))
	return NewLinkSlug(encoded)
}

func (k LinkKey) Value() int64 {
	return int64(k.value) //nolint:gosec // its validated
}
