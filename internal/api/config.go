package api

type Config struct {
	Host string
	HTTP configHTTP `mapstructure:"HTTP"`
}

type configHTTP struct {
	InternalPort int
}
