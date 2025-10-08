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

var dateTimeParameterDeclarationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DateTimeParameterDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), ""),
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
				"time_granularity": stringEnumSchema[awstypes.TimeGranularity](attrOptional),
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
							"value_when_unset_option": stringEnumSchema[awstypes.ValueWhenUnsetOption](attrOptional),
						},
					},
				},
			},
		},
	}
})

var decimalParameterDeclarationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DecimalParameterDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), ""),
					),
				},
				"parameter_value_type": stringEnumSchema[awstypes.ParameterValueType](attrRequired),
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
							"value_when_unset_option": stringEnumSchema[awstypes.ValueWhenUnsetOption](attrOptional),
						},
					},
				},
			},
		},
	}
})

var integerParameterDeclarationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_IntegerParameterDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), ""),
					),
				},
				"parameter_value_type": stringEnumSchema[awstypes.ParameterValueType](attrRequired),
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
							"value_when_unset_option": stringEnumSchema[awstypes.ValueWhenUnsetOption](attrOptional),
						},
					},
				},
			},
		},
	}
})

var stringParameterDeclarationSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_StringParameterDeclaration.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 2048),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), ""),
					),
				},
				"parameter_value_type": stringEnumSchema[awstypes.ParameterValueType](attrRequired),
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
							"value_when_unset_option": stringEnumSchema[awstypes.ValueWhenUnsetOption](attrOptional),
						},
					},
				},
			},
		},
	}
})

var dynamicValueSchema = sync.OnceValue(func() *schema.Schema {
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
})

var parameterControlsSchema = sync.OnceValue(func() *schema.Schema {
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
							"title":                 stringLenBetweenSchema(attrRequired, 1, 2048),
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
							"title":                           stringLenBetweenSchema(attrRequired, 1, 2048),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 dropDownControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DropDownControlDisplayOptions.html
							"selectable_values":               parameterSelectableValuesSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSelectableValues.html
							names.AttrType:                    stringEnumSchema[awstypes.SheetControlListType](attrOptional),
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
							"title":                           stringLenBetweenSchema(attrRequired, 1, 2048),
							"cascading_control_configuration": cascadingControlConfigurationSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CascadingControlConfiguration.html
							"display_options":                 listControlDisplayOptionsSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ListControlDisplayOptions.html
							"selectable_values":               parameterSelectableValuesSchema(),     // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSelectableValues.html
							names.AttrType:                    stringEnumSchema[awstypes.SheetControlListType](attrOptional),
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
							"title":                 stringLenBetweenSchema(attrRequired, 1, 2048),
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
							"title":                 stringLenBetweenSchema(attrRequired, 1, 2048),
							"display_options":       textAreaControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextAreaControlDisplayOptions.html
							"delimiter":             stringLenBetweenSchema(attrOptional, 1, 2048),
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
							"title":                 stringLenBetweenSchema(attrRequired, 1, 2048),
							"display_options":       textFieldControlDisplayOptionsSchema(), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_TextFieldControlDisplayOptions.html
						},
					},
				},
			},
		},
	}
})

var parameterSelectableValuesSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ParameterSelectableValues.html
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"link_to_data_set_column": columnSchema(false), // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_ColumnIdentifier.html
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

func parameterNameSchema(required bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Required: required,
		Optional: !required,
		ValidateFunc: validation.All(
			validation.StringLenBetween(1, 2048),
			validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), ""),
		),
	}
}

func expandDateTimeParameterDeclaration(tfList []any) *awstypes.DateTimeParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateTimeParameterDeclaration{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["default_values"].([]any); ok && len(v) > 0 {
		apiObject.DefaultValues = expandDateTimeDefaultValues(v)
	}
	if v, ok := tfMap["time_granularity"].(string); ok && v != "" {
		apiObject.TimeGranularity = awstypes.TimeGranularity(v)
	}
	if v, ok := tfMap["values_when_unset"].([]any); ok && len(v) > 0 {
		apiObject.ValueWhenUnset = expandDateTimeValueWhenUnsetConfiguration(v)
	}

	return apiObject
}

func expandDateTimeDefaultValues(tfList []any) *awstypes.DateTimeDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateTimeDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]any); ok && len(v) > 0 {
		apiObject.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["rolling_date"].([]any); ok && len(v) > 0 {
		apiObject.RollingDate = expandRollingDateConfiguration(v)
	}
	if v, ok := tfMap["static_values"].([]any); ok && len(v) > 0 {
		apiObject.StaticValues = flex.ExpandStringTimeValueList(v, time.RFC3339)
	}

	return apiObject
}

