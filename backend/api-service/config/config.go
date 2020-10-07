package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

type ApiServiceConfig struct {
	WebListenPort     int32  `env:"LISTEN_PORT" envDefault:"9999"`
	WebDomain         string `env:"WEB_DOMAIN" envDefault:"https://rrd.rzd"`
	FabricServiceHost string `env:"FABRIC_SERVICE_HOST,required"`
	FabricServicePort int32  `env:"FABRIC_SERVICE_PORT" envDefault:"5400"`
}

func GetConfig() ApiServiceConfig {
	cfg := &ApiServiceConfig{}
	if err := env.Parse(cfg); err != nil {
		log.Fatal("Cannot parse initial ENV vars: ", err)
	}
	return *cfg
}
