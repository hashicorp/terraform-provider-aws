// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func filtersSchema() *schema.Schema {
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
}

func categoryFilterSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CategoryFilter.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"column": columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CategoryFilterConfiguration.html
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
										"match_operator": stringSchema(true, enum.Validate[types.CategoryFilterMatchOperator]()),
										"null_option":    stringSchema(true, enum.Validate[types.FilterNullOption]()),
										"category_value": {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 512)),
										},
										"parameter_name": {
											Type:     schema.TypeString,
											Optional: true,
											ValidateDiagFunc: validation.AllDiag(
												validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
												validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+`), "")),
											),
										},
										"select_all_options": stringSchema(false, enum.Validate[types.CategoryFilterSelectAllOptions]()),
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
										"match_operator": stringSchema(true, enum.Validate[types.CategoryFilterMatchOperator]()),
										"null_option":    stringSchema(true, enum.Validate[types.FilterNullOption]()),
										"category_values": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100000,
											Elem: &schema.Schema{
												Type:             schema.TypeString,
												ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 512)),
											},
										},
										"select_all_options": stringSchema(false, enum.Validate[types.SelectAllValueOptions]()),
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
										"match_operator": stringSchema(true, enum.Validate[types.CategoryFilterMatchOperator]()),
										"category_values": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 1,
											MaxItems: 100000,
											Elem: &schema.Schema{
												Type:             schema.TypeString,
												ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 512)),
											},
										},
										"select_all_options": stringSchema(false, enum.Validate[types.SelectAllValueOptions]()),
									},
								},
							},
						},
					},
				},
				"filter_id": idSchema(),
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
				"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":            idSchema(),
				"match_operator":       stringSchema(true, enum.Validate[types.CategoryFilterMatchOperator]()),
				"null_option":          stringSchema(true, enum.Validate[types.FilterNullOption]()),
				"aggregation_function": aggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
				"parameter_name":       parameterNameSchema(false),
				"select_all_options":   stringSchema(false, enum.Validate[types.CategoryFilterSelectAllOptions]()),
				"value": {
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
				"column":               columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":            idSchema(),
				"null_option":          stringSchema(true, enum.Validate[types.FilterNullOption]()),
				"aggregation_function": aggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
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
				"select_all_options": stringSchema(false, enum.Validate[types.CategoryFilterSelectAllOptions]()),
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
							"anchor_option":  stringSchema(false, enum.Validate[types.AnchorOption]()),
							"parameter_name": parameterNameSchema(false),
						},
					},
				},
				"column":                       columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                    idSchema(),
				"null_option":                  stringSchema(true, enum.Validate[types.FilterNullOption]()),
				"relative_date_type":           stringSchema(true, enum.Validate[types.RelativeDateType]()),
				"time_granularity":             stringSchema(true, enum.Validate[types.TimeGranularity]()),
				"exclude_period_configuration": excludePeriodConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExcludePeriodConfiguration.html
				"minimum_granularity":          stringSchema(true, enum.Validate[types.TimeGranularity]()),
				"parameter_name":               parameterNameSchema(false),
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
				"column":           columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":        idSchema(),
				"time_granularity": stringSchema(true, enum.Validate[types.TimeGranularity]()),
				"parameter_name":   parameterNameSchema(false),
				"value": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validation.ToDiagFunc(verify.ValidUTCTimestamp),
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
				"column":                       columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                    idSchema(),
				"null_option":                  stringSchema(true, enum.Validate[types.FilterNullOption]()),
				"exclude_period_configuration": excludePeriodConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExcludePeriodConfiguration.html
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
				"time_granularity":    stringSchema(true, enum.Validate[types.TimeGranularity]()),
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
							"sort_direction":       stringSchema(true, enum.Validate[types.SortDirection]()),
						},
					},
				},
				"column":    columnSchema(true), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id": idSchema(),
				"limit": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"parameter_name":   parameterNameSchema(false),
				"time_granularity": stringSchema(true, enum.Validate[types.TimeGranularity]()),
			},
		},
	}
}

func excludePeriodConfigurationSchema() *schema.Schema {
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
				"granularity": stringSchema(true, enum.Validate[types.TimeGranularity]()),
				"status":      stringSchema(false, enum.Validate[types.WidgetStatus]()),
			},
		},
	}
}

func numericRangeFilterValueSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericRangeFilterValue.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"parameter": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateDiagFunc: validation.AllDiag(
						validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
						validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "")),
					),
				},
				"static_value": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
			},
		},
	}
}

func timeRangeFilterValueSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TimeRangeFilterValue.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"parameter": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateDiagFunc: validation.AllDiag(
						validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
						validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "")),
					),
				},
				"rolling_date": rollingDateConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RollingDateConfiguration.html,
				"static_value": stringSchema(false, validation.ToDiagFunc(verify.ValidUTCTimestamp)),
			},
		},
	}
}

func drillDownFilterSchema() *schema.Schema {
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
									Type:             schema.TypeString,
									ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 512)),
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
							"value": {
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
							"range_maximum":    stringSchema(true, validation.ToDiagFunc(verify.ValidUTCTimestamp)),
							"range_minimum":    stringSchema(true, validation.ToDiagFunc(verify.ValidUTCTimestamp)),
							"time_granularity": stringSchema(true, enum.Validate[types.TimeGranularity]()),
						},
					},
				},
			},
		},
	}
}

func filterSelectableValuesSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterSelectableValues.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"values": {
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
}

func filterScopeConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterScopeConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
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
										"scope":    stringSchema(true, enum.Validate[types.FilterVisualScope]()),
										"sheet_id": idSchema(),
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
}

func expandFilters(tfList []interface{}) []types.Filter {
	if len(tfList) == 0 {
		return nil
	}

	var filters []types.Filter
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		filter := expandFilter(tfMap)
		if filter == nil {
			continue
		}

		filters = append(filters, *filter)
	}

	return filters
}

func expandFilter(tfMap map[string]interface{}) *types.Filter {
	if tfMap == nil {
		return nil
	}

	filter := &types.Filter{}

	if v, ok := tfMap["category_filter"].([]interface{}); ok && len(v) > 0 {
		filter.CategoryFilter = expandCategoryFilter(v)
	}
	if v, ok := tfMap["numeric_equality_filter"].([]interface{}); ok && len(v) > 0 {
		filter.NumericEqualityFilter = expandNumericEqualityFilter(v)
	}
	if v, ok := tfMap["numeric_range_filter"].([]interface{}); ok && len(v) > 0 {
		filter.NumericRangeFilter = expandNumericRangeFilter(v)
	}
	if v, ok := tfMap["relative_dates_filter"].([]interface{}); ok && len(v) > 0 {
		filter.RelativeDatesFilter = expandRelativeDatesFilter(v)
	}
	if v, ok := tfMap["time_equality_filter"].([]interface{}); ok && len(v) > 0 {
		filter.TimeEqualityFilter = expandTimeEqualityFilter(v)
	}
	if v, ok := tfMap["time_range_filter"].([]interface{}); ok && len(v) > 0 {
		filter.TimeRangeFilter = expandTimeRangeFilter(v)
	}
	if v, ok := tfMap["top_bottom_filter"].([]interface{}); ok && len(v) > 0 {
		filter.TopBottomFilter = expandTopBottomFilter(v)
	}

	return filter
}

func expandCategoryFilter(tfList []interface{}) *types.CategoryFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.CategoryFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["configuration"].([]interface{}); ok && len(v) > 0 {
		filter.Configuration = expandCategoryFilterConfiguration(v)
	}

	return filter
}

func expandCategoryFilterConfiguration(tfList []interface{}) *types.CategoryFilterConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.CategoryFilterConfiguration{}

	if v, ok := tfMap["custom_filter_configuration"].([]interface{}); ok && len(v) > 0 {
		config.CustomFilterConfiguration = expandCustomFilterConfiguration(v)
	}
	if v, ok := tfMap["custom_filter_list_configuration"].([]interface{}); ok && len(v) > 0 {
		config.CustomFilterListConfiguration = expandCustomFilterListConfiguration(v)
	}
	if v, ok := tfMap["filter_list_configuration"].([]interface{}); ok && len(v) > 0 {
		config.FilterListConfiguration = expandFilterListConfiguration(v)
	}

	return config
}

func expandCustomFilterConfiguration(tfList []interface{}) *types.CustomFilterConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.CustomFilterConfiguration{}

	if v, ok := tfMap["category_value"].(string); ok && v != "" {
		config.CategoryValue = aws.String(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		config.MatchOperator = types.CategoryFilterMatchOperator(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		config.NullOption = types.FilterNullOption(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		config.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		config.SelectAllOptions = types.CategoryFilterSelectAllOptions(v)
	}

	return config
}

func expandCustomFilterListConfiguration(tfList []interface{}) *types.CustomFilterListConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.CustomFilterListConfiguration{}

	if v, ok := tfMap["category_values"].([]interface{}); ok && len(v) > 0 {
		config.CategoryValues = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		config.MatchOperator = types.CategoryFilterMatchOperator(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		config.NullOption = types.FilterNullOption(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		config.SelectAllOptions = types.CategoryFilterSelectAllOptions(v)
	}

	return config
}

func expandFilterListConfiguration(tfList []interface{}) *types.FilterListConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilterListConfiguration{}

	if v, ok := tfMap["category_values"].([]interface{}); ok && len(v) > 0 {
		config.CategoryValues = flex.ExpandStringValueList(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		config.MatchOperator = types.CategoryFilterMatchOperator(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		config.SelectAllOptions = types.CategoryFilterSelectAllOptions(v)
	}

	return config
}

func expandNumericEqualityFilter(tfList []interface{}) *types.NumericEqualityFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.NumericEqualityFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		filter.MatchOperator = types.NumericEqualityMatchOperator(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		filter.NullOption = types.FilterNullOption(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		filter.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		filter.SelectAllOptions = types.NumericFilterSelectAllOptions(v)
	}
	if v, ok := tfMap["value"].(float64); ok {
		filter.Value = aws.Float64(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		filter.AggregationFunction = expandAggregationFunction(v)
	}

	return filter
}

func expandFilterScopeConfiguration(tfList []interface{}) *types.FilterScopeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilterScopeConfiguration{}

	if v, ok := tfMap["selected_sheets"].([]interface{}); ok && len(v) > 0 {
		config.SelectedSheets = expandSelectedSheetsFilterScopeConfiguration(v)
	}

	return config
}

func expandSelectedSheetsFilterScopeConfiguration(tfList []interface{}) *types.SelectedSheetsFilterScopeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SelectedSheetsFilterScopeConfiguration{}

	if v, ok := tfMap["sheet_visual_scoping_configurations"].([]interface{}); ok && len(v) > 0 {
		config.SheetVisualScopingConfigurations = expandSheetVisualScopingConfigurations(v)
	}

	return config
}

func expandSheetVisualScopingConfigurations(tfList []interface{}) []types.SheetVisualScopingConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []types.SheetVisualScopingConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandSheetVisualScopingConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, *config)
	}

	return configs
}

func expandSheetVisualScopingConfiguration(tfMap map[string]interface{}) *types.SheetVisualScopingConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &types.SheetVisualScopingConfiguration{}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
		config.Scope = types.FilterVisualScope(v)
	}
	if v, ok := tfMap["sheet_id"].(string); ok && v != "" {
		config.SheetId = aws.String(v)
	}
	if v, ok := tfMap["visual_ids"].(*schema.Set); ok && v.Len() > 0 {
		config.VisualIds = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandNumericRangeFilter(tfList []interface{}) *types.NumericRangeFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.NumericRangeFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		filter.NullOption = types.FilterNullOption(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		filter.SelectAllOptions = types.NumericFilterSelectAllOptions(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		filter.AggregationFunction = expandAggregationFunction(v)
	}
	if v, ok := tfMap["include_maximum"].(bool); ok {
		filter.IncludeMaximum = aws.Bool(v)
	}
	if v, ok := tfMap["include_minimum"].(bool); ok {
		filter.IncludeMinimum = aws.Bool(v)
	}
	if v, ok := tfMap["range_maximum"].([]interface{}); ok && len(v) > 0 {
		filter.RangeMaximum = expandNumericRangeFilterValue(v)
	}
	if v, ok := tfMap["range_minimum"].([]interface{}); ok && len(v) > 0 {
		filter.RangeMinimum = expandNumericRangeFilterValue(v)
	}

	return filter
}

func expandNumericRangeFilterValue(tfList []interface{}) *types.NumericRangeFilterValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.NumericRangeFilterValue{}

	if v, ok := tfMap["parameter"].(string); ok && v != "" {
		filter.Parameter = aws.String(v)
	}
	if v, ok := tfMap["static_value"].(float64); ok {
		filter.StaticValue = aws.Float64(v)
	}

	return filter
}

func expandRelativeDatesFilter(tfList []interface{}) *types.RelativeDatesFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.RelativeDatesFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		filter.NullOption = types.FilterNullOption(v)
	}
	if v, ok := tfMap["relative_date_type"].(string); ok && v != "" {
		filter.RelativeDateType = types.RelativeDateType(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["minimum_granularity"].(string); ok && v != "" {
		filter.MinimumGranularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		filter.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["relative_date_value"].(int); ok {
		filter.RelativeDateValue = aws.Int32(int32(v))
	}
	if v, ok := tfMap["anchor_date_configuration"].([]interface{}); ok && len(v) > 0 {
		filter.AnchorDateConfiguration = expandAnchorDateConfiguration(v)
	}
	if v, ok := tfMap["exclude_period_configuration"].([]interface{}); ok && len(v) > 0 {
		filter.ExcludePeriodConfiguration = expandExcludePeriodConfiguration(v)
	}

	return filter
}

func expandAnchorDateConfiguration(tfList []interface{}) *types.AnchorDateConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.AnchorDateConfiguration{}

	if v, ok := tfMap["anchor_option"].(string); ok && v != "" {
		config.AnchorOption = types.AnchorOption(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		config.ParameterName = aws.String(v)
	}

	return config
}

func expandExcludePeriodConfiguration(tfList []interface{}) *types.ExcludePeriodConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.ExcludePeriodConfiguration{}

	if v, ok := tfMap["amount"].(int); ok {
		config.Amount = aws.Int32(int32(v))
	}
	if v, ok := tfMap["granularity"].(string); ok && v != "" {
		config.Granularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["status"].(string); ok && v != "" {
		config.Status = types.WidgetStatus(v)
	}

	return config
}

func expandTimeEqualityFilter(tfList []interface{}) *types.TimeEqualityFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.TimeEqualityFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		filter.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["value"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		filter.Value = aws.Time(t)
	}

	return filter
}

func expandTimeRangeFilter(tfList []interface{}) *types.TimeRangeFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.TimeRangeFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		filter.NullOption = types.FilterNullOption(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["exclude_period_configuration"].([]interface{}); ok && len(v) > 0 {
		filter.ExcludePeriodConfiguration = expandExcludePeriodConfiguration(v)
	}
	if v, ok := tfMap["include_maximum"].(bool); ok {
		filter.IncludeMaximum = aws.Bool(v)
	}
	if v, ok := tfMap["include_minimum"].(bool); ok {
		filter.IncludeMinimum = aws.Bool(v)
	}
	if v, ok := tfMap["range_maximum_value"].([]interface{}); ok && len(v) > 0 {
		filter.RangeMaximumValue = expandTimeRangeFilterValue(v)
	}
	if v, ok := tfMap["range_minimum_value"].([]interface{}); ok && len(v) > 0 {
		filter.RangeMinimumValue = expandTimeRangeFilterValue(v)
	}

	return filter
}

func expandTimeRangeFilterValue(tfList []interface{}) *types.TimeRangeFilterValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.TimeRangeFilterValue{}

	if v, ok := tfMap["parameter"].(string); ok && v != "" {
		filter.Parameter = aws.String(v)
	}
	if v, ok := tfMap["static_value"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		filter.StaticValue = aws.Time(t)
	}
	if v, ok := tfMap["rolling_date"].([]interface{}); ok && len(v) > 0 {
		filter.RollingDate = expandRollingDateConfiguration(v)
	}

	return filter
}

func expandTopBottomFilter(tfList []interface{}) *types.TopBottomFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.TopBottomFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["limit"].(int); ok {
		filter.Limit = aws.Int32(int32(v))
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		filter.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["aggregation_sort_configuration"].([]interface{}); ok && len(v) > 0 {
		filter.AggregationSortConfigurations = expandAggregationSortConfigurations(v)
	}

	return filter
}

func expandAggregationSortConfigurations(tfList []interface{}) []types.AggregationSortConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []types.AggregationSortConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandAggregationSortConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, *config)
	}

	return configs
}

func expandAggregationSortConfiguration(tfMap map[string]interface{}) *types.AggregationSortConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &types.AggregationSortConfiguration{}

	if v, ok := tfMap["sort_direction"].(string); ok && v != "" {
		config.SortDirection = types.SortDirection(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		config.AggregationFunction = expandAggregationFunction(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		config.Column = expandColumnIdentifier(v)
	}

	return config
}

func expandDrillDownFilters(tfList []interface{}) []types.DrillDownFilter {
	if len(tfList) == 0 {
		return nil
	}

	var options []types.DrillDownFilter
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandDrillDownFilter(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, *opts)
	}

	return options
}

func expandDrillDownFilter(tfMap map[string]interface{}) *types.DrillDownFilter {
	if tfMap == nil {
		return nil
	}

	options := &types.DrillDownFilter{}

	if v, ok := tfMap["category_filter"].([]interface{}); ok && len(v) > 0 {
		options.CategoryFilter = expandCategoryDrillDownFilter(v)
	}
	if v, ok := tfMap["numeric_equality_filter"].([]interface{}); ok && len(v) > 0 {
		options.NumericEqualityFilter = expandNumericEqualityDrillDownFilter(v)
	}
	if v, ok := tfMap["time_range_filter"].([]interface{}); ok && len(v) > 0 {
		options.TimeRangeFilter = expandTimeRangeDrillDownFilter(v)
	}

	return options
}

func expandCategoryDrillDownFilter(tfList []interface{}) *types.CategoryDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.CategoryDrillDownFilter{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["category_values"].([]interface{}); ok && len(v) > 0 {
		filter.CategoryValues = flex.ExpandStringValueList(v)
	}

	return filter
}

func expandNumericEqualityDrillDownFilter(tfList []interface{}) *types.NumericEqualityDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.NumericEqualityDrillDownFilter{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["value"].(float64); ok {
		filter.Value = v
	}

	return filter
}

func expandTimeRangeDrillDownFilter(tfList []interface{}) *types.TimeRangeDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &types.TimeRangeDrillDownFilter{}

	if v, ok := tfMap["range_maximum"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		filter.RangeMaximum = aws.Time(t)
	}
	if v, ok := tfMap["range_minimum"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		filter.RangeMinimum = aws.Time(t)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}

	return filter
}

func flattenFilters(apiObject []types.Filter) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, filter := range apiObject {
		tfMap := map[string]interface{}{}
		if filter.CategoryFilter != nil {
			tfMap["category_filter"] = flattenCategoryFilter(filter.CategoryFilter)
		}
		if filter.NumericEqualityFilter != nil {
			tfMap["numeric_equality_filter"] = flattenNumericEqualityFilter(filter.NumericEqualityFilter)
		}
		if filter.NumericRangeFilter != nil {
			tfMap["numeric_range_filter"] = flattenNumericRangeFilter(filter.NumericRangeFilter)
		}
		if filter.RelativeDatesFilter != nil {
			tfMap["relative_dates_filter"] = flattenRelativeDatesFilter(filter.RelativeDatesFilter)
		}
		if filter.TimeEqualityFilter != nil {
			tfMap["time_equality_filter"] = flattenTimeEqualityFilter(filter.TimeEqualityFilter)
		}
		if filter.TimeRangeFilter != nil {
			tfMap["time_range_filter"] = flattenTimeRangeFilter(filter.TimeRangeFilter)
		}
		if filter.TopBottomFilter != nil {
			tfMap["top_bottom_filter"] = flattenTopBottomFilter(filter.TopBottomFilter)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCategoryFilter(apiObject *types.CategoryFilter) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Column != nil {
		tfMap["column"] = flattenColumnIdentifier(apiObject.Column)
	}
	if apiObject.Configuration != nil {
		tfMap["configuration"] = flattenCategoryFilterConfiguration(apiObject.Configuration)
	}
	if apiObject.FilterId != nil {
		tfMap["filter_id"] = aws.ToString(apiObject.FilterId)
	}

	return []interface{}{tfMap}
}

func flattenCategoryFilterConfiguration(apiObject *types.CategoryFilterConfiguration) []interface{} {
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

func flattenCustomFilterConfiguration(apiObject *types.CustomFilterConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryValue != nil {
		tfMap["category_value"] = aws.ToString(apiObject.CategoryValue)
	}
	tfMap["match_operator"] = types.CategoryFilterMatchOperator(apiObject.MatchOperator)
	tfMap["null_option"] = types.FilterNullOption(apiObject.NullOption)
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}
	tfMap["select_all_options"] = types.CategoryFilterSelectAllOptions(apiObject.SelectAllOptions)

	return []interface{}{tfMap}
}

func flattenCustomFilterListConfiguration(apiObject *types.CustomFilterListConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryValues != nil {
		tfMap["category_values"] = flex.FlattenStringValueList(apiObject.CategoryValues)
	}
	tfMap["match_operator"] = types.CategoryFilterMatchOperator(apiObject.MatchOperator)
	tfMap["null_option"] = types.FilterNullOption(apiObject.NullOption)
	tfMap["select_all_options"] = types.CategoryFilterSelectAllOptions(apiObject.SelectAllOptions)

	return []interface{}{tfMap}
}

func flattenFilterListConfiguration(apiObject *types.FilterListConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CategoryValues != nil {
		tfMap["category_values"] = flex.FlattenStringValueList(apiObject.CategoryValues)
	}
	tfMap["match_operator"] = types.CategoryFilterMatchOperator(apiObject.MatchOperator)
	tfMap["select_all_options"] = types.CategoryFilterSelectAllOptions(apiObject.SelectAllOptions)

	return []interface{}{tfMap}
}

func flattenNumericEqualityFilter(apiObject *types.NumericEqualityFilter) []interface{} {
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
	tfMap["match_operator"] = types.NumericEqualityMatchOperator(apiObject.MatchOperator)
	tfMap["null_option"] = types.FilterNullOption(apiObject.NullOption)
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}
	tfMap["select_all_options"] = types.NumericFilterSelectAllOptions(apiObject.SelectAllOptions)
	if apiObject.Value != nil {
		tfMap["value"] = *apiObject.Value
	} else {
		tfMap["value"] = 0
	}

	return []interface{}{tfMap}
}

func flattenNumericRangeFilter(apiObject *types.NumericRangeFilter) []interface{} {
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
	tfMap["null_option"] = types.FilterNullOption(apiObject.NullOption)
	if apiObject.RangeMaximum != nil {
		tfMap["range_maximum"] = flattenNumericRangeFilterValue(apiObject.RangeMaximum)
	}
	if apiObject.RangeMinimum != nil {
		tfMap["range_minimum"] = flattenNumericRangeFilterValue(apiObject.RangeMinimum)
	}
	tfMap["select_all_options"] = types.NumericFilterSelectAllOptions(apiObject.SelectAllOptions)

	return []interface{}{tfMap}
}

func flattenNumericRangeFilterValue(apiObject *types.NumericRangeFilterValue) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Parameter != nil {
		tfMap["parameter"] = aws.ToString(apiObject.Parameter)
	}
	if apiObject.StaticValue != nil {
		tfMap["static_value"] = aws.ToFloat64(apiObject.StaticValue)
	}

	return []interface{}{tfMap}
}

func flattenRelativeDatesFilter(apiObject *types.RelativeDatesFilter) []interface{} {
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
	tfMap["minimum_granularity"] = types.TimeGranularity(apiObject.MinimumGranularity)
	tfMap["null_option"] = types.FilterNullOption(apiObject.NullOption)
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}
	tfMap["relative_date_type"] = types.RelativeDateType(apiObject.RelativeDateType)
	if apiObject.RelativeDateValue != nil {
		tfMap["relative_date_value"] = aws.ToInt32(apiObject.RelativeDateValue)
	}
	tfMap["time_granularity"] = types.TimeGranularity(apiObject.TimeGranularity)

	return []interface{}{tfMap}
}

func flattenAnchorDateConfiguration(apiObject *types.AnchorDateConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["anchor_option"] = types.AnchorOption(apiObject.AnchorOption)
	if apiObject.ParameterName != nil {
		tfMap["parameter_name"] = aws.ToString(apiObject.ParameterName)
	}

	return []interface{}{tfMap}
}

func flattenExcludePeriodConfiguration(apiObject *types.ExcludePeriodConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Amount != nil {
		tfMap["amount"] = aws.ToInt32(apiObject.Amount)
	}
	tfMap["granularity"] = types.TimeGranularity(apiObject.Granularity)
	tfMap["status"] = types.WidgetStatus(apiObject.Status)

	return []interface{}{tfMap}
}

func flattenTimeEqualityFilter(apiObject *types.TimeEqualityFilter) []interface{} {
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
	tfMap["time_granularity"] = types.TimeGranularity(apiObject.TimeGranularity)
	if apiObject.Value != nil {
		tfMap["value"] = apiObject.Value.Format(time.RFC3339)
	}

	return []interface{}{tfMap}
}

func flattenTimeRangeFilter(apiObject *types.TimeRangeFilter) []interface{} {
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
	tfMap["null_option"] = types.FilterNullOption(apiObject.NullOption)
	if apiObject.RangeMaximumValue != nil {
		tfMap["range_maximum_value"] = flattenTimeRangeFilterValue(apiObject.RangeMaximumValue)
	}
	if apiObject.RangeMinimumValue != nil {
		tfMap["range_minimum_value"] = flattenTimeRangeFilterValue(apiObject.RangeMinimumValue)
	}
	tfMap["time_granularity"] = types.TimeGranularity(apiObject.TimeGranularity)

	return []interface{}{tfMap}
}

func flattenTimeRangeFilterValue(apiObject *types.TimeRangeFilterValue) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.Parameter != nil {
		tfMap["parameter"] = aws.ToString(apiObject.Parameter)
	}
	if apiObject.RollingDate != nil {
		tfMap["rolling_date"] = flattenRollingDateConfiguration(apiObject.RollingDate)
	}
	if apiObject.StaticValue != nil {
		tfMap["static_value"] = apiObject.StaticValue.Format(time.RFC3339)
	}

	return []interface{}{tfMap}
}

func flattenTopBottomFilter(apiObject *types.TopBottomFilter) []interface{} {
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
	tfMap["time_granularity"] = types.TimeGranularity(apiObject.TimeGranularity)

	return []interface{}{tfMap}
}

func flattenAggregationSortConfigurations(apiObject []types.AggregationSortConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		if config.AggregationFunction != nil {
			tfMap["aggregation_function"] = flattenAggregationFunction(config.AggregationFunction)
		}
		if config.Column != nil {
			tfMap["column"] = flattenColumnIdentifier(config.Column)
		}
		tfMap["sort_direction"] = types.SortDirection(config.SortDirection)
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenFilterScopeConfiguration(apiObject *types.FilterScopeConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SelectedSheets != nil {
		tfMap["selected_sheets"] = flattenSelectedSheetsFilterScopeConfiguration(apiObject.SelectedSheets)
	}

	return []interface{}{tfMap}
}

func flattenSelectedSheetsFilterScopeConfiguration(apiObject *types.SelectedSheetsFilterScopeConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SheetVisualScopingConfigurations != nil {
		tfMap["sheet_visual_scoping_configurations"] = flattenSheetVisualScopingConfigurations(apiObject.SheetVisualScopingConfigurations)
	}

	return []interface{}{tfMap}
}

func flattenSheetVisualScopingConfigurations(apiObject []types.SheetVisualScopingConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{}
		tfMap["scope"] = types.FilterVisualScope(config.Scope)
		if config.SheetId != nil {
			tfMap["sheet_id"] = aws.ToString(config.SheetId)
		}
		if config.VisualIds != nil {
			tfMap["visual_ids"] = flex.FlattenStringValueSet(config.VisualIds)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}
