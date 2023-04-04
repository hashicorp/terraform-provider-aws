package quicksight

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
