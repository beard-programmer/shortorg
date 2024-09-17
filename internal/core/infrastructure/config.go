package infrastructure

type Config struct {
	TokenStore tokenStoreConfig `mapstructure:"TokenStore"`
}

type tokenStoreConfig struct {
	BufferSize int
}
