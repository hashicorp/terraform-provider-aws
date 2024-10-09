// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandRegions(regions []interface{}) map[string]types.RegionMapInputValue {
	if len(regions) == 0 {
		return nil
	}

	regionMap := make(map[string]types.RegionMapInputValue)
	for _, region := range regions {
		regionData := region.(map[string]interface{})

		input := types.RegionMapInputValue{}

		if kmsKey := regionData[names.AttrKMSKeyARN].(string); kmsKey != "DefaultKey" {
			input.SseKmsKeyId = aws.String(kmsKey)
		}

		regionMap[regionData[names.AttrName].(string)] = input
	}

	return regionMap
}

func flattenRegions(regions map[string]types.RegionInfo) []map[string]interface{} {
	if len(regions) == 0 {
		return nil
	}

	tfRegionData := make([]map[string]interface{}, 0)
	for regionName, regionData := range regions {
		region := make(map[string]interface{})

		region[names.AttrName] = regionName
		region[names.AttrStatus] = regionData.Status
		region[names.AttrKMSKeyARN] = aws.ToString(regionData.SseKmsKeyId)

		if v := regionData.StatusMessage; v != nil {
			region[names.AttrStatusMessage] = aws.ToString(v)
		}

		tfRegionData = append(tfRegionData, region)
	}

	return tfRegionData
}

func expandIncidentTemplate(config []interface{}) *types.IncidentTemplate {
	// we require exactly one item so we grab first in list
	templateConfig := config[0].(map[string]interface{})

	template := &types.IncidentTemplate{}

	if v, ok := templateConfig["title"].(string); ok && v != "" {
		template.Title = aws.String(v)
	}

	if v, ok := templateConfig["impact"].(int); ok && v != 0 {
		template.Impact = aws.Int32(int32(v))
	}

	// dedupe string can be updated to have no value (denoted as "")
	if v, ok := templateConfig["dedupe_string"].(string); ok {
		template.DedupeString = aws.String(v)
	}

	if v, ok := templateConfig["incident_tags"].(map[string]interface{}); ok && len(v) > 0 {
		template.IncidentTags = flex.ExpandStringValueMap(v)
	}

	// summary can be updated to have no value (denoted as "")
	if v, ok := templateConfig["summary"].(string); ok {
		template.Summary = aws.String(v)
	}

	if v, ok := templateConfig["notification_target"].(*schema.Set); ok && v.Len() > 0 {
		template.NotificationTargets = expandNotificationTargets(v.List())
	}

	return template
}

func flattenIncidentTemplate(template *types.IncidentTemplate) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	tfTemplate := make(map[string]interface{})

	tfTemplate["impact"] = aws.ToInt32(template.Impact)
	tfTemplate["title"] = aws.ToString(template.Title)

	if v := template.DedupeString; v != nil {
		tfTemplate["dedupe_string"] = aws.ToString(v)
	}

	if v := template.IncidentTags; v != nil {
		tfTemplate["incident_tags"] = template.IncidentTags
	}

	if v := template.Summary; v != nil {
		tfTemplate["summary"] = aws.ToString(template.Summary)
	}

	if v := template.NotificationTargets; v != nil {
		tfTemplate["notification_target"] = flattenNotificationTargets(template.NotificationTargets)
	}

	result = append(result, tfTemplate)
	return result
}

func expandNotificationTargets(targets []interface{}) []types.NotificationTargetItem {
	if len(targets) == 0 {
		return nil
	}

	notificationTargets := make([]types.NotificationTargetItem, len(targets))

	for i, target := range targets {
		targetData := target.(map[string]interface{})

		targetItem := &types.NotificationTargetItemMemberSnsTopicArn{
			Value: targetData[names.AttrSNSTopicARN].(string),
		}

		notificationTargets[i] = targetItem
	}

	return notificationTargets
}

