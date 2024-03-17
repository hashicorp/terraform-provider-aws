// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
													"selected_field_option": stringSchema(false, enum.Validate[types.SelectedFieldOptions]()),
													"selected_fields": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 1,
														MaxItems: 20,
														Elem: &schema.Schema{
															Type:             schema.TypeString,
															ValidateDiagFunc: validation.ToDiagFunc(validation.StringLenBetween(1, 512)),
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
																"target_visual_option": stringSchema(false, enum.Validate[types.TargetVisualOptions]()),
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
													"value": { // https://docs.aws.amazon.com/quicksight/latest/APIReference/API_DestinationParameterValueConfiguration.html
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
																								Type:             schema.TypeString,
																								ValidateDiagFunc: validation.ToDiagFunc(verify.ValidUTCTimestamp),
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
																"select_all_value_options": stringSchema(false, enum.Validate[types.SelectAllValueOptions]()),
																"source_field":             stringSchema(false, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
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
										"url_target":   stringSchema(true, enum.Validate[types.URLTargetConfiguration]()),
										"url_template": stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 2048))),
									},
								},
							},
						},
					},
				},
				"custom_action_id": idSchema(),
				"name":             stringSchema(true, validation.ToDiagFunc(validation.StringLenBetween(1, 256))),
				"trigger":          stringSchema(true, enum.Validate[types.VisualCustomActionTrigger]()),
				"status":           stringSchema(true, enum.Validate[types.Status]()),
			},
		},
	}
}

func expandVisualCustomActions(tfList []interface{}) []types.VisualCustomAction {
	if len(tfList) == 0 {
		return nil
	}

	var actions []types.VisualCustomAction
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		action := expandVisualCustomAction(tfMap)
		if action == nil {
			continue
		}

		actions = append(actions, *action)
	}

	return actions
}

func expandVisualCustomAction(tfMap map[string]interface{}) *types.VisualCustomAction {
	if tfMap == nil {
		return nil
	}

	action := &types.VisualCustomAction{}

	if v, ok := tfMap["custom_action_id"].(string); ok && v != "" {
		action.CustomActionId = aws.String(v)
	}
	if v, ok := tfMap["name"].(string); ok && v != "" {
		action.Name = aws.String(v)
	}
	if v, ok := tfMap["trigger"].(string); ok && v != "" {
		action.Trigger = types.VisualCustomActionTrigger(v)
	}
	if v, ok := tfMap["status"].(string); ok && v != "" {
		action.Status = types.WidgetStatus(v)
	}
	if v, ok := tfMap["action_operations"].([]interface{}); ok && len(v) > 0 {
		action.ActionOperations = expandVisualCustomActionOperations(v)
	}

	return action
}

func expandVisualCustomActionOperations(tfList []interface{}) []types.VisualCustomActionOperation {
	if len(tfList) == 0 {
		return nil
	}

	var actions []types.VisualCustomActionOperation
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		action := expandVisualCustomActionOperation(tfMap)
		if action == nil {
			continue
		}

		actions = append(actions, *action)
	}

	return actions
}

func expandVisualCustomActionOperation(tfMap map[string]interface{}) *types.VisualCustomActionOperation {
	if tfMap == nil {
		return nil
	}

	action := &types.VisualCustomActionOperation{}

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

func expandCustomActionFilterOperation(tfList []interface{}) *types.CustomActionFilterOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	action := &types.CustomActionFilterOperation{}

	if v, ok := tfMap["selected_fields_configuration"].([]interface{}); ok && len(v) > 0 {
		action.SelectedFieldsConfiguration = expandFilterOperationSelectedFieldsConfiguration(v)
	}
	if v, ok := tfMap["target_visuals_configuration"].([]interface{}); ok && len(v) > 0 {
		action.TargetVisualsConfiguration = expandFilterOperationTargetVisualsConfiguration(v)
	}

	return action
}

func expandFilterOperationSelectedFieldsConfiguration(tfList []interface{}) *types.FilterOperationSelectedFieldsConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilterOperationSelectedFieldsConfiguration{}

	if v, ok := tfMap["selected_field_option"].(string); ok && v != "" {
		config.SelectedFieldOptions = types.SelectedFieldOptions(v)
	}
	if v, ok := tfMap["selected_fields"].([]interface{}); ok {
		config.SelectedFields = flex.ExpandStringValueList(v)
	}

	return config
}

func expandFilterOperationTargetVisualsConfiguration(tfList []interface{}) *types.FilterOperationTargetVisualsConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.FilterOperationTargetVisualsConfiguration{}

	if v, ok := tfMap["same_sheet_target_visual_configuration"].([]interface{}); ok && len(v) > 0 {
		config.SameSheetTargetVisualConfiguration = expandSameSheetTargetVisualConfiguration(v)
	}

	return config
}