func expandDynamicDefaultValue(tfList []any) *awstypes.DynamicDefaultValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DynamicDefaultValue{}

	if v, ok := tfMap["default_value_column"].([]any); ok && len(v) > 0 {
		apiObject.DefaultValueColumn = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["group_name_column"].([]any); ok && len(v) > 0 {
		apiObject.GroupNameColumn = expandColumnIdentifier(v)
	}
	if v, ok := tfMap["user_name_column"].([]any); ok && len(v) > 0 {
		apiObject.UserNameColumn = expandColumnIdentifier(v)
	}

	return apiObject
}

func expandDateTimeValueWhenUnsetConfiguration(tfList []any) *awstypes.DateTimeValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DateTimeValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(string); ok && v != "" {
		t, _ := time.Parse(time.RFC3339, v) // Format validated with validateFunc
		apiObject.CustomValue = aws.Time(t)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		apiObject.ValueWhenUnsetOption = awstypes.ValueWhenUnsetOption(v)
	}

	return apiObject
}

func expandDecimalParameterDeclaration(tfList []any) *awstypes.DecimalParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DecimalParameterDeclaration{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		apiObject.ParameterValueType = awstypes.ParameterValueType(v)
	}
	if v, ok := tfMap["default_values"].([]any); ok && len(v) > 0 {
		apiObject.DefaultValues = expandDecimalDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]any); ok && len(v) > 0 {
		apiObject.ValueWhenUnset = expandDecimalValueWhenUnsetConfiguration(v)
	}

	return apiObject
}

func expandDecimalValueWhenUnsetConfiguration(tfList []any) *awstypes.DecimalValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DecimalValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(float64); ok && v != 0.0 {
		apiObject.CustomValue = aws.Float64(v)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		apiObject.ValueWhenUnsetOption = awstypes.ValueWhenUnsetOption(v)
	}

	return apiObject
}

func expandDecimalDefaultValues(tfList []any) *awstypes.DecimalDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.DecimalDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]any); ok && len(v) > 0 {
		apiObject.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]any); ok && len(v) > 0 {
		apiObject.StaticValues = flex.ExpandFloat64ValueList(v)
	}

	return apiObject
}

func expandIntegerParameterDeclaration(tfList []any) *awstypes.IntegerParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.IntegerParameterDeclaration{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		apiObject.ParameterValueType = awstypes.ParameterValueType(v)
	}
	if v, ok := tfMap["default_values"].([]any); ok && len(v) > 0 {
		apiObject.DefaultValues = expandIntegerDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]any); ok && len(v) > 0 {
		apiObject.ValueWhenUnset = expandIntegerValueWhenUnsetConfiguration(v)
	}

	return apiObject
}

func expandIntegerValueWhenUnsetConfiguration(tfList []any) *awstypes.IntegerValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.IntegerValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(int); ok && v != 0 {
		apiObject.CustomValue = aws.Int64(int64(v))
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		apiObject.ValueWhenUnsetOption = awstypes.ValueWhenUnsetOption(v)
	}

	return apiObject
}

func expandIntegerDefaultValues(tfList []any) *awstypes.IntegerDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.IntegerDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]any); ok && len(v) > 0 {
		apiObject.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]any); ok && len(v) > 0 {
		apiObject.StaticValues = flex.ExpandInt64ValueList(v)
	}

	return apiObject
}

func expandStringParameterDeclaration(tfList []any) *awstypes.StringParameterDeclaration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.StringParameterDeclaration{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["parameter_value_type"].(string); ok && v != "" {
		apiObject.ParameterValueType = awstypes.ParameterValueType(v)
	}
	if v, ok := tfMap["default_values"].([]any); ok && len(v) > 0 {
		apiObject.DefaultValues = expandStringDefaultValues(v)
	}
	if v, ok := tfMap["values_when_unset"].([]any); ok && len(v) > 0 {
		apiObject.ValueWhenUnset = expandStringValueWhenUnsetConfiguration(v)
	}

	return apiObject
}

func expandStringValueWhenUnsetConfiguration(tfList []any) *awstypes.StringValueWhenUnsetConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.StringValueWhenUnsetConfiguration{}

	if v, ok := tfMap["custom_value"].(string); ok && v != "" {
		apiObject.CustomValue = aws.String(v)
	}
	if v, ok := tfMap["value_when_unset_option"].(string); ok && v != "" {
		apiObject.ValueWhenUnsetOption = awstypes.ValueWhenUnsetOption(v)
	}

	return apiObject
}

