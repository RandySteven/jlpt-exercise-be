package configs

import (
	"os"

	"gopkg.in/yaml.v2"
)

// NewConfig creates a new Config instance by reading and parsing a YAML configuration file.
// It accepts a file path and returns the parsed configuration or an error if the file
// cannot be opened or parsed.
//
// Example:
//
//	cfg, err := config.NewConfig("./config.yml")
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewConfig(configPath string) (*Config, error) {
	config := &Config{}
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}
