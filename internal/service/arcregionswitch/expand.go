// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandWorkflows(tfList []any) []types.Workflow {
	if len(tfList) == 0 {
		return nil
	}

	var workflows []types.Workflow

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		workflow := types.Workflow{
			WorkflowTargetAction: types.WorkflowTargetAction(tfMap["workflow_target_action"].(string)),
		}

		if v, ok := tfMap["workflow_target_region"].(string); ok && v != "" {
			workflow.WorkflowTargetRegion = aws.String(v)
		}

		if v, ok := tfMap["workflow_description"].(string); ok && v != "" {
			workflow.WorkflowDescription = aws.String(v)
		}

		if v, ok := tfMap["step"].([]any); ok && len(v) > 0 {
			workflow.Steps = expandSteps(v, string(workflow.WorkflowTargetAction))
		}

		workflows = append(workflows, workflow)
	}

	return workflows
}

func expandSteps(tfList []any, workflowAction string) []types.Step {
	if len(tfList) == 0 {
		return nil
	}

	var steps []types.Step

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		step := types.Step{
			Name:               aws.String(tfMap[names.AttrName].(string)),
			ExecutionBlockType: types.ExecutionBlockType(tfMap["execution_block_type"].(string)),
		}

		if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
			step.Description = aws.String(v)
		}

		// Set the execution block configuration based on the block type
		blockType := tfMap["execution_block_type"].(string)
		switch blockType {
		case "ManualApproval":
			if v, ok := tfMap["execution_approval_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandExecutionApprovalConfig(v[0].(map[string]any))
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
					Value: config,
				}
			}
		case "CustomActionLambda":
			if v, ok := tfMap["custom_action_lambda_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandCustomActionLambdaConfig(v[0].(map[string]any))
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
					Value: config,
				}
			}
		case "ARCRoutingControl":
			if v, ok := tfMap["arc_routing_control_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandArcRoutingControlConfig(v[0].(map[string]any), workflowAction)
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberArcRoutingControlConfig{
					Value: config,
				}
			}
		case "EC2AutoScaling":
			if v, ok := tfMap["ec2_asg_capacity_increase_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandEC2ASGCapacityIncreaseConfig(v[0].(map[string]any))
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig{
					Value: config,
				}
			}
		case "AuroraGlobalDatabase":
			if v, ok := tfMap["global_aurora_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandGlobalAuroraConfig(v[0].(map[string]any))
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberGlobalAuroraConfig{
					Value: config,
				}
			}
		case "Parallel":
			if v, ok := tfMap["parallel_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandParallelConfig(v[0].(map[string]any), workflowAction)
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberParallelConfig{
					Value: config,
				}
			}

		case "ECSServiceScaling":
			if v, ok := tfMap["ecs_capacity_increase_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandECSCapacityIncreaseConfig(v[0].(map[string]any))
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig{
					Value: config,
				}
			}
		case "EKSResourceScaling":
			if v, ok := tfMap["eks_resource_scaling_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandEKSResourceScalingConfig(v[0].(map[string]any))
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberEksResourceScalingConfig{
					Value: config,
				}
			}
		case "Route53HealthCheck":
			if v, ok := tfMap["route53_health_check_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandRoute53HealthCheckConfig(v[0].(map[string]any))
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{
					Value: config,
				}
			}
		case "ARCRegionSwitchPlan":
			if v, ok := tfMap["region_switch_plan_config"].([]any); ok && len(v) > 0 && v[0] != nil {
				config := expandRegionSwitchPlanConfig(v[0].(map[string]any))
				step.ExecutionBlockConfiguration = &types.ExecutionBlockConfigurationMemberRegionSwitchPlanConfig{
					Value: config,
				}
			}
		}

		steps = append(steps, step)
	}

	return steps
}

