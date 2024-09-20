package core

type URL interface {
	Scheme() string
	Hostname() string
	Path() string
	String() string
}

type URLParser interface {
	Parse(string) (URL, error)
}
