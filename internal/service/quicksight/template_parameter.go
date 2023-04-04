package quicksight

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

func dateTimeParameterDeclarationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeParameterDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), ""),
					),
				},
				"default_values": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeDefaultValues.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"dynamic_value": dynamicValueSchema(),             // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DynamicDefaultValue.html
							"rolling_date":  rollingDateConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_RollingDateConfiguration.html,
							"static_values": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 50000,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: verify.ValidUTCTimestamp,
								},
							},
						},
					},
				},
				"time_granularity": stringSchema(false, validation.StringInSlice(quicksight.TimeGranularity_Values(), false)),
				"values_when_unset": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeValueWhenUnsetConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"custom_value": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: verify.ValidUTCTimestamp,
							},
							"value_when_unset_option": stringSchema(false, validation.StringInSlice(quicksight.ValueWhenUnsetOption_Values(), false)),
						},
					},
				},
			},
		},
	}
}

func decimalParameterDeclarationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalParameterDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), ""),
					),
				},
				"parameter_value_type": stringSchema(true, validation.StringInSlice(quicksight.ParameterValueType_Values(), false)),
				"default_values": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalDefaultValues.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"dynamic_value": dynamicValueSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DynamicDefaultValue.html
							"static_values": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 50000,
								Elem: &schema.Schema{
									Type: schema.TypeFloat,
								},
							},
						},
					},
				},
				"values_when_unset": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalValueWhenUnsetConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"custom_value": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
							"value_when_unset_option": stringSchema(false, validation.StringInSlice(quicksight.ValueWhenUnsetOption_Values(), false)),
						},
					},
				},
			},
		},
	}
}

func integerParameterDeclarationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_IntegerParameterDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), ""),
					),
				},
				"parameter_value_type": stringSchema(true, validation.StringInSlice(quicksight.ParameterValueType_Values(), false)),
				"default_values": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_IntegerDefaultValues.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"dynamic_value": dynamicValueSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DynamicDefaultValue.html
							"static_values": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 1,
								MaxItems: 50000,
								Elem: &schema.Schema{
									Type: schema.TypeInt,
								},
							},
						},
					},
				},
				"values_when_unset": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_IntegerValueWhenUnsetConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"custom_value": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"value_when_unset_option": stringSchema(false, validation.StringInSlice(quicksight.ValueWhenUnsetOption_Values(), false)),
						},
					},
				},
			},
		},
	}
}

func stringParameterDeclarationSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringParameterDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), ""),
					),
				},
				"parameter_value_type": stringSchema(true, validation.StringInSlice(quicksight.ParameterValueType_Values(), false)),
				"default_values": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringDefaultValues.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"dynamic_value": dynamicValueSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DynamicDefaultValue.html
							"static_values": {
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
				},
				"values_when_unset": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringValueWhenUnsetConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"custom_value": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"value_when_unset_option": stringSchema(false, validation.StringInSlice(quicksight.ValueWhenUnsetOption_Values(), false)),
						},
					},
				},
			},
		},
	}
}

func dynamicValueSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DynamicDefaultValue.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"default_value_column": columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"group_name_column":    columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"user_name_column":     columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
			},
		},
	}
}

func parameterControlsSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterControl.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 200,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_time_picker": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterDateTimePickerControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"parameter_control_id":  idSchema(),
							"source_parameter_name": parameterNameSchema(true),
							"title":                 stringSchema(true, validation.StringLenBetween(1, 2048)),
							"display_options":       dateTimePickerControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimePickerControlDisplayOptions.html
						},
					},
				},
				"dropdown": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterDropDownControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"parameter_control_id":            idSchema(),
							"source_parameter_name":           parameterNameSchema(true),
							"title":                           stringSchema(true, validation.StringLenBetween(1, 2048)),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 dropDownControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DropDownControlDisplayOptions.html
							"selectable_values":               parameterSelectableValuesSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSelectableValues.html
							"type":                            stringSchema(false, validation.StringInSlice(quicksight.SheetControlListType_Values(), false)),
						},
					},
				},
				"list": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterListControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"parameter_control_id":            idSchema(),
							"source_parameter_name":           parameterNameSchema(true),
							"title":                           stringSchema(true, validation.StringLenBetween(1, 2048)),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 listControlDisplayOptionsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlDisplayOptions.html
							"selectable_values":               parameterSelectableValuesSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSelectableValues.html
							"type":                            stringSchema(false, validation.StringInSlice(quicksight.SheetControlListType_Values(), false)),
						},
					},
				},
				"slider": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSliderControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"parameter_control_id":  idSchema(),
							"source_parameter_name": parameterNameSchema(true),
							"title":                 stringSchema(true, validation.StringLenBetween(1, 2048)),
							"display_options":       sliderControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SliderControlDisplayOptions.html
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
						},
					},
				},
				"text_area": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterTextAreaControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"parameter_control_id":  idSchema(),
							"source_parameter_name": parameterNameSchema(true),
							"title":                 stringSchema(true, validation.StringLenBetween(1, 2048)),
							"display_options":       textAreaControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextAreaControlDisplayOptions.html
							"delimiter":             stringSchema(false, validation.StringLenBetween(1, 2048)),
						},
					},
				},
				"text_field": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterTextFieldControl.html
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"parameter_control_id":  idSchema(),
							"source_parameter_name": parameterNameSchema(true),
							"title":                 stringSchema(true, validation.StringLenBetween(1, 2048)),
							"display_options":       textFieldControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextFieldControlDisplayOptions.html
						},
					},
				},
			},
		},
	}
}

