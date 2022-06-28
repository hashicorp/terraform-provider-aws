package autoscaling

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceGroupCreate,
		Read:   resourceGroupRead,
		Update: resourceGroupUpdate,
		Delete: resourceGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
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
			"desired_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
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
						"name": {
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
						"role_arn": {
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
									"skip_matching": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
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
				Type:     schema.TypeSet,
				Optional: true,
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
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[1-9][0-9]{0,2}$`), "see https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_LaunchTemplateOverrides.html"),
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
				ValidateFunc:  validation.StringLenBetween(0, 255-resource.UniqueIDSuffixLength),
				ConflictsWith: []string{"name"},
			},
			"placement_group": {
				Type:     schema.TypeString,
				Optional: true,
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
				// This should be removable, but wait until other tags work is being done.
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["key"].(string)))
					buf.WriteString(fmt.Sprintf("%s-", m["value"].(string)))
					buf.WriteString(fmt.Sprintf("%t-", m["propagate_at_launch"].(bool)))

					return create.StringHashcode(buf.String())
				},
				ConflictsWith: []string{"tags"},
			},
			"tags": {
				Deprecated: "Use tag instead",
				Type:       schema.TypeSet,
				Optional:   true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
				// Terraform 0.11 and earlier can provide incorrect type
				// information during difference handling, in which boolean
				// values are represented as "0" and "1". This Set function
				// normalizes these hashing variations, while the Terraform
				// Plugin SDK automatically suppresses the boolean/string
				// difference in the value itself.
				Set: func(v interface{}) int {
					var buf bytes.Buffer

					m, ok := v.(map[string]interface{})

					if !ok {
						return 0
					}

					if v, ok := m["key"].(string); ok {
						buf.WriteString(fmt.Sprintf("%s-", v))
					}

					if v, ok := m["value"].(string); ok {
						buf.WriteString(fmt.Sprintf("%s-", v))
					}

					if v, ok := m["propagate_at_launch"].(bool); ok {
						buf.WriteString(fmt.Sprintf("%t-", v))
					} else if v, ok := m["propagate_at_launch"].(string); ok {
						if b, err := strconv.ParseBool(v); err == nil {
							buf.WriteString(fmt.Sprintf("%t-", b))
						} else {
							buf.WriteString(fmt.Sprintf("%s-", v))
						}
					}

					return create.StringHashcode(buf.String())
				},
				ConflictsWith: []string{"tag"},
			},
			"target_group_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"termination_policies": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("launch_template.0.id", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("launch_template.0.name")
			}),
			customdiff.ComputedIf("launch_template.0.name", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				return diff.HasChange("launch_template.0.id")
			}),
		),
	}
}

func resourceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

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

	maxSize := aws.Int64(int64(d.Get("max_size").(int)))
	minSize := aws.Int64(int64(d.Get("min_size").(int)))

	if twoPhases {
		createInput.MaxSize = aws.Int64(0)
		createInput.MinSize = aws.Int64(0)

		updateInput.MaxSize = maxSize
		updateInput.MinSize = minSize

		if v, ok := d.GetOk("desired_capacity"); ok {
			updateInput.DesiredCapacity = aws.Int64(int64(v.(int)))
		}
	} else {
		createInput.MaxSize = maxSize
		createInput.MinSize = minSize

		if v, ok := d.GetOk("desired_capacity"); ok {
			createInput.DesiredCapacity = aws.Int64(int64(v.(int)))
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
		createInput.Tags = Tags(KeyValueTags(v, asgName, TagResourceTypeGroup).IgnoreAWS())
	}

	if v, ok := d.GetOk("tags"); ok {
		createInput.Tags = Tags(KeyValueTags(v, asgName, TagResourceTypeGroup).IgnoreAWS())
	}

	if v, ok := d.GetOk("target_group_arns"); ok && len(v.(*schema.Set).List()) > 0 {
		createInput.TargetGroupARNs = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("termination_policies"); ok && len(v.([]interface{})) > 0 {
		createInput.TerminationPolicies = flex.ExpandStringList(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_zone_identifier"); ok && v.(*schema.Set).Len() > 0 {
		createInput.VPCZoneIdentifier = expandVPCZoneIdentifiers(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Creating Auto Scaling Group: %s", createInput)
	_, err := tfresource.RetryWhenAWSErrMessageContains(propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateAutoScalingGroup(createInput)
		},
		// ValidationError: You must use a valid fully-formed launch template. Value (tf-acc-test-6643732652421074386) for parameter iamInstanceProfile.name is invalid. Invalid IAM Instance Profile name
		ErrCodeValidationError, "Invalid IAM Instance Profile")

	if err != nil {
		return fmt.Errorf("creating Auto Scaling Group (%s): %w", asgName, err)
	}

	d.SetId(asgName)

	if twoPhases {
		for _, input := range expandPutLifecycleHookInputs(asgName, initialLifecycleHooks) {
			_, err := tfresource.RetryWhenAWSErrMessageContains(5*time.Minute,
				func() (interface{}, error) {
					return conn.PutLifecycleHook(input)
				},
				ErrCodeValidationError, "Unable to publish test message to notification target")

			if err != nil {
				return fmt.Errorf("creating Auto Scaling Group (%s) Lifecycle Hook: %w", d.Id(), err)
			}
		}

		_, err = conn.UpdateAutoScalingGroup(updateInput)

		if err != nil {
			return fmt.Errorf("setting Auto Scaling Group (%s) initial capacity: %w", d.Id(), err)
		}
	}

	if err := waitForASGCapacity(d, meta, CapacitySatisfiedCreate); err != nil {
		return err
	}

	if v, ok := d.GetOk("suspended_processes"); ok && v.(*schema.Set).Len() > 0 {
		input := &autoscaling.ScalingProcessQuery{
			AutoScalingGroupName: aws.String(d.Id()),
			ScalingProcesses:     flex.ExpandStringSet(v.(*schema.Set)),
		}

		_, err := conn.SuspendProcesses(input)

		if err != nil {
			return fmt.Errorf("suspending Auto Scaling Group (%s) scaling processes: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("enabled_metrics"); ok && v.(*schema.Set).Len() > 0 {
		input := &autoscaling.EnableMetricsCollectionInput{
			AutoScalingGroupName: aws.String(d.Id()),
			Granularity:          aws.String(d.Get("metrics_granularity").(string)),
			Metrics:              flex.ExpandStringSet(v.(*schema.Set)),
		}

		_, err := conn.EnableMetricsCollection(input)

		if err != nil {
			return fmt.Errorf("enabling Auto Scaling Group (%s) metrics collection: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("warm_pool"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		_, err := conn.PutWarmPool(expandPutWarmPoolInput(d.Id(), v.([]interface{})[0].(map[string]interface{})))

		if err != nil {
			return fmt.Errorf("creating Auto Scaling Warm Pool (%s): %w", d.Id(), err)
		}
	}

	return resourceGroupRead(d, meta)
}

func resourceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	g, err := FindGroupByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Auto Scaling Group %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Auto Scaling Group (%s): %w", d.Id(), err)
	}

	d.Set("arn", g.AutoScalingGroupARN)
	d.Set("availability_zones", aws.StringValueSlice(g.AvailabilityZones))
	d.Set("capacity_rebalance", g.CapacityRebalance)
	d.Set("context", g.Context)
	d.Set("default_cooldown", g.DefaultCooldown)
	d.Set("desired_capacity", g.DesiredCapacity)
	if len(g.EnabledMetrics) > 0 {
		d.Set("enabled_metrics", flattenEnabledMetrics(g.EnabledMetrics))
		d.Set("metrics_granularity", g.EnabledMetrics[0].Granularity)
	} else {
		d.Set("enabled_metrics", nil)
		d.Set("metrics_granularity", DefaultEnabledMetricsGranularity)
	}
	d.Set("health_check_grace_period", g.HealthCheckGracePeriod)
	d.Set("health_check_type", g.HealthCheckType)
	d.Set("load_balancers", aws.StringValueSlice(g.LoadBalancerNames))
	d.Set("launch_configuration", g.LaunchConfigurationName)
	if g.LaunchTemplate != nil {
		if err := d.Set("launch_template", []interface{}{flattenLaunchTemplateSpecification(g.LaunchTemplate)}); err != nil {
			return fmt.Errorf("setting launch_template: %w", err)
		}
	} else {
		d.Set("launch_template", nil)
	}
	d.Set("max_instance_lifetime", g.MaxInstanceLifetime)
	d.Set("max_size", g.MaxSize)
	d.Set("min_size", g.MinSize)
	if g.MixedInstancesPolicy != nil {
		if err := d.Set("mixed_instances_policy", []interface{}{flattenMixedInstancesPolicy(g.MixedInstancesPolicy)}); err != nil {
			return fmt.Errorf("setting mixed_instances_policy: %w", err)
		}
	} else {
		d.Set("mixed_instances_policy", nil)
	}
	d.Set("name", g.AutoScalingGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(g.AutoScalingGroupName)))
	d.Set("placement_group", g.PlacementGroup)
	d.Set("protect_from_scale_in", g.NewInstancesProtectedFromScaleIn)
	d.Set("service_linked_role_arn", g.ServiceLinkedRoleARN)
	d.Set("suspended_processes", flattenSuspendedProcesses(g.SuspendedProcesses))
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
			return fmt.Errorf("setting warm_pool: %w", err)
		}
	} else {
		d.Set("warm_pool", nil)
	}

	var tagOk, tagsOk bool
	var v interface{}

	// Deprecated: In a future major version, this should always set all tags except those ignored.
	//             Remove d.GetOk() and Only() handling.
	if v, tagOk = d.GetOk("tag"); tagOk {
		proposedStateTags := KeyValueTags(v, d.Id(), TagResourceTypeGroup)

		if err := d.Set("tag", ListOfMap(KeyValueTags(g.Tags, d.Id(), TagResourceTypeGroup).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Only(proposedStateTags))); err != nil {
			return fmt.Errorf("setting tag: %w", err)
		}
	}

	if v, tagsOk = d.GetOk("tags"); tagsOk {
		proposedStateTags := KeyValueTags(v, d.Id(), TagResourceTypeGroup)

		if err := d.Set("tags", ListOfStringMap(KeyValueTags(g.Tags, d.Id(), TagResourceTypeGroup).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Only(proposedStateTags))); err != nil {
			return fmt.Errorf("setting tags: %w", err)
		}
	}

	if !tagOk && !tagsOk {
		if err := d.Set("tag", ListOfMap(KeyValueTags(g.Tags, d.Id(), TagResourceTypeGroup).IgnoreAWS().IgnoreConfig(ignoreTagsConfig))); err != nil {
			return fmt.Errorf("setting tag: %w", err)
		}
	}

	return nil
}

func resourceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	var shouldWaitForCapacity bool
	var shouldRefreshInstances bool

	if d.HasChangesExcept(
		"enabled_metrics",
		"load_balancers",
		"suspended_processes",
		"tag",
		"tags",
		"target_group_arns",
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

		if d.HasChange("desired_capacity") {
			input.DesiredCapacity = aws.Int64(int64(d.Get("desired_capacity").(int)))
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

		log.Printf("[DEBUG] Updating Auto Scaling Group: %s", input)
		_, err := conn.UpdateAutoScalingGroup(input)

		if err != nil {
			return fmt.Errorf("updating Auto Scaling Group (%s): %w", d.Id(), err)
		}
	}

	if d.HasChanges("tag", "tags") {
		oTagRaw, nTagRaw := d.GetChange("tag")
		oTagsRaw, nTagsRaw := d.GetChange("tags")

		oTag := KeyValueTags(oTagRaw, d.Id(), TagResourceTypeGroup)
		oTags := KeyValueTags(oTagsRaw, d.Id(), TagResourceTypeGroup)
		oldTags := Tags(oTag.Merge(oTags))

		nTag := KeyValueTags(nTagRaw, d.Id(), TagResourceTypeGroup)
		nTags := KeyValueTags(nTagsRaw, d.Id(), TagResourceTypeGroup)
		newTags := Tags(nTag.Merge(nTags))

		if err := UpdateTags(conn, d.Id(), TagResourceTypeGroup, oldTags, newTags); err != nil {
			return fmt.Errorf("updating tags for Auto Scaling Group (%s): %w", d.Id(), err)
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

		if remove := flex.ExpandStringSet(os.Difference(ns)); len(remove) > 0 {
			// API only supports removing 10 at a time.
			batchSize := 10

			var batches [][]*string

			for batchSize < len(remove) {
				remove, batches = remove[batchSize:], append(batches, remove[0:batchSize:batchSize])
			}
			batches = append(batches, remove)

			for _, batch := range batches {
				_, err := conn.DetachLoadBalancers(&autoscaling.DetachLoadBalancersInput{
					AutoScalingGroupName: aws.String(d.Id()),
					LoadBalancerNames:    batch,
				})

				if err != nil {
					return fmt.Errorf("detaching Auto Scaling Group (%s) load balancers: %w", d.Id(), err)
				}

				if _, err := waitLoadBalancersRemoved(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("waiting for Auto Scaling Group (%s) load balancers removed: %s", d.Id(), err)
				}
			}
		}

		if add := flex.ExpandStringSet(ns.Difference(os)); len(add) > 0 {
			// API only supports adding 10 at a time.
			batchSize := 10

			var batches [][]*string

			for batchSize < len(add) {
				add, batches = add[batchSize:], append(batches, add[0:batchSize:batchSize])
			}
			batches = append(batches, add)

			for _, batch := range batches {
				_, err := conn.AttachLoadBalancers(&autoscaling.AttachLoadBalancersInput{
					AutoScalingGroupName: aws.String(d.Id()),
					LoadBalancerNames:    batch,
				})

				if err != nil {
					return fmt.Errorf("attaching Auto Scaling Group (%s) load balancers: %w", d.Id(), err)
				}

				if _, err := waitLoadBalancersAdded(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("waiting for Auto Scaling Group (%s) load balancers added: %s", d.Id(), err)
				}
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

		if remove := flex.ExpandStringSet(os.Difference(ns)); len(remove) > 0 {
			// AWS API only supports adding/removing 10 at a time.
			batchSize := 10

			var batches [][]*string

			for batchSize < len(remove) {
				remove, batches = remove[batchSize:], append(batches, remove[0:batchSize:batchSize])
			}
			batches = append(batches, remove)

			for _, batch := range batches {
				_, err := conn.DetachLoadBalancerTargetGroups(&autoscaling.DetachLoadBalancerTargetGroupsInput{
					AutoScalingGroupName: aws.String(d.Id()),
					TargetGroupARNs:      batch,
				})

				if err != nil {
					return fmt.Errorf("detaching Auto Scaling Group (%s) target groups: %w", d.Id(), err)
				}

				if _, err := waitLoadBalancerTargetGroupsRemoved(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("waiting for Auto Scaling Group (%s) target groups removed: %s", d.Id(), err)
				}
			}

		}

		if add := flex.ExpandStringSet(ns.Difference(os)); len(add) > 0 {
			// AWS API only supports adding/removing 10 at a time.
			batchSize := 10

			var batches [][]*string

			for batchSize < len(add) {
				add, batches = add[batchSize:], append(batches, add[0:batchSize:batchSize])
			}
			batches = append(batches, add)

			for _, batch := range batches {
				_, err := conn.AttachLoadBalancerTargetGroups(&autoscaling.AttachLoadBalancerTargetGroupsInput{
					AutoScalingGroupName: aws.String(d.Id()),
					TargetGroupARNs:      batch,
				})

				if err != nil {
					return fmt.Errorf("attaching Auto Scaling Group (%s) target groups: %w", d.Id(), err)
				}

				if _, err := waitLoadBalancerTargetGroupsAdded(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
					return fmt.Errorf("waiting for Auto Scaling Group (%s) target groups added: %s", d.Id(), err)
				}
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

				if v.Contains("tag") && !v.Contains("tags") {
					triggers = append(triggers, "tags") // nozero
				} else if !v.Contains("tag") && v.Contains("tags") {
					triggers = append(triggers, "tag") // nozero
				}

				shouldRefreshInstances = d.HasChanges(triggers...)
			}
		}

		if shouldRefreshInstances {
			if err := startInstanceRefresh(conn, expandStartInstanceRefreshInput(d.Id(), tfMap)); err != nil {
				return err
			}
		}
	}

	if d.HasChange("warm_pool") {
		w := d.Get("warm_pool").([]interface{})

		// No warm pool exists in new config. Delete it.
		if len(w) == 0 || w[0] == nil {
			forceDeleteWarmPool := d.Get("force_delete").(bool) || d.Get("force_delete_warm_pool").(bool)

			if err := deleteWarmPool(conn, d.Id(), forceDeleteWarmPool, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return err
			}
		} else {
			_, err := conn.PutWarmPool(expandPutWarmPoolInput(d.Id(), w[0].(map[string]interface{})))

			if err != nil {
				return fmt.Errorf("updating Auto Scaling Warm Pool (%s): %w", d.Id(), err)
			}
		}
	}

	if shouldWaitForCapacity {
		if err := waitForASGCapacity(d, meta, CapacitySatisfiedUpdate); err != nil {
			return fmt.Errorf("error waiting for Auto Scaling Group Capacity: %w", err)
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

			_, err := conn.DisableMetricsCollection(input)

			if err != nil {
				return fmt.Errorf("disabling Auto Scaling Group (%s) metrics collection: %w", d.Id(), err)
			}
		}

		if enableMetrics := ns.Difference(os); enableMetrics.Len() != 0 {
			input := &autoscaling.EnableMetricsCollectionInput{
				AutoScalingGroupName: aws.String(d.Id()),
				Granularity:          aws.String(d.Get("metrics_granularity").(string)),
				Metrics:              flex.ExpandStringSet(enableMetrics),
			}

			_, err := conn.EnableMetricsCollection(input)

			if err != nil {
				return fmt.Errorf("enabling Auto Scaling Group (%s) metrics collection: %w", d.Id(), err)
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

			_, err := conn.ResumeProcesses(input)

			if err != nil {
				return fmt.Errorf("resuming Auto Scaling Group (%s) scaling processes: %w", d.Id(), err)
			}
		}

		if suspendProcesses := ns.Difference(os); suspendProcesses.Len() != 0 {
			input := &autoscaling.ScalingProcessQuery{
				AutoScalingGroupName: aws.String(d.Id()),
				ScalingProcesses:     flex.ExpandStringSet(suspendProcesses),
			}

			_, err := conn.SuspendProcesses(input)

			if err != nil {
				return fmt.Errorf("suspending Auto Scaling Group (%s) scaling processes: %w", d.Id(), err)
			}
		}
	}

	return resourceGroupRead(d, meta)
}

func resourceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AutoScalingConn

	forceDeleteGroup := d.Get("force_delete").(bool)
	forceDeleteWarmPool := forceDeleteGroup || d.Get("force_delete_warm_pool").(bool)

	group, err := FindGroupByName(conn, d.Id())

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Auto Scaling Group (%s): %w", d.Id(), err)
	}

	if group.WarmPoolConfiguration != nil {
		err = deleteWarmPool(conn, d.Id(), forceDeleteWarmPool, d.Timeout(schema.TimeoutDelete))

		if err != nil {
			return err
		}
	}

	if !forceDeleteGroup {
		err = drainGroup(conn, d.Id(), group.Instances, d.Timeout(schema.TimeoutDelete))

		if err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Auto Scaling Group: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrCodeEquals(d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteAutoScalingGroup(&autoscaling.DeleteAutoScalingGroupInput{
				AutoScalingGroupName: aws.String(d.Id()),
				ForceDelete:          aws.Bool(forceDeleteGroup),
			})
		},
		autoscaling.ErrCodeResourceInUseFault, autoscaling.ErrCodeScalingActivityInProgressFault)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationError, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Auto Scaling Group (%s): %w", d.Id(), err)
	}

	_, err = tfresource.RetryUntilNotFound(d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return FindGroupByName(conn, d.Id())
		})

	if err != nil {
		return fmt.Errorf("waiting for Auto Scaling Group (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func drainGroup(conn *autoscaling.AutoScaling, name string, instances []*autoscaling.Instance, timeout time.Duration) error {
	input := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(name),
		DesiredCapacity:      aws.Int64(0),
		MinSize:              aws.Int64(0),
		MaxSize:              aws.Int64(0),
	}

	log.Printf("[DEBUG] Draining Auto Scaling Group: %s", name)
	if _, err := conn.UpdateAutoScalingGroup(input); err != nil {
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

		if _, err := conn.SetInstanceProtection(input); err != nil {
			return fmt.Errorf("disabling Auto Scaling Group (%s) scale-in protections: %w", name, err)
		}
	}

	if _, err := waitGroupDrained(conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Auto Scaling Group (%s) drain: %w", name, err)
	}

	return nil
}

func deleteWarmPool(conn *autoscaling.AutoScaling, name string, force bool, timeout time.Duration) error {
	if !force {
		if err := drainWarmPool(conn, name, timeout); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Deleting Auto Scaling Warm Pool: %s", name)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(timeout,
		func() (interface{}, error) {
			return conn.DeleteWarmPool(&autoscaling.DeleteWarmPoolInput{
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

	if _, err := waitWarmPoolDeleted(conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Auto Scaling Warm Pool (%s) delete: %w", name, err)
	}

	return nil
}

func drainWarmPool(conn *autoscaling.AutoScaling, name string, timeout time.Duration) error {
	input := &autoscaling.PutWarmPoolInput{
		AutoScalingGroupName:     aws.String(name),
		MaxGroupPreparedCapacity: aws.Int64(0),
		MinSize:                  aws.Int64(0),
	}

	log.Printf("[DEBUG] Draining Auto Scaling Warm Pool: %s", name)
	if _, err := conn.PutWarmPool(input); err != nil {
		return fmt.Errorf("setting Auto Scaling Warm Pool (%s) capacity to 0: %w", name, err)
	}

	if _, err := waitWarmPoolDrained(conn, name, timeout); err != nil {
		return fmt.Errorf("waiting for Auto Scaling Warm Pool (%s) drain: %w", name, err)
	}

	return nil
}

func findGroup(conn *autoscaling.AutoScaling, input *autoscaling.DescribeAutoScalingGroupsInput) (*autoscaling.Group, error) {
	output, err := findGroups(conn, input)

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

func findGroups(conn *autoscaling.AutoScaling, input *autoscaling.DescribeAutoScalingGroupsInput) ([]*autoscaling.Group, error) {
	var output []*autoscaling.Group

	err := conn.DescribeAutoScalingGroupsPages(input, func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
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

func FindGroupByName(conn *autoscaling.AutoScaling, name string) (*autoscaling.Group, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: aws.StringSlice([]string{name}),
	}

	output, err := findGroup(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.AutoScalingGroupName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findLoadBalancerStates(conn *autoscaling.AutoScaling, name string) ([]*autoscaling.LoadBalancerState, error) {
	input := &autoscaling.DescribeLoadBalancersInput{
		AutoScalingGroupName: aws.String(name),
	}
	var output []*autoscaling.LoadBalancerState

	err := describeLoadBalancersPages(conn, input, func(page *autoscaling.DescribeLoadBalancersOutput, lastPage bool) bool {
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
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findLoadBalancerTargetGroupStates(conn *autoscaling.AutoScaling, name string) ([]*autoscaling.LoadBalancerTargetGroupState, error) {
	input := &autoscaling.DescribeLoadBalancerTargetGroupsInput{
		AutoScalingGroupName: aws.String(name),
	}
	var output []*autoscaling.LoadBalancerTargetGroupState

	err := describeLoadBalancerTargetGroupsPages(conn, input, func(page *autoscaling.DescribeLoadBalancerTargetGroupsOutput, lastPage bool) bool {
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
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func findWarmPool(conn *autoscaling.AutoScaling, name string) (*autoscaling.DescribeWarmPoolOutput, error) {
	input := &autoscaling.DescribeWarmPoolInput{
		AutoScalingGroupName: aws.String(name),
	}
	var output *autoscaling.DescribeWarmPoolOutput

	err := describeWarmPoolPages(conn, input, func(page *autoscaling.DescribeWarmPoolOutput, lastPage bool) bool {
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
		return nil, &resource.NotFoundError{
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

func findInstanceRefresh(conn *autoscaling.AutoScaling, input *autoscaling.DescribeInstanceRefreshesInput) (*autoscaling.InstanceRefresh, error) {
	output, err := FindInstanceRefreshes(conn, input)

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

func FindInstanceRefreshes(conn *autoscaling.AutoScaling, input *autoscaling.DescribeInstanceRefreshesInput) ([]*autoscaling.InstanceRefresh, error) {
	var output []*autoscaling.InstanceRefresh

	err := describeInstanceRefreshesPages(conn, input, func(page *autoscaling.DescribeInstanceRefreshesOutput, lastPage bool) bool {
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
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusGroupInstanceCount(conn *autoscaling.AutoScaling, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindGroupByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.Itoa(len(output.Instances)), nil
	}
}

func statusInstanceRefresh(conn *autoscaling.AutoScaling, name, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &autoscaling.DescribeInstanceRefreshesInput{
			AutoScalingGroupName: aws.String(name),
			InstanceRefreshIds:   aws.StringSlice([]string{id}),
		}

		output, err := findInstanceRefresh(conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusLoadBalancerInStateCount(conn *autoscaling.AutoScaling, name string, states ...string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLoadBalancerStates(conn, name)

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

func statusLoadBalancerTargetGroupInStateCount(conn *autoscaling.AutoScaling, name string, states ...string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLoadBalancerTargetGroupStates(conn, name)

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

func statusWarmPool(conn *autoscaling.AutoScaling, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findWarmPool(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output.WarmPoolConfiguration, aws.StringValue(output.WarmPoolConfiguration.Status), nil
	}
}

func statusWarmPoolInstanceCount(conn *autoscaling.AutoScaling, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findWarmPool(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.Itoa(len(output.Instances)), nil
	}
}

func waitGroupDrained(conn *autoscaling.AutoScaling, name string, timeout time.Duration) (*autoscaling.Group, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusGroupInstanceCount(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*autoscaling.Group); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancersAdded(conn *autoscaling.AutoScaling, name string, timeout time.Duration) ([]*autoscaling.LoadBalancerState, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerInStateCount(conn, name, LoadBalancerStateAdding),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*autoscaling.LoadBalancerState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancersRemoved(conn *autoscaling.AutoScaling, name string, timeout time.Duration) ([]*autoscaling.LoadBalancerState, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerInStateCount(conn, name, LoadBalancerStateRemoving),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*autoscaling.LoadBalancerState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancerTargetGroupsAdded(conn *autoscaling.AutoScaling, name string, timeout time.Duration) ([]*autoscaling.LoadBalancerTargetGroupState, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerTargetGroupInStateCount(conn, name, LoadBalancerTargetGroupStateAdding),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*autoscaling.LoadBalancerTargetGroupState); ok {
		return output, err
	}

	return nil, err
}

func waitLoadBalancerTargetGroupsRemoved(conn *autoscaling.AutoScaling, name string, timeout time.Duration) ([]*autoscaling.LoadBalancerTargetGroupState, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusLoadBalancerTargetGroupInStateCount(conn, name, LoadBalancerTargetGroupStateRemoving),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*autoscaling.LoadBalancerTargetGroupState); ok {
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

func waitInstanceRefreshCancelled(conn *autoscaling.AutoScaling, name, id string, timeout time.Duration) (*autoscaling.InstanceRefresh, error) {
	stateConf := &resource.StateChangeConf{
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
		Refresh: statusInstanceRefresh(conn, name, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*autoscaling.InstanceRefresh); ok {
		return output, err
	}

	return nil, err
}

func waitWarmPoolDeleted(conn *autoscaling.AutoScaling, name string, timeout time.Duration) (*autoscaling.WarmPoolConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscaling.WarmPoolStatusPendingDelete},
		Target:  []string{},
		Refresh: statusWarmPool(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*autoscaling.WarmPoolConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitWarmPoolDrained(conn *autoscaling.AutoScaling, name string, timeout time.Duration) (*autoscaling.DescribeWarmPoolOutput, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"0"},
		Refresh: statusWarmPoolInstanceCount(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

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
	if v, ok := tfMap["launch_template_id"]; ok && v != "" {
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
	if v, ok := tfMap["id"]; ok && v != "" {
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

func expandStartInstanceRefreshInput(name string, tfMap map[string]interface{}) *autoscaling.StartInstanceRefreshInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &autoscaling.StartInstanceRefreshInput{
		AutoScalingGroupName: aws.String(name),
	}

	if v, ok := tfMap["preferences"].([]interface{}); ok && len(v) > 0 {
		apiObject.Preferences = expandRefreshPreferences(v[0].(map[string]interface{}))
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

	if v, ok := tfMap["skip_matching"].(bool); ok {
		apiObject.SkipMatching = aws.Bool(v)
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

func cancelInstanceRefresh(conn *autoscaling.AutoScaling, name string) error {
	input := &autoscaling.CancelInstanceRefreshInput{
		AutoScalingGroupName: aws.String(name),
	}

	output, err := conn.CancelInstanceRefresh(input)

	if tfawserr.ErrCodeEquals(err, autoscaling.ErrCodeActiveInstanceRefreshNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cancelling Auto Scaling Group (%s) instance refresh: %w", name, err)
	}

	_, err = waitInstanceRefreshCancelled(conn, name, aws.StringValue(output.InstanceRefreshId), instanceRefreshCancelledTimeout)

	if err != nil {
		return fmt.Errorf("waiting for Auto Scaling Group (%s) instance refresh cancel: %w", name, err)
	}

	return nil
}

func startInstanceRefresh(conn *autoscaling.AutoScaling, input *autoscaling.StartInstanceRefreshInput) error {
	name := aws.StringValue(input.AutoScalingGroupName)

	_, err := tfresource.RetryWhen(instanceRefreshStartedTimeout,
		func() (interface{}, error) {
			return conn.StartInstanceRefresh(input)
		},
		func(err error) (bool, error) {
			if tfawserr.ErrCodeEquals(err, autoscaling.ErrCodeInstanceRefreshInProgressFault) {
				if err := cancelInstanceRefresh(conn, name); err != nil {
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

	schema := ResourceGroup().Schema
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
