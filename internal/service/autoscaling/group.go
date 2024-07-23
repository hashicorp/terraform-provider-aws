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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	elasticloadbalancingv2types "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_autoscaling_group", name="Group")
func resourceGroup() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[desiredCapacityType](),
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
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[lifecycleHookDefaultResult](),
						},
						"heartbeat_timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(30, 7200),
						},
						"lifecycle_transition": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[lifecycleHookLifecycleTransition](),
						},
						names.AttrName: {
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
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"instance_maintenance_policy": {
				Type:             schema.TypeList,
				MaxItems:         1,
				Optional:         true,
				DiffSuppressFunc: instanceMaintenancePolicyDiffSupress,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_healthy_percentage": {
							Type:     schema.TypeInt,
							Required: true,
							ValidateFunc: validation.Any(
								validation.IntBetween(100, 200),
								validation.IntBetween(-1, -1),
							),
							// When value is -1, instance maintenance policy is removed, state file will not contain any value.
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return old == "" && new == "-1"
							},
						},
						"min_healthy_percentage": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(-1, 100),
							// When value is -1, instance maintenance policy is removed, state file will not contain any value.
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return old == "" && new == "-1"
							},
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
									"alarm_specification": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"alarms": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
											},
										},
									},
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
									"max_healthy_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      100,
										ValidateFunc: validation.IntBetween(100, 200),
									},
									"min_healthy_percentage": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      90,
										ValidateFunc: validation.IntBetween(0, 100),
									},
									"scale_in_protected_instances": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          awstypes.ScaleInProtectedInstancesIgnore,
										ValidateDiagFunc: enum.Validate[awstypes.ScaleInProtectedInstances](),
									},
									"skip_matching": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"standby_instances": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          awstypes.StandbyInstancesIgnore,
										ValidateDiagFunc: enum.Validate[awstypes.StandbyInstances](),
									},
								},
							},
						},
						"strategy": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.RefreshStrategy](),
						},
						names.AttrTriggers: {
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
				ExactlyOneOf: []string{"launch_configuration", names.AttrLaunchTemplate, "mixed_instances_policy"},
			},
			names.AttrLaunchTemplate: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrID: {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  verify.ValidLaunchTemplateID,
							ConflictsWith: []string{"launch_template.0.name"},
						},
						names.AttrName: {
							Type:          schema.TypeString,
							Optional:      true,
							Computed:      true,
							ValidateFunc:  verify.ValidLaunchTemplateName,
							ConflictsWith: []string{"launch_template.0.id"},
						},
						names.AttrVersion: {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
				ExactlyOneOf: []string{"launch_configuration", names.AttrLaunchTemplate, "mixed_instances_policy"},
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
													Computed: true,
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
																		names.AttrMax: {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(0),
																		},
																		names.AttrMin: {
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
																	Type:             schema.TypeString,
																	ValidateDiagFunc: enum.Validate[awstypes.AcceleratorManufacturer](),
																},
															},
															"accelerator_names": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type:             schema.TypeString,
																	ValidateDiagFunc: enum.Validate[awstypes.AcceleratorName](),
																},
															},
															"accelerator_total_memory_mib": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		names.AttrMin: {
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
																	Type:             schema.TypeString,
																	ValidateDiagFunc: enum.Validate[awstypes.AcceleratorType](),
																},
															},
															"allowed_instance_types": {
																Type:     schema.TypeSet,
																Optional: true,
																MaxItems: 400,
																Elem:     &schema.Schema{Type: schema.TypeString},
															},
															"bare_metal": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[awstypes.BareMetal](),
															},
															"baseline_ebs_bandwidth_mbps": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		names.AttrMin: {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																	},
																},
															},
															"burstable_performance": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[awstypes.BurstablePerformance](),
															},
															"cpu_manufacturers": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type:             schema.TypeString,
																	ValidateDiagFunc: enum.Validate[awstypes.CpuManufacturer](),
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
																	Type:             schema.TypeString,
																	ValidateDiagFunc: enum.Validate[awstypes.InstanceGeneration](),
																},
															},
															"local_storage": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[awstypes.LocalStorage](),
															},
															"local_storage_types": {
																Type:     schema.TypeSet,
																Optional: true,
																Elem: &schema.Schema{
																	Type:             schema.TypeString,
																	ValidateDiagFunc: enum.Validate[awstypes.LocalStorageType](),
																},
															},
															"max_spot_price_as_percentage_of_optimal_on_demand_price": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"memory_gib_per_vcpu": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		names.AttrMax: {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
																		},
																		names.AttrMin: {
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
																		names.AttrMax: {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		names.AttrMin: {
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
																		names.AttrMax: {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
																		},
																		names.AttrMin: {
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
																		names.AttrMax: {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		names.AttrMin: {
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
																		names.AttrMax: {
																			Type:         schema.TypeFloat,
																			Optional:     true,
																			ValidateFunc: verify.FloatGreaterThan(0.0),
																		},
																		names.AttrMin: {
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
																		names.AttrMax: {
																			Type:         schema.TypeInt,
																			Optional:     true,
																			ValidateFunc: validation.IntAtLeast(1),
																		},
																		names.AttrMin: {
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
																Computed: true,
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
				ExactlyOneOf: []string{"launch_configuration", names.AttrLaunchTemplate, "mixed_instances_policy"},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringLenBetween(0, 255),
				ConflictsWith: []string{names.AttrNamePrefix},
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  validation.StringLenBetween(0, 255-id.UniqueIDSuffixLength),
				ConflictsWith: []string{names.AttrName},
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
						names.AttrIdentifier: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 2048),
						},
						names.AttrType: {
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
				ConflictsWith: []string{names.AttrAvailabilityZones},
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
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      defaultWarmPoolMaxGroupPreparedCapacity,
							ValidateFunc: validation.IntAtLeast(defaultWarmPoolMaxGroupPreparedCapacity),
						},
						"min_size": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntAtLeast(0),
						},
						"pool_state": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.WarmPoolStateStopped,
							ValidateDiagFunc: enum.Validate[awstypes.WarmPoolState](),
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
			launchTemplateCustomDiff(names.AttrLaunchTemplate, "launch_template.0.name"),
			launchTemplateCustomDiff("mixed_instances_policy", "mixed_instances_policy.0.launch_template.0.launch_template_specification.0.launch_template_name"),
			launchTemplateCustomDiff("mixed_instances_policy", "mixed_instances_policy.0.launch_template.0.override"),
		),
	}
}

func instanceMaintenancePolicyDiffSupress(k, old, new string, d *schema.ResourceData) bool {
	o, n := d.GetChange("instance_maintenance_policy")
	oList := o.([]interface{})
	nList := n.([]interface{})

	if len(oList) == 0 && len(nList) != 0 {
		tfMap := nList[0].(map[string]interface{})
		if int64(tfMap["min_healthy_percentage"].(int)) == -1 || int64(tfMap["max_healthy_percentage"].(int)) == -1 {
			return true
		}
	}
	return false
}

func launchTemplateCustomDiff(baseAttribute, subAttribute string) schema.CustomizeDiffFunc {
	return func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
		if diff.HasChange(subAttribute) {
			n := diff.Get(baseAttribute)
			ba, ok := n.([]interface{})
			if !ok {
				return nil
			}

			if baseAttribute == names.AttrLaunchTemplate {
				launchTemplate := ba[0].(map[string]interface{})
				launchTemplate[names.AttrID] = launchTemplateIDUnknown

				if err := diff.SetNew(baseAttribute, ba); err != nil {
					return err
				}
			}

			if baseAttribute == "mixed_instances_policy" && !strings.Contains(subAttribute, "override") {
				launchTemplate := ba[0].(map[string]interface{})[names.AttrLaunchTemplate].([]interface{})[0].(map[string]interface{})["launch_template_specification"].([]interface{})[0]
				launchTemplateSpecification := launchTemplate.(map[string]interface{})
				launchTemplateSpecification["launch_template_id"] = launchTemplateIDUnknown

				if err := diff.SetNew(baseAttribute, ba); err != nil {
					return err
				}
			}

			if baseAttribute == "mixed_instances_policy" && strings.Contains(subAttribute, "override") {
				launchTemplate := ba[0].(map[string]interface{})[names.AttrLaunchTemplate].([]interface{})[0].(map[string]interface{})["override"].([]interface{})

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
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	startTime := time.Now()

	asgName := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	inputCASG := &autoscaling.CreateAutoScalingGroupInput{
		AutoScalingGroupName:             aws.String(asgName),
		NewInstancesProtectedFromScaleIn: aws.Bool(d.Get("protect_from_scale_in").(bool)),
	}
	inputUASG := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(asgName),
	}

	initialLifecycleHooks := d.Get("initial_lifecycle_hook").(*schema.Set).List()
	twoPhases := len(initialLifecycleHooks) > 0

	maxSize := d.Get("max_size").(int)
	minSize := d.Get("min_size").(int)
	desiredCapacity := d.Get("desired_capacity").(int)

	if twoPhases {
		inputCASG.MaxSize = aws.Int32(0)
		inputCASG.MinSize = aws.Int32(0)

		inputUASG.MaxSize = aws.Int32(int32(maxSize))
		inputUASG.MinSize = aws.Int32(int32(minSize))

		if desiredCapacity > 0 {
			inputUASG.DesiredCapacity = aws.Int32(int32(desiredCapacity))
		}

		if v, ok := d.GetOk("desired_capacity_type"); ok {
			inputUASG.DesiredCapacityType = aws.String(v.(string))
		}
	} else {
		inputCASG.MaxSize = aws.Int32(int32(maxSize))
		inputCASG.MinSize = aws.Int32(int32(minSize))

		if desiredCapacity > 0 {
			inputCASG.DesiredCapacity = aws.Int32(int32(desiredCapacity))
		}

		if v, ok := d.GetOk("desired_capacity_type"); ok {
			inputCASG.DesiredCapacityType = aws.String(v.(string))
		}
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZones); ok && v.(*schema.Set).Len() > 0 {
		inputCASG.AvailabilityZones = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("capacity_rebalance"); ok {
		inputCASG.CapacityRebalance = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("context"); ok {
		inputCASG.Context = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_cooldown"); ok {
		inputCASG.DefaultCooldown = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("default_instance_warmup"); ok {
		inputCASG.DefaultInstanceWarmup = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("health_check_type"); ok {
		inputCASG.HealthCheckType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("health_check_grace_period"); ok {
		inputCASG.HealthCheckGracePeriod = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("instance_maintenance_policy"); ok {
		inputCASG.InstanceMaintenancePolicy = expandInstanceMaintenancePolicy(v.([]interface{}))
	}

	if v, ok := d.GetOk("launch_configuration"); ok {
		inputCASG.LaunchConfigurationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrLaunchTemplate); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		inputCASG.LaunchTemplate = expandLaunchTemplateSpecification(v.([]interface{})[0].(map[string]interface{}), false)
	}

	if v, ok := d.GetOk("load_balancers"); ok && v.(*schema.Set).Len() > 0 {
		inputCASG.LoadBalancerNames = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("max_instance_lifetime"); ok {
		inputCASG.MaxInstanceLifetime = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("mixed_instances_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		inputCASG.MixedInstancesPolicy = expandMixedInstancesPolicy(v.([]interface{})[0].(map[string]interface{}), true)
	}

	if v, ok := d.GetOk("placement_group"); ok {
		inputCASG.PlacementGroup = aws.String(v.(string))
	}

	if v, ok := d.GetOk("service_linked_role_arn"); ok {
		inputCASG.ServiceLinkedRoleARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("tag"); ok {
		inputCASG.Tags = Tags(KeyValueTags(ctx, v, asgName, TagResourceTypeGroup).IgnoreAWS())
	}

	if v, ok := d.GetOk("target_group_arns"); ok && len(v.(*schema.Set).List()) > 0 {
		inputCASG.TargetGroupARNs = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("termination_policies"); ok && len(v.([]interface{})) > 0 {
		inputCASG.TerminationPolicies = flex.ExpandStringValueList(v.([]interface{}))
	}

	if v, ok := d.GetOk("traffic_source"); ok && v.(*schema.Set).Len() > 0 {
		inputCASG.TrafficSources = expandTrafficSourceIdentifiers(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("vpc_zone_identifier"); ok && v.(*schema.Set).Len() > 0 {
		inputCASG.VPCZoneIdentifier = expandVPCZoneIdentifiers(v.(*schema.Set).List())
	}

	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateAutoScalingGroup(ctx, inputCASG)
		},
		// ValidationError: You must use a valid fully-formed launch template. Value (tf-acc-test-6643732652421074386) for parameter iamInstanceProfile.name is invalid. Invalid IAM Instance Profile name
		errCodeValidationError, "Invalid IAM Instance Profile")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Group (%s): %s", asgName, err)
	}

	d.SetId(asgName)

	if twoPhases {
		for _, input := range expandPutLifecycleHookInputs(asgName, initialLifecycleHooks) {
			const (
				timeout = 5 * time.Minute
			)
			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, timeout,
				func() (interface{}, error) {
					return conn.PutLifecycleHook(ctx, input)
				},
				errCodeValidationError, "Unable to publish test message to notification target")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Group (%s) Lifecycle Hook: %s", d.Id(), err)
			}
		}

		_, err = conn.UpdateAutoScalingGroup(ctx, inputUASG)

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

			if err := waitGroupCapacitySatisfied(ctx, conn, meta.(*conns.AWSClient).ELBClient(ctx), meta.(*conns.AWSClient).ELBV2Client(ctx), d.Id(), f, startTime, d.Get("ignore_failed_scaling_activities").(bool), timeout); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) capacity satisfied: %s", d.Id(), err)
			}
		}
	}

	if v, ok := d.GetOk("suspended_processes"); ok && v.(*schema.Set).Len() > 0 {
		input := &autoscaling.SuspendProcessesInput{
			AutoScalingGroupName: aws.String(d.Id()),
			ScalingProcesses:     flex.ExpandStringValueSet(v.(*schema.Set)),
		}

		_, err := conn.SuspendProcesses(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "suspending Auto Scaling Group (%s) scaling processes: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("enabled_metrics"); ok && v.(*schema.Set).Len() > 0 {
		input := &autoscaling.EnableMetricsCollectionInput{
			AutoScalingGroupName: aws.String(d.Id()),
			Granularity:          aws.String(d.Get("metrics_granularity").(string)),
			Metrics:              flex.ExpandStringValueSet(v.(*schema.Set)),
		}

		_, err := conn.EnableMetricsCollection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "enabling Auto Scaling Group (%s) metrics collection: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("warm_pool"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		_, err := conn.PutWarmPool(ctx, expandPutWarmPoolInput(d.Id(), v.([]interface{})[0].(map[string]interface{})))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Auto Scaling Warm Pool (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	g, err := findGroupByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Group %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Group (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, g.AutoScalingGroupARN)
	d.Set(names.AttrAvailabilityZones, g.AvailabilityZones)
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
		d.Set("metrics_granularity", defaultEnabledMetricsGranularity)
	}
	d.Set("health_check_grace_period", g.HealthCheckGracePeriod)
	d.Set("health_check_type", g.HealthCheckType)
	if err := d.Set("instance_maintenance_policy", flattenInstanceMaintenancePolicy(g.InstanceMaintenancePolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_maintenance_policy: %s", err)
	}
	d.Set("launch_configuration", g.LaunchConfigurationName)
	if g.LaunchTemplate != nil {
		if err := d.Set(names.AttrLaunchTemplate, []interface{}{flattenLaunchTemplateSpecification(g.LaunchTemplate)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting launch_template: %s", err)
		}
	} else {
		d.Set(names.AttrLaunchTemplate, nil)
	}
	d.Set("load_balancers", g.LoadBalancerNames)
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
	d.Set(names.AttrName, g.AutoScalingGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(g.AutoScalingGroupName)))
	d.Set("placement_group", g.PlacementGroup)
	d.Set("predicted_capacity", g.PredictedCapacity)
	d.Set("protect_from_scale_in", g.NewInstancesProtectedFromScaleIn)
	d.Set("service_linked_role_arn", g.ServiceLinkedRoleARN)
	d.Set("suspended_processes", flattenSuspendedProcesses(g.SuspendedProcesses))
	if err := d.Set("traffic_source", flattenTrafficSourceIdentifiers(g.TrafficSources)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting traffic_source: %s", err)
	}
	d.Set("target_group_arns", g.TargetGroupARNs)
	// If no termination polices are explicitly configured and the upstream state
	// is only using the "Default" policy, clear the state to make it consistent
	// with the default AWS Create API behavior.
	if _, ok := d.GetOk("termination_policies"); !ok && len(g.TerminationPolicies) == 1 && g.TerminationPolicies[0] == defaultTerminationPolicy {
		d.Set("termination_policies", nil)
	} else {
		d.Set("termination_policies", g.TerminationPolicies)
	}
	if len(aws.ToString(g.VPCZoneIdentifier)) > 0 {
		d.Set("vpc_zone_identifier", strings.Split(aws.ToString(g.VPCZoneIdentifier), ","))
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

	if err := d.Set("tag", listOfMap(KeyValueTags(ctx, g.Tags, d.Id(), TagResourceTypeGroup).IgnoreAWS().IgnoreConfig(ignoreTagsConfig))); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tag: %s", err)
	}

	return diags
}

func resourceGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

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

		if d.HasChange(names.AttrAvailabilityZones) {
			if v, ok := d.GetOk(names.AttrAvailabilityZones); ok && v.(*schema.Set).Len() > 0 {
				input.AvailabilityZones = flex.ExpandStringValueSet(v.(*schema.Set))
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
			input.DefaultCooldown = aws.Int32(int32(d.Get("default_cooldown").(int)))
		}

		if d.HasChange("default_instance_warmup") {
			input.DefaultInstanceWarmup = aws.Int32(int32(d.Get("default_instance_warmup").(int)))
		}

		if d.HasChange("desired_capacity") {
			input.DesiredCapacity = aws.Int32(int32(d.Get("desired_capacity").(int)))
			shouldWaitForCapacity = true
		}

		if d.HasChange("desired_capacity_type") {
			input.DesiredCapacityType = aws.String(d.Get("desired_capacity_type").(string))
			shouldWaitForCapacity = true
		}

		if d.HasChange("health_check_grace_period") {
			input.HealthCheckGracePeriod = aws.Int32(int32(d.Get("health_check_grace_period").(int)))
		}

		if d.HasChange("health_check_type") {
			input.HealthCheckGracePeriod = aws.Int32(int32(d.Get("health_check_grace_period").(int)))
			input.HealthCheckType = aws.String(d.Get("health_check_type").(string))
		}

		if d.HasChange("instance_maintenance_policy") {
			input.InstanceMaintenancePolicy = expandInstanceMaintenancePolicy(d.Get("instance_maintenance_policy").([]interface{}))
		}

		if d.HasChange("launch_configuration") {
			if v, ok := d.GetOk("launch_configuration"); ok {
				input.LaunchConfigurationName = aws.String(v.(string))
			}
			shouldRefreshInstances = true
		}

		if d.HasChange(names.AttrLaunchTemplate) {
			if v, ok := d.GetOk(names.AttrLaunchTemplate); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.LaunchTemplate = expandLaunchTemplateSpecification(v.([]interface{})[0].(map[string]interface{}), false)
			}
			shouldRefreshInstances = true
		}

		if d.HasChange("max_instance_lifetime") {
			input.MaxInstanceLifetime = aws.Int32(int32(d.Get("max_instance_lifetime").(int)))
		}

		if d.HasChange("max_size") {
			input.MaxSize = aws.Int32(int32(d.Get("max_size").(int)))
		}

		if d.HasChange("min_size") {
			input.MinSize = aws.Int32(int32(d.Get("min_size").(int)))
			shouldWaitForCapacity = true
		}

		if d.HasChange("mixed_instances_policy") {
			if v, ok := d.GetOk("mixed_instances_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				input.MixedInstancesPolicy = expandMixedInstancesPolicy(v.([]interface{})[0].(map[string]interface{}), true)
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
				input.TerminationPolicies = flex.ExpandStringValueList(v.([]interface{}))
			} else {
				input.TerminationPolicies = []string{defaultTerminationPolicy}
			}
		}

		if d.HasChange("vpc_zone_identifier") {
			input.VPCZoneIdentifier = expandVPCZoneIdentifiers(d.Get("vpc_zone_identifier").(*schema.Set).List())
		}

		_, err := conn.UpdateAutoScalingGroup(ctx, input)

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
		for _, chunk := range tfslices.Chunks(expandTrafficSourceIdentifiers(os.Difference(ns).List()), batchSize) {
			_, err := conn.DetachTrafficSources(ctx, &autoscaling.DetachTrafficSourcesInput{
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

		for _, chunk := range tfslices.Chunks(expandTrafficSourceIdentifiers(ns.Difference(os).List()), batchSize) {
			_, err := conn.AttachTrafficSources(ctx, &autoscaling.AttachTrafficSourcesInput{
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
		for _, chunk := range tfslices.Chunks(flex.ExpandStringValueSet(os.Difference(ns)), batchSize) {
			_, err := conn.DetachLoadBalancers(ctx, &autoscaling.DetachLoadBalancersInput{
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

		for _, chunk := range tfslices.Chunks(flex.ExpandStringValueSet(ns.Difference(os)), batchSize) {
			_, err := conn.AttachLoadBalancers(ctx, &autoscaling.AttachLoadBalancersInput{
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
		for _, chunk := range tfslices.Chunks(flex.ExpandStringValueSet(os.Difference(ns)), batchSize) {
			_, err := conn.DetachLoadBalancerTargetGroups(ctx, &autoscaling.DetachLoadBalancerTargetGroupsInput{
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

		for _, chunk := range tfslices.Chunks(flex.ExpandStringValueSet(ns.Difference(os)), batchSize) {
			_, err := conn.AttachLoadBalancerTargetGroups(ctx, &autoscaling.AttachLoadBalancerTargetGroupsInput{
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
			if v, ok := tfMap[names.AttrTriggers].(*schema.Set); ok && v.Len() > 0 {
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
			var launchTemplate *awstypes.LaunchTemplateSpecification

			if v, ok := d.GetOk(names.AttrLaunchTemplate); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				launchTemplate = expandLaunchTemplateSpecification(v.([]interface{})[0].(map[string]interface{}), false)
			}

			var mixedInstancesPolicy *awstypes.MixedInstancesPolicy

			if v, ok := d.GetOk("mixed_instances_policy"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				mixedInstancesPolicy = expandMixedInstancesPolicy(v.([]interface{})[0].(map[string]interface{}), true)
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
			forceDeleteWarmPool := d.Get(names.AttrForceDelete).(bool) || d.Get("force_delete_warm_pool").(bool)

			if err := deleteWarmPool(ctx, conn, d.Id(), forceDeleteWarmPool, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		} else {
			_, err := conn.PutWarmPool(ctx, expandPutWarmPoolInput(d.Id(), w[0].(map[string]interface{})))

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

				if err := waitGroupCapacitySatisfied(ctx, conn, meta.(*conns.AWSClient).ELBClient(ctx), meta.(*conns.AWSClient).ELBV2Client(ctx), d.Id(), f, startTime, d.Get("ignore_failed_scaling_activities").(bool), timeout); err != nil {
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
				Metrics:              flex.ExpandStringValueSet(disableMetrics),
			}

			_, err := conn.DisableMetricsCollection(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "disabling Auto Scaling Group (%s) metrics collection: %s", d.Id(), err)
			}
		}

		if enableMetrics := ns.Difference(os); enableMetrics.Len() != 0 {
			input := &autoscaling.EnableMetricsCollectionInput{
				AutoScalingGroupName: aws.String(d.Id()),
				Granularity:          aws.String(d.Get("metrics_granularity").(string)),
				Metrics:              flex.ExpandStringValueSet(enableMetrics),
			}

			_, err := conn.EnableMetricsCollection(ctx, input)

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
			input := &autoscaling.ResumeProcessesInput{
				AutoScalingGroupName: aws.String(d.Id()),
				ScalingProcesses:     flex.ExpandStringValueSet(resumeProcesses),
			}

			_, err := conn.ResumeProcesses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "resuming Auto Scaling Group (%s) scaling processes: %s", d.Id(), err)
			}
		}

		if suspendProcesses := ns.Difference(os); suspendProcesses.Len() != 0 {
			input := &autoscaling.SuspendProcessesInput{
				AutoScalingGroupName: aws.String(d.Id()),
				ScalingProcesses:     flex.ExpandStringValueSet(suspendProcesses),
			}

			_, err := conn.SuspendProcesses(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "suspending Auto Scaling Group (%s) scaling processes: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceGroupRead(ctx, d, meta)...)
}

func resourceGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AutoScalingClient(ctx)

	forceDeleteGroup := d.Get(names.AttrForceDelete).(bool)
	forceDeleteWarmPool := forceDeleteGroup || d.Get("force_delete_warm_pool").(bool)

	group, err := findGroupByName(ctx, conn, d.Id())

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
			return conn.DeleteAutoScalingGroup(ctx, &autoscaling.DeleteAutoScalingGroupInput{
				AutoScalingGroupName: aws.String(d.Id()),
				ForceDelete:          aws.Bool(forceDeleteGroup),
			})
		},
		errCodeResourceInUseFault, errCodeScalingActivityInProgressFault)

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Auto Scaling Group (%s): %s", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return findGroupByName(ctx, conn, d.Id())
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Auto Scaling Group (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func drainGroup(ctx context.Context, conn *autoscaling.Client, name string, instances []awstypes.Instance, timeout time.Duration) error {
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(name),
		DesiredCapacity:      aws.Int32(0),
		MinSize:              aws.Int32(0),
		MaxSize:              aws.Int32(0),
	}

	log.Printf("[DEBUG] Draining Auto Scaling Group: %s", name)
	if _, err := conn.UpdateAutoScalingGroup(ctx, input); err != nil {
		return fmt.Errorf("setting Auto Scaling Group (%s) capacity to 0: %w", name, err)
	}

	// Next, ensure that instances are not prevented from scaling in.
	//
	// The ASG's own scale-in protection setting doesn't make a difference here,
	// as it only affects new instances, which won't be launched now that the
	// desired capacity is set to 0. There is also the possibility that this ASG
	// no longer applies scale-in protection to new instances, but there's still
	// old ones that have it.
	//
	// Filter by ProtectedFromScaleIn to avoid unnecessary API calls (#36584)
	var instanceIDs []string
	for _, instance := range instances {
		if aws.ToBool(instance.ProtectedFromScaleIn) {
			instanceIDs = append(instanceIDs, aws.ToString(instance.InstanceId))
		}
	}
	const batchSize = 50 // API limit.
	for _, chunk := range tfslices.Chunks(instanceIDs, batchSize) {
		input := &autoscaling.SetInstanceProtectionInput{
			AutoScalingGroupName: aws.String(name),
			InstanceIds:          chunk,
			ProtectedFromScaleIn: aws.Bool(false),
		}

		_, err := conn.SetInstanceProtection(ctx, input)

		// Ignore ValidationError when instance is already fully terminated
		// and is not a part of Auto Scaling Group anymore.
		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not part of Auto Scaling group") {
			continue
		}

		if err != nil {
			return fmt.Errorf("disabling Auto Scaling Group (%s) scale-in protections: %w", name, err)
		}
	}

	if _, err := waitGroupDrained(ctx, conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Auto Scaling Group (%s) drain: %w", name, err)
	}

	return nil
}

func deleteWarmPool(ctx context.Context, conn *autoscaling.Client, name string, force bool, timeout time.Duration) error {
	if !force {
		if err := drainWarmPool(ctx, conn, name, timeout); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Auto Scaling Warm Pool: %s", name)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout,
		func() (interface{}, error) {
			return conn.DeleteWarmPool(ctx, &autoscaling.DeleteWarmPoolInput{
				AutoScalingGroupName: aws.String(name),
				ForceDelete:          aws.Bool(force),
			})
		},
		errCodeResourceInUseFault, errCodeScalingActivityInProgressFault)

	if tfawserr.ErrMessageContains(err, errCodeValidationError, "No warm pool found") {
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

func drainWarmPool(ctx context.Context, conn *autoscaling.Client, name string, timeout time.Duration) error {
	input := &autoscaling.PutWarmPoolInput{
		AutoScalingGroupName:     aws.String(name),
		MaxGroupPreparedCapacity: aws.Int32(0),
		MinSize:                  aws.Int32(0),
	}

	log.Printf("[DEBUG] Draining Auto Scaling Warm Pool: %s", name)
	if _, err := conn.PutWarmPool(ctx, input); err != nil {
		return fmt.Errorf("setting Auto Scaling Warm Pool (%s) capacity to 0: %w", name, err)
	}

	if _, err := waitWarmPoolDrained(ctx, conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Auto Scaling Warm Pool (%s) drain: %w", name, err)
	}

	return nil
}

func findELBInstanceStates(ctx context.Context, conn *elasticloadbalancing.Client, g *awstypes.AutoScalingGroup) (map[string]map[string]string, error) {
	instanceStates := make(map[string]map[string]string)

	for _, v := range g.LoadBalancerNames {
		lbName := v
		input := &elasticloadbalancing.DescribeInstanceHealthInput{
			LoadBalancerName: aws.String(lbName),
		}

		output, err := conn.DescribeInstanceHealth(ctx, input)

		if err != nil {
			return nil, fmt.Errorf("reading load balancer (%s) instance health: %w", lbName, err)
		}

		instanceStates[lbName] = make(map[string]string)

		for _, v := range output.InstanceStates {
			instanceID := aws.ToString(v.InstanceId)
			if instanceID == "" {
				continue
			}
			state := aws.ToString(v.State)
			if state == "" {
				continue
			}

			instanceStates[lbName][instanceID] = state
		}
	}

	return instanceStates, nil
}

func findELBV2InstanceStates(ctx context.Context, conn *elasticloadbalancingv2.Client, g *awstypes.AutoScalingGroup) (map[string]map[string]string, error) {
	instanceStates := make(map[string]map[string]string)

	for _, v := range g.TargetGroupARNs {
		targetGroupARN := v
		input := &elasticloadbalancingv2.DescribeTargetHealthInput{
			TargetGroupArn: aws.String(targetGroupARN),
		}

		output, err := conn.DescribeTargetHealth(ctx, input)

		if err != nil {
			return nil, fmt.Errorf("reading target group (%s) instance health: %w", targetGroupARN, err)
		}

		instanceStates[targetGroupARN] = make(map[string]string)

		for _, v := range output.TargetHealthDescriptions {
			if v.Target == nil || v.TargetHealth == nil {
				continue
			}

			instanceID := aws.ToString(v.Target.Id)
			if instanceID == "" {
				continue
			}
			state := string(v.TargetHealth.State)
			if state == "" {
				continue
			}

			instanceStates[targetGroupARN][instanceID] = state
		}
	}

	return instanceStates, nil
}

func findGroup(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeAutoScalingGroupsInput) (*awstypes.AutoScalingGroup, error) {
	output, err := findGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findGroups(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeAutoScalingGroupsInput) ([]awstypes.AutoScalingGroup, error) {
	var output []awstypes.AutoScalingGroup

	pages := autoscaling.NewDescribeAutoScalingGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.AutoScalingGroups...)
	}

	return output, nil
}

func findGroupByName(ctx context.Context, conn *autoscaling.Client, name string) (*awstypes.AutoScalingGroup, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{name},
	}

	output, err := findGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.AutoScalingGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findInstanceRefresh(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeInstanceRefreshesInput) (*awstypes.InstanceRefresh, error) {
	output, err := findInstanceRefreshes(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceRefreshes(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeInstanceRefreshesInput) ([]awstypes.InstanceRefresh, error) {
	var output []awstypes.InstanceRefresh

	pages := autoscaling.NewDescribeInstanceRefreshesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.InstanceRefreshes...)
	}

	return output, nil
}

func findLoadBalancerStates(ctx context.Context, conn *autoscaling.Client, name string) ([]awstypes.LoadBalancerState, error) {
	input := &autoscaling.DescribeLoadBalancersInput{
		AutoScalingGroupName: aws.String(name),
	}
	var output []awstypes.LoadBalancerState

	pages := autoscaling.NewDescribeLoadBalancersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.LoadBalancers...)
	}

	return output, nil
}

func findLoadBalancerTargetGroupStates(ctx context.Context, conn *autoscaling.Client, name string) ([]awstypes.LoadBalancerTargetGroupState, error) {
	input := &autoscaling.DescribeLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(name),
	}
	var output []awstypes.LoadBalancerTargetGroupState

	pages := autoscaling.NewDescribeLoadBalancerTargetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.LoadBalancerTargetGroups...)
	}

	return output, nil
}

func findScalingActivities(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeScalingActivitiesInput, startTime time.Time) ([]awstypes.Activity, error) {
	var output []awstypes.Activity

	pages := autoscaling.NewDescribeScalingActivitiesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, activity := range page.Activities {
			if startTime.Before(aws.ToTime(activity.StartTime)) {
				output = append(output, activity)
			}
		}
	}

	return output, nil
}

func findScalingActivitiesByName(ctx context.Context, conn *autoscaling.Client, name string, startTime time.Time) ([]awstypes.Activity, error) {
	input := &autoscaling.DescribeScalingActivitiesInput{
		AutoScalingGroupName: aws.String(name),
	}

	return findScalingActivities(ctx, conn, input, startTime)
}

func findTrafficSourceStates(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeTrafficSourcesInput) ([]awstypes.TrafficSourceState, error) {
	var output []awstypes.TrafficSourceState

	pages := autoscaling.NewDescribeTrafficSourcesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TrafficSources...)
	}

	return output, nil
}

func findTrafficSourceStatesByTwoPartKey(ctx context.Context, conn *autoscaling.Client, asgName, trafficSourceType string) ([]awstypes.TrafficSourceState, error) {
	input := &autoscaling.DescribeTrafficSourcesInput{
		AutoScalingGroupName: aws.String(asgName),
	}
	if trafficSourceType != "" {
		input.TrafficSourceType = aws.String(trafficSourceType)
	}

	return findTrafficSourceStates(ctx, conn, input)
}

func findWarmPool(ctx context.Context, conn *autoscaling.Client, input *autoscaling.DescribeWarmPoolInput) (*autoscaling.DescribeWarmPoolOutput, error) {
	var output *autoscaling.DescribeWarmPoolOutput

	pages := autoscaling.NewDescribeWarmPoolPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrMessageContains(err, errCodeValidationError, "not found") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		if output == nil {
			output = page
		} else {
			output.Instances = append(output.Instances, page.Instances...)
		}
	}

	if output == nil || output.WarmPoolConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findWarmPoolByName(ctx context.Context, conn *autoscaling.Client, name string) (*autoscaling.DescribeWarmPoolOutput, error) {
	input := &autoscaling.DescribeWarmPoolInput{
		AutoScalingGroupName: aws.String(name),
	}

	return findWarmPool(ctx, conn, input)
}

func statusGroupCapacity(ctx context.Context, conn *autoscaling.Client, elbconn *elasticloadbalancing.Client, elbv2conn *elasticloadbalancingv2.Client, name string, cb func(int, int) error, startTime time.Time, ignoreFailedScalingActivities bool) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		if !ignoreFailedScalingActivities {
			// Check for fatal error in activity logs.
			scalingActivities, err := findScalingActivitiesByName(ctx, conn, name, startTime)

			if err != nil {
				return nil, "", fmt.Errorf("reading scaling activities: %w", err)
			}

			var errs []error

			for _, v := range scalingActivities {
				if v.StatusCode == awstypes.ScalingActivityStatusCodeFailed && aws.ToInt32(v.Progress) == 100 {
					if strings.Contains(aws.ToString(v.StatusMessage), "Invalid IAM Instance Profile") {
						// the activity will likely be retried
						continue
					}
					errs = append(errs, fmt.Errorf("scaling activity (%s): %s: %s", aws.ToString(v.ActivityId), v.StatusCode, aws.ToString(v.StatusMessage)))
				}
			}

			err = errors.Join(errs...)

			if err != nil {
				return nil, "", err
			}
		}

		g, err := findGroupByName(ctx, conn, name)

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
			instanceID := aws.ToString(v.InstanceId)
			if instanceID == "" {
				continue
			}

			if aws.ToString(v.HealthStatus) != InstanceHealthStatusHealthy {
				continue
			}

			if v.LifecycleState != awstypes.LifecycleStateInService {
				continue
			}

			increment := 1
			if v := aws.ToString(v.WeightedCapacity); v != "" {
				v, _ := strconv.Atoi(v)
				increment = v
			}

			nASG += increment

			inAll := true

			for _, v := range lbInstanceStates {
				if state, ok := v[instanceID]; ok && state != elbInstanceStateInService {
					inAll = false
					break
				}
			}

			if inAll {
				for _, v := range targetGroupInstanceStates {
					if state, ok := v[instanceID]; ok && state != string(elasticloadbalancingv2types.TargetHealthStateEnumHealthy) {
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

func statusGroupInstanceCount(ctx context.Context, conn *autoscaling.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.Itoa(len(output.Instances)), nil
	}
}

func statusInstanceRefresh(ctx context.Context, conn *autoscaling.Client, name, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: aws.String(name),
			InstanceRefreshIds:   []string{id},
		}

		output, err := findInstanceRefresh(ctx, conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusLoadBalancerInStateCount(ctx context.Context, conn *autoscaling.Client, name string, states ...string) retry.StateRefreshFunc {
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
				if aws.ToString(v.State) == state {
					count++
					break
				}
			}
		}

		return output, strconv.Itoa(count), nil
	}
}

func statusLoadBalancerTargetGroupInStateCount(ctx context.Context, conn *autoscaling.Client, name string, states ...string) retry.StateRefreshFunc {
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
				if aws.ToString(v.State) == state {
					count++
					break
				}
			}
		}

		return output, strconv.Itoa(count), nil
	}
}

func statusTrafficSourcesInStateCount(ctx context.Context, conn *autoscaling.Client, asgName, trafficSourceType string, states ...string) retry.StateRefreshFunc {
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
				if aws.ToString(v.State) == state {
					count++
					break
				}
			}
		}

		return output, strconv.Itoa(count), nil
	}
}

func statusWarmPool(ctx context.Context, conn *autoscaling.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findWarmPoolByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.WarmPoolConfiguration, string(output.WarmPoolConfiguration.Status), nil
	}
}

func statusWarmPoolInstanceCount(ctx context.Context, conn *autoscaling.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findWarmPoolByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.Itoa(len(output.Instances)), nil
	}
}

func waitGroupCapacitySatisfied(ctx context.Context, conn *autoscaling.Client, elbconn *elasticloadbalancing.Client, elbv2conn *elasticloadbalancingv2.Client, name string, cb func(int, int) error, startTime time.Time, ignoreFailedScalingActivities bool, timeout time.Duration) error {
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

func waitGroupDrained(ctx context.Context, conn *autoscaling.Client, name string, timeout time.Duration) (*awstypes.AutoScalingGroup, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusGroupInstanceCount(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AutoScalingGroup); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancersAdded(ctx context.Context, conn *autoscaling.Client, name string, timeout time.Duration) ([]*awstypes.LoadBalancerState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerInStateCount(ctx, conn, name, LoadBalancerStateAdding),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*awstypes.LoadBalancerState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancersRemoved(ctx context.Context, conn *autoscaling.Client, name string, timeout time.Duration) ([]*awstypes.LoadBalancerState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerInStateCount(ctx, conn, name, LoadBalancerStateRemoving),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*awstypes.LoadBalancerState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancerTargetGroupsAdded(ctx context.Context, conn *autoscaling.Client, name string, timeout time.Duration) ([]*awstypes.LoadBalancerTargetGroupState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerTargetGroupInStateCount(ctx, conn, name, LoadBalancerTargetGroupStateAdding),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*awstypes.LoadBalancerTargetGroupState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancerTargetGroupsRemoved(ctx context.Context, conn *autoscaling.Client, name string, timeout time.Duration) ([]*awstypes.LoadBalancerTargetGroupState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerTargetGroupInStateCount(ctx, conn, name, LoadBalancerTargetGroupStateRemoving),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*awstypes.LoadBalancerTargetGroupState); ok {
		return output, err
	}

	return nil, err
}

func waitTrafficSourcesCreated(ctx context.Context, conn *autoscaling.Client, asgName, trafficSourceType string, timeout time.Duration) ([]*awstypes.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusTrafficSourcesInStateCount(ctx, conn, asgName, trafficSourceType, TrafficSourceStateAdding),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*awstypes.TrafficSourceState); ok {
		return output, err
	}

	return nil, err
}

func waitTrafficSourcesDeleted(ctx context.Context, conn *autoscaling.Client, asgName, trafficSourceType string, timeout time.Duration) ([]*awstypes.TrafficSourceState, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusTrafficSourcesInStateCount(ctx, conn, asgName, trafficSourceType, TrafficSourceStateRemoving),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*awstypes.TrafficSourceState); ok {
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

func waitInstanceRefreshCancelled(ctx context.Context, conn *autoscaling.Client, name, id string, timeout time.Duration) (*awstypes.InstanceRefresh, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.InstanceRefreshStatusCancelling,
			awstypes.InstanceRefreshStatusInProgress,
			awstypes.InstanceRefreshStatusPending,
		),
		Target: enum.Slice(
			awstypes.InstanceRefreshStatusCancelled,
			awstypes.InstanceRefreshStatusFailed,
			awstypes.InstanceRefreshStatusSuccessful,
		),
		Refresh: statusInstanceRefresh(ctx, conn, name, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InstanceRefresh); ok {
		return output, err
	}

	return nil, err
}

func waitWarmPoolDeleted(ctx context.Context, conn *autoscaling.Client, name string, timeout time.Duration) (*awstypes.WarmPoolConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WarmPoolStatusPendingDelete),
		Target:  []string{},
		Refresh: statusWarmPool(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.WarmPoolConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitWarmPoolDrained(ctx context.Context, conn *autoscaling.Client, name string, timeout time.Duration) (*autoscaling.DescribeWarmPoolOutput, error) {
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

func expandInstancesDistribution(tfMap map[string]interface{}) *awstypes.InstancesDistribution {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstancesDistribution{}

	if v, ok := tfMap["on_demand_allocation_strategy"].(string); ok && v != "" {
		apiObject.OnDemandAllocationStrategy = aws.String(v)
	}

	if v, ok := tfMap["on_demand_base_capacity"].(int); ok {
		apiObject.OnDemandBaseCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["on_demand_percentage_above_base_capacity"].(int); ok {
		apiObject.OnDemandPercentageAboveBaseCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["spot_allocation_strategy"].(string); ok && v != "" {
		apiObject.SpotAllocationStrategy = aws.String(v)
	}

	if v, ok := tfMap["spot_instance_pools"].(int); ok && v != 0 {
		apiObject.SpotInstancePools = aws.Int32(int32(v))
	}

	if v, ok := tfMap["spot_max_price"].(string); ok {
		apiObject.SpotMaxPrice = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplate(tfMap map[string]interface{}, hasDefaultVersion bool) *awstypes.LaunchTemplate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplate{}

	if v, ok := tfMap["launch_template_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplateSpecification = expandLaunchTemplateSpecificationForMixedInstancesPolicy(v[0].(map[string]interface{}), hasDefaultVersion)
	}

	if v, ok := tfMap["override"].([]interface{}); ok && len(v) > 0 {
		apiObject.Overrides = expandLaunchTemplateOverrideses(v, hasDefaultVersion)
	}

	return apiObject
}

func expandLaunchTemplateOverrides(tfMap map[string]interface{}, hasDefaultVersion bool) awstypes.LaunchTemplateOverrides {
	apiObject := awstypes.LaunchTemplateOverrides{}

	if v, ok := tfMap["instance_requirements"].([]interface{}); ok && len(v) > 0 {
		apiObject.InstanceRequirements = expandInstanceRequirements(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["launch_template_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplateSpecification = expandLaunchTemplateSpecificationForMixedInstancesPolicy(v[0].(map[string]interface{}), hasDefaultVersion)
	}

	if v, ok := tfMap[names.AttrInstanceType].(string); ok && v != "" {
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := tfMap["weighted_capacity"].(string); ok && v != "" {
		apiObject.WeightedCapacity = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateOverrideses(tfList []interface{}, hasDefaultVersion bool) []awstypes.LaunchTemplateOverrides {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateOverrides

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateOverrides(tfMap, hasDefaultVersion)
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandInstanceRequirements(tfMap map[string]interface{}) *awstypes.InstanceRequirements {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstanceRequirements{}

	if v, ok := tfMap["accelerator_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.AcceleratorCount = expandAcceleratorCountRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["accelerator_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorManufacturers = flex.ExpandStringyValueSet[awstypes.AcceleratorManufacturer](v)
	}

	if v, ok := tfMap["accelerator_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorNames = flex.ExpandStringyValueSet[awstypes.AcceleratorName](v)
	}

	if v, ok := tfMap["accelerator_total_memory_mib"].([]interface{}); ok && len(v) > 0 {
		apiObject.AcceleratorTotalMemoryMiB = expandAcceleratorTotalMemoryMiBRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["accelerator_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorTypes = flex.ExpandStringyValueSet[awstypes.AcceleratorType](v)
	}

	if v, ok := tfMap["allowed_instance_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AllowedInstanceTypes = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["bare_metal"].(string); ok && v != "" {
		apiObject.BareMetal = awstypes.BareMetal(v)
	}

	if v, ok := tfMap["baseline_ebs_bandwidth_mbps"].([]interface{}); ok && len(v) > 0 {
		apiObject.BaselineEbsBandwidthMbps = expandBaselineEBSBandwidthMbpsRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["burstable_performance"].(string); ok && v != "" {
		apiObject.BurstablePerformance = awstypes.BurstablePerformance(v)
	}

	if v, ok := tfMap["cpu_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CpuManufacturers = flex.ExpandStringyValueSet[awstypes.CpuManufacturer](v)
	}

	if v, ok := tfMap["excluded_instance_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ExcludedInstanceTypes = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["instance_generations"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.InstanceGenerations = flex.ExpandStringyValueSet[awstypes.InstanceGeneration](v)
	}

	if v, ok := tfMap["local_storage"].(string); ok && v != "" {
		apiObject.LocalStorage = awstypes.LocalStorage(v)
	}

	if v, ok := tfMap["local_storage_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.LocalStorageTypes = flex.ExpandStringyValueSet[awstypes.LocalStorageType](v)
	}

	if v, ok := tfMap["max_spot_price_as_percentage_of_optimal_on_demand_price"].(int); ok && v != 0 {
		apiObject.MaxSpotPriceAsPercentageOfOptimalOnDemandPrice = aws.Int32(int32(v))
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
		apiObject.OnDemandMaxPricePercentageOverLowestPrice = aws.Int32(int32(v))
	}

	if v, ok := tfMap["require_hibernate_support"].(bool); ok && v {
		apiObject.RequireHibernateSupport = aws.Bool(v)
	}

	if v, ok := tfMap["spot_max_price_percentage_over_lowest_price"].(int); ok && v != 0 {
		apiObject.SpotMaxPricePercentageOverLowestPrice = aws.Int32(int32(v))
	}

	if v, ok := tfMap["total_local_storage_gb"].([]interface{}); ok && len(v) > 0 {
		apiObject.TotalLocalStorageGB = expandTotalLocalStorageGBRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["vcpu_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.VCpuCount = expandVCPUCountRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAcceleratorCountRequest(tfMap map[string]interface{}) *awstypes.AcceleratorCountRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AcceleratorCountRequest{}

	var min int
	if v, ok := tfMap[names.AttrMin].(int); ok {
		min = v
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v >= min {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandAcceleratorTotalMemoryMiBRequest(tfMap map[string]interface{}) *awstypes.AcceleratorTotalMemoryMiBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AcceleratorTotalMemoryMiBRequest{}

	var min int
	if v, ok := tfMap[names.AttrMin].(int); ok {
		min = v
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v >= min {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandBaselineEBSBandwidthMbpsRequest(tfMap map[string]interface{}) *awstypes.BaselineEbsBandwidthMbpsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.BaselineEbsBandwidthMbpsRequest{}

	var min int
	if v, ok := tfMap[names.AttrMin].(int); ok {
		min = v
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v >= min {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandMemoryGiBPerVCPURequest(tfMap map[string]interface{}) *awstypes.MemoryGiBPerVCpuRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MemoryGiBPerVCpuRequest{}

	var min float64
	if v, ok := tfMap[names.AttrMin].(float64); ok {
		min = v
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrMax].(float64); ok && v >= min {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandMemoryMiBRequest(tfMap map[string]interface{}) *awstypes.MemoryMiBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MemoryMiBRequest{}

	var min int
	if v, ok := tfMap[names.AttrMin].(int); ok {
		min = v
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v >= min {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandNetworkBandwidthGbpsRequest(tfMap map[string]interface{}) *awstypes.NetworkBandwidthGbpsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkBandwidthGbpsRequest{}

	var min float64
	if v, ok := tfMap[names.AttrMin].(float64); ok {
		min = v
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrMax].(float64); ok && v >= min {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandNetworkInterfaceCountRequest(tfMap map[string]interface{}) *awstypes.NetworkInterfaceCountRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkInterfaceCountRequest{}

	var min int
	if v, ok := tfMap[names.AttrMin].(int); ok {
		min = v
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v >= min {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandTotalLocalStorageGBRequest(tfMap map[string]interface{}) *awstypes.TotalLocalStorageGBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TotalLocalStorageGBRequest{}

	var min float64
	if v, ok := tfMap[names.AttrMin].(float64); ok {
		min = v
		apiObject.Min = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrMax].(float64); ok && v >= min {
		apiObject.Max = aws.Float64(v)
	}

	return apiObject
}

func expandVCPUCountRequest(tfMap map[string]interface{}) *awstypes.VCpuCountRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VCpuCountRequest{}

	min := 0
	if v, ok := tfMap[names.AttrMin].(int); ok {
		min = v
		apiObject.Min = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMax].(int); ok && v >= min {
		apiObject.Max = aws.Int32(int32(v))
	}

	return apiObject
}

func expandLaunchTemplateSpecificationForMixedInstancesPolicy(tfMap map[string]interface{}, hasDefaultVersion bool) *awstypes.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateSpecification{}

	// API returns both ID and name, which Terraform saves to state. Next update returns:
	// ValidationError: Valid requests must contain either launchTemplateId or LaunchTemplateName
	// Prefer the ID if we have both.
	if v, ok := tfMap["launch_template_id"]; ok && v != "" && v != launchTemplateIDUnknown {
		apiObject.LaunchTemplateId = aws.String(v.(string))
	} else if v, ok := tfMap["launch_template_name"]; ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v.(string))
	}

	if hasDefaultVersion {
		apiObject.Version = aws.String("$Default")
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateSpecification(tfMap map[string]interface{}, hasDefaultVersion bool) *awstypes.LaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateSpecification{}

	// DescribeAutoScalingGroups returns both name and id but LaunchTemplateSpecification
	// allows only one of them to be set.
	if v, ok := tfMap[names.AttrID]; ok && v != "" && v != launchTemplateIDUnknown {
		apiObject.LaunchTemplateId = aws.String(v.(string))
	} else if v, ok := tfMap[names.AttrName]; ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v.(string))
	}

	if hasDefaultVersion {
		apiObject.Version = aws.String("$Default")
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandMixedInstancesPolicy(tfMap map[string]interface{}, hasDefaultVersion bool) *awstypes.MixedInstancesPolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MixedInstancesPolicy{}

	if v, ok := tfMap["instances_distribution"].([]interface{}); ok && len(v) > 0 {
		apiObject.InstancesDistribution = expandInstancesDistribution(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrLaunchTemplate].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplate = expandLaunchTemplate(v[0].(map[string]interface{}), hasDefaultVersion)
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
		apiObject.HeartbeatTimeout = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
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

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
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

	if v, ok := tfMap["max_group_prepared_capacity"].(int); ok {
		apiObject.MaxGroupPreparedCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_size"].(int); ok && v != 0 {
		apiObject.MinSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap["pool_state"].(string); ok && v != "" {
		apiObject.PoolState = awstypes.WarmPoolState(v)
	}

	return apiObject
}

func expandInstanceReusePolicy(tfMap map[string]interface{}) *awstypes.InstanceReusePolicy {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstanceReusePolicy{}

	if v, ok := tfMap["reuse_on_scale_in"].(bool); ok {
		apiObject.ReuseOnScaleIn = aws.Bool(v)
	}

	return apiObject
}

func expandStartInstanceRefreshInput(name string, tfMap map[string]interface{}, launchTemplate *awstypes.LaunchTemplateSpecification, mixedInstancesPolicy *awstypes.MixedInstancesPolicy) *autoscaling.StartInstanceRefreshInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.StartInstanceRefreshInput{
		AutoScalingGroupName: aws.String(name),
	}

	if v, ok := tfMap["preferences"].([]interface{}); ok && len(v) > 0 {
		apiObject.Preferences = expandRefreshPreferences(v[0].(map[string]interface{}))

		// "The AutoRollback parameter cannot be set to true when the DesiredConfiguration parameter is empty".
		if aws.ToBool(apiObject.Preferences.AutoRollback) {
			apiObject.DesiredConfiguration = &awstypes.DesiredConfiguration{
				LaunchTemplate:       launchTemplate,
				MixedInstancesPolicy: mixedInstancesPolicy,
			}
		}
	}

	if v, ok := tfMap["strategy"].(string); ok && v != "" {
		apiObject.Strategy = awstypes.RefreshStrategy(v)
	}

	return apiObject
}

func expandRefreshPreferences(tfMap map[string]interface{}) *awstypes.RefreshPreferences {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.RefreshPreferences{}

	if v, ok := tfMap["alarm_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.AlarmSpecification = expandAlarmSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["auto_rollback"].(bool); ok {
		apiObject.AutoRollback = aws.Bool(v)
	}

	if v, ok := tfMap["checkpoint_delay"].(string); ok {
		if v, null, _ := nullable.Int(v).ValueInt32(); !null {
			apiObject.CheckpointDelay = aws.Int32(v)
		}
	}

	if v, ok := tfMap["checkpoint_percentages"].([]interface{}); ok && len(v) > 0 {
		apiObject.CheckpointPercentages = flex.ExpandInt32ValueList(v)
	}

	if v, ok := tfMap["instance_warmup"].(string); ok {
		if v, null, _ := nullable.Int(v).ValueInt32(); !null {
			apiObject.InstanceWarmup = aws.Int32(v)
		}
	}

	if v, ok := tfMap["max_healthy_percentage"].(int); ok {
		apiObject.MaxHealthyPercentage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["min_healthy_percentage"].(int); ok {
		apiObject.MinHealthyPercentage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["scale_in_protected_instances"].(string); ok {
		apiObject.ScaleInProtectedInstances = awstypes.ScaleInProtectedInstances(v)
	}

	if v, ok := tfMap["skip_matching"].(bool); ok {
		apiObject.SkipMatching = aws.Bool(v)
	}

	if v, ok := tfMap["standby_instances"].(string); ok {
		apiObject.StandbyInstances = awstypes.StandbyInstances(v)
	}

	return apiObject
}

func expandAlarmSpecification(tfMap map[string]interface{}) *awstypes.AlarmSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AlarmSpecification{}

	if v, ok := tfMap["alarms"].([]interface{}); ok {
		apiObject.Alarms = flex.ExpandStringValueList(v)
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

func expandTrafficSourceIdentifier(tfMap map[string]interface{}) awstypes.TrafficSourceIdentifier {
	apiObject := awstypes.TrafficSourceIdentifier{}

	if v, ok := tfMap[names.AttrIdentifier].(string); ok && v != "" {
		apiObject.Identifier = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandTrafficSourceIdentifiers(tfList []interface{}) []awstypes.TrafficSourceIdentifier {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.TrafficSourceIdentifier

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandTrafficSourceIdentifier(tfMap)
		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenEnabledMetrics(apiObjects []awstypes.EnabledMetric) []string {
	var tfList []string

	for _, apiObject := range apiObjects {
		if v := apiObject.Metric; v != nil {
			tfList = append(tfList, aws.ToString(v))
		}
	}

	return tfList
}

func flattenLaunchTemplateSpecification(apiObject *awstypes.LaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap[names.AttrID] = aws.ToString(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap[names.AttrVersion] = aws.ToString(v)
	}

	return tfMap
}

func flattenMixedInstancesPolicy(apiObject *awstypes.MixedInstancesPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.InstancesDistribution; v != nil {
		tfMap["instances_distribution"] = []interface{}{flattenInstancesDistribution(v)}
	}

	if v := apiObject.LaunchTemplate; v != nil {
		tfMap[names.AttrLaunchTemplate] = []interface{}{flattenLaunchTemplate(v)}
	}

	return tfMap
}

func flattenInstancesDistribution(apiObject *awstypes.InstancesDistribution) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OnDemandAllocationStrategy; v != nil {
		tfMap["on_demand_allocation_strategy"] = aws.ToString(v)
	}

	if v := apiObject.OnDemandBaseCapacity; v != nil {
		tfMap["on_demand_base_capacity"] = aws.ToInt32(v)
	}

	if v := apiObject.OnDemandPercentageAboveBaseCapacity; v != nil {
		tfMap["on_demand_percentage_above_base_capacity"] = aws.ToInt32(v)
	}

	if v := apiObject.SpotAllocationStrategy; v != nil {
		tfMap["spot_allocation_strategy"] = aws.ToString(v)
	}

	if v := apiObject.SpotInstancePools; v != nil {
		tfMap["spot_instance_pools"] = aws.ToInt32(v)
	}

	if v := apiObject.SpotMaxPrice; v != nil {
		tfMap["spot_max_price"] = aws.ToString(v)
	}

	return tfMap
}

func expandInstanceMaintenancePolicy(l []interface{}) *awstypes.InstanceMaintenancePolicy {
	if len(l) == 0 {
		//Empty InstanceMaintenancePolicy block will reset already assigned values
		return &awstypes.InstanceMaintenancePolicy{
			MinHealthyPercentage: aws.Int32(-1),
			MaxHealthyPercentage: aws.Int32(-1),
		}
	}

	tfMap := l[0].(map[string]interface{})

	return &awstypes.InstanceMaintenancePolicy{
		MinHealthyPercentage: aws.Int32(int32(tfMap["min_healthy_percentage"].(int))),
		MaxHealthyPercentage: aws.Int32(int32(tfMap["max_healthy_percentage"].(int))),
	}
}

func flattenInstanceMaintenancePolicy(instanceMaintenancePolicy *awstypes.InstanceMaintenancePolicy) []interface{} {
	if instanceMaintenancePolicy == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"min_healthy_percentage": instanceMaintenancePolicy.MinHealthyPercentage,
		"max_healthy_percentage": instanceMaintenancePolicy.MaxHealthyPercentage,
	}

	return []interface{}{m}
}

func flattenLaunchTemplate(apiObject *awstypes.LaunchTemplate) map[string]interface{} {
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

func flattenLaunchTemplateSpecificationForMixedInstancesPolicy(apiObject *awstypes.LaunchTemplateSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateId; v != nil {
		tfMap["launch_template_id"] = aws.ToString(v)
	}

	if v := apiObject.LaunchTemplateName; v != nil {
		tfMap["launch_template_name"] = aws.ToString(v)
	}

	if v := apiObject.Version; v != nil {
		tfMap[names.AttrVersion] = aws.ToString(v)
	}

	return tfMap
}

func flattenLaunchTemplateOverrides(apiObject awstypes.LaunchTemplateOverrides) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.InstanceRequirements; v != nil {
		tfMap["instance_requirements"] = []interface{}{flattenInstanceRequirements(v)}
	}

	if v := apiObject.InstanceType; v != nil {
		tfMap[names.AttrInstanceType] = aws.ToString(v)
	}

	if v := apiObject.LaunchTemplateSpecification; v != nil {
		tfMap["launch_template_specification"] = []interface{}{flattenLaunchTemplateSpecificationForMixedInstancesPolicy(v)}
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.ToString(v)
	}

	return tfMap
}

func flattenLaunchTemplateOverrideses(apiObjects []awstypes.LaunchTemplateOverrides) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateOverrides(apiObject))
	}

	return tfList
}

func flattenInstanceRequirements(apiObject *awstypes.InstanceRequirements) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AcceleratorCount; v != nil {
		tfMap["accelerator_count"] = []interface{}{flattenAcceleratorCount(v)}
	}

	if apiObject.AcceleratorManufacturers != nil {
		tfMap["accelerator_manufacturers"] = apiObject.AcceleratorManufacturers
	}

	if apiObject.AcceleratorNames != nil {
		tfMap["accelerator_names"] = apiObject.AcceleratorNames
	}

	if v := apiObject.AcceleratorTotalMemoryMiB; v != nil {
		tfMap["accelerator_total_memory_mib"] = []interface{}{flattenAcceleratorTotalMemoryMiB(v)}
	}

	if apiObject.AcceleratorTypes != nil {
		tfMap["accelerator_types"] = apiObject.AcceleratorTypes
	}

	if apiObject.AllowedInstanceTypes != nil {
		tfMap["allowed_instance_types"] = apiObject.AllowedInstanceTypes
	}

	tfMap["bare_metal"] = apiObject.BareMetal

	if v := apiObject.BaselineEbsBandwidthMbps; v != nil {
		tfMap["baseline_ebs_bandwidth_mbps"] = []interface{}{flattenBaselineEBSBandwidthMbps(v)}
	}

	tfMap["burstable_performance"] = apiObject.BurstablePerformance

	if v := apiObject.CpuManufacturers; v != nil {
		tfMap["cpu_manufacturers"] = apiObject.CpuManufacturers
	}

	if v := apiObject.ExcludedInstanceTypes; v != nil {
		tfMap["excluded_instance_types"] = apiObject.ExcludedInstanceTypes
	}

	if v := apiObject.InstanceGenerations; v != nil {
		tfMap["instance_generations"] = apiObject.InstanceGenerations
	}

	tfMap["local_storage"] = apiObject.LocalStorage

	if v := apiObject.LocalStorageTypes; v != nil {
		tfMap["local_storage_types"] = apiObject.LocalStorageTypes
	}

	if v := apiObject.MaxSpotPriceAsPercentageOfOptimalOnDemandPrice; v != nil {
		tfMap["max_spot_price_as_percentage_of_optimal_on_demand_price"] = aws.ToInt32(v)
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
		tfMap["on_demand_max_price_percentage_over_lowest_price"] = aws.ToInt32(v)
	}

	if v := apiObject.RequireHibernateSupport; v != nil {
		tfMap["require_hibernate_support"] = aws.ToBool(v)
	}

	if v := apiObject.SpotMaxPricePercentageOverLowestPrice; v != nil {
		tfMap["spot_max_price_percentage_over_lowest_price"] = aws.ToInt32(v)
	}

	if v := apiObject.TotalLocalStorageGB; v != nil {
		tfMap["total_local_storage_gb"] = []interface{}{flattentTotalLocalStorageGB(v)}
	}

	if v := apiObject.VCpuCount; v != nil {
		tfMap["vcpu_count"] = []interface{}{flattenVCPUCount(v)}
	}

	return tfMap
}

func flattenAcceleratorCount(apiObject *awstypes.AcceleratorCountRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToInt32(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenAcceleratorTotalMemoryMiB(apiObject *awstypes.AcceleratorTotalMemoryMiBRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToInt32(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenBaselineEBSBandwidthMbps(apiObject *awstypes.BaselineEbsBandwidthMbpsRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToInt32(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenMemoryGiBPerVCPU(apiObject *awstypes.MemoryGiBPerVCpuRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToFloat64(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToFloat64(v)
	}

	return tfMap
}

func flattenMemoryMiB(apiObject *awstypes.MemoryMiBRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToInt32(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenNetworkBandwidthGbps(apiObject *awstypes.NetworkBandwidthGbpsRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToFloat64(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToFloat64(v)
	}

	return tfMap
}

func flattenNetworkInterfaceCount(apiObject *awstypes.NetworkInterfaceCountRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToInt32(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToInt32(v)
	}

	return tfMap
}

func flattentTotalLocalStorageGB(apiObject *awstypes.TotalLocalStorageGBRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToFloat64(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToFloat64(v)
	}

	return tfMap
}

func flattenTrafficSourceIdentifier(apiObject awstypes.TrafficSourceIdentifier) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Identifier; v != nil {
		tfMap[names.AttrIdentifier] = aws.ToString(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	return tfMap
}

func flattenTrafficSourceIdentifiers(apiObjects []awstypes.TrafficSourceIdentifier) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenTrafficSourceIdentifier(apiObject))
	}

	return tfList
}

func flattenVCPUCount(apiObject *awstypes.VCpuCountRequest) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Max; v != nil {
		tfMap[names.AttrMax] = aws.ToInt32(v)
	}

	if v := apiObject.Min; v != nil {
		tfMap[names.AttrMin] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenSuspendedProcesses(apiObjects []awstypes.SuspendedProcess) []string {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []string

	for _, apiObject := range apiObjects {
		if v := apiObject.ProcessName; v != nil {
			tfList = append(tfList, aws.ToString(v))
		}
	}

	return tfList
}

func flattenWarmPoolConfiguration(apiObject *awstypes.WarmPoolConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.InstanceReusePolicy; v != nil {
		tfMap["instance_reuse_policy"] = []interface{}{flattenWarmPoolInstanceReusePolicy(v)}
	}

	if v := apiObject.MaxGroupPreparedCapacity; v != nil {
		tfMap["max_group_prepared_capacity"] = aws.ToInt32(v)
	} else {
		tfMap["max_group_prepared_capacity"] = int64(defaultWarmPoolMaxGroupPreparedCapacity)
	}

	if v := apiObject.MinSize; v != nil {
		tfMap["min_size"] = aws.ToInt32(v)
	}

	tfMap["pool_state"] = apiObject.PoolState

	return tfMap
}

func flattenWarmPoolInstanceReusePolicy(apiObject *awstypes.InstanceReusePolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ReuseOnScaleIn; v != nil {
		tfMap["reuse_on_scale_in"] = aws.ToBool(v)
	}

	return tfMap
}

func cancelInstanceRefresh(ctx context.Context, conn *autoscaling.Client, name string) error {
	input := &autoscaling.CancelInstanceRefreshInput{
		AutoScalingGroupName: aws.String(name),
	}

	output, err := conn.CancelInstanceRefresh(ctx, input)

	if errs.IsA[*awstypes.ActiveInstanceRefreshNotFoundFault](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cancelling Auto Scaling Group (%s) instance refresh: %w", name, err)
	}

	_, err = waitInstanceRefreshCancelled(ctx, conn, name, aws.ToString(output.InstanceRefreshId), instanceRefreshCancelledTimeout)

	if err != nil {
		return fmt.Errorf("waiting for Auto Scaling Group (%s) instance refresh cancel: %w", name, err)
	}

	return nil
}

func startInstanceRefresh(ctx context.Context, conn *autoscaling.Client, input *autoscaling.StartInstanceRefreshInput) error {
	name := aws.ToString(input.AutoScalingGroupName)

	_, err := tfresource.RetryWhen(ctx, instanceRefreshStartedTimeout,
		func() (interface{}, error) {
			return conn.StartInstanceRefresh(ctx, input)
		},
		func(err error) (bool, error) {
			if errs.IsA[*awstypes.InstanceRefreshInProgressFault](err) {
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
	var diags diag.Diagnostics

	v, ok := i.(string)
	if !ok {
		return sdkdiag.AppendErrorf(diags, "expected type to be string")
	}

	if v == "launch_configuration" || v == names.AttrLaunchTemplate || v == "mixed_instances_policy" {
		return diag.Diagnostics{
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("'%s' always triggers an instance refresh and can be removed", v),
			},
		}
	}

	schema := resourceGroup().SchemaMap()
	for attr, attrSchema := range schema {
		if v == attr {
			if attrSchema.Computed && !attrSchema.Optional {
				return sdkdiag.AppendErrorf(diags, "'%s' is a read-only parameter and cannot be used to trigger an instance refresh", v)
			}
			return diags
		}
	}

	return sdkdiag.AppendErrorf(diags, "'%s' is not a recognized parameter name for aws_autoscaling_group", v)
}
