// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
)

func ThemeConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ThemeConfiguration.html
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"data_color_palette": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DataColorPalette.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"colors": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 8, // Colors size needs to be in the range between 8 and 20
								MaxItems: 20,
								Elem:     hexColorSchema(attrElem),
							},
							"empty_fill_color": hexColorSchema(attrOptional),
							"min_max_gradient": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 2, // MinMaxGradient size needs to be 2
								MaxItems: 2,
								Elem:     hexColorSchema(attrElem),
							},
						},
					},
				},
				"sheet": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetStyle.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"tile": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TileStyle.html
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"border": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BorderStyle.html
											Type:     schema.TypeList,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"show": {
														Type:     schema.TypeBool,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
							"tile_layout": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TileLayoutStyle.html
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"gutter": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GutterStyle.html
											Type:     schema.TypeList,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"show": {
														Type:     schema.TypeBool,
														Optional: true,
													},
												},
											},
										},
										"margin": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_MarginStyle.html
											Type:     schema.TypeList,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"show": {
														Type:     schema.TypeBool,
														Optional: true,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				"typography": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Typography.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"font_families": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Font.html
								Type:     schema.TypeList,
								MaxItems: 5,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"font_family": {
											Type:     schema.TypeString,
											Optional: true,
										},
									},
								},
							},
						},
					},
				},
				"ui_color_palette": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_UIColorPalette.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"accent":               hexColorSchema(attrOptional),
							"accent_foreground":    hexColorSchema(attrOptional),
							"danger":               hexColorSchema(attrOptional),
							"danger_foreground":    hexColorSchema(attrOptional),
							"dimension":            hexColorSchema(attrOptional),
							"dimension_foreground": hexColorSchema(attrOptional),
							"measure":              hexColorSchema(attrOptional),
							"measure_foreground":   hexColorSchema(attrOptional),
							"primary_background":   hexColorSchema(attrOptional),
							"primary_foreground":   hexColorSchema(attrOptional),
							"secondary_background": hexColorSchema(attrOptional),
							"secondary_foreground": hexColorSchema(attrOptional),
							"success":              hexColorSchema(attrOptional),
							"success_foreground":   hexColorSchema(attrOptional),
							"warning":              hexColorSchema(attrOptional),
							"warning_foreground":   hexColorSchema(attrOptional),
						},
					},
				},
			},
		},
	}
}

func ThemeConfigurationDataSourceSchema() *schema.Schema {
	return sdkv2.DataSourcePropertyFromResourceProperty(ThemeConfigurationSchema())
}

func ExpandThemeConfiguration(tfList []any) *awstypes.ThemeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ThemeConfiguration{}

	if v, ok := tfMap["data_color_palette"].([]any); ok && len(v) > 0 {
		apiObject.DataColorPalette = expandDataColorPalette(v)
	}
	if v, ok := tfMap["sheet"].([]any); ok && len(v) > 0 {
		apiObject.Sheet = expandSheetStyle(v)
	}
	if v, ok := tfMap["typography"].([]any); ok && len(v) > 0 {
		apiObject.Typography = expandTypography(v)
	}
	if v, ok := tfMap["ui_color_palette"].([]any); ok && len(v) > 0 {
		apiObject.UIColorPalette = expandUIColorPalette(v)
	}

	return apiObject
}

func expandDataColorPalette(tfList []any) *awstypes.DataColorPalette {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DataColorPalette{}

	if v, ok := tfMap["colors"].([]any); ok {
		apiObject.Colors = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["empty_fill_color"].(string); ok && v != "" {
		apiObject.EmptyFillColor = aws.String(v)
	}
	if v, ok := tfMap["min_max_gradient"].([]any); ok {
		apiObject.MinMaxGradient = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandSheetStyle(tfList []any) *awstypes.SheetStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.SheetStyle{}

	if v, ok := tfMap["tile"].([]any); ok && len(v) > 0 {
		apiObject.Tile = expandTileStyle(v)
	}
	if v, ok := tfMap["tile_layout"].([]any); ok && len(v) > 0 {
		apiObject.TileLayout = expandTileLayoutStyle(v)
	}

	return apiObject
}

func expandTileStyle(tfList []any) *awstypes.TileStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TileStyle{}

	if v, ok := tfMap["border"].([]any); ok && len(v) > 0 {
		apiObject.Border = expandBorderStyle(v)
	}

	return apiObject
}

func expandBorderStyle(tfList []any) *awstypes.BorderStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.BorderStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		apiObject.Show = aws.Bool(v)
	}

	return apiObject
}