func parameterSelectableValuesSchema() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSelectableValues.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"link_to_data_set_column": columnSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
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

func parameterNameSchema(required bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: required,
		Optional: !required,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 2048),
			validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9]+$`), ""),
		),
	}
}

func expandDateTimeParameterDeclaration(tfList []interface{}) *quicksight.DateTimeParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	param := &quicksight.DateTimeParameterDeclaration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		param.Name = aws.String(v)
	}
	if v, ok := tfMap["default_values"].([]interface{}); ok && len(v) > 0 {
		param.DefaultValues = expandDateTimeDefaultValues(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		param.TimeGranularity = aws.String(v)
	}
	if v, ok := tfMap["values_when_unset"].([]interface{}); ok && len(v) > 0 {
		param.ValueWhenUnset = expandDateTimeValueWhenUnsetConfiguration(v)
	}

	return param
}

func expandDateTimeDefaultValues(tfList []interface{}) *quicksight.DateTimeDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &quicksight.DateTimeDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]interface{}); ok && len(v) > 0 {
		values.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["rolling_date"].([]interface{}); ok && len(v) > 0 {
		values.RollingDate = expandRollingDateConfiguration(v)
	}
	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 {
		values.StaticValues = flex.ExpandStringTimeList(v, time.RFC3339)
	}

	return values
}

func expandDynamicDefaultValue(tfList []interface{}) *quicksight.DynamicDefaultValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	value := &quicksight.DynamicDefaultValue{}

	if v, ok := tfMap["default_value_column"].([]interface{}); ok && len(v) > 0 {
		value.DefaultValueColumn = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["group_name_column"].([]interface{}); ok && len(v) > 0 {
		value.GroupNameColumn = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["user_name_column"].([]interface{}); ok && len(v) > 0 {
		value.UserNameColumn = expandColumnIdentifier(v)
	}

	return value
}

func expandDateTimeValueWhenUnsetConfiguration(tfList []interface{}) *quicksight.DateTimeValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DateTimeValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		config.CustomValue = aws.Time(t)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		config.ValueWhenUnsetOption = aws.String(v)
	}

	return config
}

func expandDecimalParameterDeclaration(tfList []interface{}) *quicksight.DecimalParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	param := &quicksight.DecimalParameterDeclaration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		param.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		param.ParameterValueType = aws.String(v)
	}
	if v, ok := tfMap["default_values"].([]interface{}); ok && len(v) > 0 {
		param.DefaultValues = expandDecimalDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]interface{}); ok && len(v) > 0 {
		param.ValueWhenUnset = expandDecimalValueWhenUnsetConfiguration(v)
	}

	return param
}

func expandDecimalValueWhenUnsetConfiguration(tfList []interface{}) *quicksight.DecimalValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DecimalValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(float64); ok {
		config.CustomValue = aws.Float64(v)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		config.ValueWhenUnsetOption = aws.String(v)
	}

	return config
}

func expandDecimalDefaultValues(tfList []interface{}) *quicksight.DecimalDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &quicksight.DecimalDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]interface{}); ok && len(v) > 0 {
		values.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 {
		values.StaticValues = flex.ExpandFloat64List(v)
	}

	return values
}

func expandIntegerParameterDeclaration(tfList []interface{}) *quicksight.IntegerParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	param := &quicksight.IntegerParameterDeclaration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		param.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		param.ParameterValueType = aws.String(v)
	}
	if v, ok := tfMap["default_values"].([]interface{}); ok && len(v) > 0 {
		param.DefaultValues = expandIntegerDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]interface{}); ok && len(v) > 0 {
		param.ValueWhenUnset = expandIntegerValueWhenUnsetConfiguration(v)
	}

	return param
}

func expandIntegerValueWhenUnsetConfiguration(tfList []interface{}) *quicksight.IntegerValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.IntegerValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(int64); ok {
		config.CustomValue = aws.Int64(v)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		config.ValueWhenUnsetOption = aws.String(v)
	}

	return config
}

func expandIntegerDefaultValues(tfList []interface{}) *quicksight.IntegerDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &quicksight.IntegerDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]interface{}); ok && len(v) > 0 {
		values.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 {
		values.StaticValues = flex.ExpandInt64List(v)
	}

	return values
}

func expandStringParameterDeclaration(tfList []interface{}) *quicksight.StringParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	param := &quicksight.StringParameterDeclaration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		param.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		param.ParameterValueType = aws.String(v)
	}
	if v, ok := tfMap["default_values"].([]interface{}); ok && len(v) > 0 {
		param.DefaultValues = expandStringDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]interface{}); ok && len(v) > 0 {
		param.ValueWhenUnset = expandStringValueWhenUnsetConfiguration(v)
	}

	return param
}

func expandStringValueWhenUnsetConfiguration(tfList []interface{}) *quicksight.StringValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.StringValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(string); ok {
		config.CustomValue = aws.String(v)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		config.ValueWhenUnsetOption = aws.String(v)
	}

	return config
}

func expandStringDefaultValues(tfList []interface{}) *quicksight.StringDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &quicksight.StringDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]interface{}); ok && len(v) > 0 {
		values.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 {
		values.StaticValues = flex.ExpandStringList(v)
	}

	return values
}

func expandParameterSelectableValues(tfList []interface{}) *quicksight.ParameterSelectableValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &quicksight.ParameterSelectableValues{}

	if v, ok := tfMap["link_to_data_set_column"].([]interface{}); ok && len(v) > 0 {
		values.LinkToDataSetColumn = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		values.Values = flex.ExpandStringList(v)
	}

	return values
}
