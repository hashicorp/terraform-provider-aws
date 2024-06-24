// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func visualCustomActionsSchema(maxItems int) *schema.Schema {
	return &schema.Schema{ // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomAction.html
		Type:     schema.TypeList,
		MinItems: 1,
		MaxItems: maxItems,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"action_operations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_VisualCustomActionOperation.html
					Type:     schema.TypeList,
					MinItems: 1,
					MaxItems: 2,
					Required: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"filter_operation": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomActionFilterOperation.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"selected_fields_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterOperationSelectedFieldsConfiguration.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Required: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"selected_field_option": stringSchema(false, validation.StringInSlice(quicksight.SelectedFieldOptions_Values(), false)),
													"selected_fields": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 20,
														Elem: &schema.Schema{
															Type:         schema.TypeString,
															ValidateFunc: validation.StringLenBetween(1, 512),
														},
													},
												},
											},
										},
										"target_visuals_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_FilterOperationTargetVisualsConfiguration.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Required: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"same_sheet_target_visual_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SameSheetTargetVisualConfiguration.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Optional: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"target_visual_option": stringSchema(false, validation.StringInSlice(quicksight.TargetVisualOptions_Values(), false)),
																"target_visuals": {
																	Type:     schema.TypeSet,
																	Optional: true,
																	MinItems: 1,
																	MaxItems: 30,
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
							},
							"navigation_operation": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomActionNavigationOperation.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"local_navigation_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_LocalNavigationConfiguration.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 1,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"target_sheet_id": idSchema(),
												},
											},
										},
									},
								},
							},
							"set_parameters_operation": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomActionSetParametersOperation.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"parameter_value_configurations": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_SetParameterValueConfiguration.html
											Type:     schema.TypeList,
											MinItems: 1,
											MaxItems: 200,
											Required: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"destination_parameter_name": parameterNameSchema(true),
													names.AttrValue: { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DestinationParameterValueConfiguration.html
														Type:     schema.TypeList,
														MinItems: 1,
														MaxItems: 1,
														Required: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"custom_values_configuration": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomValuesConfiguration.html
																	Type:     schema.TypeList,
																	MinItems: 1,
																	MaxItems: 1,
																	Optional: true,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"custom_values": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomParameterValues.html
																				Type:     schema.TypeList,
																				MinItems: 1,
																				MaxItems: 1,
																				Required: true,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"date_time_values": {
																							Type:     schema.TypeList,
																							Optional: true,
																							MinItems: 1,
																							MaxItems: 50000,
																							Elem: &schema.Schema{
																								Type:         schema.TypeString,
																								ValidateFunc: verify.ValidUTCTimestamp,
																							},
																						},
																						"decimal_values": {
																							Type:     schema.TypeList,
																							Optional: true,
																							MinItems: 1,
																							MaxItems: 50000,
																							Elem: &schema.Schema{
																								Type: schema.TypeFloat,
																							},
																						},
																						"integer_values": {
																							Type:     schema.TypeList,
																							Optional: true,
																							MinItems: 1,
																							MaxItems: 50000,
																							Elem: &schema.Schema{
																								Type: schema.TypeInt,
																							},
																						},
																						"string_values": {
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
																			"include_null_value": {
																				Type:     schema.TypeBool,
																				Optional: true,
																			},
																		},
																	},
																},
																"select_all_value_options": stringSchema(false, validation.StringInSlice(quicksight.SelectAllValueOptions_Values(), false)),
																"source_field":             stringSchema(false, validation.StringLenBetween(1, 2048)),
																"source_parameter_name": {
																	Type:     schema.TypeString,
																	Optional: true,
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
							"url_operation": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_CustomActionURLOperation.html
								Type:     schema.TypeList,
								MinItems: 1,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"url_target":   stringSchema(true, validation.StringInSlice(quicksight.URLTargetConfiguration_Values(), false)),
										"url_template": stringSchema(true, validation.StringLenBetween(1, 2048)),
									},
								},
							},
						},
					},
				},
				"custom_action_id": idSchema(),
				names.AttrName:     stringSchema(true, validation.StringLenBetween(1, 256)),
				"trigger":          stringSchema(true, validation.StringInSlice(quicksight.VisualCustomActionTrigger_Values(), false)),
				names.AttrStatus:   stringSchema(true, validation.StringInSlice(quicksight.Status_Values(), false)),
			},
		},
	}
}

func expandVisualCustomActions(tfList []interface{}) []*quicksight.VisualCustomAction {
	if len(tfList) == 0 {
		return nil
	}

	var actions []*quicksight.VisualCustomAction
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		action := expandVisualCustomAction(tfMap)
		if action == nil {
			continue
		}

		actions = append(actions, action)
	}

	return actions
}

