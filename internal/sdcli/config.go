package sdcli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

type Config struct {
	// The Stability API key to use for generating images.
	APIKey string `json:"api_key"`

	// The directory to output images to.  This can be an absolute or relative path,
	// but it will not expand tilde for home directories nor will it interpret environment
	// variables.
	//
	// Images will be saved by Unix timestamp with an appropriate file ending.
	OutputDirectory string `json:"output_directory"`

	// The command to run after generating an image.  This command will be invoked with
	// the path to the image as an argument.  E.g. putting "firefox" in here will result
	// in "firefox /path/to/image" being called after the image is generated.
	PostGenerationCommand string `json:"post_generation_command"`
}

// DefaultConfigPath returns the default path to the config file for sdcli.
func DefaultConfigPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(home, ".config", "sdcli", "config.json"), nil
}

// ParseConfigFile parses the given configuration file and returns a Config.
func ParseConfigFile(configPath string) (*Config, error) {
	fd, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open configuration file %q: %w", configPath, err)
	}

	defer fd.Close()

	var config Config

	err = json.NewDecoder(fd).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON in configuration file %q: %w", configPath, err)
	}

	return &config, nil
}
