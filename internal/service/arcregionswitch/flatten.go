// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func flattenExecutionApprovalConfig(config *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig, diags *fwdiag.Diagnostics) types.List {
	approvalAttrs := map[string]attr.Value{
		"approval_role": types.StringValue(aws.ToString(config.Value.ApprovalRole)),
	}
	if config.Value.TimeoutMinutes != nil {
		approvalAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
	} else {
		approvalAttrs["timeout_minutes"] = types.Int64Null()
	}

	approvalObj, approvalDiags := types.ObjectValue(getExecutionApprovalConfigObjectType().AttrTypes, approvalAttrs)
	diags.Append(approvalDiags...)
	return types.ListValueMust(getExecutionApprovalConfigObjectType(), []attr.Value{approvalObj})
}

func flattenRoute53HealthCheckConfig(config *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig, diags *fwdiag.Diagnostics) types.List {
	healthCheckAttrs := map[string]attr.Value{
		"hosted_zone_id": types.StringValue(aws.ToString(config.Value.HostedZoneId)),
		"record_name":    types.StringValue(aws.ToString(config.Value.RecordName)),
	}

	if config.Value.CrossAccountRole != nil {
		healthCheckAttrs["cross_account_role"] = types.StringValue(aws.ToString(config.Value.CrossAccountRole))
	} else {
		healthCheckAttrs["cross_account_role"] = types.StringNull()
	}

	if config.Value.ExternalId != nil {
		healthCheckAttrs["external_id"] = types.StringValue(aws.ToString(config.Value.ExternalId))
	} else {
		healthCheckAttrs["external_id"] = types.StringNull()
	}

	if config.Value.TimeoutMinutes != nil {
		healthCheckAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
	} else {
		healthCheckAttrs["timeout_minutes"] = types.Int64Null()
	}

	if len(config.Value.RecordSets) > 0 {
		recordSetElements := make([]attr.Value, len(config.Value.RecordSets))
		for k, recordSet := range config.Value.RecordSets {
			recordSetAttrs := map[string]attr.Value{
				"record_set_identifier": types.StringValue(aws.ToString(recordSet.RecordSetIdentifier)),
				"region":                types.StringValue(aws.ToString(recordSet.Region)),
			}
			recordSetObj, recordSetDiags := types.ObjectValue(getRecordSetObjectType().AttrTypes, recordSetAttrs)
			diags.Append(recordSetDiags...)
			recordSetElements[k] = recordSetObj
		}
		healthCheckAttrs["record_sets"] = types.ListValueMust(getRecordSetObjectType(), recordSetElements)
	} else {
		healthCheckAttrs["record_sets"] = types.ListNull(getRecordSetObjectType())
	}

	healthCheckObj, healthCheckDiags := types.ObjectValue(getRoute53HealthCheckConfigObjectType().AttrTypes, healthCheckAttrs)
	diags.Append(healthCheckDiags...)
	return types.ListValueMust(getRoute53HealthCheckConfigObjectType(), []attr.Value{healthCheckObj})
}

