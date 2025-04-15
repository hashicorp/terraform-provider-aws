// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package namevaluesfilters

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// NameValuesFilters is a standard implementation for AWS resource filters.
// The AWS Go SDK is split into multiple service packages, each service with
// its own Go struct type representing a resource filter. To standardize logic
// across all these Go types, we convert them into this Go type.
type NameValuesFilters map[string][]string

// Add adds missing and updates existing filters from common Terraform Provider SDK types.
// Supports map[string]string, map[string][]string, *schema.Set.
func (filters NameValuesFilters) Add(v any) NameValuesFilters {
	switch v := v.(type) {
	case map[string]string:
		for name, v := range v {
			if values, ok := filters[name]; ok {
				filters[name] = append(values, v)
			} else {
				values = []string{v}
				filters[name] = values
			}
		}

	case map[string][]string:
		// We can't use fallthrough here, so recurse.
		return filters.Add(NameValuesFilters(v))

	case NameValuesFilters:
		for name, v := range v {
			if values, ok := filters[name]; ok {
				filters[name] = append(values, v...)
			} else {
				values = make([]string, len(v))
				copy(values, v)
				filters[name] = values
			}
		}

	case *schema.Set:
		// The set of filters described by Schema().
		for _, tfMapRaw := range v.List() {
			tfMap := tfMapRaw.(map[string]any)
			name := tfMap[names.AttrName].(string)

			for _, v := range tfMap[names.AttrValues].(*schema.Set).List() {
				v := v.(string)
				if values, ok := filters[name]; ok {
					filters[name] = append(values, v)
				} else {
					values = []string{v}
					filters[name] = values
				}
			}
		}
	}

	return filters
}

// Map returns filter names mapped to their values.
// Duplicate values are eliminated and empty values removed.
func (filters NameValuesFilters) Map() map[string][]string {
	result := make(map[string][]string)

	for k, v := range filters {
		targetValues := make([]string, 0)

	SOURCE_VALUES:
		for _, sourceValue := range v {
			if sourceValue == "" {
				continue
			}

			for _, targetValue := range targetValues {
				if sourceValue == targetValue {
					continue SOURCE_VALUES
				}
			}

			targetValues = append(targetValues, sourceValue)
		}

		if len(targetValues) == 0 {
			continue
		}

		result[k] = targetValues
	}

	return result
}

// New creates NameValuesFilters from common Terraform Provider SDK types.
// Supports map[string]string, map[string][]string, *schema.Set.
func New(v any) NameValuesFilters {
	return make(NameValuesFilters).Add(v)
}

// Schema returns a *schema.Schema that represents a set of custom filtering criteria
// that a user can specify as input to a data source.
// It is conventional for an attribute of this type to be included as a top-level attribute called "filter".
func Schema() *schema.Schema {
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
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}