func expandStringDefaultValues(tfList []any) *awstypes.StringDefaultValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.StringDefaultValues{}

	if v, ok := tfMap["dynamic_value"].([]any); ok && len(v) > 0 {
		apiObject.DynamicValue = expandDynamicDefaultValue(v)
	}
	if v, ok := tfMap["static_values"].([]any); ok && len(v) > 0 {
		apiObject.StaticValues = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandParameterSelectableValues(tfList []any) *awstypes.ParameterSelectableValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &awstypes.ParameterSelectableValues{}

	if v, ok := tfMap["link_to_data_set_column"].([]any); ok && len(v) > 0 {
		apiObject.LinkToDataSetColumn = expandColumnIdentifier(v)
	}
	if v, ok := tfMap[names.AttrValues].([]any); ok && len(v) > 0 {
		apiObject.Values = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func flattenDateTimeParameterDeclaration(apiObject *awstypes.DateTimeParameterDeclaration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DefaultValues != nil {
		tfMap["default_values"] = flattenDateTimeDefaultValues(apiObject.DefaultValues)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	tfMap["time_granularity"] = apiObject.TimeGranularity
	if apiObject.ValueWhenUnset != nil {
		tfMap["values_when_unset"] = flattenDateTimeValueWhenUnsetConfiguration(apiObject.ValueWhenUnset)
	}

	return []any{tfMap}
}

func flattenDateTimeDefaultValues(apiObject *awstypes.DateTimeDefaultValues) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DynamicValue != nil {
		tfMap["dynamic_value"] = flattenDynamicDefaultValue(apiObject.DynamicValue)
	}
	if apiObject.RollingDate != nil {
		tfMap["rolling_date"] = flattenRollingDateConfiguration(apiObject.RollingDate)
	}
	if len(apiObject.StaticValues) > 0 {
		tfMap["static_values"] = flex.FlattenTimeStringValueList(apiObject.StaticValues, time.RFC3339)
	}

	return []any{tfMap}
}

func flattenDynamicDefaultValue(apiObject *awstypes.DynamicDefaultValue) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DefaultValueColumn != nil {
		tfMap["default_value_column"] = flattenColumnIdentifier(apiObject.DefaultValueColumn)
	}
	if apiObject.GroupNameColumn != nil {
		tfMap["group_name_column"] = flattenColumnIdentifier(apiObject.GroupNameColumn)
	}
	if apiObject.UserNameColumn != nil {
		tfMap["user_name_column"] = flattenColumnIdentifier(apiObject.UserNameColumn)
	}

	return []any{tfMap}
}

func flattenDateTimeValueWhenUnsetConfiguration(apiObject *awstypes.DateTimeValueWhenUnsetConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CustomValue != nil {
		tfMap["custom_value"] = apiObject.CustomValue.Format(time.RFC3339)
	}
	tfMap["value_when_unset_option"] = apiObject.ValueWhenUnsetOption

	return []any{tfMap}
}

func flattenDecimalParameterDeclaration(apiObject *awstypes.DecimalParameterDeclaration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DefaultValues != nil {
		tfMap["default_values"] = flattenDecimalDefaultValues(apiObject.DefaultValues)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	tfMap["parameter_value_type"] = apiObject.ParameterValueType
	if apiObject.ValueWhenUnset != nil {
		tfMap["values_when_unset"] = flattenDecimalValueWhenUnsetConfiguration(apiObject.ValueWhenUnset)
	}

	return []any{tfMap}
}

func flattenDecimalDefaultValues(apiObject *awstypes.DecimalDefaultValues) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DynamicValue != nil {
		tfMap["dynamic_value"] = flattenDynamicDefaultValue(apiObject.DynamicValue)
	}
	if len(apiObject.StaticValues) > 0 {
		tfMap["static_values"] = apiObject.StaticValues
	}

	return []any{tfMap}
}

func flattenDecimalValueWhenUnsetConfiguration(apiObject *awstypes.DecimalValueWhenUnsetConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CustomValue != nil {
		tfMap["custom_value"] = aws.ToFloat64(apiObject.CustomValue)
	}
	tfMap["value_when_unset_option"] = apiObject.ValueWhenUnsetOption

	return []any{tfMap}
}

func flattenIntegerParameterDeclaration(apiObject *awstypes.IntegerParameterDeclaration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DefaultValues != nil {
		tfMap["default_values"] = flattenIntegerDefaultValues(apiObject.DefaultValues)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	tfMap["parameter_value_type"] = apiObject.ParameterValueType
	if apiObject.ValueWhenUnset != nil {
		tfMap["values_when_unset"] = flattenIntegerValueWhenUnsetConfiguration(apiObject.ValueWhenUnset)
	}

	return []any{tfMap}
}

func flattenIntegerDefaultValues(apiObject *awstypes.IntegerDefaultValues) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DynamicValue != nil {
		tfMap["dynamic_value"] = flattenDynamicDefaultValue(apiObject.DynamicValue)
	}
	if len(apiObject.StaticValues) > 0 {
		tfMap["static_values"] = apiObject.StaticValues
	}

	return []any{tfMap}
}

func flattenIntegerValueWhenUnsetConfiguration(apiObject *awstypes.IntegerValueWhenUnsetConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CustomValue != nil {
		tfMap["custom_value"] = aws.ToInt64(apiObject.CustomValue)
	}
	tfMap["value_when_unset_option"] = apiObject.ValueWhenUnsetOption

	return []any{tfMap}
}

func flattenStringParameterDeclaration(apiObject *awstypes.StringParameterDeclaration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DefaultValues != nil {
		tfMap["default_values"] = flattenStringDefaultValues(apiObject.DefaultValues)
	}
	if apiObject.Name != nil {
		tfMap[names.AttrName] = aws.ToString(apiObject.Name)
	}
	tfMap["parameter_value_type"] = apiObject.ParameterValueType
	if apiObject.ValueWhenUnset != nil {
		tfMap["values_when_unset"] = flattenStringValueWhenUnsetConfiguration(apiObject.ValueWhenUnset)
	}

	return []any{tfMap}
}

func flattenStringDefaultValues(apiObject *awstypes.StringDefaultValues) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.DynamicValue != nil {
		tfMap["dynamic_value"] = flattenDynamicDefaultValue(apiObject.DynamicValue)
	}
	if len(apiObject.StaticValues) > 0 {
		tfMap["static_values"] = apiObject.StaticValues
	}

	return []any{tfMap}
}

