package encode

import (
	"github.com/beard-programmer/shortorg/internal/simple_types"
)

type TokenEncoded = simple_types.StringBase58Exp5To6

func NewTokenEncoded(value string) (*TokenEncoded, error) {
	return simple_types.NewStringBase58Exp5To6(value)
}
