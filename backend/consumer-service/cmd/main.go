package main

import (
	"consumer.local/config"
	"consumer.local/service"
)

func main() {
	cfg := config.GetConfig()

	service := service.NewConsumerService(&cfg)
	service.Start()
}
