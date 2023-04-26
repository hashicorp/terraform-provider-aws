package schema

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func numericFormatConfigurationSchema() *schema.Schema {
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
							"number_scale":                    stringSchema(false, validation.StringInSlice(quicksight.NumberScale_Values(), false)),
							"prefix":                          stringSchema(false, validation.StringLenBetween(1, 128)),
							"separator_configuration":         separatorConfigurationSchema(),
							"suffix":                          stringSchema(false, validation.StringLenBetween(1, 128)),
							"symbol":                          stringSchema(false, validation.StringMatch(regexp.MustCompile(`[A-Z]{3}`), "must be a 3 character currency symbol")),
						},
					},
				},
				"number_display_format_configuration":     numberDisplayFormatConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumberDisplayFormatConfiguration.html
				"percentage_display_format_configuration": percentageDisplayFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_PercentageDisplayFormatConfiguration.html
			},
		},
	}
}

func dateTimeFormatConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_format":                stringSchema(false, validation.StringLenBetween(1, 128)),
				"null_value_format_configuration": nullValueConfigurationSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NullValueFormatConfiguration.html
				"numeric_format_configuration":    numericFormatConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericFormatConfiguration.html
			},
		},
	}
}

func numberDisplayFormatConfigurationSchema() *schema.Schema {
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
				"number_scale":                    stringSchema(false, validation.StringInSlice(quicksight.NumberScale_Values(), false)),
				"prefix":                          stringSchema(false, validation.StringLenBetween(1, 128)),
				"separator_configuration":         separatorConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericSeparatorConfiguration.html
				"suffix":                          stringSchema(false, validation.StringLenBetween(1, 128)),
			},
		},
	}
}

func percentageDisplayFormatConfigurationSchema() *schema.Schema {
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
				"prefix":                          stringSchema(false, validation.StringLenBetween(1, 128)),
				"separator_configuration":         separatorConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_NumericSeparatorConfiguration.html
				"suffix":                          stringSchema(false, validation.StringLenBetween(1, 128)),
			},
		},
	}
}

func numberFormatConfigurationSchema() *schema.Schema {
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
}

func stringFormatConfigurationSchema() *schema.Schema {
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
}

func decimalPlacesConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalPlacesConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"decimal_places": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntBetween(0, 20),
				},
			},
		},
	}
}

func negativeValueConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"display_mode": stringSchema(true, validation.StringInSlice(quicksight.NegativeValueDisplayMode_Values(), false)),
			},
		},
	}
}

func nullValueConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"null_string": stringSchema(true, validation.StringLenBetween(1, 128)),
			},
		},
	}
}

func separatorConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"decimal_separator": stringSchema(false, validation.StringInSlice(quicksight.NumericSeparatorSymbol_Values(), false)),
				"thousands_separator": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ThousandSeparatorOptions.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"symbol":     stringSchema(false, validation.StringInSlice(quicksight.NumericSeparatorSymbol_Values(), false)),
							"visibility": stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
						},
					},
				},
			},
		},
	}
}

func labelOptionsSchema() *schema.Schema {
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
				"visibility":         stringSchema(false, validation.StringInSlice(quicksight.Visibility_Values(), false)),
			},
		},
	}
}

func fontConfigurationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontConfiguration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"font_color":      stringSchema(false, validation.StringMatch(regexp.MustCompile(`^#[A-F0-9]{6}$`), "")),
				"font_decoration": stringSchema(false, validation.StringInSlice(quicksight.FontDecoration_Values(), false)),
				"font_size": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontSize.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"relative": stringSchema(false, validation.StringInSlice(quicksight.RelativeFontSize_Values(), false)),
						},
					},
				},
				"font_style": stringSchema(false, validation.StringInSlice(quicksight.FontStyle_Values(), false)),
				"font_weight": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FontWeight.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": stringSchema(false, validation.StringInSlice(quicksight.FontWeightName_Values(), false)),
						},
					},
				},
			},
		},
	}
}

func formatConfigurationSchema() *schema.Schema {
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
}