func expandTileLayoutStyle(tfList []any) *awstypes.TileLayoutStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.TileLayoutStyle{}

	if v, ok := tfMap["gutter"].([]any); ok && len(v) > 0 {
		apiObject.Gutter = expandGutterStyle(v)
	}
	if v, ok := tfMap["margin"].([]any); ok && len(v) > 0 {
		apiObject.Margin = expandMarginStyle(v)
	}

	return apiObject
}

func expandGutterStyle(tfList []any) *awstypes.GutterStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.GutterStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		apiObject.Show = aws.Bool(v)
	}

	return apiObject
}

func expandMarginStyle(tfList []any) *awstypes.MarginStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.MarginStyle{}

	if v, ok := tfMap["show"].(bool); ok {
		apiObject.Show = aws.Bool(v)
	}

	return apiObject
}

func expandTypography(tfList []any) *awstypes.Typography {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.Typography{}

	if v, ok := tfMap["font_families"].([]any); ok && len(v) > 0 {
		apiObject.FontFamilies = expandFonts(v)
	}

	return apiObject
}

func expandFonts(tfList []any) []awstypes.Font {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Font

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandFont(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFont(tfMap map[string]any) *awstypes.Font {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Font{}

	if v, ok := tfMap["font_family"].(string); ok && v != "" {
		apiObject.FontFamily = aws.String(v)
	}

	return apiObject
}

func expandUIColorPalette(tfList []any) *awstypes.UIColorPalette {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.UIColorPalette{}

	if v, ok := tfMap["accent"].(string); ok && v != "" {
		apiObject.Accent = aws.String(v)
	}
	if v, ok := tfMap["accent_foreground"].(string); ok && v != "" {
		apiObject.AccentForeground = aws.String(v)
	}
	if v, ok := tfMap["danger"].(string); ok && v != "" {
		apiObject.Danger = aws.String(v)
	}
	if v, ok := tfMap["danger_foreground"].(string); ok && v != "" {
		apiObject.DangerForeground = aws.String(v)
	}
	if v, ok := tfMap["dimension"].(string); ok && v != "" {
		apiObject.Dimension = aws.String(v)
	}
	if v, ok := tfMap["dimension_foreground"].(string); ok && v != "" {
		apiObject.DimensionForeground = aws.String(v)
	}
	if v, ok := tfMap["measure"].(string); ok && v != "" {
		apiObject.Measure = aws.String(v)
	}
	if v, ok := tfMap["measure_foreground"].(string); ok && v != "" {
		apiObject.MeasureForeground = aws.String(v)
	}
	if v, ok := tfMap["primary_background"].(string); ok && v != "" {
		apiObject.PrimaryBackground = aws.String(v)
	}
	if v, ok := tfMap["primary_foreground"].(string); ok && v != "" {
		apiObject.PrimaryForeground = aws.String(v)
	}
	if v, ok := tfMap["secondary_background"].(string); ok && v != "" {
		apiObject.SecondaryBackground = aws.String(v)
	}
	if v, ok := tfMap["secondary_foreground"].(string); ok && v != "" {
		apiObject.SecondaryForeground = aws.String(v)
	}
	if v, ok := tfMap["success"].(string); ok && v != "" {
		apiObject.Success = aws.String(v)
	}
	if v, ok := tfMap["success_foreground"].(string); ok && v != "" {
		apiObject.SuccessForeground = aws.String(v)
	}
	if v, ok := tfMap["warning"].(string); ok && v != "" {
		apiObject.Warning = aws.String(v)
	}
	if v, ok := tfMap["warning_foreground"].(string); ok && v != "" {
		apiObject.WarningForeground = aws.String(v)
	}

	return apiObject
}

func FlattenThemeConfiguration(apiObject *awstypes.ThemeConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DataColorPalette != nil {
		tfMap["data_color_palette"] = flattenDataColorPalette(apiObject.DataColorPalette)
	}
	if apiObject.Sheet != nil {
		tfMap["sheet"] = flattenSheetStyle(apiObject.Sheet)
	}
	if apiObject.Typography != nil {
		tfMap["typography"] = flattenTypography(apiObject.Typography)
	}
	if apiObject.UIColorPalette != nil {
		tfMap["ui_color_palette"] = flattenUIColorPalette(apiObject.UIColorPalette)
	}

	return []any{tfMap}
}

func flattenDataColorPalette(apiObject *awstypes.DataColorPalette) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Colors != nil {
		tfMap["colors"] = apiObject.Colors
	}
	if apiObject.EmptyFillColor != nil {
		tfMap["empty_fill_color"] = aws.ToString(apiObject.EmptyFillColor)
	}
	if apiObject.MinMaxGradient != nil {
		tfMap["min_max_gradient"] = apiObject.MinMaxGradient
	}

	return []any{tfMap}
}

func flattenSheetStyle(apiObject *awstypes.SheetStyle) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Tile != nil {
		tfMap["tile"] = flattenTileStyle(apiObject.Tile)
	}
	if apiObject.TileLayout != nil {
		tfMap["tile_layout"] = flattenTileLayoutStyle(apiObject.TileLayout)
	}

	return []any{tfMap}
}

