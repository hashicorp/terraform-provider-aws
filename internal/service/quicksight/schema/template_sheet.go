// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var analysisDefaultSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"default_new_sheet_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultNewSheetConfiguration.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"interactive_layout_configuration": interactiveLayoutConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultInteractiveLayoutConfiguration.html
							"paginated_layout_configuration":   paginatedLayoutConfigurationSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultPaginatedLayoutConfiguration.html,
							"sheet_content_type":               stringEnumSchema[awstypes.SheetContentType](attrOptional),
						},
					},
				},
			},
		},
	}
})

func interactiveLayoutConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultInteractiveLayoutConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"free_form": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"canvas_size_options": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"screen_canvas_size_options": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"optimized_view_port_width": {
														Type:     schema.TypeString,
														Required: true,
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
				"grid": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"canvas_size_options": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"screen_canvas_size_options": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"optimized_view_port_width": {
														Type:     schema.TypeString,
														Optional: true,
													},
													"resize_option": stringEnumSchema[awstypes.ResizeOption](attrRequired),
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
		},
	}
}

func paginatedLayoutConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultPaginatedLayoutConfiguration.html,
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"section_based": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"canvas_size_options": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"paper_canvas_size_options": paperCanvasSizeOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionBasedLayoutPaperCanvasSizeOptions.html
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

var paperCanvasSizeOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"paper_margin":      spacingSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Spacing.html
				"paper_orientation": stringEnumSchema[awstypes.PaperOrientation](attrOptional),
				"paper_size":        stringEnumSchema[awstypes.PaperSize](attrOptional),
			},
		},
	}
})

var sheetControlLayoutsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetControlLayout.html
		Type:     schema.TypeList,
		MinItems: 0,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrConfiguration: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetControlLayoutConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"grid_layout": gridLayoutConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GridLayoutConfiguration.html,
						},
					},
				},
			},
		},
	}
})

var layoutSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Layout.html
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrConfiguration: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LayoutConfiguration.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"free_form_layout": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"elements": freeFormLayoutElementsSchema(),
										"canvas_size_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutCanvasSizeOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"screen_canvas_size_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutScreenCanvasSizeOptions.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"optimized_view_port_width": {
																	Type:     schema.TypeString,
																	Required: true,
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
							"grid_layout": gridLayoutConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GridLayoutConfiguration.html,
							"section_based_layout": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionBasedLayoutConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"body_sections": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BodySectionConfiguration.html
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 28,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrContent: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BodySectionContent.html
														Type:     schema.TypeList,
														Required: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"layout": sectionLayoutConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionLayoutConfiguration.html
															},
														},
													},
													"section_id": idSchema(),
													"page_break_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionPageBreakConfiguration.html
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"after": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionAfterPageBreak.html
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			names.AttrStatus: stringEnumSchema[awstypes.Status](attrOptional),
																		},
																	},
																},
															},
														},
													},
													"style": sectionStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionStyle.html
												},
											},
										},
										"canvas_size_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionBasedLayoutCanvasSizeOptions.html
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"paper_canvas_size_options": paperCanvasSizeOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionBasedLayoutPaperCanvasSizeOptions.html

												},
											},
										},
										"footer_sections": headerFooterSectionConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeaderFooterSectionConfiguration.html
										"header_sections": headerFooterSectionConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeaderFooterSectionConfiguration.html
									},
								},
							},
						},
					},
				},
			},
		},
	}
})

var gridLayoutConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GridLayoutConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"elements": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GridLayoutElement.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 430,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column_span":  intBetweenSchema(attrRequired, 1, 36),
							"element_id":   idSchema(),
							"element_type": stringEnumSchema[awstypes.LayoutElementType](attrRequired),
							"row_span":     intBetweenSchema(attrRequired, 1, 21),
							"column_index": {
								Type:         nullable.TypeNullableInt,
								Optional:     true,
								ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 35),
							},
							"row_index": {
								Type:         nullable.TypeNullableInt,
								Optional:     true,
								ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 9009),
							},
						},
					},
				},
				"canvas_size_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GridLayoutCanvasSizeOptions.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"screen_canvas_size_options": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_GridLayoutScreenCanvasSizeOptions.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"optimized_view_port_width": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"resize_option": stringEnumSchema[awstypes.ResizeOption](attrRequired),
									},
								},
							},
						},
					},
				},
			},
		},
	}
})

var headerFooterSectionConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_HeaderFooterSectionConfiguration.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"layout":     sectionLayoutConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionLayoutConfiguration.html
				"section_id": idSchema(),
				"style":      sectionStyleSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionStyle.html
			},
		},
	}
})

var sectionStyleSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionStyle.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"height": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"padding": spacingSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Spacing.html
			},
		},
	}
})

var freeFormLayoutElementsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutElement.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 430,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"element_id":   idSchema(),
				"element_type": stringEnumSchema[awstypes.LayoutElementType](attrRequired),
				"height": {
					Type:     schema.TypeString,
					Required: true,
				},
				"width": {
					Type:     schema.TypeString,
					Required: true,
				},
				"x_axis_location": {
					Type:     schema.TypeString,
					Required: true,
				},
				"y_axis_location": {
					Type:     schema.TypeString,
					Required: true,
				},
				"background_style": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutElementBackgroundStyle.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":      stringMatchSchema(attrOptional, `^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`, ""),
							"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
						},
					},
				},
				"border_style": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutElementBorderStyle.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":      stringMatchSchema(attrOptional, `^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`, ""),
							"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
						},
					},
				},
				"loading_animation": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LoadingAnimation.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
						},
					},
				},
				"rendering_rules": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetElementRenderingRule.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 10000,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"configuration_overrides": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetElementConfigurationOverrides.html
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
									},
								},
							},
							names.AttrExpression: stringLenBetweenSchema(attrRequired, 1, 4096),
						},
					},
				},
				"selected_border_style": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutElementBorderStyle.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"color":      stringMatchSchema(attrOptional, `^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`, ""),
							"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
						},
					},
				},
				"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

var sectionLayoutConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SectionLayoutConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"free_form_layout": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormSectionLayoutConfiguration.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"elements": freeFormLayoutElementsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutElement.html
						},
					},
				},
			},
		},
	}
})

var spacingSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Spacing.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"bottom": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"left": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"right": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"top": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
})

func expandAnalysisDefaults(tfList []interface{}) *awstypes.AnalysisDefaults {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.AnalysisDefaults{}

	if v, ok := tfMap["default_new_sheet_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultNewSheetConfiguration = expandDefaultNewSheetConfiguration(v)
	}

	return apiObject
}

func expandDefaultNewSheetConfiguration(tfList []interface{}) *awstypes.DefaultNewSheetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultNewSheetConfiguration{}

	if v, ok := tfMap["interactive_layout_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.InteractiveLayoutConfiguration = expandDefaultInteractiveLayoutConfiguration(v)
	}

	if v, ok := tfMap["paginated_layout_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.PaginatedLayoutConfiguration = expandDefaultPaginatedLayoutConfiguration(v)
	}

	if v, ok := tfMap["sheet_content_type"].(string); ok && v != "" {
		apiObject.SheetContentType = awstypes.SheetContentType(v)
	}

	return apiObject
}

func expandDefaultInteractiveLayoutConfiguration(tfList []interface{}) *awstypes.DefaultInteractiveLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultInteractiveLayoutConfiguration{}

	if v, ok := tfMap["free_form"].([]interface{}); ok && len(v) > 0 {
		apiObject.FreeForm = expandDefaultFreeFormLayoutConfiguration(v)
	}

	if v, ok := tfMap["grid"].([]interface{}); ok && len(v) > 0 {
		apiObject.Grid = expandDefaultGridLayoutConfiguration(v)
	}

	return apiObject
}

func expandDefaultFreeFormLayoutConfiguration(tfList []interface{}) *awstypes.DefaultFreeFormLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultFreeFormLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.CanvasSizeOptions = expandFreeFormLayoutCanvasSizeOptions(v)
	}

	return apiObject
}

func expandFreeFormLayoutCanvasSizeOptions(tfList []interface{}) *awstypes.FreeFormLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FreeFormLayoutCanvasSizeOptions{}

	if v, ok := tfMap["screen_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.ScreenCanvasSizeOptions = expandFreeFormLayoutScreenCanvasSizeOptions(v)
	}

	return apiObject
}

func expandFreeFormLayoutScreenCanvasSizeOptions(tfList []interface{}) *awstypes.FreeFormLayoutScreenCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FreeFormLayoutScreenCanvasSizeOptions{}

	if v, ok := tfMap["optimized_view_port_width"].(string); ok && v != "" {
		apiObject.OptimizedViewPortWidth = aws.String(v)
	}

	return apiObject
}

func expandDefaultGridLayoutConfiguration(tfList []interface{}) *awstypes.DefaultGridLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultGridLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.CanvasSizeOptions = expandGridLayoutCanvasSizeOptions(v)
	}

	return apiObject
}

func expandGridLayoutCanvasSizeOptions(tfList []interface{}) *awstypes.GridLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GridLayoutCanvasSizeOptions{}

	if v, ok := tfMap["screen_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.ScreenCanvasSizeOptions = expandGridLayoutScreenCanvasSizeOptions(v)
	}

	return apiObject
}

