// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func encryptionConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		MaxItems: 1,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrKeyID: {
					Type:     schema.TypeString,
					Optional: true,
				},
				names.AttrType: {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[awstypes.EncryptionType](),
				},
			},
		},
	}
}

func expandEncryptionConfiguration(tfList []interface{}) *awstypes.EncryptionConfiguration {
	apiObject := &awstypes.EncryptionConfiguration{
		Type: awstypes.EncryptionTypeAwsOwnedKmsKey,
	}

	if len(tfList) == 1 && tfList[0] != nil {
		tfMap := tfList[0].(map[string]interface{})

		if v, ok := tfMap[names.AttrKeyID].(string); ok {
			apiObject.KeyId = aws.String(v)
		}
		if v, ok := tfMap[names.AttrType].(string); ok {
			apiObject.Type = awstypes.EncryptionType(v)
		}
	}

	return apiObject
}

func flattenEncryptionConfiguration(apiObject *awstypes.EncryptionConfiguration) []interface{} {
	if apiObject == nil || apiObject.Type == "" {
		return nil
	}

	if apiObject.Type == awstypes.EncryptionTypeAwsOwnedKmsKey {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrKeyID: aws.ToString(apiObject.KeyId),
		names.AttrType:  apiObject.Type,
	}

	return []interface{}{tfMap}
}

func customActionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"action_definition": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"publish_metric_action": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"dimension": {
											Type:     schema.TypeSet,
											Required: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrValue: {
														Type:     schema.TypeString,
														Required: true,
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
				"action_name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]+$`), "must contain only alphanumeric characters"),
				},
			},
		},
	}
}

func expandCustomActions(tfList []interface{}) []awstypes.CustomAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.CustomAction, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.CustomAction{}

		if v, ok := tfMap["action_definition"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			apiObject.ActionDefinition = expandActionDefinition(v)
		}
		if v, ok := tfMap["action_name"].(string); ok && v != "" {
			apiObject.ActionName = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandActionDefinition(tfList []interface{}) *awstypes.ActionDefinition {
	if tfList == nil || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.ActionDefinition{}

	if v, ok := tfMap["publish_metric_action"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.PublishMetricAction = expandPublishMetricAction(v)
	}

	return apiObject
}

func expandPublishMetricAction(tfList []interface{}) *awstypes.PublishMetricAction {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.PublishMetricAction{}

	if v, ok := tfMap["dimension"].(*schema.Set); ok && v.Len() > 0 {
		tfList := v.List()
		dimensions := make([]awstypes.Dimension, 0, len(tfList))

		for _, tfMapRaw := range tfList {
			tfMap, ok := tfMapRaw.(map[string]interface{})
			if !ok {
				continue
			}

			dimensions = append(dimensions, awstypes.Dimension{
				Value: aws.String(tfMap[names.AttrValue].(string)),
			})
		}

		apiObject.Dimensions = dimensions
	}

	return apiObject
}

func flattenCustomActions(apiObjects []awstypes.CustomAction) []interface{} {
	if apiObjects == nil {
		return []interface{}{}
	}

	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"action_definition": flattenActionDefinition(apiObject.ActionDefinition),
			"action_name":       aws.ToString(apiObject.ActionName),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenActionDefinition(apiObject *awstypes.ActionDefinition) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"publish_metric_action": flattenPublishMetricAction(apiObject.PublishMetricAction),
	}

	return []interface{}{tfMap}
}

func flattenPublishMetricAction(apiObject *awstypes.PublishMetricAction) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"dimension": flattenDimensions(apiObject.Dimensions),
	}

	return []interface{}{tfMap}
}

func flattenDimensions(apiObjects []awstypes.Dimension) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			names.AttrValue: aws.ToString(apiObject.Value),
		})
	}

	return tfList
}

func forceNewIfNotRuleOrderDefault(key string, d *schema.ResourceDiff) error {
	if d.Id() != "" && d.HasChange(key) {
		old, new := d.GetChange(key)
		defaultRuleOrderOld := old == nil || old.(string) == "" || old.(string) == string(awstypes.RuleOrderDefaultActionOrder)
		defaultRuleOrderNew := new == nil || new.(string) == "" || new.(string) == string(awstypes.RuleOrderDefaultActionOrder)

		if (defaultRuleOrderOld && !defaultRuleOrderNew) || (defaultRuleOrderNew && !defaultRuleOrderOld) {
			return d.ForceNew(key)
		}
	}

	return nil
}

func expandIPSets(tfList []interface{}) map[string]awstypes.IPSet {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObject := make(map[string]awstypes.IPSet)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if k, ok := tfMap[names.AttrKey].(string); ok && k != "" {
			if tfList, ok := tfMap["ip_set"].([]interface{}); ok && len(tfList) > 0 && tfList[0] != nil {
				tfMap, ok := tfList[0].(map[string]interface{})
				if !ok {
					continue
				}

				if v, ok := tfMap["definition"].(*schema.Set); ok && v.Len() > 0 {
					apiObject[k] = awstypes.IPSet{
						Definition: flex.ExpandStringValueSet(v),
					}
				}
			}
		}
	}

	return apiObject
}

func flattenIPSets(tfMap map[string]awstypes.IPSet) []interface{} {
	if tfMap == nil {
		return []interface{}{}
	}

	tfList := make([]interface{}, 0, len(tfMap))

	for k, v := range tfMap {
		tfList = append(tfList, map[string]interface{}{
			names.AttrKey: k,
			"ip_set":      flattenIPSet(&v),
		})
	}

	return tfList
}

func flattenIPSet(apiObject *awstypes.IPSet) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"definition": apiObject.Definition,
	}

	return []interface{}{tfMap}
}