func expandVisualCustomAction(tfMap map[string]interface{}) *quicksight.VisualCustomAction {
	if tfMap == nil {
		return nil
	}

	action := &quicksight.VisualCustomAction{}

	if v, ok := tfMap["custom_action_id"].(string); ok && v != "" {
		action.CustomActionId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		action.Name = aws.String(v)
	}
	if v, ok := tfMap["trigger"].(string); ok && v != "" {
		action.Trigger = aws.String(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		action.Status = aws.String(v)
	}
	if v, ok := tfMap["action_operations"].([]interface{}); ok && len(v) > 0 {
		action.ActionOperations = expandVisualCustomActionOperations(v)
	}

	return action
}

func expandVisualCustomActionOperations(tfList []interface{}) []*quicksight.VisualCustomActionOperation {
	if len(tfList) == 0 {
		return nil
	}

	var actions []*quicksight.VisualCustomActionOperation
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		action := expandVisualCustomActionOperation(tfMap)
		if action == nil {
			continue
		}

		actions = append(actions, action)
	}

	return actions
}

func expandVisualCustomActionOperation(tfMap map[string]interface{}) *quicksight.VisualCustomActionOperation {
	if tfMap == nil {
		return nil
	}

	action := &quicksight.VisualCustomActionOperation{}

	if v, ok := tfMap["filter_operation"].([]interface{}); ok && len(v) > 0 {
		action.FilterOperation = expandCustomActionFilterOperation(v)
	}
	if v, ok := tfMap["navigation_operation"].([]interface{}); ok && len(v) > 0 {
		action.NavigationOperation = expandCustomActionNavigationOperation(v)
	}
	if v, ok := tfMap["set_parameters_operation"].([]interface{}); ok && len(v) > 0 {
		action.SetParametersOperation = expandCustomActionSetParametersOperation(v)
	}
	if v, ok := tfMap["url_operation"].([]interface{}); ok && len(v) > 0 {
		action.URLOperation = expandCustomActionURLOperation(v)
	}

	return action
}

func expandCustomActionFilterOperation(tfList []interface{}) *quicksight.CustomActionFilterOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	action := &quicksight.CustomActionFilterOperation{}

	if v, ok := tfMap["selected_fields_configuration"].([]interface{}); ok && len(v) > 0 {
		action.SelectedFieldsConfiguration = expandFilterOperationSelectedFieldsConfiguration(v)
	}
	if v, ok := tfMap["target_visuals_configuration"].([]interface{}); ok && len(v) > 0 {
		action.TargetVisualsConfiguration = expandFilterOperationTargetVisualsConfiguration(v)
	}

	return action
}

func expandFilterOperationSelectedFieldsConfiguration(tfList []interface{}) *quicksight.FilterOperationSelectedFieldsConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FilterOperationSelectedFieldsConfiguration{}

	if v, ok := tfMap["selected_field_option"].(string); ok && v != "" {
		config.SelectedFieldOptions = aws.String(v)
	}
	if v, ok := tfMap["selected_fields"].([]interface{}); ok {
		config.SelectedFields = flex.ExpandStringList(v)
	}

	return config
}

func expandFilterOperationTargetVisualsConfiguration(tfList []interface{}) *quicksight.FilterOperationTargetVisualsConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.FilterOperationTargetVisualsConfiguration{}

	if v, ok := tfMap["same_sheet_target_visual_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SameSheetTargetVisualConfiguration = expandSameSheetTargetVisualConfiguration(v)
	}

	return config
}

func expandSameSheetTargetVisualConfiguration(tfList []interface{}) *quicksight.SameSheetTargetVisualConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.SameSheetTargetVisualConfiguration{}

	if v, ok := tfMap["target_visual_option"].(string); ok && v != "" {
		config.TargetVisualOptions = aws.String(v)
	}
	if v, ok := tfMap["target_visuals"].(*schema.Set); ok {
		config.TargetVisuals = flex.ExpandStringSet(v)
	}

	return config
}

func expandCustomActionNavigationOperation(tfList []interface{}) *quicksight.CustomActionNavigationOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	action := &quicksight.CustomActionNavigationOperation{}

	if v, ok := tfMap["local_navigation_configuration"].([]interface{}); ok && len(v) > 0 {
		action.LocalNavigationConfiguration = expandLocalNavigationConfiguration(v)
	}

	return action
}

func expandLocalNavigationConfiguration(tfList []interface{}) *quicksight.LocalNavigationConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.LocalNavigationConfiguration{}

	if v, ok := tfMap["target_sheet_id"].(string); ok && v != "" {
		config.TargetSheetId = aws.String(v)
	}
	return config
}

