package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger     LoggerConf     `yaml:"logger"`
	Storage    StorageConf    `yaml:"storage"`
	SQL        SQLConf        `yaml:"sql"`
	AMQPClient AMQPClientConf `yaml:"amqpClient"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type StorageConf struct {
	Type string `yaml:"type"` // memory | sql
}

type SQLConf struct {
	DSN string `yaml:"dsn"`
}

type AMQPClientConf struct {
	URL          string    `yaml:"url"`
	Queue        QueueConf `yaml:"queue"`
	PollInterval int       `yaml:"pollInterval"`
}

type QueueConf struct {
	Name string `yaml:"name"`
}

func NewConfig(configFile string) (Config, error) {
	cfg, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(cfg, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