func expandGridLayoutScreenCanvasSizeOptions(tfList []interface{}) *awstypes.GridLayoutScreenCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GridLayoutScreenCanvasSizeOptions{}

	if v, ok := tfMap["optimized_view_port_width"].(string); ok && v != "" {
		apiObject.OptimizedViewPortWidth = aws.String(v)
	}
	if v, ok := tfMap["resize_option"].(string); ok && v != "" {
		apiObject.ResizeOption = awstypes.ResizeOption(v)
	}

	return apiObject
}

func expandDefaultPaginatedLayoutConfiguration(tfList []interface{}) *awstypes.DefaultPaginatedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultPaginatedLayoutConfiguration{}

	if v, ok := tfMap["section_based"].([]interface{}); ok && len(v) > 0 {
		apiObject.SectionBased = expandDefaultSectionBasedLayoutConfiguration(v)
	}

	return apiObject
}

func expandDefaultSectionBasedLayoutConfiguration(tfList []interface{}) *awstypes.DefaultSectionBasedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultSectionBasedLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.CanvasSizeOptions = expandSectionBasedLayoutCanvasSizeOptions(v)
	}

	return apiObject
}

func expandSectionBasedLayoutCanvasSizeOptions(tfList []interface{}) *awstypes.SectionBasedLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SectionBasedLayoutCanvasSizeOptions{}

	if v, ok := tfMap["paper_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.PaperCanvasSizeOptions = expandSectionBasedLayoutPaperCanvasSizeOptions(v)
	}

	return apiObject
}

func expandSectionBasedLayoutPaperCanvasSizeOptions(tfList []interface{}) *awstypes.SectionBasedLayoutPaperCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SectionBasedLayoutPaperCanvasSizeOptions{}

	if v, ok := tfMap["paper_margin"].([]interface{}); ok && len(v) > 0 {
		apiObject.PaperMargin = expandSpacing(v)
	}
	if v, ok := tfMap["paper_orientation"].(string); ok && v != "" {
		apiObject.PaperOrientation = awstypes.PaperOrientation(v)
	}
	if v, ok := tfMap["paper_size"].(string); ok && v != "" {
		apiObject.PaperSize = awstypes.PaperSize(v)
	}

	return apiObject
}

func expandSpacing(tfList []interface{}) *awstypes.Spacing {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.Spacing{}

	if v, ok := tfMap["bottom"].(string); ok && v != "" {
		apiObject.Bottom = aws.String(v)
	}

	if v, ok := tfMap["left"].(string); ok && v != "" {
		apiObject.Left = aws.String(v)
	}

	if v, ok := tfMap["right"].(string); ok && v != "" {
		apiObject.Right = aws.String(v)
	}

	if v, ok := tfMap["top"].(string); ok && v != "" {
		apiObject.Top = aws.String(v)
	}

	return apiObject
}

func expandSheetDefinition(tfMap map[string]interface{}) *awstypes.SheetDefinition {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SheetDefinition{}

	if v, ok := tfMap["sheet_id"].(string); ok && v != "" {
		apiObject.SheetId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrContentType].(string); ok && v != "" {
		apiObject.ContentType = awstypes.SheetContentType(v)
	}
	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}
	if v, ok := tfMap["filter_controls"].([]interface{}); ok && len(v) > 0 {
		apiObject.FilterControls = expandFilterControls(v)
	}
	if v, ok := tfMap["layouts"].([]interface{}); ok && len(v) > 0 {
		apiObject.Layouts = expandLayouts(v)
	}
	if v, ok := tfMap["parameter_controls"].([]interface{}); ok && len(v) > 0 {
		apiObject.ParameterControls = expandParameterControls(v)
	}
	if v, ok := tfMap["sheet_control_layouts"].([]interface{}); ok && len(v) > 0 {
		apiObject.SheetControlLayouts = expandSheetControlLayouts(v)
	}
	if v, ok := tfMap["text_boxes"].([]interface{}); ok && len(v) > 0 {
		apiObject.TextBoxes = expandSheetTextBoxes(v)
	}
	if v, ok := tfMap["visuals"].([]interface{}); ok && len(v) > 0 {
		apiObject.Visuals = expandVisuals(v)
	}

	return apiObject
}

