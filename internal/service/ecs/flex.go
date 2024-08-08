// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandCapacityProviderStrategyItems(tfSet *schema.Set) []awstypes.CapacityProviderStrategyItem {
	tfList := tfSet.List()
	apiObjects := make([]awstypes.CapacityProviderStrategyItem, 0)

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObject := awstypes.CapacityProviderStrategyItem{}

		if v, ok := tfMap["base"]; ok {
			apiObject.Base = int32(v.(int))
		}

		if v, ok := tfMap["capacity_provider"]; ok {
			apiObject.CapacityProvider = aws.String(v.(string))
		}

		if v, ok := tfMap[names.AttrWeight]; ok {
			apiObject.Weight = int32(v.(int))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenCapacityProviderStrategyItems(apiObjects []awstypes.CapacityProviderStrategyItem) []interface{} {
	if apiObjects == nil {
		return nil
	}

	tfList := make([]interface{}, 0)

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{})

		tfMap["base"] = apiObject.Base
		tfMap["capacity_provider"] = aws.ToString(apiObject.CapacityProvider)
		tfMap[names.AttrWeight] = apiObject.Weight

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandLoadBalancers(tfList []interface{}) []awstypes.LoadBalancer {
	apiObjects := make([]awstypes.LoadBalancer, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := awstypes.LoadBalancer{
			ContainerName: aws.String(tfMap["container_name"].(string)),
			ContainerPort: aws.Int32(int32(tfMap["container_port"].(int))),
		}

		if v, ok := tfMap["elb_name"]; ok && v.(string) != "" {
			apiObject.LoadBalancerName = aws.String(v.(string))
		}

		if v, ok := tfMap["target_group_arn"]; ok && v.(string) != "" {
			apiObject.TargetGroupArn = aws.String(v.(string))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenLoadBalancers(apiObjects []awstypes.LoadBalancer) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"container_name": aws.ToString(apiObject.ContainerName),
			"container_port": aws.ToInt32(apiObject.ContainerPort),
		}

		if apiObject.LoadBalancerName != nil {
			tfMap["elb_name"] = aws.ToString(apiObject.LoadBalancerName)
		}

		if apiObject.TargetGroupArn != nil {
			tfMap["target_group_arn"] = aws.ToString(apiObject.TargetGroupArn)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandTaskSetLoadBalancers(tfList []interface{}) []awstypes.LoadBalancer {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	apiObjects := make([]awstypes.LoadBalancer, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})

		apiObject := awstypes.LoadBalancer{}

		if v, ok := tfMap["container_name"].(string); ok && v != "" {
			apiObject.ContainerName = aws.String(v)
		}

		if v, ok := tfMap["container_port"].(int); ok {
			apiObject.ContainerPort = aws.Int32(int32(v))
		}

		if v, ok := tfMap["load_balancer_name"]; ok && v.(string) != "" {
			apiObject.LoadBalancerName = aws.String(v.(string))
		}

		if v, ok := tfMap["target_group_arn"]; ok && v.(string) != "" {
			apiObject.TargetGroupArn = aws.String(v.(string))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenTaskSetLoadBalancers(apiObjects []awstypes.LoadBalancer) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"container_name": aws.ToString(apiObject.ContainerName),
			"container_port": aws.ToInt32(apiObject.ContainerPort),
		}

		if apiObject.LoadBalancerName != nil {
			tfMap["load_balancer_name"] = aws.ToString(apiObject.LoadBalancerName)
		}

		if apiObject.TargetGroupArn != nil {
			tfMap["target_group_arn"] = aws.ToString(apiObject.TargetGroupArn)
		}

		tfList = append(tfList, tfMap)
	}
	return tfList
}

func expandServiceRegistries(tfList []interface{}) []awstypes.ServiceRegistry {
	apiObjects := make([]awstypes.ServiceRegistry, 0, len(tfList))

	for _, tfMapRaw := range tfList {
		tfMap := tfMapRaw.(map[string]interface{})
		apiObject := awstypes.ServiceRegistry{
			RegistryArn: aws.String(tfMap["registry_arn"].(string)),
		}

		if v, ok := tfMap["container_name"].(string); ok && v != "" {
			apiObject.ContainerName = aws.String(v)
		}

		if v, ok := tfMap["container_port"].(int); ok && v > 0 {
			apiObject.ContainerPort = aws.Int32(int32(v))
		}

		if v, ok := tfMap[names.AttrPort].(int); ok && v > 0 {
			apiObject.Port = aws.Int32(int32(v))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandScale(tfList []interface{}) *awstypes.Scale {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &awstypes.Scale{}

	if v, ok := tfMap[names.AttrUnit].(string); ok && v != "" {
		apiObject.Unit = awstypes.ScaleUnit(v)
	}

	if v, ok := tfMap[names.AttrValue].(float64); ok {
		apiObject.Value = v
	}

	return apiObject
}

func flattenScale(apiObject *awstypes.Scale) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := make(map[string]interface{})
	tfMap[names.AttrUnit] = string(apiObject.Unit)
	tfMap[names.AttrValue] = apiObject.Value

	return []interface{}{tfMap}
}
