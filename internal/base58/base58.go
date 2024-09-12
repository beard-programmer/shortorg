package base58

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/itchyny/base58-go"
)

type Codec struct{}

func (Codec) Encode(key int64) string {
	return string(base58.BitcoinEncoding.EncodeUint64(uint64(key)))
}

func (Codec) Decode(token string) (int64, error) {
	result, err := base58.BitcoinEncoding.DecodeUint64([]byte(token))
	return int64(result), err
}

type IntegerExp5To6 struct {
	value int64
}

const (
	MinBase58Exp5 = 656356768   // 58^5
	MaxBase58Exp6 = 38068692543 // 58^6
)

func (IntegerExp5To6) New(value int64) (*IntegerExp5To6, error) {
	if value < MinBase58Exp5 || MaxBase58Exp6 <= value {
		return nil, fmt.Errorf("value %d is out of range: must be included in %d .. %d", value, MinBase58Exp5, MaxBase58Exp6-1)
	}

	return &IntegerExp5To6{value: value}, nil
}

func (i IntegerExp5To6) Value() int64 {
	return i.value
}

var Pattern = regexp.MustCompile(`^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`)

type StringExp5To6 struct {
	value string
}

func (StringExp5To6) New(value string) (*StringExp5To6, error) {
	if len(value) != 6 {
		return nil, fmt.Errorf("value length must 6 characters, got %d", len(value))
	}

	if !Pattern.MatchString(value) {
		return nil, errors.New("value contains invalid characters: must only contain Base58 characters")
	}

	return &StringExp5To6{value: value}, nil
}

func (s StringExp5To6) Value() string {
	return s.value
}
