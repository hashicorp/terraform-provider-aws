// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func geospatialMapStyleOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialMapStyleOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"base_map_style": stringSchema(false, validation.StringInSlice(quicksight.BaseMapStyleType_Values(), false)),
			},
		},
	}
}

func geospatialWindowOptionsSchema() *schema.Schema {
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
							"east": {
								Type:         schema.TypeFloat,
								Required:     true,
								ValidateFunc: validation.IntBetween(-1800, 1800),
							},
							"north": {
								Type:         schema.TypeFloat,
								Required:     true,
								ValidateFunc: validation.IntBetween(-90, 90),
							},
							"south": {
								Type:         schema.TypeFloat,
								Required:     true,
								ValidateFunc: validation.IntBetween(-90, 90),
							},
							"west": {
								Type:         schema.TypeFloat,
								Required:     true,
								ValidateFunc: validation.IntBetween(-1800, 1800),
							},
						},
					},
				},
				"map_zoom_mode": stringSchema(false, validation.StringInSlice(quicksight.MapZoomMode_Values(), false)),
			},
		},
	}
}

func expandGeospatialMapStyleOptions(tfList []interface{}) *quicksight.GeospatialMapStyleOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.GeospatialMapStyleOptions{}

	if v, ok := tfMap["base_map_style"].(string); ok && v != "" {
		options.BaseMapStyle = aws.String(v)
	}

	return options
}

func expandGeospatialWindowOptions(tfList []interface{}) *quicksight.GeospatialWindowOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.GeospatialWindowOptions{}

	if v, ok := tfMap["map_zoom_mode"].(string); ok && v != "" {
		options.MapZoomMode = aws.String(v)
	}
	if v, ok := tfMap["bounds"].([]interface{}); ok && len(v) > 0 {
		options.Bounds = expandGeospatialCoordinateBounds(v)
	}

	return options
}

func expandGeospatialCoordinateBounds(tfList []interface{}) *quicksight.GeospatialCoordinateBounds {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.GeospatialCoordinateBounds{}

	if v, ok := tfMap["east"].(float64); ok {
		config.East = aws.Float64(v)
	}
	if v, ok := tfMap["north"].(float64); ok {
		config.North = aws.Float64(v)
	}
	if v, ok := tfMap["south"].(float64); ok {
		config.South = aws.Float64(v)
	}
	if v, ok := tfMap["west"].(float64); ok {
		config.West = aws.Float64(v)
	}

	return config
}

func flattenGeospatialMapStyleOptions(apiObject *quicksight.GeospatialMapStyleOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.BaseMapStyle != nil {
		tfMap["base_map_style"] = aws.StringValue(apiObject.BaseMapStyle)
	}

	return []interface{}{tfMap}
}

func flattenGeospatialWindowOptions(apiObject *quicksight.GeospatialWindowOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Bounds != nil {
		tfMap["bounds"] = flattenGeospatialCoordinateBounds(apiObject.Bounds)
	}
	if apiObject.MapZoomMode != nil {
		tfMap["map_zoom_mode"] = aws.StringValue(apiObject.MapZoomMode)
	}

	return []interface{}{tfMap}
}

func flattenGeospatialCoordinateBounds(apiObject *quicksight.GeospatialCoordinateBounds) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.East != nil {
		tfMap["east"] = aws.Float64Value(apiObject.East)
	}
	if apiObject.North != nil {
		tfMap["north"] = aws.Float64Value(apiObject.North)
	}
	if apiObject.South != nil {
		tfMap["south"] = aws.Float64Value(apiObject.South)
	}
	if apiObject.West != nil {
		tfMap["west"] = aws.Float64Value(apiObject.West)
	}

	return []interface{}{tfMap}
}
