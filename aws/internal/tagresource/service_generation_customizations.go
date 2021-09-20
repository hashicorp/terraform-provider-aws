package tagresource

import (
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
)

// ServiceIdentifierAttributeName determines the schema identifier attribute name.
func ServiceIdentifierAttributeName(serviceName string) string {
	switch serviceName {
	case "ec2":
		return "resource_id"
	default:
		return toSnakeCase(keyvaluetags.ServiceTagInputIdentifierField(serviceName))
	}
}
