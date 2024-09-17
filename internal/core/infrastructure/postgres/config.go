package postgres

type ClientsConfig struct {
	TokenIdentifier config `mapstructure:"TokenIdentifier"`
	ShortOrg        config `mapstructure:"ShortOrg"`
}

type config struct {
	User               string
	Password           string
	Host               string
	DBName             string
	Port               int
	MaxConnections     int
	MaxIdleConnections int
}
