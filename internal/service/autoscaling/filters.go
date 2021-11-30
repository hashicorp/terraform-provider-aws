package autoscaling

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildFiltersDataSource(set *schema.Set) []*autoscaling.Filter {
	var filters []*autoscaling.Filter
	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []*string
		for _, e := range m["values"].([]interface{}) {
			filterValues = append(filterValues, aws.String(e.(string)))
		}

		// In previous iterations, users were expected to provide "key" and "value" tag names.
		// With the addition of asgs filters, the signature is "tag-key" and "tag-value", so these conditions prevent breaking changes.
		// https://docs.aws.amazon.com/sdk-for-go/api/service/autoscaling/#Filter
		name := m["name"].(string)
		if name == "key" {
			name = "tag-key"
		}
		if name == "value" {
			name = "tag-value"
		}
		filters = append(filters, &autoscaling.Filter{
			Name:   aws.String(name),
			Values: filterValues,
		})
	}
	return filters
}

func dataSourceFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
				},

				"values": {
					Type:     schema.TypeList,
					Required: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}