func expandFilterControls(tfList []interface{}) []awstypes.FilterControl {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FilterControl

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandFilterControl(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandLayouts(tfList []interface{}) []awstypes.Layout {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Layout

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandLayout(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandLayout(tfMap map[string]interface{}) *awstypes.Layout {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Layout{}

	if v, ok := tfMap[names.AttrConfiguration].([]interface{}); ok && len(v) > 0 {
		apiObject.Configuration = expandLayoutConfiguration(v)
	}

	return apiObject
}

func expandLayoutConfiguration(tfList []interface{}) *awstypes.LayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.LayoutConfiguration{}

	if v, ok := tfMap["free_form_layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.FreeFormLayout = expandFreeFormLayoutConfiguration(v)
	}
	if v, ok := tfMap["grid_layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.GridLayout = expandGridLayoutConfiguration(v)
	}
	if v, ok := tfMap["section_based_layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.SectionBasedLayout = expandSectionBasedLayoutConfiguration(v)
	}

	return apiObject
}

func expandFreeFormLayoutConfiguration(tfList []interface{}) *awstypes.FreeFormLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FreeFormLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		apiObject.Elements = expandFreeFormLayoutElements(v)
	}
	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.CanvasSizeOptions = expandFreeFormLayoutCanvasSizeOptions(v)
	}

	return apiObject
}

