package tagresource

import (
	"regexp"
	"strings"

	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
)

// ServiceIdentifierAttributeName determines the schema identifier attribute name.
func ServiceIdentifierAttributeName(serviceName string) string {
	switch serviceName {
	case "ec2":
		return "resource_id"
	default:
		return toSnakeCase(tftags.ServiceTagInputIdentifierField(serviceName))
	}
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
