package config

import (
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type RedisConfig struct {
	Host        string `yaml:"host" env:"REDIS_HOST" env-description:"Database host"`
	Port        string `yaml:"port" env:"REDIS_PORT" env-description:"Database port"`
	DB          int    `yaml:"db" env:"REDIS_DB" env-description:"Database to load"`
	ReadTimeout int    `yaml:"readTimeout"`
}

type AppConfigs struct {
	Redis RedisConfig `yaml:"redis"`
}

// Args command-line parameters
type Args struct {
	ConfigPath string
}

// ProcessArgs processes and handles CLI arguments
func ProcessArgs(cfg interface{}) Args {
	var a Args

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
		return Args{}
	}
	return a
}
