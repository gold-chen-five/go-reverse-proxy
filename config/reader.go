package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config struct to hold the settings from settings.yaml
type Config struct {
	Servers []ServerConfig `yaml:"servers"`
}

type ServerConfig struct {
	Listen string        `yaml:"listen"`
	Routes []RouteConfig `yaml:"routes"`
}

type RouteConfig struct {
	Match RouteMatch  `yaml:"match"`
	Proxy ProxyConfig `yaml:"proxy"`
}

type RouteMatch struct {
	Host string `yaml:"host"`
	Path string `yaml:"path"`
}

type ProxyConfig struct {
	Upstream []string `yaml:"upstream"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