func expandFormatConfiguration(tfList []interface{}) *quicksight.FormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FormatConfiguration{}

	if v, ok := tfMap["date_time_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.DateTimeFormatConfiguration = expandDateTimeFormatConfiguration(v)
	}
	if v, ok := tfMap["number_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NumberFormatConfiguration = expandNumberFormatConfiguration(v)
	}
	if v, ok := tfMap["string_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.StringFormatConfiguration = expandStringFormatConfiguration(v)
	}

	return config
}

func expandDateTimeFormatConfiguration(tfList []interface{}) *quicksight.DateTimeFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DateTimeFormatConfiguration{}

	if v, ok := tfMap["date_time_format"].(string); ok && v != "" {
		config.DateTimeFormat = aws.String(v)
	}
	if v, ok := tfMap["null_value_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["numeric_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NumericFormatConfiguration = expandNumericFormatConfiguration(v)
	}

	return config
}

func expandNullValueFormatConfiguration(tfList []interface{}) *quicksight.NullValueFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.NullValueFormatConfiguration{}

	if v, ok := tfMap["null_string"].(string); ok && v != "" {
		config.NullString = aws.String(v)
	}

	return config
}

func expandNumericFormatConfiguration(tfList []interface{}) *quicksight.NumericFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.NumericFormatConfiguration{}

	if v, ok := tfMap["currency_display_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.CurrencyDisplayFormatConfiguration = expandCurrencyDisplayFormatConfiguration(v)
	}

	return config
}

func expandCurrencyDisplayFormatConfiguration(tfList []interface{}) *quicksight.CurrencyDisplayFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.CurrencyDisplayFormatConfiguration{}

	if v, ok := tfMap["decimal_places_configuration"].([]interface{}); ok && len(v) > 0 {
		config.DecimalPlacesConfiguration = expandDecimalPlacesConfiguration(v)
	}
	if v, ok := tfMap["negative_value_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NegativeValueConfiguration = expandNegativeValueConfiguration(v)
	}
	if v, ok := tfMap["null_value_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["number_scale"].(string); ok && v != "" {
		config.NumberScale = aws.String(v)
	}
	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		config.Prefix = aws.String(v)
	}
	if v, ok := tfMap["separator_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SeparatorConfiguration = expandNumericSeparatorConfiguration(v)
	}
	if v, ok := tfMap["suffix"].(string); ok && v != "" {
		config.Suffix = aws.String(v)
	}
	if v, ok := tfMap["symbol"].(string); ok && v != "" {
		config.Symbol = aws.String(v)
	}

	return config
}

func expandDecimalPlacesConfiguration(tfList []interface{}) *quicksight.DecimalPlacesConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DecimalPlacesConfiguration{}

	if v, ok := tfMap["decimal_places"].(int64); ok {
		config.DecimalPlaces = aws.Int64(v)
	}

	return config
}

func expandNegativeValueConfiguration(tfList []interface{}) *quicksight.NegativeValueConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.NegativeValueConfiguration{}

	if v, ok := tfMap["display_mode"].(string); ok {
		config.DisplayMode = aws.String(v)
	}

	return config
}

func expandNumericSeparatorConfiguration(tfList []interface{}) *quicksight.NumericSeparatorConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.NumericSeparatorConfiguration{}

	if v, ok := tfMap["decimal_separator"].(string); ok {
		config.DecimalSeparator = aws.String(v)
	}
	if v, ok := tfMap["thousands_separator"].([]interface{}); ok && len(v) > 0 {
		config.ThousandsSeparator = expandThousandSeparatorOptions(v)
	}

	return config
}

func expandThousandSeparatorOptions(tfList []interface{}) *quicksight.ThousandSeparatorOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ThousandSeparatorOptions{}

	if v, ok := tfMap["symbol"].(string); ok {
		config.Symbol = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok {
		config.Visibility = aws.String(v)
	}

	return config
}

func expandNumberFormatConfiguration(tfList []interface{}) *quicksight.NumberFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.NumberFormatConfiguration{}

	if v, ok := tfMap["numeric_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.FormatConfiguration = expandNumericFormatConfiguration(v)
	}

	return config
}

func expandStringFormatConfiguration(tfList []interface{}) *quicksight.StringFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.StringFormatConfiguration{}

	if v, ok := tfMap["null_value_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["numeric_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NumericFormatConfiguration = expandNumericFormatConfiguration(v)
	}

	return config
}

