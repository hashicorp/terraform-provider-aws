// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
							"sheet_content_type":               stringSchema(false, enum.Validate[types.SheetContentType]()),
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
													"resize_option": stringSchema(true, enum.Validate[types.ResizeOption]()),
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
				"paper_orientation": stringSchema(false, enum.Validate[types.PaperOrientation]()),
				"paper_size":        stringSchema(false, enum.Validate[types.PaperSize]()),
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
																			"status": stringSchema(false, enum.Validate[types.SectionPageBreakStatus]()),
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
								Type:             schema.TypeInt,
								Required:         true,
								ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 36)),
							},
							"element_id":   idSchema(),
							"element_type": stringSchema(true, enum.Validate[types.LayoutElementType]()),
							"row_span": {
								Type:             schema.TypeInt,
								Required:         true,
								ValidateDiagFunc: validation.ToDiagFunc(validation.IntBetween(1, 21)),
							},
							"column_index": {
								Type:             nullable.TypeNullableInt,
								Optional:         true,
								ValidateDiagFunc: validation.ToDiagFunc(nullable.ValidateTypeStringNullableIntBetween(0, 35)),
							},
							"row_index": {
								Type:             nullable.TypeNullableInt,
								Optional:         true,
								ValidateDiagFunc: validation.ToDiagFunc(nullable.ValidateTypeStringNullableIntBetween(0, 9009)),
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
										"resize_option": stringSchema(true, enum.Validate[types.ResizeOption]()),
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
				"element_type": stringSchema(true, enum.Validate[types.LayoutElementType]()),
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
							"color":      stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`), ""))),
							"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
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
							"color":      stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`), ""))),
							"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
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
							"visibility": stringSchema(false, enum.Validate[types.Visibility]()),
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
										"visibility": stringSchema(false, enum.Validate[types.Visibility]())},
								},
							},
							"expression": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 4096))),
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
							"color":      stringSchema(false, validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^#[0-9A-F]{6}(?:[0-9A-F]{2})?$`), ""))),
							"visibility": stringSchema(false, enum.Validate[types.Visibility]())},
					},
				},
				"visibility": stringSchema(false, enum.Validate[types.Visibility]())},
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

func expandAnalysisDefaults(tfList []interface{}) *types.AnalysisDefaults {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	defaults := &types.AnalysisDefaults{}

	if v, ok := tfMap["default_new_sheet_configuration"].([]interface{}); ok && len(v) > 0 {
		defaults.DefaultNewSheetConfiguration = expandDefaultNewSheetConfiguration(v)
	}

	return defaults
}

func expandDefaultNewSheetConfiguration(tfList []interface{}) *types.DefaultNewSheetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DefaultNewSheetConfiguration{}

	if v, ok := tfMap["interactive_layout_configuration"].([]interface{}); ok && len(v) > 0 {
		config.InteractiveLayoutConfiguration = expandDefaultInteractiveLayoutConfiguration(v)
	}

	if v, ok := tfMap["paginated_layout_configuration"].([]interface{}); ok && len(v) > 0 {
		config.PaginatedLayoutConfiguration = expandDefaultPaginatedLayoutConfiguration(v)
	}

	if v, ok := tfMap["sheet_content_type"].(string); ok && v != "" {
		config.SheetContentType = types.SheetContentType(v)
	}

	return config
}

func expandDefaultInteractiveLayoutConfiguration(tfList []interface{}) *types.DefaultInteractiveLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DefaultInteractiveLayoutConfiguration{}

	if v, ok := tfMap["free_form"].([]interface{}); ok && len(v) > 0 {
		config.FreeForm = expandDefaultFreeFormLayoutConfiguration(v)
	}

	if v, ok := tfMap["grid"].([]interface{}); ok && len(v) > 0 {
		config.Grid = expandDefaultGridLayoutConfiguration(v)
	}

	return config
}

