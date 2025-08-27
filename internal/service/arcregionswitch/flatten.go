// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenSteps(apiObject []types.Step) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, step := range apiObject {
		tfMap := map[string]any{
			"execution_block_type": string(step.ExecutionBlockType),
		}

		if step.Name != nil {
			tfMap[names.AttrName] = aws.ToString(step.Name)
		}

		if step.Description != nil {
			tfMap[names.AttrDescription] = aws.ToString(step.Description)
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
				tfMap["ec2_asg_capacity_increase_config"] = flattenEC2ASGCapacityIncreaseConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberGlobalAuroraConfig:
				tfMap["global_aurora_config"] = flattenGlobalAuroraConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberParallelConfig:
				tfMap["parallel_config"] = flattenParallelConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig:
				tfMap["ecs_capacity_increase_config"] = flattenECSCapacityIncreaseConfig(config.Value)
			case *types.ExecutionBlockConfigurationMemberEksResourceScalingConfig:
				tfMap["eks_resource_scaling_config"] = flattenEKSResourceScalingConfig(config.Value)
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

func flattenExecutionApprovalConfig(apiObject types.ExecutionApprovalConfiguration) []any {
	tfMap := map[string]any{}

	if apiObject.ApprovalRole != nil {
		tfMap["approval_role"] = aws.ToString(apiObject.ApprovalRole)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	return []any{tfMap}
}

func flattenCustomActionLambdaConfig(apiObject types.CustomActionLambdaConfiguration) []any {
	tfMap := map[string]any{
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
		tfMap["ungraceful"] = []any{
			map[string]any{
				"behavior": string(apiObject.Ungraceful.Behavior),
			},
		}
	}

	return []any{tfMap}
}

func flattenLambdas(apiObject []types.Lambdas) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, lambda := range apiObject {
		tfMap := map[string]any{}

		if lambda.Arn != nil {
			tfMap[names.AttrARN] = aws.ToString(lambda.Arn)
		}

		if lambda.CrossAccountRole != nil {
			tfMap["cross_account_role"] = aws.ToString(lambda.CrossAccountRole)
		}

		if lambda.ExternalId != nil {
			tfMap[names.AttrExternalID] = aws.ToString(lambda.ExternalId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenArcRoutingControlConfig(apiObject types.ArcRoutingControlConfiguration) []any {
	tfMap := map[string]any{}

	if len(apiObject.RegionAndRoutingControls) > 0 {
		// Flatten to list of objects with region and routing_control_arns
		var regionControlsList []any

		// Sort regions for consistent ordering
		var regions []string
		for region := range apiObject.RegionAndRoutingControls {
			regions = append(regions, region)
		}
		slices.Sort(regions)

		for _, region := range regions {
			controls := apiObject.RegionAndRoutingControls[region]
			var controlArns []string
			for _, control := range controls {
				if control.RoutingControlArn != nil {
					controlArns = append(controlArns, aws.ToString(control.RoutingControlArn))
				}
			}
			regionControlsList = append(regionControlsList, map[string]any{
				names.AttrRegion:       region,
				"routing_control_arns": controlArns,
			})
		}
		tfMap["region_and_routing_controls"] = regionControlsList
	}

	if apiObject.CrossAccountRole != nil {
		tfMap["cross_account_role"] = aws.ToString(apiObject.CrossAccountRole)
	}

	if apiObject.ExternalId != nil {
		tfMap[names.AttrExternalID] = aws.ToString(apiObject.ExternalId)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	return []any{tfMap}
}

func flattenEC2ASGCapacityIncreaseConfig(apiObject types.Ec2AsgCapacityIncreaseConfiguration) []any {
	tfMap := map[string]any{}

	if len(apiObject.Asgs) > 0 {
		tfMap["asgs"] = flattenASGs(apiObject.Asgs)
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
		tfMap["ungraceful"] = []any{
			map[string]any{
				"minimum_success_percentage": aws.ToInt32(apiObject.Ungraceful.MinimumSuccessPercentage),
			},
		}
	}

	return []any{tfMap}
}

func flattenASGs(apiObject []types.Asg) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, asg := range apiObject {
		tfMap := map[string]any{}

		if asg.Arn != nil {
			tfMap[names.AttrARN] = aws.ToString(asg.Arn)
		}

		if asg.CrossAccountRole != nil {
			tfMap["cross_account_role"] = aws.ToString(asg.CrossAccountRole)
		}

		if asg.ExternalId != nil {
			tfMap[names.AttrExternalID] = aws.ToString(asg.ExternalId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenGlobalAuroraConfig(apiObject types.GlobalAuroraConfiguration) []any {
	tfMap := map[string]any{
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
		tfMap[names.AttrExternalID] = aws.ToString(apiObject.ExternalId)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	if apiObject.Ungraceful != nil {
		tfMap["ungraceful"] = []any{
			map[string]any{
				"ungraceful": string(apiObject.Ungraceful.Ungraceful),
			},
		}
	}

	return []any{tfMap}
}

func flattenParallelConfig(apiObject types.ParallelExecutionBlockConfiguration) []any {
	tfMap := map[string]any{}

	if len(apiObject.Steps) > 0 {
		tfMap["step"] = flattenSteps(apiObject.Steps)
	}

	return []any{tfMap}
}

func flattenECSCapacityIncreaseConfig(apiObject types.EcsCapacityIncreaseConfiguration) []any {
	tfMap := map[string]any{}

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
		tfMap["ungraceful"] = []any{
			map[string]any{
				"minimum_success_percentage": aws.ToInt32(apiObject.Ungraceful.MinimumSuccessPercentage),
			},
		}
	}

	return []any{tfMap}
}

func flattenServices(apiObject []types.Service) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, service := range apiObject {
		tfMap := map[string]any{}

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
			tfMap[names.AttrExternalID] = aws.ToString(service.ExternalId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenEKSResourceScalingConfig(apiObject types.EksResourceScalingConfiguration) []any {
	tfMap := map[string]any{}

	if apiObject.KubernetesResourceType != nil {
		tfMap["kubernetes_resource_type"] = []any{
			map[string]any{
				"api_version": aws.ToString(apiObject.KubernetesResourceType.ApiVersion),
				"kind":        aws.ToString(apiObject.KubernetesResourceType.Kind),
			},
		}
	}

	if len(apiObject.EksClusters) > 0 {
		tfMap["eks_clusters"] = flattenEKSClusters(apiObject.EksClusters)
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
		tfMap["ungraceful"] = []any{
			map[string]any{
				"minimum_success_percentage": aws.ToInt32(apiObject.Ungraceful.MinimumSuccessPercentage),
			},
		}
	}

	// Handle complex scaling_resources structure
	if len(apiObject.ScalingResources) > 0 {
		var scalingResourcesList []any
		for _, resourceMap := range apiObject.ScalingResources {
			for namespace, resources := range resourceMap {
				var resourcesList []any
				for resourceName, resource := range resources {
					resourceData := map[string]any{
						"resource_name":     resourceName,
						names.AttrName:      aws.ToString(resource.Name),
						names.AttrNamespace: aws.ToString(resource.Namespace),
					}
					if resource.HpaName != nil {
						resourceData["hpa_name"] = aws.ToString(resource.HpaName)
					}
					resourcesList = append(resourcesList, resourceData)
				}
				scalingResourcesList = append(scalingResourcesList, map[string]any{
					names.AttrNamespace: namespace,
					names.AttrResources: resourcesList,
				})
			}
		}
		tfMap["scaling_resources"] = scalingResourcesList
	}

	return []any{tfMap}
}

func flattenEKSClusters(apiObject []types.EksCluster) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, cluster := range apiObject {
		tfMap := map[string]any{}

		if cluster.ClusterArn != nil {
			tfMap["cluster_arn"] = aws.ToString(cluster.ClusterArn)
		}

		if cluster.CrossAccountRole != nil {
			tfMap["cross_account_role"] = aws.ToString(cluster.CrossAccountRole)
		}

		if cluster.ExternalId != nil {
			tfMap[names.AttrExternalID] = aws.ToString(cluster.ExternalId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRoute53HealthCheckConfig(apiObject types.Route53HealthCheckConfiguration) []any {
	tfMap := map[string]any{}

	if apiObject.HostedZoneId != nil {
		tfMap[names.AttrHostedZoneID] = aws.ToString(apiObject.HostedZoneId)
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
		tfMap[names.AttrExternalID] = aws.ToString(apiObject.ExternalId)
	}

	if apiObject.TimeoutMinutes != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(apiObject.TimeoutMinutes)
	}

	return []any{tfMap}
}

func flattenRoute53ResourceRecordSets(apiObject []types.Route53ResourceRecordSet) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, recordSet := range apiObject {
		tfMap := map[string]any{}

		if recordSet.RecordSetIdentifier != nil {
			tfMap["record_set_identifier"] = aws.ToString(recordSet.RecordSetIdentifier)
		}

		if recordSet.Region != nil {
			tfMap[names.AttrRegion] = aws.ToString(recordSet.Region)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenRegionSwitchPlanConfig(apiObject types.RegionSwitchPlanConfiguration) []any {
	tfMap := map[string]any{}

	if apiObject.Arn != nil {
		tfMap[names.AttrARN] = aws.ToString(apiObject.Arn)
	}

	if apiObject.CrossAccountRole != nil {
		tfMap["cross_account_role"] = aws.ToString(apiObject.CrossAccountRole)
	}

	if apiObject.ExternalId != nil {
		tfMap[names.AttrExternalID] = aws.ToString(apiObject.ExternalId)
	}

	return []any{tfMap}
}

func flattenWorkflows(apiObject []types.Workflow) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, workflow := range apiObject {
		tfMap := map[string]any{
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

func flattenAssociatedAlarms(apiObject map[string]types.AssociatedAlarm) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for alarmName, alarm := range apiObject {
		alarmMap := map[string]any{
			names.AttrName:        alarmName,
			"alarm_type":          string(alarm.AlarmType),
			"resource_identifier": aws.ToString(alarm.ResourceIdentifier),
		}

		if alarm.CrossAccountRole != nil {
			alarmMap["cross_account_role"] = aws.ToString(alarm.CrossAccountRole)
		}

		if alarm.ExternalId != nil {
			alarmMap[names.AttrExternalID] = aws.ToString(alarm.ExternalId)
		}

		tfList = append(tfList, alarmMap)
	}

	return tfList
}

func flattenTriggers(apiObject []types.Trigger) []any {
	if len(apiObject) == 0 {
		return nil
	}

	var tfList []any

	for _, trigger := range apiObject {
		tfMap := map[string]any{
			names.AttrAction:                       string(trigger.Action),
			"min_delay_minutes_between_executions": aws.ToInt32(trigger.MinDelayMinutesBetweenExecutions),
			"target_region":                        aws.ToString(trigger.TargetRegion),
		}

		if trigger.Description != nil {
			tfMap[names.AttrDescription] = aws.ToString(trigger.Description)
		}

		if len(trigger.Conditions) > 0 {
			var conditions []any
			for _, condition := range trigger.Conditions {
				conditionMap := map[string]any{
					"associated_alarm_name": aws.ToString(condition.AssociatedAlarmName),
					names.AttrCondition:     string(condition.Condition),
				}
				conditions = append(conditions, conditionMap)
			}
			tfMap["conditions"] = conditions
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
