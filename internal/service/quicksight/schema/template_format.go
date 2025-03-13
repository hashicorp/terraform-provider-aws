// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var numericFormatConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"currency_display_format_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CurrencyDisplayFormatConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"decimal_places_configuration":    decimalPlacesConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalPlacesConfiguration.html
							"negative_value_configuration":    negativeValueConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NegativeValueConfiguration.html
							"null_value_format_configuration": nullValueConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NullValueFormatConfiguration.html
							"number_scale":                    stringEnumSchema[awstypes.NumberScale](attrOptional),
							names.AttrPrefix:                  stringLenBetweenSchema(attrOptional, 1, 128),
							"separator_configuration":         separatorConfigurationSchema(),
							"suffix":                          stringLenBetweenSchema(attrOptional, 1, 128),
							"symbol":                          stringMatchSchema(attrOptional, `[A-Z]{3}`, "must be a 3 character currency symbol"),
						},
					},
				},
				"number_display_format_configuration":     numberDisplayFormatConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberDisplayFormatConfiguration.html
				"percentage_display_format_configuration": percentageDisplayFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PercentageDisplayFormatConfiguration.html
			},
		},
	}
})

var dateTimeFormatConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_format":                stringLenBetweenSchema(attrOptional, 1, 128),
				"null_value_format_configuration": nullValueConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NullValueFormatConfiguration.html
				"numeric_format_configuration":    numericFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericFormatConfiguration.html
			},
		},
	}
})

var numberDisplayFormatConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberDisplayFormatConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"decimal_places_configuration":    decimalPlacesConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalPlacesConfiguration.html
				"negative_value_configuration":    negativeValueConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NegativeValueConfiguration.html
				"null_value_format_configuration": nullValueConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NullValueFormatConfiguration.html
				"number_scale":                    stringEnumSchema[awstypes.NumberScale](attrOptional),
				names.AttrPrefix:                  stringLenBetweenSchema(attrOptional, 1, 128),
				"separator_configuration":         separatorConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericSeparatorConfiguration.html
				"suffix":                          stringLenBetweenSchema(attrOptional, 1, 128),
			},
		},
	}
})

var percentageDisplayFormatConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PercentageDisplayFormatConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"decimal_places_configuration":    decimalPlacesConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalPlacesConfiguration.html
				"negative_value_configuration":    negativeValueConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NegativeValueConfiguration.html
				"null_value_format_configuration": nullValueConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NullValueFormatConfiguration.html
				names.AttrPrefix:                  stringLenBetweenSchema(attrOptional, 1, 128),
				"separator_configuration":         separatorConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericSeparatorConfiguration.html
				"suffix":                          stringLenBetweenSchema(attrOptional, 1, 128),
			},
		},
	}
})

var numberFormatConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberFormatConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"numeric_format_configuration": numericFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericFormatConfiguration.html
			},
		},
	}
})

var stringFormatConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"null_value_format_configuration": nullValueConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NullValueFormatConfiguration.html
				"numeric_format_configuration":    numericFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericFormatConfiguration.html
			},
		},
	}
})

var decimalPlacesConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalPlacesConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"decimal_places": intBetweenSchema(attrRequired, 0, 20),
			},
		},
	}
})

var negativeValueConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"display_mode": stringEnumSchema[awstypes.NegativeValueDisplayMode](attrRequired),
			},
		},
	}
})

var nullValueConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"null_string": stringLenBetweenSchema(attrRequired, 1, 128),
			},
		},
	}
})

var separatorConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"decimal_separator": stringEnumSchema[awstypes.NumericSeparatorSymbol](attrOptional),
				"thousands_separator": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ThousandSeparatorOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"symbol":     stringEnumSchema[awstypes.NumericSeparatorSymbol](attrOptional),
							"visibility": stringEnumSchema[awstypes.Visibility](attrOptional),
						},
					},
				},
			},
		},
	}
})

var labelOptionsSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LabelOptions.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"custom_label": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"font_configuration": fontConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
				"visibility":         stringEnumSchema[awstypes.Visibility](attrOptional),
			},
		},
	}
})

var fontConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"font_color":      hexColorSchema(attrOptional),
				"font_decoration": stringEnumSchema[awstypes.FontDecoration](attrOptional),
				"font_size": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontSize.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"relative": stringEnumSchema[awstypes.RelativeFontSize](attrOptional),
						},
					},
				},
				"font_style": stringEnumSchema[awstypes.FontStyle](attrOptional),
				"font_weight": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontWeight.html
					Type:     schema.TypeList,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrName: stringEnumSchema[awstypes.FontWeightName](attrOptional),
						},
					},
				},
			},
		},
	}
})

var formatConfigurationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FormatConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_format_configuration": dateTimeFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeFormatConfiguration.html
				"number_format_configuration":    numberFormatConfigurationSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberFormatConfiguration.html
				"string_format_configuration":    stringFormatConfigurationSchema(),   // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringFormatConfiguration.html
			},
		},
	}
})

func expandFormatConfiguration(tfList []any) *awstypes.FormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FormatConfiguration{}

	if v, ok := tfMap["date_time_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.DateTimeFormatConfiguration = expandDateTimeFormatConfiguration(v)
	}
	if v, ok := tfMap["number_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NumberFormatConfiguration = expandNumberFormatConfiguration(v)
	}
	if v, ok := tfMap["string_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.StringFormatConfiguration = expandStringFormatConfiguration(v)
	}

	return apiObject
}

func expandDateTimeFormatConfiguration(tfList []any) *awstypes.DateTimeFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateTimeFormatConfiguration{}

	if v, ok := tfMap["date_time_format"].(string); ok && v != "" {
		apiObject.DateTimeFormat = aws.String(v)
	}
	if v, ok := tfMap["null_value_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["numeric_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NumericFormatConfiguration = expandNumericFormatConfiguration(v)
	}

	return apiObject
}

func expandNullValueFormatConfiguration(tfList []any) *awstypes.NullValueFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NullValueFormatConfiguration{}

	if v, ok := tfMap["null_string"].(string); ok && v != "" {
		apiObject.NullString = aws.String(v)
	}

	return apiObject
}

func expandNumericFormatConfiguration(tfList []any) *awstypes.NumericFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericFormatConfiguration{}

	if v, ok := tfMap["currency_display_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.CurrencyDisplayFormatConfiguration = expandCurrencyDisplayFormatConfiguration(v)
	}
	if v, ok := tfMap["number_display_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NumberDisplayFormatConfiguration = expandNumberDisplayFormatConfiguration(v)
	}
	if v, ok := tfMap["percentage_display_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.PercentageDisplayFormatConfiguration = expandPercentageDisplayFormatConfiguration(v)
	}

	return apiObject
}

func expandCurrencyDisplayFormatConfiguration(tfList []any) *awstypes.CurrencyDisplayFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.CurrencyDisplayFormatConfiguration{}

	if v, ok := tfMap["decimal_places_configuration"].([]any); ok && len(v) > 0 {
		apiObject.DecimalPlacesConfiguration = expandDecimalPlacesConfiguration(v)
	}
	if v, ok := tfMap["negative_value_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NegativeValueConfiguration = expandNegativeValueConfiguration(v)
	}
	if v, ok := tfMap["null_value_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["number_scale"].(string); ok && v != "" {
		apiObject.NumberScale = awstypes.NumberScale(v)
	}
	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}
	if v, ok := tfMap["separator_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SeparatorConfiguration = expandNumericSeparatorConfiguration(v)
	}
	if v, ok := tfMap["suffix"].(string); ok && v != "" {
		apiObject.Suffix = aws.String(v)
	}
	if v, ok := tfMap["symbol"].(string); ok && v != "" {
		apiObject.Symbol = aws.String(v)
	}

	return apiObject
}

func expandDecimalPlacesConfiguration(tfList []any) *awstypes.DecimalPlacesConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DecimalPlacesConfiguration{}

	if v, ok := tfMap["decimal_places"].(int); ok {
		apiObject.DecimalPlaces = aws.Int64(int64(v))
	}

	return apiObject
}

func expandNegativeValueConfiguration(tfList []any) *awstypes.NegativeValueConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NegativeValueConfiguration{}

	if v, ok := tfMap["display_mode"].(string); ok {
		apiObject.DisplayMode = awstypes.NegativeValueDisplayMode(v)
	}

	return apiObject
}

