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
					ValidateDiagFunc: validation.AllDiag(
						validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
						validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "")),
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
									Type:             schema.TypeString,
									ValidateDiagFunc: validation.ToDiagFunc(verify.ValidUTCTimestamp),
								},
							},
						},
					},
				},
				"time_granularity": stringSchema(false, enum.Validate[types.TimeGranularity]()),
				"values_when_unset": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeValueWhenUnsetConfiguration.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 1,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"custom_value": {
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validation.ToDiagFunc(verify.ValidUTCTimestamp),
							},
							"value_when_unset_option": stringSchema(false, enum.Validate[types.ValueWhenUnsetOption]()),
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
					ValidateDiagFunc: validation.AllDiag(
						validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
						validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "")),
					),
				},
				"parameter_value_type": stringSchema(true, enum.Validate[types.ParameterValueType]()),
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
							"value_when_unset_option": stringSchema(false, enum.Validate[types.ValueWhenUnsetOption]()),
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
					ValidateDiagFunc: validation.AllDiag(
						validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
						validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "")),
					),
				},
				"parameter_value_type": stringSchema(true, enum.Validate[types.ParameterValueType]()),
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
							"value_when_unset_option": stringSchema(false, enum.Validate[types.ValueWhenUnsetOption]()),
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
					ValidateDiagFunc: validation.AllDiag(
						validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
						validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "")),
					),
				},
				"parameter_value_type": stringSchema(true, enum.Validate[types.ValueWhenUnsetOption]()),
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
							"value_when_unset_option": stringSchema(false, enum.Validate[types.ValueWhenUnsetOption]()),
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
				"default_value_column": columnSchema(true),  // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"group_name_column":    columnSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
				"user_name_column":     columnSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
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
							"title":                 stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
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
							"title":                           stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 dropDownControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DropDownControlDisplayOptions.html
							"selectable_values":               parameterSelectableValuesSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSelectableValues.html
							"type":                            stringSchema(false, enum.Validate[types.SheetControlListType]()),
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
							"title":                           stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 listControlDisplayOptionsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlDisplayOptions.html
							"selectable_values":               parameterSelectableValuesSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSelectableValues.html
							"type":                            stringSchema(false, enum.Validate[types.SheetControlListType]()),
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
							"title":                 stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
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
							"title":                 stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
							"display_options":       textAreaControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextAreaControlDisplayOptions.html
							"delimiter":             stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
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
							"title":                 stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
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
				"link_to_data_set_column": columnSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
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
		ValidateDiagFunc: validation.AllDiag(
			validation.ToDiagFunc(validation.StringLenBetween(1, 2048)),
			validation.ToDiagFunc(validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "")),
		),
	}
}

func expandDateTimeParameterDeclaration(tfList []interface{}) *types.DateTimeParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	param := &types.DateTimeParameterDeclaration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		param.Name = aws.String(v)
	}
	if v, ok := tfMap["default_values"].([]interface{}); ok && len(v) > 0 {
		param.DefaultValues = expandDateTimeDefaultValues(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		param.TimeGranularity = types.TimeGranularity(v)
	}
	if v, ok := tfMap["values_when_unset"].([]interface{}); ok && len(v) > 0 {
		param.ValueWhenUnset = expandDateTimeValueWhenUnsetConfiguration(v)
	}

	return param
}

func expandDateTimeDefaultValues(tfList []interface{}) *types.DateTimeDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &types.DateTimeDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]interface{}); ok && len(v) > 0 {
		values.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["rolling_date"].([]interface{}); ok && len(v) > 0 {
		values.RollingDate = expandRollingDateConfiguration(v)
	}
	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 {
		values.StaticValues = flex.ExpandStringTimeValueList(v, time.RFC3339)
	}

	return values
}

func expandDynamicDefaultValue(tfList []interface{}) *types.DynamicDefaultValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	value := &types.DynamicDefaultValue{}

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

func expandDateTimeValueWhenUnsetConfiguration(tfList []interface{}) *types.DateTimeValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DateTimeValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		config.CustomValue = aws.Time(t)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		config.ValueWhenUnsetOption = types.ValueWhenUnsetOption(v)
	}

	return config
}

