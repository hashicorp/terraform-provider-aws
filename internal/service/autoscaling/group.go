// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_autoscaling_group")
func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGroupCreate,
		ReadWithoutTimeout:   resourceGroupRead,
		UpdateWithoutTimeout: resourceGroupUpdate,
		DeleteWithoutTimeout: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceGroupV0().CoreConfigSchema().ImpliedType(),
				Upgrade: GroupStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zones": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"vpc_zone_identifier"},
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(DesiredCapacityType_Values(), false),
			},
			"enabled_metrics": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"force_delete": {
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
			"ignore_failed_scaling_activities": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"initial_lifecycle_hook": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_result": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(lifecycleHookDefaultResult_Values(), false),
						},
						"heartbeat_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(30, 7200),
						},
						"lifecycle_transition": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(lifecycleHookLifecycleTransition_Values(), false),
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 255),
								validation.StringMatch(regexache.MustCompile(`[A-Za-z0-9\-_\/]+`),
									`no spaces or special characters except "-", "_", and "/"`),
							),
						},
						"notification_metadata": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"notification_target_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
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
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
									},
									"checkpoint_percentages": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
										},
									},
									"instance_warmup": {
										Type:         nullable.TypeNullableInt,
										Optional:     true,
										ValidateFunc: nullable.ValidateTypeStringNullableIntAtLeast(0),
									},
									"min_healthy_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      90,
										ValidateFunc: validation.IntBetween(0, 100),
									},
									"scale_in_protected_instances": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      autoscaling.ScaleInProtectedInstancesIgnore,
										ValidateFunc: validation.StringInSlice(autoscaling.ScaleInProtectedInstances_Values(), false),
									},
									"skip_matching": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"standby_instances": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      autoscaling.StandbyInstancesIgnore,
										ValidateFunc: validation.StringInSlice(autoscaling.StandbyInstances_Values(), false),
									},
								},
							},
						},
						"strategy": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(autoscaling.RefreshStrategy_Values(), false),
						},
						"triggers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validateGroupInstanceRefreshTriggerFields,
							},
						},
					},
				},
			},
			"launch_configuration": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"launch_configuration", "launch_template", "mixed_instances_policy"},
			},
			"launch_template": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  verify.ValidLaunchTemplateID,
							ConflictsWith: []string{"launch_template.0.name"},
						},
						"name": {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  verify.ValidLaunchTemplateName,
							ConflictsWith: []string{"launch_template.0.id"},
						},
						"version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
				ExactlyOneOf: []string{"launch_configuration", "launch_template", "mixed_instances_policy"},
			},
			"load_balancers": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"traffic_source"},
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
				Default:  DefaultEnabledMetricsGranularity,
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
				Computed: true,
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
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"on_demand_percentage_above_base_capacity": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntBetween(0, 100),
									},
									"spot_allocation_strategy": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"spot_instance_pools": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"spot_max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"launch_template": {
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
												"version": {
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
										Computed: true,
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
																		"max": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(0),
																		},
																		"min": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																	},
																},
															},
															"accelerator_manufacturers": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringInSlice(autoscaling.AcceleratorManufacturer_Values(), false),
																},
															},
															"accelerator_names": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringInSlice(autoscaling.AcceleratorName_Values(), false),
																},
															},
															"accelerator_total_memory_mib": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"max": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		"min": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																	},
																},
															},
															"accelerator_types": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringInSlice(autoscaling.AcceleratorType_Values(), false),
																},
															},
															"allowed_instance_types": {
																Type:     schema.TypeSet,
																Optional: true,
																MaxItems: 400,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"bare_metal": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(autoscaling.BareMetal_Values(), false),
															},
															"baseline_ebs_bandwidth_mbps": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"max": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		"min": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																	},
																},
															},
															"burstable_performance": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(autoscaling.BurstablePerformance_Values(), false),
															},
															"cpu_manufacturers": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringInSlice(autoscaling.CpuManufacturer_Values(), false),
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
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringInSlice(autoscaling.InstanceGeneration_Values(), false),
																},
															},
															"local_storage": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(autoscaling.LocalStorage_Values(), false),
															},
															"local_storage_types": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringInSlice(autoscaling.LocalStorageType_Values(), false),
																},
															},
															"memory_gib_per_vcpu": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"max": {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
																		},
																		"min": {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
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
																		"max": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		"min": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
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
																		"max": {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
																		},
																		"min": {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
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
																		"max": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		"min": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																	},
																},
															},
															"on_demand_max_price_percentage_over_lowest_price": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"require_hibernate_support": {
																Type:     schema.TypeBool,
																Optional: true,
															},
															"spot_max_price_percentage_over_lowest_price": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"total_local_storage_gb": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"max": {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
																		},
																		"min": {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
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
																		"max": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		"min": {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																	},
																},
															},
														},
													},
												},
												"instance_type": {
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
															"version": {
																Type:     schema.TypeString,
																Optional: true,
																Default:  "$Default",
															},
														},
													},
												},
												"weighted_capacity": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[1-9][0-9]{0,2}$`), "see https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_LaunchTemplateOverrides.html"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
				ExactlyOneOf: []string{"launch_configuration", "launch_template", "mixed_instances_policy"},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringLenBetween(0, 255),
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringLenBetween(0, 255-id.UniqueIDSuffixLength),
				ConflictsWith: []string{"name"},
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
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"propagate_at_launch": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"target_group_arns": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"traffic_source"},
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
						"identifier": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
					},
				},
				ConflictsWith: []string{"load_balancers", "target_group_arns"},
			},
			"vpc_zone_identifier": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"availability_zones"},
			},
			"wait_for_capacity_timeout": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "10m",
				ValidateFunc: verify.ValidDuration,
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
							Default:  DefaultWarmPoolMaxGroupPreparedCapacity,
						},
						"min_size": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"pool_state": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      autoscaling.WarmPoolStateStopped,
							ValidateFunc: validation.StringInSlice(autoscaling.WarmPoolState_Values(), false),
						},
					},
				},
			},
			"warm_pool_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		CustomizeDiff: customdiff.Sequence(
			launchTemplateCustomDiff("launch_template", "launch_template.0.name"),
			launchTemplateCustomDiff("mixed_instances_policy", "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name"),
			launchTemplateCustomDiff("mixed_instances_policy", "mixed_instances_policy.0.launch_template.0.override"),
		),
	}
}

func launchTemplateCustomDiff(baseAttribute, subAttribute string) schema.CustomizeDiffFunc {
	return func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
		if diff.HasChange(subAttribute) {
			n := diff.Get(baseAttribute)
			ba, ok := n.([]interface{})
			if !ok {
				return nil
			}

			if baseAttribute == "launch_template" {
				launchTemplate := ba[0].(map[string]interface{})
				launchTemplate["id"] = launchTemplateIDUnknown

				if err := diff.SetNew(baseAttribute, ba); err != nil {
					return err
				}
			}

			if baseAttribute == "mixed_instances_policy" && !strings.Contains(subAttribute, "override") {
				launchTemplate := ba[0].(map[string]interface{})["launch_template"].([]interface{})[0].(map[string]interface{})["launch_template_specification"].([]interface{})[0]
				launchTemplateSpecification := launchTemplate.(map[string]interface{})
				launchTemplateSpecification["launch_template_id"] = launchTemplateIDUnknown

				if err := diff.SetNew(baseAttribute, ba); err != nil {
					return err
				}
			}

			if baseAttribute == "mixed_instances_policy" && strings.Contains(subAttribute, "override") {
				launchTemplate := ba[0].(map[string]interface{})["launch_template"].([]interface{})[0].(map[string]interface{})["override"].([]interface{})

				for i := range launchTemplate {
					key := fmt.Sprintf("mixed_instances_policy.0.launch_template.0.override.%d.launch_template_specification.0.launch_template_name", i)

					if diff.HasChange(key) {
						launchTemplateSpecification := launchTemplate[i].(map[string]interface{})["launch_template_specification"].([]interface{})[0].(map[string]interface{})
						launchTemplateSpecification["launch_template_id"] = launchTemplateIDUnknown
					}
				}

				if err := diff.SetNew(baseAttribute, ba); err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func resourceGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)

	startTime := time.Now()

	asgName := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	createInput := &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName:             aws.String(asgName),
		NewInstancesProtectedFromScaleIn: aws.Bool(d.Get("protect_from_scale_in").(bool)),
	}
	updateInput := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asgName),
	}

	initialLifecycleHooks := d.Get("initial_lifecycle_hook").(*schema.Set).List()
	twoPhases := len(initialLifecycleHooks) > 0

	maxSize := d.Get("max_size").(int)
	minSize := d.Get("min_size").(int)
	desiredCapacity := d.Get("desired_capacity").(int)

	if twoPhases {
		createInput.MaxSize = aws.Int64(0)
		createInput.MinSize = aws.Int64(0)

		updateInput.MaxSize = aws.Int64(int64(maxSize))
		updateInput.MinSize = aws.Int64(int64(minSize))

		if desiredCapacity > 0 {
			updateInput.DesiredCapacity = aws.Int64(int64(desiredCapacity))
		}

		if v, ok := d.GetOk("desired_capacity_type"); ok {
			updateInput.DesiredCapacityType = aws.String(v.(string))
		}
	} else {
		createInput.MaxSize = aws.Int64(int64(maxSize))
		createInput.MinSize = aws.Int64(int64(minSize))

		if desiredCapacity > 0 {
			createInput.DesiredCapacity = aws.Int64(int64(desiredCapacity))
		}

		if v, ok := d.GetOk("desired_capacity_type"); ok {
			createInput.DesiredCapacityType = aws.String(v.(string))
		}
	}

	if v, ok := d.GetOk("availability_zones"); ok && v.(*schema.Set).Len() > 0 {
		createInput.AvailabilityZones = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("capacity_rebalance"); ok {
		createInput.CapacityRebalance = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("context"); ok {
		createInput.Context = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_cooldown"); ok {
		createInput.DefaultCooldown = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("default_instance_warmup"); ok {
		createInput.DefaultInstanceWarmup = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("health_check_type"); ok {
		createInput.HealthCheckType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_grace_period"); ok {
		createInput.HealthCheckGracePeriod = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("launch_configuration"); ok {
		createInput.LaunchConfigurationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_template"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		createInput.LaunchTemplate = expandLaunchTemplateSpecification(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("load_balancers"); ok && v.(*schema.Set).Len() > 0 {
		createInput.LoadBalancerNames = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("max_instance_lifetime"); ok {
		createInput.MaxInstanceLifetime = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("mixed_instances_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		createInput.MixedInstancesPolicy = expandMixedInstancesPolicy(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("placement_group"); ok {
		createInput.PlacementGroup = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_linked_role_arn"); ok {
		createInput.ServiceLinkedRoleARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag"); ok {
		createInput.Tags = Tags(KeyValueTags(ctx, v, asgName, TagResourceTypeGroup).IgnoreAWS())
	}

	if v, ok := d.GetOk("target_group_arns"); ok && len(v.(*schema.Set).List()) > 0 {
		createInput.TargetGroupARNs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("termination_policies"); ok && len(v.([]interface{})) > 0 {
		createInput.TerminationPolicies = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("traffic_source"); ok && v.(*schema.Set).Len() > 0 {
		createInput.TrafficSources = expandTrafficSourceIdentifiers(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("vpc_zone_identifier"); ok && v.(*schema.Set).Len() > 0 {
		createInput.VPCZoneIdentifier = expandVPCZoneIdentifiers(v.(*schema.Set).List())
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateAutoScalingGroupWithContext(ctx, createInput)
		},
		// ValidationError: You must use a valid fully-formed launch template. Value (tf-acc-test-6643732652421074386) for parameter iamInstanceProfile.name is invalid. Invalid IAM Instance Profile name
		ErrCodeValidationError, "Invalid IAM Instance Profile")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Group (%s): %s", asgName, err)
	}

	d.SetId(asgName)

	if twoPhases {
		for _, input := range expandPutLifecycleHookInputs(asgName, initialLifecycleHooks) {
			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 5*time.Minute,
				func() (interface{}, error) {
					return conn.PutLifecycleHookWithContext(ctx, input)
				},
				ErrCodeValidationError, "Unable to publish test message to notification target")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Group (%s) Lifecycle Hook: %s", d.Id(), err)
			}
		}

		_, err = conn.UpdateAutoScalingGroupWithContext(ctx, updateInput)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting Auto Scaling Group (%s) initial capacity: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("wait_for_capacity_timeout"); ok {
		if timeout, _ := time.ParseDuration(v.(string)); timeout > 0 {
			// On creation all targets are minimums.
			f := func(nASG, nELB int) error {
				minSize := minSize
				if desiredCapacity > 0 {
					minSize = desiredCapacity
				}

				if nASG < minSize {
					return fmt.Errorf("want at least %d healthy instance(s) in Auto Scaling Group, have %d", minSize, nASG)
				}

				minELBCapacity := d.Get("min_elb_capacity").(int)
				if waitForELBCapacity := d.Get("wait_for_elb_capacity").(int); waitForELBCapacity > 0 {
					minELBCapacity = waitForELBCapacity
				}

				if nELB < minELBCapacity {
					return fmt.Errorf("want at least %d healthy instance(s) registered to Load Balancer, have %d", minELBCapacity, nELB)
				}

				return nil
			}

			if err := waitGroupCapacitySatisfied(ctx, conn, meta.(*conns.AWSClient).ELBConn(ctx), meta.(*conns.AWSClient).ELBV2Conn(ctx), d.Id(), f, startTime, d.Get("ignore_failed_scaling_activities").(bool), timeout); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) capacity satisfied: %s", d.Id(), err)
			}
		}
	}

	if v, ok := d.GetOk("suspended_processes"); ok && v.(*schema.Set).Len() > 0 {
		input := &autoscaling.ScalingProcessQuery{
			AutoScalingGroupName: aws.String(d.Id()),
			ScalingProcesses:     flex.ExpandStringSet(v.(*schema.Set)),
		}

		_, err := conn.SuspendProcessesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "suspending Auto Scaling Group (%s) scaling processes: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("enabled_metrics"); ok && v.(*schema.Set).Len() > 0 {
		input := &autoscaling.EnableMetricsCollectionInput{
			AutoScalingGroupName: aws.String(d.Id()),
			Granularity:          aws.String(d.Get("metrics_granularity").(string)),
			Metrics:              flex.ExpandStringSet(v.(*schema.Set)),
		}

		_, err := conn.EnableMetricsCollectionWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Auto Scaling Group (%s) metrics collection: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("warm_pool"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		_, err := conn.PutWarmPoolWithContext(ctx, expandPutWarmPoolInput(d.Id(), v.([]interface{})[0].(map[string]interface{})))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Warm Pool (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	g, err := FindGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Group %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Group (%s): %s", d.Id(), err)
	}

	d.Set("arn", g.AutoScalingGroupARN)
	d.Set("availability_zones", aws.StringValueSlice(g.AvailabilityZones))
	d.Set("capacity_rebalance", g.CapacityRebalance)
	d.Set("context", g.Context)
	d.Set("default_cooldown", g.DefaultCooldown)
	d.Set("default_instance_warmup", g.DefaultInstanceWarmup)
	d.Set("desired_capacity", g.DesiredCapacity)
	d.Set("desired_capacity_type", g.DesiredCapacityType)
	if len(g.EnabledMetrics) > 0 {
		d.Set("enabled_metrics", flattenEnabledMetrics(g.EnabledMetrics))
		d.Set("metrics_granularity", g.EnabledMetrics[0].Granularity)
	} else {
		d.Set("enabled_metrics", nil)
		d.Set("metrics_granularity", DefaultEnabledMetricsGranularity)
	}
	d.Set("health_check_grace_period", g.HealthCheckGracePeriod)
	d.Set("health_check_type", g.HealthCheckType)
	d.Set("launch_configuration", g.LaunchConfigurationName)
	if g.LaunchTemplate != nil {
		if err := d.Set("launch_template", []interface{}{flattenLaunchTemplateSpecification(g.LaunchTemplate)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting launch_template: %s", err)
		}
	} else {
		d.Set("launch_template", nil)
	}
	d.Set("load_balancers", aws.StringValueSlice(g.LoadBalancerNames))
	d.Set("max_instance_lifetime", g.MaxInstanceLifetime)
	d.Set("max_size", g.MaxSize)
	d.Set("min_size", g.MinSize)
	if g.MixedInstancesPolicy != nil {
		if err := d.Set("mixed_instances_policy", []interface{}{flattenMixedInstancesPolicy(g.MixedInstancesPolicy)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting mixed_instances_policy: %s", err)
		}
	} else {
		d.Set("mixed_instances_policy", nil)
	}
	d.Set("name", g.AutoScalingGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(g.AutoScalingGroupName)))
	d.Set("placement_group", g.PlacementGroup)
	d.Set("predicted_capacity", g.PredictedCapacity)
	d.Set("protect_from_scale_in", g.NewInstancesProtectedFromScaleIn)
	d.Set("service_linked_role_arn", g.ServiceLinkedRoleARN)
	d.Set("suspended_processes", flattenSuspendedProcesses(g.SuspendedProcesses))
	if err := d.Set("traffic_source", flattenTrafficSourceIdentifiers(g.TrafficSources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting traffic_source: %s", err)
	}
	d.Set("target_group_arns", aws.StringValueSlice(g.TargetGroupARNs))
	// If no termination polices are explicitly configured and the upstream state
	// is only using the "Default" policy, clear the state to make it consistent
	// with the default AWS Create API behavior.
	if _, ok := d.GetOk("termination_policies"); !ok && len(g.TerminationPolicies) == 1 && aws.StringValue(g.TerminationPolicies[0]) == DefaultTerminationPolicy {
		d.Set("termination_policies", nil)
	} else {
		d.Set("termination_policies", aws.StringValueSlice(g.TerminationPolicies))
	}
	if len(aws.StringValue(g.VPCZoneIdentifier)) > 0 {
		d.Set("vpc_zone_identifier", strings.Split(aws.StringValue(g.VPCZoneIdentifier), ","))
	} else {
		d.Set("vpc_zone_identifier", nil)
	}
	if g.WarmPoolConfiguration != nil {
		if err := d.Set("warm_pool", []interface{}{flattenWarmPoolConfiguration(g.WarmPoolConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting warm_pool: %s", err)
		}
	} else {
		d.Set("warm_pool", nil)
	}
	d.Set("warm_pool_size", g.WarmPoolSize)

	if err := d.Set("tag", ListOfMap(KeyValueTags(ctx, g.Tags, d.Id(), TagResourceTypeGroup).IgnoreAWS().IgnoreConfig(ignoreTagsConfig))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tag: %s", err)
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)

	startTime := time.Now()

	var shouldWaitForCapacity bool
	var shouldRefreshInstances bool

	if d.HasChangesExcept(
		"enabled_metrics",
		"load_balancers",
		"suspended_processes",
		"tag",
		"target_group_arns",
		"traffic_source",
		"warm_pool",
	) {
		input := &autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName:             aws.String(d.Id()),
			NewInstancesProtectedFromScaleIn: aws.Bool(d.Get("protect_from_scale_in").(bool)),
		}

		if d.HasChange("availability_zones") {
			if v, ok := d.GetOk("availability_zones"); ok && v.(*schema.Set).Len() > 0 {
				input.AvailabilityZones = flex.ExpandStringSet(v.(*schema.Set))
			}
		}

		if d.HasChange("capacity_rebalance") {
			// If the capacity rebalance field is set to null, we need to explicitly set
			// it back to "false", or the API won't reset it for us.
			if v, ok := d.GetOk("capacity_rebalance"); ok {
				input.CapacityRebalance = aws.Bool(v.(bool))
			} else {
				input.CapacityRebalance = aws.Bool(false)
			}
		}

		if d.HasChange("context") {
			input.Context = aws.String(d.Get("context").(string))
		}

		if d.HasChange("default_cooldown") {
			input.DefaultCooldown = aws.Int64(int64(d.Get("default_cooldown").(int)))
		}

		if d.HasChange("default_instance_warmup") {
			input.DefaultInstanceWarmup = aws.Int64(int64(d.Get("default_instance_warmup").(int)))
		}

		if d.HasChange("desired_capacity") {
			input.DesiredCapacity = aws.Int64(int64(d.Get("desired_capacity").(int)))
			shouldWaitForCapacity = true
		}

		if d.HasChange("desired_capacity_type") {
			input.DesiredCapacityType = aws.String(d.Get("desired_capacity_type").(string))
			shouldWaitForCapacity = true
		}

		if d.HasChange("health_check_grace_period") {
			input.HealthCheckGracePeriod = aws.Int64(int64(d.Get("health_check_grace_period").(int)))
		}

		if d.HasChange("health_check_type") {
			input.HealthCheckGracePeriod = aws.Int64(int64(d.Get("health_check_grace_period").(int)))
			input.HealthCheckType = aws.String(d.Get("health_check_type").(string))
		}

		if d.HasChange("launch_configuration") {
			if v, ok := d.GetOk("launch_configuration"); ok {
				input.LaunchConfigurationName = aws.String(v.(string))
			}
			shouldRefreshInstances = true
		}

		if d.HasChange("launch_template") {
			if v, ok := d.GetOk("launch_template"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.LaunchTemplate = expandLaunchTemplateSpecification(v.([]interface{})[0].(map[string]interface{}))
			}
			shouldRefreshInstances = true
		}

		if d.HasChange("max_instance_lifetime") {
			input.MaxInstanceLifetime = aws.Int64(int64(d.Get("max_instance_lifetime").(int)))
		}

		if d.HasChange("max_size") {
			input.MaxSize = aws.Int64(int64(d.Get("max_size").(int)))
		}

		if d.HasChange("min_size") {
			input.MinSize = aws.Int64(int64(d.Get("min_size").(int)))
			shouldWaitForCapacity = true
		}

		if d.HasChange("mixed_instances_policy") {
			if v, ok := d.GetOk("mixed_instances_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.MixedInstancesPolicy = expandMixedInstancesPolicy(v.([]interface{})[0].(map[string]interface{}))
			}
			shouldRefreshInstances = true
		}

		if d.HasChange("placement_group") {
			input.PlacementGroup = aws.String(d.Get("placement_group").(string))
		}

		if d.HasChange("service_linked_role_arn") {
			input.ServiceLinkedRoleARN = aws.String(d.Get("service_linked_role_arn").(string))
		}

		if d.HasChange("termination_policies") {
			// If the termination policy is set to null, we need to explicitly set
			// it back to "Default", or the API won't reset it for us.
			if v, ok := d.GetOk("termination_policies"); ok && len(v.([]interface{})) > 0 {
				input.TerminationPolicies = flex.ExpandStringList(v.([]interface{}))
			} else {
				input.TerminationPolicies = aws.StringSlice([]string{DefaultTerminationPolicy})
			}
		}

		if d.HasChange("vpc_zone_identifier") {
			input.VPCZoneIdentifier = expandVPCZoneIdentifiers(d.Get("vpc_zone_identifier").(*schema.Set).List())
		}

		_, err := conn.UpdateAutoScalingGroupWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChanges("tag") {
		oTagRaw, nTagRaw := d.GetChange("tag")

		oldTags := Tags(KeyValueTags(ctx, oTagRaw, d.Id(), TagResourceTypeGroup))
		newTags := Tags(KeyValueTags(ctx, nTagRaw, d.Id(), TagResourceTypeGroup))

		if err := updateTags(ctx, conn, d.Id(), TagResourceTypeGroup, oldTags, newTags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags for Auto Scaling Group (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("traffic_source") {
		o, n := d.GetChange("traffic_source")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// API only supports adding or removing 10 at a time.
		batchSize := 10
		for _, chunk := range slices.Chunks(expandTrafficSourceIdentifiers(os.Difference(ns).List()), batchSize) {
			_, err := conn.DetachTrafficSourcesWithContext(ctx, &autoscaling.DetachTrafficSourcesInput{
				AutoScalingGroupName: aws.String(d.Id()),
				TrafficSources:       chunk,
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "detaching Auto Scaling Group (%s) traffic sources: %s", d.Id(), err)
			}

			if _, err := waitTrafficSourcesDeleted(ctx, conn, d.Id(), "", d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) traffic sources removed: %s", d.Id(), err)
			}
		}

		for _, chunk := range slices.Chunks(expandTrafficSourceIdentifiers(ns.Difference(os).List()), batchSize) {
			_, err := conn.AttachTrafficSourcesWithContext(ctx, &autoscaling.AttachTrafficSourcesInput{
				AutoScalingGroupName: aws.String(d.Id()),
				TrafficSources:       chunk,
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "attaching Auto Scaling Group (%s) traffic sources: %s", d.Id(), err)
			}

			if _, err := waitTrafficSourcesCreated(ctx, conn, d.Id(), "", d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) traffic sources added: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("load_balancers") {
		o, n := d.GetChange("load_balancers")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// API only supports adding or removing 10 at a time.
		batchSize := 10
		for _, chunk := range slices.Chunks(flex.ExpandStringSet(os.Difference(ns)), batchSize) {
			_, err := conn.DetachLoadBalancersWithContext(ctx, &autoscaling.DetachLoadBalancersInput{
				AutoScalingGroupName: aws.String(d.Id()),
				LoadBalancerNames:    chunk,
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "detaching Auto Scaling Group (%s) load balancers: %s", d.Id(), err)
			}

			if _, err := waitLoadBalancersRemoved(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) load balancers removed: %s", d.Id(), err)
			}
		}

		for _, chunk := range slices.Chunks(flex.ExpandStringSet(ns.Difference(os)), batchSize) {
			_, err := conn.AttachLoadBalancersWithContext(ctx, &autoscaling.AttachLoadBalancersInput{
				AutoScalingGroupName: aws.String(d.Id()),
				LoadBalancerNames:    chunk,
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "attaching Auto Scaling Group (%s) load balancers: %s", d.Id(), err)
			}

			if _, err := waitLoadBalancersAdded(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) load balancers added: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("target_group_arns") {
		o, n := d.GetChange("target_group_arns")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		// API only supports adding or removing 10 at a time.
		batchSize := 10
		for _, chunk := range slices.Chunks(flex.ExpandStringSet(os.Difference(ns)), batchSize) {
			_, err := conn.DetachLoadBalancerTargetGroupsWithContext(ctx, &autoscaling.DetachLoadBalancerTargetGroupsInput{
				AutoScalingGroupName: aws.String(d.Id()),
				TargetGroupARNs:      chunk,
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "detaching Auto Scaling Group (%s) target groups: %s", d.Id(), err)
			}

			if _, err := waitLoadBalancerTargetGroupsRemoved(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) target groups removed: %s", d.Id(), err)
			}
		}

		for _, chunk := range slices.Chunks(flex.ExpandStringSet(ns.Difference(os)), batchSize) {
			_, err := conn.AttachLoadBalancerTargetGroupsWithContext(ctx, &autoscaling.AttachLoadBalancerTargetGroupsInput{
				AutoScalingGroupName: aws.String(d.Id()),
				TargetGroupARNs:      chunk,
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "attaching Auto Scaling Group (%s) target groups: %s", d.Id(), err)
			}

			if _, err := waitLoadBalancerTargetGroupsAdded(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) target groups added: %s", d.Id(), err)
			}
		}
	}

	if v, ok := d.GetOk("instance_refresh"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		if !shouldRefreshInstances {
			if v, ok := tfMap["triggers"].(*schema.Set); ok && v.Len() > 0 {
				var triggers []string

				for _, v := range v.List() {
					if v := v.(string); v != "" {
						triggers = append(triggers, v)
					}
				}

				shouldRefreshInstances = d.HasChanges(triggers...)
			}
		}

		if shouldRefreshInstances {
			var launchTemplate *autoscaling.LaunchTemplateSpecification

			if v, ok := d.GetOk("launch_template"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				launchTemplate = expandLaunchTemplateSpecification(v.([]interface{})[0].(map[string]interface{}))
			}

			var mixedInstancesPolicy *autoscaling.MixedInstancesPolicy

			if v, ok := d.GetOk("mixed_instances_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				mixedInstancesPolicy = expandMixedInstancesPolicy(v.([]interface{})[0].(map[string]interface{}))
			}

			if err := startInstanceRefresh(ctx, conn, expandStartInstanceRefreshInput(d.Id(), tfMap, launchTemplate, mixedInstancesPolicy)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("warm_pool") {
		w := d.Get("warm_pool").([]interface{})

		// No warm pool exists in new config. Delete it.
		if len(w) == 0 || w[0] == nil {
			forceDeleteWarmPool := d.Get("force_delete").(bool) || d.Get("force_delete_warm_pool").(bool)

			if err := deleteWarmPool(ctx, conn, d.Id(), forceDeleteWarmPool, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			_, err := conn.PutWarmPoolWithContext(ctx, expandPutWarmPoolInput(d.Id(), w[0].(map[string]interface{})))

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Auto Scaling Warm Pool (%s): %s", d.Id(), err)
			}
		}
	}

	if shouldWaitForCapacity {
		if v, ok := d.GetOk("wait_for_capacity_timeout"); ok {
			if timeout, _ := time.ParseDuration(v.(string)); timeout > 0 {
				// On update all targets are specific.
				f := func(nASG, nELB int) error {
					minSize := d.Get("min_size").(int)
					if desiredCapacity := d.Get("desired_capacity").(int); desiredCapacity > minSize {
						minSize = desiredCapacity
					}

					if nASG != minSize {
						return fmt.Errorf("want exactly %d healthy instance(s) in Auto Scaling Group, have %d", minSize, nASG)
					}

					if waitForELBCapacity := d.Get("wait_for_elb_capacity").(int); waitForELBCapacity > 0 {
						if nELB != waitForELBCapacity {
							return fmt.Errorf("want exactly %d healthy instance(s) registered to Load Balancer, have %d", waitForELBCapacity, nELB)
						}
					}

					return nil
				}

				if err := waitGroupCapacitySatisfied(ctx, conn, meta.(*conns.AWSClient).ELBConn(ctx), meta.(*conns.AWSClient).ELBV2Conn(ctx), d.Id(), f, startTime, d.Get("ignore_failed_scaling_activities").(bool), timeout); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) capacity satisfied: %s", d.Id(), err)
				}
			}
		}
	}

	if d.HasChange("enabled_metrics") {
		o, n := d.GetChange("enabled_metrics")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		if disableMetrics := os.Difference(ns); disableMetrics.Len() != 0 {
			input := &autoscaling.DisableMetricsCollectionInput{
				AutoScalingGroupName: aws.String(d.Id()),
				Metrics:              flex.ExpandStringSet(disableMetrics),
			}

			_, err := conn.DisableMetricsCollectionWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling Auto Scaling Group (%s) metrics collection: %s", d.Id(), err)
			}
		}

		if enableMetrics := ns.Difference(os); enableMetrics.Len() != 0 {
			input := &autoscaling.EnableMetricsCollectionInput{
				AutoScalingGroupName: aws.String(d.Id()),
				Granularity:          aws.String(d.Get("metrics_granularity").(string)),
				Metrics:              flex.ExpandStringSet(enableMetrics),
			}

			_, err := conn.EnableMetricsCollectionWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling Auto Scaling Group (%s) metrics collection: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("suspended_processes") {
		o, n := d.GetChange("suspended_processes")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		if resumeProcesses := os.Difference(ns); resumeProcesses.Len() != 0 {
			input := &autoscaling.ScalingProcessQuery{
				AutoScalingGroupName: aws.String(d.Id()),
				ScalingProcesses:     flex.ExpandStringSet(resumeProcesses),
			}

			_, err := conn.ResumeProcessesWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "resuming Auto Scaling Group (%s) scaling processes: %s", d.Id(), err)
			}
		}

		if suspendProcesses := ns.Difference(os); suspendProcesses.Len() != 0 {
			input := &autoscaling.ScalingProcessQuery{
				AutoScalingGroupName: aws.String(d.Id()),
				ScalingProcesses:     flex.ExpandStringSet(suspendProcesses),
			}

			_, err := conn.SuspendProcessesWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "suspending Auto Scaling Group (%s) scaling processes: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingConn(ctx)

	forceDeleteGroup := d.Get("force_delete").(bool)
	forceDeleteWarmPool := forceDeleteGroup || d.Get("force_delete_warm_pool").(bool)

	group, err := FindGroupByName(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Group (%s): %s", d.Id(), err)
	}

	if group.WarmPoolConfiguration != nil {
		err = deleteWarmPool(ctx, conn, d.Id(), forceDeleteWarmPool, d.Timeout(schema.TimeoutDelete))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if !forceDeleteGroup {
		err = drainGroup(ctx, conn, d.Id(), group.Instances, d.Timeout(schema.TimeoutDelete))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[DEBUG] Deleting Auto Scaling Group: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteAutoScalingGroupWithContext(ctx, &autoscaling.DeleteAutoScalingGroupInput{
				AutoScalingGroupName: aws.String(d.Id()),
				ForceDelete:          aws.Bool(forceDeleteGroup),
			})
		},
		autoscaling.ErrCodeResourceInUseFault, autoscaling.ErrCodeScalingActivityInProgressFault)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Group (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return FindGroupByName(ctx, conn, d.Id())
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func drainGroup(ctx context.Context, conn *autoscaling.AutoScaling, name string, instances []*autoscaling.Instance, timeout time.Duration) error {
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(name),
		DesiredCapacity:      aws.Int64(0),
		MinSize:              aws.Int64(0),
		MaxSize:              aws.Int64(0),
	}

	log.Printf("[DEBUG] Draining Auto Scaling Group: %s", name)
	if _, err := conn.UpdateAutoScalingGroupWithContext(ctx, input); err != nil {
		return fmt.Errorf("setting Auto Scaling Group (%s) capacity to 0: %w", name, err)
	}

	// Next, ensure that instances are not prevented from scaling in.
	//
	// The ASG's own scale-in protection setting doesn't make a difference here,
	// as it only affects new instances, which won't be launched now that the
	// desired capacity is set to 0. There is also the possibility that this ASG
	// no longer applies scale-in protection to new instances, but there's still
	// old ones that have it.

	const chunkSize = 50 // API limit.
	for i, n := 0, len(instances); i < n; i += chunkSize {
		j := i + chunkSize
		if j > n {
			j = n
		}

		var instanceIDs []string

		for k := i; k < j; k++ {
			instanceIDs = append(instanceIDs, aws.StringValue(instances[k].InstanceId))
		}

		input := &autoscaling.SetInstanceProtectionInput{
			AutoScalingGroupName: aws.String(name),
			InstanceIds:          aws.StringSlice(instanceIDs),
			ProtectedFromScaleIn: aws.Bool(false),
		}

		if _, err := conn.SetInstanceProtectionWithContext(ctx, input); err != nil {
			return fmt.Errorf("disabling Auto Scaling Group (%s) scale-in protections: %w", name, err)
		}
	}

	if _, err := waitGroupDrained(ctx, conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Auto Scaling Group (%s) drain: %w", name, err)
	}

	return nil
}

func deleteWarmPool(ctx context.Context, conn *autoscaling.AutoScaling, name string, force bool, timeout time.Duration) error {
	if !force {
		if err := drainWarmPool(ctx, conn, name, timeout); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Auto Scaling Warm Pool: %s", name)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout,
		func() (interface{}, error) {
			return conn.DeleteWarmPoolWithContext(ctx, &autoscaling.DeleteWarmPoolInput{
				AutoScalingGroupName: aws.String(name),
				ForceDelete:          aws.Bool(force),
			})
		},
		autoscaling.ErrCodeResourceInUseFault, autoscaling.ErrCodeScalingActivityInProgressFault)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "No warm pool found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Auto Scaling Warm Pool (%s): %w", name, err)
	}

	if _, err := waitWarmPoolDeleted(ctx, conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Auto Scaling Warm Pool (%s) delete: %w", name, err)
	}

	return nil
}

func drainWarmPool(ctx context.Context, conn *autoscaling.AutoScaling, name string, timeout time.Duration) error {
	input := &autoscaling.PutWarmPoolInput{
		AutoScalingGroupName:     aws.String(name),
		MaxGroupPreparedCapacity: aws.Int64(0),
		MinSize:                  aws.Int64(0),
	}

	log.Printf("[DEBUG] Draining Auto Scaling Warm Pool: %s", name)
	if _, err := conn.PutWarmPoolWithContext(ctx, input); err != nil {
		return fmt.Errorf("setting Auto Scaling Warm Pool (%s) capacity to 0: %w", name, err)
	}

	if _, err := waitWarmPoolDrained(ctx, conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Auto Scaling Warm Pool (%s) drain: %w", name, err)
	}

	return nil
}

func findELBInstanceStates(ctx context.Context, conn *elb.ELB, g *autoscaling.Group) (map[string]map[string]string, error) {
	instanceStates := make(map[string]map[string]string)

	for _, v := range g.LoadBalancerNames {
		lbName := aws.StringValue(v)
		input := &elb.DescribeInstanceHealthInput{
			LoadBalancerName: aws.String(lbName),
		}

		output, err := conn.DescribeInstanceHealthWithContext(ctx, input)

		if err != nil {
			return nil, fmt.Errorf("reading load balancer (%s) instance health: %w", lbName, err)
		}

		instanceStates[lbName] = make(map[string]string)

		for _, v := range output.InstanceStates {
			instanceID := aws.StringValue(v.InstanceId)
			if instanceID == "" {
				continue
			}
			state := aws.StringValue(v.State)
			if state == "" {
				continue
			}

			instanceStates[lbName][instanceID] = state
		}
	}

	return instanceStates, nil
}

func findELBV2InstanceStates(ctx context.Context, conn *elbv2.ELBV2, g *autoscaling.Group) (map[string]map[string]string, error) {
	instanceStates := make(map[string]map[string]string)

	for _, v := range g.TargetGroupARNs {
		targetGroupARN := aws.StringValue(v)
		input := &elbv2.DescribeTargetHealthInput{
			TargetGroupArn: aws.String(targetGroupARN),
		}

		output, err := conn.DescribeTargetHealthWithContext(ctx, input)

		if err != nil {
			return nil, fmt.Errorf("reading target group (%s) instance health: %w", targetGroupARN, err)
		}

		instanceStates[targetGroupARN] = make(map[string]string)

		for _, v := range output.TargetHealthDescriptions {
			if v.Target == nil || v.TargetHealth == nil {
				continue
			}

			instanceID := aws.StringValue(v.Target.Id)
			if instanceID == "" {
				continue
			}
			state := aws.StringValue(v.TargetHealth.State)
			if state == "" {
				continue
			}

			instanceStates[targetGroupARN][instanceID] = state
		}
	}

	return instanceStates, nil
}

func findGroup(ctx context.Context, conn *autoscaling.AutoScaling, input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.Group, error) {
	output, err := findGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func findGroups(ctx context.Context, conn *autoscaling.AutoScaling, input *autoscaling.DescribeAutoScalingGroupsInput) ([]*autoscaling.Group, error) {
	var output []*autoscaling.Group

	err := conn.DescribeAutoScalingGroupsPagesWithContext(ctx, input, func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.AutoScalingGroups {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindGroupByName(ctx context.Context, conn *autoscaling.AutoScaling, name string) (*autoscaling.Group, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: aws.StringSlice([]string{name}),
	}

	output, err := findGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.AutoScalingGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findInstanceRefresh(ctx context.Context, conn *autoscaling.AutoScaling, input *autoscaling.DescribeInstanceRefreshesInput) (*autoscaling.InstanceRefresh, error) {
	output, err := FindInstanceRefreshes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindInstanceRefreshes(ctx context.Context, conn *autoscaling.AutoScaling, input *autoscaling.DescribeInstanceRefreshesInput) ([]*autoscaling.InstanceRefresh, error) {
	var output []*autoscaling.InstanceRefresh

	err := describeInstanceRefreshesPages(ctx, conn, input, func(page *autoscaling.DescribeInstanceRefreshesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceRefreshes {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findLoadBalancerStates(ctx context.Context, conn *autoscaling.AutoScaling, name string) ([]*autoscaling.LoadBalancerState, error) {
	input := &autoscaling.DescribeLoadBalancersInput{
		AutoScalingGroupName: aws.String(name),
	}
	var output []*autoscaling.LoadBalancerState

	err := describeLoadBalancersPages(ctx, conn, input, func(page *autoscaling.DescribeLoadBalancersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LoadBalancers {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findLoadBalancerTargetGroupStates(ctx context.Context, conn *autoscaling.AutoScaling, name string) ([]*autoscaling.LoadBalancerTargetGroupState, error) {
	input := &autoscaling.DescribeLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(name),
	}
	var output []*autoscaling.LoadBalancerTargetGroupState

	err := describeLoadBalancerTargetGroupsPages(ctx, conn, input, func(page *autoscaling.DescribeLoadBalancerTargetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LoadBalancerTargetGroups {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findScalingActivities(ctx context.Context, conn *autoscaling.AutoScaling, input *autoscaling.DescribeScalingActivitiesInput, startTime time.Time) ([]*autoscaling.Activity, error) {
	var output []*autoscaling.Activity

	err := conn.DescribeScalingActivitiesPagesWithContext(ctx, input, func(page *autoscaling.DescribeScalingActivitiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, activity := range page.Activities {
			if activity == nil {
				continue
			}

			if startTime.Before(aws.TimeValue(activity.StartTime)) {
				output = append(output, activity)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findScalingActivitiesByName(ctx context.Context, conn *autoscaling.AutoScaling, name string, startTime time.Time) ([]*autoscaling.Activity, error) {
	input := &autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(name),
	}

	return findScalingActivities(ctx, conn, input, startTime)
}

func findTrafficSourceStates(ctx context.Context, conn *autoscaling.AutoScaling, input *autoscaling.DescribeTrafficSourcesInput) ([]*autoscaling.TrafficSourceState, error) {
	var output []*autoscaling.TrafficSourceState

	err := conn.DescribeTrafficSourcesPagesWithContext(ctx, input, func(page *autoscaling.DescribeTrafficSourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TrafficSources {
			if v == nil {
				continue
			}

			output = append(output, v)
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findTrafficSourceStatesByTwoPartKey(ctx context.Context, conn *autoscaling.AutoScaling, asgName, trafficSourceType string) ([]*autoscaling.TrafficSourceState, error) {
	input := &autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
	}
	if trafficSourceType != "" {
		input.TrafficSourceType = aws.String(trafficSourceType)
	}

	return findTrafficSourceStates(ctx, conn, input)
}

func findWarmPool(ctx context.Context, conn *autoscaling.AutoScaling, name string) (*autoscaling.DescribeWarmPoolOutput, error) {
	input := &autoscaling.DescribeWarmPoolInput{
		AutoScalingGroupName: aws.String(name),
	}
	var output *autoscaling.DescribeWarmPoolOutput

	err := describeWarmPoolPages(ctx, conn, input, func(page *autoscaling.DescribeWarmPoolOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		if output == nil {
			output = page
		} else {
			output.Instances = append(output.Instances, page.Instances...)
		}

		return !lastPage
	})

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.WarmPoolConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output, nil
}

func statusGroupCapacity(ctx context.Context, conn *autoscaling.AutoScaling, elbconn *elb.ELB, elbv2conn *elbv2.ELBV2, name string, cb func(int, int) error, startTime time.Time, ignoreFailedScalingActivities bool) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		if !ignoreFailedScalingActivities {
			// Check for fatal error in activity logs.
			scalingActivities, err := findScalingActivitiesByName(ctx, conn, name, startTime)

			if err != nil {
				return nil, "", fmt.Errorf("reading scaling activities: %w", err)
			}

			var errs []error

			for _, v := range scalingActivities {
				if statusCode := aws.StringValue(v.StatusCode); statusCode == autoscaling.ScalingActivityStatusCodeFailed && aws.Int64Value(v.Progress) == 100 {
					if strings.Contains(aws.StringValue(v.StatusMessage), "Invalid IAM Instance Profile") {
						// the activity will likely be retried
						continue
					}
					errs = append(errs, fmt.Errorf("scaling activity (%s): %s: %s", aws.StringValue(v.ActivityId), statusCode, aws.StringValue(v.StatusMessage)))
				}
			}

			err = errors.Join(errs...)

			if err != nil {
				return nil, "", err
			}
		}

		g, err := FindGroupByName(ctx, conn, name)

		if err != nil {
			return nil, "", fmt.Errorf("reading Auto Scaling Group (%s): %w", name, err)
		}

		lbInstanceStates, err := findELBInstanceStates(ctx, elbconn, g)

		if err != nil {
			return nil, "", err
		}

		targetGroupInstanceStates, err := findELBV2InstanceStates(ctx, elbv2conn, g)

		if err != nil {
			return nil, "", err
		}

		nASG := 0
		nELB := 0

		for _, v := range g.Instances {
			instanceID := aws.StringValue(v.InstanceId)
			if instanceID == "" {
				continue
			}

			if aws.StringValue(v.HealthStatus) != InstanceHealthStatusHealthy {
				continue
			}

			if aws.StringValue(v.LifecycleState) != autoscaling.LifecycleStateInService {
				continue
			}

			increment := 1
			if v := aws.StringValue(v.WeightedCapacity); v != "" {
				v, _ := strconv.Atoi(v)
				increment = v
			}

			nASG += increment

			inAll := true

			for _, v := range lbInstanceStates {
				if state, ok := v[instanceID]; ok && state != tfelb.InstanceStateInService {
					inAll = false
					break
				}
			}

			if inAll {
				for _, v := range targetGroupInstanceStates {
					if state, ok := v[instanceID]; ok && state != elbv2.TargetHealthStateEnumHealthy {
						inAll = false
						break
					}
				}
			}

			if inAll {
				nELB += increment
			}
		}

		err = cb(nASG, nELB)

		if err != nil {
			return struct{}{}, err.Error(), nil //nolint:nilerr // err is passed via the result State
		}

		return struct{}{}, "ok", nil
	}
}

func statusGroupInstanceCount(ctx context.Context, conn *autoscaling.AutoScaling, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.Itoa(len(output.Instances)), nil
	}
}

func statusInstanceRefresh(ctx context.Context, conn *autoscaling.AutoScaling, name, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: aws.String(name),
			InstanceRefreshIds:   aws.StringSlice([]string{id}),
		}

		output, err := findInstanceRefresh(ctx, conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusLoadBalancerInStateCount(ctx context.Context, conn *autoscaling.AutoScaling, name string, states ...string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLoadBalancerStates(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		var count int

		for _, v := range output {
			for _, state := range states {
				if aws.StringValue(v.State) == state {
					count++
					break
				}
			}
		}

		return output, strconv.Itoa(count), nil
	}
}

func statusLoadBalancerTargetGroupInStateCount(ctx context.Context, conn *autoscaling.AutoScaling, name string, states ...string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLoadBalancerTargetGroupStates(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		var count int

		for _, v := range output {
			for _, state := range states {
				if aws.StringValue(v.State) == state {
					count++
					break
				}
			}
		}

		return output, strconv.Itoa(count), nil
	}
}

func statusTrafficSourcesInStateCount(ctx context.Context, conn *autoscaling.AutoScaling, asgName, trafficSourceType string, states ...string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTrafficSourceStatesByTwoPartKey(ctx, conn, asgName, trafficSourceType)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		var count int

		for _, v := range output {
			for _, state := range states {
				if aws.StringValue(v.State) == state {
					count++
					break
				}
			}
		}

		return output, strconv.Itoa(count), nil
	}
}

func statusWarmPool(ctx context.Context, conn *autoscaling.AutoScaling, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findWarmPool(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.WarmPoolConfiguration, aws.StringValue(output.WarmPoolConfiguration.Status), nil
	}
}

func statusWarmPoolInstanceCount(ctx context.Context, conn *autoscaling.AutoScaling, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findWarmPool(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.Itoa(len(output.Instances)), nil
	}
}

func waitGroupCapacitySatisfied(ctx context.Context, conn *autoscaling.AutoScaling, elbconn *elb.ELB, elbv2conn *elbv2.ELBV2, name string, cb func(int, int) error, startTime time.Time, ignoreFailedScalingActivities bool, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"ok"},
		Refresh: statusGroupCapacity(ctx, conn, elbconn, elbv2conn, name, cb, startTime, ignoreFailedScalingActivities),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(struct{ err error }); ok {
		tfresource.SetLastError(err, output.err)
	}

	return err
}

func waitGroupDrained(ctx context.Context, conn *autoscaling.AutoScaling, name string, timeout time.Duration) (*autoscaling.Group, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusGroupInstanceCount(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscaling.Group); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancersAdded(ctx context.Context, conn *autoscaling.AutoScaling, name string, timeout time.Duration) ([]*autoscaling.LoadBalancerState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerInStateCount(ctx, conn, name, LoadBalancerStateAdding),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*autoscaling.LoadBalancerState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancersRemoved(ctx context.Context, conn *autoscaling.AutoScaling, name string, timeout time.Duration) ([]*autoscaling.LoadBalancerState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerInStateCount(ctx, conn, name, LoadBalancerStateRemoving),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*autoscaling.LoadBalancerState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancerTargetGroupsAdded(ctx context.Context, conn *autoscaling.AutoScaling, name string, timeout time.Duration) ([]*autoscaling.LoadBalancerTargetGroupState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerTargetGroupInStateCount(ctx, conn, name, LoadBalancerTargetGroupStateAdding),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*autoscaling.LoadBalancerTargetGroupState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancerTargetGroupsRemoved(ctx context.Context, conn *autoscaling.AutoScaling, name string, timeout time.Duration) ([]*autoscaling.LoadBalancerTargetGroupState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerTargetGroupInStateCount(ctx, conn, name, LoadBalancerTargetGroupStateRemoving),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*autoscaling.LoadBalancerTargetGroupState); ok {
		return output, err
	}

	return nil, err
}

func waitTrafficSourcesCreated(ctx context.Context, conn *autoscaling.AutoScaling, asgName, trafficSourceType string, timeout time.Duration) ([]*autoscaling.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusTrafficSourcesInStateCount(ctx, conn, asgName, trafficSourceType, TrafficSourceStateAdding),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*autoscaling.TrafficSourceState); ok {
		return output, err
	}

	return nil, err
}

func waitTrafficSourcesDeleted(ctx context.Context, conn *autoscaling.AutoScaling, asgName, trafficSourceType string, timeout time.Duration) ([]*autoscaling.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusTrafficSourcesInStateCount(ctx, conn, asgName, trafficSourceType, TrafficSourceStateRemoving),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*autoscaling.TrafficSourceState); ok {
		return output, err
	}

	return nil, err
}

const (
	// Maximum amount of time to wait for an InstanceRefresh to be started
	// Must be at least as long as instanceRefreshCancelledTimeout, since we try to cancel any
	// existing Instance Refreshes when starting.
	instanceRefreshStartedTimeout = instanceRefreshCancelledTimeout

	// Maximum amount of time to wait for an Instance Refresh to be Cancelled
	instanceRefreshCancelledTimeout = 15 * time.Minute
)

func waitInstanceRefreshCancelled(ctx context.Context, conn *autoscaling.AutoScaling, name, id string, timeout time.Duration) (*autoscaling.InstanceRefresh, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			autoscaling.InstanceRefreshStatusCancelling,
			autoscaling.InstanceRefreshStatusInProgress,
			autoscaling.InstanceRefreshStatusPending,
		},
		Target: []string{
			autoscaling.InstanceRefreshStatusCancelled,
			autoscaling.InstanceRefreshStatusFailed,
			autoscaling.InstanceRefreshStatusSuccessful,
		},
		Refresh: statusInstanceRefresh(ctx, conn, name, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscaling.InstanceRefresh); ok {
		return output, err
	}

	return nil, err
}

func waitWarmPoolDeleted(ctx context.Context, conn *autoscaling.AutoScaling, name string, timeout time.Duration) (*autoscaling.WarmPoolConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{autoscaling.WarmPoolStatusPendingDelete},
		Target:  []string{},
		Refresh: statusWarmPool(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscaling.WarmPoolConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitWarmPoolDrained(ctx context.Context, conn *autoscaling.AutoScaling, name string, timeout time.Duration) (*autoscaling.DescribeWarmPoolOutput, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusWarmPoolInstanceCount(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*autoscaling.DescribeWarmPoolOutput); ok {
		return output, err
	}

	return nil, err
}

func expandInstancesDistribution(tfMap map[string]interface{}) *autoscaling.InstancesDistribution {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.InstancesDistribution{}

	if v, ok := tfMap["on_demand_allocation_strategy"].(string); ok && v != "" {
		apiObject.OnDemandAllocationStrategy = aws.String(v)
	}

	if v, ok := tfMap["on_demand_base_capacity"].(int); ok {
		apiObject.OnDemandBaseCapacity = aws.Int64(int64(v))
	}

	if v, ok := tfMap["on_demand_percentage_above_base_capacity"].(int); ok {
		apiObject.OnDemandPercentageAboveBaseCapacity = aws.Int64(int64(v))
	}

	if v, ok := tfMap["spot_allocation_strategy"].(string); ok && v != "" {
		apiObject.SpotAllocationStrategy = aws.String(v)
	}

	if v, ok := tfMap["spot_instance_pools"].(int); ok && v != 0 {
		apiObject.SpotInstancePools = aws.Int64(int64(v))
	}

	if v, ok := tfMap["spot_max_price"].(string); ok {
		apiObject.SpotMaxPrice = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplate(tfMap map[string]interface{}) *autoscaling.LaunchTemplate {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.LaunchTemplate{}

	if v, ok := tfMap["launch_template_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplateSpecification = expandLaunchTemplateSpecificationForMixedInstancesPolicy(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["override"].([]interface{}); ok && len(v) > 0 {
		apiObject.Overrides = expandLaunchTemplateOverrideses(v)
	}

	return apiObject
}

func expandLaunchTemplateOverrides(tfMap map[string]interface{}) *autoscaling.LaunchTemplateOverrides {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.LaunchTemplateOverrides{}

	if v, ok := tfMap["instance_requirements"].([]interface{}); ok && len(v) > 0 {
		apiObject.InstanceRequirements = expandInstanceRequirements(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["launch_template_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplateSpecification = expandLaunchTemplateSpecificationForMixedInstancesPolicy(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["instance_type"].(string); ok && v != "" {
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := tfMap["weighted_capacity"].(string); ok && v != "" {
		apiObject.WeightedCapacity = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateOverrideses(tfList []interface{}) []*autoscaling.LaunchTemplateOverrides {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*autoscaling.LaunchTemplateOverrides

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateOverrides(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandInstanceRequirements(tfMap map[string]interface{}) *autoscaling.InstanceRequirements {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.InstanceRequirements{}

	if v, ok := tfMap["accelerator_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.AcceleratorCount = expandAcceleratorCountRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["accelerator_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorManufacturers = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["accelerator_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorNames = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["accelerator_total_memory_mib"].([]interface{}); ok && len(v) > 0 {
		apiObject.AcceleratorTotalMemoryMiB = expandAcceleratorTotalMemoryMiBRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["accelerator_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorTypes = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["allowed_instance_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowedInstanceTypes = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["bare_metal"].(string); ok && v != "" {
		apiObject.BareMetal = aws.String(v)
	}

	if v, ok := tfMap["baseline_ebs_bandwidth_mbps"].([]interface{}); ok && len(v) > 0 {
		apiObject.BaselineEbsBandwidthMbps = expandBaselineEBSBandwidthMbpsRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["burstable_performance"].(string); ok && v != "" {
		apiObject.BurstablePerformance = aws.String(v)
	}

	if v, ok := tfMap["cpu_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CpuManufacturers = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["excluded_instance_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ExcludedInstanceTypes = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["instance_generations"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceGenerations = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["local_storage"].(string); ok && v != "" {
		apiObject.LocalStorage = aws.String(v)
	}

	if v, ok := tfMap["local_storage_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.LocalStorageTypes = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["memory_gib_per_vcpu"].([]interface{}); ok && len(v) > 0 {
		apiObject.MemoryGiBPerVCpu = expandMemoryGiBPerVCPURequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["memory_mib"].([]interface{}); ok && len(v) > 0 {
		apiObject.MemoryMiB = expandMemoryMiBRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["network_bandwidth_gbps"].([]interface{}); ok && len(v) > 0 {
		apiObject.NetworkBandwidthGbps = expandNetworkBandwidthGbpsRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["network_interface_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.NetworkInterfaceCount = expandNetworkInterfaceCountRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["on_demand_max_price_percentage_over_lowest_price"].(int); ok && v != 0 {
		apiObject.OnDemandMaxPricePercentageOverLowestPrice = aws.Int64(int64(v))
	}

	if v, ok := tfMap["require_hibernate_support"].(bool); ok && v {
		apiObject.RequireHibernateSupport = aws.Bool(v)
	}

	if v, ok := tfMap["spot_max_price_percentage_over_lowest_price"].(int); ok && v != 0 {
		apiObject.SpotMaxPricePercentageOverLowestPrice = aws.Int64(int64(v))
	}

	if v, ok := tfMap["total_local_storage_gb"].([]interface{}); ok && len(v) > 0 {
		apiObject.TotalLocalStorageGB = expandTotalLocalStorageGBRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["vcpu_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.VCpuCount = expandVCPUCountRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAcceleratorCountRequest(tfMap map[string]interface{}) *autoscaling.AcceleratorCountRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.AcceleratorCountRequest{}

	var min int
	if v, ok := tfMap["min"].(int); ok {
		min = v
		apiObject.Min = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max"].(int); ok && v >= min {
		apiObject.Max = aws.Int64(int64(v))
	}

	return apiObject
}

func expandAcceleratorTotalMemoryMiBRequest(tfMap map[string]interface{}) *autoscaling.AcceleratorTotalMemoryMiBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.AcceleratorTotalMemoryMiBRequest{}

	var min int
	if v, ok := tfMap["min"].(int); ok {
		min = v
		apiObject.Min = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max"].(int); ok && v >= min {
		apiObject.Max = aws.Int64(int64(v))
	}

	return apiObject
}

func expandBaselineEBSBandwidthMbpsRequest(tfMap map[string]interface{}) *autoscaling.BaselineEbsBandwidthMbpsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.BaselineEbsBandwidthMbpsRequest{}

	var min int
	if v, ok := tfMap["min"].(int); ok {
		min = v
		apiObject.Min = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max"].(int); ok && v >= min {
		apiObject.Max = aws.Int64(int64(v))
	}

	return apiObject
}

func expandMemoryGiBPerVCPURequest(tfMap map[string]interface{}) *autoscaling.MemoryGiBPerVCpuRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.MemoryGiBPerVCpuRequest{}

	var min float64
	if v, ok := tfMap["min"].(float64); ok {
		min = v
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap["max"].(float64); ok && v >= min {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandMemoryMiBRequest(tfMap map[string]interface{}) *autoscaling.MemoryMiBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.MemoryMiBRequest{}

	var min int
	if v, ok := tfMap["min"].(int); ok {
		min = v
		apiObject.Min = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max"].(int); ok && v >= min {
		apiObject.Max = aws.Int64(int64(v))
	}

	return apiObject
}

func expandNetworkBandwidthGbpsRequest(tfMap map[string]interface{}) *autoscaling.NetworkBandwidthGbpsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.NetworkBandwidthGbpsRequest{}

	var min float64
	if v, ok := tfMap["min"].(float64); ok {
		min = v
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap["max"].(float64); ok && v >= min {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandNetworkInterfaceCountRequest(tfMap map[string]interface{}) *autoscaling.NetworkInterfaceCountRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.NetworkInterfaceCountRequest{}

	var min int
	if v, ok := tfMap["min"].(int); ok {
		min = v
		apiObject.Min = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max"].(int); ok && v >= min {
		apiObject.Max = aws.Int64(int64(v))
	}

	return apiObject
}

func expandTotalLocalStorageGBRequest(tfMap map[string]interface{}) *autoscaling.TotalLocalStorageGBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.TotalLocalStorageGBRequest{}

	var min float64
	if v, ok := tfMap["min"].(float64); ok {
		min = v
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap["max"].(float64); ok && v >= min {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandVCPUCountRequest(tfMap map[string]interface{}) *autoscaling.VCpuCountRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.VCpuCountRequest{}

	min := 0
	if v, ok := tfMap["min"].(int); ok {
		min = v
		apiObject.Min = aws.Int64(int64(v))
	}

	if v, ok := tfMap["max"].(int); ok && v >= min {
		apiObject.Max = aws.Int64(int64(v))
	}

	return apiObject
}

func expandLaunchTemplateSpecificationForMixedInstancesPolicy(tfMap map[string]interface{}) *autoscaling.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.LaunchTemplateSpecification{}

	// API returns both ID and name, which Terraform saves to state. Next update returns:
	// ValidationError: Valid requests must contain either launchTemplateId or LaunchTemplateName
	// Prefer the ID if we have both.
	if v, ok := tfMap["launch_template_id"]; ok && v != "" && v != launchTemplateIDUnknown {
		apiObject.LaunchTemplateId = aws.String(v.(string))
	} else if v, ok := tfMap["launch_template_name"]; ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v.(string))
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateSpecification(tfMap map[string]interface{}) *autoscaling.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.LaunchTemplateSpecification{}

	// DescribeAutoScalingGroups returns both name and id but LaunchTemplateSpecification
	// allows only one of them to be set.
	if v, ok := tfMap["id"]; ok && v != "" && v != launchTemplateIDUnknown {
		apiObject.LaunchTemplateId = aws.String(v.(string))
	} else if v, ok := tfMap["name"]; ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v.(string))
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandMixedInstancesPolicy(tfMap map[string]interface{}) *autoscaling.MixedInstancesPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.MixedInstancesPolicy{}

	if v, ok := tfMap["instances_distribution"].([]interface{}); ok && len(v) > 0 {
		apiObject.InstancesDistribution = expandInstancesDistribution(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["launch_template"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplate = expandLaunchTemplate(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPutLifecycleHookInput(name string, tfMap map[string]interface{}) *autoscaling.PutLifecycleHookInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.PutLifecycleHookInput{
		AutoScalingGroupName: aws.String(name),
	}

	if v, ok := tfMap["default_result"].(string); ok && v != "" {
		apiObject.DefaultResult = aws.String(v)
	}

	if v, ok := tfMap["heartbeat_timeout"].(int); ok && v != 0 {
		apiObject.HeartbeatTimeout = aws.Int64(int64(v))
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.LifecycleHookName = aws.String(v)
	}

	if v, ok := tfMap["lifecycle_transition"].(string); ok && v != "" {
		apiObject.LifecycleTransition = aws.String(v)
	}

	if v, ok := tfMap["notification_metadata"].(string); ok && v != "" {
		apiObject.NotificationMetadata = aws.String(v)
	}

	if v, ok := tfMap["notification_target_arn"].(string); ok && v != "" {
		apiObject.NotificationTargetARN = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleARN = aws.String(v)
	}

	return apiObject
}

func expandPutLifecycleHookInputs(name string, tfList []interface{}) []*autoscaling.PutLifecycleHookInput {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*autoscaling.PutLifecycleHookInput

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPutLifecycleHookInput(name, tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandPutWarmPoolInput(name string, tfMap map[string]interface{}) *autoscaling.PutWarmPoolInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.PutWarmPoolInput{
		AutoScalingGroupName: aws.String(name),
	}

	if v, ok := tfMap["instance_reuse_policy"].([]interface{}); ok && len(v) > 0 {
		apiObject.InstanceReusePolicy = expandInstanceReusePolicy(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["max_group_prepared_capacity"].(int); ok && v != 0 {
		apiObject.MaxGroupPreparedCapacity = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min_size"].(int); ok && v != 0 {
		apiObject.MinSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["pool_state"].(string); ok && v != "" {
		apiObject.PoolState = aws.String(v)
	}

	return apiObject
}

func expandInstanceReusePolicy(tfMap map[string]interface{}) *autoscaling.InstanceReusePolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.InstanceReusePolicy{}

	if v, ok := tfMap["reuse_on_scale_in"].(bool); ok {
		apiObject.ReuseOnScaleIn = aws.Bool(v)
	}

	return apiObject
}

func expandStartInstanceRefreshInput(name string, tfMap map[string]interface{}, launchTemplate *autoscaling.LaunchTemplateSpecification, mixedInstancesPolicy *autoscaling.MixedInstancesPolicy) *autoscaling.StartInstanceRefreshInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.StartInstanceRefreshInput{
		AutoScalingGroupName: aws.String(name),
	}

	if v, ok := tfMap["preferences"].([]interface{}); ok && len(v) > 0 {
		apiObject.Preferences = expandRefreshPreferences(v[0].(map[string]interface{}))

		// "The AutoRollback parameter cannot be set to true when the DesiredConfiguration parameter is empty".
		if aws.BoolValue(apiObject.Preferences.AutoRollback) {
			apiObject.DesiredConfiguration = &autoscaling.DesiredConfiguration{
				LaunchTemplate:       launchTemplate,
				MixedInstancesPolicy: mixedInstancesPolicy,
			}
		}
	}

	if v, ok := tfMap["strategy"].(string); ok && v != "" {
		apiObject.Strategy = aws.String(v)
	}

	return apiObject
}

func expandRefreshPreferences(tfMap map[string]interface{}) *autoscaling.RefreshPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.RefreshPreferences{}

	if v, ok := tfMap["auto_rollback"].(bool); ok {
		apiObject.AutoRollback = aws.Bool(v)
	}

	if v, ok := tfMap["checkpoint_delay"].(string); ok {
		if v, null, _ := nullable.Int(v).Value(); !null {
			apiObject.CheckpointDelay = aws.Int64(v)
		}
	}

	if v, ok := tfMap["checkpoint_percentages"].([]interface{}); ok && len(v) > 0 {
		apiObject.CheckpointPercentages = flex.ExpandInt64List(v)
	}

	if v, ok := tfMap["instance_warmup"].(string); ok {
		if v, null, _ := nullable.Int(v).Value(); !null {
			apiObject.InstanceWarmup = aws.Int64(v)
		}
	}

	if v, ok := tfMap["min_healthy_percentage"].(int); ok {
		apiObject.MinHealthyPercentage = aws.Int64(int64(v))
	}

	if v, ok := tfMap["scale_in_protected_instances"].(string); ok {
		apiObject.ScaleInProtectedInstances = aws.String(v)
	}

	if v, ok := tfMap["skip_matching"].(bool); ok {
		apiObject.SkipMatching = aws.Bool(v)
	}

	if v, ok := tfMap["standby_instances"].(string); ok {
		apiObject.StandbyInstances = aws.String(v)
	}

	return apiObject
}

func expandVPCZoneIdentifiers(tfList []interface{}) *string {
	vpcZoneIDs := make([]string, len(tfList))

	for i, v := range tfList {
		vpcZoneIDs[i] = v.(string)
	}

	return aws.String(strings.Join(vpcZoneIDs, ","))
}

func expandTrafficSourceIdentifier(tfMap map[string]interface{}) *autoscaling.TrafficSourceIdentifier {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.TrafficSourceIdentifier{}

	if v, ok := tfMap["identifier"].(string); ok && v != "" {
		apiObject.Identifier = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandTrafficSourceIdentifiers(tfList []interface{}) []*autoscaling.TrafficSourceIdentifier {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*autoscaling.TrafficSourceIdentifier

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandTrafficSourceIdentifier(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEnabledMetrics(apiObjects []*autoscaling.EnabledMetric) []string {
	var tfList []string

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if v := apiObject.Metric; v != nil {
			tfList = append(tfList, aws.StringValue(v))
		}
	}

	return tfList
}

func flattenLaunchTemplateSpecification(apiObject *autoscaling.LaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["id"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap["version"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenMixedInstancesPolicy(apiObject *autoscaling.MixedInstancesPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.InstancesDistribution; v != nil {
		tfMap["instances_distribution"] = []interface{}{flattenInstancesDistribution(v)}
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap["launch_template"] = []interface{}{flattenLaunchTemplate(v)}
	}

	return tfMap
}

func flattenInstancesDistribution(apiObject *autoscaling.InstancesDistribution) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OnDemandAllocationStrategy; v != nil {
		tfMap["on_demand_allocation_strategy"] = aws.StringValue(v)
	}

	if v := apiObject.OnDemandBaseCapacity; v != nil {
		tfMap["on_demand_base_capacity"] = aws.Int64Value(v)
	}

	if v := apiObject.OnDemandPercentageAboveBaseCapacity; v != nil {
		tfMap["on_demand_percentage_above_base_capacity"] = aws.Int64Value(v)
	}

	if v := apiObject.SpotAllocationStrategy; v != nil {
		tfMap["spot_allocation_strategy"] = aws.StringValue(v)
	}

	if v := apiObject.SpotInstancePools; v != nil {
		tfMap["spot_instance_pools"] = aws.Int64Value(v)
	}

	if v := apiObject.SpotMaxPrice; v != nil {
		tfMap["spot_max_price"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplate(apiObject *autoscaling.LaunchTemplate) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateSpecification; v != nil {
		tfMap["launch_template_specification"] = []interface{}{flattenLaunchTemplateSpecificationForMixedInstancesPolicy(v)}
	}

	if v := apiObject.Overrides; v != nil {
		tfMap["override"] = flattenLaunchTemplateOverrideses(v)
	}

	return tfMap
}

func flattenLaunchTemplateSpecificationForMixedInstancesPolicy(apiObject *autoscaling.LaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.StringValue(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap["version"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateOverrides(apiObject *autoscaling.LaunchTemplateOverrides) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.InstanceRequirements; v != nil {
		tfMap["instance_requirements"] = []interface{}{flattenInstanceRequirements(v)}
	}

	if v := apiObject.InstanceType; v != nil {
		tfMap["instance_type"] = aws.StringValue(v)
	}

	if v := apiObject.LaunchTemplateSpecification; v != nil {
		tfMap["launch_template_specification"] = []interface{}{flattenLaunchTemplateSpecificationForMixedInstancesPolicy(v)}
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateOverrideses(apiObjects []*autoscaling.LaunchTemplateOverrides) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchTemplateOverrides(apiObject))
	}

	return tfList
}

func flattenInstanceRequirements(apiObject *autoscaling.InstanceRequirements) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AcceleratorCount; v != nil {
		tfMap["accelerator_count"] = []interface{}{flattenAcceleratorCount(v)}
	}

	if v := apiObject.AcceleratorManufacturers; v != nil {
		tfMap["accelerator_manufacturers"] = aws.StringValueSlice(v)
	}

	if v := apiObject.AcceleratorNames; v != nil {
		tfMap["accelerator_names"] = aws.StringValueSlice(v)
	}

	if v := apiObject.AcceleratorTotalMemoryMiB; v != nil {
		tfMap["accelerator_total_memory_mib"] = []interface{}{flattenAcceleratorTotalMemoryMiB(v)}
	}

	if v := apiObject.AcceleratorTypes; v != nil {
		tfMap["accelerator_types"] = aws.StringValueSlice(v)
	}

	if v := apiObject.AllowedInstanceTypes; v != nil {
		tfMap["allowed_instance_types"] = aws.StringValueSlice(v)
	}

	if v := apiObject.BareMetal; v != nil {
		tfMap["bare_metal"] = aws.StringValue(v)
	}

	if v := apiObject.BaselineEbsBandwidthMbps; v != nil {
		tfMap["baseline_ebs_bandwidth_mbps"] = []interface{}{flattenBaselineEBSBandwidthMbps(v)}
	}

	if v := apiObject.BurstablePerformance; v != nil {
		tfMap["burstable_performance"] = aws.StringValue(v)
	}

	if v := apiObject.CpuManufacturers; v != nil {
		tfMap["cpu_manufacturers"] = aws.StringValueSlice(v)
	}

	if v := apiObject.ExcludedInstanceTypes; v != nil {
		tfMap["excluded_instance_types"] = aws.StringValueSlice(v)
	}

	if v := apiObject.InstanceGenerations; v != nil {
		tfMap["instance_generations"] = aws.StringValueSlice(v)
	}

	if v := apiObject.LocalStorage; v != nil {
		tfMap["local_storage"] = aws.StringValue(v)
	}

	if v := apiObject.LocalStorageTypes; v != nil {
		tfMap["local_storage_types"] = aws.StringValueSlice(v)
	}

	if v := apiObject.MemoryGiBPerVCpu; v != nil {
		tfMap["memory_gib_per_vcpu"] = []interface{}{flattenMemoryGiBPerVCPU(v)}
	}

	if v := apiObject.MemoryMiB; v != nil {
		tfMap["memory_mib"] = []interface{}{flattenMemoryMiB(v)}
	}

	if v := apiObject.NetworkBandwidthGbps; v != nil {
		tfMap["network_bandwidth_gbps"] = []interface{}{flattenNetworkBandwidthGbps(v)}
	}

	if v := apiObject.NetworkInterfaceCount; v != nil {
		tfMap["network_interface_count"] = []interface{}{flattenNetworkInterfaceCount(v)}
	}

	if v := apiObject.OnDemandMaxPricePercentageOverLowestPrice; v != nil {
		tfMap["on_demand_max_price_percentage_over_lowest_price"] = aws.Int64Value(v)
	}

	if v := apiObject.RequireHibernateSupport; v != nil {
		tfMap["require_hibernate_support"] = aws.BoolValue(v)
	}

	if v := apiObject.SpotMaxPricePercentageOverLowestPrice; v != nil {
		tfMap["spot_max_price_percentage_over_lowest_price"] = aws.Int64Value(v)
	}

	if v := apiObject.TotalLocalStorageGB; v != nil {
		tfMap["total_local_storage_gb"] = []interface{}{flattentTotalLocalStorageGB(v)}
	}

	if v := apiObject.VCpuCount; v != nil {
		tfMap["vcpu_count"] = []interface{}{flattenVCPUCount(v)}
	}

	return tfMap
}

func flattenAcceleratorCount(apiObject *autoscaling.AcceleratorCountRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Int64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenAcceleratorTotalMemoryMiB(apiObject *autoscaling.AcceleratorTotalMemoryMiBRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Int64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenBaselineEBSBandwidthMbps(apiObject *autoscaling.BaselineEbsBandwidthMbpsRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Int64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenMemoryGiBPerVCPU(apiObject *autoscaling.MemoryGiBPerVCpuRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Float64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Float64Value(v)
	}

	return tfMap
}

func flattenMemoryMiB(apiObject *autoscaling.MemoryMiBRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Int64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenNetworkBandwidthGbps(apiObject *autoscaling.NetworkBandwidthGbpsRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Float64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Float64Value(v)
	}

	return tfMap
}

func flattenNetworkInterfaceCount(apiObject *autoscaling.NetworkInterfaceCountRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Int64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattentTotalLocalStorageGB(apiObject *autoscaling.TotalLocalStorageGBRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Float64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Float64Value(v)
	}

	return tfMap
}

func flattenTrafficSourceIdentifier(apiObject *autoscaling.TrafficSourceIdentifier) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Identifier; v != nil {
		tfMap["identifier"] = aws.StringValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenTrafficSourceIdentifiers(apiObjects []*autoscaling.TrafficSourceIdentifier) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenTrafficSourceIdentifier(apiObject))
	}

	return tfList
}

func flattenVCPUCount(apiObject *autoscaling.VCpuCountRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap["max"] = aws.Int64Value(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap["min"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenSuspendedProcesses(apiObjects []*autoscaling.SuspendedProcess) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if v := apiObject.ProcessName; v != nil {
			tfList = append(tfList, aws.StringValue(v))
		}
	}

	return tfList
}

func flattenWarmPoolConfiguration(apiObject *autoscaling.WarmPoolConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.InstanceReusePolicy; v != nil {
		tfMap["instance_reuse_policy"] = []interface{}{flattenWarmPoolInstanceReusePolicy(v)}
	}

	if v := apiObject.MaxGroupPreparedCapacity; v != nil {
		tfMap["max_group_prepared_capacity"] = aws.Int64Value(v)
	} else {
		tfMap["max_group_prepared_capacity"] = int64(DefaultWarmPoolMaxGroupPreparedCapacity)
	}

	if v := apiObject.MinSize; v != nil {
		tfMap["min_size"] = aws.Int64Value(v)
	}

	if v := apiObject.PoolState; v != nil {
		tfMap["pool_state"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenWarmPoolInstanceReusePolicy(apiObject *autoscaling.InstanceReusePolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ReuseOnScaleIn; v != nil {
		tfMap["reuse_on_scale_in"] = aws.BoolValue(v)
	}

	return tfMap
}

func cancelInstanceRefresh(ctx context.Context, conn *autoscaling.AutoScaling, name string) error {
	input := &autoscaling.CancelInstanceRefreshInput{
		AutoScalingGroupName: aws.String(name),
	}

	output, err := conn.CancelInstanceRefreshWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, autoscaling.ErrCodeActiveInstanceRefreshNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cancelling Auto Scaling Group (%s) instance refresh: %w", name, err)
	}

	_, err = waitInstanceRefreshCancelled(ctx, conn, name, aws.StringValue(output.InstanceRefreshId), instanceRefreshCancelledTimeout)

	if err != nil {
		return fmt.Errorf("waiting for Auto Scaling Group (%s) instance refresh cancel: %w", name, err)
	}

	return nil
}

func startInstanceRefresh(ctx context.Context, conn *autoscaling.AutoScaling, input *autoscaling.StartInstanceRefreshInput) error {
	name := aws.StringValue(input.AutoScalingGroupName)

	_, err := tfresource.RetryWhen(ctx, instanceRefreshStartedTimeout,
		func() (interface{}, error) {
			return conn.StartInstanceRefreshWithContext(ctx, input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, autoscaling.ErrCodeInstanceRefreshInProgressFault) {
				if err := cancelInstanceRefresh(ctx, conn, name); err != nil {
					return false, err
				}

				return true, err
			}

			return false, err
		})

	if err != nil {
		return fmt.Errorf("starting Auto Scaling Group (%s) instance refresh: %w", name, err)
	}

	return nil
}

func validateGroupInstanceRefreshTriggerFields(i interface{}, path cty.Path) diag.Diagnostics {
	v, ok := i.(string)
	if !ok {
		return diag.Errorf("expected type to be string")
	}

	if v == "launch_configuration" || v == "launch_template" || v == "mixed_instances_policy" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("'%s' always triggers an instance refresh and can be removed", v),
			},
		}
	}

	schema := ResourceGroup().SchemaMap()
	for attr, attrSchema := range schema {
		if v == attr {
			if attrSchema.Computed && !attrSchema.Optional {
				return diag.Errorf("'%s' is a read-only parameter and cannot be used to trigger an instance refresh", v)
			}
			return nil
		}
	}

	return diag.Errorf("'%s' is not a recognized parameter name for aws_autoscaling_group", v)
}
