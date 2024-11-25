// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
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
													"selected_field_option": stringEnumSchema[awstypes.SelectedFieldOptions](attrOptional),
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
																"target_visual_option": stringEnumSchema[awstypes.TargetVisualOptions](attrOptional),
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
																"select_all_value_options": stringEnumSchema[awstypes.SelectAllValueOptions](attrOptional),
																"source_field":             stringLenBetweenSchema(attrOptional, 1, 2048),
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
										"url_target":   stringEnumSchema[awstypes.URLTargetConfiguration](attrRequired),
										"url_template": stringLenBetweenSchema(attrRequired, 1, 2048),
									},
								},
							},
						},
					},
				},
				"custom_action_id": idSchema(),
				names.AttrName:     stringLenBetweenSchema(attrRequired, 1, 256),
				"trigger":          stringEnumSchema[awstypes.VisualCustomActionTrigger](attrRequired),
				names.AttrStatus:   stringEnumSchema[awstypes.Status](attrRequired),
			},
		},
	}
}

func expandVisualCustomActions(tfList []interface{}) []awstypes.VisualCustomAction {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.VisualCustomAction

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandVisualCustomAction(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandVisualCustomAction(tfMap map[string]interface{}) *awstypes.VisualCustomAction {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VisualCustomAction{}

	if v, ok := tfMap["custom_action_id"].(string); ok && v != "" {
		apiObject.CustomActionId = aws.String(v)
	}
	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}
	if v, ok := tfMap["trigger"].(string); ok && v != "" {
		apiObject.Trigger = awstypes.VisualCustomActionTrigger(v)
	}
	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		apiObject.Status = awstypes.WidgetStatus(v)
	}
	if v, ok := tfMap["action_operations"].([]interface{}); ok && len(v) > 0 {
		apiObject.ActionOperations = expandVisualCustomActionOperations(v)
	}

	return apiObject
}

func expandVisualCustomActionOperations(tfList []interface{}) []awstypes.VisualCustomActionOperation {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.VisualCustomActionOperation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandVisualCustomActionOperation(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandVisualCustomActionOperation(tfMap map[string]interface{}) *awstypes.VisualCustomActionOperation {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VisualCustomActionOperation{}

	if v, ok := tfMap["filter_operation"].([]interface{}); ok && len(v) > 0 {
		apiObject.FilterOperation = expandCustomActionFilterOperation(v)
	}
	if v, ok := tfMap["navigation_operation"].([]interface{}); ok && len(v) > 0 {
		apiObject.NavigationOperation = expandCustomActionNavigationOperation(v)
	}
	if v, ok := tfMap["set_parameters_operation"].([]interface{}); ok && len(v) > 0 {
		apiObject.SetParametersOperation = expandCustomActionSetParametersOperation(v)
	}
	if v, ok := tfMap["url_operation"].([]interface{}); ok && len(v) > 0 {
		apiObject.URLOperation = expandCustomActionURLOperation(v)
	}

	return apiObject
}

func expandCustomActionFilterOperation(tfList []interface{}) *awstypes.CustomActionFilterOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomActionFilterOperation{}

	if v, ok := tfMap["selected_fields_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.SelectedFieldsConfiguration = expandFilterOperationSelectedFieldsConfiguration(v)
	}
	if v, ok := tfMap["target_visuals_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.TargetVisualsConfiguration = expandFilterOperationTargetVisualsConfiguration(v)
	}

	return apiObject
}

func expandFilterOperationSelectedFieldsConfiguration(tfList []interface{}) *awstypes.FilterOperationSelectedFieldsConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterOperationSelectedFieldsConfiguration{}

	if v, ok := tfMap["selected_field_option"].(string); ok && v != "" {
		apiObject.SelectedFieldOptions = awstypes.SelectedFieldOptions(v)
	}
	if v, ok := tfMap["selected_fields"].([]interface{}); ok {
		apiObject.SelectedFields = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandFilterOperationTargetVisualsConfiguration(tfList []interface{}) *awstypes.FilterOperationTargetVisualsConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.FilterOperationTargetVisualsConfiguration{}

	if v, ok := tfMap["same_sheet_target_visual_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.SameSheetTargetVisualConfiguration = expandSameSheetTargetVisualConfiguration(v)
	}

	return apiObject
}

func expandSameSheetTargetVisualConfiguration(tfList []interface{}) *awstypes.SameSheetTargetVisualConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.SameSheetTargetVisualConfiguration{}

	if v, ok := tfMap["target_visual_option"].(string); ok && v != "" {
		apiObject.TargetVisualOptions = awstypes.TargetVisualOptions(v)
	}
	if v, ok := tfMap["target_visuals"].(*schema.Set); ok {
		apiObject.TargetVisuals = flex.ExpandStringValueSet(v)
	}

	return apiObject
}

