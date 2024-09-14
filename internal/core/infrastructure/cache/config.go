package cache

type Config struct {
	MaxNumberOfElements int64 `toml:"max_number_of_elements" envconfig:"MAX_NUMBER_OF_ELEMENTS"`
	MaxMbSize           int64 `toml:"max_mb_size" envconfig:"MAX_MB_SIZE"`
}
