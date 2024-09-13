package cache

type Config struct {
	MaxNumberOfElements int64 `config:"maxNumberOfElements"  default:"100000"`
	MaxMbSize           int64 `config:"maxMbSize"  default:"128"`
}
