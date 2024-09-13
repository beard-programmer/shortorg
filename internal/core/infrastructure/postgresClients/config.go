package postgresClients

type ClientsConfig struct {
	TokenIdentifier Config
	ShortOrg        Config
}

type Config struct {
	User                string `json:"user"`
	Password            string `json:"password"`
	Host                string `json:"host"`
	DBName              string `json:"dbname"`
	Port                int    `json:"port"`
	MaxConnections      int    `json:"maxConnections"`
	MaxIddleConnections int    `json:"minConnections"`
	MigrationsPath      string `json:"migrationsPath"`
}
