package postgresClients

type ClientsConfig struct {
	TokenIdentifier Config `toml:"token_identifier" envconfig:"TOKEN_IDENTIFIER"`
	ShortOrg        Config `toml:"short_org" envconfig:"SHORT_ORG"`
}

type Config struct {
	User               string `toml:"user" envconfig:"USER" default:"postgres"`
	Password           string `toml:"password" envconfig:"PASSWORD"`
	Host               string `toml:"host" envconfig:"HOST" default:"localhost"`
	DBName             string `toml:"dbname" envconfig:"DBNAME"`
	Port               int    `toml:"port" envconfig:"PORT"`
	MaxConnections     int    `toml:"max_connections" envconfig:"MAX_CONNECTIONS"`
	MaxIdleConnections int    `toml:"max_idle_connections" envconfig:"MAX_IDLE_CONNECTIONS"`
	MigrationsPath     string `toml:"migrations_path" envconfig:"MIGRATIONS_PATH"`
}
