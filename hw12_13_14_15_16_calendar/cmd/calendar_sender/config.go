package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger     LoggerConf     `yaml:"logger"`
	AMQPClient AMQPClientConf `yaml:"amqpClient"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type AMQPClientConf struct {
	URL   string    `yaml:"url"`
	Queue QueueConf `yaml:"queue"`
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
