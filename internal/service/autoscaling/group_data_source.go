// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_autoscaling_group", name="Group")
func dataSourceGroup() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGroupRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_cooldown": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"desired_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"desired_capacity_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled_metrics": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"health_check_grace_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"health_check_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_maintenance_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_healthy_percentage": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"min_healthy_percentage": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"launch_configuration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLaunchTemplate: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVersion: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"load_balancers": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"max_instance_lifetime": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"min_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"mixed_instances_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instances_distribution": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"on_demand_allocation_strategy": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"on_demand_base_capacity": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"on_demand_percentage_above_base_capacity": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"spot_allocation_strategy": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"spot_instance_pools": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"spot_max_price": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						names.AttrLaunchTemplate: {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_specification": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"launch_template_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"launch_template_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrVersion: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									"override": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_requirements": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"accelerator_count": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																	},
																},
															},
															"accelerator_manufacturers": {
																Type:     schema.TypeSet,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"accelerator_names": {
																Type:     schema.TypeSet,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"accelerator_total_memory_mib": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																	},
																},
															},
															"accelerator_types": {
																Type:     schema.TypeSet,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"allowed_instance_types": {
																Type:     schema.TypeSet,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"bare_metal": {
																Type:     schema.TypeString,
																Computed: true,
															},
															"baseline_ebs_bandwidth_mbps": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																	},
																},
															},
															"burstable_performance": {
																Type:     schema.TypeString,
																Computed: true,
															},
															"cpu_manufacturers": {
																Type:     schema.TypeSet,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"excluded_instance_types": {
																Type:     schema.TypeSet,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"instance_generations": {
																Type:     schema.TypeSet,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"local_storage": {
																Type:     schema.TypeString,
																Computed: true,
															},
															"local_storage_types": {
																Type:     schema.TypeSet,
																Computed: true,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"max_spot_price_as_percentage_of_optimal_on_demand_price": {
																Type:     schema.TypeInt,
																Computed: true,
															},
															"memory_gib_per_vcpu": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeFloat,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeFloat,
																			Computed: true,
																		},
																	},
																},
															},
															"memory_mib": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																	},
																},
															},
															"network_bandwidth_gbps": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeFloat,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeFloat,
																			Computed: true,
																		},
																	},
																},
															},
															"network_interface_count": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																	},
																},
															},
															"on_demand_max_price_percentage_over_lowest_price": {
																Type:     schema.TypeInt,
																Computed: true,
															},
															"require_hibernate_support": {
																Type:     schema.TypeBool,
																Computed: true,
															},
															"spot_max_price_percentage_over_lowest_price": {
																Type:     schema.TypeInt,
																Computed: true,
															},
															"total_local_storage_gb": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeFloat,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeFloat,
																			Computed: true,
																		},
																	},
																},
															},
															"vcpu_count": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Computed: true,
																		},
																	},
																},
															},
														},
													},
												},
												names.AttrInstanceType: {
													Type:     schema.TypeString,
													Computed: true,
												},
												"launch_template_specification": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"launch_template_id": {
																Type:     schema.TypeString,
																Computed: true,
															},
															"launch_template_name": {
																Type:     schema.TypeString,
																Computed: true,
															},
															names.AttrVersion: {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},
												"weighted_capacity": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"new_instances_protected_from_scale_in": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"placement_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"predicted_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"service_linked_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"suspended_processes": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tag": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"propagate_at_launch": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"target_group_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"termination_policies": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"traffic_source": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIdentifier: {
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
			"vpc_zone_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"warm_pool": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_reuse_policy": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"reuse_on_scale_in": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"max_group_prepared_capacity": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"min_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"pool_state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"warm_pool_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	groupName := d.Get(names.AttrName).(string)
	group, err := findGroupByName(ctx, conn, groupName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Group (%s): %s", groupName, err)
	}

	d.SetId(aws.ToString(group.AutoScalingGroupName))
	d.Set(names.AttrARN, group.AutoScalingGroupARN)
	d.Set(names.AttrAvailabilityZones, group.AvailabilityZones)
	d.Set("default_cooldown", group.DefaultCooldown)
	d.Set("desired_capacity", group.DesiredCapacity)
	d.Set("desired_capacity_type", group.DesiredCapacityType)
	d.Set("enabled_metrics", flattenEnabledMetrics(group.EnabledMetrics))
	d.Set("health_check_grace_period", group.HealthCheckGracePeriod)
	d.Set("health_check_type", group.HealthCheckType)
	if err := d.Set("instance_maintenance_policy", flattenInstanceMaintenancePolicy(group.InstanceMaintenancePolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_maintenance_policy: %s", err)
	}
	d.Set("launch_configuration", group.LaunchConfigurationName)
	if group.LaunchTemplate != nil {
		if err := d.Set(names.AttrLaunchTemplate, []interface{}{flattenLaunchTemplateSpecification(group.LaunchTemplate)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting launch_template: %s", err)
		}
	} else {
		d.Set(names.AttrLaunchTemplate, nil)
	}
	d.Set("load_balancers", group.LoadBalancerNames)
	d.Set("max_instance_lifetime", group.MaxInstanceLifetime)
	d.Set("max_size", group.MaxSize)
	d.Set("min_size", group.MinSize)
	if group.MixedInstancesPolicy != nil {
		if err := d.Set("mixed_instances_policy", []interface{}{flattenMixedInstancesPolicy(group.MixedInstancesPolicy)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting mixed_instances_policy: %s", err)
		}
	} else {
		d.Set("mixed_instances_policy", nil)
	}
	d.Set(names.AttrName, group.AutoScalingGroupName)
	d.Set("new_instances_protected_from_scale_in", group.NewInstancesProtectedFromScaleIn)
	d.Set("placement_group", group.PlacementGroup)
	d.Set("predicted_capacity", group.PredictedCapacity)
	d.Set("service_linked_role_arn", group.ServiceLinkedRoleARN)
	d.Set(names.AttrStatus, group.Status)
	d.Set("suspended_processes", flattenSuspendedProcesses(group.SuspendedProcesses))
	if err := d.Set("tag", listOfMap(KeyValueTags(ctx, group.Tags, d.Id(), TagResourceTypeGroup).IgnoreAWS().IgnoreConfig(ignoreTagsConfig))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tag: %s", err)
	}
	d.Set("target_group_arns", group.TargetGroupARNs)
	d.Set("termination_policies", group.TerminationPolicies)
	if err := d.Set("traffic_source", flattenTrafficSourceIdentifiers(group.TrafficSources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting traffic_source: %s", err)
	}
	d.Set("vpc_zone_identifier", group.VPCZoneIdentifier)
	if group.WarmPoolConfiguration != nil {
		if err := d.Set("warm_pool", []interface{}{flattenWarmPoolConfiguration(group.WarmPoolConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting warm_pool: %s", err)
		}
	} else {
		d.Set("warm_pool", nil)
	}
	d.Set("warm_pool_size", group.WarmPoolSize)

	return diags
}