func expandDefaultFreeFormLayoutConfiguration(tfList []interface{}) *types.DefaultFreeFormLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DefaultFreeFormLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandFreeFormLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandFreeFormLayoutCanvasSizeOptions(tfList []interface{}) *types.FreeFormLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.FreeFormLayoutCanvasSizeOptions{}

	if v, ok := tfMap["screen_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		options.ScreenCanvasSizeOptions = expandFreeFormLayoutScreenCanvasSizeOptions(v)
	}

	return options
}

func expandFreeFormLayoutScreenCanvasSizeOptions(tfList []interface{}) *types.FreeFormLayoutScreenCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.FreeFormLayoutScreenCanvasSizeOptions{}

	if v, ok := tfMap["optimized_view_port_width"].(string); ok && v != "" {
		options.OptimizedViewPortWidth = aws.String(v)
	}

	return options
}

func expandDefaultGridLayoutConfiguration(tfList []interface{}) *types.DefaultGridLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DefaultGridLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandGridLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandGridLayoutCanvasSizeOptions(tfList []interface{}) *types.GridLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.GridLayoutCanvasSizeOptions{}

	if v, ok := tfMap["screen_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		options.ScreenCanvasSizeOptions = expandGridLayoutScreenCanvasSizeOptions(v)
	}

	return options
}

func expandGridLayoutScreenCanvasSizeOptions(tfList []interface{}) *types.GridLayoutScreenCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.GridLayoutScreenCanvasSizeOptions{}

	if v, ok := tfMap["optimized_view_port_width"].(string); ok && v != "" {
		options.OptimizedViewPortWidth = aws.String(v)
	}
	if v, ok := tfMap["resize_option"].(string); ok && v != "" {
		options.ResizeOption = types.ResizeOption(v)
	}

	return options
}

func expandDefaultPaginatedLayoutConfiguration(tfList []interface{}) *types.DefaultPaginatedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DefaultPaginatedLayoutConfiguration{}

	if v, ok := tfMap["section_based"].([]interface{}); ok && len(v) > 0 {
		config.SectionBased = expandDefaultSectionBasedLayoutConfiguration(v)
	}

	return config
}

func expandDefaultSectionBasedLayoutConfiguration(tfList []interface{}) *types.DefaultSectionBasedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DefaultSectionBasedLayoutConfiguration{}

	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandSectionBasedLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandSectionBasedLayoutCanvasSizeOptions(tfList []interface{}) *types.SectionBasedLayoutCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.SectionBasedLayoutCanvasSizeOptions{}

	if v, ok := tfMap["paper_canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		options.PaperCanvasSizeOptions = expandSectionBasedLayoutPaperCanvasSizeOptions(v)
	}

	return options
}

func expandSectionBasedLayoutPaperCanvasSizeOptions(tfList []interface{}) *types.SectionBasedLayoutPaperCanvasSizeOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &types.SectionBasedLayoutPaperCanvasSizeOptions{}

	if v, ok := tfMap["paper_margin"].([]interface{}); ok && len(v) > 0 {
		options.PaperMargin = expandSpacing(v)
	}
	if v, ok := tfMap["paper_orientation"].(string); ok && v != "" {
		options.PaperOrientation = types.PaperOrientation(v)
	}
	if v, ok := tfMap["paper_size"].(string); ok && v != "" {
		options.PaperSize = types.PaperSize(v)
	}

	return options
}

