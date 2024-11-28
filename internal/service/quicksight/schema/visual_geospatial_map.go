// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func geospatialMapVisualSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialMapVisual.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"visual_id":       idSchema(),
				names.AttrActions: visualCustomActionsSchema(customActionsMaxItems), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
				"chart_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialMapConfiguration.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialMapFieldWells.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"geospatial_map_aggregated_field_wells": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialMapAggregatedFieldWells.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"colors":         dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													"geospatial":     dimensionFieldSchema(dimensionsFieldMaxItems200), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DimensionField.html
													names.AttrValues: measureFieldSchema(measureFieldsMaxItems200),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MeasureField.html
												},
											},
										},
									},
								},
							},
							"legend":            legendOptionsSchema(),             // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LegendOptions.html
							"map_style_options": geospatialMapStyleOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialMapStyleOptions.html
							"point_style_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialPointStyleOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"cluster_marker_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ClusterMarkerConfiguration.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"cluster_marker": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ClusterMarker.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"simple_cluster_marker": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SimpleClusterMarker.html
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"color": hexColorSchema(attrOptional),
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
										"selected_point_style": stringEnumSchema[awstypes.GeospatialSelectedPointStyle](attrOptional),
									},
								},
							},
							"tooltip":        tooltipOptionsSchema(),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TooltipOptions.html
							"visual_palette": visualPaletteSchema(),           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualPalette.html
							"window_options": geospatialWindowOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GeospatialWindowOptions.html
						},
					},
				},
				"column_hierarchies": columnHierarchiesSchema(),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnHierarchy.html
				"subtitle":           visualSubtitleLabelOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualSubtitleLabelOptions.html
				"title":              visualTitleLabelOptionsSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualTitleLabelOptions.html
			},
		},
	}
}

func expandGeospatialMapVisual(tfList []interface{}) *awstypes.GeospatialMapVisual {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GeospatialMapVisual{}

	if v, ok := tfMap["visual_id"].(string); ok && v != "" {
		apiObject.VisualId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrActions].([]interface{}); ok && len(v) > 0 {
		apiObject.Actions = expandVisualCustomActions(v)
	}
	if v, ok := tfMap["chart_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ChartConfiguration = expandGeospatialMapConfiguration(v)
	}
	if v, ok := tfMap["column_hierarchies"].([]interface{}); ok && len(v) > 0 {
		apiObject.ColumnHierarchies = expandColumnHierarchies(v)
	}
	if v, ok := tfMap["subtitle"].([]interface{}); ok && len(v) > 0 {
		apiObject.Subtitle = expandVisualSubtitleLabelOptions(v)
	}
	if v, ok := tfMap["title"].([]interface{}); ok && len(v) > 0 {
		apiObject.Title = expandVisualTitleLabelOptions(v)
	}

	return apiObject
}

func expandGeospatialMapConfiguration(tfList []interface{}) *awstypes.GeospatialMapConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GeospatialMapConfiguration{}

	if v, ok := tfMap["field_wells"].([]interface{}); ok && len(v) > 0 {
		apiObject.FieldWells = expandGeospatialMapFieldWells(v)
	}
	if v, ok := tfMap["legend"].([]interface{}); ok && len(v) > 0 {
		apiObject.Legend = expandLegendOptions(v)
	}
	if v, ok := tfMap["map_style_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.MapStyleOptions = expandGeospatialMapStyleOptions(v)
	}
	if v, ok := tfMap["point_style_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.PointStyleOptions = expandGeospatialPointStyleOptions(v)
	}
	if v, ok := tfMap["tooltip"].([]interface{}); ok && len(v) > 0 {
		apiObject.Tooltip = expandTooltipOptions(v)
	}
	if v, ok := tfMap["visual_palatte"].([]interface{}); ok && len(v) > 0 {
		apiObject.VisualPalette = expandVisualPalette(v)
	}
	if v, ok := tfMap["window_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.WindowOptions = expandGeospatialWindowOptions(v)
	}

	return apiObject
}

func expandGeospatialMapFieldWells(tfList []interface{}) *awstypes.GeospatialMapFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GeospatialMapFieldWells{}

	if v, ok := tfMap["geospatial_map_aggregated_field_wells"].([]interface{}); ok && len(v) > 0 {
		apiObject.GeospatialMapAggregatedFieldWells = expandGeospatialMapAggregatedFieldWells(v)
	}

	return apiObject
}

func expandGeospatialMapAggregatedFieldWells(tfList []interface{}) *awstypes.GeospatialMapAggregatedFieldWells {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GeospatialMapAggregatedFieldWells{}

	if v, ok := tfMap["colors"].([]interface{}); ok && len(v) > 0 {
		apiObject.Colors = expandDimensionFields(v)
	}
	if v, ok := tfMap["geospatial"].([]interface{}); ok && len(v) > 0 {
		apiObject.Geospatial = expandDimensionFields(v)
	}
	if v, ok := tfMap[names.AttrValues].([]interface{}); ok && len(v) > 0 {
		apiObject.Values = expandMeasureFields(v)
	}

	return apiObject
}