func expandNumericSeparatorConfiguration(tfList []any) *awstypes.NumericSeparatorConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumericSeparatorConfiguration{}

	if v, ok := tfMap["decimal_separator"].(string); ok {
		apiObject.DecimalSeparator = awstypes.NumericSeparatorSymbol(v)
	}
	if v, ok := tfMap["thousands_separator"].([]any); ok && len(v) > 0 {
		apiObject.ThousandsSeparator = expandThousandSeparatorOptions(v)
	}

	return apiObject
}

func expandThousandSeparatorOptions(tfList []any) *awstypes.ThousandSeparatorOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ThousandSeparatorOptions{}

	if v, ok := tfMap["symbol"].(string); ok {
		apiObject.Symbol = awstypes.NumericSeparatorSymbol(v)
	}
	if v, ok := tfMap["visibility"].(string); ok {
		apiObject.Visibility = awstypes.Visibility(v)
	}

	return apiObject
}

func expandNumberFormatConfiguration(tfList []any) *awstypes.NumberFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumberFormatConfiguration{}

	if v, ok := tfMap["numeric_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FormatConfiguration = expandNumericFormatConfiguration(v)
	}

	return apiObject
}

func expandStringFormatConfiguration(tfList []any) *awstypes.StringFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.StringFormatConfiguration{}

	if v, ok := tfMap["null_value_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["numeric_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NumericFormatConfiguration = expandNumericFormatConfiguration(v)
	}

	return apiObject
}

func expandLabelOptions(tfList []any) *awstypes.LabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.LabelOptions{}

	if v, ok := tfMap["custom_label"].(string); ok {
		apiObject.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok {
		apiObject.Visibility = awstypes.Visibility(v)
	}
	if v, ok := tfMap["font_configuration"].([]any); ok && len(v) > 0 {
		apiObject.FontConfiguration = expandFontConfiguration(v)
	}

	return apiObject
}

func expandFontConfiguration(tfList []any) *awstypes.FontConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FontConfiguration{}

	if v, ok := tfMap["font_color"].(string); ok && v != "" {
		apiObject.FontColor = aws.String(v)
	}
	if v, ok := tfMap["font_decoration"].(string); ok && v != "" {
		apiObject.FontDecoration = awstypes.FontDecoration(v)
	}
	if v, ok := tfMap["font_style"].(string); ok && v != "" {
		apiObject.FontStyle = awstypes.FontStyle(v)
	}
	if v, ok := tfMap["font_size"].([]any); ok && len(v) > 0 {
		apiObject.FontSize = expandFontSize(v)
	}
	if v, ok := tfMap["font_weight"].([]any); ok && len(v) > 0 {
		apiObject.FontWeight = expandFontWeight(v)
	}

	return apiObject
}

func expandFontSize(tfList []any) *awstypes.FontSize {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FontSize{}

	if v, ok := tfMap["relative"].(string); ok {
		apiObject.Relative = awstypes.RelativeFontSize(v)
	}

	return apiObject
}

func expandFontWeight(tfList []any) *awstypes.FontWeight {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.FontWeight{}

	if v, ok := tfMap[names.AttrName].(string); ok {
		apiObject.Name = awstypes.FontWeightName(v)
	}

	return apiObject
}

func expandComparisonFormatConfiguration(tfList []any) *awstypes.ComparisonFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ComparisonFormatConfiguration{}

	if v, ok := tfMap["number_display_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NumberDisplayFormatConfiguration = expandNumberDisplayFormatConfiguration(v)
	}
	if v, ok := tfMap["percentage_display_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.PercentageDisplayFormatConfiguration = expandPercentageDisplayFormatConfiguration(v)
	}

	return apiObject
}

func expandNumberDisplayFormatConfiguration(tfList []any) *awstypes.NumberDisplayFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.NumberDisplayFormatConfiguration{}

	if v, ok := tfMap["number_scale"].(string); ok && v != "" {
		apiObject.NumberScale = awstypes.NumberScale(v)
	}
	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}
	if v, ok := tfMap["suffix"].(string); ok && v != "" {
		apiObject.Suffix = aws.String(v)
	}
	if v, ok := tfMap["decimal_places_configuration"].([]any); ok && len(v) > 0 {
		apiObject.DecimalPlacesConfiguration = expandDecimalPlacesConfiguration(v)
	}
	if v, ok := tfMap["negative_value_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NegativeValueConfiguration = expandNegativeValueConfiguration(v)
	}
	if v, ok := tfMap["null_value_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["separator_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SeparatorConfiguration = expandNumericSeparatorConfiguration(v)
	}

	return apiObject
}