func expandSpacing(tfList []interface{}) *types.Spacing {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	spacing := &types.Spacing{}

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

func expandSheetDefinition(tfMap map[string]interface{}) *types.SheetDefinition {
	if tfMap == nil {
		return nil
	}

	sheet := &types.SheetDefinition{}

	if v, ok := tfMap["sheet_id"].(string); ok && v != "" {
		sheet.SheetId = aws.String(v)
	}
	if v, ok := tfMap["content_type"].(string); ok && v != "" {
		sheet.ContentType = types.SheetContentType(v)
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

func expandFilterControls(tfList []interface{}) []types.FilterControl {
	if len(tfList) == 0 {
		return nil
	}

	var controls []types.FilterControl
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		control := expandFilterControl(tfMap)
		if control == nil {
			continue
		}

		controls = append(controls, *control)
	}

	return controls
}

func expandLayouts(tfList []interface{}) []types.Layout {
	if len(tfList) == 0 {
		return nil
	}

	var layouts []types.Layout
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		layout := expandLayout(tfMap)
		if layout == nil {
			continue
		}

		layouts = append(layouts, *layout)
	}

	return layouts
}

func expandLayout(tfMap map[string]interface{}) *types.Layout {
	if tfMap == nil {
		return nil
	}

	layout := &types.Layout{}

	if v, ok := tfMap["configuration"].([]interface{}); ok && len(v) > 0 {
		layout.Configuration = expandLayoutConfiguration(v)
	}

	return layout
}

func expandLayoutConfiguration(tfList []interface{}) *types.LayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.LayoutConfiguration{}

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

func expandFreeFormLayoutConfiguration(tfList []interface{}) *types.FreeFormLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FreeFormLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		config.Elements = expandFreeFormLayoutElements(v)
	}
	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandFreeFormLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandFreeFormLayoutElements(tfList []interface{}) []types.FreeFormLayoutElement {
	if len(tfList) == 0 {
		return nil
	}

	var layouts []types.FreeFormLayoutElement
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		layout := expandFreeFormLayoutElement(tfMap)
		if layout == nil {
			continue
		}

		layouts = append(layouts, *layout)
	}

	return layouts
}

func expandFreeFormLayoutElement(tfMap map[string]interface{}) *types.FreeFormLayoutElement {
	if tfMap == nil {
		return nil
	}

	layout := &types.FreeFormLayoutElement{}

	if v, ok := tfMap["element_id"].(string); ok && v != "" {
		layout.ElementId = aws.String(v)
	}
	if v, ok := tfMap["element_type"].(string); ok && v != "" {
		layout.ElementType = types.LayoutElementType(v)
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
		layout.Visibility = types.Visibility(v)
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

func expandFreeFormLayoutElementBackgroundStyle(tfList []interface{}) *types.FreeFormLayoutElementBackgroundStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FreeFormLayoutElementBackgroundStyle{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		config.Color = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = types.Visibility(v)
	}
	return config
}

func expandFreeFormLayoutElementBorderStyle(tfList []interface{}) *types.FreeFormLayoutElementBorderStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FreeFormLayoutElementBorderStyle{}

	if v, ok := tfMap["color"].(string); ok && v != "" {
		config.Color = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = types.Visibility(v)
	}
	return config
}

func expandLoadingAnimation(tfList []interface{}) *types.LoadingAnimation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.LoadingAnimation{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = types.Visibility(v)
	}
	return config
}

func expandSheetElementRenderingRules(tfList []interface{}) []types.SheetElementRenderingRule {
	if len(tfList) == 0 {
		return nil
	}

	var rules []types.SheetElementRenderingRule
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := expandSheetElementRenderingRule(tfMap)
		if rule == nil {
			continue
		}

		rules = append(rules, *rule)
	}

	return rules
}

func expandSheetElementRenderingRule(tfMap map[string]interface{}) *types.SheetElementRenderingRule {
	if tfMap == nil {
		return nil
	}

	layout := &types.SheetElementRenderingRule{}

	if v, ok := tfMap["expression"].(string); ok && v != "" {
		layout.Expression = aws.String(v)
	}
	if v, ok := tfMap["configuration_overrides"].([]interface{}); ok && len(v) > 0 {
		layout.ConfigurationOverrides = expandSheetElementConfigurationOverrides(v)
	}

	return layout
}

