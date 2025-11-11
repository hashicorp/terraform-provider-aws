// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
							Type:     schema.TypeInt,
							Computed: true,
						},
						"canary_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"canary_bake_time_in_minutes": {
										Type:     schema.TypeInt,
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
										Type:     schema.TypeInt,
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
			names.AttrPropagateTags: {
				Type:     schema.TypeString,
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
			names.AttrServiceName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"task_definition": {
				Type:     schema.TypeString,
				Computed: true,
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
	if service.DeploymentConfiguration != nil {
		if err := d.Set("deployment_configuration", flattenDeploymentConfiguration(service.DeploymentConfiguration)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting deployment_configuration: %s", err)
		}
	}
	if err := d.Set("deployment_controller", flattenDeploymentController(service.DeploymentController)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting deployment_controller: %s", err)
	}
	d.Set("desired_count", service.DesiredCount)
	d.Set("enable_ecs_managed_tags", service.EnableECSManagedTags)
	d.Set("enable_execute_command", service.EnableExecuteCommand)
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
	d.Set(names.AttrPropagateTags, service.PropagateTags)
	d.Set("scheduling_strategy", service.SchedulingStrategy)
	if err := d.Set("service_registries", flattenServiceRegistries(service.ServiceRegistries)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_registries: %s", err)
	}
	d.Set(names.AttrServiceName, service.ServiceName)
	d.Set("task_definition", service.TaskDefinition)

	setTagsOut(ctx, service.Tags)

	return diags
}
