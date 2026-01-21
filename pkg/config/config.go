package config

import (
	"os"
	"path/filepath"

	"github.com/modern-tooling/aloc/internal/model"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Overrides map[model.Role][]string `yaml:"overrides"`
	Exclude   []string                `yaml:"exclude"`
	Options   Options                 `yaml:"options"`
}

type Options struct {
	HeaderProbe  bool `yaml:"header_probe"`
	Neighborhood bool `yaml:"neighborhood"`
}

func DefaultConfig() *Config {
	return &Config{
		Overrides: nil,
		Exclude: []string{
			"vendor/**",
			"node_modules/**",
			".git/**",
		},
		Options: Options{
			HeaderProbe:  false,
			Neighborhood: true,
		},
	}
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

func LoadFromDir(dir string) (*Config, error) {
	candidates := []string{
		filepath.Join(dir, "aloc.yaml"),
		filepath.Join(dir, "aloc.yml"),
		filepath.Join(dir, ".aloc.yaml"),
		filepath.Join(dir, ".aloc.yml"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return Load(path)
		}
	}

	return DefaultConfig(), nil
}
