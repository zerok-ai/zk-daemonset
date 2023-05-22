package main

import (
	"fmt"
	"zk-daemonset/internal/config"
	"zk-daemonset/internal/detector"
)

func main() {

	fmt.Printf("Hello from daemonset\n")

	// read configuration from the file and environment variables
	var cfg config.AppConfigs
	if err := config.ProcessArgs(&cfg); err != nil {
		panic(err)
	}

	fmt.Printf("config redis host:%s \n", cfg.Redis.Host)

	// start business logic
	if err := detector.Start(cfg); err != nil {
		panic(err)
	}
}