func flattenNotificationTargets(targets []types.NotificationTargetItem) []map[string]interface{} {
	if len(targets) == 0 {
		return nil
	}

	notificationTargets := make([]map[string]interface{}, len(targets))

	for i, target := range targets {
		targetItem := make(map[string]interface{})

		targetItem[names.AttrSNSTopicARN] = target.(*types.NotificationTargetItemMemberSnsTopicArn).Value

		notificationTargets[i] = targetItem
	}

	return notificationTargets
}

func expandChatChannel(chatChannels *schema.Set) types.ChatChannel {
	chatChannelList := flex.ExpandStringValueSet(chatChannels)

	if len(chatChannelList) == 0 {
		return &types.ChatChannelMemberEmpty{
			Value: types.EmptyChatChannel{},
		}
	}

	return &types.ChatChannelMemberChatbotSns{
		Value: chatChannelList,
	}
}

func flattenChatChannel(chatChannel types.ChatChannel) *schema.Set {
	if _, ok := chatChannel.(*types.ChatChannelMemberEmpty); ok {
		return &schema.Set{}
	}

	if chatBotSns, ok := chatChannel.(*types.ChatChannelMemberChatbotSns); ok {
		return flex.FlattenStringValueSet(chatBotSns.Value)
	}
	return nil
}

func expandAction(actions []interface{}) []types.Action {
	if len(actions) == 0 {
		return nil
	}

	result := make([]types.Action, 0)

	actionConfig := actions[0].(map[string]interface{})
	if v, ok := actionConfig["ssm_automation"].([]interface{}); ok {
		result = append(result, expandSSMAutomations(v)...)
	}

	return result
}

func flattenAction(actions []types.Action) []interface{} {
	if len(actions) == 0 {
		return nil
	}

	result := make([]interface{}, 0)

	action := make(map[string]interface{})
	action["ssm_automation"] = flattenSSMAutomations(actions)
	result = append(result, action)

	return result
}

func expandSSMAutomations(automations []interface{}) []types.Action {
	var result []types.Action
	for _, automation := range automations {
		ssmAutomation := types.SsmAutomation{}
		automationData := automation.(map[string]interface{})

		if v, ok := automationData["document_name"].(string); ok {
			ssmAutomation.DocumentName = aws.String(v)
		}

		if v, ok := automationData[names.AttrRoleARN].(string); ok {
			ssmAutomation.RoleArn = aws.String(v)
		}

		if v, ok := automationData["document_version"].(string); ok {
			ssmAutomation.DocumentVersion = aws.String(v)
		}

		if v, ok := automationData["target_account"].(string); ok {
			ssmAutomation.TargetAccount = types.SsmTargetAccount(v)
		}

		if v, ok := automationData[names.AttrParameter].(*schema.Set); ok {
			ssmAutomation.Parameters = expandParameters(v)
		}

		if v, ok := automationData["dynamic_parameters"].(map[string]interface{}); ok {
			ssmAutomation.DynamicParameters = expandDynamicParameters(v)
		}

		result = append(
			result,
			&types.ActionMemberSsmAutomation{Value: ssmAutomation},
		)
	}
	return result
}

func flattenSSMAutomations(actions []types.Action) []interface{} {
	var result []interface{}

	for _, action := range actions {
		if ssmAutomationAction, ok := action.(*types.ActionMemberSsmAutomation); ok {
			ssmAutomation := ssmAutomationAction.Value

			a := map[string]interface{}{}

			if v := ssmAutomation.DocumentName; v != nil {
				a["document_name"] = aws.ToString(v)
			}

			if v := ssmAutomation.RoleArn; v != nil {
				a[names.AttrRoleARN] = aws.ToString(v)
			}

			if v := ssmAutomation.DocumentVersion; v != nil {
				a["document_version"] = aws.ToString(v)
			}

			if v := ssmAutomation.TargetAccount; v != "" {
				a["target_account"] = ssmAutomation.TargetAccount
			}

			if v := ssmAutomation.Parameters; v != nil {
				a[names.AttrParameter] = flattenParameters(v)
			}

			if v := ssmAutomation.DynamicParameters; v != nil {
				a["dynamic_parameters"] = flattenDynamicParameters(v)
			}

			result = append(result, a)
		}
	}
	return result
}