func flattenTileStyle(apiObject *awstypes.TileStyle) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Border != nil {
		tfMap["border"] = flattenBorderStyle(apiObject.Border)
	}

	return []any{tfMap}
}

func flattenBorderStyle(apiObject *awstypes.BorderStyle) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Show != nil {
		tfMap["show"] = aws.ToBool(apiObject.Show)
	}

	return []any{tfMap}
}

func flattenTileLayoutStyle(apiObject *awstypes.TileLayoutStyle) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Gutter != nil {
		tfMap["gutter"] = flattenGutterStyle(apiObject.Gutter)
	}
	if apiObject.Margin != nil {
		tfMap["margin"] = flattenMarginStyle(apiObject.Margin)
	}

	return []any{tfMap}
}

func flattenGutterStyle(apiObject *awstypes.GutterStyle) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Show != nil {
		tfMap["show"] = aws.ToBool(apiObject.Show)
	}

	return []any{tfMap}
}

func flattenMarginStyle(apiObject *awstypes.MarginStyle) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Show != nil {
		tfMap["show"] = aws.ToBool(apiObject.Show)
	}

	return []any{tfMap}
}

func flattenTypography(apiObject *awstypes.Typography) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FontFamilies != nil {
		tfMap["font_families"] = flattenFonts(apiObject.FontFamilies)
	}

	return []any{tfMap}
}

func flattenFonts(apiObject []awstypes.Font) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObject {
		tfMap := map[string]any{}

		if apiObject.FontFamily != nil {
			tfMap["font_family"] = aws.ToString(apiObject.FontFamily)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenUIColorPalette(apiObject *awstypes.UIColorPalette) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Accent != nil {
		tfMap["accent"] = aws.ToString(apiObject.Accent)
	}
	if apiObject.AccentForeground != nil {
		tfMap["accent_foreground"] = aws.ToString(apiObject.AccentForeground)
	}
	if apiObject.Danger != nil {
		tfMap["danger"] = aws.ToString(apiObject.Danger)
	}
	if apiObject.DangerForeground != nil {
		tfMap["danger_foreground"] = aws.ToString(apiObject.DangerForeground)
	}
	if apiObject.Dimension != nil {
		tfMap["dimension"] = aws.ToString(apiObject.Dimension)
	}
	if apiObject.DimensionForeground != nil {
		tfMap["dimension_foreground"] = aws.ToString(apiObject.DimensionForeground)
	}
	if apiObject.Measure != nil {
		tfMap["measure"] = aws.ToString(apiObject.Measure)
	}
	if apiObject.MeasureForeground != nil {
		tfMap["measure_foreground"] = aws.ToString(apiObject.MeasureForeground)
	}
	if apiObject.PrimaryBackground != nil {
		tfMap["primary_background"] = aws.ToString(apiObject.PrimaryBackground)
	}
	if apiObject.PrimaryForeground != nil {
		tfMap["primary_foreground"] = aws.ToString(apiObject.PrimaryForeground)
	}
	if apiObject.SecondaryBackground != nil {
		tfMap["secondary_background"] = aws.ToString(apiObject.SecondaryBackground)
	}
	if apiObject.SecondaryForeground != nil {
		tfMap["secondary_foreground"] = aws.ToString(apiObject.SecondaryForeground)
	}
	if apiObject.Success != nil {
		tfMap["success"] = aws.ToString(apiObject.Success)
	}
	if apiObject.SuccessForeground != nil {
		tfMap["success_foreground"] = aws.ToString(apiObject.SuccessForeground)
	}
	if apiObject.Warning != nil {
		tfMap["warning"] = aws.ToString(apiObject.Warning)
	}
	if apiObject.WarningForeground != nil {
		tfMap["warning_foreground"] = aws.ToString(apiObject.WarningForeground)
	}

	return []any{tfMap}
}
