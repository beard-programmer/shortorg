package core

type Encoder interface {
	Encode(int64) string
}

type Decoder interface {
	Decode(string) (int64, error)
}
