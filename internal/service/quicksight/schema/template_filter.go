package schema

import (
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				"column": columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
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
										"match_operator": stringSchema(true, validation.StringInSlice(quicksight.CategoryFilterMatchOperator_Values(), false)),
										"null_option":    stringSchema(true, validation.StringInSlice(quicksight.FilterNullOption_Values(), false)),
										"category_value": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(1, 512),
										},
										"parameter_name": {
											Type:     schema.TypeString,
											Optional: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 2048),
												validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+`), ""),
											),
										},
										"select_all_options": stringSchema(false, validation.StringInSlice(quicksight.SelectAllValueOptions_Values(), false)),
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
										"match_operator": stringSchema(true, validation.StringInSlice(quicksight.CategoryFilterMatchOperator_Values(), false)),
										"null_option":    stringSchema(true, validation.StringInSlice(quicksight.FilterNullOption_Values(), false)),
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
										"select_all_options": stringSchema(false, validation.StringInSlice(quicksight.SelectAllValueOptions_Values(), false)),
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
										"match_operator": stringSchema(true, validation.StringInSlice(quicksight.CategoryFilterMatchOperator_Values(), false)),
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
										"select_all_options": stringSchema(false, validation.StringInSlice(quicksight.SelectAllValueOptions_Values(), false)),
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
				"column":               columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":            idSchema(),
				"match_operator":       stringSchema(true, validation.StringInSlice(quicksight.CategoryFilterMatchOperator_Values(), false)),
				"null_option":          stringSchema(true, validation.StringInSlice(quicksight.FilterNullOption_Values(), false)),
				"aggregation_function": aggregationFunctionSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_AggregationFunction.html
				"parameter_name":       parameterNameSchema(false),
				"select_all_options":   stringSchema(false, validation.StringInSlice(quicksight.SelectAllValueOptions_Values(), false)),
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
				"column":               columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":            idSchema(),
				"null_option":          stringSchema(true, validation.StringInSlice(quicksight.FilterNullOption_Values(), false)),
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
				"select_all_options": stringSchema(false, validation.StringInSlice(quicksight.SelectAllValueOptions_Values(), false)),
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
							"anchor_option":  stringSchema(false, validation.StringInSlice(quicksight.AnchorOption_Values(), false)),
							"parameter_name": parameterNameSchema(false),
						},
					},
				},
				"column":                       columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                    idSchema(),
				"null_option":                  stringSchema(true, validation.StringInSlice(quicksight.FilterNullOption_Values(), false)),
				"relative_date_type":           stringSchema(true, validation.StringInSlice(quicksight.RelativeDateType_Values(), false)),
				"time_granularity":             stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
				"exclude_period_configuration": excludePeriodConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ExcludePeriodConfiguration.html
				"minimum_granularity":          stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
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
				"column":           columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":        idSchema(),
				"time_granularity": stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
				"parameter_name":   parameterNameSchema(false),
				"value": {
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
				"column":                       columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id":                    idSchema(),
				"null_option":                  stringSchema(true, validation.StringInSlice(quicksight.FilterNullOption_Values(), false)),
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
				"time_granularity":    stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
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
							"column":               columnSchema(),                  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"sort_direction":       stringSchema(true, validation.StringInSlice(quicksight.SortDirection_Values(), false)),
						},
					},
				},
				"column":    columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"filter_id": idSchema(),
				"limit": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"parameter_name":   parameterNameSchema(false),
				"time_granularity": stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
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
				"granularity": stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
				"status":      stringSchema(false, validation.StringInSlice(quicksight.Status_Values(), false)),
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
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), ""),
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
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), ""),
					),
				},
				"rolling_date": rollingDateConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RollingDateConfiguration.html,
				"static_value": stringSchema(false, verify.ValidUTCTimestamp),
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
									Type:         schema.TypeString,
									ValidateFunc: validation.StringLenBetween(1, 512),
								},
							},
							"column": columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
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
							"column": columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
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
							"column":           columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
							"range_maximum":    stringSchema(true, verify.ValidUTCTimestamp),
							"range_minimum":    stringSchema(true, verify.ValidUTCTimestamp),
							"time_granularity": stringSchema(true, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
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
										"scope":    stringSchema(true, validation.StringInSlice(quicksight.FilterVisualScope_Values(), false)),
										"sheet_id": idSchema(),
										"visual_ids": {
											Type:     schema.TypeSet,
											Optional: true,
											MinItems: 1,
											MaxItems: 50,
											Elem:     schema.TypeString,
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

func expandFilters(tfList []interface{}) []*quicksight.Filter {
	if len(tfList) == 0 {
		return nil
	}

	var filters []*quicksight.Filter
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		filter := expandFilter(tfMap)
		if filter == nil {
			continue
		}

		filters = append(filters, filter)
	}

	return filters
}

func expandFilter(tfMap map[string]interface{}) *quicksight.Filter {
	if tfMap == nil {
		return nil
	}

	filter := &quicksight.Filter{}

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

func expandCategoryFilter(tfList []interface{}) *quicksight.CategoryFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.CategoryFilter{}

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

func expandCategoryFilterConfiguration(tfList []interface{}) *quicksight.CategoryFilterConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.CategoryFilterConfiguration{}

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

func expandCustomFilterConfiguration(tfList []interface{}) *quicksight.CustomFilterConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.CustomFilterConfiguration{}

	if v, ok := tfMap["category_value"].(string); ok && v != "" {
		config.CategoryValue = aws.String(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		config.MatchOperator = aws.String(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		config.NullOption = aws.String(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		config.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		config.SelectAllOptions = aws.String(v)
	}

	return config
}

func expandCustomFilterListConfiguration(tfList []interface{}) *quicksight.CustomFilterListConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.CustomFilterListConfiguration{}

	if v, ok := tfMap["category_values"].([]interface{}); ok {
		config.CategoryValues = flex.ExpandStringList(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		config.MatchOperator = aws.String(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		config.NullOption = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		config.SelectAllOptions = aws.String(v)
	}

	return config
}

func expandFilterListConfiguration(tfList []interface{}) *quicksight.FilterListConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FilterListConfiguration{}

	if v, ok := tfMap["category_values"].([]interface{}); ok {
		config.CategoryValues = flex.ExpandStringList(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		config.MatchOperator = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		config.SelectAllOptions = aws.String(v)
	}

	return config
}

func expandNumericEqualityFilter(tfList []interface{}) *quicksight.NumericEqualityFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.NumericEqualityFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["match_operator"].(string); ok && v != "" {
		filter.MatchOperator = aws.String(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		filter.NullOption = aws.String(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		filter.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		filter.SelectAllOptions = aws.String(v)
	}
	if v, ok := tfMap["value"].(float64); ok {
		filter.Value = aws.Float64(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		filter.AggregationFunction = expandAggregationFunction(v)
	}

	return filter
}

func expandFilterScopeConfiguration(tfList []interface{}) *quicksight.FilterScopeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FilterScopeConfiguration{}

	if v, ok := tfMap["selected_sheets"].([]interface{}); ok && len(v) > 0 {
		config.SelectedSheets = expandSelectedSheetsFilterScopeConfiguration(v)
	}

	return config
}

func expandSelectedSheetsFilterScopeConfiguration(tfList []interface{}) *quicksight.SelectedSheetsFilterScopeConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SelectedSheetsFilterScopeConfiguration{}

	if v, ok := tfMap["sheet_visual_scoping_configurations"].([]interface{}); ok && len(v) > 0 {
		config.SheetVisualScopingConfigurations = expandSheetVisualScopingConfigurations(v)
	}

	return config
}

func expandSheetVisualScopingConfigurations(tfList []interface{}) []*quicksight.SheetVisualScopingConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.SheetVisualScopingConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandSheetVisualScopingConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, config)
	}

	return configs
}

func expandSheetVisualScopingConfiguration(tfMap map[string]interface{}) *quicksight.SheetVisualScopingConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &quicksight.SheetVisualScopingConfiguration{}

	if v, ok := tfMap["scope"].(string); ok && v != "" {
		config.Scope = aws.String(v)
	}
	if v, ok := tfMap["sheet_id"].(string); ok && v != "" {
		config.SheetId = aws.String(v)
	}
	if v, ok := tfMap["visual_ids"].(*schema.Set); ok {
		config.VisualIds = flex.ExpandStringSet(v)
	}

	return config
}

func expandNumericRangeFilter(tfList []interface{}) *quicksight.NumericRangeFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.NumericRangeFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		filter.NullOption = aws.String(v)
	}
	if v, ok := tfMap["select_all_options"].(string); ok && v != "" {
		filter.SelectAllOptions = aws.String(v)
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

func expandNumericRangeFilterValue(tfList []interface{}) *quicksight.NumericRangeFilterValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.NumericRangeFilterValue{}

	if v, ok := tfMap["parameter"].(string); ok && v != "" {
		filter.Parameter = aws.String(v)
	}
	if v, ok := tfMap["static_value"].(float64); ok {
		filter.StaticValue = aws.Float64(v)
	}

	return filter
}

func expandRelativeDatesFilter(tfList []interface{}) *quicksight.RelativeDatesFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.RelativeDatesFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		filter.NullOption = aws.String(v)
	}
	if v, ok := tfMap["relative_date_type"].(string); ok && v != "" {
		filter.RelativeDateType = aws.String(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = aws.String(v)
	}
	if v, ok := tfMap["minimum_granularity"].(string); ok && v != "" {
		filter.MinimumGranularity = aws.String(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		filter.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["relative_date_value"].(int64); ok {
		filter.RelativeDateValue = aws.Int64(v)
	}
	if v, ok := tfMap["anchor_date_configuration"].([]interface{}); ok && len(v) > 0 {
		filter.AnchorDateConfiguration = expandAnchorDateConfiguration(v)
	}
	if v, ok := tfMap["exclude_period_configuration"].([]interface{}); ok && len(v) > 0 {
		filter.ExcludePeriodConfiguration = expandExcludePeriodConfiguration(v)
	}

	return filter
}

func expandAnchorDateConfiguration(tfList []interface{}) *quicksight.AnchorDateConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.AnchorDateConfiguration{}

	if v, ok := tfMap["anchor_option"].(string); ok && v != "" {
		config.AnchorOption = aws.String(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		config.ParameterName = aws.String(v)
	}

	return config
}

func expandExcludePeriodConfiguration(tfList []interface{}) *quicksight.ExcludePeriodConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ExcludePeriodConfiguration{}

	if v, ok := tfMap["amount"].(int64); ok {
		config.Amount = aws.Int64(v)
	}
	if v, ok := tfMap["granularity"].(string); ok && v != "" {
		config.Granularity = aws.String(v)
	}
	if v, ok := tfMap["status"].(string); ok && v != "" {
		config.Status = aws.String(v)
	}

	return config
}

func expandTimeEqualityFilter(tfList []interface{}) *quicksight.TimeEqualityFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.TimeEqualityFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = aws.String(v)
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

func expandTimeRangeFilter(tfList []interface{}) *quicksight.TimeRangeFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.TimeRangeFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["null_option"].(string); ok && v != "" {
		filter.NullOption = aws.String(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = aws.String(v)
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

func expandTimeRangeFilterValue(tfList []interface{}) *quicksight.TimeRangeFilterValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.TimeRangeFilterValue{}

	if v, ok := tfMap["parameter"].(string); ok && v != "" {
		filter.Parameter = aws.String(v)
	}
	if v, ok := tfMap["static_value"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		filter.StaticValue = aws.Time(t)
	}
	if v, ok := tfMap["range_minimum_value"].([]interface{}); ok && len(v) > 0 {
		filter.RollingDate = expandRollingDateConfiguration(v)
	}

	return filter
}

func expandTopBottomFilter(tfList []interface{}) *quicksight.TopBottomFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.TopBottomFilter{}

	if v, ok := tfMap["filter_id"].(string); ok && v != "" {
		filter.FilterId = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["limit"].(int64); ok {
		filter.Limit = aws.Int64(v)
	}
	if v, ok := tfMap["parameter_name"].(string); ok && v != "" {
		filter.ParameterName = aws.String(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = aws.String(v)
	}
	if v, ok := tfMap["aggregation_sort_configuration"].([]interface{}); ok && len(v) > 0 {
		filter.AggregationSortConfigurations = expandAggregationSortConfigurations(v)
	}

	return filter
}

func expandAggregationSortConfigurations(tfList []interface{}) []*quicksight.AggregationSortConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.AggregationSortConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandAggregationSortConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, config)
	}

	return configs
}

func expandAggregationSortConfiguration(tfMap map[string]interface{}) *quicksight.AggregationSortConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &quicksight.AggregationSortConfiguration{}

	if v, ok := tfMap["sort_direction"].(string); ok && v != "" {
		config.SortDirection = aws.String(v)
	}
	if v, ok := tfMap["aggregation_function"].([]interface{}); ok && len(v) > 0 {
		config.AggregationFunction = expandAggregationFunction(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		config.Column = expandColumnIdentifier(v)
	}

	return config
}

func expandDrillDownFilters(tfList []interface{}) []*quicksight.DrillDownFilter {
	if len(tfList) == 0 {
		return nil
	}

	var options []*quicksight.DrillDownFilter
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		opts := expandDrillDownFilter(tfMap)
		if opts == nil {
			continue
		}

		options = append(options, opts)
	}

	return options
}

func expandDrillDownFilter(tfMap map[string]interface{}) *quicksight.DrillDownFilter {
	if tfMap == nil {
		return nil
	}

	options := &quicksight.DrillDownFilter{}

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

func expandCategoryDrillDownFilter(tfList []interface{}) *quicksight.CategoryDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.CategoryDrillDownFilter{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["category_values"].([]interface{}); ok && len(v) > 0 {
		filter.CategoryValues = flex.ExpandStringList(v)
	}

	return filter
}

func expandNumericEqualityDrillDownFilter(tfList []interface{}) *quicksight.NumericEqualityDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.NumericEqualityDrillDownFilter{}

	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["value"].(float64); ok {
		filter.Value = aws.Float64(v)
	}

	return filter
}

func expandTimeRangeDrillDownFilter(tfList []interface{}) *quicksight.TimeRangeDrillDownFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filter := &quicksight.TimeRangeDrillDownFilter{}

	if v, ok := tfMap["range_maximum"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		filter.RangeMaximum = aws.Time(t)
	}
	if v, ok := tfMap["range_minimum"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		filter.RangeMinimum = aws.Time(t)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		filter.TimeGranularity = aws.String(v)
	}
	if v, ok := tfMap["column"].([]interface{}); ok && len(v) > 0 {
		filter.Column = expandColumnIdentifier(v)
	}

	return filter
}
