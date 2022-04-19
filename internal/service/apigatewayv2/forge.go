package apigatewayv2

import (
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/create"
)

// hashStringCaseInsensitive hashes strings in a case insensitive manner.
// If you want a Set of strings and are case inensitive, this is the SchemaSetFunc you want.
func hashStringCaseInsensitive(v interface{}) int {
	return create.StringHashcode(strings.ToLower(v.(string)))
}