func expandLabelOptions(tfList []interface{}) *quicksight.LabelOptions {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	options := &quicksight.LabelOptions{}

	if v, ok := tfMap["custom_label"].(string); ok {
		options.CustomLabel = aws.String(v)
	}
	if v, ok := tfMap["visibility"].(string); ok {
		options.Visibility = aws.String(v)
	}
	if v, ok := tfMap["font_configuration"].([]interface{}); ok && len(v) > 0 {
		options.FontConfiguration = expandFontConfiguration(v)
	}

	return options
}

func expandFontConfiguration(tfList []interface{}) *quicksight.FontConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FontConfiguration{}

	if v, ok := tfMap["font_color"].(string); ok {
		config.FontColor = aws.String(v)
	}
	if v, ok := tfMap["font_decoration"].(string); ok {
		config.FontDecoration = aws.String(v)
	}
	if v, ok := tfMap["font_style"].(string); ok {
		config.FontStyle = aws.String(v)
	}
	if v, ok := tfMap["font_size"].([]interface{}); ok && len(v) > 0 {
		config.FontSize = expandFontSize(v)
	}
	if v, ok := tfMap["font_weight"].([]interface{}); ok && len(v) > 0 {
		config.FontWeight = expandFontWeight(v)
	}

	return config
}

func expandFontSize(tfList []interface{}) *quicksight.FontSize {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FontSize{}

	if v, ok := tfMap["relative"].(string); ok {
		config.Relative = aws.String(v)
	}

	return config
}

func expandFontWeight(tfList []interface{}) *quicksight.FontWeight {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FontWeight{}

	if v, ok := tfMap["name"].(string); ok {
		config.Name = aws.String(v)
	}

	return config
}

func expandComparisonFormatConfiguration(tfList []interface{}) *quicksight.ComparisonFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.ComparisonFormatConfiguration{}

	if v, ok := tfMap["number_display_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NumberDisplayFormatConfiguration = expandNumberDisplayFormatConfiguration(v)
	}
	if v, ok := tfMap["percentage_display_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.PercentageDisplayFormatConfiguration = expandPercentageDisplayFormatConfiguration(v)
	}

	return config
}

func expandNumberDisplayFormatConfiguration(tfList []interface{}) *quicksight.NumberDisplayFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.NumberDisplayFormatConfiguration{}

	if v, ok := tfMap["number_scale"].(string); ok {
		config.NumberScale = aws.String(v)
	}
	if v, ok := tfMap["prefix"].(string); ok {
		config.Prefix = aws.String(v)
	}
	if v, ok := tfMap["suffix"].(string); ok {
		config.Suffix = aws.String(v)
	}
	if v, ok := tfMap["decimal_places_configuration"].([]interface{}); ok && len(v) > 0 {
		config.DecimalPlacesConfiguration = expandDecimalPlacesConfiguration(v)
	}
	if v, ok := tfMap["negative_value_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NegativeValueConfiguration = expandNegativeValueConfiguration(v)
	}
	if v, ok := tfMap["null_value_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["separator_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SeparatorConfiguration = expandNumericSeparatorConfiguration(v)
	}

	return config
}

func expandPercentageDisplayFormatConfiguration(tfList []interface{}) *quicksight.PercentageDisplayFormatConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.PercentageDisplayFormatConfiguration{}

	if v, ok := tfMap["prefix"].(string); ok {
		config.Prefix = aws.String(v)
	}
	if v, ok := tfMap["suffix"].(string); ok {
		config.Suffix = aws.String(v)
	}
	if v, ok := tfMap["decimal_places_configuration"].([]interface{}); ok && len(v) > 0 {
		config.DecimalPlacesConfiguration = expandDecimalPlacesConfiguration(v)
	}
	if v, ok := tfMap["negative_value_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NegativeValueConfiguration = expandNegativeValueConfiguration(v)
	}
	if v, ok := tfMap["null_value_format_configuration"].([]interface{}); ok && len(v) > 0 {
		config.NullValueFormatConfiguration = expandNullValueFormatConfiguration(v)
	}
	if v, ok := tfMap["separator_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SeparatorConfiguration = expandNumericSeparatorConfiguration(v)
	}

	return config
}
