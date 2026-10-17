package configs

import (
	"flag"
	"fmt"
	"os"
)

// ParseFlags parses command-line flags and returns the configuration file path.
// It defaults to "./config.yml" if no --config flag is provided.
// Returns an error if the config path is invalid.
func ParseFlags() (string, error) {
	var configPath string
	if flag.Lookup("config") == nil {
		flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
		flag.Parse()
	} else {
		configPath = flag.Lookup("config").Value.String()
	}
	if err := ValidateConfigPath(configPath); err != nil {
		return "", err
	}
	return configPath, nil
}

// ValidateConfigPath validates that the given path exists and is a file (not a directory).
// Returns an error if the path does not exist or is a directory.
func ValidateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("invalid file format")
	}
	return nil
}
