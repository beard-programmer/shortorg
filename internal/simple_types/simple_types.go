package simple_types

import (
	"errors"
	"fmt"
	"regexp"
)

type IntegerBase58Exp5To6 struct {
	value int64
}

const (
	MinBase58Exp5 = 656356768   // 58^5
	MaxBase58Exp6 = 38068692543 // 58^6
)

func (IntegerBase58Exp5To6) New(value int64) (*IntegerBase58Exp5To6, error) {
	if value < MinBase58Exp5 || MaxBase58Exp6 <= value {
		return nil, fmt.Errorf("value %d is out of range: must be included in %d .. %d", value, MinBase58Exp5, MaxBase58Exp6-1)
	}

	return &IntegerBase58Exp5To6{value: value}, nil
}

func (i IntegerBase58Exp5To6) Value() int64 {
	return i.value
}

var Base58Pattern = regexp.MustCompile(`^[123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz]+$`)

type StringBase58Exp5To6 struct {
	value string
}

func (StringBase58Exp5To6) New(value string) (*StringBase58Exp5To6, error) {
	if len(value) != 6 {
		return nil, fmt.Errorf("value length must 6 characters, got %d", len(value))
	}

	if !Base58Pattern.MatchString(value) {
		return nil, errors.New("value contains invalid characters: must only contain Base58 characters")
	}

	return &StringBase58Exp5To6{value: value}, nil
}

func (s StringBase58Exp5To6) Value() string {
	return s.value
}
