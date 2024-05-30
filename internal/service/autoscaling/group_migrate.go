// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// aws_autoscaling_group resource's Schema @v5.11.0 minus validators.
func resourceGroupV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"capacity_rebalance": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"context": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_cooldown": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"default_instance_warmup": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"desired_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"desired_capacity_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled_metrics": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrForceDelete: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"force_delete_warm_pool": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"health_check_grace_period": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
			},
			"health_check_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"initial_lifecycle_hook": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_result": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"heartbeat_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"lifecycle_transition": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"notification_metadata": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"notification_target_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrRoleARN: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"instance_refresh": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"preferences": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_rollback": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"checkpoint_delay": {
										Type:     nullable.TypeNullableInt,
										Optional: true,
									},
									"checkpoint_percentages": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
									},
									"instance_warmup": {
										Type:     nullable.TypeNullableInt,
										Optional: true,
									},
									"min_healthy_percentage": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  90,
									},
									"skip_matching": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"strategy": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrTriggers: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"launch_configuration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrLaunchTemplate: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						names.AttrVersion: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"load_balancers": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"max_instance_lifetime": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"max_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"metrics_granularity": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  defaultEnabledMetricsGranularity,
			},
			"min_elb_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"min_size": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"mixed_instances_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instances_distribution": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Computed: true,
							// Ideally we'd want to detect drift detection,
							// but a DiffSuppressFunc here does not behave nicely
							// for detecting missing configuration blocks
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									// These fields are returned from calls to the API
									// even if not provided at input time and can be omitted in requests;
									// thus, to prevent non-empty plans, we set these
									// to Computed and remove Defaults
									"on_demand_allocation_strategy": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"on_demand_base_capacity": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"on_demand_percentage_above_base_capacity": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"spot_allocation_strategy": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"spot_instance_pools": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"spot_max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						names.AttrLaunchTemplate: {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_specification": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"launch_template_id": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"launch_template_name": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												names.AttrVersion: {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "$Default",
												},
											},
										},
									},
									"override": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"instance_requirements": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"accelerator_count": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
															"accelerator_manufacturers": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
															"accelerator_names": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
															"accelerator_total_memory_mib": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
															"accelerator_types": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
															"allowed_instance_types": {
																Type:     schema.TypeSet,
																Optional: true,
																MaxItems: 400,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"bare_metal": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"baseline_ebs_bandwidth_mbps": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
															"burstable_performance": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"cpu_manufacturers": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
															"excluded_instance_types": {
																Type:     schema.TypeSet,
																Optional: true,
																MaxItems: 400,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"instance_generations": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
															"local_storage": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"local_storage_types": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																},
															},
															"memory_gib_per_vcpu": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeFloat,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeFloat,
																			Optional: true,
																		},
																	},
																},
															},
															"memory_mib": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
															"network_bandwidth_gbps": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeFloat,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeFloat,
																			Optional: true,
																		},
																	},
																},
															},
															"network_interface_count": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
															"on_demand_max_price_percentage_over_lowest_price": {
																Type:     schema.TypeInt,
																Optional: true,
															},
															"require_hibernate_support": {
																Type:     schema.TypeBool,
																Optional: true,
															},
															"spot_max_price_percentage_over_lowest_price": {
																Type:     schema.TypeInt,
																Optional: true,
															},
															"total_local_storage_gb": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeFloat,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeFloat,
																			Optional: true,
																		},
																	},
																},
															},
															"vcpu_count": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																		names.AttrMin: {
																			Type:     schema.TypeInt,
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												names.AttrInstanceType: {
													Type:     schema.TypeString,
													Optional: true,
												},
												"launch_template_specification": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"launch_template_id": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															"launch_template_name": {
																Type:     schema.TypeString,
																Optional: true,
																Computed: true,
															},
															names.AttrVersion: {
																Type:     schema.TypeString,
																Optional: true,
																Default:  "$Default",
															},
														},
													},
												},
												"weighted_capacity": {
													Type:     schema.TypeString,
													Optional: true,
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
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrNamePrefix: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"placement_group": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"predicted_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"protect_from_scale_in": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"service_linked_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"suspended_processes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tag": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
						"propagate_at_launch": {
							Type:     schema.TypeBool,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"target_group_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"termination_policies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"traffic_source": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrIdentifier: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"vpc_zone_identifier": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"wait_for_capacity_timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "10m",
			},
			"wait_for_elb_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"warm_pool": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_reuse_policy": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"reuse_on_scale_in": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"max_group_prepared_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  defaultWarmPoolMaxGroupPreparedCapacity,
						},
						"min_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"pool_state": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  awstypes.WarmPoolStateStopped,
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

func GroupStateUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if rawState == nil {
		rawState = map[string]interface{}{}
	}

	if _, ok := rawState["ignore_failed_scaling_activities"]; !ok {
		rawState["ignore_failed_scaling_activities"] = "false"
	}

	return rawState, nil
}
