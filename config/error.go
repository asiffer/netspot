package config

import "fmt"

// ConfigNotFoundError is a basic error that is triggered
// when the given config file does not exist
type ConfigNotFoundError struct {
	Path string
}

func (err *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("config file '%s' has not been found", err.Path)
}