func expandDecimalParameterDeclaration(tfList []interface{}) *types.DecimalParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	param := &types.DecimalParameterDeclaration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		param.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		param.ParameterValueType = types.ParameterValueType(v)
	}
	if v, ok := tfMap["default_values"].([]interface{}); ok && len(v) > 0 {
		param.DefaultValues = expandDecimalDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]interface{}); ok && len(v) > 0 {
		param.ValueWhenUnset = expandDecimalValueWhenUnsetConfiguration(v)
	}

	return param
}

func expandDecimalValueWhenUnsetConfiguration(tfList []interface{}) *types.DecimalValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DecimalValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(float64); ok && v != 0.0 {
		config.CustomValue = aws.Float64(v)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		config.ValueWhenUnsetOption = types.ValueWhenUnsetOption(v)
	}

	return config
}

func expandDecimalDefaultValues(tfList []interface{}) *types.DecimalDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &types.DecimalDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]interface{}); ok && len(v) > 0 {
		values.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 {
		values.StaticValues = flex.ExpandFloat64ValueList(v)
	}

	return values
}

func expandIntegerParameterDeclaration(tfList []interface{}) *types.IntegerParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	param := &types.IntegerParameterDeclaration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		param.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		param.ParameterValueType = types.ParameterValueType(v)
	}
	if v, ok := tfMap["default_values"].([]interface{}); ok && len(v) > 0 {
		param.DefaultValues = expandIntegerDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]interface{}); ok && len(v) > 0 {
		param.ValueWhenUnset = expandIntegerValueWhenUnsetConfiguration(v)
	}

	return param
}

func expandIntegerValueWhenUnsetConfiguration(tfList []interface{}) *types.IntegerValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.IntegerValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(int); ok && v != 0 {
		config.CustomValue = aws.Int64(int64(v))
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		config.ValueWhenUnsetOption = types.ValueWhenUnsetOption(v)
	}

	return config
}

func expandIntegerDefaultValues(tfList []interface{}) *types.IntegerDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &types.IntegerDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]interface{}); ok && len(v) > 0 {
		values.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 {
		values.StaticValues = flex.ExpandInt64ValueList(v)
	}

	return values
}

func expandStringParameterDeclaration(tfList []interface{}) *types.StringParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	param := &types.StringParameterDeclaration{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		param.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		param.ParameterValueType = types.ParameterValueType(v)
	}
	if v, ok := tfMap["default_values"].([]interface{}); ok && len(v) > 0 {
		param.DefaultValues = expandStringDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]interface{}); ok && len(v) > 0 {
		param.ValueWhenUnset = expandStringValueWhenUnsetConfiguration(v)
	}

	return param
}

func expandStringValueWhenUnsetConfiguration(tfList []interface{}) *types.StringValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.StringValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(string); ok && v != "" {
		config.CustomValue = aws.String(v)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		config.ValueWhenUnsetOption = types.ValueWhenUnsetOption(v)
	}

	return config
}

func expandStringDefaultValues(tfList []interface{}) *types.StringDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &types.StringDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]interface{}); ok && len(v) > 0 {
		values.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]interface{}); ok && len(v) > 0 {
		values.StaticValues = flex.ExpandStringValueList(v)
	}

	return values
}

func expandParameterSelectableValues(tfList []interface{}) *types.ParameterSelectableValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	values := &types.ParameterSelectableValues{}

	if v, ok := tfMap["link_to_data_set_column"].([]interface{}); ok && len(v) > 0 {
		values.LinkToDataSetColumn = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["values"].([]interface{}); ok && len(v) > 0 {
		values.Values = flex.ExpandStringValueList(v)
	}

	return values
}

func flattenDateTimeParameterDeclaration(apiObject *types.DateTimeParameterDeclaration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DefaultValues != nil {
		tfMap["default_values"] = flattenDateTimeDefaultValues(apiObject.DefaultValues)
	}
	if apiObject.Name != nil {
		tfMap["name"] = aws.ToString(apiObject.Name)
	}
	tfMap["time_granularity"] = types.ParameterValueType(apiObject.TimeGranularity)
	if apiObject.ValueWhenUnset != nil {
		tfMap["values_when_unset"] = flattenDateTimeValueWhenUnsetConfiguration(apiObject.ValueWhenUnset)
	}

	return []interface{}{tfMap}
}