func expandGeospatialPointStyleOptions(tfList []interface{}) *awstypes.GeospatialPointStyleOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GeospatialPointStyleOptions{}

	if v, ok := tfMap["selected_point_style"].(string); ok && v != "" {
		apiObject.SelectedPointStyle = awstypes.GeospatialSelectedPointStyle(v)
	}
	if v, ok := tfMap["cluster_marker_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ClusterMarkerConfiguration = expandClusterMarkerConfiguration(v)
	}

	return apiObject
}

func expandClusterMarkerConfiguration(tfList []interface{}) *awstypes.ClusterMarkerConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ClusterMarkerConfiguration{}

	if v, ok := tfMap["cluster_marker"].([]interface{}); ok && len(v) > 0 {
		apiObject.ClusterMarker = expandClusterMarker(v)
	}

	return apiObject
}

func expandClusterMarker(tfList []interface{}) *awstypes.ClusterMarker {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ClusterMarker{}

	if v, ok := tfMap["simple_cluster_marker"].([]interface{}); ok && len(v) > 0 {
		apiObject.SimpleClusterMarker = expandSimpleClusterMarker(v)
	}

	return apiObject
}

func expandSimpleClusterMarker(tfList []interface{}) *awstypes.SimpleClusterMarker {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SimpleClusterMarker{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}

	return apiObject
}

func flattenGeospatialMapVisual(apiObject *awstypes.GeospatialMapVisual) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"visual_id": aws.ToString(apiObject.VisualId),
	}

	if apiObject.Actions != nil {
		tfMap[names.AttrActions] = flattenVisualCustomAction(apiObject.Actions)
	}
	if apiObject.ChartConfiguration != nil {
		tfMap["chart_configuration"] = flattenGeospatialMapConfiguration(apiObject.ChartConfiguration)
	}
	if apiObject.ColumnHierarchies != nil {
		tfMap["column_hierarchies"] = flattenColumnHierarchy(apiObject.ColumnHierarchies)
	}
	if apiObject.Subtitle != nil {
		tfMap["subtitle"] = flattenVisualSubtitleLabelOptions(apiObject.Subtitle)
	}
	if apiObject.Title != nil {
		tfMap["title"] = flattenVisualTitleLabelOptions(apiObject.Title)
	}

	return []interface{}{tfMap}
}

func flattenGeospatialMapConfiguration(apiObject *awstypes.GeospatialMapConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.FieldWells != nil {
		tfMap["field_wells"] = flattenGeospatialMapFieldWells(apiObject.FieldWells)
	}
	if apiObject.Legend != nil {
		tfMap["legend"] = flattenLegendOptions(apiObject.Legend)
	}
	if apiObject.MapStyleOptions != nil {
		tfMap["map_style_options"] = flattenGeospatialMapStyleOptions(apiObject.MapStyleOptions)
	}
	if apiObject.PointStyleOptions != nil {
		tfMap["point_style_options"] = flattenGeospatialPointStyleOptions(apiObject.PointStyleOptions)
	}
	if apiObject.Tooltip != nil {
		tfMap["tooltip"] = flattenTooltipOptions(apiObject.Tooltip)
	}
	if apiObject.WindowOptions != nil {
		tfMap["window_options"] = flattenGeospatialWindowOptions(apiObject.WindowOptions)
	}
	if apiObject.VisualPalette != nil {
		tfMap["visual_palette"] = flattenVisualPalette(apiObject.VisualPalette)
	}

	return []interface{}{tfMap}
}

func flattenGeospatialMapFieldWells(apiObject *awstypes.GeospatialMapFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.GeospatialMapAggregatedFieldWells != nil {
		tfMap["geospatial_map_aggregated_field_wells"] = flattenGeospatialMapAggregatedFieldWells(apiObject.GeospatialMapAggregatedFieldWells)
	}

	return []interface{}{tfMap}
}

func flattenGeospatialMapAggregatedFieldWells(apiObject *awstypes.GeospatialMapAggregatedFieldWells) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Colors != nil {
		tfMap["colors"] = flattenDimensionFields(apiObject.Colors)
	}
	if apiObject.Geospatial != nil {
		tfMap["geospatial"] = flattenDimensionFields(apiObject.Geospatial)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = flattenMeasureFields(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenGeospatialPointStyleOptions(apiObject *awstypes.GeospatialPointStyleOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ClusterMarkerConfiguration != nil {
		tfMap["cluster_marker_configuration"] = flattenClusterMarkerConfiguration(apiObject.ClusterMarkerConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenClusterMarkerConfiguration(apiObject *awstypes.ClusterMarkerConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ClusterMarker != nil {
		tfMap["cluster_marker"] = flattenClusterMarker(apiObject.ClusterMarker)
	}

	return []interface{}{tfMap}
}

func flattenClusterMarker(apiObject *awstypes.ClusterMarker) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SimpleClusterMarker != nil {
		tfMap["simple_cluster_marker"] = flattenSimpleClusterMarker(apiObject.SimpleClusterMarker)
	}

	return []interface{}{tfMap}
}

func flattenSimpleClusterMarker(apiObject *awstypes.SimpleClusterMarker) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}

	return []interface{}{tfMap}
}