func expandSheetElementConfigurationOverrides(tfList []interface{}) *types.SheetElementConfigurationOverrides {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SheetElementConfigurationOverrides{}

	if v, ok := tfMap["visibility"].(string); ok && v != "" {
		config.Visibility = types.Visibility(v)
	}
	return config
}

func expandGridLayoutConfiguration(tfList []interface{}) *types.GridLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.GridLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		config.Elements = expandGridLayoutElements(v)
	}
	if v, ok := tfMap["canvas_size_options"].([]interface{}); ok && len(v) > 0 {
		config.CanvasSizeOptions = expandGridLayoutCanvasSizeOptions(v)
	}

	return config
}

func expandGridLayoutElements(tfList []interface{}) []types.GridLayoutElement {
	if len(tfList) == 0 {
		return nil
	}

	var layouts []types.GridLayoutElement
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		layout := expandGridLayoutElement(tfMap)
		if layout == nil {
			continue
		}

		layouts = append(layouts, *layout)
	}

	return layouts
}

func expandGridLayoutElement(tfMap map[string]interface{}) *types.GridLayoutElement {
	if tfMap == nil {
		return nil
	}

	layout := &types.GridLayoutElement{}

	if v, ok := tfMap["element_id"].(string); ok && v != "" {
		layout.ElementId = aws.String(v)
	}
	if v, ok := tfMap["element_type"].(string); ok && v != "" {
		layout.ElementType = types.LayoutElementType(v)
	}
	if v, ok := tfMap["column_span"].(int); ok && v != 0 {
		layout.ColumnSpan = aws.Int32(int32(v))
	}
	if v, ok := tfMap["row_span"].(int); ok && v != 0 {
		layout.RowSpan = aws.Int32(int32(v))
	}
	if v, null, _ := nullable.Int(tfMap["column_index"].(string)).Value(); !null {
		layout.ColumnIndex = aws.Int32(int32(v))
	}
	if v, null, _ := nullable.Int(tfMap["row_index"].(string)).Value(); !null {
		layout.RowIndex = aws.Int32(int32(v))
	}

	return layout
}

func expandSectionBasedLayoutConfiguration(tfList []interface{}) *types.SectionBasedLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SectionBasedLayoutConfiguration{}

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

func expandBodySectionConfigurations(tfList []interface{}) []types.BodySectionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []types.BodySectionConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandBodySectionConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, *config)
	}

	return configs
}

func expandBodySectionConfiguration(tfMap map[string]interface{}) *types.BodySectionConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &types.BodySectionConfiguration{}

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

func expandBodySectionContent(tfList []interface{}) *types.BodySectionContent {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.BodySectionContent{}

	if v, ok := tfMap["layout"].([]interface{}); ok && len(v) > 0 {
		config.Layout = expandSectionLayoutConfiguration(v)
	}

	return config
}

func expandSectionLayoutConfiguration(tfList []interface{}) *types.SectionLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SectionLayoutConfiguration{}

	if v, ok := tfMap["free_form_layout"].([]interface{}); ok && len(v) > 0 {
		config.FreeFormLayout = expandFreeFormSectionLayoutConfiguration(v)
	}

	return config
}

func expandFreeFormSectionLayoutConfiguration(tfList []interface{}) *types.FreeFormSectionLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FreeFormSectionLayoutConfiguration{}

	if v, ok := tfMap["elements"].([]interface{}); ok && len(v) > 0 {
		config.Elements = expandFreeFormLayoutElements(v)
	}

	return config
}

func expandSectionPageBreakConfiguration(tfList []interface{}) *types.SectionPageBreakConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SectionPageBreakConfiguration{}

	if v, ok := tfMap["after"].([]interface{}); ok && len(v) > 0 {
		config.After = expandSectionAfterPageBreak(v)
	}

	return config
}

func expandSectionAfterPageBreak(tfList []interface{}) *types.SectionAfterPageBreak {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SectionAfterPageBreak{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		config.Status = types.SectionPageBreakStatus(v)
	}

	return config
}

