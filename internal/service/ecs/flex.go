package ecs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

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
func flattenECSLoadBalancers(list []*ecs.LoadBalancer) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, loadBalancer := range list {
		l := map[string]interface{}{
			"container_name": *loadBalancer.ContainerName,
			"container_port": *loadBalancer.ContainerPort,
		}

		if loadBalancer.LoadBalancerName != nil {
			l["elb_name"] = *loadBalancer.LoadBalancerName
		}

		if loadBalancer.TargetGroupArn != nil {
			l["target_group_arn"] = *loadBalancer.TargetGroupArn
		}

		result = append(result, l)
	}
	return result
}
