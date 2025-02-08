// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var filtersSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 20,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"category_filter":         categoryFilterSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CategoryFilter.html
				"numeric_equality_filter": numericEqualityFilterSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericEqualityFilter.html
				"numeric_range_filter":    numericRangeFilterSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericRangeFilter.html
				"relative_dates_filter":   relativeDatesFilterSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RelativeDatesFilter.html
				"time_equality_filter":    timeEqualityFilterSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeEqualityFilter.html
				"time_range_filter":       timeRangeFilterSchema(),       // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeRangeFilter.html
				"top_bottom_filter":       topBottomFilterSchema(),       // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TopBottomFilter.html
			},
		},
	}
})

func categoryFilterSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CategoryFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column": columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				names.AttrConfiguration: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CategoryFilterConfiguration.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"custom_filter_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomFilterConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"match_operator": stringEnumSchema[awstypes.CategoryFilterMatchOperator](attrRequired),
										"null_option":    stringEnumSchema[awstypes.FilterNullOption](attrRequired),
										"category_value": stringLenBetweenSchema(attrOptional, 1, 512),
										"parameter_name": {
											Type:     schema.TypeString,
											Optional: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 2048),
												validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+`), ""),
											),
										},
										"select_all_options": stringEnumSchema[awstypes.CategoryFilterSelectAllOptions](attrOptional),
									},
								},
							},
							"custom_filter_list_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomFilterListConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"match_operator": stringEnumSchema[awstypes.CategoryFilterMatchOperator](attrRequired),
										"null_option":    stringEnumSchema[awstypes.FilterNullOption](attrRequired),
										"category_values": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100000,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validation.StringLenBetween(1, 512),
											},
										},
										"select_all_options": stringEnumSchema[awstypes.CategoryFilterSelectAllOptions](attrOptional),
									},
								},
							},
							"filter_list_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterListConfiguration.html
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"match_operator": stringEnumSchema[awstypes.CategoryFilterMatchOperator](attrRequired),
										"category_values": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100000,
											Elem: &schema.Schema{
												Type:         schema.TypeString,
												ValidateFunc: validation.StringLenBetween(1, 512),
											},
										},
										"select_all_options": stringEnumSchema[awstypes.CategoryFilterSelectAllOptions](attrOptional),
									},
								},
							},
						},
					},
				},
				"filter_id":                            idSchema(),
				"default_filter_control_configuration": defaultFilterControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlConfiguration.html
			},
		},
	}
}

func numericEqualityFilterSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericEqualityFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column":                               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                            idSchema(),
				"match_operator":                       stringEnumSchema[awstypes.CategoryFilterMatchOperator](attrRequired),
				"null_option":                          stringEnumSchema[awstypes.FilterNullOption](attrRequired),
				"aggregation_function":                 aggregationFunctionSchema(false),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
				"default_filter_control_configuration": defaultFilterControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlConfiguration.html
				"parameter_name":                       parameterNameSchema(false),
				"select_all_options":                   stringEnumSchema[awstypes.NumericFilterSelectAllOptions](attrOptional),
				names.AttrValue: {
					Type:     schema.TypeFloat,
					Optional: true,
				},
			},
		},
	}
}

func numericRangeFilterSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericRangeFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column":                               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                            idSchema(),
				"null_option":                          stringEnumSchema[awstypes.FilterNullOption](attrRequired),
				"aggregation_function":                 aggregationFunctionSchema(false),          // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
				"default_filter_control_configuration": defaultFilterControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlConfiguration.html
				"include_maximum": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"include_minimum": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"range_maximum":      numericRangeFilterValueSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericRangeFilterValue.html
				"range_minimum":      numericRangeFilterValueSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericRangeFilterValue.html
				"select_all_options": stringEnumSchema[awstypes.NumericFilterSelectAllOptions](attrOptional),
			},
		},
	}
}

func relativeDatesFilterSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RelativeDatesFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"anchor_date_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AnchorDateConfiguration.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"anchor_option":  stringEnumSchema[awstypes.AnchorOption](attrOptional),
							"parameter_name": parameterNameSchema(false),
						},
					},
				},
				"column":                               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                            idSchema(),
				"null_option":                          stringEnumSchema[awstypes.FilterNullOption](attrRequired),
				"relative_date_type":                   stringEnumSchema[awstypes.RelativeDateType](attrRequired),
				"time_granularity":                     stringEnumSchema[awstypes.TimeGranularity](attrRequired),
				"default_filter_control_configuration": defaultFilterControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlConfiguration.html
				"exclude_period_configuration":         excludePeriodConfigurationSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExcludePeriodConfiguration.html
				"minimum_granularity":                  stringEnumSchema[awstypes.TimeGranularity](attrOptional),
				"parameter_name":                       parameterNameSchema(false),
				"relative_date_value": {
					Type:     schema.TypeInt,
					Optional: true,
				},
			},
		},
	}
}

func timeEqualityFilterSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeEqualityFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column":                               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                            idSchema(),
				"default_filter_control_configuration": defaultFilterControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlConfiguration.html
				"parameter_name":                       parameterNameSchema(false),
				"time_granularity":                     stringEnumSchema[awstypes.TimeGranularity](attrOptional),
				names.AttrValue: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidUTCTimestamp,
				},
			},
		},
	}
}

func timeRangeFilterSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeRangeFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column":                               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                            idSchema(),
				"null_option":                          stringEnumSchema[awstypes.FilterNullOption](attrRequired),
				"default_filter_control_configuration": defaultFilterControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlConfiguration.html
				"exclude_period_configuration":         excludePeriodConfigurationSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExcludePeriodConfiguration.html
				"include_maximum": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"include_minimum": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"range_maximum_value": timeRangeFilterValueSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeRangeFilterValue.html
				"range_minimum_value": timeRangeFilterValueSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeRangeFilterValue.html
				"time_granularity":    stringEnumSchema[awstypes.TimeGranularity](attrOptional),
			},
		},
	}
}

func topBottomFilterSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TopBottomFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"aggregation_sort_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AnchorDateConfiguration.html
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 100,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"aggregation_function": aggregationFunctionSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
							"column":               columnSchema(true),              // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"sort_direction":       stringEnumSchema[awstypes.SortDirection](attrRequired),
						},
					},
				},
				"column":                               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                            idSchema(),
				"default_filter_control_configuration": defaultFilterControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlConfiguration.html
				"limit": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"parameter_name":   parameterNameSchema(false),
				"time_granularity": stringEnumSchema[awstypes.TimeGranularity](attrOptional),
			},
		},
	}
}

var excludePeriodConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExcludePeriodConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"amount": {
					Type:     schema.TypeInt,
					Required: true,
				},
				"granularity":    stringEnumSchema[awstypes.TimeGranularity](attrRequired),
				names.AttrStatus: stringEnumSchema[awstypes.Status](attrOptional),
			},
		},
	}
})

var numericRangeFilterValueSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericRangeFilterValue.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrParameter: {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), ""),
					),
				},
				"static_value": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
			},
		},
	}
})

var timeRangeFilterValueSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeRangeFilterValue.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrParameter: {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), ""),
					),
				},
				"rolling_date": rollingDateConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RollingDateConfiguration.html,
				"static_value": utcTimestampStringSchema(attrOptional),
			},
		},
	}
})

var drillDownFilterSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DrillDownFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 10,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"category_filter": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CategoryDrillDownFilter.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"category_values": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 100000,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(1, 512),
								},
							},
							"column": columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
						},
					},
				},
				"numeric_equality_filter": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericEqualityDrillDownFilter.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column": columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							names.AttrValue: {
								Type:     schema.TypeFloat,
								Required: true,
							},
						},
					},
				},
				"time_range_filter": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeRangeDrillDownFilter.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"column":           columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"range_maximum":    utcTimestampStringSchema(attrRequired),
							"range_minimum":    utcTimestampStringSchema(attrRequired),
							"time_granularity": stringEnumSchema[awstypes.TimeGranularity](attrRequired),
						},
					},
				},
			},
		},
	}
})

var filterSelectableValuesSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSelectableValues.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrValues: {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 50000,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
})

var filterScopeConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterScopeConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"all_sheets": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AllSheetsFilterScopeConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem:     &schema.Resource{},
				},
				"selected_sheets": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SelectedSheetsFilterScopeConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"sheet_visual_scoping_configurations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SheetVisualScopingConfiguration.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 50,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrScope: stringEnumSchema[awstypes.FilterVisualScope](attrRequired),
										"sheet_id":      idSchema(),
										"visual_ids": {
											Type:     schema.TypeSet,
											Optional: true,
											MinItems: 1,
											MaxItems: 50,
											Elem:     &schema.Schema{Type: schema.TypeString},
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
})

func defaultFilterControlConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlConfiguration.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"control_options": defaultFilterControlOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlOptions.html
				"title":           stringLenBetweenSchema(attrRequired, 1, 2048),
			},
		},
	}
}

func defaultFilterControlOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterControlOptions.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"default_date_time_picker_options":   defaultDateTimePickerControlOptionsSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultDateTimePickerControlOptions.html
				"default_dropdown_options":           defaultFilterDropDownControlOptionsSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterDropDownControlOptions.html
				"default_list_options":               defaultFilterListControlOptionsSchema(),       // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterListControlOptions.html
				"default_relative_date_time_options": defaultRelativeDateTimeControlOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultRelativeDateTimeControlOptions.html
				"default_slider_options":             defaultSliderControlOptionsSchema(),           // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultSliderControlOptions.html
				"default_text_area_options":          defaultTextAreaControlOptionsSchema(),         // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultTextAreaControlOptions.html
				"default_text_field_options":         defaultTextFieldControlOptionsSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultTextFieldControlOptions.html
			},
		},
	}
}

func defaultDateTimePickerControlOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultDateTimePickerControlOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"commit_mode":     stringEnumSchema[awstypes.CommitMode](attrOptional),
				"display_options": dateTimePickerControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimePickerControlDisplayOptions.html
				names.AttrType:    stringEnumSchema[awstypes.SheetControlDateTimePickerType](attrOptional),
			},
		},
	}
}

func defaultFilterDropDownControlOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterDropDownControlOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"commit_mode":       stringEnumSchema[awstypes.CommitMode](attrOptional),
				"display_options":   dropDownControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DropDownControlDisplayOptions.html
				"selectable_values": filterSelectableValuesSchema(),        // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSelectableValues.html
				names.AttrType:      stringEnumSchema[awstypes.SheetControlListType](attrOptional),
			},
		},
	}
}

func defaultFilterListControlOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultFilterListControlOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"display_options":   listControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlDisplayOptions.html
				"selectable_values": filterSelectableValuesSchema(),    // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSelectableValues.html
				names.AttrType:      stringEnumSchema[awstypes.SheetControlListType](attrOptional),
			},
		},
	}
}

func defaultRelativeDateTimeControlOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultRelativeDateTimeControlOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"commit_mode":     stringEnumSchema[awstypes.CommitMode](attrOptional),
				"display_options": relativeDateTimeControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RelativeDateTimeControlDisplayOptions.html
			},
		},
	}
}

func defaultSliderControlOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultSliderControlOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"maximum_value": {
					Type:     schema.TypeFloat,
					Required: true,
				},
				"minimum_value": {
					Type:     schema.TypeFloat,
					Required: true,
				},
				"step_size": {
					Type:     schema.TypeFloat,
					Required: true,
				},
				"display_options": sliderControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SliderControlDisplayOptions.html
				names.AttrType:    stringEnumSchema[awstypes.SheetControlSliderType](attrOptional),
			},
		},
	}
}

func defaultTextAreaControlOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultTextAreaControlOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"delimiter":       stringLenBetweenSchema(attrOptional, 1, 2048),
				"display_options": textAreaControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextAreaControlDisplayOptions.html
			},
		},
	}
}

func defaultTextFieldControlOptionsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DefaultTextFieldControlOptions.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"display_options": textFieldControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextFieldControlDisplayOptions.html
			},
		},
	}
}

func expandFilters(tfList []interface{}) []awstypes.Filter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Filter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandFilter(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandFilter(tfMap map[string]interface{}) *awstypes.Filter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Filter{}

	if v, ok := tfMap["category_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.CategoryFilter = expandCategoryFilter(v)
	}
	if v, ok := tfMap["numeric_equality_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.NumericEqualityFilter = expandNumericEqualityFilter(v)
	}
	if v, ok := tfMap["numeric_range_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.NumericRangeFilter = expandNumericRangeFilter(v)
	}
	if v, ok := tfMap["relative_dates_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.RelativeDatesFilter = expandRelativeDatesFilter(v)
	}
	if v, ok := tfMap["time_equality_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.TimeEqualityFilter = expandTimeEqualityFilter(v)
	}
	if v, ok := tfMap["time_range_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.TimeRangeFilter = expandTimeRangeFilter(v)
	}
	if v, ok := tfMap["top_bottom_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.TopBottomFilter = expandTopBottomFilter(v)
	}

	return apiObject
}

func expandCategoryFilter(tfList []interface{}) *awstypes.CategoryFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CategoryFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		apiObject.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap[names.AttrConfiguration].([]interface{}); ok && len(v) > 0 {
		apiObject.Configuration = expandCategoryFilterConfiguration(v)
	}
	if v, ok := tfMap["default_filter_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultFilterControlConfiguration = expandDefaultFilterControlConfiguration(v)
	}

	return apiObject
}

func expandCategoryFilterConfiguration(tfList []interface{}) *awstypes.CategoryFilterConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CategoryFilterConfiguration{}

	if v, ok := tfMap["custom_filter_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.CustomFilterConfiguration = expandCustomFilterConfiguration(v)
	}
	if v, ok := tfMap["custom_filter_list_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.CustomFilterListConfiguration = expandCustomFilterListConfiguration(v)
	}
	if v, ok := tfMap["filter_list_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.FilterListConfiguration = expandFilterListConfiguration(v)
	}

	return apiObject
}

func expandCustomFilterConfiguration(tfList []interface{}) *awstypes.CustomFilterConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomFilterConfiguration{}

	if v, ok := tfMap["category_value"].(string); ok && v != "" {
		apiObject.CategoryValue = aws.String(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		apiObject.MatchOperator = awstypes.CategoryFilterMatchOperator(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		apiObject.NullOption = awstypes.FilterNullOption(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		apiObject.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		apiObject.SelectAllOptions = awstypes.CategoryFilterSelectAllOptions(v)
	}

	return apiObject
}

func expandCustomFilterListConfiguration(tfList []interface{}) *awstypes.CustomFilterListConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomFilterListConfiguration{}

	if v, ok := tfMap["category_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.CategoryValues = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		apiObject.MatchOperator = awstypes.CategoryFilterMatchOperator(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		apiObject.NullOption = awstypes.FilterNullOption(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		apiObject.SelectAllOptions = awstypes.CategoryFilterSelectAllOptions(v)
	}

	return apiObject
}

func expandFilterListConfiguration(tfList []interface{}) *awstypes.FilterListConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterListConfiguration{}

	if v, ok := tfMap["category_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.CategoryValues = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		apiObject.MatchOperator = awstypes.CategoryFilterMatchOperator(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		apiObject.SelectAllOptions = awstypes.CategoryFilterSelectAllOptions(v)
	}

	return apiObject
}

func expandNumericEqualityFilter(tfList []interface{}) *awstypes.NumericEqualityFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericEqualityFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		apiObject.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		apiObject.MatchOperator = awstypes.NumericEqualityMatchOperator(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		apiObject.NullOption = awstypes.FilterNullOption(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		apiObject.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		apiObject.SelectAllOptions = awstypes.NumericFilterSelectAllOptions(v)
	}
	if v, ok := tfMap[names.AttrValue].(float64); ok {
		apiObject.Value = aws.Float64(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		apiObject.AggregationFunction = expandAggregationFunction(v)
	}
	if v, ok := tfMap["default_filter_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultFilterControlConfiguration = expandDefaultFilterControlConfiguration(v)
	}

	return apiObject
}

func expandFilterScopeConfiguration(tfList []interface{}) *awstypes.FilterScopeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterScopeConfiguration{}

	if v, ok := tfMap["all_sheets"].([]interface{}); ok && len(v) > 0 {
		apiObject.AllSheets = &awstypes.AllSheetsFilterScopeConfiguration{}
	}

	if v, ok := tfMap["selected_sheets"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectedSheets = expandSelectedSheetsFilterScopeConfiguration(v)
	}

	return apiObject
}

func expandSelectedSheetsFilterScopeConfiguration(tfList []interface{}) *awstypes.SelectedSheetsFilterScopeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SelectedSheetsFilterScopeConfiguration{}

	if v, ok := tfMap["sheet_visual_scoping_configurations"].([]interface{}); ok && len(v) > 0 {
		apiObject.SheetVisualScopingConfigurations = expandSheetVisualScopingConfigurations(v)
	}

	return apiObject
}

func expandSheetVisualScopingConfigurations(tfList []interface{}) []awstypes.SheetVisualScopingConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SheetVisualScopingConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandSheetVisualScopingConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandSheetVisualScopingConfiguration(tfMap map[string]interface{}) *awstypes.SheetVisualScopingConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SheetVisualScopingConfiguration{}

	if v, ok := tfMap[names.AttrScope].(string); ok && v != "" {
		apiObject.Scope = awstypes.FilterVisualScope(v)
	}
	if v, ok := tfMap["sheet_id"].(string); ok && v != "" {
		apiObject.SheetId = aws.String(v)
	}
	if v, ok := tfMap["visual_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.VisualIds = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandNumericRangeFilter(tfList []interface{}) *awstypes.NumericRangeFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericRangeFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		apiObject.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		apiObject.NullOption = awstypes.FilterNullOption(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		apiObject.SelectAllOptions = awstypes.NumericFilterSelectAllOptions(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		apiObject.AggregationFunction = expandAggregationFunction(v)
	}
	if v, ok := tfMap["include_maximum"].(bool); ok {
		apiObject.IncludeMaximum = aws.Bool(v)
	}
	if v, ok := tfMap["include_minimum"].(bool); ok {
		apiObject.IncludeMinimum = aws.Bool(v)
	}
	if v, ok := tfMap["range_maximum"].([]interface{}); ok && len(v) > 0 {
		apiObject.RangeMaximum = expandNumericRangeFilterValue(v)
	}
	if v, ok := tfMap["range_minimum"].([]interface{}); ok && len(v) > 0 {
		apiObject.RangeMinimum = expandNumericRangeFilterValue(v)
	}
	if v, ok := tfMap["default_filter_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultFilterControlConfiguration = expandDefaultFilterControlConfiguration(v)
	}

	return apiObject
}

func expandNumericRangeFilterValue(tfList []interface{}) *awstypes.NumericRangeFilterValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericRangeFilterValue{}

	if v, ok := tfMap[names.AttrParameter].(string); ok && v != "" {
		apiObject.Parameter = aws.String(v)
	}
	if v, ok := tfMap["static_value"].(float64); ok {
		apiObject.StaticValue = aws.Float64(v)
	}

	return apiObject
}

func expandRelativeDatesFilter(tfList []interface{}) *awstypes.RelativeDatesFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.RelativeDatesFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		apiObject.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		apiObject.NullOption = awstypes.FilterNullOption(v)
	}
	if v, ok := tfMap["relative_date_type"].(string); ok && v != "" {
		apiObject.RelativeDateType = awstypes.RelativeDateType(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		apiObject.TimeGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["minimum_granularity"].(string); ok && v != "" {
		apiObject.MinimumGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		apiObject.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["relative_date_value"].(int); ok {
		apiObject.RelativeDateValue = aws.Int32(int32(v))
	}
	if v, ok := tfMap["anchor_date_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.AnchorDateConfiguration = expandAnchorDateConfiguration(v)
	}
	if v, ok := tfMap["exclude_period_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ExcludePeriodConfiguration = expandExcludePeriodConfiguration(v)
	}
	if v, ok := tfMap["default_filter_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultFilterControlConfiguration = expandDefaultFilterControlConfiguration(v)
	}

	return apiObject
}

func expandAnchorDateConfiguration(tfList []interface{}) *awstypes.AnchorDateConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.AnchorDateConfiguration{}

	if v, ok := tfMap["anchor_option"].(string); ok && v != "" {
		apiObject.AnchorOption = awstypes.AnchorOption(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		apiObject.ParameterName = aws.String(v)
	}

	return apiObject
}

func expandExcludePeriodConfiguration(tfList []interface{}) *awstypes.ExcludePeriodConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ExcludePeriodConfiguration{}

	if v, ok := tfMap["amount"].(int); ok {
		apiObject.Amount = aws.Int32(int32(v))
	}
	if v, ok := tfMap["granularity"].(string); ok && v != "" {
		apiObject.Granularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.WidgetStatus(v)
	}

	return apiObject
}

func expandTimeEqualityFilter(tfList []interface{}) *awstypes.TimeEqualityFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TimeEqualityFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		apiObject.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		apiObject.TimeGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		apiObject.ParameterName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		apiObject.Value = aws.Time(t)
	}
	if v, ok := tfMap["default_filter_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultFilterControlConfiguration = expandDefaultFilterControlConfiguration(v)
	}

	return apiObject
}

func expandTimeRangeFilter(tfList []interface{}) *awstypes.TimeRangeFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TimeRangeFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		apiObject.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		apiObject.NullOption = awstypes.FilterNullOption(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		apiObject.TimeGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["exclude_period_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.ExcludePeriodConfiguration = expandExcludePeriodConfiguration(v)
	}
	if v, ok := tfMap["include_maximum"].(bool); ok {
		apiObject.IncludeMaximum = aws.Bool(v)
	}
	if v, ok := tfMap["include_minimum"].(bool); ok {
		apiObject.IncludeMinimum = aws.Bool(v)
	}
	if v, ok := tfMap["range_maximum_value"].([]interface{}); ok && len(v) > 0 {
		apiObject.RangeMaximumValue = expandTimeRangeFilterValue(v)
	}
	if v, ok := tfMap["range_minimum_value"].([]interface{}); ok && len(v) > 0 {
		apiObject.RangeMinimumValue = expandTimeRangeFilterValue(v)
	}
	if v, ok := tfMap["default_filter_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultFilterControlConfiguration = expandDefaultFilterControlConfiguration(v)
	}

	return apiObject
}

func expandTimeRangeFilterValue(tfList []interface{}) *awstypes.TimeRangeFilterValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TimeRangeFilterValue{}

	if v, ok := tfMap[names.AttrParameter].(string); ok && v != "" {
		apiObject.Parameter = aws.String(v)
	}
	if v, ok := tfMap["static_value"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		apiObject.StaticValue = aws.Time(t)
	}
	if v, ok := tfMap["rolling_date"].([]interface{}); ok && len(v) > 0 {
		apiObject.RollingDate = expandRollingDateConfiguration(v)
	}

	return apiObject
}

func expandTopBottomFilter(tfList []interface{}) *awstypes.TopBottomFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TopBottomFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		apiObject.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["limit"].(int); ok {
		apiObject.Limit = aws.Int32(int32(v))
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		apiObject.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		apiObject.TimeGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["aggregation_sort_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.AggregationSortConfigurations = expandAggregationSortConfigurations(v)
	}
	if v, ok := tfMap["default_filter_control_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultFilterControlConfiguration = expandDefaultFilterControlConfiguration(v)
	}

	return apiObject
}

func expandAggregationSortConfigurations(tfList []interface{}) []awstypes.AggregationSortConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.AggregationSortConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandAggregationSortConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandAggregationSortConfiguration(tfMap map[string]interface{}) *awstypes.AggregationSortConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AggregationSortConfiguration{}

	if v, ok := tfMap["sort_direction"].(string); ok && v != "" {
		apiObject.SortDirection = awstypes.SortDirection(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		apiObject.AggregationFunction = expandAggregationFunction(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}

	return apiObject
}

func expandDefaultFilterControlConfiguration(tfList []interface{}) *awstypes.DefaultFilterControlConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultFilterControlConfiguration{}

	if v, ok := tfMap["control_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.ControlOptions = expandDefaultFilterControlOptions(v)
	}
	if v, ok := tfMap["title"].(string); ok && v != "" {
		apiObject.Title = aws.String(v)
	}

	return apiObject
}

func expandDefaultFilterControlOptions(tfList []interface{}) *awstypes.DefaultFilterControlOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultFilterControlOptions{}

	if v, ok := tfMap["default_date_time_picker_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultDateTimePickerOptions = expandDefaultDateTimePickerControlOptions(v)
	}
	if v, ok := tfMap["default_dropdown_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultDropdownOptions = expandDefaultFilterDropDownControlOptions(v)
	}
	if v, ok := tfMap["default_list_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultListOptions = expandDefaultFilterListControlOptions(v)
	}
	if v, ok := tfMap["default_relative_date_time_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultRelativeDateTimeOptions = expandDefaultRelativeDateTimeControlOptions(v)
	}
	if v, ok := tfMap["default_slider_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultSliderOptions = expandDefaultSliderControlOptions(v)
	}
	if v, ok := tfMap["default_text_area_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultTextAreaOptions = expandDefaultTextAreaControlOptions(v)
	}
	if v, ok := tfMap["default_text_field_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DefaultTextFieldOptions = expandDefaultTextFieldControlOptions(v)
	}

	return apiObject
}

func expandDefaultDateTimePickerControlOptions(tfList []interface{}) *awstypes.DefaultDateTimePickerControlOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultDateTimePickerControlOptions{}

	if v, ok := tfMap["commit_mode"].(string); ok && v != "" {
		apiObject.CommitMode = awstypes.CommitMode(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDateTimePickerControlDisplayOptions(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlDateTimePickerType(v)
	}

	return apiObject
}

func expandDefaultFilterDropDownControlOptions(tfList []interface{}) *awstypes.DefaultFilterDropDownControlOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultFilterDropDownControlOptions{}

	if v, ok := tfMap["commit_mode"].(string); ok && v != "" {
		apiObject.CommitMode = awstypes.CommitMode(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandDropDownControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectableValues = expandFilterSelectableValues(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlListType(v)
	}

	return apiObject
}

func expandDefaultFilterListControlOptions(tfList []interface{}) *awstypes.DefaultFilterListControlOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultFilterListControlOptions{}

	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandListControlDisplayOptions(v)
	}
	if v, ok := tfMap["selectable_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectableValues = expandFilterSelectableValues(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlListType(v)
	}

	return apiObject
}

func expandDefaultRelativeDateTimeControlOptions(tfList []interface{}) *awstypes.DefaultRelativeDateTimeControlOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultRelativeDateTimeControlOptions{}

	if v, ok := tfMap["commit_mode"].(string); ok && v != "" {
		apiObject.CommitMode = awstypes.CommitMode(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandRelativeDateTimeControlDisplayOptions(v)
	}

	return apiObject
}

func expandDefaultSliderControlOptions(tfList []interface{}) *awstypes.DefaultSliderControlOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultSliderControlOptions{}

	if v, ok := tfMap["maximum_value"].(float64); ok {
		apiObject.MaximumValue = v
	}
	if v, ok := tfMap["minimum_value"].(float64); ok {
		apiObject.MinimumValue = v
	}
	if v, ok := tfMap["step_size"].(float64); ok {
		apiObject.StepSize = v
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandSliderControlDisplayOptions(v)
	}
	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.SheetControlSliderType(v)
	}

	return apiObject
}

func expandDefaultTextAreaControlOptions(tfList []interface{}) *awstypes.DefaultTextAreaControlOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultTextAreaControlOptions{}

	if v, ok := tfMap["delimiter"].(string); ok && v != "" {
		apiObject.Delimiter = aws.String(v)
	}
	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextAreaControlDisplayOptions(v)
	}

	return apiObject
}

func expandDefaultTextFieldControlOptions(tfList []interface{}) *awstypes.DefaultTextFieldControlOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DefaultTextFieldControlOptions{}

	if v, ok := tfMap["display_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.DisplayOptions = expandTextFieldControlDisplayOptions(v)
	}

	return apiObject
}

func expandDrillDownFilters(tfList []interface{}) []awstypes.DrillDownFilter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DrillDownFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandDrillDownFilter(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandDrillDownFilter(tfMap map[string]interface{}) *awstypes.DrillDownFilter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.DrillDownFilter{}

	if v, ok := tfMap["category_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.CategoryFilter = expandCategoryDrillDownFilter(v)
	}
	if v, ok := tfMap["numeric_equality_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.NumericEqualityFilter = expandNumericEqualityDrillDownFilter(v)
	}
	if v, ok := tfMap["time_range_filter"].([]interface{}); ok && len(v) > 0 {
		apiObject.TimeRangeFilter = expandTimeRangeDrillDownFilter(v)
	}

	return apiObject
}

func expandCategoryDrillDownFilter(tfList []interface{}) *awstypes.CategoryDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CategoryDrillDownFilter{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["category_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.CategoryValues = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandNumericEqualityDrillDownFilter(tfList []interface{}) *awstypes.NumericEqualityDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericEqualityDrillDownFilter{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap[names.AttrValue].(float64); ok {
		apiObject.Value = v
	}

	return apiObject
}

func expandTimeRangeDrillDownFilter(tfList []interface{}) *awstypes.TimeRangeDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.TimeRangeDrillDownFilter{}

	if v, ok := tfMap["range_maximum"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		apiObject.RangeMaximum = aws.Time(t)
	}
	if v, ok := tfMap["range_minimum"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		apiObject.RangeMinimum = aws.Time(t)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		apiObject.TimeGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		apiObject.Column = expandColumnIdentifier(v)
	}

	return apiObject
}

func flattenFilters(apiObjects []awstypes.Filter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.CategoryFilter != nil {
			tfMap["category_filter"] = flattenCategoryFilter(apiObject.CategoryFilter)
		}
		if apiObject.NumericEqualityFilter != nil {
			tfMap["numeric_equality_filter"] = flattenNumericEqualityFilter(apiObject.NumericEqualityFilter)
		}
		if apiObject.NumericRangeFilter != nil {
			tfMap["numeric_range_filter"] = flattenNumericRangeFilter(apiObject.NumericRangeFilter)
		}
		if apiObject.RelativeDatesFilter != nil {
			tfMap["relative_dates_filter"] = flattenRelativeDatesFilter(apiObject.RelativeDatesFilter)
		}
		if apiObject.TimeEqualityFilter != nil {
			tfMap["time_equality_filter"] = flattenTimeEqualityFilter(apiObject.TimeEqualityFilter)
		}
		if apiObject.TimeRangeFilter != nil {
			tfMap["time_range_filter"] = flattenTimeRangeFilter(apiObject.TimeRangeFilter)
		}
		if apiObject.TopBottomFilter != nil {
			tfMap["top_bottom_filter"] = flattenTopBottomFilter(apiObject.TopBottomFilter)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCategoryFilter(apiObject *awstypes.CategoryFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.Configuration != nil {
		tfMap[names.AttrConfiguration] = flattenCategoryFilterConfiguration(apiObject.Configuration)
	}
	if apiObject.FilterId != nil {
		tfMap["filter_id"] = aws.ToString(apiObject.FilterId)
	}
	if apiObject.DefaultFilterControlConfiguration != nil {
		tfMap["default_filter_control_configuration"] = flattenDefaultFilterControlConfiguration(apiObject.DefaultFilterControlConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenCategoryFilterConfiguration(apiObject *awstypes.CategoryFilterConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CustomFilterConfiguration != nil {
		tfMap["custom_filter_configuration"] = flattenCustomFilterConfiguration(apiObject.CustomFilterConfiguration)
	}
	if apiObject.CustomFilterListConfiguration != nil {
		tfMap["custom_filter_list_configuration"] = flattenCustomFilterListConfiguration(apiObject.CustomFilterListConfiguration)
	}
	if apiObject.FilterListConfiguration != nil {
		tfMap["filter_list_configuration"] = flattenFilterListConfiguration(apiObject.FilterListConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenCustomFilterConfiguration(apiObject *awstypes.CustomFilterConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CategoryValue != nil {
		tfMap["category_value"] = aws.ToString(apiObject.CategoryValue)
	}
	tfMap["match_operator"] = apiObject.MatchOperator
	tfMap["null_option"] = apiObject.NullOption
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}
	tfMap["select_all_options"] = apiObject.SelectAllOptions

	return []interface{}{tfMap}
}

func flattenCustomFilterListConfiguration(apiObject *awstypes.CustomFilterListConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CategoryValues != nil {
		tfMap["category_values"] = apiObject.CategoryValues
	}
	tfMap["match_operator"] = apiObject.MatchOperator
	tfMap["null_option"] = apiObject.NullOption
	tfMap["select_all_options"] = apiObject.SelectAllOptions

	return []interface{}{tfMap}
}

func flattenFilterListConfiguration(apiObject *awstypes.FilterListConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CategoryValues != nil {
		tfMap["category_values"] = apiObject.CategoryValues
	}
	tfMap["match_operator"] = apiObject.MatchOperator
	tfMap["select_all_options"] = apiObject.SelectAllOptions

	return []interface{}{tfMap}
}

func flattenNumericEqualityFilter(apiObject *awstypes.NumericEqualityFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.AggregationFunction != nil {
		tfMap["aggregation_function"] = flattenAggregationFunction(apiObject.AggregationFunction)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FilterId != nil {
		tfMap["filter_id"] = aws.ToString(apiObject.FilterId)
	}
	tfMap["match_operator"] = apiObject.MatchOperator
	tfMap["null_option"] = apiObject.NullOption
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}
	tfMap["select_all_options"] = apiObject.SelectAllOptions
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = aws.ToFloat64(apiObject.Value)
	}
	if apiObject.DefaultFilterControlConfiguration != nil {
		tfMap["default_filter_control_configuration"] = flattenDefaultFilterControlConfiguration(apiObject.DefaultFilterControlConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenNumericRangeFilter(apiObject *awstypes.NumericRangeFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.AggregationFunction != nil {
		tfMap["aggregation_function"] = flattenAggregationFunction(apiObject.AggregationFunction)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FilterId != nil {
		tfMap["filter_id"] = aws.ToString(apiObject.FilterId)
	}
	if apiObject.IncludeMaximum != nil {
		tfMap["include_maximum"] = aws.ToBool(apiObject.IncludeMaximum)
	}
	if apiObject.IncludeMinimum != nil {
		tfMap["include_minimum"] = aws.ToBool(apiObject.IncludeMinimum)
	}
	tfMap["null_option"] = apiObject.NullOption
	if apiObject.RangeMaximum != nil {
		tfMap["range_maximum"] = flattenNumericRangeFilterValue(apiObject.RangeMaximum)
	}
	if apiObject.RangeMinimum != nil {
		tfMap["range_minimum"] = flattenNumericRangeFilterValue(apiObject.RangeMinimum)
	}
	tfMap["select_all_options"] = apiObject.SelectAllOptions
	if apiObject.DefaultFilterControlConfiguration != nil {
		tfMap["default_filter_control_configuration"] = flattenDefaultFilterControlConfiguration(apiObject.DefaultFilterControlConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenNumericRangeFilterValue(apiObject *awstypes.NumericRangeFilterValue) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Parameter != nil {
		tfMap[names.AttrParameter] = aws.ToString(apiObject.Parameter)
	}
	if apiObject.StaticValue != nil {
		tfMap["static_value"] = aws.ToFloat64(apiObject.StaticValue)
	}

	return []interface{}{tfMap}
}

func flattenRelativeDatesFilter(apiObject *awstypes.RelativeDatesFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.AnchorDateConfiguration != nil {
		tfMap["anchor_date_configuration"] = flattenAnchorDateConfiguration(apiObject.AnchorDateConfiguration)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.ExcludePeriodConfiguration != nil {
		tfMap["exclude_period_configuration"] = flattenExcludePeriodConfiguration(apiObject.ExcludePeriodConfiguration)
	}
	if apiObject.FilterId != nil {
		tfMap["filter_id"] = aws.ToString(apiObject.FilterId)
	}
	tfMap["minimum_granularity"] = apiObject.MinimumGranularity
	tfMap["null_option"] = apiObject.NullOption
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}
	tfMap["relative_date_type"] = apiObject.RelativeDateType
	if apiObject.RelativeDateValue != nil {
		tfMap["relative_date_value"] = aws.ToInt32(apiObject.RelativeDateValue)
	}
	tfMap["time_granularity"] = apiObject.TimeGranularity
	if apiObject.DefaultFilterControlConfiguration != nil {
		tfMap["default_filter_control_configuration"] = flattenDefaultFilterControlConfiguration(apiObject.DefaultFilterControlConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenAnchorDateConfiguration(apiObject *awstypes.AnchorDateConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["anchor_option"] = apiObject.AnchorOption
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}

	return []interface{}{tfMap}
}

func flattenExcludePeriodConfiguration(apiObject *awstypes.ExcludePeriodConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Amount != nil {
		tfMap["amount"] = aws.ToInt32(apiObject.Amount)
	}
	tfMap["granularity"] = apiObject.Granularity
	tfMap[names.AttrStatus] = apiObject.Status

	return []interface{}{tfMap}
}

func flattenTimeEqualityFilter(apiObject *awstypes.TimeEqualityFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FilterId != nil {
		tfMap["filter_id"] = aws.ToString(apiObject.FilterId)
	}
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}
	tfMap["time_granularity"] = apiObject.TimeGranularity
	if apiObject.Value != nil {
		tfMap[names.AttrValue] = apiObject.Value.Format(time.RFC3339)
	}
	if apiObject.DefaultFilterControlConfiguration != nil {
		tfMap["default_filter_control_configuration"] = flattenDefaultFilterControlConfiguration(apiObject.DefaultFilterControlConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenTimeRangeFilter(apiObject *awstypes.TimeRangeFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.ExcludePeriodConfiguration != nil {
		tfMap["exclude_period_configuration"] = flattenExcludePeriodConfiguration(apiObject.ExcludePeriodConfiguration)
	}
	if apiObject.FilterId != nil {
		tfMap["filter_id"] = aws.ToString(apiObject.FilterId)
	}
	if apiObject.IncludeMaximum != nil {
		tfMap["include_maximum"] = aws.ToBool(apiObject.IncludeMaximum)
	}
	if apiObject.IncludeMinimum != nil {
		tfMap["include_minimum"] = aws.ToBool(apiObject.IncludeMinimum)
	}
	tfMap["null_option"] = apiObject.NullOption
	if apiObject.RangeMaximumValue != nil {
		tfMap["range_maximum_value"] = flattenTimeRangeFilterValue(apiObject.RangeMaximumValue)
	}
	if apiObject.RangeMinimumValue != nil {
		tfMap["range_minimum_value"] = flattenTimeRangeFilterValue(apiObject.RangeMinimumValue)
	}
	tfMap["time_granularity"] = apiObject.TimeGranularity
	if apiObject.DefaultFilterControlConfiguration != nil {
		tfMap["default_filter_control_configuration"] = flattenDefaultFilterControlConfiguration(apiObject.DefaultFilterControlConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenTimeRangeFilterValue(apiObject *awstypes.TimeRangeFilterValue) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Parameter != nil {
		tfMap[names.AttrParameter] = aws.ToString(apiObject.Parameter)
	}
	if apiObject.RollingDate != nil {
		tfMap["rolling_date"] = flattenRollingDateConfiguration(apiObject.RollingDate)
	}
	if apiObject.StaticValue != nil {
		tfMap["static_value"] = apiObject.StaticValue.Format(time.RFC3339)
	}

	return []interface{}{tfMap}
}

func flattenTopBottomFilter(apiObject *awstypes.TopBottomFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.AggregationSortConfigurations != nil {
		tfMap["aggregation_sort_configuration"] = flattenAggregationSortConfigurations(apiObject.AggregationSortConfigurations)
	}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.FilterId != nil {
		tfMap["filter_id"] = aws.ToString(apiObject.FilterId)
	}
	if apiObject.Limit != nil {
		tfMap["limit"] = aws.ToInt32(apiObject.Limit)
	}
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}
	tfMap["time_granularity"] = apiObject.TimeGranularity
	if apiObject.DefaultFilterControlConfiguration != nil {
		tfMap["default_filter_control_configuration"] = flattenDefaultFilterControlConfiguration(apiObject.DefaultFilterControlConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenAggregationSortConfigurations(apiObjects []awstypes.AggregationSortConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.AggregationFunction != nil {
			tfMap["aggregation_function"] = flattenAggregationFunction(apiObject.AggregationFunction)
		}
		if apiObject.Column != nil {
			tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
		}
		tfMap["sort_direction"] = apiObject.SortDirection

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDefaultFilterControlConfiguration(apiObject *awstypes.DefaultFilterControlConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.ControlOptions != nil {
		tfMap["control_options"] = flattenDefaultFilterControlOptions(apiObject.ControlOptions)
	}
	if apiObject.Title != nil {
		tfMap["title"] = aws.ToString(apiObject.Title)
	}

	return []interface{}{tfMap}
}

func flattenDefaultFilterControlOptions(apiObject *awstypes.DefaultFilterControlOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DefaultDateTimePickerOptions != nil {
		tfMap["default_date_time_picker_options"] = flattenDefaultDateTimePickerControlOptions(apiObject.DefaultDateTimePickerOptions)
	}
	if apiObject.DefaultDropdownOptions != nil {
		tfMap["default_dropdown_options"] = flattenDefaultFilterDropDownControlOptions(apiObject.DefaultDropdownOptions)
	}
	if apiObject.DefaultListOptions != nil {
		tfMap["default_list_options"] = flattenDefaultFilterListControlOptions(apiObject.DefaultListOptions)
	}
	if apiObject.DefaultRelativeDateTimeOptions != nil {
		tfMap["default_relative_date_time_options"] = flattenDefaultRelativeDateTimeControlOptions(apiObject.DefaultRelativeDateTimeOptions)
	}
	if apiObject.DefaultSliderOptions != nil {
		tfMap["default_slider_options"] = flattenDefaultSliderControlOptions(apiObject.DefaultSliderOptions)
	}
	if apiObject.DefaultTextAreaOptions != nil {
		tfMap["default_text_area_options"] = flattenDefaultTextAreaControlOptions(apiObject.DefaultTextAreaOptions)
	}
	if apiObject.DefaultTextFieldOptions != nil {
		tfMap["default_text_field_options"] = flattenDefaultTextFieldControlOptions(apiObject.DefaultTextFieldOptions)
	}

	return []interface{}{tfMap}
}

func flattenDefaultDateTimePickerControlOptions(apiObject *awstypes.DefaultDateTimePickerControlOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["commit_mode"] = apiObject.CommitMode
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDateTimePickerControlDisplayOptions(apiObject.DisplayOptions)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []interface{}{tfMap}
}

func flattenDefaultFilterDropDownControlOptions(apiObject *awstypes.DefaultFilterDropDownControlOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["commit_mode"] = apiObject.CommitMode
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDropDownControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.SelectableValues != nil {
		tfMap["selectable_values"] = flattenFilterSelectableValues(apiObject.SelectableValues)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []interface{}{tfMap}
}

func flattenDefaultFilterListControlOptions(apiObject *awstypes.DefaultFilterListControlOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenListControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.SelectableValues != nil {
		tfMap["selectable_values"] = flattenFilterSelectableValues(apiObject.SelectableValues)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []interface{}{tfMap}
}

func flattenDefaultRelativeDateTimeControlOptions(apiObject *awstypes.DefaultRelativeDateTimeControlOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["commit_mode"] = apiObject.CommitMode
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenRelativeDateTimeControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenDefaultSliderControlOptions(apiObject *awstypes.DefaultSliderControlOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["maximum_value"] = apiObject.MaximumValue
	tfMap["minimum_value"] = apiObject.MinimumValue
	tfMap["step_size"] = apiObject.StepSize
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenSliderControlDisplayOptions(apiObject.DisplayOptions)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []interface{}{tfMap}
}

func flattenDefaultTextAreaControlOptions(apiObject *awstypes.DefaultTextAreaControlOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.Delimiter != nil {
		tfMap["delimiter"] = aws.ToString(apiObject.Delimiter)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextAreaControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenDefaultTextFieldControlOptions(apiObject *awstypes.DefaultTextFieldControlOptions) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextFieldControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenFilterScopeConfiguration(apiObject *awstypes.FilterScopeConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.AllSheets != nil {
		tfMap["all_sheets"] = []interface{}{}
	}

	if apiObject.SelectedSheets != nil {
		tfMap["selected_sheets"] = flattenSelectedSheetsFilterScopeConfiguration(apiObject.SelectedSheets)
	}

	return []interface{}{tfMap}
}

func flattenSelectedSheetsFilterScopeConfiguration(apiObject *awstypes.SelectedSheetsFilterScopeConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SheetVisualScopingConfigurations != nil {
		tfMap["sheet_visual_scoping_configurations"] = flattenSheetVisualScopingConfigurations(apiObject.SheetVisualScopingConfigurations)
	}

	return []interface{}{tfMap}
}

func flattenSheetVisualScopingConfigurations(apiObjects []awstypes.SheetVisualScopingConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		tfMap[names.AttrScope] = apiObject.Scope
		if apiObject.SheetId != nil {
			tfMap["sheet_id"] = aws.ToString(apiObject.SheetId)
		}
		if apiObject.VisualIds != nil {
			tfMap["visual_ids"] = apiObject.VisualIds
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