func expandParameters(parameters *schema.Set) map[string][]string {
	parameterMap := make(map[string][]string)
	for _, parameter := range parameters.List() {
		parameterData := parameter.(map[string]interface{})
		name := parameterData[names.AttrName].(string)
		values := flex.ExpandStringValueSet(parameterData[names.AttrValues].(*schema.Set))
		parameterMap[name] = values
	}
	return parameterMap
}

func flattenParameters(parameterMap map[string][]string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for name, values := range parameterMap {
		data := make(map[string]interface{})
		data[names.AttrName] = name
		data[names.AttrValues] = flex.FlattenStringValueList(values)
		result = append(result, data)
	}
	return result
}

func expandDynamicParameters(parameterMap map[string]interface{}) map[string]types.DynamicSsmParameterValue {
	result := make(map[string]types.DynamicSsmParameterValue)
	for key, value := range parameterMap {
		parameterValue := &types.DynamicSsmParameterValueMemberVariable{
			Value: types.VariableType(value.(string)),
		}
		result[key] = parameterValue
	}
	return result
}

func flattenDynamicParameters(parameterMap map[string]types.DynamicSsmParameterValue) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range parameterMap {
		parameterValue := value.(*types.DynamicSsmParameterValueMemberVariable)
		result[key] = parameterValue.Value
	}

	return result
}

func expandIntegration(integrations []interface{}) []types.Integration {
	if len(integrations) == 0 {
		return nil
	}

	// we require exactly one integration item
	integrationData := integrations[0].(map[string]interface{})
	result := make([]types.Integration, 0)

	if v, ok := integrationData["pagerduty"].([]interface{}); ok {
		result = append(result, expandPagerDutyIntegration(v)...)
	}

	return result
}

func flattenIntegration(integrations []types.Integration) []interface{} {
	if len(integrations) == 0 {
		return nil
	}

	result := make([]interface{}, 0)

	integration := make(map[string]interface{})
	integration["pagerduty"] = flattenPagerDutyIntegration(integrations)
	result = append(result, integration)

	return result
}

func expandPagerDutyIntegration(integrations []interface{}) []types.Integration {
	result := make([]types.Integration, 0)

	for _, integration := range integrations {
		if integration == nil {
			continue
		}
		integrationData := integration.(map[string]interface{})

		pagerDutyIntegration := types.PagerDutyConfiguration{}

		if v, ok := integrationData[names.AttrName].(string); ok && v != "" {
			pagerDutyIntegration.Name = aws.String(v)
		}

		if v, ok := integrationData["service_id"].(string); ok && v != "" {
			pagerDutyIntegration.PagerDutyIncidentConfiguration =
				&types.PagerDutyIncidentConfiguration{
					ServiceId: aws.String(v),
				}
		}

		if v, ok := integrationData["secret_id"].(string); ok && v != "" {
			pagerDutyIntegration.SecretId = aws.String(v)
		}

		result = append(result, &types.IntegrationMemberPagerDutyConfiguration{Value: pagerDutyIntegration})
	}

	return result
}

func flattenPagerDutyIntegration(integrations []types.Integration) []interface{} {
	result := make([]interface{}, 0)

	for _, integration := range integrations {
		if v, ok := integration.(*types.IntegrationMemberPagerDutyConfiguration); ok {
			pagerDutyConfiguration := v.Value
			pagerDutyData := map[string]interface{}{}

			if v := pagerDutyConfiguration.Name; v != nil {
				pagerDutyData[names.AttrName] = v
			}

			if v := pagerDutyConfiguration.PagerDutyIncidentConfiguration.ServiceId; v != nil {
				pagerDutyData["service_id"] = v
			}

			if v := pagerDutyConfiguration.SecretId; v != nil {
				pagerDutyData["secret_id"] = v
			}

			result = append(result, pagerDutyData)
		}
	}
	return result
}