func flattenDateTimeDefaultValues(apiObject *types.DateTimeDefaultValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DynamicValue != nil {
		tfMap["dynamic_value"] = flattenDynamicDefaultValue(apiObject.DynamicValue)
	}
	if apiObject.RollingDate != nil {
		tfMap["rolling_date"] = flattenRollingDateConfiguration(apiObject.RollingDate)
	}
	if len(apiObject.StaticValues) > 0 {
		tfMap["static_values"] = flex.FlattenTimeStringValueList(apiObject.StaticValues, time.RFC3339)
	}

	return []interface{}{tfMap}
}

func flattenDynamicDefaultValue(apiObject *types.DynamicDefaultValue) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DefaultValueColumn != nil {
		tfMap["default_value_column"] = flattenColumnIdentifier(apiObject.DefaultValueColumn)
	}
	if apiObject.GroupNameColumn != nil {
		tfMap["group_name_column"] = flattenColumnIdentifier(apiObject.GroupNameColumn)
	}
	if apiObject.UserNameColumn != nil {
		tfMap["user_name_column"] = flattenColumnIdentifier(apiObject.UserNameColumn)
	}

	return []interface{}{tfMap}
}

func flattenDateTimeValueWhenUnsetConfiguration(apiObject *types.DateTimeValueWhenUnsetConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomValue != nil {
		tfMap["custom_value"] = apiObject.CustomValue.Format(time.RFC3339)
	}
	tfMap["value_when_unset_option"] = types.ValueWhenUnsetOption(apiObject.ValueWhenUnsetOption)

	return []interface{}{tfMap}
}

func flattenDecimalParameterDeclaration(apiObject *types.DecimalParameterDeclaration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DefaultValues != nil {
		tfMap["default_values"] = flattenDecimalDefaultValues(apiObject.DefaultValues)
	}
	if apiObject.Name != nil {
		tfMap["name"] = aws.ToString(apiObject.Name)
	}
	tfMap["parameter_value_type"] = types.ParameterValueType(apiObject.ParameterValueType)
	if apiObject.ValueWhenUnset != nil {
		tfMap["values_when_unset"] = flattenDecimalValueWhenUnsetConfiguration(apiObject.ValueWhenUnset)
	}

	return []interface{}{tfMap}
}

func flattenDecimalDefaultValues(apiObject *types.DecimalDefaultValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DynamicValue != nil {
		tfMap["dynamic_value"] = flattenDynamicDefaultValue(apiObject.DynamicValue)
	}
	if len(apiObject.StaticValues) > 0 {
		tfMap["static_values"] = apiObject.StaticValues
	}

	return []interface{}{tfMap}
}

func flattenDecimalValueWhenUnsetConfiguration(apiObject *types.DecimalValueWhenUnsetConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomValue != nil {
		tfMap["custom_value"] = aws.ToFloat64(apiObject.CustomValue)
	}
	tfMap["value_when_unset_option"] = types.ValueWhenUnsetOption(apiObject.ValueWhenUnsetOption)

	return []interface{}{tfMap}
}

func flattenIntegerParameterDeclaration(apiObject *types.IntegerParameterDeclaration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DefaultValues != nil {
		tfMap["default_values"] = flattenIntegerDefaultValues(apiObject.DefaultValues)
	}
	if apiObject.Name != nil {
		tfMap["name"] = aws.ToString(apiObject.Name)
	}
	tfMap["parameter_value_type"] = types.ParameterValueType(apiObject.ParameterValueType)
	if apiObject.ValueWhenUnset != nil {
		tfMap["values_when_unset"] = flattenIntegerValueWhenUnsetConfiguration(apiObject.ValueWhenUnset)
	}

	return []interface{}{tfMap}
}

func flattenIntegerDefaultValues(apiObject *types.IntegerDefaultValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DynamicValue != nil {
		tfMap["dynamic_value"] = flattenDynamicDefaultValue(apiObject.DynamicValue)
	}
	if len(apiObject.StaticValues) > 0 {
		tfMap["static_values"] = apiObject.StaticValues
	}

	return []interface{}{tfMap}
}

func flattenIntegerValueWhenUnsetConfiguration(apiObject *types.IntegerValueWhenUnsetConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomValue != nil {
		tfMap["custom_value"] = aws.ToInt64(apiObject.CustomValue)
	}
	tfMap["value_when_unset_option"] = types.ValueWhenUnsetOption(apiObject.ValueWhenUnsetOption)

	return []interface{}{tfMap}
}