func flattenCustomActionLambdaConfig(config *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig, diags *fwdiag.Diagnostics) types.List {
	lambdaAttrs := map[string]attr.Value{
		"region_to_run":          types.StringValue(string(config.Value.RegionToRun)),
		"retry_interval_minutes": types.Float64Value(float64(aws.ToFloat32(config.Value.RetryIntervalMinutes))),
	}

	if config.Value.TimeoutMinutes != nil {
		lambdaAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
	} else {
		lambdaAttrs["timeout_minutes"] = types.Int64Null()
	}

	if len(config.Value.Lambdas) > 0 {
		lambdaElements := make([]attr.Value, len(config.Value.Lambdas))
		for l, lambda := range config.Value.Lambdas {
			lambdaElementAttrs := map[string]attr.Value{
				"arn": types.StringValue(aws.ToString(lambda.Arn)),
			}
			if lambda.CrossAccountRole != nil {
				lambdaElementAttrs["cross_account_role"] = types.StringValue(aws.ToString(lambda.CrossAccountRole))
			} else {
				lambdaElementAttrs["cross_account_role"] = types.StringNull()
			}
			if lambda.ExternalId != nil {
				lambdaElementAttrs["external_id"] = types.StringValue(aws.ToString(lambda.ExternalId))
			} else {
				lambdaElementAttrs["external_id"] = types.StringNull()
			}

			lambdaObj, lambdaDiags := types.ObjectValue(getLambdaObjectType().AttrTypes, lambdaElementAttrs)
			diags.Append(lambdaDiags...)
			lambdaElements[l] = lambdaObj
		}
		lambdaAttrs["lambda"] = types.ListValueMust(getLambdaObjectType(), lambdaElements)
	} else {
		lambdaAttrs["lambda"] = types.ListNull(getLambdaObjectType())
	}

	if config.Value.Ungraceful != nil {
		ungracefulAttrs := map[string]attr.Value{
			"behavior": types.StringValue(string(config.Value.Ungraceful.Behavior)),
		}
		ungracefulObj, ungracefulDiags := types.ObjectValue(getUngracefulObjectType().AttrTypes, ungracefulAttrs)
		diags.Append(ungracefulDiags...)
		lambdaAttrs["ungraceful"] = types.ListValueMust(getUngracefulObjectType(), []attr.Value{ungracefulObj})
	} else {
		lambdaAttrs["ungraceful"] = types.ListNull(getUngracefulObjectType())
	}

	lambdaObj, lambdaDiags := types.ObjectValue(getCustomActionLambdaConfigObjectType().AttrTypes, lambdaAttrs)
	diags.Append(lambdaDiags...)
	return types.ListValueMust(getCustomActionLambdaConfigObjectType(), []attr.Value{lambdaObj})
}

func flattenGlobalAuroraConfig(ctx context.Context, config *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig, diags *fwdiag.Diagnostics) types.List {
	auroraAttrs := map[string]attr.Value{
		"behavior":                  types.StringValue(string(config.Value.Behavior)),
		"global_cluster_identifier": types.StringValue(aws.ToString(config.Value.GlobalClusterIdentifier)),
	}

	if config.Value.TimeoutMinutes != nil {
		auroraAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
	} else {
		auroraAttrs["timeout_minutes"] = types.Int64Null()
	}

	if config.Value.CrossAccountRole != nil {
		auroraAttrs["cross_account_role"] = types.StringValue(aws.ToString(config.Value.CrossAccountRole))
	} else {
		auroraAttrs["cross_account_role"] = types.StringNull()
	}

	if config.Value.ExternalId != nil {
		auroraAttrs["external_id"] = types.StringValue(aws.ToString(config.Value.ExternalId))
	} else {
		auroraAttrs["external_id"] = types.StringNull()
	}

	if len(config.Value.DatabaseClusterArns) > 0 {
		clusterArns, clusterArnsDiags := types.ListValueFrom(ctx, types.StringType, config.Value.DatabaseClusterArns)
		diags.Append(clusterArnsDiags...)
		auroraAttrs["database_cluster_arns"] = clusterArns
	} else {
		auroraAttrs["database_cluster_arns"] = types.ListNull(types.StringType)
	}

	if config.Value.Ungraceful != nil {
		ungracefulAttrs := map[string]attr.Value{
			"ungraceful": types.StringValue(string(config.Value.Ungraceful.Ungraceful)),
		}
		ungracefulObj, ungracefulDiags := types.ObjectValue(getGlobalAuroraUngracefulObjectType().AttrTypes, ungracefulAttrs)
		diags.Append(ungracefulDiags...)
		auroraAttrs["ungraceful"] = types.ListValueMust(getGlobalAuroraUngracefulObjectType(), []attr.Value{ungracefulObj})
	} else {
		auroraAttrs["ungraceful"] = types.ListNull(getGlobalAuroraUngracefulObjectType())
	}

	auroraObj, auroraDiags := types.ObjectValue(getGlobalAuroraConfigObjectType().AttrTypes, auroraAttrs)
	diags.Append(auroraDiags...)
	return types.ListValueMust(getGlobalAuroraConfigObjectType(), []attr.Value{auroraObj})
}