func expandCustomActionSetParametersOperation(tfList []interface{}) *quicksight.CustomActionSetParametersOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	action := &quicksight.CustomActionSetParametersOperation{}

	if v, ok := tfMap["parameter_value_configurations"].([]interface{}); ok && len(v) > 0 {
		action.ParameterValueConfigurations = expandSetParameterValueConfigurations(v)
	}

	return action
}

func expandSetParameterValueConfigurations(tfList []interface{}) []*quicksight.SetParameterValueConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []*quicksight.SetParameterValueConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandSetParameterValueConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, config)
	}

	return configs
}

func expandSetParameterValueConfiguration(tfMap map[string]interface{}) *quicksight.SetParameterValueConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &quicksight.SetParameterValueConfiguration{}

	if v, ok := tfMap["destination_parameter_name"].(string); ok && v != "" {
		config.DestinationParameterName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		config.Value = expandDestinationParameterValueConfiguration(v)
	}

	return config
}

func expandDestinationParameterValueConfiguration(tfList []interface{}) *quicksight.DestinationParameterValueConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.DestinationParameterValueConfiguration{}

	if v, ok := tfMap["custom_values_configuration"].([]interface{}); ok && len(v) > 0 {
		config.CustomValuesConfiguration = expandCustomValuesConfiguration(v)
	}
	if v, ok := tfMap["select_all_value_options"].(string); ok && v != "" {
		config.SelectAllValueOptions = aws.String(v)
	}
	if v, ok := tfMap["source_field"].(string); ok && v != "" {
		config.SourceField = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		config.SourceParameterName = aws.String(v)
	}

	return config
}

func expandCustomValuesConfiguration(tfList []interface{}) *quicksight.CustomValuesConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.CustomValuesConfiguration{}

	if v, ok := tfMap["custom_values"].([]interface{}); ok && len(v) > 0 {
		config.CustomValues = expandCustomParameterValues(v)
	}
	if v, ok := tfMap["include_null_value"].(bool); ok {
		config.IncludeNullValue = aws.Bool(v)
	}

	return config
}

func expandCustomParameterValues(tfList []interface{}) *quicksight.CustomParameterValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &quicksight.CustomParameterValues{}

	if v, ok := tfMap["date_time_values"].([]interface{}); ok {
		config.DateTimeValues = flex.ExpandStringTimeList(v, time.RFC3339)
	}
	if v, ok := tfMap["decimal_values"].([]interface{}); ok {
		config.DecimalValues = flex.ExpandFloat64List(v)
	}
	if v, ok := tfMap["integer_values"].([]interface{}); ok {
		config.IntegerValues = flex.ExpandInt64List(v)
	}
	if v, ok := tfMap["string_values"].([]interface{}); ok {
		config.StringValues = flex.ExpandStringList(v)
	}

	return config
}

func expandCustomActionURLOperation(tfList []interface{}) *quicksight.CustomActionURLOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	action := &quicksight.CustomActionURLOperation{}

	if v, ok := tfMap["url_target"].(string); ok && v != "" {
		action.URLTarget = aws.String(v)
	}
	if v, ok := tfMap["url_template"].(string); ok && v != "" {
		action.URLTemplate = aws.String(v)
	}

	return action
}

