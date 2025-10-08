// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var geospatialMapStyleOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialMapStyleOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"base_map_style": stringEnumSchema[awstypes.BaseMapStyleType](attrOptional),
			},
		},
	}
})

var geospatialWindowOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialWindowOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"bounds": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialCoordinateBounds.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"east":  floatBetweenSchema(attrRequired, -1800, 1800),
							"north": floatBetweenSchema(attrRequired, -90, 90),
							"south": floatBetweenSchema(attrRequired, -90, 90),
							"west":  floatBetweenSchema(attrRequired, -1800, 1800),
						},
					},
				},
				"map_zoom_mode": stringEnumSchema[awstypes.MapZoomMode](attrOptional),
			},
		},
	}
})

func expandGeospatialMapStyleOptions(tfList []any) *awstypes.GeospatialMapStyleOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GeospatialMapStyleOptions{}

	if v, ok := tfMap["base_map_style"].(string); ok && v != "" {
		apiObject.BaseMapStyle = awstypes.BaseMapStyleType(v)
	}

	return apiObject
}

func expandGeospatialWindowOptions(tfList []any) *awstypes.GeospatialWindowOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GeospatialWindowOptions{}

	if v, ok := tfMap["map_zoom_mode"].(string); ok && v != "" {
		apiObject.MapZoomMode = awstypes.MapZoomMode(v)
	}
	if v, ok := tfMap["bounds"].([]any); ok && len(v) > 0 {
		apiObject.Bounds = expandGeospatialCoordinateBounds(v)
	}

	return apiObject
}

func expandGeospatialCoordinateBounds(tfList []any) *awstypes.GeospatialCoordinateBounds {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GeospatialCoordinateBounds{}

	if v, ok := tfMap["east"].(float64); ok {
		apiObject.East = aws.Float64(v)
	}
	if v, ok := tfMap["north"].(float64); ok {
		apiObject.North = aws.Float64(v)
	}
	if v, ok := tfMap["south"].(float64); ok {
		apiObject.South = aws.Float64(v)
	}
	if v, ok := tfMap["west"].(float64); ok {
		apiObject.West = aws.Float64(v)
	}

	return apiObject
}

func flattenGeospatialMapStyleOptions(apiObject *awstypes.GeospatialMapStyleOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"base_map_style": apiObject.BaseMapStyle,
	}

	return []any{tfMap}
}

func flattenGeospatialWindowOptions(apiObject *awstypes.GeospatialWindowOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Bounds != nil {
		tfMap["bounds"] = flattenGeospatialCoordinateBounds(apiObject.Bounds)
	}
	tfMap["map_zoom_mode"] = apiObject.MapZoomMode

	return []any{tfMap}
}

func flattenGeospatialCoordinateBounds(apiObject *awstypes.GeospatialCoordinateBounds) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.East != nil {
		tfMap["east"] = aws.ToFloat64(apiObject.East)
	}
	if apiObject.North != nil {
		tfMap["north"] = aws.ToFloat64(apiObject.North)
	}
	if apiObject.South != nil {
		tfMap["south"] = aws.ToFloat64(apiObject.South)
	}
	if apiObject.West != nil {
		tfMap["west"] = aws.ToFloat64(apiObject.West)
	}

	return []any{tfMap}
}
