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
	value int64
}

func NewLinkKey(value int64) (*LinkKey, error) {
	if value < minBase58Exp5 || maxBase58Exp6 <= value {
		return nil, fmt.Errorf(
			"value %d is out of range: must be included in %d .. %d", value,
			minBase58Exp5, maxBase58Exp6-1,
		)
	}

	//		return nil, false, fmt.Errorf("%w: failed to build token %v", errApplication, err)

	return &LinkKey{value: value}, nil
}

func NewLinkKeyFromLinkSlug(slug LinkSlug) (*LinkKey, error) {
	value, err := base58.BitcoinEncoding.DecodeUint64([]byte(slug.Value()))
	if err != nil {
		return nil, err
	}
	return NewLinkKey(int64(value))
}

func (k LinkKey) Value() int64 {
	return k.value
}