func flattenEc2AsgCapacityIncreaseConfig(config *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig, diags *fwdiag.Diagnostics) types.List {
	ec2Attrs := map[string]attr.Value{
		"capacity_monitoring_approach": types.StringValue(string(config.Value.CapacityMonitoringApproach)),
		"target_percent":               types.Int64Value(int64(aws.ToInt32(config.Value.TargetPercent))),
	}

	if config.Value.TimeoutMinutes != nil {
		ec2Attrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
	} else {
		ec2Attrs["timeout_minutes"] = types.Int64Null()
	}

	if len(config.Value.Asgs) > 0 {
		asgElements := make([]attr.Value, len(config.Value.Asgs))
		for a, asg := range config.Value.Asgs {
			asgAttrs := map[string]attr.Value{
				"arn": types.StringValue(aws.ToString(asg.Arn)),
			}
			if asg.CrossAccountRole != nil {
				asgAttrs["cross_account_role"] = types.StringValue(aws.ToString(asg.CrossAccountRole))
			} else {
				asgAttrs["cross_account_role"] = types.StringNull()
			}
			if asg.ExternalId != nil {
				asgAttrs["external_id"] = types.StringValue(aws.ToString(asg.ExternalId))
			} else {
				asgAttrs["external_id"] = types.StringNull()
			}

			asgObj, asgDiags := types.ObjectValue(getAsgObjectType().AttrTypes, asgAttrs)
			diags.Append(asgDiags...)
			asgElements[a] = asgObj
		}
		ec2Attrs["asgs"] = types.ListValueMust(getAsgObjectType(), asgElements)
	} else {
		ec2Attrs["asgs"] = types.ListNull(getAsgObjectType())
	}

	if config.Value.Ungraceful != nil {
		ungracefulAttrs := map[string]attr.Value{
			"minimum_success_percentage": types.Int64Value(int64(aws.ToInt32(config.Value.Ungraceful.MinimumSuccessPercentage))),
		}
		ungracefulObj, ungracefulDiags := types.ObjectValue(getEc2UngracefulObjectType().AttrTypes, ungracefulAttrs)
		diags.Append(ungracefulDiags...)
		ec2Attrs["ungraceful"] = types.ListValueMust(getEc2UngracefulObjectType(), []attr.Value{ungracefulObj})
	} else {
		ec2Attrs["ungraceful"] = types.ListNull(getEc2UngracefulObjectType())
	}

	ec2Obj, ec2Diags := types.ObjectValue(getEc2AsgCapacityIncreaseConfigObjectType().AttrTypes, ec2Attrs)
	diags.Append(ec2Diags...)
	return types.ListValueMust(getEc2AsgCapacityIncreaseConfigObjectType(), []attr.Value{ec2Obj})
}

func flattenEcsCapacityIncreaseConfig(config *awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig, diags *fwdiag.Diagnostics) types.List {
	ecsAttrs := map[string]attr.Value{
		"capacity_monitoring_approach": types.StringValue(string(config.Value.CapacityMonitoringApproach)),
		"target_percent":               types.Int64Value(int64(aws.ToInt32(config.Value.TargetPercent))),
	}

	if config.Value.TimeoutMinutes != nil {
		ecsAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
	} else {
		ecsAttrs["timeout_minutes"] = types.Int64Null()
	}

	if len(config.Value.Services) > 0 {
		serviceElements := make([]attr.Value, len(config.Value.Services))
		for s, service := range config.Value.Services {
			serviceAttrs := map[string]attr.Value{
				"cluster_arn": types.StringValue(aws.ToString(service.ClusterArn)),
				"service_arn": types.StringValue(aws.ToString(service.ServiceArn)),
			}
			if service.CrossAccountRole != nil {
				serviceAttrs["cross_account_role"] = types.StringValue(aws.ToString(service.CrossAccountRole))
			} else {
				serviceAttrs["cross_account_role"] = types.StringNull()
			}
			if service.ExternalId != nil {
				serviceAttrs["external_id"] = types.StringValue(aws.ToString(service.ExternalId))
			} else {
				serviceAttrs["external_id"] = types.StringNull()
			}

			serviceObj, serviceDiags := types.ObjectValue(getServiceObjectType().AttrTypes, serviceAttrs)
			diags.Append(serviceDiags...)
			serviceElements[s] = serviceObj
		}
		ecsAttrs["services"] = types.ListValueMust(getServiceObjectType(), serviceElements)
	} else {
		ecsAttrs["services"] = types.ListNull(getServiceObjectType())
	}

	if config.Value.Ungraceful != nil {
		ungracefulAttrs := map[string]attr.Value{
			"minimum_success_percentage": types.Int64Value(int64(aws.ToInt32(config.Value.Ungraceful.MinimumSuccessPercentage))),
		}
		ungracefulObj, ungracefulDiags := types.ObjectValue(getEcsUngracefulObjectType().AttrTypes, ungracefulAttrs)
		diags.Append(ungracefulDiags...)
		ecsAttrs["ungraceful"] = types.ListValueMust(getEcsUngracefulObjectType(), []attr.Value{ungracefulObj})
	} else {
		ecsAttrs["ungraceful"] = types.ListNull(getEcsUngracefulObjectType())
	}

	ecsObj, ecsDiags := types.ObjectValue(getEcsCapacityIncreaseConfigObjectType().AttrTypes, ecsAttrs)
	diags.Append(ecsDiags...)
	return types.ListValueMust(getEcsCapacityIncreaseConfigObjectType(), []attr.Value{ecsObj})
}

