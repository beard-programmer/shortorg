package infrastructure

type Config struct {
	PostgresClients postgresClientsConfig `mapstructure:"PostgresClients"`
	Cache           cacheConfig           `mapstructure:"Cache"`
	TokenStore      tokenStoreConfig      `mapstructure:"TokenStore"`
}

type postgresClientsConfig struct {
	TokenIdentifier postgresClientConfig `mapstructure:"TokenIdentifier"`
	ShortOrg        postgresClientConfig `mapstructure:"ShortOrg"`
}

type postgresClientConfig struct {
	User               string
	Password           string
	Host               string
	DBName             string
	Port               int
	MaxConnections     int
	MaxIdleConnections int
}

type cacheConfig struct {
	UseCache            bool
	MaxNumberOfElements int64
	MaxMbSize           int64
}

type tokenStoreConfig struct {
	BufferSize int
}