func expandFreeFormLayoutElements(tfList []interface{}) []awstypes.FreeFormLayoutElement {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FreeFormLayoutElement

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandFreeFormLayoutElement(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFreeFormLayoutElement(tfMap map[string]interface{}) *awstypes.FreeFormLayoutElement {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FreeFormLayoutElement{}

	if v, ok := tfMap["element_id"].(string); ok && v != "" {
		apiObject.ElementId = aws.String(v)
	}
	if v, ok := tfMap["element_type"].(string); ok && v != "" {
		apiObject.ElementType = awstypes.LayoutElementType(v)
	}
	if v, ok := tfMap["height"].(string); ok && v != "" {
		apiObject.Height = aws.String(v)
	}
	if v, ok := tfMap["width"].(string); ok && v != "" {
		apiObject.Width = aws.String(v)
	}
	if v, ok := tfMap["x_axis_location"].(string); ok && v != "" {
		apiObject.XAxisLocation = aws.String(v)
	}
	if v, ok := tfMap["y_axis_location"].(string); ok && v != "" {
		apiObject.YAxisLocation = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["background_style"].([]interface{}); ok && len(v) > 0 {
		apiObject.BackgroundStyle = expandFreeFormLayoutElementBackgroundStyle(v)
	}
	if v, ok := tfMap["border_style"].([]interface{}); ok && len(v) > 0 {
		apiObject.BorderStyle = expandFreeFormLayoutElementBorderStyle(v)
	}
	if v, ok := tfMap["loading_animation"].([]interface{}); ok && len(v) > 0 {
		apiObject.LoadingAnimation = expandLoadingAnimation(v)
	}
	if v, ok := tfMap["rendering_rules"].([]interface{}); ok && len(v) > 0 {
		apiObject.RenderingRules = expandSheetElementRenderingRules(v)
	}
	if v, ok := tfMap["selected_border_style"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectedBorderStyle = expandFreeFormLayoutElementBorderStyle(v)
	}

	return apiObject
}

func expandFreeFormLayoutElementBackgroundStyle(tfList []interface{}) *awstypes.FreeFormLayoutElementBackgroundStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FreeFormLayoutElementBackgroundStyle{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandFreeFormLayoutElementBorderStyle(tfList []interface{}) *awstypes.FreeFormLayoutElementBorderStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FreeFormLayoutElementBorderStyle{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		apiObject.Color = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandLoadingAnimation(tfList []interface{}) *awstypes.LoadingAnimation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.LoadingAnimation{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandSheetElementRenderingRules(tfList []interface{}) []awstypes.SheetElementRenderingRule {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SheetElementRenderingRule

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandSheetElementRenderingRule(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandSheetElementRenderingRule(tfMap map[string]interface{}) *awstypes.SheetElementRenderingRule {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SheetElementRenderingRule{}

	if v, ok := tfMap[names.AttrExpression].(string); ok && v != "" {
		apiObject.Expression = aws.String(v)
	}
	if v, ok := tfMap["configuration_overrides"].([]interface{}); ok && len(v) > 0 {
		apiObject.ConfigurationOverrides = expandSheetElementConfigurationOverrides(v)
	}

	return apiObject
}

func expandSheetElementConfigurationOverrides(tfList []interface{}) *awstypes.SheetElementConfigurationOverrides {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SheetElementConfigurationOverrides{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandGridLayoutConfiguration(tfList []interface{}) *awstypes.GridLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.GridLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		apiObject.Elements = expandGridLayoutElements(v)
	}
	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.CanvasSizeOptions = expandGridLayoutCanvasSizeOptions(v)
	}

	return apiObject
}

func expandGridLayoutElements(tfList []interface{}) []awstypes.GridLayoutElement {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.GridLayoutElement

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandGridLayoutElement(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandGridLayoutElement(tfMap map[string]interface{}) *awstypes.GridLayoutElement {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GridLayoutElement{}

	if v, ok := tfMap["element_id"].(string); ok && v != "" {
		apiObject.ElementId = aws.String(v)
	}
	if v, ok := tfMap["element_type"].(string); ok && v != "" {
		apiObject.ElementType = awstypes.LayoutElementType(v)
	}
	if v, ok := tfMap["column_span"].(int); ok && v != 0 {
		apiObject.ColumnSpan = aws.Int32(int32(v))
	}
	if v, ok := tfMap["row_span"].(int); ok && v != 0 {
		apiObject.RowSpan = aws.Int32(int32(v))
	}
	if v, ok := tfMap["column_index"].(string); ok && v != "" {
		if i, null, _ := nullable.Int(v).ValueInt32(); !null {
			apiObject.ColumnIndex = aws.Int32(i)
		}
	}
	if v, ok := tfMap["row_index"].(string); ok && v != "" {
		if i, null, _ := nullable.Int(v).ValueInt32(); !null {
			apiObject.RowIndex = aws.Int32(i)
		}
	}

	return apiObject
}

func expandSectionBasedLayoutConfiguration(tfList []interface{}) *awstypes.SectionBasedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SectionBasedLayoutConfiguration{}

	if v, ok := tfMap["body_sections"].([]interface{}); ok && len(v) > 0 {
		apiObject.BodySections = expandBodySectionConfigurations(v)
	}
	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.CanvasSizeOptions = expandSectionBasedLayoutCanvasSizeOptions(v)
	}
	if v, ok := tfMap["footer_sections"].([]interface{}); ok && len(v) > 0 {
		apiObject.FooterSections = expandHeaderFooterSectionConfigurations(v)
	}
	if v, ok := tfMap["header_sections"].([]interface{}); ok && len(v) > 0 {
		apiObject.HeaderSections = expandHeaderFooterSectionConfigurations(v)
	}

	return apiObject
}

func expandBodySectionConfigurations(tfList []interface{}) []awstypes.BodySectionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.BodySectionConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandBodySectionConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandBodySectionConfiguration(tfMap map[string]interface{}) *awstypes.BodySectionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.BodySectionConfiguration{}

	if v, ok := tfMap["section_id"].(string); ok && v != "" {
		apiObject.SectionId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrContent].([]interface{}); ok && len(v) > 0 {
		apiObject.Content = expandBodySectionContent(v)
	}
	if v, ok := tfMap["page_break_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.PageBreakConfiguration = expandSectionPageBreakConfiguration(v)
	}
	if v, ok := tfMap["style"].([]interface{}); ok && len(v) > 0 {
		apiObject.Style = expandSectionStyle(v)
	}

	return apiObject
}

func expandBodySectionContent(tfList []interface{}) *awstypes.BodySectionContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.BodySectionContent{}

	if v, ok := tfMap["layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.Layout = expandSectionLayoutConfiguration(v)
	}

	return apiObject
}

func expandSectionLayoutConfiguration(tfList []interface{}) *awstypes.SectionLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SectionLayoutConfiguration{}

	if v, ok := tfMap["free_form_layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.FreeFormLayout = expandFreeFormSectionLayoutConfiguration(v)
	}

	return apiObject
}

func expandFreeFormSectionLayoutConfiguration(tfList []interface{}) *awstypes.FreeFormSectionLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FreeFormSectionLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		apiObject.Elements = expandFreeFormLayoutElements(v)
	}

	return apiObject
}

func expandSectionPageBreakConfiguration(tfList []interface{}) *awstypes.SectionPageBreakConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SectionPageBreakConfiguration{}

	if v, ok := tfMap["after"].([]interface{}); ok && len(v) > 0 {
		apiObject.After = expandSectionAfterPageBreak(v)
	}

	return apiObject
}

func expandSectionAfterPageBreak(tfList []interface{}) *awstypes.SectionAfterPageBreak {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SectionAfterPageBreak{}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.SectionPageBreakStatus(v)
	}

	return apiObject
}

func expandSectionStyle(tfList []interface{}) *awstypes.SectionStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SectionStyle{}

	if v, ok := tfMap["height"].(string); ok && v != "" {
		apiObject.Height = aws.String(v)
	}
	if v, ok := tfMap["padding"].([]interface{}); ok && len(v) > 0 {
		apiObject.Padding = expandSpacing(v)
	}

	return apiObject
}