func expandPercentageDisplayFormatConfiguration(tfList []any) *awstypes.PercentageDisplayFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.PercentageDisplayFormatConfiguration{}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		apiObject.Prefix = aws.String(v)
	}
	if v, ok := tfMap["suffix"].(string); ok && v != "" {
		apiObject.Suffix = aws.String(v)
	}
	if v, ok := tfMap["decimal_places_configuration"].([]any); ok && len(v) > 0 {
		apiObject.DecimalPlacesConfiguration = expandDecimalPlacesConfiguration(v)
	}
	if v, ok := tfMap["negative_value_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NegativeValueConfiguration = expandNegativeValueConfiguration(v)
	}
	if v, ok := tfMap["null_value_format_configuration"].([]any); ok && len(v) > 0 {
		apiObject.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["separator_configuration"].([]any); ok && len(v) > 0 {
		apiObject.SeparatorConfiguration = expandNumericSeparatorConfiguration(v)
	}

	return apiObject
}

func flattenFormatConfiguration(apiObject *awstypes.FormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DateTimeFormatConfiguration != nil {
		tfMap["date_time_format_configuration"] = flattenDateTimeFormatConfiguration(apiObject.DateTimeFormatConfiguration)
	}
	if apiObject.NumberFormatConfiguration != nil {
		tfMap["number_format_configuration"] = flattenNumberFormatConfiguration(apiObject.NumberFormatConfiguration)
	}
	if apiObject.StringFormatConfiguration != nil {
		tfMap["string_format_configuration"] = flattenStringFormatConfiguration(apiObject.StringFormatConfiguration)
	}

	return []any{tfMap}
}

func flattenDateTimeFormatConfiguration(apiObject *awstypes.DateTimeFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DateTimeFormat != nil {
		tfMap["date_time_format"] = aws.ToString(apiObject.DateTimeFormat)
	}
	if apiObject.NullValueFormatConfiguration != nil {
		tfMap["null_value_format_configuration"] = flattenNullValueFormatConfiguration(apiObject.NullValueFormatConfiguration)
	}
	if apiObject.NumericFormatConfiguration != nil {
		tfMap["numeric_format_configuration"] = flattenNumericFormatConfiguration(apiObject.NumericFormatConfiguration)
	}

	return []any{tfMap}
}

func flattenNullValueFormatConfiguration(apiObject *awstypes.NullValueFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.NullString != nil {
		tfMap["null_string"] = aws.ToString(apiObject.NullString)
	}

	return []any{tfMap}
}

func flattenNumericFormatConfiguration(apiObject *awstypes.NumericFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CurrencyDisplayFormatConfiguration != nil {
		tfMap["currency_display_format_configuration"] = flattenCurrencyDisplayFormatConfiguration(apiObject.CurrencyDisplayFormatConfiguration)
	}
	if apiObject.NumberDisplayFormatConfiguration != nil {
		tfMap["number_display_format_configuration"] = flattenNumberDisplayFormatConfiguration(apiObject.NumberDisplayFormatConfiguration)
	}
	if apiObject.PercentageDisplayFormatConfiguration != nil {
		tfMap["percentage_display_format_configuration"] = flattenPercentageDisplayFormatConfiguration(apiObject.PercentageDisplayFormatConfiguration)
	}

	return []any{tfMap}
}

func flattenCurrencyDisplayFormatConfiguration(apiObject *awstypes.CurrencyDisplayFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DecimalPlacesConfiguration != nil {
		tfMap["decimal_places_configuration"] = flattenDecimalPlacesConfiguration(apiObject.DecimalPlacesConfiguration)
	}
	if apiObject.NegativeValueConfiguration != nil {
		tfMap["negative_value_configuration"] = flattenNegativeValueConfiguration(apiObject.NegativeValueConfiguration)
	}
	if apiObject.NullValueFormatConfiguration != nil {
		tfMap["null_value_format_configuration"] = flattenNullValueFormatConfiguration(apiObject.NullValueFormatConfiguration)
	}
	tfMap["number_scale"] = apiObject.NumberScale
	if apiObject.Prefix != nil {
		tfMap[names.AttrPrefix] = aws.ToString(apiObject.Prefix)
	}
	if apiObject.SeparatorConfiguration != nil {
		tfMap["separator_configuration"] = flattenNumericSeparatorConfiguration(apiObject.SeparatorConfiguration)
	}
	if apiObject.Suffix != nil {
		tfMap["suffix"] = aws.ToString(apiObject.Suffix)
	}
	if apiObject.Symbol != nil {
		tfMap["symbol"] = aws.ToString(apiObject.Symbol)
	}

	return []any{tfMap}
}

