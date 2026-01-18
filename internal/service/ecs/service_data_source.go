// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecs_service", name="Service")
// @Tags
func dataSourceService() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceServiceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_rebalancing": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCapacityProviderStrategy: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"capacity_provider": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrWeight: {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alarms": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"alarm_names": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"enable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"rollback": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"bake_time_in_minutes": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"canary_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"canary_bake_time_in_minutes": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"canary_percent": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
						},
						"deployment_circuit_breaker": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"rollback": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"linear_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"step_bake_time_in_minutes": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"step_percent": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
						},
						"lifecycle_hook": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hook_details": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"hook_target_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"lifecycle_stages": {
										Type:     schema.TypeList,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrRoleARN: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"maximum_percent": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"minimum_healthy_percent": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"strategy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"deployment_controller": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"deployments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCreatedAt: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"desired_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"pending_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"running_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"task_definition": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"desired_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"enable_ecs_managed_tags": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_execute_command": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"events": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCreatedAt: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrMessage: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"health_check_grace_period_seconds": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"iam_role": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrNetworkConfiguration: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnets: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"load_balancer": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"container_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"elb_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_group_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"advanced_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"alternate_target_group_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"production_listener_rule": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"test_listener_rule": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrRoleARN: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"ordered_placement_strategy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrField: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"placement_constraints": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrExpression: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"platform_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pending_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"platform_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPropagateTags: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"running_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"scheduling_strategy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_registries": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"container_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrPort: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"registry_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"task_definition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"task_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrCreatedAt: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"pending_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"running_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"stability_status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"task_definition": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceServiceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECSClient(ctx)

	service, err := findServiceByTwoPartKey(ctx, conn, d.Get(names.AttrServiceName).(string), d.Get("cluster_arn").(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("ECS Service", err))
	}

	arn := aws.ToString(service.ServiceArn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set("availability_zone_rebalancing", service.AvailabilityZoneRebalancing)
	if err := d.Set(names.AttrCapacityProviderStrategy, flattenCapacityProviderStrategyItems(service.CapacityProviderStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting capacity_provider_strategy: %s", err)
	}
	d.Set("cluster_arn", service.ClusterArn)
	if service.CreatedAt != nil {
		d.Set(names.AttrCreatedAt, service.CreatedAt.Format(time.RFC3339))
	}
	d.Set("created_by", service.CreatedBy)
	if service.DeploymentConfiguration != nil {
		if err := d.Set("deployment_configuration", flattenDeploymentConfigurationForDataSource(service.DeploymentConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting deployment_configuration: %s", err)
		}
	}
	if err := d.Set("deployment_controller", flattenDeploymentController(service.DeploymentController)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_controller: %s", err)
	}
	if err := d.Set("deployments", flattenDeployments(service.Deployments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployments: %s", err)
	}
	d.Set("desired_count", service.DesiredCount)
	d.Set("enable_ecs_managed_tags", service.EnableECSManagedTags)
	d.Set("enable_execute_command", service.EnableExecuteCommand)
	if err := d.Set("events", flattenServiceEvents(service.Events)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting events: %s", err)
	}
	d.Set("health_check_grace_period_seconds", service.HealthCheckGracePeriodSeconds)
	d.Set("iam_role", service.RoleArn)
	d.Set("launch_type", service.LaunchType)
	if service.LoadBalancers != nil {
		if err := d.Set("load_balancer", flattenServiceLoadBalancers(service.LoadBalancers)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting load_balancer: %s", err)
		}
	}
	if err := d.Set(names.AttrNetworkConfiguration, flattenNetworkConfiguration(service.NetworkConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_configuration: %s", err)
	}
	if err := d.Set("ordered_placement_strategy", flattenPlacementStrategy(service.PlacementStrategy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ordered_placement_strategy: %s", err)
	}
	if err := d.Set("placement_constraints", flattenServicePlacementConstraints(service.PlacementConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting placement_constraints: %s", err)
	}
	d.Set("platform_version", service.PlatformVersion)
	d.Set("pending_count", service.PendingCount)
	d.Set("platform_family", service.PlatformFamily)
	d.Set(names.AttrPropagateTags, service.PropagateTags)
	d.Set("running_count", service.RunningCount)
	d.Set("scheduling_strategy", service.SchedulingStrategy)
	if err := d.Set("service_registries", flattenServiceRegistries(service.ServiceRegistries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_registries: %s", err)
	}
	d.Set(names.AttrStatus, service.Status)
	d.Set(names.AttrServiceName, service.ServiceName)
	d.Set("task_definition", service.TaskDefinition)
	if err := d.Set("task_sets", flattenTaskSets(service.TaskSets)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting task_sets: %s", err)
	}

	setTagsOut(ctx, service.Tags)

	return diags
}

// flattenDeploymentConfigurationForDataSource flattens DeploymentConfiguration for the data source
// which includes all fields from the API, unlike the resource which has some fields at the top level
func flattenDeploymentConfigurationForDataSource(apiObject *awstypes.DeploymentConfiguration) []map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Alarms; v != nil {
		tfMap["alarms"] = []map[string]any{flattenAlarmsForDataSource(v)}
	}

	if v := apiObject.BakeTimeInMinutes; v != nil {
		tfMap["bake_time_in_minutes"] = flex.Int32ToStringValue(v)
	}

	if v := apiObject.CanaryConfiguration; v != nil {
		tfMap["canary_configuration"] = flattenCanaryConfiguration(v)
	}

	if v := apiObject.DeploymentCircuitBreaker; v != nil {
		tfMap["deployment_circuit_breaker"] = []map[string]any{flattenDeploymentCircuitBreakerForDataSource(v)}
	}

	if v := apiObject.LinearConfiguration; v != nil {
		tfMap["linear_configuration"] = flattenLinearConfiguration(v)
	}

	if v := apiObject.LifecycleHooks; len(v) > 0 {
		tfMap["lifecycle_hook"] = flattenLifecycleHooks(v)
	}

	if v := apiObject.MaximumPercent; v != nil {
		tfMap["maximum_percent"] = aws.ToInt32(v)
	}

	if v := apiObject.MinimumHealthyPercent; v != nil {
		tfMap["minimum_healthy_percent"] = aws.ToInt32(v)
	}

	if v := apiObject.Strategy; v != "" {
		tfMap["strategy"] = string(v)
	}

	return []map[string]any{tfMap}
}

func flattenAlarmsForDataSource(apiObject *awstypes.DeploymentAlarms) map[string]any {
	tfMap := map[string]any{
		"enable":   apiObject.Enable,
		"rollback": apiObject.Rollback,
	}

	if v := apiObject.AlarmNames; len(v) > 0 {
		tfMap["alarm_names"] = v
	}

	return tfMap
}

func flattenDeploymentCircuitBreakerForDataSource(apiObject *awstypes.DeploymentCircuitBreaker) map[string]any {
	return map[string]any{
		"enable":   apiObject.Enable,
		"rollback": apiObject.Rollback,
	}
}

func flattenDeployments(apiObjects []awstypes.Deployment) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]any
	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"desired_count":   apiObject.DesiredCount,
			names.AttrID:      aws.ToString(apiObject.Id),
			"pending_count":   apiObject.PendingCount,
			"running_count":   apiObject.RunningCount,
			names.AttrStatus:  aws.ToString(apiObject.Status),
			"task_definition": aws.ToString(apiObject.TaskDefinition),
		}
		if apiObject.CreatedAt != nil {
			tfMap[names.AttrCreatedAt] = apiObject.CreatedAt.Format(time.RFC3339)
		}
		if apiObject.UpdatedAt != nil {
			tfMap["updated_at"] = apiObject.UpdatedAt.Format(time.RFC3339)
		}
		tfList = append(tfList, tfMap)
	}
	return tfList
}

func flattenServiceEvents(apiObjects []awstypes.ServiceEvent) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]any
	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrID:      aws.ToString(apiObject.Id),
			names.AttrMessage: aws.ToString(apiObject.Message),
		}
		if apiObject.CreatedAt != nil {
			tfMap[names.AttrCreatedAt] = apiObject.CreatedAt.Format(time.RFC3339)
		}
		tfList = append(tfList, tfMap)
	}
	return tfList
}

func flattenTaskSets(apiObjects []awstypes.TaskSet) []map[string]any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]any
	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrARN:      aws.ToString(apiObject.TaskSetArn),
			names.AttrID:       aws.ToString(apiObject.Id),
			"pending_count":    apiObject.PendingCount,
			"running_count":    apiObject.RunningCount,
			"stability_status": string(apiObject.StabilityStatus),
			names.AttrStatus:   aws.ToString(apiObject.Status),
			"task_definition":  aws.ToString(apiObject.TaskDefinition),
		}
		if apiObject.CreatedAt != nil {
			tfMap[names.AttrCreatedAt] = apiObject.CreatedAt.Format(time.RFC3339)
		}
		if apiObject.UpdatedAt != nil {
			tfMap["updated_at"] = apiObject.UpdatedAt.Format(time.RFC3339)
		}
		tfList = append(tfList, tfMap)
	}
	return tfList
}
