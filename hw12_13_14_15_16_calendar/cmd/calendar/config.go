package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger  LoggerConf  `yaml:"logger"`
	Server  ServerConf  `yaml:"server"`
	Storage StorageConf `yaml:"storage"`
	SQL     SQLConf     `yaml:"sql"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type HTTPServer struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type GRPCServer struct {
	Port string `yaml:"port"`
}

type ServerConf struct {
	HTTPServer `yaml:"http"`
	GRPCServer `yaml:"grpc"`
}

type StorageConf struct {
	Type string `yaml:"type"` // memory | sql
}

type SQLConf struct {
	DSN string `yaml:"dsn"`
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
