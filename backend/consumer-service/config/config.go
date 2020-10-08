package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type ConsumerServiceConfig struct {
	ServicePort       int32  `env:"SERVICE_PORT" envDefault:"9993"`
	QueueName         string `env:"QUEUE_NAME,required"`
	FabricServiceHost string `env:"FABRIC_SERVICE_HOST,required"`
	FabricServicePort int32  `env:"FABRIC_SERVICE_PORT" envDefault:"5400"`
}

func GetConfig() ConsumerServiceConfig {
	cfg := &ConsumerServiceConfig{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal("Cannot parse initial ENV vars: ", err)
	}
	return *cfg
}
