package apiServer

type Config struct {
	Host string     `toml:"host" envconfig:"HOST" default:"0.0.0.0"`
	HTTP HTTPConfig `toml:"http" envconfig:"HTTP"`
}

//type Config struct {
//    Host string     `json:"host" envconfig:"HOST" default:"0.0.0.0"`
//    HTTP HTTPConfig `json:"http" envconfig:"HTTP"`
//}
//
//type HTTPConfig struct {
//    InternalPort int `json:"internal_port" envconfig:"INTERNAL_PORT" default:"8080"`
//}

type HTTPConfig struct {
	InternalPort int `toml:"internal_port" envconfig:"INTERNAL_PORT" default:"8080"`
}
