package attrmap

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// AttributeMap represents a map of Terraform resource attribute name to AWS API attribute name.
// Useful for SQS Queue or SNS Topic attribute handling.
type AttributeMap map[string]string

// ResourceDataToApiAttributes returns a map of AWS API attributes from Terraform ResourceData.
func (m AttributeMap) ResourceDataToApiAttributes(d *schema.ResourceData) (map[string]string, error) {
	apiAttributes := map[string]string{}

	for tfAttributeName, apiAttributeName := range m {
		if v, ok := d.GetOk(tfAttributeName); ok {
			var apiAttributeValue string

			switch v := v.(type) {
			case int:
				apiAttributeValue = strconv.Itoa(v)
			case bool:
				apiAttributeValue = strconv.FormatBool(v)
			case string:
				apiAttributeValue = v
			default:
				return nil, fmt.Errorf("attribute %s is of unsupported type: %T", tfAttributeName, v)
			}

			apiAttributes[apiAttributeName] = apiAttributeValue
		}
	}

	return apiAttributes, nil
}
