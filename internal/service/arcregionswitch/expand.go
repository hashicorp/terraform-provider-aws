// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
)

func expandAssociatedAlarmsFromFramework(alarms []associatedAlarmModel) map[string]awstypes.AssociatedAlarm {
	if len(alarms) == 0 {
		return nil
	}

	result := make(map[string]awstypes.AssociatedAlarm)
	for _, alarm := range alarms {
		awsAlarm := awstypes.AssociatedAlarm{
			AlarmType:          awstypes.AlarmType(alarm.AlarmType.ValueString()),
			ResourceIdentifier: aws.String(alarm.ResourceIdentifier.ValueString()),
		}

		if !alarm.CrossAccountRole.IsNull() && !alarm.CrossAccountRole.IsUnknown() {
			awsAlarm.CrossAccountRole = aws.String(alarm.CrossAccountRole.ValueString())
		}

		if !alarm.ExternalId.IsNull() && !alarm.ExternalId.IsUnknown() {
			awsAlarm.ExternalId = aws.String(alarm.ExternalId.ValueString())
		}

		result[alarm.Name.ValueString()] = awsAlarm
	}
	return result
}

func expandExecutionApprovalConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig {
	if step.ExecutionApprovalConfig.IsNull() {
		return nil
	}

	var approvalConfigs []executionApprovalConfigModel
	step.ExecutionApprovalConfig.ElementsAs(context.Background(), &approvalConfigs, false)

	if len(approvalConfigs) == 0 {
		return nil
	}

	approvalConfig := approvalConfigs[0]
	config := awstypes.ExecutionApprovalConfiguration{
		ApprovalRole: aws.String(approvalConfig.ApprovalRole.ValueString()),
	}
	if !approvalConfig.TimeoutMinutes.IsNull() {
		config.TimeoutMinutes = aws.Int32(int32(approvalConfig.TimeoutMinutes.ValueInt64()))
	}

	return &awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
		Value: config,
	}
}

func expandRoute53HealthCheckConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig {
	if step.Route53HealthCheckConfig.IsNull() {
		return nil
	}

	var healthCheckConfigs []route53HealthCheckConfigModel
	step.Route53HealthCheckConfig.ElementsAs(context.Background(), &healthCheckConfigs, false)

	if len(healthCheckConfigs) == 0 {
		return nil
	}

	healthCheckConfig := healthCheckConfigs[0]
	config := awstypes.Route53HealthCheckConfiguration{
		HostedZoneId: aws.String(healthCheckConfig.HostedZoneId.ValueString()),
		RecordName:   aws.String(healthCheckConfig.RecordName.ValueString()),
	}

	if !healthCheckConfig.CrossAccountRole.IsNull() {
		config.CrossAccountRole = aws.String(healthCheckConfig.CrossAccountRole.ValueString())
	}
	if !healthCheckConfig.ExternalId.IsNull() {
		config.ExternalId = aws.String(healthCheckConfig.ExternalId.ValueString())
	}
	if !healthCheckConfig.TimeoutMinutes.IsNull() {
		config.TimeoutMinutes = aws.Int32(int32(healthCheckConfig.TimeoutMinutes.ValueInt64()))
	}

	if !healthCheckConfig.RecordSets.IsNull() {
		var recordSets []recordSetModel
		healthCheckConfig.RecordSets.ElementsAs(context.Background(), &recordSets, false)

		config.RecordSets = make([]awstypes.Route53ResourceRecordSet, len(recordSets))
		for k, recordSet := range recordSets {
			config.RecordSets[k] = awstypes.Route53ResourceRecordSet{
				RecordSetIdentifier: aws.String(recordSet.RecordSetIdentifier.ValueString()),
				Region:              aws.String(recordSet.Region.ValueString()),
			}
		}
	}

	return &awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{
		Value: config,
	}
}

func expandCustomActionLambdaConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig {
	if step.CustomActionLambdaConfig.IsNull() {
		return nil
	}

	var lambdaConfigs []customActionLambdaConfigModel
	step.CustomActionLambdaConfig.ElementsAs(context.Background(), &lambdaConfigs, false)

	if len(lambdaConfigs) == 0 {
		return nil
	}

	lambdaConfig := lambdaConfigs[0]
	config := awstypes.CustomActionLambdaConfiguration{
		RegionToRun:          awstypes.RegionToRunIn(lambdaConfig.RegionToRun.ValueString()),
		RetryIntervalMinutes: aws.Float32(float32(lambdaConfig.RetryIntervalMinutes.ValueFloat64())),
	}

	if !lambdaConfig.TimeoutMinutes.IsNull() {
		config.TimeoutMinutes = aws.Int32(int32(lambdaConfig.TimeoutMinutes.ValueInt64()))
	}

	if !lambdaConfig.Lambda.IsNull() {
		var lambdas []lambdaModel
		lambdaConfig.Lambda.ElementsAs(context.Background(), &lambdas, false)

		config.Lambdas = make([]awstypes.Lambdas, len(lambdas))
		for k, lambda := range lambdas {
			config.Lambdas[k] = awstypes.Lambdas{
				Arn: aws.String(lambda.ARN.ValueString()),
			}
			if !lambda.CrossAccountRole.IsNull() {
				config.Lambdas[k].CrossAccountRole = aws.String(lambda.CrossAccountRole.ValueString())
			}
			if !lambda.ExternalID.IsNull() {
				config.Lambdas[k].ExternalId = aws.String(lambda.ExternalID.ValueString())
			}
		}
	}

	if !lambdaConfig.Ungraceful.IsNull() {
		var ungracefuls []ungracefulModel
		lambdaConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

		if len(ungracefuls) > 0 {
			config.Ungraceful = &awstypes.LambdaUngraceful{
				Behavior: awstypes.LambdaUngracefulBehavior(ungracefuls[0].Behavior.ValueString()),
			}
		}
	}

	return &awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
		Value: config,
	}
}

func expandGlobalAuroraConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig {
	if step.GlobalAuroraConfig.IsNull() {
		return nil
	}

	var auroraConfigs []globalAuroraConfigModel
	step.GlobalAuroraConfig.ElementsAs(context.Background(), &auroraConfigs, false)

	if len(auroraConfigs) == 0 {
		return nil
	}

	auroraConfig := auroraConfigs[0]
	config := awstypes.GlobalAuroraConfiguration{
		Behavior:                awstypes.GlobalAuroraDefaultBehavior(auroraConfig.Behavior.ValueString()),
		GlobalClusterIdentifier: aws.String(auroraConfig.GlobalClusterIdentifier.ValueString()),
	}

	if !auroraConfig.DatabaseClusterArns.IsNull() {
		var arns []string
		auroraConfig.DatabaseClusterArns.ElementsAs(context.Background(), &arns, false)
		config.DatabaseClusterArns = arns
	}

	if !auroraConfig.CrossAccountRole.IsNull() {
		config.CrossAccountRole = aws.String(auroraConfig.CrossAccountRole.ValueString())
	}
	if !auroraConfig.ExternalId.IsNull() {
		config.ExternalId = aws.String(auroraConfig.ExternalId.ValueString())
	}
	if !auroraConfig.TimeoutMinutes.IsNull() {
		config.TimeoutMinutes = aws.Int32(int32(auroraConfig.TimeoutMinutes.ValueInt64()))
	}

	if !auroraConfig.Ungraceful.IsNull() {
		var ungracefuls []globalAuroraUngracefulModel
		auroraConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

		if len(ungracefuls) > 0 {
			config.Ungraceful = &awstypes.GlobalAuroraUngraceful{
				Ungraceful: awstypes.GlobalAuroraUngracefulBehavior(ungracefuls[0].Ungraceful.ValueString()),
			}
		}
	}

	return &awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig{
		Value: config,
	}
}

func expandEc2AsgCapacityIncreaseConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig {
	if step.Ec2AsgCapacityIncreaseConfig.IsNull() {
		return nil
	}

	var asgConfigs []ec2AsgCapacityIncreaseConfigModel
	step.Ec2AsgCapacityIncreaseConfig.ElementsAs(context.Background(), &asgConfigs, false)

	if len(asgConfigs) == 0 {
		return nil
	}

	asgConfig := asgConfigs[0]
	config := awstypes.Ec2AsgCapacityIncreaseConfiguration{
		CapacityMonitoringApproach: awstypes.Ec2AsgCapacityMonitoringApproach(asgConfig.CapacityMonitoringApproach.ValueString()),
		TargetPercent:              aws.Int32(int32(asgConfig.TargetPercent.ValueInt64())),
	}

	if !asgConfig.TimeoutMinutes.IsNull() {
		config.TimeoutMinutes = aws.Int32(int32(asgConfig.TimeoutMinutes.ValueInt64()))
	}

	if !asgConfig.Asgs.IsNull() {
		var asgs []asgModel
		asgConfig.Asgs.ElementsAs(context.Background(), &asgs, false)

		config.Asgs = make([]awstypes.Asg, len(asgs))
		for k, asg := range asgs {
			config.Asgs[k] = awstypes.Asg{
				Arn: aws.String(asg.ARN.ValueString()),
			}
			if !asg.CrossAccountRole.IsNull() {
				config.Asgs[k].CrossAccountRole = aws.String(asg.CrossAccountRole.ValueString())
			}
			if !asg.ExternalId.IsNull() {
				config.Asgs[k].ExternalId = aws.String(asg.ExternalId.ValueString())
			}
		}
	}

	if !asgConfig.Ungraceful.IsNull() {
		var ungracefuls []ec2UngracefulModel
		asgConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

		if len(ungracefuls) > 0 {
			config.Ungraceful = &awstypes.Ec2Ungraceful{
				MinimumSuccessPercentage: aws.Int32(int32(ungracefuls[0].MinimumSuccessPercentage.ValueInt64())),
			}
		}
	}

	return &awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig{
		Value: config,
	}
}

func expandEcsCapacityIncreaseConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig {
	if step.EcsCapacityIncreaseConfig.IsNull() {
		return nil
	}

	var ecsConfigs []ecsCapacityIncreaseConfigModel
	step.EcsCapacityIncreaseConfig.ElementsAs(context.Background(), &ecsConfigs, false)

	if len(ecsConfigs) == 0 {
		return nil
	}

	ecsConfig := ecsConfigs[0]
	config := awstypes.EcsCapacityIncreaseConfiguration{
		CapacityMonitoringApproach: awstypes.EcsCapacityMonitoringApproach(ecsConfig.CapacityMonitoringApproach.ValueString()),
		TargetPercent:              aws.Int32(int32(ecsConfig.TargetPercent.ValueInt64())),
	}

	if !ecsConfig.TimeoutMinutes.IsNull() {
		config.TimeoutMinutes = aws.Int32(int32(ecsConfig.TimeoutMinutes.ValueInt64()))
	}

	if !ecsConfig.Services.IsNull() {
		var services []serviceModel
		ecsConfig.Services.ElementsAs(context.Background(), &services, false)

		config.Services = make([]awstypes.Service, len(services))
		for k, service := range services {
			config.Services[k] = awstypes.Service{
				ClusterArn: aws.String(service.ClusterArn.ValueString()),
				ServiceArn: aws.String(service.ServiceArn.ValueString()),
			}
			if !service.CrossAccountRole.IsNull() {
				config.Services[k].CrossAccountRole = aws.String(service.CrossAccountRole.ValueString())
			}
			if !service.ExternalId.IsNull() {
				config.Services[k].ExternalId = aws.String(service.ExternalId.ValueString())
			}
		}
	}

	if !ecsConfig.Ungraceful.IsNull() {
		var ungracefuls []ecsUngracefulModel
		ecsConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

		if len(ungracefuls) > 0 {
			config.Ungraceful = &awstypes.EcsUngraceful{
				MinimumSuccessPercentage: aws.Int32(int32(ungracefuls[0].MinimumSuccessPercentage.ValueInt64())),
			}
		}
	}

	return &awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig{
		Value: config,
	}
}

func expandEksResourceScalingConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig {
	if step.EksResourceScalingConfig.IsNull() {
		return nil
	}

	var eksConfigs []eksResourceScalingConfigModel
	step.EksResourceScalingConfig.ElementsAs(context.Background(), &eksConfigs, false)

	if len(eksConfigs) == 0 {
		return nil
	}

	eksConfig := eksConfigs[0]
	config := awstypes.EksResourceScalingConfiguration{
		CapacityMonitoringApproach: awstypes.EksCapacityMonitoringApproach(eksConfig.CapacityMonitoringApproach.ValueString()),
		TargetPercent:              aws.Int32(int32(eksConfig.TargetPercent.ValueInt64())),
	}

	if !eksConfig.TimeoutMinutes.IsNull() {
		config.TimeoutMinutes = aws.Int32(int32(eksConfig.TimeoutMinutes.ValueInt64()))
	}

	if !eksConfig.KubernetesResourceType.IsNull() {
		var resourceTypes []kubernetesResourceTypeModel
		eksConfig.KubernetesResourceType.ElementsAs(context.Background(), &resourceTypes, false)

		if len(resourceTypes) > 0 {
			config.KubernetesResourceType = &awstypes.KubernetesResourceType{
				ApiVersion: aws.String(resourceTypes[0].ApiVersion.ValueString()),
				Kind:       aws.String(resourceTypes[0].Kind.ValueString()),
			}
		}
	}

	if !eksConfig.EksClusters.IsNull() {
		var clusters []eksClusterModel
		eksConfig.EksClusters.ElementsAs(context.Background(), &clusters, false)

		config.EksClusters = make([]awstypes.EksCluster, len(clusters))
		for k, cluster := range clusters {
			config.EksClusters[k] = awstypes.EksCluster{
				ClusterArn: aws.String(cluster.ClusterArn.ValueString()),
			}
			if !cluster.CrossAccountRole.IsNull() {
				config.EksClusters[k].CrossAccountRole = aws.String(cluster.CrossAccountRole.ValueString())
			}
			if !cluster.ExternalId.IsNull() {
				config.EksClusters[k].ExternalId = aws.String(cluster.ExternalId.ValueString())
			}
		}
	}

	if !eksConfig.ScalingResources.IsNull() {
		var scalingResources []scalingResourcesModel
		eksConfig.ScalingResources.ElementsAs(context.Background(), &scalingResources, false)

		config.ScalingResources = make([]map[string]map[string]awstypes.KubernetesScalingResource, len(scalingResources))
		for k, scalingResource := range scalingResources {
			if !scalingResource.Resources.IsNull() {
				var resources []kubernetesScalingResourceModel
				scalingResource.Resources.ElementsAs(context.Background(), &resources, false)

				regionMap := make(map[string]awstypes.KubernetesScalingResource)
				for _, resource := range resources {
					scalingResource := awstypes.KubernetesScalingResource{
						Name:      aws.String(resource.Name.ValueString()),
						Namespace: aws.String(resource.Namespace.ValueString()),
					}
					if !resource.HpaName.IsNull() {
						scalingResource.HpaName = aws.String(resource.HpaName.ValueString())
					}
					regionMap[resource.ResourceName.ValueString()] = scalingResource
				}
				config.ScalingResources[k] = map[string]map[string]awstypes.KubernetesScalingResource{
					scalingResource.Namespace.ValueString(): regionMap,
				}
			}
		}
	}

	if !eksConfig.Ungraceful.IsNull() {
		var ungracefuls []eksUngracefulModel
		eksConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

		if len(ungracefuls) > 0 {
			config.Ungraceful = &awstypes.EksResourceScalingUngraceful{
				MinimumSuccessPercentage: aws.Int32(int32(ungracefuls[0].MinimumSuccessPercentage.ValueInt64())),
			}
		}
	}

	return &awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig{
		Value: config,
	}
}

func expandArcRoutingControlConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig {
	if step.ArcRoutingControlConfig.IsNull() {
		return nil
	}

	var routingConfigs []arcRoutingControlConfigModel
	step.ArcRoutingControlConfig.ElementsAs(context.Background(), &routingConfigs, false)

	if len(routingConfigs) == 0 {
		return nil
	}

	routingConfig := routingConfigs[0]
	config := awstypes.ArcRoutingControlConfiguration{
		RegionAndRoutingControls: make(map[string][]awstypes.ArcRoutingControlState),
	}

	if !routingConfig.CrossAccountRole.IsNull() {
		config.CrossAccountRole = aws.String(routingConfig.CrossAccountRole.ValueString())
	}
	if !routingConfig.ExternalId.IsNull() {
		config.ExternalId = aws.String(routingConfig.ExternalId.ValueString())
	}
	if !routingConfig.TimeoutMinutes.IsNull() {
		config.TimeoutMinutes = aws.Int32(int32(routingConfig.TimeoutMinutes.ValueInt64()))
	}

	if !routingConfig.RegionAndRoutingControls.IsNull() {
		var regionControls []regionAndRoutingControlsModel
		routingConfig.RegionAndRoutingControls.ElementsAs(context.Background(), &regionControls, false)

		for _, regionControl := range regionControls {
			var controlArns []string
			regionControl.RoutingControlArns.ElementsAs(context.Background(), &controlArns, false)

			controlStates := make([]awstypes.ArcRoutingControlState, len(controlArns))
			for l, arn := range controlArns {
				controlStates[l] = awstypes.ArcRoutingControlState{
					RoutingControlArn: aws.String(arn),
					State:             awstypes.RoutingControlStateChangeOn,
				}
			}
			config.RegionAndRoutingControls[regionControl.Region.ValueString()] = controlStates
		}
	}

	return &awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig{
		Value: config,
	}
}

func expandParallelConfig(step stepModel) *awstypes.ExecutionBlockConfigurationMemberParallelConfig {
	if step.ParallelConfig.IsNull() {
		return nil
	}

	var parallelConfigs []parallelConfigModel
	step.ParallelConfig.ElementsAs(context.Background(), &parallelConfigs, false)

	if len(parallelConfigs) == 0 {
		return nil
	}

	parallelConfig := parallelConfigs[0]
	config := awstypes.ParallelExecutionBlockConfiguration{}

	if !parallelConfig.Step.IsNull() {
		var parallelSteps []parallelStepModel
		parallelConfig.Step.ElementsAs(context.Background(), &parallelSteps, false)

		config.Steps = make([]awstypes.Step, len(parallelSteps))
		for k, parallelStep := range parallelSteps {
			config.Steps[k] = awstypes.Step{
				Name:               aws.String(parallelStep.Name.ValueString()),
				ExecutionBlockType: awstypes.ExecutionBlockType(parallelStep.ExecutionBlockType.ValueString()),
			}

			if !parallelStep.Description.IsNull() {
				config.Steps[k].Description = aws.String(parallelStep.Description.ValueString())
			}

			if !parallelStep.ExecutionApprovalConfig.IsNull() {
				var approvalConfigs []executionApprovalConfigModel
				parallelStep.ExecutionApprovalConfig.ElementsAs(context.Background(), &approvalConfigs, false)

				if len(approvalConfigs) > 0 {
					approvalConfig := approvalConfigs[0]
					config.Steps[k].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
						Value: awstypes.ExecutionApprovalConfiguration{
							ApprovalRole: aws.String(approvalConfig.ApprovalRole.ValueString()),
						},
					}
					if !approvalConfig.TimeoutMinutes.IsNull() {
						config.Steps[k].ExecutionBlockConfiguration.(*awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig).Value.TimeoutMinutes = aws.Int32(int32(approvalConfig.TimeoutMinutes.ValueInt64()))
					}
				}
			} else if !parallelStep.CustomActionLambdaConfig.IsNull() {
				var lambdaConfigs []customActionLambdaConfigModel
				parallelStep.CustomActionLambdaConfig.ElementsAs(context.Background(), &lambdaConfigs, false)

				if len(lambdaConfigs) > 0 {
					lambdaConfig := lambdaConfigs[0]
					lambdaConfigValue := awstypes.CustomActionLambdaConfiguration{
						RegionToRun:          awstypes.RegionToRunIn(lambdaConfig.RegionToRun.ValueString()),
						RetryIntervalMinutes: aws.Float32(float32(lambdaConfig.RetryIntervalMinutes.ValueFloat64())),
					}

					if !lambdaConfig.TimeoutMinutes.IsNull() {
						lambdaConfigValue.TimeoutMinutes = aws.Int32(int32(lambdaConfig.TimeoutMinutes.ValueInt64()))
					}

					if !lambdaConfig.Lambda.IsNull() {
						var lambdas []lambdaModel
						lambdaConfig.Lambda.ElementsAs(context.Background(), &lambdas, false)

						lambdaConfigValue.Lambdas = make([]awstypes.Lambdas, len(lambdas))
						for l, lambda := range lambdas {
							lambdaConfigValue.Lambdas[l] = awstypes.Lambdas{
								Arn: aws.String(lambda.ARN.ValueString()),
							}
							if !lambda.CrossAccountRole.IsNull() {
								lambdaConfigValue.Lambdas[l].CrossAccountRole = aws.String(lambda.CrossAccountRole.ValueString())
							}
							if !lambda.ExternalID.IsNull() {
								lambdaConfigValue.Lambdas[l].ExternalId = aws.String(lambda.ExternalID.ValueString())
							}
						}
					}

					config.Steps[k].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
						Value: lambdaConfigValue,
					}
				}
			}
		}
	}

	return &awstypes.ExecutionBlockConfigurationMemberParallelConfig{
		Value: config,
	}
}

