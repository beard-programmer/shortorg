package apiServer

type Config struct {
	Host string     `toml:"host" envconfig:"HOST" default:"0.0.0.0"`
	HTTP HTTPConfig `toml:"http" envconfig:"HTTP"`
}

type HTTPConfig struct {
	InternalPort int `toml:"internal_port" envconfig:"INTERNAL_PORT" default:"8080"`
}