func expandSameSheetTargetVisualConfiguration(tfList []interface{}) *types.SameSheetTargetVisualConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.SameSheetTargetVisualConfiguration{}

	if v, ok := tfMap["target_visual_option"].(string); ok && v != "" {
		config.TargetVisualOptions = types.TargetVisualOptions(v)
	}
	if v, ok := tfMap["target_visuals"].(*schema.Set); ok {
		config.TargetVisuals = flex.ExpandStringValueSet(v)
	}

	return config
}

func expandCustomActionNavigationOperation(tfList []interface{}) *types.CustomActionNavigationOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	action := &types.CustomActionNavigationOperation{}

	if v, ok := tfMap["local_navigation_configuration"].([]interface{}); ok && len(v) > 0 {
		action.LocalNavigationConfiguration = expandLocalNavigationConfiguration(v)
	}

	return action
}

func expandLocalNavigationConfiguration(tfList []interface{}) *types.LocalNavigationConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.LocalNavigationConfiguration{}

	if v, ok := tfMap["target_sheet_id"].(string); ok && v != "" {
		config.TargetSheetId = aws.String(v)
	}
	return config
}

func expandCustomActionSetParametersOperation(tfList []interface{}) *types.CustomActionSetParametersOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	action := &types.CustomActionSetParametersOperation{}

	if v, ok := tfMap["parameter_value_configurations"].([]interface{}); ok && len(v) > 0 {
		action.ParameterValueConfigurations = expandSetParameterValueConfigurations(v)
	}

	return action
}

func expandSetParameterValueConfigurations(tfList []interface{}) []types.SetParameterValueConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var configs []types.SetParameterValueConfiguration
	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		config := expandSetParameterValueConfiguration(tfMap)
		if config == nil {
			continue
		}

		configs = append(configs, *config)
	}

	return configs
}

func expandSetParameterValueConfiguration(tfMap map[string]interface{}) *types.SetParameterValueConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &types.SetParameterValueConfiguration{}

	if v, ok := tfMap["destination_parameter_name"].(string); ok && v != "" {
		config.DestinationParameterName = aws.String(v)
	}
	if v, ok := tfMap["value"].([]interface{}); ok && len(v) > 0 {
		config.Value = expandDestinationParameterValueConfiguration(v)
	}

	return config
}

func expandDestinationParameterValueConfiguration(tfList []interface{}) *types.DestinationParameterValueConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.DestinationParameterValueConfiguration{}

	if v, ok := tfMap["custom_values_configuration"].([]interface{}); ok && len(v) > 0 {
		config.CustomValuesConfiguration = expandCustomValuesConfiguration(v)
	}
	if v, ok := tfMap["select_all_value_options"].(string); ok && v != "" {
		config.SelectAllValueOptions = types.SelectAllValueOptions(v)
	}
	if v, ok := tfMap["source_field"].(string); ok && v != "" {
		config.SourceField = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		config.SourceParameterName = aws.String(v)
	}

	return config
}

func expandCustomValuesConfiguration(tfList []interface{}) *types.CustomValuesConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.CustomValuesConfiguration{}

	if v, ok := tfMap["custom_values"].([]interface{}); ok && len(v) > 0 {
		config.CustomValues = expandCustomParameterValues(v)
	}
	if v, ok := tfMap["include_null_value"].(bool); ok {
		config.IncludeNullValue = aws.Bool(v)
	}

	return config
}

func expandCustomParameterValues(tfList []interface{}) *types.CustomParameterValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	config := &types.CustomParameterValues{}

	if v, ok := tfMap["date_time_values"].([]interface{}); ok {
		config.DateTimeValues = flex.ExpandStringTimeValueList(v, time.RFC3339)
	}
	if v, ok := tfMap["decimal_values"].([]interface{}); ok {
		config.DecimalValues = flex.ExpandFloat64ValueList(v)
	}
	if v, ok := tfMap["integer_values"].([]interface{}); ok {
		config.IntegerValues = flex.ExpandInt64ValueList(v)
	}
	if v, ok := tfMap["string_values"].([]interface{}); ok {
		config.StringValues = flex.ExpandStringValueList(v)
	}

	return config
}

func expandCustomActionURLOperation(tfList []interface{}) *types.CustomActionURLOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	action := &types.CustomActionURLOperation{}

	if v, ok := tfMap["url_target"].(string); ok && v != "" {
		action.URLTarget = types.URLTargetConfiguration(v)
	}
	if v, ok := tfMap["url_template"].(string); ok && v != "" {
		action.URLTemplate = aws.String(v)
	}

	return action
}

