// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandCapacityProviderStrategy(cps *schema.Set) []*ecs.CapacityProviderStrategyItem {
	list := cps.List()
	results := make([]*ecs.CapacityProviderStrategyItem, 0)
	for _, raw := range list {
		cp := raw.(map[string]interface{})
		ps := &ecs.CapacityProviderStrategyItem{}
		if val, ok := cp["base"]; ok {
			ps.Base = aws.Int64(int64(val.(int)))
		}
		if val, ok := cp["weight"]; ok {
			ps.Weight = aws.Int64(int64(val.(int)))
		}
		if val, ok := cp["capacity_provider"]; ok {
			ps.CapacityProvider = aws.String(val.(string))
		}

		results = append(results, ps)
	}
	return results
}

func flattenCapacityProviderStrategy(cps []*ecs.CapacityProviderStrategyItem) []map[string]interface{} {
	if cps == nil {
		return nil
	}
	results := make([]map[string]interface{}, 0)
	for _, cp := range cps {
		s := make(map[string]interface{})
		s["capacity_provider"] = aws.StringValue(cp.CapacityProvider)
		if cp.Weight != nil {
			s["weight"] = aws.Int64Value(cp.Weight)
		}
		if cp.Base != nil {
			s["base"] = aws.Int64Value(cp.Base)
		}
		results = append(results, s)
	}
	return results
}

// Takes the result of flatmap. Expand for an array of load balancers and
// returns ecs.LoadBalancer compatible objects
func expandLoadBalancers(configured []interface{}) []*ecs.LoadBalancer {
	loadBalancers := make([]*ecs.LoadBalancer, 0, len(configured))

	// Loop over our configured load balancers and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		l := &ecs.LoadBalancer{
			ContainerName: aws.String(data["container_name"].(string)),
			ContainerPort: aws.Int64(int64(data["container_port"].(int))),
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
func flattenLoadBalancers(list []*ecs.LoadBalancer) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, loadBalancer := range list {
		l := map[string]interface{}{
			"container_name": *loadBalancer.ContainerName,
			"container_port": *loadBalancer.ContainerPort,
		}

		if loadBalancer.LoadBalancerName != nil {
			l["elb_name"] = aws.StringValue(loadBalancer.LoadBalancerName)
		}

		if loadBalancer.TargetGroupArn != nil {
			l["target_group_arn"] = aws.StringValue(loadBalancer.TargetGroupArn)
		}

		result = append(result, l)
	}
	return result
}

// Expand for an array of load balancers and
// returns ecs.LoadBalancer compatible objects for an ECS TaskSet
func expandTaskSetLoadBalancers(l []interface{}) []*ecs.LoadBalancer {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	loadBalancers := make([]*ecs.LoadBalancer, 0, len(l))

	// Loop over our configured load balancers and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range l {
		data := lRaw.(map[string]interface{})

		l := &ecs.LoadBalancer{}

		if v, ok := data["container_name"].(string); ok && v != "" {
			l.ContainerName = aws.String(v)
		}

		if v, ok := data["container_port"].(int); ok {
			l.ContainerPort = aws.Int64(int64(v))
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
func flattenTaskSetLoadBalancers(list []*ecs.LoadBalancer) []map[string]interface{} {
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
func expandServiceRegistries(l []interface{}) []*ecs.ServiceRegistry {
	result := make([]*ecs.ServiceRegistry, 0, len(l))

	for _, v := range l {
		m := v.(map[string]interface{})
		sr := &ecs.ServiceRegistry{
			RegistryArn: aws.String(m["registry_arn"].(string)),
		}
		if raw, ok := m["container_name"].(string); ok && raw != "" {
			sr.ContainerName = aws.String(raw)
		}
		if raw, ok := m["container_port"].(int); ok && raw > 0 {
			sr.ContainerPort = aws.Int64(int64(raw))
		}
		if raw, ok := m["port"].(int); ok && raw > 0 {
			sr.Port = aws.Int64(int64(raw))
		}
		result = append(result, sr)
	}

	return result
}

// Expand for an array of scale configurations and
// returns an ecs.Scale compatible object for an ECS TaskSet
func expandScale(l []interface{}) *ecs.Scale {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &ecs.Scale{}

	if v, ok := tfMap["unit"].(string); ok && v != "" {
		result.Unit = aws.String(v)
	}

	if v, ok := tfMap["value"].(float64); ok {
		result.Value = aws.Float64(v)
	}

	return result
}

// Flattens an ECS Scale configuration into a []map[string]interface{}
func flattenScale(scale *ecs.Scale) []map[string]interface{} {
	if scale == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["unit"] = aws.StringValue(scale.Unit)
	m["value"] = aws.Float64Value(scale.Value)

	return []map[string]interface{}{m}
}
