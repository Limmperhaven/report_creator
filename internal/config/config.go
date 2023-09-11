package config

import (
	"os"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

type (
	Config struct {
		Timetta      Timetta             `yaml:"timetta"`
		Employees    []Employee          `yaml:"employees"`
		Projects     []Project           `yaml:"projects"`
		Output       string              `yaml:"-"`
		EmployeeById map[string]Employee `yaml:"-"`
		ProjectById  map[string]Project  `yaml:"-"`
	}
	Timetta struct {
		Credentials Credentials `yaml:"credentials"`
		Settings    Settings    `yaml:"settings"`
	}
	Credentials struct {
		Email    string `yaml:"email"`
		Password string `yaml:"password"`
	}
	Project struct {
		Id        string `yaml:"id"`
		Name      string `yaml:"name"`
		Initiator string `yaml:"initiator"`
		Code      string `yaml:"code"`
	}
	Settings struct {
		DateFrom     string `yaml:"dateFrom"`
		DateTo       string `yaml:"dateTo"`
		DocumentDate string `yaml:"documentDate"`
	}
	Employee struct {
		Id       string `yaml:"id"`
		Fio      string `yaml:"fio"`
		Category string `yaml:"category"`
		Salary   int    `yaml:"salary"`
	}
)

func (c *Config) Parse(data []byte) error {
	return yaml.Unmarshal(data, c)
}

func ReadConfig(cfgPath, outPath string) (*Config, error) {
	pflag.Parse()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var cfg Config

	err = cfg.Parse(data)
	if err != nil {
		return nil, err
	}

	cfg.Output = outPath
	cfg.ProjectById = make(map[string]Project, len(cfg.Projects))
	for _, project := range cfg.Projects {
		cfg.ProjectById[project.Id] = project
	}
	cfg.EmployeeById = make(map[string]Employee, len(cfg.Employees))
	for _, employee := range cfg.Employees {
		cfg.EmployeeById[employee.Id] = employee
	}
	return &cfg, nil
}
