package main

import (
	"github.com/ilyakaznacheev/cleanenv"
	"zerok-deamonset/internal/config"
	"zerok-deamonset/internal/detector"
)

func main() {

	// read configuration from the file and environment variables
	var cfg config.AppConfigs
	args := config.ProcessArgs(&cfg)
	if err := cleanenv.ReadConfig(args.ConfigPath, &cfg); err != nil {
		panic(err)
	}

	// start business logic
	detector.Start(cfg)
}
