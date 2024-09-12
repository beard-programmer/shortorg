package decode

type Decoder interface {
	Decode(string) (int64, error)
}

type UrlParser interface {
	Parse(string) (URL, error)
}

type URL interface {
	Scheme() string
	Hostname() string
	Path() string
	String() string
}
