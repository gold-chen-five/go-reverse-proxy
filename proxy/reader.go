package proxy

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
	Ssl    bool          `yaml:"ssl"`
	Host   string        `yaml:"host"`
	Routes []RouteConfig `yaml:"routes"`
}

type RouteConfig struct {
	Match RouteMatch  `yaml:"match"`
	Proxy ProxyConfig `yaml:"proxy"`
}

type RouteMatch struct {
	Path string `yaml:"path"`
}

type ProxyConfig struct {
	Upstream []string       `yaml:"upstream"`
	Strategy StrategyConfig `yaml:"strategy"`
}

type StrategyConfig struct {
	Type   Strategy               `yaml:"type"`
	Config map[string]interface{} `yaml:"config,omitempty"`
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

func (cfg *Config) GetAllDomains() []string {
	var domains []string
	for _, server := range cfg.Servers {
		domains = append(domains, server.Host)
	}

	return domains
}
