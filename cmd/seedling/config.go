package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version string `yaml:"version"`

	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`

	Output struct {
		Format    string `yaml:"format"`
		File      string `yaml:"file"`
		BatchSize int    `yaml:"batch_size"`
		UseCopy   bool   `yaml:"use_copy"`
	} `yaml:"output"`

	Generation struct {
		Count    int   `yaml:"count"`
		Seed     int64 `yaml:"seed"`
		Truncate bool  `yaml:"truncate"`
		Parallel bool  `yaml:"parallel"`
		DryRun   bool  `yaml:"dry_run"`
		Verbose  bool  `yaml:"verbose"`
	} `yaml:"generation"`

	Schema struct {
		File string `yaml:"file"`
	} `yaml:"schema"`

	Generators struct {
		File string `yaml:"file"`
	} `yaml:"generators"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	expanded := os.Expand(string(data), os.Getenv)

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
