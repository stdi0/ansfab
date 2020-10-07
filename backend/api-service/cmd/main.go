package main

import (
	"api.local/config"
	"api.local/service"
)

func main() {
	cfg := config.GetConfig()

	service := service.NewApiService(&cfg)
	service.Start()
}
