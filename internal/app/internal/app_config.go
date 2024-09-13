package internal

import "github.com/beard-programmer/shortorg/internal/core/infrastructure/postgresClients"

type AppConfig struct {
	//Vault             vault.VaultConfig     `envconfig:"VAULT"`
	//Server            apiServer.Config      `envconfig:"SERVER"`
	//ElasticClients    elasticClients.Config `envconfig:"ELASTIC_CLIENTS"`
	//Tracer            tracer.Config         `envconfig:"TRACER"`
	//ElasticRepository repository.Config     `envconfig:"ELASTIC_REPO"`
	postgresClientsConfig postgresClients.Config
	Debug                 bool   `envconfig:"IS_DEBUG"`
	ENV                   string `envconfig:"ENV" default:"dev"`
}

func (c AppConfig) IsProdEnv() bool {
	return c.ENV == "production"
}
