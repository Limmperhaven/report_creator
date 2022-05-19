package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config struct {
	Timetta struct {
		Credentials struct {
			Email    string `yaml:"email"`
			Password string `yaml:"password"`
		}
		Settings struct {
			DateFrom     string `yaml:"dateFrom"`
			DateTo       string `yaml:"dateTo"`
			DocumentDate string `yaml:"documentDate"`
		}
	}
	Categories map[string]interface{} `yaml:"categories"`
	Projects   map[string]interface{} `yaml:"projects"`
}

func (c *Config) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

func ReadConfig() (*Config, error) {
	data, err := ioutil.ReadFile("etc/config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = cfg.Parse(data)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ParseFlags() string {
	output := flag.String("out", "", "Result document saving path")
	flag.Parse()

	return *output
}
