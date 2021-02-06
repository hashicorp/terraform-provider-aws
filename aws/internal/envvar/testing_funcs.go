package envvar

import (
	"os"

	"github.com/mitchellh/go-testing-interface"
)

// TestFailIfAllEmpty verifies that at least one environment variable is non-empty or fails the test.
//
// If at lease one environment variable is non-empty, returns the first name and value.
func TestFailIfAllEmpty(t testing.T, names []string, usageMessage string) (string, string) {
	t.Helper()

	for _, variable := range names {
		value := os.Getenv(variable)

		if value != "" {
			return variable, value
		}
	}

	t.Fatalf("at least one environment variable of %v must be set. Usage: %s", names, usageMessage)

	return "", ""
}

// TestFailIfEmpty verifies that an environment variable is non-empty or fails the test.
//
// For acceptance tests, this function must be used outside PreCheck functions to set values for configurations.
func TestFailIfEmpty(t testing.T, name string, usageMessage string) string {
	t.Helper()

	value := os.Getenv(name)

	if value == "" {
		t.Fatalf("environment variable %s must be set. Usage: %s", name, usageMessage)
	}

	return value
}

// TestSkipIfEmpty verifies that an environment variable is non-empty or skips the test.
//
// For acceptance tests, this function must be used outside PreCheck functions to set values for configurations.
func TestSkipIfEmpty(t testing.T, name string, usageMessage string) string {
	t.Helper()

	value := os.Getenv(name)

	if value == "" {
		t.Skipf("skipping test; environment variable %s must be set. Usage: %s", name, usageMessage)
	}

	return value
}