func flattenDecimalPlacesConfiguration(apiObject *awstypes.DecimalPlacesConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DecimalPlaces != nil {
		tfMap["decimal_places"] = aws.ToInt64(apiObject.DecimalPlaces)
	}

	return []any{tfMap}
}

func flattenNegativeValueConfiguration(apiObject *awstypes.NegativeValueConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"display_mode": apiObject.DisplayMode,
	}

	return []any{tfMap}
}

func flattenNumericSeparatorConfiguration(apiObject *awstypes.NumericSeparatorConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"decimal_separator": apiObject.DecimalSeparator,
	}

	if apiObject.ThousandsSeparator != nil {
		tfMap["thousands_separator"] = flattenThousandSeparatorOptions(apiObject.ThousandsSeparator)
	}

	return []any{tfMap}
}
func flattenThousandSeparatorOptions(apiObject *awstypes.ThousandSeparatorOptions) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"symbol":     apiObject.Symbol,
		"visibility": apiObject.Visibility,
	}

	return []any{tfMap}
}

func flattenNumberDisplayFormatConfiguration(apiObject *awstypes.NumberDisplayFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DecimalPlacesConfiguration != nil {
		tfMap["decimal_places_configuration"] = flattenDecimalPlacesConfiguration(apiObject.DecimalPlacesConfiguration)
	}
	if apiObject.NegativeValueConfiguration != nil {
		tfMap["negative_value_configuration"] = flattenNegativeValueConfiguration(apiObject.NegativeValueConfiguration)
	}
	if apiObject.NullValueFormatConfiguration != nil {
		tfMap["null_value_format_configuration"] = flattenNullValueFormatConfiguration(apiObject.NullValueFormatConfiguration)
	}
	tfMap["number_scale"] = apiObject.NumberScale
	if apiObject.Prefix != nil {
		tfMap[names.AttrPrefix] = aws.ToString(apiObject.Prefix)
	}
	if apiObject.SeparatorConfiguration != nil {
		tfMap["separator_configuration"] = flattenNumericSeparatorConfiguration(apiObject.SeparatorConfiguration)
	}
	if apiObject.Suffix != nil {
		tfMap["suffix"] = aws.ToString(apiObject.Suffix)
	}

	return []any{tfMap}
}

func flattenPercentageDisplayFormatConfiguration(apiObject *awstypes.PercentageDisplayFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DecimalPlacesConfiguration != nil {
		tfMap["decimal_places_configuration"] = flattenDecimalPlacesConfiguration(apiObject.DecimalPlacesConfiguration)
	}
	if apiObject.NegativeValueConfiguration != nil {
		tfMap["negative_value_configuration"] = flattenNegativeValueConfiguration(apiObject.NegativeValueConfiguration)
	}
	if apiObject.NullValueFormatConfiguration != nil {
		tfMap["null_value_format_configuration"] = flattenNullValueFormatConfiguration(apiObject.NullValueFormatConfiguration)
	}
	if apiObject.Prefix != nil {
		tfMap[names.AttrPrefix] = aws.ToString(apiObject.Prefix)
	}
	if apiObject.SeparatorConfiguration != nil {
		tfMap["separator_configuration"] = flattenNumericSeparatorConfiguration(apiObject.SeparatorConfiguration)
	}
	if apiObject.Suffix != nil {
		tfMap["suffix"] = aws.ToString(apiObject.Suffix)
	}

	return []any{tfMap}
}

func flattenNumberFormatConfiguration(apiObject *awstypes.NumberFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.FormatConfiguration != nil {
		tfMap["numeric_format_configuration"] = flattenNumericFormatConfiguration(apiObject.FormatConfiguration)
	}

	return []any{tfMap}
}

func flattenStringFormatConfiguration(apiObject *awstypes.StringFormatConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.NullValueFormatConfiguration != nil {
		tfMap["null_value_format_configuration"] = flattenNullValueFormatConfiguration(apiObject.NullValueFormatConfiguration)
	}
	if apiObject.NumericFormatConfiguration != nil {
		tfMap["numeric_format_configuration"] = flattenNumericFormatConfiguration(apiObject.NumericFormatConfiguration)
	}

	return []any{tfMap}
}