func flattenVisualCustomAction(apiObject []*quicksight.VisualCustomAction) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"custom_action_id": aws.StringValue(config.CustomActionId),
			names.AttrName:     aws.StringValue(config.Name),
			names.AttrStatus:   aws.StringValue(config.Status),
			"trigger":          aws.StringValue(config.Trigger),
		}
		if config.ActionOperations != nil {
			tfMap["action_operations"] = flattenVisualCustomActionOperation(config.ActionOperations)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenVisualCustomActionOperation(apiObject []*quicksight.VisualCustomActionOperation) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{}
		if config.FilterOperation != nil {
			tfMap["filter_operation"] = flattenCustomActionFilterOperation(config.FilterOperation)
		}
		if config.NavigationOperation != nil {
			tfMap["navigation_operation"] = flattenCustomActionNavigationOperation(config.NavigationOperation)
		}
		if config.SetParametersOperation != nil {
			tfMap["set_parameters_operation"] = flattenCustomActionSetParametersOperation(config.SetParametersOperation)
		}
		if config.URLOperation != nil {
			tfMap["url_operation"] = flattenCustomActionURLOperation(config.URLOperation)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCustomActionFilterOperation(apiObject *quicksight.CustomActionFilterOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SelectedFieldsConfiguration != nil {
		tfMap["selected_fields_configuration"] = flattenFilterOperationSelectedFieldsConfiguration(apiObject.SelectedFieldsConfiguration)
	}
	if apiObject.TargetVisualsConfiguration != nil {
		tfMap["target_visuals_configuration"] = flattenFilterOperationTargetVisualsConfiguration(apiObject.TargetVisualsConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenFilterOperationSelectedFieldsConfiguration(apiObject *quicksight.FilterOperationSelectedFieldsConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SelectedFields != nil {
		tfMap["selected_fields"] = flex.FlattenStringList(apiObject.SelectedFields)
	}
	if apiObject.SelectedFieldOptions != nil {
		tfMap["selected_field_option"] = aws.StringValue(apiObject.SelectedFieldOptions)
	}

	return []interface{}{tfMap}
}

func flattenFilterOperationTargetVisualsConfiguration(apiObject *quicksight.FilterOperationTargetVisualsConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SameSheetTargetVisualConfiguration != nil {
		tfMap["same_sheet_target_visual_configuration"] = flattenSameSheetTargetVisualConfiguration(apiObject.SameSheetTargetVisualConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenSameSheetTargetVisualConfiguration(apiObject *quicksight.SameSheetTargetVisualConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.TargetVisualOptions != nil {
		tfMap["target_visual_option"] = aws.StringValue(apiObject.TargetVisualOptions)
	}
	if apiObject.TargetVisuals != nil {
		tfMap["target_visuals"] = flex.FlattenStringList(apiObject.TargetVisuals)
	}

	return []interface{}{tfMap}
}

func flattenCustomActionNavigationOperation(apiObject *quicksight.CustomActionNavigationOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LocalNavigationConfiguration != nil {
		tfMap["local_navigation_configuration"] = flattenLocalNavigationConfiguration(apiObject.LocalNavigationConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenLocalNavigationConfiguration(apiObject *quicksight.LocalNavigationConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"target_sheet_id": aws.StringValue(apiObject.TargetSheetId),
	}

	return []interface{}{tfMap}
}

func flattenCustomActionSetParametersOperation(apiObject *quicksight.CustomActionSetParametersOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_value_configurations": flattenSetParameterValueConfiguration(apiObject.ParameterValueConfigurations),
	}

	return []interface{}{tfMap}
}

func flattenSetParameterValueConfiguration(apiObject []*quicksight.SetParameterValueConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {
		if config == nil {
			continue
		}

		tfMap := map[string]interface{}{
			"destination_parameter_name": aws.StringValue(config.DestinationParameterName),
		}
		if config.Value != nil {
			tfMap[names.AttrValue] = flattenDestinationParameterValueConfiguration(config.Value)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDestinationParameterValueConfiguration(apiObject *quicksight.DestinationParameterValueConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomValuesConfiguration != nil {
		tfMap["custom_values_configuration"] = flattenCustomValuesConfiguration(apiObject.CustomValuesConfiguration)
	}
	if apiObject.SelectAllValueOptions != nil {
		tfMap["select_all_value_options"] = aws.StringValue(apiObject.SelectAllValueOptions)
	}
	if apiObject.SourceField != nil {
		tfMap["source_field"] = aws.StringValue(apiObject.SourceField)
	}
	if apiObject.SourceParameterName != nil {
		tfMap["source_parameter_name"] = aws.StringValue(apiObject.SourceParameterName)
	}

	return []interface{}{tfMap}
}

func flattenCustomValuesConfiguration(apiObject *quicksight.CustomValuesConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomValues != nil {
		tfMap["custom_values"] = flattenCustomParameterValues(apiObject.CustomValues)
	}
	if apiObject.IncludeNullValue != nil {
		tfMap["include_null_value"] = aws.BoolValue(apiObject.IncludeNullValue)
	}

	return []interface{}{tfMap}
}

func flattenCustomParameterValues(apiObject *quicksight.CustomParameterValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DateTimeValues != nil {
		tfMap["date_time_values"] = flex.FlattenTimeStringList(apiObject.DateTimeValues, time.RFC3339)
	}
	if apiObject.DecimalValues != nil {
		tfMap["decimal_values"] = flex.FlattenFloat64List(apiObject.DecimalValues)
	}
	if apiObject.IntegerValues != nil {
		tfMap["integer_values"] = flex.FlattenInt64List(apiObject.IntegerValues)
	}
	if apiObject.StringValues != nil {
		tfMap["string_values"] = flex.FlattenStringList(apiObject.StringValues)
	}

	return []interface{}{tfMap}
}

func flattenCustomActionURLOperation(apiObject *quicksight.CustomActionURLOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"url_target":   aws.StringValue(apiObject.URLTarget),
		"url_template": aws.StringValue(apiObject.URLTemplate),
	}

	return []interface{}{tfMap}
}
