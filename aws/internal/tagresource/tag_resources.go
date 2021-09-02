package tagresource

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// Separator used in resource identifiers
	resourceIDSeparator = `,`
)

// GetResourceID parses a given resource identifier for tag identifier and tag key.
func GetResourceID(resourceID string) (string, string, error) {
	parts := strings.SplitN(resourceID, resourceIDSeparator, 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid resource identifier (%[1]s), expected ID%[2]sKEY", resourceID, resourceIDSeparator)
	}

	return parts[0], parts[1], nil
}

// SetResourceID creates a resource identifier given a tag identifier and a tag key.
func SetResourceID(identifier string, key string) string {
	parts := []string{identifier, key}
	resourceID := strings.Join(parts, resourceIDSeparator)

	return resourceID
}

// toSnakeCase converts a string to snake case.
//
// For example, AWS Go SDK field names are in PascalCase,
// while Terraform schema attribute names are in snake_case.
func toSnakeCase(str string) string {
	result := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	result = regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}
