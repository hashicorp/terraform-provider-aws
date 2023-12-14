// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandCapacityProviderStrategy(cps *schema.Set) []types.CapacityProviderStrategyItem {
	list := cps.List()
	results := make([]types.CapacityProviderStrategyItem, 0)
	for _, raw := range list {
		cp := raw.(map[string]interface{})
		ps := types.CapacityProviderStrategyItem{}
		if val, ok := cp["base"]; ok {
			ps.Base = int32(val.(int))
		}
		if val, ok := cp["weight"]; ok {
			ps.Weight = int32(val.(int))
		}
		if val, ok := cp["capacity_provider"]; ok {
			ps.CapacityProvider = aws.String(val.(string))
		}

		results = append(results, ps)
	}
	return results
}

func flattenCapacityProviderStrategy(cps []*types.CapacityProviderStrategyItem) []map[string]interface{} {
	if cps == nil {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, cp := range cps {
		s := make(map[string]interface{})
		s["capacity_provider"] = aws.ToString(cp.CapacityProvider)
		s["weight"] = cp.Weight
		s["base"] = cp.Base
		results = append(results, s)
	}
	return results
}

// Takes the result of flatmap. Expand for an array of load balancers and
// returns ecs.LoadBalancer compatible objects
func expandLoadBalancers(configured []interface{}) []*types.LoadBalancer {
	loadBalancers := make([]*types.LoadBalancer, 0, len(configured))

	// Loop over our configured load balancers and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &types.LoadBalancer{
			ContainerName: aws.String(data["container_name"].(string)),
			ContainerPort: aws.Int32(int32(data["container_port"].(int))),
		}

		if v, ok := data["elb_name"]; ok && v.(string) != "" {
			l.LoadBalancerName = aws.String(v.(string))
		}
		if v, ok := data["target_group_arn"]; ok && v.(string) != "" {
			l.TargetGroupArn = aws.String(v.(string))
		}

		loadBalancers = append(loadBalancers, l)
	}

	return loadBalancers
}

// Flattens an array of ECS LoadBalancers into a []map[string]interface{}
func flattenLoadBalancers(list []*types.LoadBalancer) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, loadBalancer := range list {
		l := map[string]interface{}{
			"container_name": *loadBalancer.ContainerName,
			"container_port": *loadBalancer.ContainerPort,
		}

		if loadBalancer.LoadBalancerName != nil {
			l["elb_name"] = aws.ToString(loadBalancer.LoadBalancerName)
		}

		if loadBalancer.TargetGroupArn != nil {
			l["target_group_arn"] = aws.ToString(loadBalancer.TargetGroupArn)
		}

		result = append(result, l)
	}
	return result
}

// Expand for an array of load balancers and
// returns ecs.LoadBalancer compatible objects for an ECS TaskSet
func expandTaskSetLoadBalancers(l []interface{}) []*types.LoadBalancer {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	loadBalancers := make([]*types.LoadBalancer, 0, len(l))

	// Loop over our configured load balancers and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range l {
		data := lRaw.(map[string]interface{})

		l := &types.LoadBalancer{}

		if v, ok := data["container_name"].(string); ok && v != "" {
			l.ContainerName = aws.String(v)
		}

		if v, ok := data["container_port"].(int); ok {
			l.ContainerPort = aws.Int32(int32(v))
		}

		if v, ok := data["load_balancer_name"]; ok && v.(string) != "" {
			l.LoadBalancerName = aws.String(v.(string))
		}
		if v, ok := data["target_group_arn"]; ok && v.(string) != "" {
			l.TargetGroupArn = aws.String(v.(string))
		}

		loadBalancers = append(loadBalancers, l)
	}

	return loadBalancers
}

// Flattens an array of ECS LoadBalancers (of an ECS TaskSet) into a []map[string]interface{}
func flattenTaskSetLoadBalancers(list []*types.LoadBalancer) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, loadBalancer := range list {
		l := map[string]interface{}{
			"container_name": loadBalancer.ContainerName,
			"container_port": loadBalancer.ContainerPort,
		}

		if loadBalancer.LoadBalancerName != nil {
			l["load_balancer_name"] = loadBalancer.LoadBalancerName
		}

		if loadBalancer.TargetGroupArn != nil {
			l["target_group_arn"] = loadBalancer.TargetGroupArn
		}

		result = append(result, l)
	}
	return result
}

// Expand for an array of service registries and
// returns ecs.ServiceRegistry compatible objects for an ECS TaskSet
func expandServiceRegistries(l []interface{}) []*types.ServiceRegistry {
	result := make([]*types.ServiceRegistry, 0, len(l))

	for _, v := range l {
		m := v.(map[string]interface{})
		sr := &types.ServiceRegistry{
			RegistryArn: aws.String(m["registry_arn"].(string)),
		}
		if raw, ok := m["container_name"].(string); ok && raw != "" {
			sr.ContainerName = aws.String(raw)
		}
		if raw, ok := m["container_port"].(int); ok && raw > 0 {
			sr.ContainerPort = aws.Int32(int32(raw))
		}
		if raw, ok := m["port"].(int); ok && raw > 0 {
			sr.Port = aws.Int32(int32(raw))
		}
		result = append(result, sr)
	}

	return result
}

// Expand for an array of scale configurations and
// returns an ecs.Scale compatible object for an ECS TaskSet
func expandScale(l []interface{}) *types.Scale {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Scale{}

	if v, ok := tfMap["unit"].(string); ok && v != "" {
		result.Unit = types.ScaleUnit(v)
	}

	if v, ok := tfMap["value"].(float64); ok {
		result.Value = v
	}

	return result
}

// Flattens an ECS Scale configuration into a []map[string]interface{}
func flattenScale(scale *types.Scale) []map[string]interface{} {
	if scale == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["unit"] = scale.Unit
	m["value"] = scale.Value

	return []map[string]interface{}{m}
}