func expandSectionStyle(tfList []interface{}) *types.SectionStyle {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SectionStyle{}

	if v, ok := tfMap["height"].(string); ok && v != "" {
		config.Height = aws.String(v)
	}
	if v, ok := tfMap["padding"].([]interface{}); ok && len(v) > 0 {
		config.Padding = expandSpacing(v)
	}

	return config
}

func expandHeaderFooterSectionConfigurations(tfList []interface{}) []types.HeaderFooterSectionConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []types.HeaderFooterSectionConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandHeaderFooterSectionConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, *config)
	}

	return configs
}

func expandHeaderFooterSectionConfiguration(tfMap map[string]interface{}) *types.HeaderFooterSectionConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &types.HeaderFooterSectionConfiguration{}

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

func expandParameterControls(tfList []interface{}) []types.ParameterControl {
	if len(tfList) == 0 {
		return nil
	}

	var controls []types.ParameterControl
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		control := expandParameterControl(tfMap)
		if control == nil {
			continue
		}

		controls = append(controls, *control)
	}

	return controls
}

func expandSheetControlLayouts(tfList []interface{}) []types.SheetControlLayout {
	if len(tfList) == 0 {
		return nil
	}

	var layouts []types.SheetControlLayout
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		layout := expandSheetControlLayout(tfMap)
		if layout == nil {
			continue
		}

		layouts = append(layouts, *layout)
	}

	return layouts
}

func expandSheetControlLayout(tfMap map[string]interface{}) *types.SheetControlLayout {
	if tfMap == nil {
		return nil
	}

	layout := &types.SheetControlLayout{}

	if v, ok := tfMap["configuration"].([]interface{}); ok && len(v) > 0 {
		layout.Configuration = expandSheetControlLayoutConfiguration(v)
	}

	return layout
}

func expandSheetControlLayoutConfiguration(tfList []interface{}) *types.SheetControlLayoutConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SheetControlLayoutConfiguration{}

	if v, ok := tfMap["grid_layout"].([]interface{}); ok && len(v) > 0 {
		config.GridLayout = expandGridLayoutConfiguration(v)
	}

	return config
}

func expandSheetTextBoxes(tfList []interface{}) []types.SheetTextBox {
	if len(tfList) == 0 {
		return nil
	}

	var boxes []types.SheetTextBox
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		box := expandSheetTextBox(tfMap)
		if box == nil {
			continue
		}

		boxes = append(boxes, *box)
	}

	return boxes
}

func expandSheetTextBox(tfMap map[string]interface{}) *types.SheetTextBox {
	if tfMap == nil {
		return nil
	}

	box := &types.SheetTextBox{}

	if v, ok := tfMap["sheet_text_box_id"].(string); ok && v != "" {
		box.SheetTextBoxId = aws.String(v)
	}
	if v, ok := tfMap["content"].(string); ok && v != "" {
		box.Content = aws.String(v)
	}

	return box
}

func expandVisuals(tfList []interface{}) []types.Visual {
	if len(tfList) == 0 {
		return nil
	}

	var visuals []types.Visual
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		visual := expandVisual(tfMap)
		if visual == nil {
			continue
		}

		visuals = append(visuals, *visual)
	}

	return visuals
}

