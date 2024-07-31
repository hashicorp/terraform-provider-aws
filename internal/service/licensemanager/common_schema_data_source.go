// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package licensemanager

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/licensemanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func BuildFiltersDataSource(set *schema.Set) []awstypes.Filter {
	var filters []awstypes.Filter
	for _, v := range set.List() {
		m := v.(map[string]interface{})
		var filterValues []string
		for _, e := range m[names.AttrValues].([]interface{}) {
			filterValues = append(filterValues, e.(string))
		}
		filters = append(filters, awstypes.Filter{
			Name:   aws.String(m[names.AttrName].(string)),
			Values: filterValues,
		})
	}
	return filters
}

func DataSourceFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},

				names.AttrValues: {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}
