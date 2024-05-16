// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
)

func analysisDefaultSchema() *schema.Schema {
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
							"sheet_content_type":               stringSchema(false, validation.StringInSlice(quicksight.SheetContentType_Values(), false)),
						},
					},
				},
			},
		},
	}
}

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
													"resize_option": stringSchema(true, validation.StringInSlice(quicksight.ResizeOption_Values(), false)),
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

func paperCanvasSizeOptionsSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"paper_margin":      spacingSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Spacing.html
				"paper_orientation": stringSchema(false, validation.StringInSlice(quicksight.PaperOrientation_Values(), false)),
				"paper_size":        stringSchema(false, validation.StringInSlice(quicksight.PaperSize_Values(), false)),
			},
		},
	}
}

func sheetControlLayoutsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetControlLayout.html
		Type:     schema.TypeList,
		MinItems: 0,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetControlLayoutConfiguration.html
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
}

func layoutSchema() *schema.Schema {
	return &schema.Schema{ // // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_Layout.html
		Type:     schema.TypeList,
		Optional: true,
		Computed: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LayoutConfiguration.html
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
													"content": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_BodySectionContent.html
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
																			"status": stringSchema(false, validation.StringInSlice(quicksight.Status_Values(), false)),
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
}

func gridLayoutConfigurationSchema() *schema.Schema {
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
							"column_span": {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntBetween(1, 36),
							},
							"element_id":   idSchema(),
							"element_type": stringSchema(true, validation.StringInSlice(quicksight.LayoutElementType_Values(), false)),
							"row_span": {
								Type:         schema.TypeInt,
								Required:     true,
								ValidateFunc: validation.IntBetween(1, 21),
							},
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
										"resize_option": stringSchema(true, validation.StringInSlice(quicksight.ResizeOption_Values(), false)),
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

func headerFooterSectionConfigurationSchema() *schema.Schema {
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
}

func sectionStyleSchema() *schema.Schema {
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
}

func freeFormLayoutElementsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FreeFormLayoutElement.html
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 430,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"element_id":   idSchema(),
				"element_type": stringSchema(true, validation.StringInSlice(quicksight.LayoutElementType_Values(), false)),
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
							"color":      stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`), "")),
							"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
							"color":      stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`), "")),
							"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
							"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
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
										"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
									},
								},
							},
							"expression": stringSchema(true, validation.StringLenBetween(1, 4096)),
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
							"color":      stringSchema(false, validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`), "")),
							"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
						},
					},
				},
				"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
			},
		},
	}
}

func sectionLayoutConfigurationSchema() *schema.Schema {
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
}

func spacingSchema() *schema.Schema {
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
}

func expandAnalysisDefaults(tfList []interface{}) *quicksight.AnalysisDefaults {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	defaults := &quicksight.AnalysisDefaults{}

	if v, ok := tfMap["default_new_sheet_configuration"].([]interface{}); ok && len(v) > 0 {
		defaults.DefaultNewSheetConfiguration = expandDefaultNewSheetConfiguration(v)
	}

	return defaults
}

func expandDefaultNewSheetConfiguration(tfList []interface{}) *quicksight.DefaultNewSheetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DefaultNewSheetConfiguration{}

	if v, ok := tfMap["interactive_layout_configuration"].([]interface{}); ok && len(v) > 0 {
		config.InteractiveLayoutConfiguration = expandDefaultInteractiveLayoutConfiguration(v)
	}

	if v, ok := tfMap["paginated_layout_configuration"].([]interface{}); ok && len(v) > 0 {
		config.PaginatedLayoutConfiguration = expandDefaultPaginatedLayoutConfiguration(v)
	}

	if v, ok := tfMap["sheet_content_type"].(string); ok && v != "" {
		config.SheetContentType = aws.String(v)
	}

	return config
}

func expandDefaultInteractiveLayoutConfiguration(tfList []interface{}) *quicksight.DefaultInteractiveLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DefaultInteractiveLayoutConfiguration{}

	if v, ok := tfMap["free_form"].([]interface{}); ok && len(v) > 0 {
		config.FreeForm = expandDefaultFreeFormLayoutConfiguration(v)
	}

	if v, ok := tfMap["grid"].([]interface{}); ok && len(v) > 0 {
		config.Grid = expandDefaultGridLayoutConfiguration(v)
	}

	return config
}

func expandDefaultFreeFormLayoutConfiguration(tfList []interface{}) *quicksight.DefaultFreeFormLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DefaultFreeFormLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandFreeFormLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandFreeFormLayoutCanvasSizeOptions(tfList []interface{}) *quicksight.FreeFormLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.FreeFormLayoutCanvasSizeOptions{}

	if v, ok := tfMap["screen_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		options.ScreenCanvasSizeOptions = expandFreeFormLayoutScreenCanvasSizeOptions(v)
	}

	return options
}

func expandFreeFormLayoutScreenCanvasSizeOptions(tfList []interface{}) *quicksight.FreeFormLayoutScreenCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.FreeFormLayoutScreenCanvasSizeOptions{}

	if v, ok := tfMap["optimized_view_port_width"].(string); ok && v != "" {
		options.OptimizedViewPortWidth = aws.String(v)
	}

	return options
}

func expandDefaultGridLayoutConfiguration(tfList []interface{}) *quicksight.DefaultGridLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DefaultGridLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandGridLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandGridLayoutCanvasSizeOptions(tfList []interface{}) *quicksight.GridLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.GridLayoutCanvasSizeOptions{}

	if v, ok := tfMap["screen_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		options.ScreenCanvasSizeOptions = expandGridLayoutScreenCanvasSizeOptions(v)
	}

	return options
}

func expandGridLayoutScreenCanvasSizeOptions(tfList []interface{}) *quicksight.GridLayoutScreenCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.GridLayoutScreenCanvasSizeOptions{}

	if v, ok := tfMap["optimized_view_port_width"].(string); ok && v != "" {
		options.OptimizedViewPortWidth = aws.String(v)
	}
	if v, ok := tfMap["resize_option"].(string); ok && v != "" {
		options.ResizeOption = aws.String(v)
	}

	return options
}

func expandDefaultPaginatedLayoutConfiguration(tfList []interface{}) *quicksight.DefaultPaginatedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DefaultPaginatedLayoutConfiguration{}

	if v, ok := tfMap["section_based"].([]interface{}); ok && len(v) > 0 {
		config.SectionBased = expandDefaultSectionBasedLayoutConfiguration(v)
	}

	return config
}

func expandDefaultSectionBasedLayoutConfiguration(tfList []interface{}) *quicksight.DefaultSectionBasedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DefaultSectionBasedLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandSectionBasedLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandSectionBasedLayoutCanvasSizeOptions(tfList []interface{}) *quicksight.SectionBasedLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.SectionBasedLayoutCanvasSizeOptions{}

	if v, ok := tfMap["paper_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		options.PaperCanvasSizeOptions = expandSectionBasedLayoutPaperCanvasSizeOptions(v)
	}

	return options
}

func expandSectionBasedLayoutPaperCanvasSizeOptions(tfList []interface{}) *quicksight.SectionBasedLayoutPaperCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.SectionBasedLayoutPaperCanvasSizeOptions{}

	if v, ok := tfMap["paper_margin"].([]interface{}); ok && len(v) > 0 {
		options.PaperMargin = expandSpacing(v)
	}
	if v, ok := tfMap["paper_orientation"].(string); ok && v != "" {
		options.PaperOrientation = aws.String(v)
	}
	if v, ok := tfMap["paper_size"].(string); ok && v != "" {
		options.PaperSize = aws.String(v)
	}

	return options
}

func expandSpacing(tfList []interface{}) *quicksight.Spacing {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	spacing := &quicksight.Spacing{}

	if v, ok := tfMap["bottom"].(string); ok && v != "" {
		spacing.Bottom = aws.String(v)
	}

	if v, ok := tfMap["left"].(string); ok && v != "" {
		spacing.Left = aws.String(v)
	}

	if v, ok := tfMap["right"].(string); ok && v != "" {
		spacing.Right = aws.String(v)
	}

	if v, ok := tfMap["top"].(string); ok && v != "" {
		spacing.Top = aws.String(v)
	}

	return spacing
}

func expandSheetDefinition(tfMap map[string]interface{}) *quicksight.SheetDefinition {
	if tfMap == nil {
		return nil
	}

	sheet := &quicksight.SheetDefinition{}

	if v, ok := tfMap["sheet_id"].(string); ok && v != "" {
		sheet.SheetId = aws.String(v)
	}
	if v, ok := tfMap["content_type"].(string); ok && v != "" {
		sheet.ContentType = aws.String(v)
	}
	if v, ok := tfMap["description"].(string); ok && v != "" {
		sheet.Description = aws.String(v)
	}
	if v, ok := tfMap["name"].(string); ok && v != "" {
		sheet.Name = aws.String(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		sheet.Title = aws.String(v)
	}
	if v, ok := tfMap["filter_controls"].([]interface{}); ok && len(v) > 0 {
		sheet.FilterControls = expandFilterControls(v)
	}
	if v, ok := tfMap["layouts"].([]interface{}); ok && len(v) > 0 {
		sheet.Layouts = expandLayouts(v)
	}
	if v, ok := tfMap["parameter_controls"].([]interface{}); ok && len(v) > 0 {
		sheet.ParameterControls = expandParameterControls(v)
	}
	if v, ok := tfMap["sheet_control_layouts"].([]interface{}); ok && len(v) > 0 {
		sheet.SheetControlLayouts = expandSheetControlLayouts(v)
	}
	if v, ok := tfMap["text_boxes"].([]interface{}); ok && len(v) > 0 {
		sheet.TextBoxes = expandSheetTextBoxes(v)
	}
	if v, ok := tfMap["visuals"].([]interface{}); ok && len(v) > 0 {
		sheet.Visuals = expandVisuals(v)
	}

	return sheet
}

func expandFilterControls(tfList []interface{}) []*quicksight.FilterControl {
	if len(tfList) == 0 {
		return nil
	}

	var controls []*quicksight.FilterControl
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		control := expandFilterControl(tfMap)
		if control == nil {
			continue
		}

		controls = append(controls, control)
	}

	return controls
}

func expandLayouts(tfList []interface{}) []*quicksight.Layout {
	if len(tfList) == 0 {
		return nil
	}

	var layouts []*quicksight.Layout
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		layout := expandLayout(tfMap)
		if layout == nil {
			continue
		}

		layouts = append(layouts, layout)
	}

	return layouts
}

func expandLayout(tfMap map[string]interface{}) *quicksight.Layout {
	if tfMap == nil {
		return nil
	}

	layout := &quicksight.Layout{}

	if v, ok := tfMap["configuration"].([]interface{}); ok && len(v) > 0 {
		layout.Configuration = expandLayoutConfiguration(v)
	}

	return layout
}

func expandLayoutConfiguration(tfList []interface{}) *quicksight.LayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.LayoutConfiguration{}

	if v, ok := tfMap["free_form_layout"].([]interface{}); ok && len(v) > 0 {
		config.FreeFormLayout = expandFreeFormLayoutConfiguration(v)
	}
	if v, ok := tfMap["grid_layout"].([]interface{}); ok && len(v) > 0 {
		config.GridLayout = expandGridLayoutConfiguration(v)
	}
	if v, ok := tfMap["section_based_layout"].([]interface{}); ok && len(v) > 0 {
		config.SectionBasedLayout = expandSectionBasedLayoutConfiguration(v)
	}

	return config
}

func expandFreeFormLayoutConfiguration(tfList []interface{}) *quicksight.FreeFormLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FreeFormLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		config.Elements = expandFreeFormLayoutElements(v)
	}
	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandFreeFormLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandFreeFormLayoutElements(tfList []interface{}) []*quicksight.FreeFormLayoutElement {
	if len(tfList) == 0 {
		return nil
	}

	var layouts []*quicksight.FreeFormLayoutElement
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		layout := expandFreeFormLayoutElement(tfMap)
		if layout == nil {
			continue
		}

		layouts = append(layouts, layout)
	}

	return layouts
}

func expandFreeFormLayoutElement(tfMap map[string]interface{}) *quicksight.FreeFormLayoutElement {
	if tfMap == nil {
		return nil
	}

	layout := &quicksight.FreeFormLayoutElement{}

	if v, ok := tfMap["element_id"].(string); ok && v != "" {
		layout.ElementId = aws.String(v)
	}
	if v, ok := tfMap["element_type"].(string); ok && v != "" {
		layout.ElementType = aws.String(v)
	}
	if v, ok := tfMap["height"].(string); ok && v != "" {
		layout.Height = aws.String(v)
	}
	if v, ok := tfMap["width"].(string); ok && v != "" {
		layout.Width = aws.String(v)
	}
	if v, ok := tfMap["x_axis_location"].(string); ok && v != "" {
		layout.XAxisLocation = aws.String(v)
	}
	if v, ok := tfMap["y_axis_location"].(string); ok && v != "" {
		layout.YAxisLocation = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		layout.Visibility = aws.String(v)
	}
	if v, ok := tfMap["background_style"].([]interface{}); ok && len(v) > 0 {
		layout.BackgroundStyle = expandFreeFormLayoutElementBackgroundStyle(v)
	}
	if v, ok := tfMap["border_style"].([]interface{}); ok && len(v) > 0 {
		layout.BorderStyle = expandFreeFormLayoutElementBorderStyle(v)
	}
	if v, ok := tfMap["loading_animation"].([]interface{}); ok && len(v) > 0 {
		layout.LoadingAnimation = expandLoadingAnimation(v)
	}
	if v, ok := tfMap["rendering_rules"].([]interface{}); ok && len(v) > 0 {
		layout.RenderingRules = expandSheetElementRenderingRules(v)
	}
	if v, ok := tfMap["selected_border_style"].([]interface{}); ok && len(v) > 0 {
		layout.SelectedBorderStyle = expandFreeFormLayoutElementBorderStyle(v)
	}

	return layout
}

func expandFreeFormLayoutElementBackgroundStyle(tfList []interface{}) *quicksight.FreeFormLayoutElementBackgroundStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FreeFormLayoutElementBackgroundStyle{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		config.Color = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = aws.String(v)
	}
	return config
}

func expandFreeFormLayoutElementBorderStyle(tfList []interface{}) *quicksight.FreeFormLayoutElementBorderStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FreeFormLayoutElementBorderStyle{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		config.Color = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = aws.String(v)
	}
	return config
}

func expandLoadingAnimation(tfList []interface{}) *quicksight.LoadingAnimation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.LoadingAnimation{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = aws.String(v)
	}
	return config
}

func expandSheetElementRenderingRules(tfList []interface{}) []*quicksight.SheetElementRenderingRule {
	if len(tfList) == 0 {
		return nil
	}

	var rules []*quicksight.SheetElementRenderingRule
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := expandSheetElementRenderingRule(tfMap)
		if rule == nil {
			continue
		}

		rules = append(rules, rule)
	}

	return rules
}

func expandSheetElementRenderingRule(tfMap map[string]interface{}) *quicksight.SheetElementRenderingRule {
	if tfMap == nil {
		return nil
	}

	layout := &quicksight.SheetElementRenderingRule{}

	if v, ok := tfMap["expression"].(string); ok && v != "" {
		layout.Expression = aws.String(v)
	}
	if v, ok := tfMap["configuration_overrides"].([]interface{}); ok && len(v) > 0 {
		layout.ConfigurationOverrides = expandSheetElementConfigurationOverrides(v)
	}

	return layout
}

func expandSheetElementConfigurationOverrides(tfList []interface{}) *quicksight.SheetElementConfigurationOverrides {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SheetElementConfigurationOverrides{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = aws.String(v)
	}
	return config
}

func expandGridLayoutConfiguration(tfList []interface{}) *quicksight.GridLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.GridLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		config.Elements = expandGridLayoutElements(v)
	}
	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandGridLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandGridLayoutElements(tfList []interface{}) []*quicksight.GridLayoutElement {
	if len(tfList) == 0 {
		return nil
	}

	var layouts []*quicksight.GridLayoutElement
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		layout := expandGridLayoutElement(tfMap)
		if layout == nil {
			continue
		}

		layouts = append(layouts, layout)
	}

	return layouts
}

func expandGridLayoutElement(tfMap map[string]interface{}) *quicksight.GridLayoutElement {
	if tfMap == nil {
		return nil
	}

	layout := &quicksight.GridLayoutElement{}

	if v, ok := tfMap["element_id"].(string); ok && v != "" {
		layout.ElementId = aws.String(v)
	}
	if v, ok := tfMap["element_type"].(string); ok && v != "" {
		layout.ElementType = aws.String(v)
	}
	if v, ok := tfMap["column_span"].(int); ok && v != 0 {
		layout.ColumnSpan = aws.Int64(int64(v))
	}
	if v, ok := tfMap["row_span"].(int); ok && v != 0 {
		layout.RowSpan = aws.Int64(int64(v))
	}
	if v, ok := tfMap["column_index"].(string); ok && v != "" {
		if i, null, _ := nullable.Int(v).Value(); !null {
			layout.ColumnIndex = aws.Int64(i)
		}
	}
	if v, ok := tfMap["row_index"].(string); ok && v != "" {
		if i, null, _ := nullable.Int(v).Value(); !null {
			layout.RowIndex = aws.Int64(i)
		}
	}

	return layout
}

func expandSectionBasedLayoutConfiguration(tfList []interface{}) *quicksight.SectionBasedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SectionBasedLayoutConfiguration{}

	if v, ok := tfMap["body_sections"].([]interface{}); ok && len(v) > 0 {
		config.BodySections = expandBodySectionConfigurations(v)
	}
	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandSectionBasedLayoutCanvasSizeOptions(v)
	}
	if v, ok := tfMap["footer_sections"].([]interface{}); ok && len(v) > 0 {
		config.FooterSections = expandHeaderFooterSectionConfigurations(v)
	}
	if v, ok := tfMap["header_sections"].([]interface{}); ok && len(v) > 0 {
		config.HeaderSections = expandHeaderFooterSectionConfigurations(v)
	}

	return config
}

func expandBodySectionConfigurations(tfList []interface{}) []*quicksight.BodySectionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.BodySectionConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandBodySectionConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, config)
	}

	return configs
}

func expandBodySectionConfiguration(tfMap map[string]interface{}) *quicksight.BodySectionConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &quicksight.BodySectionConfiguration{}

	if v, ok := tfMap["section_id"].(string); ok && v != "" {
		config.SectionId = aws.String(v)
	}
	if v, ok := tfMap["content"].([]interface{}); ok && len(v) > 0 {
		config.Content = expandBodySectionContent(v)
	}
	if v, ok := tfMap["page_break_configuration"].([]interface{}); ok && len(v) > 0 {
		config.PageBreakConfiguration = expandSectionPageBreakConfiguration(v)
	}

	if v, ok := tfMap["style"].([]interface{}); ok && len(v) > 0 {
		config.Style = expandSectionStyle(v)
	}

	return config
}

func expandBodySectionContent(tfList []interface{}) *quicksight.BodySectionContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.BodySectionContent{}

	if v, ok := tfMap["layout"].([]interface{}); ok && len(v) > 0 {
		config.Layout = expandSectionLayoutConfiguration(v)
	}

	return config
}

func expandSectionLayoutConfiguration(tfList []interface{}) *quicksight.SectionLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SectionLayoutConfiguration{}

	if v, ok := tfMap["free_form_layout"].([]interface{}); ok && len(v) > 0 {
		config.FreeFormLayout = expandFreeFormSectionLayoutConfiguration(v)
	}

	return config
}

func expandFreeFormSectionLayoutConfiguration(tfList []interface{}) *quicksight.FreeFormSectionLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FreeFormSectionLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		config.Elements = expandFreeFormLayoutElements(v)
	}

	return config
}

func expandSectionPageBreakConfiguration(tfList []interface{}) *quicksight.SectionPageBreakConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SectionPageBreakConfiguration{}

	if v, ok := tfMap["after"].([]interface{}); ok && len(v) > 0 {
		config.After = expandSectionAfterPageBreak(v)
	}

	return config
}

func expandSectionAfterPageBreak(tfList []interface{}) *quicksight.SectionAfterPageBreak {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SectionAfterPageBreak{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		config.Status = aws.String(v)
	}

	return config
}

func expandSectionStyle(tfList []interface{}) *quicksight.SectionStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SectionStyle{}

	if v, ok := tfMap["height"].(string); ok && v != "" {
		config.Height = aws.String(v)
	}
	if v, ok := tfMap["padding"].([]interface{}); ok && len(v) > 0 {
		config.Padding = expandSpacing(v)
	}

	return config
}

func expandHeaderFooterSectionConfigurations(tfList []interface{}) []*quicksight.HeaderFooterSectionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.HeaderFooterSectionConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandHeaderFooterSectionConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, config)
	}

	return configs
}

func expandHeaderFooterSectionConfiguration(tfMap map[string]interface{}) *quicksight.HeaderFooterSectionConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &quicksight.HeaderFooterSectionConfiguration{}

	if v, ok := tfMap["section_id"].(string); ok && v != "" {
		config.SectionId = aws.String(v)
	}
	if v, ok := tfMap["layout"].([]interface{}); ok && len(v) > 0 {
		config.Layout = expandSectionLayoutConfiguration(v)
	}
	if v, ok := tfMap["style"].([]interface{}); ok && len(v) > 0 {
		config.Style = expandSectionStyle(v)
	}

	return config
}

func expandParameterControls(tfList []interface{}) []*quicksight.ParameterControl {
	if len(tfList) == 0 {
		return nil
	}

	var controls []*quicksight.ParameterControl
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		control := expandParameterControl(tfMap)
		if control == nil {
			continue
		}

		controls = append(controls, control)
	}

	return controls
}

func expandSheetControlLayouts(tfList []interface{}) []*quicksight.SheetControlLayout {
	if len(tfList) == 0 {
		return nil
	}

	var layouts []*quicksight.SheetControlLayout
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		layout := expandSheetControlLayout(tfMap)
		if layout == nil {
			continue
		}

		layouts = append(layouts, layout)
	}

	return layouts
}

func expandSheetControlLayout(tfMap map[string]interface{}) *quicksight.SheetControlLayout {
	if tfMap == nil {
		return nil
	}

	layout := &quicksight.SheetControlLayout{}

	if v, ok := tfMap["configuration"].([]interface{}); ok && len(v) > 0 {
		layout.Configuration = expandSheetControlLayoutConfiguration(v)
	}

	return layout
}

func expandSheetControlLayoutConfiguration(tfList []interface{}) *quicksight.SheetControlLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SheetControlLayoutConfiguration{}

	if v, ok := tfMap["grid_layout"].([]interface{}); ok && len(v) > 0 {
		config.GridLayout = expandGridLayoutConfiguration(v)
	}

	return config
}

func expandSheetTextBoxes(tfList []interface{}) []*quicksight.SheetTextBox {
	if len(tfList) == 0 {
		return nil
	}

	var boxes []*quicksight.SheetTextBox
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		box := expandSheetTextBox(tfMap)
		if box == nil {
			continue
		}

		boxes = append(boxes, box)
	}

	return boxes
}

func expandSheetTextBox(tfMap map[string]interface{}) *quicksight.SheetTextBox {
	if tfMap == nil {
		return nil
	}

	box := &quicksight.SheetTextBox{}

	if v, ok := tfMap["sheet_text_box_id"].(string); ok && v != "" {
		box.SheetTextBoxId = aws.String(v)
	}
	if v, ok := tfMap["content"].(string); ok && v != "" {
		box.Content = aws.String(v)
	}

	return box
}

func expandVisuals(tfList []interface{}) []*quicksight.Visual {
	if len(tfList) == 0 {
		return nil
	}

	var visuals []*quicksight.Visual
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		visual := expandVisual(tfMap)
		if visual == nil {
			continue
		}

		visuals = append(visuals, visual)
	}

	return visuals
}

func flattenAnalysisDefaults(apiObject *quicksight.AnalysisDefaults) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DefaultNewSheetConfiguration != nil {
		tfMap["default_new_sheet_configuration"] = flattenDefaultNewSheetConfiguration(apiObject.DefaultNewSheetConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenDefaultNewSheetConfiguration(apiObject *quicksight.DefaultNewSheetConfiguration) []interface{} {
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
	if apiObject.SheetContentType != nil {
		tfMap["sheet_content_type"] = aws.StringValue(apiObject.SheetContentType)
	}

	return []interface{}{tfMap}
}

func flattenDefaultInteractiveLayoutConfiguration(apiObject *quicksight.DefaultInteractiveLayoutConfiguration) []interface{} {
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

func flattenDefaultFreeFormLayoutConfiguration(apiObject *quicksight.DefaultFreeFormLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenFreeFormLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutCanvasSizeOptions(apiObject *quicksight.FreeFormLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ScreenCanvasSizeOptions != nil {
		tfMap["screen_canvas_size_options"] = flattenFreeFormLayoutScreenCanvasSizeOptions(apiObject.ScreenCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutScreenCanvasSizeOptions(apiObject *quicksight.FreeFormLayoutScreenCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.OptimizedViewPortWidth != nil {
		tfMap["optimized_view_port_width"] = aws.StringValue(apiObject.OptimizedViewPortWidth)
	}

	return []interface{}{tfMap}
}

func flattenDefaultGridLayoutConfiguration(apiObject *quicksight.DefaultGridLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenGridLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenGridLayoutCanvasSizeOptions(apiObject *quicksight.GridLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ScreenCanvasSizeOptions != nil {
		tfMap["screen_canvas_size_options"] = flattenGridLayoutScreenCanvasSizeOptions(apiObject.ScreenCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenGridLayoutScreenCanvasSizeOptions(apiObject *quicksight.GridLayoutScreenCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.OptimizedViewPortWidth != nil {
		tfMap["optimized_view_port_width"] = aws.StringValue(apiObject.OptimizedViewPortWidth)
	}
	if apiObject.ResizeOption != nil {
		tfMap["resize_option"] = aws.StringValue(apiObject.ResizeOption)
	}

	return []interface{}{tfMap}
}

func flattenDefaultPaginatedLayoutConfiguration(apiObject *quicksight.DefaultPaginatedLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SectionBased != nil {
		tfMap["section_based"] = flattenDefaultSectionBasedLayoutConfiguration(apiObject.SectionBased)
	}

	return []interface{}{tfMap}
}

func flattenDefaultSectionBasedLayoutConfiguration(apiObject *quicksight.DefaultSectionBasedLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenSectionBasedLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenSectionBasedLayoutCanvasSizeOptions(apiObject *quicksight.SectionBasedLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PaperCanvasSizeOptions != nil {
		tfMap["paper_canvas_size_options"] = flattenSectionBasedLayoutPaperCanvasSizeOptions(apiObject.PaperCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenSectionBasedLayoutPaperCanvasSizeOptions(apiObject *quicksight.SectionBasedLayoutPaperCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PaperMargin != nil {
		tfMap["paper_margin"] = flattenSpacing(apiObject.PaperMargin)
	}
	if apiObject.PaperOrientation != nil {
		tfMap["paper_orientation"] = aws.StringValue(apiObject.PaperOrientation)
	}
	if apiObject.PaperSize != nil {
		tfMap["paper_size"] = aws.StringValue(apiObject.PaperSize)
	}

	return []interface{}{tfMap}
}

func flattenSpacing(apiObject *quicksight.Spacing) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Bottom != nil {
		tfMap["bottom"] = aws.StringValue(apiObject.Bottom)
	}
	if apiObject.Left != nil {
		tfMap["left"] = aws.StringValue(apiObject.Left)
	}
	if apiObject.Right != nil {
		tfMap["right"] = aws.StringValue(apiObject.Right)
	}
	if apiObject.Top != nil {
		tfMap["top"] = aws.StringValue(apiObject.Top)
	}

	return []interface{}{tfMap}
}

func flattenLayouts(apiObject []*quicksight.Layout) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"configuration": flattenLayoutConfiguration(config.Configuration),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenLayoutConfiguration(apiObject *quicksight.LayoutConfiguration) []interface{} {
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

func flattenFreeFormLayoutConfiguration(apiObject *quicksight.FreeFormLayoutConfiguration) []interface{} {
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

func flattenFreeFormLayoutElement(apiObject []*quicksight.FreeFormLayoutElement) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"element_id":      aws.StringValue(config.ElementId),
			"element_type":    aws.StringValue(config.ElementType),
			"height":          aws.StringValue(config.Height),
			"width":           aws.StringValue(config.Width),
			"x_axis_location": aws.StringValue(config.XAxisLocation),
			"y_axis_location": aws.StringValue(config.YAxisLocation),
		}
		if config.BackgroundStyle != nil {
			tfMap["background_style"] = flattenFreeFormLayoutElementBackgroundStyle(config.BackgroundStyle)
		}
		if config.BorderStyle != nil {
			tfMap["border_style"] = flattenFreeFormLayoutElementBorderStyle(config.BorderStyle)
		}
		if config.LoadingAnimation != nil {
			tfMap["loading_animation"] = flattenLoadingAnimation(config.LoadingAnimation)
		}
		if config.RenderingRules != nil {
			tfMap["rendering_rules"] = flattenSheetElementRenderingRule(config.RenderingRules)
		}
		if config.SelectedBorderStyle != nil {
			tfMap["selected_border_style"] = flattenFreeFormLayoutElementBorderStyle(config.SelectedBorderStyle)
		}
		if config.Visibility != nil {
			tfMap["visibility"] = aws.StringValue(config.Visibility)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFreeFormLayoutElementBackgroundStyle(apiObject *quicksight.FreeFormLayoutElementBackgroundStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.StringValue(apiObject.Color)
	}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutElementBorderStyle(apiObject *quicksight.FreeFormLayoutElementBorderStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.StringValue(apiObject.Color)
	}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenLoadingAnimation(apiObject *quicksight.LoadingAnimation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenSheetElementRenderingRule(apiObject []*quicksight.SheetElementRenderingRule) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.ConfigurationOverrides != nil {
			tfMap["configuration_overrides"] = flattenSheetElementConfigurationOverrides(config.ConfigurationOverrides)
		}
		if config.Expression != nil {
			tfMap["expression"] = aws.StringValue(config.Expression)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetElementConfigurationOverrides(apiObject *quicksight.SheetElementConfigurationOverrides) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Visibility != nil {
		tfMap["visibility"] = aws.StringValue(apiObject.Visibility)
	}

	return []interface{}{tfMap}
}

func flattenGridLayoutConfiguration(apiObject *quicksight.GridLayoutConfiguration) []interface{} {
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

func flattenGridLayoutElement(apiObject []*quicksight.GridLayoutElement) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"column_span":  aws.Int64Value(config.ColumnSpan),
			"element_id":   aws.StringValue(config.ElementId),
			"element_type": aws.StringValue(config.ElementType),
			"row_span":     aws.Int64Value(config.RowSpan),
		}
		if config.ColumnIndex != nil {
			tfMap["column_index"] = strconv.FormatInt(aws.Int64Value(config.ColumnIndex), 10)
		}
		if config.RowIndex != nil {
			tfMap["row_index"] = strconv.FormatInt(aws.Int64Value(config.RowIndex), 10)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSectionBasedLayoutConfiguration(apiObject *quicksight.SectionBasedLayoutConfiguration) []interface{} {
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

func flattenBodySectionConfiguration(apiObject []*quicksight.BodySectionConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"content":    flattenBodySectionContent(config.Content),
			"section_id": aws.StringValue(config.SectionId),
		}
		if config.PageBreakConfiguration != nil {
			tfMap["page_break_configuration"] = flattenSectionPageBreakConfiguration(config.PageBreakConfiguration)
		}
		if config.Style != nil {
			tfMap["style"] = flattenSectionStyle(config.Style)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenBodySectionContent(apiObject *quicksight.BodySectionContent) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Layout != nil {
		tfMap["layout"] = flattenSectionLayoutConfiguration(apiObject.Layout)
	}

	return []interface{}{tfMap}
}

func flattenSectionLayoutConfiguration(apiObject *quicksight.SectionLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FreeFormLayout != nil {
		tfMap["free_form_layout"] = flattenFreeFormSectionLayoutConfiguration(apiObject.FreeFormLayout)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormSectionLayoutConfiguration(apiObject *quicksight.FreeFormSectionLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Elements != nil {
		tfMap["free_form_layout"] = flattenFreeFormLayoutElement(apiObject.Elements)
	}

	return []interface{}{tfMap}
}

func flattenSectionPageBreakConfiguration(apiObject *quicksight.SectionPageBreakConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.After != nil {
		tfMap["after"] = flattenSectionAfterPageBreak(apiObject.After)
	}

	return []interface{}{tfMap}
}

func flattenSectionAfterPageBreak(apiObject *quicksight.SectionAfterPageBreak) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Status != nil {
		tfMap["status"] = aws.StringValue(apiObject.Status)
	}

	return []interface{}{tfMap}
}

func flattenSectionStyle(apiObject *quicksight.SectionStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Height != nil {
		tfMap["height"] = aws.StringValue(apiObject.Height)
	}
	if apiObject.Padding != nil {
		tfMap["padding"] = flattenSpacing(apiObject.Padding)
	}

	return []interface{}{tfMap}
}

func flattenHeaderFooterSectionConfiguration(apiObject []*quicksight.HeaderFooterSectionConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"section_id": aws.StringValue(config.SectionId),
		}
		if config.Layout != nil {
			tfMap["layout"] = flattenSectionLayoutConfiguration(config.Layout)
		}
		if config.Style != nil {
			tfMap["style"] = flattenSectionStyle(config.Style)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetControlLayouts(apiObject []*quicksight.SheetControlLayout) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"configuration": flattenSheetControlLayoutConfiguration(config.Configuration),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetControlLayoutConfiguration(apiObject *quicksight.SheetControlLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.GridLayout != nil {
		tfMap["grid_layout"] = flattenGridLayoutConfiguration(apiObject.GridLayout)
	}

	return []interface{}{tfMap}
}