func flattenStringParameterDeclaration(apiObject *types.StringParameterDeclaration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DefaultValues != nil {
		tfMap["default_values"] = flattenStringDefaultValues(apiObject.DefaultValues)
	}
	if apiObject.Name != nil {
		tfMap["name"] = aws.ToString(apiObject.Name)
	}
	tfMap["parameter_value_type"] = types.ParameterValueType(apiObject.ParameterValueType)
	if apiObject.ValueWhenUnset != nil {
		tfMap["values_when_unset"] = flattenStringValueWhenUnsetConfiguration(apiObject.ValueWhenUnset)
	}

	return []interface{}{tfMap}
}

func flattenStringDefaultValues(apiObject *types.StringDefaultValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DynamicValue != nil {
		tfMap["dynamic_value"] = flattenDynamicDefaultValue(apiObject.DynamicValue)
	}
	if len(apiObject.StaticValues) > 0 {
		tfMap["static_values"] = flex.FlattenStringValueList(apiObject.StaticValues)
	}

	return []interface{}{tfMap}
}

func flattenStringValueWhenUnsetConfiguration(apiObject *types.StringValueWhenUnsetConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomValue != nil {
		tfMap["custom_value"] = aws.ToString(apiObject.CustomValue)
	}
	tfMap["value_when_unset_option"] = types.ValueWhenUnsetOption(apiObject.ValueWhenUnsetOption)

	return []interface{}{tfMap}
}

func flattenParameterControls(apiObject []types.ParameterControl) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		tfMap := map[string]interface{}{}
		if config.DateTimePicker != nil {
			tfMap["date_time_picker"] = flattenParameterDateTimePickerControl(config.DateTimePicker)
		}
		if config.Dropdown != nil {
			tfMap["dropdown"] = flattenParameterDropDownControl(config.Dropdown)
		}
		if config.List != nil {
			tfMap["list"] = flattenParameterListControl(config.List)
		}
		if config.Slider != nil {
			tfMap["slider"] = flattenParameterSliderControl(config.Slider)
		}
		if config.TextArea != nil {
			tfMap["text_area"] = flattenParameterTextAreaControl(config.TextArea)
		}
		if config.TextField != nil {
			tfMap["text_field"] = flattenParameterTextFieldControl(config.TextField)
		}
		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenParameterDateTimePickerControl(apiObject *types.ParameterDateTimePickerControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDateTimePickerControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenParameterDropDownControl(apiObject *types.ParameterDropDownControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
	}
	if apiObject.CascadingControlConfiguration != nil {
		tfMap["cascading_control_configuration"] = flattenCascadingControlConfiguration(apiObject.CascadingControlConfiguration)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDropDownControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.SelectableValues != nil {
		tfMap["selectable_values"] = flattenParameterSelectableValues(apiObject.SelectableValues)
	}
	tfMap["type"] = types.SheetControlListType(apiObject.Type)

	return []interface{}{tfMap}
}

func flattenParameterSelectableValues(apiObject *types.ParameterSelectableValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LinkToDataSetColumn != nil {
		tfMap["link_to_data_set_column"] = flattenColumnIdentifier(apiObject.LinkToDataSetColumn)
	}
	if apiObject.Values != nil {
		tfMap["values"] = flex.FlattenStringValueList(apiObject.Values)
	}

	return []interface{}{tfMap}
}

func flattenParameterListControl(apiObject *types.ParameterListControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
	}
	if apiObject.CascadingControlConfiguration != nil {
		tfMap["cacading_control_configuration"] = flattenCascadingControlConfiguration(apiObject.CascadingControlConfiguration)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenListControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.SelectableValues != nil {
		tfMap["selectable_values"] = flattenParameterSelectableValues(apiObject.SelectableValues)
	}
	tfMap["type"] = types.SheetControlListType(apiObject.Type)

	return []interface{}{tfMap}
}

func flattenParameterSliderControl(apiObject *types.ParameterSliderControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
		"maximum_value":         apiObject.MaximumValue,
		"minimum_value":         apiObject.MinimumValue,
		"step_size":             apiObject.StepSize,
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenSliderControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenParameterTextAreaControl(apiObject *types.ParameterTextAreaControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
	}
	if apiObject.Delimiter != nil {
		tfMap["delimiter"] = aws.ToString(apiObject.Delimiter)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextAreaControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}

func flattenParameterTextFieldControl(apiObject *types.ParameterTextFieldControl) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextFieldControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []interface{}{tfMap}
}
