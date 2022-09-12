package util

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Loads and parses the YAML file in the specified path and stores the result into the dest object.
//
// The dest object must not be nil.
// Returns nil on success or an error, if any occurs.
func ParseYamlFile(path string, dest interface{}) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("the specified path is not a file, but a directory: %s", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(dest); err != nil {
		return err
	}

	return nil
}