func flattenEksResourceScalingConfig(config *awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig, diags *fwdiag.Diagnostics) types.List {
	eksAttrs := map[string]attr.Value{
		"capacity_monitoring_approach": types.StringValue(string(config.Value.CapacityMonitoringApproach)),
		"target_percent":               types.Int64Value(int64(aws.ToInt32(config.Value.TargetPercent))),
	}

	if config.Value.TimeoutMinutes != nil {
		eksAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
	} else {
		eksAttrs["timeout_minutes"] = types.Int64Null()
	}

	if len(config.Value.EksClusters) > 0 {
		clusterElements := make([]attr.Value, len(config.Value.EksClusters))
		for c, cluster := range config.Value.EksClusters {
			clusterAttrs := map[string]attr.Value{
				"cluster_arn": types.StringValue(aws.ToString(cluster.ClusterArn)),
			}
			if cluster.CrossAccountRole != nil {
				clusterAttrs["cross_account_role"] = types.StringValue(aws.ToString(cluster.CrossAccountRole))
			} else {
				clusterAttrs["cross_account_role"] = types.StringNull()
			}
			if cluster.ExternalId != nil {
				clusterAttrs["external_id"] = types.StringValue(aws.ToString(cluster.ExternalId))
			} else {
				clusterAttrs["external_id"] = types.StringNull()
			}

			clusterObj, clusterDiags := types.ObjectValue(getEksClusterObjectType().AttrTypes, clusterAttrs)
			diags.Append(clusterDiags...)
			clusterElements[c] = clusterObj
		}
		eksAttrs["eks_clusters"] = types.ListValueMust(getEksClusterObjectType(), clusterElements)
	} else {
		eksAttrs["eks_clusters"] = types.ListNull(getEksClusterObjectType())
	}

	if config.Value.KubernetesResourceType != nil {
		k8sAttrs := map[string]attr.Value{
			"api_version": types.StringValue(aws.ToString(config.Value.KubernetesResourceType.ApiVersion)),
			"kind":        types.StringValue(aws.ToString(config.Value.KubernetesResourceType.Kind)),
		}
		k8sObj, k8sDiags := types.ObjectValue(getKubernetesResourceTypeObjectType().AttrTypes, k8sAttrs)
		diags.Append(k8sDiags...)
		eksAttrs["kubernetes_resource_type"] = types.ListValueMust(getKubernetesResourceTypeObjectType(), []attr.Value{k8sObj})
	} else {
		eksAttrs["kubernetes_resource_type"] = types.ListNull(getKubernetesResourceTypeObjectType())
	}

	if len(config.Value.ScalingResources) > 0 {
		scalingElements := make([]attr.Value, 0)
		for _, scalingResourceMap := range config.Value.ScalingResources {
			for namespace, resourceMap := range scalingResourceMap {
				scalingAttrs := map[string]attr.Value{
					"namespace": types.StringValue(namespace),
				}

				resourceElements := make([]attr.Value, 0)
				resourceNames := make([]string, 0, len(resourceMap))
				for resourceName := range resourceMap {
					resourceNames = append(resourceNames, resourceName)
				}
				sort.Strings(resourceNames)

				for _, resourceName := range resourceNames {
					resource := resourceMap[resourceName]
					resourceAttrs := map[string]attr.Value{
						"resource_name": types.StringValue(resourceName),
						"name":          types.StringValue(aws.ToString(resource.Name)),
						"namespace":     types.StringValue(aws.ToString(resource.Namespace)),
						"hpa_name":      types.StringValue(aws.ToString(resource.HpaName)),
					}
					resourceObj, resourceDiags := types.ObjectValue(getKubernetesScalingResourceObjectType().AttrTypes, resourceAttrs)
					diags.Append(resourceDiags...)
					resourceElements = append(resourceElements, resourceObj)
				}
				scalingAttrs["resources"] = types.ListValueMust(getKubernetesScalingResourceObjectType(), resourceElements)

				scalingObj, scalingDiags := types.ObjectValue(getScalingResourcesObjectType().AttrTypes, scalingAttrs)
				diags.Append(scalingDiags...)
				scalingElements = append(scalingElements, scalingObj)
			}
		}
		eksAttrs["scaling_resources"] = types.ListValueMust(getScalingResourcesObjectType(), scalingElements)
	} else {
		eksAttrs["scaling_resources"] = types.ListNull(getScalingResourcesObjectType())
	}

	if config.Value.Ungraceful != nil {
		ungracefulAttrs := map[string]attr.Value{
			"minimum_success_percentage": types.Int64Value(int64(aws.ToInt32(config.Value.Ungraceful.MinimumSuccessPercentage))),
		}
		ungracefulObj, ungracefulDiags := types.ObjectValue(getEksUngracefulObjectType().AttrTypes, ungracefulAttrs)
		diags.Append(ungracefulDiags...)
		eksAttrs["ungraceful"] = types.ListValueMust(getEksUngracefulObjectType(), []attr.Value{ungracefulObj})
	} else {
		eksAttrs["ungraceful"] = types.ListNull(getEksUngracefulObjectType())
	}

	eksObj, eksDiags := types.ObjectValue(getEksResourceScalingConfigObjectType().AttrTypes, eksAttrs)
	diags.Append(eksDiags...)
	return types.ListValueMust(getEksResourceScalingConfigObjectType(), []attr.Value{eksObj})
}