func expandHeaderFooterSectionConfigurations(tfList []interface{}) []awstypes.HeaderFooterSectionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.HeaderFooterSectionConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandHeaderFooterSectionConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandHeaderFooterSectionConfiguration(tfMap map[string]interface{}) *awstypes.HeaderFooterSectionConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.HeaderFooterSectionConfiguration{}

	if v, ok := tfMap["section_id"].(string); ok && v != "" {
		apiObject.SectionId = aws.String(v)
	}
	if v, ok := tfMap["layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.Layout = expandSectionLayoutConfiguration(v)
	}
	if v, ok := tfMap["style"].([]interface{}); ok && len(v) > 0 {
		apiObject.Style = expandSectionStyle(v)
	}

	return apiObject
}

func expandParameterControls(tfList []interface{}) []awstypes.ParameterControl {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ParameterControl

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandParameterControl(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandSheetControlLayouts(tfList []interface{}) []awstypes.SheetControlLayout {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SheetControlLayout

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandSheetControlLayout(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandSheetControlLayout(tfMap map[string]interface{}) *awstypes.SheetControlLayout {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SheetControlLayout{}

	if v, ok := tfMap[names.AttrConfiguration].([]interface{}); ok && len(v) > 0 {
		apiObject.Configuration = expandSheetControlLayoutConfiguration(v)
	}

	return apiObject
}

func expandSheetControlLayoutConfiguration(tfList []interface{}) *awstypes.SheetControlLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SheetControlLayoutConfiguration{}

	if v, ok := tfMap["grid_layout"].([]interface{}); ok && len(v) > 0 {
		apiObject.GridLayout = expandGridLayoutConfiguration(v)
	}

	return apiObject
}

func expandSheetTextBoxes(tfList []interface{}) []awstypes.SheetTextBox {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SheetTextBox

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandSheetTextBox(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandSheetTextBox(tfMap map[string]interface{}) *awstypes.SheetTextBox {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SheetTextBox{}

	if v, ok := tfMap["sheet_text_box_id"].(string); ok && v != "" {
		apiObject.SheetTextBoxId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrContent].(string); ok && v != "" {
		apiObject.Content = aws.String(v)
	}

	return apiObject
}

func expandVisuals(tfList []interface{}) []awstypes.Visual {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Visual

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandVisual(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenAnalysisDefaults(apiObject *awstypes.AnalysisDefaults) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DefaultNewSheetConfiguration != nil {
		tfMap["default_new_sheet_configuration"] = flattenDefaultNewSheetConfiguration(apiObject.DefaultNewSheetConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenDefaultNewSheetConfiguration(apiObject *awstypes.DefaultNewSheetConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.InteractiveLayoutConfiguration != nil {
		tfMap["interactive_layout_configuration"] = flattenDefaultInteractiveLayoutConfiguration(apiObject.InteractiveLayoutConfiguration)
	}
	if apiObject.PaginatedLayoutConfiguration != nil {
		tfMap["paginated_layout_configuration"] = flattenDefaultPaginatedLayoutConfiguration(apiObject.PaginatedLayoutConfiguration)
	}
	tfMap["sheet_content_type"] = apiObject.SheetContentType

	return []interface{}{tfMap}
}

func flattenDefaultInteractiveLayoutConfiguration(apiObject *awstypes.DefaultInteractiveLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.FreeForm != nil {
		tfMap["free_form"] = flattenDefaultFreeFormLayoutConfiguration(apiObject.FreeForm)
	}
	if apiObject.Grid != nil {
		tfMap["grid"] = flattenDefaultGridLayoutConfiguration(apiObject.Grid)
	}

	return []interface{}{tfMap}
}

func flattenDefaultFreeFormLayoutConfiguration(apiObject *awstypes.DefaultFreeFormLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenFreeFormLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutCanvasSizeOptions(apiObject *awstypes.FreeFormLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ScreenCanvasSizeOptions != nil {
		tfMap["screen_canvas_size_options"] = flattenFreeFormLayoutScreenCanvasSizeOptions(apiObject.ScreenCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutScreenCanvasSizeOptions(apiObject *awstypes.FreeFormLayoutScreenCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.OptimizedViewPortWidth != nil {
		tfMap["optimized_view_port_width"] = aws.ToString(apiObject.OptimizedViewPortWidth)
	}

	return []interface{}{tfMap}
}

func flattenDefaultGridLayoutConfiguration(apiObject *awstypes.DefaultGridLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenGridLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenGridLayoutCanvasSizeOptions(apiObject *awstypes.GridLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ScreenCanvasSizeOptions != nil {
		tfMap["screen_canvas_size_options"] = flattenGridLayoutScreenCanvasSizeOptions(apiObject.ScreenCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenGridLayoutScreenCanvasSizeOptions(apiObject *awstypes.GridLayoutScreenCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.OptimizedViewPortWidth != nil {
		tfMap["optimized_view_port_width"] = aws.ToString(apiObject.OptimizedViewPortWidth)
	}
	tfMap["resize_option"] = apiObject.ResizeOption

	return []interface{}{tfMap}
}

func flattenDefaultPaginatedLayoutConfiguration(apiObject *awstypes.DefaultPaginatedLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SectionBased != nil {
		tfMap["section_based"] = flattenDefaultSectionBasedLayoutConfiguration(apiObject.SectionBased)
	}

	return []interface{}{tfMap}
}

func flattenDefaultSectionBasedLayoutConfiguration(apiObject *awstypes.DefaultSectionBasedLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenSectionBasedLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenSectionBasedLayoutCanvasSizeOptions(apiObject *awstypes.SectionBasedLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.PaperCanvasSizeOptions != nil {
		tfMap["paper_canvas_size_options"] = flattenSectionBasedLayoutPaperCanvasSizeOptions(apiObject.PaperCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenSectionBasedLayoutPaperCanvasSizeOptions(apiObject *awstypes.SectionBasedLayoutPaperCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.PaperMargin != nil {
		tfMap["paper_margin"] = flattenSpacing(apiObject.PaperMargin)
	}
	tfMap["paper_orientation"] = apiObject.PaperOrientation
	tfMap["paper_size"] = apiObject.PaperSize

	return []interface{}{tfMap}
}

func flattenSpacing(apiObject *awstypes.Spacing) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Bottom != nil {
		tfMap["bottom"] = aws.ToString(apiObject.Bottom)
	}
	if apiObject.Left != nil {
		tfMap["left"] = aws.ToString(apiObject.Left)
	}
	if apiObject.Right != nil {
		tfMap["right"] = aws.ToString(apiObject.Right)
	}
	if apiObject.Top != nil {
		tfMap["top"] = aws.ToString(apiObject.Top)
	}

	return []interface{}{tfMap}
}

func flattenLayouts(apiObjects []awstypes.Layout) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrConfiguration: flattenLayoutConfiguration(apiObject.Configuration),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenLayoutConfiguration(apiObject *awstypes.LayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.FreeFormLayout != nil {
		tfMap["free_form_layout"] = flattenFreeFormLayoutConfiguration(apiObject.FreeFormLayout)
	}
	if apiObject.GridLayout != nil {
		tfMap["grid_layout"] = flattenGridLayoutConfiguration(apiObject.GridLayout)
	}
	if apiObject.SectionBasedLayout != nil {
		tfMap["section_based_layout"] = flattenSectionBasedLayoutConfiguration(apiObject.SectionBasedLayout)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutConfiguration(apiObject *awstypes.FreeFormLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenFreeFormLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}
	if apiObject.Elements != nil {
		tfMap["elements"] = flattenFreeFormLayoutElement(apiObject.Elements)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutElement(apiObjects []awstypes.FreeFormLayoutElement) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"element_id":      aws.ToString(apiObject.ElementId),
			"element_type":    apiObject.ElementType,
			"height":          aws.ToString(apiObject.Height),
			"width":           aws.ToString(apiObject.Width),
			"x_axis_location": aws.ToString(apiObject.XAxisLocation),
			"y_axis_location": aws.ToString(apiObject.YAxisLocation),
		}

		if apiObject.BackgroundStyle != nil {
			tfMap["background_style"] = flattenFreeFormLayoutElementBackgroundStyle(apiObject.BackgroundStyle)
		}
		if apiObject.BorderStyle != nil {
			tfMap["border_style"] = flattenFreeFormLayoutElementBorderStyle(apiObject.BorderStyle)
		}
		if apiObject.LoadingAnimation != nil {
			tfMap["loading_animation"] = flattenLoadingAnimation(apiObject.LoadingAnimation)
		}
		if apiObject.RenderingRules != nil {
			tfMap["rendering_rules"] = flattenSheetElementRenderingRule(apiObject.RenderingRules)
		}
		if apiObject.SelectedBorderStyle != nil {
			tfMap["selected_border_style"] = flattenFreeFormLayoutElementBorderStyle(apiObject.SelectedBorderStyle)
		}
		tfMap["visibility"] = apiObject.Visibility

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFreeFormLayoutElementBackgroundStyle(apiObject *awstypes.FreeFormLayoutElementBackgroundStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutElementBorderStyle(apiObject *awstypes.FreeFormLayoutElementBorderStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	tfMap["visibility"] = apiObject.Visibility

	return []interface{}{tfMap}
}

func flattenLoadingAnimation(apiObject *awstypes.LoadingAnimation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["visibility"] = apiObject.Visibility

	return []interface{}{tfMap}
}

func flattenSheetElementRenderingRule(apiObjects []awstypes.SheetElementRenderingRule) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.ConfigurationOverrides != nil {
			tfMap["configuration_overrides"] = flattenSheetElementConfigurationOverrides(apiObject.ConfigurationOverrides)
		}
		if apiObject.Expression != nil {
			tfMap[names.AttrExpression] = aws.ToString(apiObject.Expression)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetElementConfigurationOverrides(apiObject *awstypes.SheetElementConfigurationOverrides) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["visibility"] = apiObject.Visibility

	return []interface{}{tfMap}
}

func flattenGridLayoutConfiguration(apiObject *awstypes.GridLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenGridLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}
	if apiObject.Elements != nil {
		tfMap["elements"] = flattenGridLayoutElement(apiObject.Elements)
	}

	return []interface{}{tfMap}
}

func flattenGridLayoutElement(apiObjects []awstypes.GridLayoutElement) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"column_span":  aws.ToInt32(apiObject.ColumnSpan),
			"element_id":   aws.ToString(apiObject.ElementId),
			"element_type": apiObject.ElementType,
			"row_span":     aws.ToInt32(apiObject.RowSpan),
		}

		if apiObject.ColumnIndex != nil {
			tfMap["column_index"] = flex.Int32ToStringValue(apiObject.ColumnIndex)
		}
		if apiObject.RowIndex != nil {
			tfMap["row_index"] = flex.Int32ToStringValue(apiObject.RowIndex)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSectionBasedLayoutConfiguration(apiObject *awstypes.SectionBasedLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.BodySections != nil {
		tfMap["body_sections"] = flattenBodySectionConfiguration(apiObject.BodySections)
	}
	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenSectionBasedLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}
	if apiObject.FooterSections != nil {
		tfMap["footer_sections"] = flattenHeaderFooterSectionConfiguration(apiObject.FooterSections)
	}
	if apiObject.HeaderSections != nil {
		tfMap["header_sections"] = flattenHeaderFooterSectionConfiguration(apiObject.HeaderSections)
	}

	return []interface{}{tfMap}
}

func flattenBodySectionConfiguration(apiObjects []awstypes.BodySectionConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrContent: flattenBodySectionContent(apiObject.Content),
			"section_id":      aws.ToString(apiObject.SectionId),
		}

		if apiObject.PageBreakConfiguration != nil {
			tfMap["page_break_configuration"] = flattenSectionPageBreakConfiguration(apiObject.PageBreakConfiguration)
		}
		if apiObject.Style != nil {
			tfMap["style"] = flattenSectionStyle(apiObject.Style)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenBodySectionContent(apiObject *awstypes.BodySectionContent) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Layout != nil {
		tfMap["layout"] = flattenSectionLayoutConfiguration(apiObject.Layout)
	}

	return []interface{}{tfMap}
}

func flattenSectionLayoutConfiguration(apiObject *awstypes.SectionLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.FreeFormLayout != nil {
		tfMap["free_form_layout"] = flattenFreeFormSectionLayoutConfiguration(apiObject.FreeFormLayout)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormSectionLayoutConfiguration(apiObject *awstypes.FreeFormSectionLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Elements != nil {
		tfMap["free_form_layout"] = flattenFreeFormLayoutElement(apiObject.Elements)
	}

	return []interface{}{tfMap}
}

func flattenSectionPageBreakConfiguration(apiObject *awstypes.SectionPageBreakConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.After != nil {
		tfMap["after"] = flattenSectionAfterPageBreak(apiObject.After)
	}

	return []interface{}{tfMap}
}

func flattenSectionAfterPageBreak(apiObject *awstypes.SectionAfterPageBreak) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap[names.AttrStatus] = apiObject.Status

	return []interface{}{tfMap}
}

func flattenSectionStyle(apiObject *awstypes.SectionStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Height != nil {
		tfMap["height"] = aws.ToString(apiObject.Height)
	}
	if apiObject.Padding != nil {
		tfMap["padding"] = flattenSpacing(apiObject.Padding)
	}

	return []interface{}{tfMap}
}

func flattenHeaderFooterSectionConfiguration(apiObjects []awstypes.HeaderFooterSectionConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"section_id": aws.ToString(apiObject.SectionId),
		}

		if apiObject.Layout != nil {
			tfMap["layout"] = flattenSectionLayoutConfiguration(apiObject.Layout)
		}
		if apiObject.Style != nil {
			tfMap["style"] = flattenSectionStyle(apiObject.Style)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetControlLayouts(apiObjects []awstypes.SheetControlLayout) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrConfiguration: flattenSheetControlLayoutConfiguration(apiObject.Configuration),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetControlLayoutConfiguration(apiObject *awstypes.SheetControlLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.GridLayout != nil {
		tfMap["grid_layout"] = flattenGridLayoutConfiguration(apiObject.GridLayout)
	}

	return []interface{}{tfMap}
}
