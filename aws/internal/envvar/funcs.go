package envvar

import (
	"os"
)

// GetWithDefault gets an environment variable value if non-empty or returns the default.
func GetWithDefault(variable string, defaultValue string) string {
	value := os.Getenv(variable)

	if value == "" {
		return defaultValue
	}

	return value
}