func expandCustomActionNavigationOperation(tfList []interface{}) *awstypes.CustomActionNavigationOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomActionNavigationOperation{}

	if v, ok := tfMap["local_navigation_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.LocalNavigationConfiguration = expandLocalNavigationConfiguration(v)
	}

	return apiObject
}

func expandLocalNavigationConfiguration(tfList []interface{}) *awstypes.LocalNavigationConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.LocalNavigationConfiguration{}

	if v, ok := tfMap["target_sheet_id"].(string); ok && v != "" {
		apiObject.TargetSheetId = aws.String(v)
	}
	return apiObject
}

func expandCustomActionSetParametersOperation(tfList []interface{}) *awstypes.CustomActionSetParametersOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomActionSetParametersOperation{}

	if v, ok := tfMap["parameter_value_configurations"].([]interface{}); ok && len(v) > 0 {
		apiObject.ParameterValueConfigurations = expandSetParameterValueConfigurations(v)
	}

	return apiObject
}

func expandSetParameterValueConfigurations(tfList []interface{}) []awstypes.SetParameterValueConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.SetParameterValueConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := expandSetParameterValueConfiguration(tfMap)
		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandSetParameterValueConfiguration(tfMap map[string]interface{}) *awstypes.SetParameterValueConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SetParameterValueConfiguration{}

	if v, ok := tfMap["destination_parameter_name"].(string); ok && v != "" {
		apiObject.DestinationParameterName = aws.String(v)
	}
	if v, ok := tfMap[names.AttrValue].([]interface{}); ok && len(v) > 0 {
		apiObject.Value = expandDestinationParameterValueConfiguration(v)
	}

	return apiObject
}

func expandDestinationParameterValueConfiguration(tfList []interface{}) *awstypes.DestinationParameterValueConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.DestinationParameterValueConfiguration{}

	if v, ok := tfMap["custom_values_configuration"].([]interface{}); ok && len(v) > 0 {
		apiObject.CustomValuesConfiguration = expandCustomValuesConfiguration(v)
	}
	if v, ok := tfMap["select_all_value_options"].(string); ok && v != "" {
		apiObject.SelectAllValueOptions = awstypes.SelectAllValueOptions(v)
	}
	if v, ok := tfMap["source_field"].(string); ok && v != "" {
		apiObject.SourceField = aws.String(v)
	}
	if v, ok := tfMap["source_parameter_name"].(string); ok && v != "" {
		apiObject.SourceParameterName = aws.String(v)
	}

	return apiObject
}

func expandCustomValuesConfiguration(tfList []interface{}) *awstypes.CustomValuesConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomValuesConfiguration{}

	if v, ok := tfMap["custom_values"].([]interface{}); ok && len(v) > 0 {
		apiObject.CustomValues = expandCustomParameterValues(v)
	}
	if v, ok := tfMap["include_null_value"].(bool); ok {
		apiObject.IncludeNullValue = aws.Bool(v)
	}

	return apiObject
}

func expandCustomParameterValues(tfList []interface{}) *awstypes.CustomParameterValues {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomParameterValues{}

	if v, ok := tfMap["date_time_values"].([]interface{}); ok {
		apiObject.DateTimeValues = flex.ExpandStringTimeValueList(v, time.RFC3339)
	}
	if v, ok := tfMap["decimal_values"].([]interface{}); ok {
		apiObject.DecimalValues = flex.ExpandFloat64ValueList(v)
	}
	if v, ok := tfMap["integer_values"].([]interface{}); ok {
		apiObject.IntegerValues = flex.ExpandInt64ValueList(v)
	}
	if v, ok := tfMap["string_values"].([]interface{}); ok {
		apiObject.StringValues = flex.ExpandStringValueList(v)
	}

	return apiObject
}

func expandCustomActionURLOperation(tfList []interface{}) *awstypes.CustomActionURLOperation {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.CustomActionURLOperation{}

	if v, ok := tfMap["url_target"].(string); ok && v != "" {
		apiObject.URLTarget = awstypes.URLTargetConfiguration(v)
	}
	if v, ok := tfMap["url_template"].(string); ok && v != "" {
		apiObject.URLTemplate = aws.String(v)
	}

	return apiObject
}

