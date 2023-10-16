package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	logsConfig "github.com/zerok-ai/zk-utils-go/logs/config"
	storage "github.com/zerok-ai/zk-utils-go/storage/redis/config"
	"os"
)

type ServerConfig struct {
	Host string `yaml:"host" env:"SRV_HOST,HOST" env-description:"Server host" env-default:"localhost"`
	Port string `yaml:"port" env:"SRV_PORT,PORT" env-description:"Server port" env-default:"8080"`
}

// AppConfigs is an application configuration structure
type AppConfigs struct {
	Redis      storage.RedisConfig   `yaml:"redis"`
	Server     ServerConfig          `yaml:"server"`
	LogsConfig logsConfig.LogsConfig `yaml:"logs"`
}

// Args command-line parameters
type Args struct {
	ConfigPath string
}

// ProcessArgs processes and handles CLI arguments
func ProcessArgs() (*AppConfigs, error) {
	var a Args

	var cfg AppConfigs

	flagSet := flag.NewFlagSet("server", 1)
	flagSet.StringVar(&a.ConfigPath, "c", "config.yaml", "Path to configuration file")

	fu := flagSet.Usage
	flagSet.Usage = func() {
		fu()
		envHelp, _ := cleanenv.GetDescription(cfg, nil)
		if _, err := fmt.Fprintln(flagSet.Output()); err != nil {
			return
		}

		_, err := fmt.Fprintln(flagSet.Output(), envHelp)
		if err != nil {
			return
		}
	}

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	err := cleanenv.ReadConfig(a.ConfigPath, &cfg)
	return &cfg, err
}