func flattenArcRoutingControlConfig(ctx context.Context, config *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig, diags *fwdiag.Diagnostics) types.List {
	arcAttrs := map[string]attr.Value{}

	if config.Value.CrossAccountRole != nil {
		arcAttrs["cross_account_role"] = types.StringValue(aws.ToString(config.Value.CrossAccountRole))
	} else {
		arcAttrs["cross_account_role"] = types.StringNull()
	}

	if config.Value.ExternalId != nil {
		arcAttrs["external_id"] = types.StringValue(aws.ToString(config.Value.ExternalId))
	} else {
		arcAttrs["external_id"] = types.StringNull()
	}

	if config.Value.TimeoutMinutes != nil {
		arcAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
	} else {
		arcAttrs["timeout_minutes"] = types.Int64Null()
	}

	if len(config.Value.RegionAndRoutingControls) > 0 {
		regionElements := make([]attr.Value, 0)
		regions := make([]string, 0, len(config.Value.RegionAndRoutingControls))
		for region := range config.Value.RegionAndRoutingControls {
			regions = append(regions, region)
		}
		sort.Strings(regions)

		for _, region := range regions {
			routingControlStates := config.Value.RegionAndRoutingControls[region]
			regionAttrs := map[string]attr.Value{
				"region": types.StringValue(region),
			}

			routingControlArns := make([]string, len(routingControlStates))
			for i, state := range routingControlStates {
				routingControlArns[i] = aws.ToString(state.RoutingControlArn)
			}

			if len(routingControlArns) > 0 {
				controlArns, controlArnsDiags := types.ListValueFrom(ctx, types.StringType, routingControlArns)
				diags.Append(controlArnsDiags...)
				regionAttrs["routing_control_arns"] = controlArns
			} else {
				regionAttrs["routing_control_arns"] = types.ListNull(types.StringType)
			}

			regionObj, regionDiags := types.ObjectValue(getRegionAndRoutingControlsObjectType().AttrTypes, regionAttrs)
			diags.Append(regionDiags...)
			regionElements = append(regionElements, regionObj)
		}
		arcAttrs["region_and_routing_controls"] = types.ListValueMust(getRegionAndRoutingControlsObjectType(), regionElements)
	} else {
		arcAttrs["region_and_routing_controls"] = types.ListNull(getRegionAndRoutingControlsObjectType())
	}

	arcObj, arcDiags := types.ObjectValue(getArcRoutingControlConfigObjectType().AttrTypes, arcAttrs)
	diags.Append(arcDiags...)
	return types.ListValueMust(getArcRoutingControlConfigObjectType(), []attr.Value{arcObj})
}