func flattenVisualCustomAction(apiObjects []awstypes.VisualCustomAction) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"custom_action_id": aws.ToString(apiObject.CustomActionId),
			names.AttrName:     aws.ToString(apiObject.Name),
			names.AttrStatus:   apiObject.Status,
			"trigger":          apiObject.Trigger,
		}

		if apiObject.ActionOperations != nil {
			tfMap["action_operations"] = flattenVisualCustomActionOperation(apiObject.ActionOperations)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenVisualCustomActionOperation(apiObjects []awstypes.VisualCustomActionOperation) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{}

		if apiObject.FilterOperation != nil {
			tfMap["filter_operation"] = flattenCustomActionFilterOperation(apiObject.FilterOperation)
		}
		if apiObject.NavigationOperation != nil {
			tfMap["navigation_operation"] = flattenCustomActionNavigationOperation(apiObject.NavigationOperation)
		}
		if apiObject.SetParametersOperation != nil {
			tfMap["set_parameters_operation"] = flattenCustomActionSetParametersOperation(apiObject.SetParametersOperation)
		}
		if apiObject.URLOperation != nil {
			tfMap["url_operation"] = flattenCustomActionURLOperation(apiObject.URLOperation)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenCustomActionFilterOperation(apiObject *awstypes.CustomActionFilterOperation) []interface{} {
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

func flattenFilterOperationSelectedFieldsConfiguration(apiObject *awstypes.FilterOperationSelectedFieldsConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SelectedFields != nil {
		tfMap["selected_fields"] = apiObject.SelectedFields
	}
	tfMap["selected_field_option"] = apiObject.SelectedFieldOptions

	return []interface{}{tfMap}
}

func flattenFilterOperationTargetVisualsConfiguration(apiObject *awstypes.FilterOperationTargetVisualsConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.SameSheetTargetVisualConfiguration != nil {
		tfMap["same_sheet_target_visual_configuration"] = flattenSameSheetTargetVisualConfiguration(apiObject.SameSheetTargetVisualConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenSameSheetTargetVisualConfiguration(apiObject *awstypes.SameSheetTargetVisualConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"target_visual_option": apiObject.TargetVisualOptions,
		"target_visuals":       apiObject.TargetVisuals,
	}

	return []interface{}{tfMap}
}

func flattenCustomActionNavigationOperation(apiObject *awstypes.CustomActionNavigationOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.LocalNavigationConfiguration != nil {
		tfMap["local_navigation_configuration"] = flattenLocalNavigationConfiguration(apiObject.LocalNavigationConfiguration)
	}

	return []interface{}{tfMap}
}

func flattenLocalNavigationConfiguration(apiObject *awstypes.LocalNavigationConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"target_sheet_id": aws.ToString(apiObject.TargetSheetId),
	}

	return []interface{}{tfMap}
}

func flattenCustomActionSetParametersOperation(apiObject *awstypes.CustomActionSetParametersOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"parameter_value_configurations": flattenSetParameterValueConfiguration(apiObject.ParameterValueConfigurations),
	}

	return []interface{}{tfMap}
}

func flattenSetParameterValueConfiguration(apiObjects []awstypes.SetParameterValueConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"destination_parameter_name": aws.ToString(apiObject.DestinationParameterName),
		}

		if apiObject.Value != nil {
			tfMap[names.AttrValue] = flattenDestinationParameterValueConfiguration(apiObject.Value)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDestinationParameterValueConfiguration(apiObject *awstypes.DestinationParameterValueConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if apiObject.CustomValuesConfiguration != nil {
		tfMap["custom_values_configuration"] = flattenCustomValuesConfiguration(apiObject.CustomValuesConfiguration)
	}
	tfMap["select_all_value_options"] = apiObject.SelectAllValueOptions
	if apiObject.SourceField != nil {
		tfMap["source_field"] = aws.ToString(apiObject.SourceField)
	}
	if apiObject.SourceParameterName != nil {
		tfMap["source_parameter_name"] = aws.ToString(apiObject.SourceParameterName)
	}

	return []interface{}{tfMap}
}

func flattenCustomValuesConfiguration(apiObject *awstypes.CustomValuesConfiguration) []interface{} {
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

func flattenCustomParameterValues(apiObject *awstypes.CustomParameterValues) []interface{} {
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

func flattenCustomActionURLOperation(apiObject *awstypes.CustomActionURLOperation) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"url_target":   apiObject.URLTarget,
		"url_template": aws.ToString(apiObject.URLTemplate),
	}

	return []interface{}{tfMap}
}