func expandWorkflowsFromFramework(workflows []workflowModel) []awstypes.Workflow {
	if len(workflows) == 0 {
		return nil
	}

	result := make([]awstypes.Workflow, len(workflows))
	for i, workflow := range workflows {
		result[i] = awstypes.Workflow{
			WorkflowTargetAction: awstypes.WorkflowTargetAction(workflow.WorkflowTargetAction.ValueString()),
		}

		if !workflow.WorkflowTargetRegion.IsNull() {
			result[i].WorkflowTargetRegion = aws.String(workflow.WorkflowTargetRegion.ValueString())
		}

		if !workflow.WorkflowDescription.IsNull() {
			result[i].WorkflowDescription = aws.String(workflow.WorkflowDescription.ValueString())
		}

		// Handle steps
		if !workflow.Step.IsNull() {
			var steps []stepModel
			workflow.Step.ElementsAs(context.Background(), &steps, false)

			result[i].Steps = make([]awstypes.Step, len(steps))
			for j, step := range steps {
				result[i].Steps[j] = awstypes.Step{
					Name:               aws.String(step.Name.ValueString()),
					ExecutionBlockType: awstypes.ExecutionBlockType(step.ExecutionBlockType.ValueString()),
				}

				if !step.Description.IsNull() {
					result[i].Steps[j].Description = aws.String(step.Description.ValueString())
				}

				// Handle execution block configurations using abstracted functions
				if config := expandExecutionApprovalConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				} else if config := expandRoute53HealthCheckConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				} else if config := expandCustomActionLambdaConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				} else if config := expandGlobalAuroraConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				} else if config := expandEc2AsgCapacityIncreaseConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				} else if config := expandEcsCapacityIncreaseConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				} else if config := expandEksResourceScalingConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				} else if config := expandArcRoutingControlConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				} else if config := expandParallelConfig(step); config != nil {
					result[i].Steps[j].ExecutionBlockConfiguration = config
				}
			}
		} else {
			result[i].Steps = []awstypes.Step{}
		}
	}

	return result
}