func flattenParallelConfig(config *awstypes.ExecutionBlockConfigurationMemberParallelConfig, diags *fwdiag.Diagnostics) types.List {
	parallelStepElements := make([]attr.Value, len(config.Value.Steps))
	for k, parallelStep := range config.Value.Steps {
		parallelStepAttrs := map[string]attr.Value{
			"name":                 types.StringValue(aws.ToString(parallelStep.Name)),
			"execution_block_type": types.StringValue(string(parallelStep.ExecutionBlockType)),
		}

		if parallelStep.Description != nil {
			parallelStepAttrs["description"] = types.StringValue(aws.ToString(parallelStep.Description))
		} else {
			parallelStepAttrs["description"] = types.StringNull()
		}

		if parallelStep.ExecutionBlockConfiguration != nil {
			switch parallelConfig := parallelStep.ExecutionBlockConfiguration.(type) {
			case *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
				parallelStepAttrs["custom_action_lambda_config"] = flattenCustomActionLambdaConfig(parallelConfig, diags)
				parallelStepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
			default:
				parallelStepAttrs["custom_action_lambda_config"] = types.ListNull(getCustomActionLambdaConfigObjectType())
				parallelStepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
			}
		} else {
			parallelStepAttrs["custom_action_lambda_config"] = types.ListNull(getCustomActionLambdaConfigObjectType())
			parallelStepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
		}

		parallelStepObj, parallelStepDiags := types.ObjectValue(getParallelStepObjectType().AttrTypes, parallelStepAttrs)
		diags.Append(parallelStepDiags...)
		parallelStepElements[k] = parallelStepObj
	}

	parallelConfigAttrs := map[string]attr.Value{
		"step": types.ListValueMust(getParallelStepObjectType(), parallelStepElements),
	}

	parallelConfigObj, parallelConfigDiags := types.ObjectValue(getParallelConfigObjectType().AttrTypes, parallelConfigAttrs)
	diags.Append(parallelConfigDiags...)
	return types.ListValueMust(getParallelConfigObjectType(), []attr.Value{parallelConfigObj})
}