func flattenVisualCustomAction(apiObject []types.VisualCustomAction) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{
			"custom_action_id": aws.ToString(config.CustomActionId),
			"name":             aws.ToString(config.Name),
			"status":           types.WidgetStatus(config.Status),
			"trigger":          types.VisualCustomActionTrigger(config.Trigger),
		}
		if config.ActionOperations != nil {
			tfMap["action_operations"] = flattenVisualCustomActionOperation(config.ActionOperations)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenVisualCustomActionOperation(apiObject []types.VisualCustomActionOperation) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

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

func flattenCustomActionFilterOperation(apiObject *types.CustomActionFilterOperation) []interface{} {
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

func flattenFilterOperationSelectedFieldsConfiguration(apiObject *types.FilterOperationSelectedFieldsConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SelectedFields != nil {
		tfMap["selected_fields"] = flex.FlattenStringValueList(apiObject.SelectedFields)
	}
	tfMap["selected_field_option"] = types.SelectedFieldOptions(apiObject.SelectedFieldOptions)

	return []interface{}{tfMap}
}

func flattenFilterOperationTargetVisualsConfiguration(apiObject *types.FilterOperationTargetVisualsConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.SameSheetTargetVisualConfiguration != nil {
		tfMap["same_sheet_target_visual_configuration"] = flattenSameSheetTargetVisualConfiguration(apiObject.SameSheetTargetVisualConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenSameSheetTargetVisualConfiguration(apiObject *types.SameSheetTargetVisualConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["target_visual_option"] = types.TargetVisualOptions(apiObject.TargetVisualOptions)
	if apiObject.TargetVisuals != nil {
		tfMap["target_visuals"] = flex.FlattenStringValueList(apiObject.TargetVisuals)
	}

	return []interface{}{tfMap}
}

func flattenCustomActionNavigationOperation(apiObject *types.CustomActionNavigationOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.LocalNavigationConfiguration != nil {
		tfMap["local_navigation_configuration"] = flattenLocalNavigationConfiguration(apiObject.LocalNavigationConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenLocalNavigationConfiguration(apiObject *types.LocalNavigationConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"target_sheet_id": aws.ToString(apiObject.TargetSheetId),
	}

	return []interface{}{tfMap}
}

func flattenCustomActionSetParametersOperation(apiObject *types.CustomActionSetParametersOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_value_configurations": flattenSetParameterValueConfiguration(apiObject.ParameterValueConfigurations),
	}

	return []interface{}{tfMap}
}

func flattenSetParameterValueConfiguration(apiObject []types.SetParameterValueConfiguration) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}
	for _, config := range apiObject {

		tfMap := map[string]interface{}{
			"destination_parameter_name": aws.ToString(config.DestinationParameterName),
		}
		if config.Value != nil {
			tfMap["value"] = flattenDestinationParameterValueConfiguration(config.Value)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDestinationParameterValueConfiguration(apiObject *types.DestinationParameterValueConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomValuesConfiguration != nil {
		tfMap["custom_values_configuration"] = flattenCustomValuesConfiguration(apiObject.CustomValuesConfiguration)
	}
	tfMap["select_all_value_options"] = types.SelectAllValueOptions(apiObject.SelectAllValueOptions)
	if apiObject.SourceField != nil {
		tfMap["source_field"] = aws.ToString(apiObject.SourceField)
	}
	if apiObject.SourceParameterName != nil {
		tfMap["source_parameter_name"] = aws.ToString(apiObject.SourceParameterName)
	}

	return []interface{}{tfMap}
}

func flattenCustomValuesConfiguration(apiObject *types.CustomValuesConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.CustomValues != nil {
		tfMap["custom_values"] = flattenCustomParameterValues(apiObject.CustomValues)
	}
	if apiObject.IncludeNullValue != nil {
		tfMap["include_null_value"] = aws.ToBool(apiObject.IncludeNullValue)
	}

	return []interface{}{tfMap}
}

func flattenCustomParameterValues(apiObject *types.CustomParameterValues) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if apiObject.DateTimeValues != nil {
		tfMap["date_time_values"] = flex.FlattenTimeStringValueList(apiObject.DateTimeValues, time.RFC3339)
	}
	if apiObject.DecimalValues != nil {
		tfMap["decimal_values"] = apiObject.DecimalValues
	}
	if apiObject.IntegerValues != nil {
		tfMap["integer_values"] = apiObject.IntegerValues
	}
	if apiObject.StringValues != nil {
		tfMap["string_values"] = apiObject.StringValues
	}

	return []interface{}{tfMap}
}

func flattenCustomActionURLOperation(apiObject *types.CustomActionURLOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"url_target":   types.URLTargetConfiguration(apiObject.URLTarget),
		"url_template": aws.ToString(apiObject.URLTemplate),
	}

	return []interface{}{tfMap}
}
