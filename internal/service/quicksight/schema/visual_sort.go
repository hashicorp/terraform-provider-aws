// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const fieldSortOptionsMaxItems100 = 100

var fieldSortOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSortOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: fieldSortOptionsMaxItems100,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column_sort": columnSortSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnSort.html
				"field_sort":  fieldSortSchema(),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSort.html
			},
		},
	}
})

var columnSortSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnSort.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"direction":            stringEnumSchema[awstypes.SortDirection](attrRequired),
				"sort_by":              columnSchema(true),               // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"aggregation_function": aggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
			},
		},
	}
})

var fieldSortSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FieldSort.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"direction": stringEnumSchema[awstypes.SortDirection](attrRequired),
				"field_id":  stringLenBetweenSchema(attrRequired, 1, 512),
			},
		},
	}
})

func expandFieldSortOptionsList(tfList []any) []awstypes.FieldSortOptions {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FieldSortOptions

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandFieldSortOptions(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFieldSortOptions(tfMap map[string]any) *awstypes.FieldSortOptions {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FieldSortOptions{}

	if v, ok := tfMap["column_sort"].([]any); ok && len(v) > 0 {
		apiObject.ColumnSort = expandColumnSort(v)
	}
	if v, ok := tfMap["field_sort"].([]any); ok && len(v) > 0 {
		apiObject.FieldSort = expandFieldSort(v)
	}

	return apiObject
}

func expandColumnSort(tfList []any) *awstypes.ColumnSort {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ColumnSort{}

	if v, ok := tfMap["direction"].(string); ok && v != "" {
		apiObject.Direction = awstypes.SortDirection(v)
	}
	if v, ok := tfMap["sort_by"].([]any); ok && len(v) > 0 {
		apiObject.SortBy = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["aggregation_function"].([]any); ok && len(v) > 0 {
		apiObject.AggregationFunction = expandAggregationFunction(v)
	}

	return apiObject
}

func expandFieldSort(tfList []any) *awstypes.FieldSort {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FieldSort{}

	if v, ok := tfMap["direction"].(string); ok && v != "" {
		apiObject.Direction = awstypes.SortDirection(v)
	}
	if v, ok := tfMap["field_id"].(string); ok && v != "" {
		apiObject.FieldId = aws.String(v)
	}

	return apiObject
}