func flattenStringValueWhenUnsetConfiguration(apiObject *awstypes.StringValueWhenUnsetConfiguration) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.CustomValue != nil {
		tfMap["custom_value"] = aws.ToString(apiObject.CustomValue)
	}

	tfMap["value_when_unset_option"] = apiObject.ValueWhenUnsetOption

	return []any{tfMap}
}

func flattenParameterControls(apiObjects []awstypes.ParameterControl) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.DateTimePicker != nil {
			tfMap["date_time_picker"] = flattenParameterDateTimePickerControl(apiObject.DateTimePicker)
		}
		if apiObject.Dropdown != nil {
			tfMap["dropdown"] = flattenParameterDropDownControl(apiObject.Dropdown)
		}
		if apiObject.List != nil {
			tfMap["list"] = flattenParameterListControl(apiObject.List)
		}
		if apiObject.Slider != nil {
			tfMap["slider"] = flattenParameterSliderControl(apiObject.Slider)
		}
		if apiObject.TextArea != nil {
			tfMap["text_area"] = flattenParameterTextAreaControl(apiObject.TextArea)
		}
		if apiObject.TextField != nil {
			tfMap["text_field"] = flattenParameterTextFieldControl(apiObject.TextField)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenParameterDateTimePickerControl(apiObject *awstypes.ParameterDateTimePickerControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenDateTimePickerControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []any{tfMap}
}

func flattenParameterDropDownControl(apiObject *awstypes.ParameterDropDownControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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
	tfMap[names.AttrType] = apiObject.Type

	return []any{tfMap}
}

func flattenParameterSelectableValues(apiObject *awstypes.ParameterSelectableValues) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.LinkToDataSetColumn != nil {
		tfMap["link_to_data_set_column"] = flattenColumnIdentifier(apiObject.LinkToDataSetColumn)
	}
	if apiObject.Values != nil {
		tfMap[names.AttrValues] = apiObject.Values
	}

	return []any{tfMap}
}

func flattenParameterListControl(apiObject *awstypes.ParameterListControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
	}

	if apiObject.CascadingControlConfiguration != nil {
		tfMap["cascading_control_configuration"] = flattenCascadingControlConfiguration(apiObject.CascadingControlConfiguration)
	}
	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenListControlDisplayOptions(apiObject.DisplayOptions)
	}
	if apiObject.SelectableValues != nil {
		tfMap["selectable_values"] = flattenParameterSelectableValues(apiObject.SelectableValues)
	}
	tfMap[names.AttrType] = apiObject.Type

	return []any{tfMap}
}

func flattenParameterSliderControl(apiObject *awstypes.ParameterSliderControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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

	return []any{tfMap}
}

func flattenParameterTextAreaControl(apiObject *awstypes.ParameterTextAreaControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
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

	return []any{tfMap}
}

func flattenParameterTextFieldControl(apiObject *awstypes.ParameterTextFieldControl) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"parameter_control_id":  aws.ToString(apiObject.ParameterControlId),
		"source_parameter_name": aws.ToString(apiObject.SourceParameterName),
		"title":                 aws.ToString(apiObject.Title),
	}

	if apiObject.DisplayOptions != nil {
		tfMap["display_options"] = flattenTextFieldControlDisplayOptions(apiObject.DisplayOptions)
	}

	return []any{tfMap}
}
