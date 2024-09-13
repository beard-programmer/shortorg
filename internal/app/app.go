package app

import (
	"github.com/beard-programmer/shortorg/internal/core/infrastructure"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/cache"
	"github.com/beard-programmer/shortorg/internal/core/infrastructure/postgresClients"
	"github.com/beard-programmer/shortorg/internal/decode"
	"github.com/beard-programmer/shortorg/internal/encode"
	"go.uber.org/zap"
)

type App struct {
	logger          *zap.Logger
	postgresClients *postgresClients.Clients
	cache           *cache.InMemory[string]
	encodedUrlStore *infrastructure.EncodedUrlStore
	tokenKeyStore   *infrastructure.TokenKeyStore
	encodeFn        encode.Fn
	decodeFn        decode.Fn
	config          Config
}

func New(l *zap.Logger) *App {
	return &App{logger: l}
}

func (App) Name() string {
	return "shortorg"
}

type Config struct {
	//Vault             vault.VaultConfig     `envconfig:"VAULT"`
	//Server            apiServer.Config      `envconfig:"SERVER"`
	//ElasticClients    elasticClients.Config `envconfig:"ELASTIC_CLIENTS"`
	//Tracer            tracer.Config         `envconfig:"TRACER"`
	//ElasticRepository repository.Config     `envconfig:"ELASTIC_REPO"`
	postgresClientsConfig postgresClients.ClientsConfig
	cacheConfig           cache.Config
	Debug                 bool   `envconfig:"IS_DEBUG"`
	ENV                   string `envconfig:"ENV" default:"dev"`
}

func (c Config) IsProdEnv() bool {
	return c.ENV == "production"
}