func flattenWorkflowsToFramework(
	ctx context.Context,
	workflows []awstypes.Workflow,
) (types.List, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	if len(workflows) == 0 {
		return types.ListNull(getWorkflowObjectType()), diags
	}

	// Sort workflows by target action to ensure consistent ordering (activate first, then deactivate)
	sortedWorkflows := make([]awstypes.Workflow, len(workflows))
	copy(sortedWorkflows, workflows)

	// Simple sort: activate workflows first, then deactivate workflows
	var activateWorkflows, deactivateWorkflows []awstypes.Workflow
	for _, workflow := range sortedWorkflows {
		if workflow.WorkflowTargetAction == awstypes.WorkflowTargetActionActivate {
			activateWorkflows = append(activateWorkflows, workflow)
		} else {
			deactivateWorkflows = append(deactivateWorkflows, workflow)
		}
	}

	// Combine: activate workflows first, then deactivate workflows
	sortedWorkflows = append(activateWorkflows, deactivateWorkflows...)

	elements := make([]attr.Value, len(sortedWorkflows))
	for i, workflow := range sortedWorkflows {
		workflowAttrs := map[string]attr.Value{
			"workflow_target_action": types.StringValue(string(workflow.WorkflowTargetAction)),
		}

		if workflow.WorkflowTargetRegion != nil {
			workflowAttrs["workflow_target_region"] = types.StringValue(aws.ToString(workflow.WorkflowTargetRegion))
		} else {
			workflowAttrs["workflow_target_region"] = types.StringNull()
		}

		if workflow.WorkflowDescription != nil {
			workflowAttrs["workflow_description"] = types.StringValue(aws.ToString(workflow.WorkflowDescription))
		} else {
			workflowAttrs["workflow_description"] = types.StringNull()
		}

		// Handle steps
		if len(workflow.Steps) > 0 {
			stepElements := make([]attr.Value, len(workflow.Steps))
			for j, step := range workflow.Steps {
				stepAttrs := map[string]attr.Value{
					"name":                 types.StringValue(aws.ToString(step.Name)),
					"execution_block_type": types.StringValue(string(step.ExecutionBlockType)),
				}

				if step.Description != nil {
					stepAttrs["description"] = types.StringValue(aws.ToString(step.Description))
				} else {
					stepAttrs["description"] = types.StringNull()
				}

				// Handle execution block configuration
				if step.ExecutionBlockConfiguration != nil {
					// Initialize all execution block configs to null first
					stepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
					stepAttrs["route53_health_check_config"] = types.ListNull(getRoute53HealthCheckConfigObjectType())
					stepAttrs["custom_action_lambda_config"] = types.ListNull(getCustomActionLambdaConfigObjectType())
					stepAttrs["global_aurora_config"] = types.ListNull(getGlobalAuroraConfigObjectType())
					stepAttrs["ec2_asg_capacity_increase_config"] = types.ListNull(getEc2AsgCapacityIncreaseConfigObjectType())
					stepAttrs["ecs_capacity_increase_config"] = types.ListNull(getEcsCapacityIncreaseConfigObjectType())
					stepAttrs["eks_resource_scaling_config"] = types.ListNull(getEksResourceScalingConfigObjectType())
					stepAttrs["arc_routing_control_config"] = types.ListNull(getArcRoutingControlConfigObjectType())
					stepAttrs["parallel_config"] = types.ListNull(getParallelConfigObjectType())

					// Now each case only needs to set its specific config
					switch config := step.ExecutionBlockConfiguration.(type) {
					case *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig:
						stepAttrs["execution_approval_config"] = flattenExecutionApprovalConfig(config, &diags)

					case *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig:
						stepAttrs["route53_health_check_config"] = flattenRoute53HealthCheckConfig(config, &diags)

					case *awstypes.ExecutionBlockConfigurationMemberParallelConfig:
						stepAttrs["parallel_config"] = flattenParallelConfig(config, &diags)

					case *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
						stepAttrs["custom_action_lambda_config"] = flattenCustomActionLambdaConfig(config, &diags)

					case *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig:
						stepAttrs["global_aurora_config"] = flattenGlobalAuroraConfig(ctx, config, &diags)

					case *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig:
						stepAttrs["ec2_asg_capacity_increase_config"] = flattenEc2AsgCapacityIncreaseConfig(config, &diags)

					case *awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig:
						stepAttrs["ecs_capacity_increase_config"] = flattenEcsCapacityIncreaseConfig(config, &diags)

					case *awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig:
						stepAttrs["eks_resource_scaling_config"] = flattenEksResourceScalingConfig(config, &diags)

					case *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig:
						stepAttrs["arc_routing_control_config"] = flattenArcRoutingControlConfig(ctx, config, &diags)
					}
				} else {
					// No execution block configuration - all configs are null (already set above)
					stepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
					stepAttrs["route53_health_check_config"] = types.ListNull(getRoute53HealthCheckConfigObjectType())
					stepAttrs["custom_action_lambda_config"] = types.ListNull(getCustomActionLambdaConfigObjectType())
					stepAttrs["global_aurora_config"] = types.ListNull(getGlobalAuroraConfigObjectType())
					stepAttrs["ec2_asg_capacity_increase_config"] = types.ListNull(getEc2AsgCapacityIncreaseConfigObjectType())
					stepAttrs["ecs_capacity_increase_config"] = types.ListNull(getEcsCapacityIncreaseConfigObjectType())
					stepAttrs["eks_resource_scaling_config"] = types.ListNull(getEksResourceScalingConfigObjectType())
					stepAttrs["arc_routing_control_config"] = types.ListNull(getArcRoutingControlConfigObjectType())
					stepAttrs["parallel_config"] = types.ListNull(getParallelConfigObjectType())
				}

				stepObj, stepDiags := types.ObjectValue(getStepObjectType().AttrTypes, stepAttrs)
				diags.Append(stepDiags...)
				stepElements[j] = stepObj
			}

			stepsList, stepsDiags := types.ListValue(getStepObjectType(), stepElements)
			diags.Append(stepsDiags...)
			workflowAttrs["step"] = stepsList
		} else {
			workflowAttrs["step"] = types.ListNull(getStepObjectType())
		}

		workflowObj, objDiags := types.ObjectValue(getWorkflowObjectType().AttrTypes, workflowAttrs)
		diags.Append(objDiags...)
		elements[i] = workflowObj
	}

	result, resultDiags := types.ListValue(getWorkflowObjectType(), elements)
	diags.Append(resultDiags...)
	return result, diags
}
