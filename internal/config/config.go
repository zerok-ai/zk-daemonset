package config

import (
	logsConfig "github.com/zerok-ai/zk-utils-go/logs/config"
	storage "github.com/zerok-ai/zk-utils-go/storage/redis/config"
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
