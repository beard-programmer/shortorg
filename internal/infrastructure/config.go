package infrastructure

type Config struct {
	PostgresClients postgresClientsConfig `mapstructure:"PostgresClients"`
	Cache           cacheConfig           `mapstructure:"Cache"`
	TokenStore      tokenStoreConfig      `mapstructure:"TokenStore"`
}

type tokenStoreConfig struct {
	BufferSize int
}

type cacheConfig struct {
	UseCache            bool
	MaxNumberOfElements int64
	MaxMbSize           int64
}

type postgresClientsConfig struct {
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