func expandExecutionApprovalConfig(tfMap map[string]any) types.ExecutionApprovalConfiguration {
	config := types.ExecutionApprovalConfiguration{
		ApprovalRole: aws.String(tfMap["approval_role"].(string)),
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v > 0 {
		config.TimeoutMinutes = aws.Int32(int32(v))
	}

	return config
}

func expandCustomActionLambdaConfig(tfMap map[string]any) types.CustomActionLambdaConfiguration {
	config := types.CustomActionLambdaConfiguration{
		RegionToRun:          types.RegionToRunIn(tfMap["region_to_run"].(string)),
		RetryIntervalMinutes: aws.Float32(float32(tfMap["retry_interval_minutes"].(float64))),
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v > 0 {
		config.TimeoutMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["lambda"].([]any); ok && len(v) > 0 {
		config.Lambdas = expandLambdas(v)
	}

	if v, ok := tfMap["ungraceful"].([]any); ok && len(v) > 0 && v[0] != nil {
		ungracefulMap := v[0].(map[string]any)
		config.Ungraceful = &types.LambdaUngraceful{
			Behavior: types.LambdaUngracefulBehavior(ungracefulMap["behavior"].(string)),
		}
	}

	return config
}

func expandLambdas(tfList []any) []types.Lambdas {
	if len(tfList) == 0 {
		return nil
	}

	var lambdas []types.Lambdas

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		lambda := types.Lambdas{
			Arn: aws.String(tfMap[names.AttrARN].(string)),
		}

		if v, ok := tfMap["cross_account_role"].(string); ok && v != "" {
			lambda.CrossAccountRole = aws.String(v)
		}

		if v, ok := tfMap[names.AttrExternalID].(string); ok && v != "" {
			lambda.ExternalId = aws.String(v)
		}

		lambdas = append(lambdas, lambda)
	}

	return lambdas
}

func expandArcRoutingControlConfig(tfMap map[string]any, workflowAction string) types.ArcRoutingControlConfiguration {
	config := types.ArcRoutingControlConfiguration{
		RegionAndRoutingControls: make(map[string][]types.ArcRoutingControlState),
	}

	// Handle region_and_routing_controls list
	// Schema: TypeList of objects with region and routing_control_arns
	if v, ok := tfMap["region_and_routing_controls"].([]any); ok && len(v) > 0 {
		for _, regionControlsInterface := range v {
			if regionControlsMap, ok := regionControlsInterface.(map[string]any); ok {
				region := regionControlsMap[names.AttrRegion].(string)
				if controlsInterface, ok := regionControlsMap["routing_control_arns"].([]any); ok && len(controlsInterface) > 0 {
					var routingControlStates []types.ArcRoutingControlState
					for _, controlInterface := range controlsInterface {
						if controlArn, ok := controlInterface.(string); ok && controlArn != "" {
							// Determine routing control state based on workflow action:
							// - "activate" workflows set controls to "On" to direct traffic to the target region
							// - "deactivate" workflows set controls to "Off" to stop traffic to the target region
							var controlState types.RoutingControlStateChange
							if workflowAction == "activate" {
								controlState = types.RoutingControlStateChangeOn
							} else {
								controlState = types.RoutingControlStateChangeOff
							}

							state := types.ArcRoutingControlState{
								RoutingControlArn: aws.String(controlArn),
								State:             controlState,
							}
							routingControlStates = append(routingControlStates, state)
						}
					}
					config.RegionAndRoutingControls[region] = routingControlStates
				}
			}
		}
	}

	if v, ok := tfMap["cross_account_role"].(string); ok && v != "" {
		config.CrossAccountRole = aws.String(v)
	}

	if v, ok := tfMap[names.AttrExternalID].(string); ok && v != "" {
		config.ExternalId = aws.String(v)
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v > 0 {
		config.TimeoutMinutes = aws.Int32(int32(v))
	}

	return config
}

func expandEC2ASGCapacityIncreaseConfig(tfMap map[string]any) types.Ec2AsgCapacityIncreaseConfiguration {
	config := types.Ec2AsgCapacityIncreaseConfiguration{}

	if v, ok := tfMap["asgs"].([]any); ok && len(v) > 0 {
		var asgs []types.Asg
		for _, asgRaw := range v {
			asgMap := asgRaw.(map[string]any)
			asg := types.Asg{
				Arn: aws.String(asgMap[names.AttrARN].(string)),
			}

			if v, ok := asgMap["cross_account_role"].(string); ok && v != "" {
				asg.CrossAccountRole = aws.String(v)
			}

			if v, ok := asgMap[names.AttrExternalID].(string); ok && v != "" {
				asg.ExternalId = aws.String(v)
			}

			asgs = append(asgs, asg)
		}
		config.Asgs = asgs
	}

	if v, ok := tfMap["capacity_monitoring_approach"].(string); ok && v != "" {
		config.CapacityMonitoringApproach = types.Ec2AsgCapacityMonitoringApproach(v)
	}

	if v, ok := tfMap["target_percent"].(int); ok && v > 0 {
		config.TargetPercent = aws.Int32(int32(v))
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v > 0 {
		config.TimeoutMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["ungraceful"].([]any); ok && len(v) > 0 && v[0] != nil {
		ungracefulMap := v[0].(map[string]any)
		config.Ungraceful = &types.Ec2Ungraceful{
			MinimumSuccessPercentage: aws.Int32(int32(ungracefulMap["minimum_success_percentage"].(int))),
		}
	}

	return config
}

func expandGlobalAuroraConfig(tfMap map[string]any) types.GlobalAuroraConfiguration {
	config := types.GlobalAuroraConfiguration{
		Behavior:                types.GlobalAuroraDefaultBehavior(tfMap["behavior"].(string)),
		GlobalClusterIdentifier: aws.String(tfMap["global_cluster_identifier"].(string)),
	}

	if v, ok := tfMap["database_cluster_arns"].([]any); ok && len(v) > 0 {
		config.DatabaseClusterArns = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["cross_account_role"].(string); ok && v != "" {
		config.CrossAccountRole = aws.String(v)
	}

	if v, ok := tfMap[names.AttrExternalID].(string); ok && v != "" {
		config.ExternalId = aws.String(v)
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v > 0 {
		config.TimeoutMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["ungraceful"].([]any); ok && len(v) > 0 && v[0] != nil {
		ungracefulMap := v[0].(map[string]any)
		config.Ungraceful = &types.GlobalAuroraUngraceful{
			Ungraceful: types.GlobalAuroraUngracefulBehavior(ungracefulMap["ungraceful"].(string)),
		}
	}

	return config
}

func expandECSCapacityIncreaseConfig(tfMap map[string]any) types.EcsCapacityIncreaseConfiguration {
	config := types.EcsCapacityIncreaseConfiguration{}

	if v, ok := tfMap["services"].([]any); ok && len(v) > 0 {
		var services []types.Service
		for _, serviceRaw := range v {
			serviceMap := serviceRaw.(map[string]any)
			service := types.Service{}

			if v, ok := serviceMap["cluster_arn"].(string); ok && v != "" {
				service.ClusterArn = aws.String(v)
			}

			if v, ok := serviceMap["service_arn"].(string); ok && v != "" {
				service.ServiceArn = aws.String(v)
			}

			if v, ok := serviceMap["cross_account_role"].(string); ok && v != "" {
				service.CrossAccountRole = aws.String(v)
			}

			if v, ok := serviceMap[names.AttrExternalID].(string); ok && v != "" {
				service.ExternalId = aws.String(v)
			}

			services = append(services, service)
		}
		config.Services = services
	}

	if v, ok := tfMap["capacity_monitoring_approach"].(string); ok && v != "" {
		config.CapacityMonitoringApproach = types.EcsCapacityMonitoringApproach(v)
	}

	if v, ok := tfMap["target_percent"].(int); ok && v > 0 {
		config.TargetPercent = aws.Int32(int32(v))
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v > 0 {
		config.TimeoutMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["ungraceful"].([]any); ok && len(v) > 0 && v[0] != nil {
		ungracefulMap := v[0].(map[string]any)
		config.Ungraceful = &types.EcsUngraceful{
			MinimumSuccessPercentage: aws.Int32(int32(ungracefulMap["minimum_success_percentage"].(int))),
		}
	}

	return config
}

func expandEKSResourceScalingConfig(tfMap map[string]any) types.EksResourceScalingConfiguration {
	config := types.EksResourceScalingConfiguration{}

	if v, ok := tfMap["kubernetes_resource_type"].([]any); ok && len(v) > 0 && v[0] != nil {
		resourceTypeMap := v[0].(map[string]any)
		config.KubernetesResourceType = &types.KubernetesResourceType{
			ApiVersion: aws.String(resourceTypeMap["api_version"].(string)),
			Kind:       aws.String(resourceTypeMap["kind"].(string)),
		}
	}

	if v, ok := tfMap["eks_clusters"].([]any); ok && len(v) > 0 {
		var clusters []types.EksCluster
		for _, clusterRaw := range v {
			clusterMap := clusterRaw.(map[string]any)
			cluster := types.EksCluster{
				ClusterArn: aws.String(clusterMap["cluster_arn"].(string)),
			}

			if v, ok := clusterMap["cross_account_role"].(string); ok && v != "" {
				cluster.CrossAccountRole = aws.String(v)
			}

			if v, ok := clusterMap[names.AttrExternalID].(string); ok && v != "" {
				cluster.ExternalId = aws.String(v)
			}

			clusters = append(clusters, cluster)
		}
		config.EksClusters = clusters
	}

	if v, ok := tfMap["capacity_monitoring_approach"].(string); ok && v != "" {
		config.CapacityMonitoringApproach = types.EksCapacityMonitoringApproach(v)
	}

	if v, ok := tfMap["target_percent"].(int); ok && v > 0 {
		config.TargetPercent = aws.Int32(int32(v))
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v > 0 {
		config.TimeoutMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["ungraceful"].([]any); ok && len(v) > 0 && v[0] != nil {
		ungracefulMap := v[0].(map[string]any)
		config.Ungraceful = &types.EksResourceScalingUngraceful{
			MinimumSuccessPercentage: aws.Int32(int32(ungracefulMap["minimum_success_percentage"].(int))),
		}
	}

	// Handle scaling_resources - complex nested map structure
	if v, ok := tfMap["scaling_resources"].([]any); ok && len(v) > 0 {
		scalingResources := make([]map[string]map[string]types.KubernetesScalingResource, len(v))
		for i, resourceRaw := range v {
			resourceMap := resourceRaw.(map[string]any)
			namespace := resourceMap[names.AttrNamespace].(string)

			if resourcesRaw, ok := resourceMap[names.AttrResources].(*schema.Set); ok {
				namespaceMap := make(map[string]types.KubernetesScalingResource)
				for _, resourceDataRaw := range resourcesRaw.List() {
					if resourceData, ok := resourceDataRaw.(map[string]any); ok {
						resourceName := resourceData["resource_name"].(string)
						kubernetesResource := types.KubernetesScalingResource{
							Name:      aws.String(resourceData[names.AttrName].(string)),
							Namespace: aws.String(resourceData[names.AttrNamespace].(string)),
						}
						if hpaName, ok := resourceData["hpa_name"].(string); ok && hpaName != "" {
							kubernetesResource.HpaName = aws.String(hpaName)
						}
						namespaceMap[resourceName] = kubernetesResource
					}
				}
				scalingResources[i] = map[string]map[string]types.KubernetesScalingResource{
					namespace: namespaceMap,
				}
			}
		}
		config.ScalingResources = scalingResources
	}

	return config
}

func expandRoute53HealthCheckConfig(tfMap map[string]any) types.Route53HealthCheckConfiguration {
	config := types.Route53HealthCheckConfiguration{
		HostedZoneId: aws.String(tfMap[names.AttrHostedZoneID].(string)),
		RecordName:   aws.String(tfMap["record_name"].(string)),
	}

	if v, ok := tfMap["cross_account_role"].(string); ok && v != "" {
		config.CrossAccountRole = aws.String(v)
	}

	if v, ok := tfMap[names.AttrExternalID].(string); ok && v != "" {
		config.ExternalId = aws.String(v)
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v > 0 {
		config.TimeoutMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["record_sets"].([]any); ok && len(v) > 0 {
		var recordSets []types.Route53ResourceRecordSet
		for _, recordSetRaw := range v {
			recordSetMap := recordSetRaw.(map[string]any)
			recordSet := types.Route53ResourceRecordSet{}

			if v, ok := recordSetMap["record_set_identifier"].(string); ok && v != "" {
				recordSet.RecordSetIdentifier = aws.String(v)
			}

			if v, ok := recordSetMap[names.AttrRegion].(string); ok && v != "" {
				recordSet.Region = aws.String(v)
			}

			recordSets = append(recordSets, recordSet)
		}
		config.RecordSets = recordSets
	}

	return config
}

func expandRegionSwitchPlanConfig(tfMap map[string]any) types.RegionSwitchPlanConfiguration {
	config := types.RegionSwitchPlanConfiguration{
		Arn: aws.String(tfMap[names.AttrARN].(string)),
	}

	if v, ok := tfMap["cross_account_role"].(string); ok && v != "" {
		config.CrossAccountRole = aws.String(v)
	}

	if v, ok := tfMap[names.AttrExternalID].(string); ok && v != "" {
		config.ExternalId = aws.String(v)
	}

	return config
}

func expandParallelConfig(tfMap map[string]any, workflowAction string) types.ParallelExecutionBlockConfiguration {
	config := types.ParallelExecutionBlockConfiguration{}

	if v, ok := tfMap["step"].([]any); ok && len(v) > 0 {
		config.Steps = expandSteps(v, workflowAction)
	}

	return config
}

func expandAssociatedAlarms(tfList []any) map[string]types.AssociatedAlarm {
	if len(tfList) == 0 {
		return nil
	}

	result := make(map[string]types.AssociatedAlarm)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		name := tfMap[names.AttrName].(string)
		alarm := types.AssociatedAlarm{
			AlarmType:          types.AlarmType(tfMap["alarm_type"].(string)),
			ResourceIdentifier: aws.String(tfMap["resource_identifier"].(string)),
		}

		if v, ok := tfMap["cross_account_role"].(string); ok && v != "" {
			alarm.CrossAccountRole = aws.String(v)
		}

		if v, ok := tfMap[names.AttrExternalID].(string); ok && v != "" {
			alarm.ExternalId = aws.String(v)
		}

		result[name] = alarm
	}

	return result
}

func expandTriggers(tfList []any) []types.Trigger {
	if len(tfList) == 0 {
		return nil
	}

	var triggers []types.Trigger

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		trigger := types.Trigger{
			Action:                           types.WorkflowTargetAction(tfMap[names.AttrAction].(string)),
			MinDelayMinutesBetweenExecutions: aws.Int32(int32(tfMap["min_delay_minutes_between_executions"].(int))),
			TargetRegion:                     aws.String(tfMap["target_region"].(string)),
		}

		if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
			trigger.Description = aws.String(v)
		}

		if v, ok := tfMap["conditions"].([]any); ok && len(v) > 0 {
			var conditions []types.TriggerCondition
			for _, condRaw := range v {
				condMap := condRaw.(map[string]any)
				condition := types.TriggerCondition{
					AssociatedAlarmName: aws.String(condMap["associated_alarm_name"].(string)),
					Condition:           types.AlarmCondition(condMap[names.AttrCondition].(string)),
				}
				conditions = append(conditions, condition)
			}
			trigger.Conditions = conditions
		}

		triggers = append(triggers, trigger)
	}

	return triggers
}
