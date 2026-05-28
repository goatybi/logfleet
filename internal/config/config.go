package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Servers []Server `yaml:"servers"`
}

type Server struct {
	Name    string   `yaml:"name"`
	Host    string   `yaml:"host"`
	Port    int      `yaml:"port"`
	User    string   `yaml:"user"`
	KeyPath string   `yaml:"key"`
	Logs    []string `yaml:"logs"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config parse error: %w", err)
	}

	for i := range cfg.Servers {
		if cfg.Servers[i].Port == 0 {
			cfg.Servers[i].Port = 22
		}
		if cfg.Servers[i].User == "" {
			cfg.Servers[i].User = "root"
		}
	}

	return &cfg, nil
}

func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return home + "/.logfleet/config.yaml"
}

func ExampleConfig() string {
	return `servers:
  - name: prod-web
    host: 1.2.3.4
    port: 22
    user: root
    key: ~/.ssh/id_rsa
    logs:
      - /var/log/nginx/access.log
      - /var/log/nginx/error.log
      - /var/log/syslog

  - name: prod-db
    host: 5.6.7.8
    user: ubuntu
    key: ~/.ssh/id_rsa
    logs:
      - /var/log/postgresql/postgresql-16-main.log
`
}