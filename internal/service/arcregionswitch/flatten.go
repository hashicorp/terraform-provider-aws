package arcregionswitch

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func flattenSteps(apiObject []types.Step) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, step := range apiObject {
		tfMap := map[string]interface{}{
			"execution_block_type": string(step.ExecutionBlockType),
		}

		if step.Name != nil {
			tfMap["name"] = aws.ToString(step.Name)
		}

		if step.Description != nil {
			tfMap["description"] = aws.ToString(step.Description)
		}

		// Handle different execution block configuration types
		if step.ExecutionBlockConfiguration != nil {
			switch config := step.ExecutionBlockConfiguration.(type) {
			case *types.ExecutionBlockConfigurationMemberExecutionApprovalConfig:
				tfMap["execution_approval_config"] = flattenExecutionApprovalConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
				tfMap["custom_action_lambda_config"] = flattenCustomActionLambdaConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberArcRoutingControlConfig:
				tfMap["arc_routing_control_config"] = flattenArcRoutingControlConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig:
				tfMap["ec2_asg_capacity_increase_config"] = flattenEc2AsgCapacityIncreaseConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberGlobalAuroraConfig:
				tfMap["global_aurora_config"] = flattenGlobalAuroraConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberParallelConfig:
				tfMap["parallel_config"] = flattenParallelConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig:
				tfMap["ecs_capacity_increase_config"] = flattenEcsCapacityIncreaseConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberEksResourceScalingConfig:
				tfMap["eks_resource_scaling_config"] = flattenEksResourceScalingConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig:
				tfMap["route53_health_check_config"] = flattenRoute53HealthCheckConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberRegionSwitchPlanConfig:
				tfMap["region_switch_plan_config"] = flattenRegionSwitchPlanConfig(config.Value)
			}
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenExecutionApprovalConfig(apiObject types.ExecutionApprovalConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if apiObject.ApprovalRole != nil {
		tfMap["approval_role"] = aws.ToString(apiObject.ApprovalRole)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	return []interface{}{tfMap}
}

func flattenCustomActionLambdaConfig(apiObject types.CustomActionLambdaConfiguration) []interface{} {
	tfMap := map[string]interface{}{
		"region_to_run": string(apiObject.RegionToRun),
	}

	if apiObject.RetryIntervalMinutes != nil {
		tfMap["retry_interval_minutes"] = aws.ToFloat32(apiObject.RetryIntervalMinutes)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	if len(apiObject.Lambdas) > 0 {
		tfMap["lambda"] = flattenLambdas(apiObject.Lambdas)
	}

	if apiObject.Ungraceful != nil {
		tfMap["ungraceful"] = []interface{}{
			map[string]interface{}{
				"behavior": string(apiObject.Ungraceful.Behavior),
			},
		}
	}

	return []interface{}{tfMap}
}

func flattenLambdas(apiObject []types.Lambdas) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, lambda := range apiObject {
		tfMap := map[string]interface{}{}

		if lambda.Arn != nil {
			tfMap["arn"] = aws.ToString(lambda.Arn)
		}

		if lambda.CrossAccountRole != nil {
			tfMap["cross_account_role"] = aws.ToString(lambda.CrossAccountRole)
		}

		if lambda.ExternalId != nil {
			tfMap["external_id"] = aws.ToString(lambda.ExternalId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenArcRoutingControlConfig(apiObject types.ArcRoutingControlConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if len(apiObject.RegionAndRoutingControls) > 0 {
		// Flatten to list of objects with region and routing_control_arns
		var regionControlsList []interface{}

		// Sort regions for consistent ordering
		var regions []string
		for region := range apiObject.RegionAndRoutingControls {
			regions = append(regions, region)
		}
		sort.Strings(regions)

		for _, region := range regions {
			controls := apiObject.RegionAndRoutingControls[region]
			var controlArns []string
			for _, control := range controls {
				if control.RoutingControlArn != nil {
					controlArns = append(controlArns, aws.ToString(control.RoutingControlArn))
				}
			}
			regionControlsList = append(regionControlsList, map[string]interface{}{
				"region":               region,
				"routing_control_arns": controlArns,
			})
		}
		tfMap["region_and_routing_controls"] = regionControlsList
	}

	if apiObject.CrossAccountRole != nil {
		tfMap["cross_account_role"] = aws.ToString(apiObject.CrossAccountRole)
	}

	if apiObject.ExternalId != nil {
		tfMap["external_id"] = aws.ToString(apiObject.ExternalId)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	return []interface{}{tfMap}
}

func flattenEc2AsgCapacityIncreaseConfig(apiObject types.Ec2AsgCapacityIncreaseConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if len(apiObject.Asgs) > 0 {
		tfMap["asgs"] = flattenAsgs(apiObject.Asgs)
	}

	if apiObject.CapacityMonitoringApproach != "" {
		tfMap["capacity_monitoring_approach"] = string(apiObject.CapacityMonitoringApproach)
	}

	if apiObject.TargetPercent != nil {
		tfMap["target_percent"] = aws.ToInt32(apiObject.TargetPercent)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	if apiObject.Ungraceful != nil {
		tfMap["ungraceful"] = []interface{}{
			map[string]interface{}{
				"minimum_success_percentage": aws.ToInt32(apiObject.Ungraceful.MinimumSuccessPercentage),
			},
		}
	}

	return []interface{}{tfMap}
}

func flattenAsgs(apiObject []types.Asg) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, asg := range apiObject {
		tfMap := map[string]interface{}{}

		if asg.Arn != nil {
			tfMap["arn"] = aws.ToString(asg.Arn)
		}

		if asg.CrossAccountRole != nil {
			tfMap["cross_account_role"] = aws.ToString(asg.CrossAccountRole)
		}

		if asg.ExternalId != nil {
			tfMap["external_id"] = aws.ToString(asg.ExternalId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenGlobalAuroraConfig(apiObject types.GlobalAuroraConfiguration) []interface{} {
	tfMap := map[string]interface{}{
		"behavior":                  string(apiObject.Behavior),
		"global_cluster_identifier": aws.ToString(apiObject.GlobalClusterIdentifier),
	}

	if len(apiObject.DatabaseClusterArns) > 0 {
		tfMap["database_cluster_arns"] = flex.FlattenStringValueList(apiObject.DatabaseClusterArns)
	}

	if apiObject.CrossAccountRole != nil {
		tfMap["cross_account_role"] = aws.ToString(apiObject.CrossAccountRole)
	}

	if apiObject.ExternalId != nil {
		tfMap["external_id"] = aws.ToString(apiObject.ExternalId)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	if apiObject.Ungraceful != nil {
		tfMap["ungraceful"] = []interface{}{
			map[string]interface{}{
				"ungraceful": string(apiObject.Ungraceful.Ungraceful),
			},
		}
	}

	return []interface{}{tfMap}
}

func flattenParallelConfig(apiObject types.ParallelExecutionBlockConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if len(apiObject.Steps) > 0 {
		tfMap["step"] = flattenSteps(apiObject.Steps)
	}

	return []interface{}{tfMap}
}

func flattenEcsCapacityIncreaseConfig(apiObject types.EcsCapacityIncreaseConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if len(apiObject.Services) > 0 {
		tfMap["services"] = flattenServices(apiObject.Services)
	}

	if apiObject.CapacityMonitoringApproach != "" {
		tfMap["capacity_monitoring_approach"] = string(apiObject.CapacityMonitoringApproach)
	}

	if apiObject.TargetPercent != nil {
		tfMap["target_percent"] = aws.ToInt32(apiObject.TargetPercent)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	if apiObject.Ungraceful != nil {
		tfMap["ungraceful"] = []interface{}{
			map[string]interface{}{
				"minimum_success_percentage": aws.ToInt32(apiObject.Ungraceful.MinimumSuccessPercentage),
			},
		}
	}

	return []interface{}{tfMap}
}

func flattenServices(apiObject []types.Service) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, service := range apiObject {
		tfMap := map[string]interface{}{}

		if service.ClusterArn != nil {
			tfMap["cluster_arn"] = aws.ToString(service.ClusterArn)
		}

		if service.ServiceArn != nil {
			tfMap["service_arn"] = aws.ToString(service.ServiceArn)
		}

		if service.CrossAccountRole != nil {
			tfMap["cross_account_role"] = aws.ToString(service.CrossAccountRole)
		}

		if service.ExternalId != nil {
			tfMap["external_id"] = aws.ToString(service.ExternalId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEksResourceScalingConfig(apiObject types.EksResourceScalingConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if apiObject.KubernetesResourceType != nil {
		tfMap["kubernetes_resource_type"] = []interface{}{
			map[string]interface{}{
				"api_version": aws.ToString(apiObject.KubernetesResourceType.ApiVersion),
				"kind":        aws.ToString(apiObject.KubernetesResourceType.Kind),
			},
		}
	}

	if len(apiObject.EksClusters) > 0 {
		tfMap["eks_clusters"] = flattenEksClusters(apiObject.EksClusters)
	}

	if apiObject.CapacityMonitoringApproach != "" {
		tfMap["capacity_monitoring_approach"] = string(apiObject.CapacityMonitoringApproach)
	}

	if apiObject.TargetPercent != nil {
		tfMap["target_percent"] = aws.ToInt32(apiObject.TargetPercent)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	if apiObject.Ungraceful != nil {
		tfMap["ungraceful"] = []interface{}{
			map[string]interface{}{
				"minimum_success_percentage": aws.ToInt32(apiObject.Ungraceful.MinimumSuccessPercentage),
			},
		}
	}

	// Handle complex scaling_resources structure
	if len(apiObject.ScalingResources) > 0 {
		var scalingResourcesList []interface{}
		for _, resourceMap := range apiObject.ScalingResources {
			for namespace, resources := range resourceMap {
				var resourcesList []interface{}
				for resourceName, resource := range resources {
					resourceData := map[string]interface{}{
						"resource_name": resourceName,
						"name":          aws.ToString(resource.Name),
						"namespace":     aws.ToString(resource.Namespace),
					}
					if resource.HpaName != nil {
						resourceData["hpa_name"] = aws.ToString(resource.HpaName)
					}
					resourcesList = append(resourcesList, resourceData)
				}
				scalingResourcesList = append(scalingResourcesList, map[string]interface{}{
					"namespace": namespace,
					"resources": resourcesList,
				})
			}
		}
		tfMap["scaling_resources"] = scalingResourcesList
	}

	return []interface{}{tfMap}
}

func flattenEksClusters(apiObject []types.EksCluster) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, cluster := range apiObject {
		tfMap := map[string]interface{}{}

		if cluster.ClusterArn != nil {
			tfMap["cluster_arn"] = aws.ToString(cluster.ClusterArn)
		}

		if cluster.CrossAccountRole != nil {
			tfMap["cross_account_role"] = aws.ToString(cluster.CrossAccountRole)
		}

		if cluster.ExternalId != nil {
			tfMap["external_id"] = aws.ToString(cluster.ExternalId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRoute53HealthCheckConfig(apiObject types.Route53HealthCheckConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if apiObject.HostedZoneId != nil {
		tfMap["hosted_zone_id"] = aws.ToString(apiObject.HostedZoneId)
	}

	if apiObject.RecordName != nil {
		tfMap["record_name"] = aws.ToString(apiObject.RecordName)
	}

	if len(apiObject.RecordSets) > 0 {
		tfMap["record_sets"] = flattenRoute53ResourceRecordSets(apiObject.RecordSets)
	}

	if apiObject.CrossAccountRole != nil {
		tfMap["cross_account_role"] = aws.ToString(apiObject.CrossAccountRole)
	}

	if apiObject.ExternalId != nil {
		tfMap["external_id"] = aws.ToString(apiObject.ExternalId)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	return []interface{}{tfMap}
}

func flattenRoute53ResourceRecordSets(apiObject []types.Route53ResourceRecordSet) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, recordSet := range apiObject {
		tfMap := map[string]interface{}{}

		if recordSet.RecordSetIdentifier != nil {
			tfMap["record_set_identifier"] = aws.ToString(recordSet.RecordSetIdentifier)
		}

		if recordSet.Region != nil {
			tfMap["region"] = aws.ToString(recordSet.Region)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRegionSwitchPlanConfig(apiObject types.RegionSwitchPlanConfiguration) []interface{} {
	tfMap := map[string]interface{}{}

	if apiObject.Arn != nil {
		tfMap["arn"] = aws.ToString(apiObject.Arn)
	}

	if apiObject.CrossAccountRole != nil {
		tfMap["cross_account_role"] = aws.ToString(apiObject.CrossAccountRole)
	}

	if apiObject.ExternalId != nil {
		tfMap["external_id"] = aws.ToString(apiObject.ExternalId)
	}

	return []interface{}{tfMap}
}

func flattenWorkflows(apiObject []types.Workflow) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, workflow := range apiObject {
		tfMap := map[string]interface{}{
			"workflow_target_action": string(workflow.WorkflowTargetAction),
		}

		if workflow.WorkflowTargetRegion != nil {
			tfMap["workflow_target_region"] = aws.ToString(workflow.WorkflowTargetRegion)
		}

		if workflow.WorkflowDescription != nil {
			tfMap["workflow_description"] = aws.ToString(workflow.WorkflowDescription)
		}

		if len(workflow.Steps) > 0 {
			tfMap["step"] = flattenSteps(workflow.Steps)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenAssociatedAlarms(apiObject map[string]types.AssociatedAlarm) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for alarmName, alarm := range apiObject {
		alarmMap := map[string]interface{}{
			"name":                alarmName,
			"alarm_type":          string(alarm.AlarmType),
			"resource_identifier": aws.ToString(alarm.ResourceIdentifier),
		}

		if alarm.CrossAccountRole != nil {
			alarmMap["cross_account_role"] = aws.ToString(alarm.CrossAccountRole)
		}

		if alarm.ExternalId != nil {
			alarmMap["external_id"] = aws.ToString(alarm.ExternalId)
		}

		tfList = append(tfList, alarmMap)
	}

	return tfList
}

func flattenTriggers(apiObject []types.Trigger) []interface{} {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, trigger := range apiObject {
		tfMap := map[string]interface{}{
			"action":                               string(trigger.Action),
			"min_delay_minutes_between_executions": aws.ToInt32(trigger.MinDelayMinutesBetweenExecutions),
			"target_region":                        aws.ToString(trigger.TargetRegion),
		}

		if trigger.Description != nil {
			tfMap["description"] = aws.ToString(trigger.Description)
		}

		if len(trigger.Conditions) > 0 {
			var conditions []interface{}
			for _, condition := range trigger.Conditions {
				conditionMap := map[string]interface{}{
					"associated_alarm_name": aws.ToString(condition.AssociatedAlarmName),
					"condition":             string(condition.Condition),
				}
				conditions = append(conditions, conditionMap)
			}
			tfMap["conditions"] = conditions
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
