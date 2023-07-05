// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

const fieldSortOptionsMaxItems100 = 100

func fieldSortOptionsSchema(maxItems int) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: maxItems,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column_sort": columnSortSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnSort.html
				"field_sort":  fieldSortSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSort.html
			},
		},
	}
}

func columnSortSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnSort.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"direction":            stringSchema(true, validation.StringInSlice(quicksight.SortDirection_Values(), false)),
				"sort_by":              columnSchema(),                   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"aggregation_function": aggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
			},
		},
	}
}

func fieldSortSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSort.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"direction": stringSchema(true, validation.StringInSlice(quicksight.SortDirection_Values(), false)),
				"field_id":  stringSchema(true, validation.StringLenBetween(1, 512)),
			},
		},
	}
}

func expandFieldSortOptionsList(tfList []interface{}) []*quicksight.FieldSortOptions {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.FieldSortOptions
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandFieldSortOptions(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandFieldSortOptions(tfMap map[string]interface{}) *quicksight.FieldSortOptions {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.FieldSortOptions{}

	if v, ok := tfMap["column_sort"].([]interface{}); ok && len(v) > 0 {
		options.ColumnSort = expandColumnSort(v)
	}
	if v, ok := tfMap["field_sort"].([]interface{}); ok && len(v) > 0 {
		options.FieldSort = expandFieldSort(v)
	}

	return options
}

func expandColumnSort(tfList []interface{}) *quicksight.ColumnSort {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ColumnSort{}

	if v, ok := tfMap["direction"].(string); ok && v != "" {
		config.Direction = aws.String(v)
	}
	if v, ok := tfMap["sort_by"].([]interface{}); ok && len(v) > 0 {
		config.SortBy = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		config.AggregationFunction = expandAggregationFunction(v)
	}

	return config
}

func expandFieldSort(tfList []interface{}) *quicksight.FieldSort {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FieldSort{}

	if v, ok := tfMap["direction"].(string); ok && v != "" {
		config.Direction = aws.String(v)
	}
	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		config.FieldId = aws.String(v)
	}

	return config
}