func flattenAnalysisDefaults(apiObject *types.AnalysisDefaults) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DefaultNewSheetConfiguration != nil {
		tfMap["default_new_sheet_configuration"] = flattenDefaultNewSheetConfiguration(apiObject.DefaultNewSheetConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenDefaultNewSheetConfiguration(apiObject *types.DefaultNewSheetConfiguration) []interface{} {
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
	tfMap["sheet_content_type"] = types.SheetContentType(apiObject.SheetContentType)

	return []interface{}{tfMap}
}

func flattenDefaultInteractiveLayoutConfiguration(apiObject *types.DefaultInteractiveLayoutConfiguration) []interface{} {
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

func flattenDefaultFreeFormLayoutConfiguration(apiObject *types.DefaultFreeFormLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenFreeFormLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutCanvasSizeOptions(apiObject *types.FreeFormLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ScreenCanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenFreeFormLayoutScreenCanvasSizeOptions(apiObject.ScreenCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutScreenCanvasSizeOptions(apiObject *types.FreeFormLayoutScreenCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.OptimizedViewPortWidth != nil {
		tfMap["optimized_view_port_width"] = apiObject.OptimizedViewPortWidth
	}

	return []interface{}{tfMap}
}

func flattenDefaultGridLayoutConfiguration(apiObject *types.DefaultGridLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenGridLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenGridLayoutCanvasSizeOptions(apiObject *types.GridLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.ScreenCanvasSizeOptions != nil {
		tfMap["screen_canvas_size_options"] = flattenGridLayoutScreenCanvasSizeOptions(apiObject.ScreenCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenGridLayoutScreenCanvasSizeOptions(apiObject *types.GridLayoutScreenCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.OptimizedViewPortWidth != nil {
		tfMap["optimized_view_port_width"] = aws.ToString(apiObject.OptimizedViewPortWidth)
	}
	tfMap["resize_option"] = types.ResizeOption(apiObject.ResizeOption)

	return []interface{}{tfMap}
}

func flattenDefaultPaginatedLayoutConfiguration(apiObject *types.DefaultPaginatedLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SectionBased != nil {
		tfMap["section_based"] = flattenDefaultSectionBasedLayoutConfiguration(apiObject.SectionBased)
	}

	return []interface{}{tfMap}
}

func flattenDefaultSectionBasedLayoutConfiguration(apiObject *types.DefaultSectionBasedLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CanvasSizeOptions != nil {
		tfMap["canvas_size_options"] = flattenSectionBasedLayoutCanvasSizeOptions(apiObject.CanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenSectionBasedLayoutCanvasSizeOptions(apiObject *types.SectionBasedLayoutCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PaperCanvasSizeOptions != nil {
		tfMap["paper_canvas_size_options"] = flattenSectionBasedLayoutPaperCanvasSizeOptions(apiObject.PaperCanvasSizeOptions)
	}

	return []interface{}{tfMap}
}

func flattenSectionBasedLayoutPaperCanvasSizeOptions(apiObject *types.SectionBasedLayoutPaperCanvasSizeOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.PaperMargin != nil {
		tfMap["paper_margin"] = flattenSpacing(apiObject.PaperMargin)
	}
	tfMap["paper_orientation"] = types.PaperOrientation(apiObject.PaperOrientation)
	tfMap["paper_size"] = types.PaperSize(apiObject.PaperSize)

	return []interface{}{tfMap}
}

func flattenSpacing(apiObject *types.Spacing) []interface{} {
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

func flattenLayouts(apiObject []types.Layout) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{
			"configuration": flattenLayoutConfiguration(config.Configuration),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenLayoutConfiguration(apiObject *types.LayoutConfiguration) []interface{} {
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

func flattenFreeFormLayoutConfiguration(apiObject *types.FreeFormLayoutConfiguration) []interface{} {
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

func flattenFreeFormLayoutElement(apiObject []types.FreeFormLayoutElement) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		tfMap := map[string]interface{}{
			"element_id":      aws.ToString(config.ElementId),
			"element_type":    types.LayoutElementType(config.ElementType),
			"height":          aws.ToString(config.Height),
			"width":           aws.ToString(config.Width),
			"x_axis_location": aws.ToString(config.XAxisLocation),
			"y_axis_location": aws.ToString(config.YAxisLocation),
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
		tfMap["visibility"] = types.Visibility(config.Visibility)

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFreeFormLayoutElementBackgroundStyle(apiObject *types.FreeFormLayoutElementBackgroundStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenFreeFormLayoutElementBorderStyle(apiObject *types.FreeFormLayoutElementBorderStyle) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Color != nil {
		tfMap["color"] = aws.ToString(apiObject.Color)
	}

	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenLoadingAnimation(apiObject *types.LoadingAnimation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenSheetElementRenderingRule(apiObject []types.SheetElementRenderingRule) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		tfMap := map[string]interface{}{}
		if config.ConfigurationOverrides != nil {
			tfMap["configuration_overrides"] = flattenSheetElementConfigurationOverrides(config.ConfigurationOverrides)
		}
		if config.Expression != nil {
			tfMap["expression"] = aws.ToString(config.Expression)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetElementConfigurationOverrides(apiObject *types.SheetElementConfigurationOverrides) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["visibility"] = types.Visibility(apiObject.Visibility)

	return []interface{}{tfMap}
}

func flattenGridLayoutConfiguration(apiObject *types.GridLayoutConfiguration) []interface{} {
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

func flattenGridLayoutElement(apiObject []types.GridLayoutElement) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		tfMap := map[string]interface{}{
			"column_span":  aws.ToInt32(config.ColumnSpan),
			"element_id":   aws.ToString(config.ElementId),
			"element_type": types.LayoutElementType(config.ElementType),
			"row_span":     aws.ToInt32(config.RowSpan),
		}
		if config.ColumnIndex != nil {
			tfMap["column_index"] = aws.ToInt32(config.ColumnIndex)
		}
		if config.RowIndex != nil {
			tfMap["row_index"] = aws.ToInt32(config.RowIndex)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSectionBasedLayoutConfiguration(apiObject *types.SectionBasedLayoutConfiguration) []interface{} {
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

func flattenBodySectionConfiguration(apiObject []types.BodySectionConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		tfMap := map[string]interface{}{
			"content":    flattenBodySectionContent(config.Content),
			"section_id": aws.ToString(config.SectionId),
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

func flattenBodySectionContent(apiObject *types.BodySectionContent) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Layout != nil {
		tfMap["layout"] = flattenSectionLayoutConfiguration(apiObject.Layout)
	}

	return []interface{}{tfMap}
}

func flattenSectionLayoutConfiguration(apiObject *types.SectionLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.FreeFormLayout != nil {
		tfMap["free_form_layout"] = flattenFreeFormSectionLayoutConfiguration(apiObject.FreeFormLayout)
	}

	return []interface{}{tfMap}
}

func flattenFreeFormSectionLayoutConfiguration(apiObject *types.FreeFormSectionLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Elements != nil {
		tfMap["free_form_layout"] = apiObject.Elements
	}

	return []interface{}{tfMap}
}

func flattenSectionPageBreakConfiguration(apiObject *types.SectionPageBreakConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.After != nil {
		tfMap["after"] = flattenSectionAfterPageBreak(apiObject.After)
	}

	return []interface{}{tfMap}
}

func flattenSectionAfterPageBreak(apiObject *types.SectionAfterPageBreak) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["status"] = types.SectionPageBreakStatus(apiObject.Status)

	return []interface{}{tfMap}
}

func flattenSectionStyle(apiObject *types.SectionStyle) []interface{} {
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

func flattenHeaderFooterSectionConfiguration(apiObject []types.HeaderFooterSectionConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		tfMap := map[string]interface{}{
			"section_id": *config.SectionId,
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

func flattenSheetControlLayouts(apiObject []types.SheetControlLayout) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		tfMap := map[string]interface{}{
			"configuration": flattenSheetControlLayoutConfiguration(config.Configuration),
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenSheetControlLayoutConfiguration(apiObject *types.SheetControlLayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.GridLayout != nil {
		tfMap["grid_layout"] = flattenGridLayoutConfiguration(apiObject.GridLayout)
	}

	return []interface{}{tfMap}
}
