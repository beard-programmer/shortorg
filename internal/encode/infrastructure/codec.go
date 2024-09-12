package infrastructure

import (
	"github.com/itchyny/base58-go"
)

type CodecBase58 struct{}

func (CodecBase58) Encode(key int64) string {
	return string(base58.BitcoinEncoding.EncodeUint64(uint64(key)))
}

func (CodecBase58) Decode(token string) (int64, error) {
	result, err := base58.BitcoinEncoding.DecodeUint64([]byte(token))
	return int64(result), err
}
