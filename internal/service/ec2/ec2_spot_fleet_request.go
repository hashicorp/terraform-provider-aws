package ec2

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSpotFleetRequest() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceSpotFleetRequestCreate,
		Read:   resourceSpotFleetRequestRead,
		Delete: resourceSpotFleetRequestDelete,
		Update: resourceSpotFleetRequestUpdate,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("instance_pools_to_use_count", 1)
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(15 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  SpotFleetRequestMigrateState,

		Schema: map[string]*schema.Schema{
			"allocation_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.AllocationStrategyLowestPrice,
				ValidateFunc: validation.StringInSlice(ec2.AllocationStrategy_Values(), false),
			},
			"client_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// Provided constants do not have the correct casing so going with hard-coded values.
			"excess_capacity_termination_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Default",
				ValidateFunc: validation.StringInSlice([]string{
					"Default",
					"NoTermination",
				}, false),
			},
			"fleet_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.FleetTypeMaintain,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.FleetType_Values(), false),
			},
			"iam_fleet_role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"instance_interruption_behaviour": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.InstanceInterruptionBehaviorTerminate,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(ec2.InstanceInterruptionBehavior_Values(), false),
			},
			"instance_pools_to_use_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
				ForceNew: true,
			},
			"launch_specification": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ami": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"associate_public_ip_address": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"ebs_block_device": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"device_name": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"iops": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"throughput": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"volume_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"volume_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(ec2.VolumeType_Values(), false),
									},
								},
							},
							Set: hashEbsBlockDevice,
						},
						"ebs_optimized": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"ephemeral_block_device": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"device_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"virtual_name": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
							Set: hashEphemeralBlockDevice,
						},
						"iam_instance_profile": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"iam_instance_profile_arn": {
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"key_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"monitoring": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"placement_group": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"placement_tenancy": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ec2.Tenancy_Values(), false),
						},
						"root_block_device": {
							// TODO: This is a set because we don't support singleton
							//       sub-resources today. We'll enforce that the set only ever has
							//       length zero or one below. When TF gains support for
							//       sub-resources this can be converted.
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								// "You can only modify the volume size, volume type, and Delete on
								// Termination flag on the block device mapping entry for the root
								// device volume." - bit.ly/ec2bdmap
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									"encrypted": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"iops": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"kms_key_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"throughput": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"volume_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"volume_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(ec2.VolumeType_Values(), false),
									},
								},
							},
							Set: hashRootBlockDevice,
						},
						"spot_price": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"tags": tftags.TagsSchemaForceNew(),
						"user_data": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							StateFunc: func(v interface{}) string {
								switch v := v.(type) {
								case string:
									return userDataHashSum(v)
								default:
									return ""
								}
							},
						},
						"vpc_security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"weighted_capacity": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
					},
				},
				Set:          hashLaunchSpecification,
				ExactlyOneOf: []string{"launch_specification", "launch_template_config"},
			},
			"launch_template_config": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"launch_template_specification": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidLaunchTemplateID,
									},
									"name": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidLaunchTemplateName,
									},
									"version": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
								},
							},
						},
						"overrides": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									"instance_requirements": {
										Type:     schema.TypeList,
										Optional: true,
										ForceNew: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"accelerator_count": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(0),
															},
															"min": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"accelerator_manufacturers": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.AcceleratorManufacturer_Values(), false),
													},
												},
												"accelerator_names": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.AcceleratorName_Values(), false),
													},
												},
												"accelerator_total_memory_mib": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"min": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"accelerator_types": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.AcceleratorType_Values(), false),
													},
												},
												"bare_metal": {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(ec2.BareMetal_Values(), false),
												},
												"baseline_ebs_bandwidth_mbps": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"min": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"burstable_performance": {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(ec2.BurstablePerformance_Values(), false),
												},
												"cpu_manufacturers": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.CpuManufacturer_Values(), false),
													},
												},
												"excluded_instance_types": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													MaxItems: 400,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"instance_generations": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.InstanceGeneration_Values(), false),
													},
												},
												"local_storage": {
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(ec2.LocalStorage_Values(), false),
												},
												"local_storage_types": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.LocalStorageType_Values(), false),
													},
												},
												"memory_gib_per_vcpu": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max": {
																Type:         schema.TypeFloat,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: verify.FloatGreaterThan(0.0),
															},
															"min": {
																Type:         schema.TypeFloat,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: verify.FloatGreaterThan(0.0),
															},
														},
													},
												},
												"memory_mib": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"min": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"network_interface_count": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"min": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"on_demand_max_price_percentage_over_lowest_price": {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"require_hibernate_support": {
													Type:     schema.TypeBool,
													Optional: true,
													ForceNew: true,
												},
												"spot_max_price_percentage_over_lowest_price": {
													Type:         schema.TypeInt,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"total_local_storage_gb": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max": {
																Type:         schema.TypeFloat,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: verify.FloatGreaterThan(0.0),
															},
															"min": {
																Type:         schema.TypeFloat,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: verify.FloatGreaterThan(0.0),
															},
														},
													},
												},
												"vcpu_count": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															"min": {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
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
										ForceNew: true,
									},
									"priority": {
										Type:     schema.TypeFloat,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"spot_price": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"weighted_capacity": {
										Type:     schema.TypeFloat,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
				ExactlyOneOf: []string{"launch_specification", "launch_template_config"},
			},
			"load_balancers": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"on_demand_allocation_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      ec2.OnDemandAllocationStrategyLowestPrice,
				ValidateFunc: validation.StringInSlice(ec2.OnDemandAllocationStrategy_Values(), false),
			},
			"on_demand_max_total_price": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"on_demand_target_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"replace_unhealthy_instances": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"spot_maintenance_strategies": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity_rebalance": {
							Type:             schema.TypeList,
							Optional:         true,
							MaxItems:         1,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"replacement_strategy": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(ec2.ReplacementStrategy_Values(), false),
									},
								},
							},
						},
					},
				},
			},
			"spot_price": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"spot_request_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"target_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"target_group_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"terminate_instances_on_delete": {
				Type:         nullable.TypeNullableBool,
				Optional:     true,
				ValidateFunc: nullable.ValidateTypeStringNullableBool,
			},
			"terminate_instances_with_expiration": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"valid_from": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"valid_until": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsRFC3339Time,
			},
			"wait_for_fulfillment": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSpotFleetRequestCreate(d *schema.ResourceData, meta interface{}) error {
	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_RequestSpotFleet.html
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	_, launchSpecificationOk := d.GetOk("launch_specification")

	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2.html#type-SpotFleetRequestConfigData
	spotFleetConfig := &ec2.SpotFleetRequestConfigData{
		IamFleetRole:                     aws.String(d.Get("iam_fleet_role").(string)),
		TargetCapacity:                   aws.Int64(int64(d.Get("target_capacity").(int))),
		ClientToken:                      aws.String(resource.UniqueId()),
		TerminateInstancesWithExpiration: aws.Bool(d.Get("terminate_instances_with_expiration").(bool)),
		ReplaceUnhealthyInstances:        aws.Bool(d.Get("replace_unhealthy_instances").(bool)),
		InstanceInterruptionBehavior:     aws.String(d.Get("instance_interruption_behaviour").(string)),
		Type:                             aws.String(d.Get("fleet_type").(string)),
		TagSpecifications:                ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeSpotFleetRequest),
	}

	if launchSpecificationOk {
		launchSpecs, err := buildSpotFleetLaunchSpecifications(d, meta)
		if err != nil {
			return err
		}
		spotFleetConfig.LaunchSpecifications = launchSpecs
	}

	if v, ok := d.GetOk("launch_template_config"); ok && v.(*schema.Set).Len() > 0 {
		spotFleetConfig.LaunchTemplateConfigs = expandLaunchTemplateConfigs(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("excess_capacity_termination_policy"); ok {
		spotFleetConfig.ExcessCapacityTerminationPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("allocation_strategy"); ok {
		spotFleetConfig.AllocationStrategy = aws.String(v.(string))
	} else {
		spotFleetConfig.AllocationStrategy = aws.String(ec2.AllocationStrategyLowestPrice)
	}

	if v, ok := d.GetOk("instance_pools_to_use_count"); ok && v.(int) != 1 {
		spotFleetConfig.InstancePoolsToUseCount = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("spot_maintenance_strategies"); ok {
		spotFleetConfig.SpotMaintenanceStrategies = expandSpotMaintenanceStrategies(v.([]interface{}))
	}

	// InvalidSpotFleetConfig: SpotMaintenanceStrategies option is only available with the spot fleet type maintain.
	if d.Get("fleet_type").(string) != ec2.FleetTypeMaintain {
		if spotFleetConfig.SpotMaintenanceStrategies != nil {
			log.Printf("[WARN] Spot Fleet (%s) has an invalid configuration and can not be requested. Capacity Rebalance maintenance strategies can only be specified for spot fleets of type maintain.", spotFleetConfig)
			return nil
		}
	}

	if v, ok := d.GetOk("spot_price"); ok {
		spotFleetConfig.SpotPrice = aws.String(v.(string))
	}

	spotFleetConfig.OnDemandTargetCapacity = aws.Int64(int64(d.Get("on_demand_target_capacity").(int)))

	if v, ok := d.GetOk("on_demand_allocation_strategy"); ok {
		spotFleetConfig.OnDemandAllocationStrategy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("on_demand_max_total_price"); ok {
		spotFleetConfig.OnDemandMaxTotalPrice = aws.String(v.(string))
	}

	if v, ok := d.GetOk("valid_from"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))

		spotFleetConfig.ValidFrom = aws.Time(v)
	}

	if v, ok := d.GetOk("valid_until"); ok {
		v, _ := time.Parse(time.RFC3339, v.(string))

		spotFleetConfig.ValidUntil = aws.Time(v)
	}

	if v, ok := d.GetOk("load_balancers"); ok && v.(*schema.Set).Len() > 0 {
		var elbNames []*ec2.ClassicLoadBalancer
		for _, v := range v.(*schema.Set).List() {
			elbNames = append(elbNames, &ec2.ClassicLoadBalancer{
				Name: aws.String(v.(string)),
			})
		}
		if spotFleetConfig.LoadBalancersConfig == nil {
			spotFleetConfig.LoadBalancersConfig = &ec2.LoadBalancersConfig{}
		}
		spotFleetConfig.LoadBalancersConfig.ClassicLoadBalancersConfig = &ec2.ClassicLoadBalancersConfig{
			ClassicLoadBalancers: elbNames,
		}
	}

	if v, ok := d.GetOk("target_group_arns"); ok && v.(*schema.Set).Len() > 0 {
		var targetGroups []*ec2.TargetGroup
		for _, v := range v.(*schema.Set).List() {
			targetGroups = append(targetGroups, &ec2.TargetGroup{
				Arn: aws.String(v.(string)),
			})
		}
		if spotFleetConfig.LoadBalancersConfig == nil {
			spotFleetConfig.LoadBalancersConfig = &ec2.LoadBalancersConfig{}
		}
		spotFleetConfig.LoadBalancersConfig.TargetGroupsConfig = &ec2.TargetGroupsConfig{
			TargetGroups: targetGroups,
		}
	}

	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2.html#type-RequestSpotFleetInput
	input := &ec2.RequestSpotFleetInput{
		SpotFleetRequestConfig: spotFleetConfig,
	}

	log.Printf("[DEBUG] Creating EC2 Spot Fleet Request: %s", input)
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(propagationTimeout,
		func() (interface{}, error) {
			return conn.RequestSpotFleet(input)
		},
		ErrCodeInvalidSpotFleetRequestConfig, "SpotFleetRequestConfig.IamFleetRole",
	)

	if err != nil {
		return fmt.Errorf("creating EC2 Spot Fleet Request: %w", err)
	}

	d.SetId(aws.StringValue(outputRaw.(*ec2.RequestSpotFleetOutput).SpotFleetRequestId))

	if _, err := WaitSpotFleetRequestCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for EC2 Spot Fleet Request (%s) create: %w", d.Id(), err)
	}

	if d.Get("wait_for_fulfillment").(bool) {
		if _, err := WaitSpotFleetRequestFulfilled(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return fmt.Errorf("waiting for EC2 Spot Fleet Request (%s) fulfillment: %w", d.Id(), err)
		}
	}

	return resourceSpotFleetRequestRead(d, meta)
}

func resourceSpotFleetRequestRead(d *schema.ResourceData, meta interface{}) error {
	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeSpotFleetRequests.html
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindSpotFleetRequestByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Spot Fleet Request %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Spot Fleet Request (%s): %w", d.Id(), err)
	}

	d.Set("spot_request_state", output.SpotFleetRequestState)

	config := output.SpotFleetRequestConfig

	if config.AllocationStrategy != nil {
		d.Set("allocation_strategy", config.AllocationStrategy)
	}

	if config.InstancePoolsToUseCount != nil {
		d.Set("instance_pools_to_use_count", config.InstancePoolsToUseCount)
	}

	if config.ClientToken != nil {
		d.Set("client_token", config.ClientToken)
	}

	if config.ExcessCapacityTerminationPolicy != nil {
		d.Set("excess_capacity_termination_policy", config.ExcessCapacityTerminationPolicy)
	}

	if config.IamFleetRole != nil {
		d.Set("iam_fleet_role", config.IamFleetRole)
	}

	d.Set("spot_maintenance_strategies", flattenSpotMaintenanceStrategies(config.SpotMaintenanceStrategies))

	if config.SpotPrice != nil {
		d.Set("spot_price", config.SpotPrice)
	}

	if config.TargetCapacity != nil {
		d.Set("target_capacity", config.TargetCapacity)
	}

	if config.TerminateInstancesWithExpiration != nil {
		d.Set("terminate_instances_with_expiration", config.TerminateInstancesWithExpiration)
	}

	if config.ValidFrom != nil {
		d.Set("valid_from",
			aws.TimeValue(config.ValidFrom).Format(time.RFC3339))
	}

	if config.ValidUntil != nil {
		d.Set("valid_until",
			aws.TimeValue(config.ValidUntil).Format(time.RFC3339))
	}

	launchSpec, err := launchSpecsToSet(conn, config.LaunchSpecifications)

	if err != nil {
		return fmt.Errorf("reading EC2 Spot Fleet Request (%s) launch specifications: %w", d.Id(), err)
	}

	d.Set("replace_unhealthy_instances", config.ReplaceUnhealthyInstances)
	d.Set("instance_interruption_behaviour", config.InstanceInterruptionBehavior)
	d.Set("fleet_type", config.Type)
	d.Set("launch_specification", launchSpec)

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	if err := d.Set("launch_template_config", flattenLaunchTemplateConfigs(config.LaunchTemplateConfigs)); err != nil {
		return fmt.Errorf("setting launch_template_config: %w", err)
	}

	d.Set("on_demand_target_capacity", config.OnDemandTargetCapacity)
	d.Set("on_demand_allocation_strategy", config.OnDemandAllocationStrategy)
	d.Set("on_demand_max_total_price", config.OnDemandMaxTotalPrice)

	if config.LoadBalancersConfig != nil {
		lbConf := config.LoadBalancersConfig

		if lbConf.ClassicLoadBalancersConfig != nil {
			flatLbs := make([]*string, 0)
			for _, lb := range lbConf.ClassicLoadBalancersConfig.ClassicLoadBalancers {
				flatLbs = append(flatLbs, lb.Name)
			}
			if err := d.Set("load_balancers", flex.FlattenStringSet(flatLbs)); err != nil {
				return fmt.Errorf("setting load_balancers: %w", err)
			}
		}

		if lbConf.TargetGroupsConfig != nil {
			flatTgs := make([]*string, 0)
			for _, tg := range lbConf.TargetGroupsConfig.TargetGroups {
				flatTgs = append(flatTgs, tg.Arn)
			}
			if err := d.Set("target_group_arns", flex.FlattenStringSet(flatTgs)); err != nil {
				return fmt.Errorf("setting target_group_arns: %w", err)
			}
		}
	}

	return nil
}

func resourceSpotFleetRequestUpdate(d *schema.ResourceData, meta interface{}) error {
	// http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySpotFleetRequest.html
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifySpotFleetRequestInput{
			SpotFleetRequestId: aws.String(d.Id()),
		}

		if d.HasChange("target_capacity") {
			input.TargetCapacity = aws.Int64(int64(d.Get("target_capacity").(int)))
		}

		if d.HasChange("on_demand_target_capacity") {
			input.OnDemandTargetCapacity = aws.Int64(int64(d.Get("on_demand_target_capacity").(int)))
		}

		if d.HasChange("excess_capacity_termination_policy") {
			if val, ok := d.GetOk("excess_capacity_termination_policy"); ok {
				input.ExcessCapacityTerminationPolicy = aws.String(val.(string))
			}
		}

		log.Printf("[DEBUG] Modifying EC2 Spot Fleet Request: %s", input)
		if _, err := conn.ModifySpotFleetRequest(input); err != nil {
			return fmt.Errorf("updating EC2 Spot Fleet Request (%s): %w", d.Id(), err)
		}

		if _, err := WaitSpotFleetRequestUpdated(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for EC2 Spot Fleet Request (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating tags: %w", err)
		}
	}

	return resourceSpotFleetRequestRead(d, meta)
}

func resourceSpotFleetRequestDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	terminateInstances := d.Get("terminate_instances_with_expiration").(bool)
	// If terminate_instances_on_delete is not null, its value is used.
	if v, null, _ := nullable.Bool(d.Get("terminate_instances_on_delete").(string)).Value(); !null {
		terminateInstances = v
	}

	log.Printf("[INFO] Deleting EC2 Spot Fleet Request: %s", d.Id())
	output, err := conn.CancelSpotFleetRequests(&ec2.CancelSpotFleetRequestsInput{
		SpotFleetRequestIds: aws.StringSlice([]string{d.Id()}),
		TerminateInstances:  aws.Bool(terminateInstances),
	})

	if err == nil && output != nil {
		err = CancelSpotFleetRequestsError(output.UnsuccessfulFleetRequests)
	}

	if tfawserr.ErrCodeEquals(err, ec2.CancelBatchErrorCodeFleetRequestIdDoesNotExist) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("cancelling EC2 Spot Fleet Request (%s): %w", d.Id(), err)
	}

	// Only wait for instance termination if requested.
	if !terminateInstances {
		return nil
	}

	_, err = tfresource.RetryUntilNotFound(d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		input := &ec2.DescribeSpotFleetInstancesInput{
			SpotFleetRequestId: aws.String(d.Id()),
		}
		output, err := FindSpotFleetInstances(conn, input)

		if err != nil {
			return nil, err
		}

		if len(output) == 0 {
			return nil, tfresource.NewEmptyResultError(input)
		}

		return output, nil
	})

	if err != nil {
		return fmt.Errorf("waiting for EC2 Spot Fleet Request (%s) active instance count to reach 0: %w", d.Id(), err)
	}

	return nil
}

func buildSpotFleetLaunchSpecification(d map[string]interface{}, meta interface{}) (*ec2.SpotFleetLaunchSpecification, error) {
	conn := meta.(*conns.AWSClient).EC2Conn

	opts := &ec2.SpotFleetLaunchSpecification{
		ImageId:      aws.String(d["ami"].(string)),
		InstanceType: aws.String(d["instance_type"].(string)),
		SpotPrice:    aws.String(d["spot_price"].(string)),
	}

	placement := new(ec2.SpotPlacement)
	if v, ok := d["availability_zone"]; ok {
		placement.AvailabilityZone = aws.String(v.(string))
		opts.Placement = placement
	}

	if v, ok := d["placement_tenancy"]; ok {
		placement.Tenancy = aws.String(v.(string))
		opts.Placement = placement
	}

	if v, ok := d["placement_group"]; ok {
		if v.(string) != "" {
			// If instanceInterruptionBehavior is set to STOP, this can't be set at all, even to an empty string, so check for "" to avoid those errors
			placement.GroupName = aws.String(v.(string))
			opts.Placement = placement
		}
	}

	if v, ok := d["ebs_optimized"]; ok {
		opts.EbsOptimized = aws.Bool(v.(bool))
	}

	if v, ok := d["monitoring"]; ok {
		opts.Monitoring = &ec2.SpotFleetMonitoring{
			Enabled: aws.Bool(v.(bool)),
		}
	}

	if v, ok := d["iam_instance_profile"]; ok {
		opts.IamInstanceProfile = &ec2.IamInstanceProfileSpecification{
			Name: aws.String(v.(string)),
		}
	}

	if v, ok := d["iam_instance_profile_arn"]; ok && v.(string) != "" {
		opts.IamInstanceProfile = &ec2.IamInstanceProfileSpecification{
			Arn: aws.String(v.(string)),
		}
	}

	if v, ok := d["user_data"]; ok {
		opts.UserData = aws.String(verify.Base64Encode([]byte(v.(string))))
	}

	if v, ok := d["key_name"]; ok && v != "" {
		opts.KeyName = aws.String(v.(string))
	}

	if v, ok := d["weighted_capacity"]; ok && v != "" {
		wc, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			return nil, err
		}
		opts.WeightedCapacity = aws.Float64(wc)
	}

	var securityGroupIds []*string
	if v, ok := d["vpc_security_group_ids"]; ok {
		if s := v.(*schema.Set); s.Len() > 0 {
			for _, v := range s.List() {
				securityGroupIds = append(securityGroupIds, aws.String(v.(string)))
			}
		}
	}

	if m, ok := d["tags"].(map[string]interface{}); ok && len(m) > 0 {
		tagsSpec := make([]*ec2.SpotFleetTagSpecification, 0)

		tags := Tags(tftags.New(m).IgnoreAWS())

		spec := &ec2.SpotFleetTagSpecification{
			ResourceType: aws.String(ec2.ResourceTypeInstance),
			Tags:         tags,
		}

		tagsSpec = append(tagsSpec, spec)

		opts.TagSpecifications = tagsSpec
	}

	subnetId, hasSubnetId := d["subnet_id"]
	if hasSubnetId {
		opts.SubnetId = aws.String(subnetId.(string))
	}

	associatePublicIpAddress, hasPublicIpAddress := d["associate_public_ip_address"]
	if hasPublicIpAddress && associatePublicIpAddress.(bool) && hasSubnetId {

		// If we have a non-default VPC / Subnet specified, we can flag
		// AssociatePublicIpAddress to get a Public IP assigned. By default these are not provided.
		// You cannot specify both SubnetId and the NetworkInterface.0.* parameters though, otherwise
		// you get: Network interfaces and an instance-level subnet ID may not be specified on the same request
		// You also need to attach Security Groups to the NetworkInterface instead of the instance,
		// to avoid: Network interfaces and an instance-level security groups may not be specified on
		// the same request
		ni := &ec2.InstanceNetworkInterfaceSpecification{
			AssociatePublicIpAddress: aws.Bool(true),
			DeleteOnTermination:      aws.Bool(true),
			DeviceIndex:              aws.Int64(0),
			SubnetId:                 aws.String(subnetId.(string)),
			Groups:                   securityGroupIds,
		}

		opts.NetworkInterfaces = []*ec2.InstanceNetworkInterfaceSpecification{ni}
		opts.SubnetId = aws.String("")
	} else {
		for _, id := range securityGroupIds {
			opts.SecurityGroups = append(opts.SecurityGroups, &ec2.GroupIdentifier{GroupId: id})
		}
	}

	blockDevices, err := readSpotFleetBlockDeviceMappingsFromConfig(d, conn)
	if err != nil {
		return nil, err
	}
	if len(blockDevices) > 0 {
		opts.BlockDeviceMappings = blockDevices
	}

	return opts, nil
}

func readSpotFleetBlockDeviceMappingsFromConfig(d map[string]interface{}, conn *ec2.EC2) ([]*ec2.BlockDeviceMapping, error) {
	blockDevices := make([]*ec2.BlockDeviceMapping, 0)

	if v, ok := d["ebs_block_device"]; ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd["delete_on_termination"].(bool)),
			}

			if v, ok := bd["snapshot_id"].(string); ok && v != "" {
				ebs.SnapshotId = aws.String(v)
			}

			if v, ok := bd["encrypted"].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd["kms_key_id"].(string); ok && v != "" {
				ebs.KmsKeyId = aws.String(v)
			}

			if v, ok := bd["volume_size"].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int64(int64(v))
			}

			if v, ok := bd["volume_type"].(string); ok && v != "" {
				ebs.VolumeType = aws.String(v)
			}

			if v, ok := bd["iops"].(int); ok && v > 0 {
				ebs.Iops = aws.Int64(int64(v))
			}

			if v, ok := bd["throughput"].(int); ok && v > 0 {
				ebs.Throughput = aws.Int64(int64(v))
			}

			blockDevices = append(blockDevices, &ec2.BlockDeviceMapping{
				DeviceName: aws.String(bd["device_name"].(string)),
				Ebs:        ebs,
			})
		}
	}

	if v, ok := d["ephemeral_block_device"]; ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			blockDevices = append(blockDevices, &ec2.BlockDeviceMapping{
				DeviceName:  aws.String(bd["device_name"].(string)),
				VirtualName: aws.String(bd["virtual_name"].(string)),
			})
		}
	}

	if v, ok := d["root_block_device"]; ok {
		vL := v.(*schema.Set).List()
		if len(vL) > 1 {
			return nil, fmt.Errorf("Cannot specify more than one root_block_device.")
		}
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &ec2.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd["delete_on_termination"].(bool)),
			}

			if v, ok := bd["encrypted"].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd["kms_key_id"].(string); ok && v != "" {
				ebs.KmsKeyId = aws.String(v)
			}

			if v, ok := bd["volume_size"].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int64(int64(v))
			}

			if v, ok := bd["volume_type"].(string); ok && v != "" {
				ebs.VolumeType = aws.String(v)
			}

			if v, ok := bd["iops"].(int); ok && v > 0 {
				ebs.Iops = aws.Int64(int64(v))
			}

			if v, ok := bd["throughput"].(int); ok && v > 0 {
				ebs.Throughput = aws.Int64(int64(v))
			}

			if dn, err := FetchRootDeviceName(conn, d["ami"].(string)); err == nil {
				if dn == nil {
					return nil, fmt.Errorf(
						"Expected 1 AMI for ID: %s, got none",
						d["ami"].(string))
				}

				blockDevices = append(blockDevices, &ec2.BlockDeviceMapping{
					DeviceName: dn,
					Ebs:        ebs,
				})
			} else {
				return nil, err
			}
		}
	}

	return blockDevices, nil
}

func buildSpotFleetLaunchSpecifications(d *schema.ResourceData, meta interface{}) ([]*ec2.SpotFleetLaunchSpecification, error) {
	userSpecs := d.Get("launch_specification").(*schema.Set).List()
	specs := make([]*ec2.SpotFleetLaunchSpecification, len(userSpecs))
	for i, userSpec := range userSpecs {
		userSpecMap := userSpec.(map[string]interface{})
		// panic: interface conversion: interface {} is map[string]interface {}, not *schema.ResourceData
		opts, err := buildSpotFleetLaunchSpecification(userSpecMap, meta)
		if err != nil {
			return nil, err
		}
		specs[i] = opts
	}

	return specs, nil
}

func expandLaunchTemplateConfig(tfMap map[string]interface{}) *ec2.LaunchTemplateConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateConfig{}

	if v, ok := tfMap["launch_template_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplateSpecification = expandFleetLaunchTemplateSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["overrides"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Overrides = expandLaunchTemplateOverrideses(v.List())
	}

	return apiObject
}

func expandLaunchTemplateConfigs(tfList []interface{}) []*ec2.LaunchTemplateConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateConfig(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandFleetLaunchTemplateSpecification(tfMap map[string]interface{}) *ec2.FleetLaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.FleetLaunchTemplateSpecification{}

	if v, ok := tfMap["id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateOverrides(tfMap map[string]interface{}) *ec2.LaunchTemplateOverrides {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateOverrides{}

	if v, ok := tfMap["availability_zone"].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["instance_requirements"].([]interface{}); ok && len(v) > 0 {
		apiObject.InstanceRequirements = expandInstanceRequirements(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["instance_type"].(string); ok && v != "" {
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := tfMap["priority"].(float64); ok && v != 0.0 {
		apiObject.Priority = aws.Float64(v)
	}

	if v, ok := tfMap["spot_price"].(string); ok && v != "" {
		apiObject.SpotPrice = aws.String(v)
	}

	if v, ok := tfMap["subnet_id"].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	if v, ok := tfMap["weighted_capacity"].(float64); ok && v != 0.0 {
		apiObject.WeightedCapacity = aws.Float64(v)
	}

	return apiObject
}

func expandLaunchTemplateOverrideses(tfList []interface{}) []*ec2.LaunchTemplateOverrides {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateOverrides

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

func expandInstanceRequirements(tfMap map[string]interface{}) *ec2.InstanceRequirements {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.InstanceRequirements{}

	if v, ok := tfMap["accelerator_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.AcceleratorCount = expandAcceleratorCount(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["accelerator_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorManufacturers = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["accelerator_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorNames = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["accelerator_total_memory_mib"].([]interface{}); ok && len(v) > 0 {
		apiObject.AcceleratorTotalMemoryMiB = expandAcceleratorTotalMemoryMiB(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["accelerator_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorTypes = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["bare_metal"].(string); ok && v != "" {
		apiObject.BareMetal = aws.String(v)
	}

	if v, ok := tfMap["baseline_ebs_bandwidth_mbps"].([]interface{}); ok && len(v) > 0 {
		apiObject.BaselineEbsBandwidthMbps = expandBaselineEbsBandwidthMbps(v[0].(map[string]interface{}))
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
		apiObject.MemoryGiBPerVCpu = expandMemoryGiBPerVCpu(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["memory_mib"].([]interface{}); ok && len(v) > 0 {
		apiObject.MemoryMiB = expandMemoryMiB(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["network_interface_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.NetworkInterfaceCount = expandNetworkInterfaceCount(v[0].(map[string]interface{}))
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
		apiObject.TotalLocalStorageGB = expandTotalLocalStorageGB(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["vcpu_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.VCpuCount = expandVCpuCountRange(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAcceleratorCount(tfMap map[string]interface{}) *ec2.AcceleratorCount {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.AcceleratorCount{}

	if v, ok := tfMap["max"].(int); ok {
		apiObject.Max = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min"].(int); ok {
		apiObject.Min = aws.Int64(int64(v))
	}

	return apiObject
}

func expandAcceleratorTotalMemoryMiB(tfMap map[string]interface{}) *ec2.AcceleratorTotalMemoryMiB {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.AcceleratorTotalMemoryMiB{}

	if v, ok := tfMap["max"].(int); ok {
		apiObject.Max = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min"].(int); ok {
		apiObject.Min = aws.Int64(int64(v))
	}

	return apiObject
}

func expandBaselineEbsBandwidthMbps(tfMap map[string]interface{}) *ec2.BaselineEbsBandwidthMbps {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.BaselineEbsBandwidthMbps{}

	if v, ok := tfMap["max"].(int); ok {
		apiObject.Max = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min"].(int); ok {
		apiObject.Min = aws.Int64(int64(v))
	}

	return apiObject
}

func expandMemoryGiBPerVCpu(tfMap map[string]interface{}) *ec2.MemoryGiBPerVCpu {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.MemoryGiBPerVCpu{}

	if v, ok := tfMap["max"].(float64); ok {
		apiObject.Max = aws.Float64(v)
	}

	if v, ok := tfMap["min"].(float64); ok {
		apiObject.Min = aws.Float64(v)
	}

	return apiObject
}

func expandMemoryMiB(tfMap map[string]interface{}) *ec2.MemoryMiB {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.MemoryMiB{}

	if v, ok := tfMap["max"].(int); ok {
		apiObject.Max = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min"].(int); ok {
		apiObject.Min = aws.Int64(int64(v))
	}

	return apiObject
}

func expandNetworkInterfaceCount(tfMap map[string]interface{}) *ec2.NetworkInterfaceCount {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.NetworkInterfaceCount{}

	if v, ok := tfMap["max"].(int); ok {
		apiObject.Max = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min"].(int); ok {
		apiObject.Min = aws.Int64(int64(v))
	}

	return apiObject
}

func expandTotalLocalStorageGB(tfMap map[string]interface{}) *ec2.TotalLocalStorageGB {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.TotalLocalStorageGB{}

	if v, ok := tfMap["max"].(float64); ok {
		apiObject.Max = aws.Float64(v)
	}

	if v, ok := tfMap["min"].(float64); ok {
		apiObject.Min = aws.Float64(v)
	}

	return apiObject
}

func expandVCpuCountRange(tfMap map[string]interface{}) *ec2.VCpuCountRange {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.VCpuCountRange{}

	if v, ok := tfMap["max"].(int); ok {
		apiObject.Max = aws.Int64(int64(v))
	}

	if v, ok := tfMap["min"].(int); ok {
		apiObject.Min = aws.Int64(int64(v))
	}

	return apiObject
}

func expandSpotMaintenanceStrategies(l []interface{}) *ec2.SpotMaintenanceStrategies {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	fleetSpotMaintenanceStrategies := &ec2.SpotMaintenanceStrategies{
		CapacityRebalance: expandSpotCapacityRebalance(m["capacity_rebalance"].([]interface{})),
	}

	return fleetSpotMaintenanceStrategies
}

func expandSpotCapacityRebalance(l []interface{}) *ec2.SpotCapacityRebalance {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	capacityRebalance := &ec2.SpotCapacityRebalance{}

	if v, ok := m["replacement_strategy"]; ok && v.(string) != "" {
		capacityRebalance.ReplacementStrategy = aws.String(v.(string))
	}

	return capacityRebalance
}

func launchSpecsToSet(conn *ec2.EC2, launchSpecs []*ec2.SpotFleetLaunchSpecification) (*schema.Set, error) {
	specSet := &schema.Set{F: hashLaunchSpecification}
	for _, spec := range launchSpecs {
		rootDeviceName, err := FetchRootDeviceName(conn, aws.StringValue(spec.ImageId))
		if err != nil {
			return nil, err
		}

		specSet.Add(launchSpecToMap(spec, rootDeviceName))
	}
	return specSet, nil
}

func launchSpecToMap(l *ec2.SpotFleetLaunchSpecification, rootDevName *string) map[string]interface{} {
	m := make(map[string]interface{})

	m["root_block_device"] = rootBlockDeviceToSet(l.BlockDeviceMappings, rootDevName)
	m["ebs_block_device"] = ebsBlockDevicesToSet(l.BlockDeviceMappings, rootDevName)
	m["ephemeral_block_device"] = ephemeralBlockDevicesToSet(l.BlockDeviceMappings)

	if l.ImageId != nil {
		m["ami"] = aws.StringValue(l.ImageId)
	}

	if l.InstanceType != nil {
		m["instance_type"] = aws.StringValue(l.InstanceType)
	}

	if l.SpotPrice != nil {
		m["spot_price"] = aws.StringValue(l.SpotPrice)
	}

	if l.EbsOptimized != nil {
		m["ebs_optimized"] = aws.BoolValue(l.EbsOptimized)
	}

	if l.Monitoring != nil && l.Monitoring.Enabled != nil {
		m["monitoring"] = aws.BoolValue(l.Monitoring.Enabled)
	}

	if l.IamInstanceProfile != nil && l.IamInstanceProfile.Name != nil {
		m["iam_instance_profile"] = aws.StringValue(l.IamInstanceProfile.Name)
	}

	if l.IamInstanceProfile != nil && l.IamInstanceProfile.Arn != nil {
		m["iam_instance_profile_arn"] = aws.StringValue(l.IamInstanceProfile.Arn)
	}

	if l.UserData != nil {
		m["user_data"] = userDataHashSum(aws.StringValue(l.UserData))
	}

	if l.KeyName != nil {
		m["key_name"] = aws.StringValue(l.KeyName)
	}

	if l.Placement != nil {
		m["availability_zone"] = aws.StringValue(l.Placement.AvailabilityZone)
	}

	if l.SubnetId != nil {
		m["subnet_id"] = aws.StringValue(l.SubnetId)
	}

	securityGroupIds := &schema.Set{F: schema.HashString}
	if len(l.NetworkInterfaces) > 0 {
		m["associate_public_ip_address"] = aws.BoolValue(l.NetworkInterfaces[0].AssociatePublicIpAddress)
		m["subnet_id"] = aws.StringValue(l.NetworkInterfaces[0].SubnetId)

		for _, group := range l.NetworkInterfaces[0].Groups {
			securityGroupIds.Add(aws.StringValue(group))
		}
	} else {
		for _, group := range l.SecurityGroups {
			securityGroupIds.Add(aws.StringValue(group.GroupId))
		}
	}
	m["vpc_security_group_ids"] = securityGroupIds

	if l.WeightedCapacity != nil {
		m["weighted_capacity"] = strconv.FormatFloat(*l.WeightedCapacity, 'f', 0, 64)
	}

	if l.TagSpecifications != nil {
		for _, tagSpecs := range l.TagSpecifications {
			// only "instance" tags are currently supported: http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SpotFleetTagSpecification.html
			if aws.StringValue(tagSpecs.ResourceType) == ec2.ResourceTypeInstance {
				m["tags"] = KeyValueTags(tagSpecs.Tags).IgnoreAWS().Map()
			}
		}
	}

	return m
}

func ebsBlockDevicesToSet(bdm []*ec2.BlockDeviceMapping, rootDevName *string) *schema.Set {
	set := &schema.Set{F: hashEbsBlockDevice}

	for _, val := range bdm {
		if val.Ebs != nil {
			m := make(map[string]interface{})

			ebs := val.Ebs

			if val.DeviceName != nil {
				if aws.StringValue(rootDevName) == aws.StringValue(val.DeviceName) {
					continue
				}

				m["device_name"] = aws.StringValue(val.DeviceName)
			}

			if ebs.DeleteOnTermination != nil {
				m["delete_on_termination"] = aws.BoolValue(ebs.DeleteOnTermination)
			}

			if ebs.SnapshotId != nil {
				m["snapshot_id"] = aws.StringValue(ebs.SnapshotId)
			}

			if ebs.Encrypted != nil {
				m["encrypted"] = aws.BoolValue(ebs.Encrypted)
			}

			if ebs.KmsKeyId != nil {
				m["kms_key_id"] = aws.StringValue(ebs.KmsKeyId)
			}

			if ebs.VolumeSize != nil {
				m["volume_size"] = aws.Int64Value(ebs.VolumeSize)
			}

			if ebs.VolumeType != nil {
				m["volume_type"] = aws.StringValue(ebs.VolumeType)
			}

			if ebs.Iops != nil {
				m["iops"] = aws.Int64Value(ebs.Iops)
			}

			if ebs.Throughput != nil {
				m["throughput"] = aws.Int64Value(ebs.Throughput)
			}

			set.Add(m)
		}
	}

	return set
}

func ephemeralBlockDevicesToSet(bdm []*ec2.BlockDeviceMapping) *schema.Set {
	set := &schema.Set{F: hashEphemeralBlockDevice}

	for _, val := range bdm {
		if val.VirtualName != nil {
			m := make(map[string]interface{})
			m["virtual_name"] = aws.StringValue(val.VirtualName)

			if val.DeviceName != nil {
				m["device_name"] = aws.StringValue(val.DeviceName)
			}

			set.Add(m)
		}
	}

	return set
}

func rootBlockDeviceToSet(bdm []*ec2.BlockDeviceMapping, rootDevName *string) *schema.Set {
	set := &schema.Set{F: hashRootBlockDevice}

	if rootDevName != nil {
		for _, val := range bdm {
			if aws.StringValue(val.DeviceName) == aws.StringValue(rootDevName) {
				m := make(map[string]interface{})
				if val.Ebs.DeleteOnTermination != nil {
					m["delete_on_termination"] = aws.BoolValue(val.Ebs.DeleteOnTermination)
				}

				if val.Ebs.Encrypted != nil {
					m["encrypted"] = aws.BoolValue(val.Ebs.Encrypted)
				}

				if val.Ebs.KmsKeyId != nil {
					m["kms_key_id"] = aws.StringValue(val.Ebs.KmsKeyId)
				}

				if val.Ebs.VolumeSize != nil {
					m["volume_size"] = aws.Int64Value(val.Ebs.VolumeSize)
				}

				if val.Ebs.VolumeType != nil {
					m["volume_type"] = aws.StringValue(val.Ebs.VolumeType)
				}

				if val.Ebs.Iops != nil {
					m["iops"] = aws.Int64Value(val.Ebs.Iops)
				}

				if val.Ebs.Throughput != nil {
					m["throughput"] = aws.Int64Value(val.Ebs.Throughput)
				}

				set.Add(m)
			}
		}
	}

	return set
}

func hashEphemeralBlockDevice(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["device_name"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["virtual_name"].(string)))
	return create.StringHashcode(buf.String())
}

func hashRootBlockDevice(v interface{}) int {
	// there can be only one root device; no need to hash anything
	return 0
}

func hashLaunchSpecification(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["ami"].(string)))
	if v, ok := m["availability_zone"].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := m["subnet_id"].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["spot_price"].(string)))
	return create.StringHashcode(buf.String())
}

func hashEbsBlockDevice(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if name, ok := m["device_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", name.(string)))
	}
	if id, ok := m["snapshot_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", id.(string)))
	}
	return create.StringHashcode(buf.String())
}

func flattenLaunchTemplateConfig(apiObject *ec2.LaunchTemplateConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateSpecification; v != nil {
		tfMap["launch_template_specification"] = []interface{}{flattenFleetLaunchTemplateSpecification(v)}
	}

	if v := apiObject.Overrides; v != nil {
		tfMap["overrides"] = flattenLaunchTemplateOverrideses(v)
	}

	return tfMap
}

func flattenLaunchTemplateConfigs(apiObjects []*ec2.LaunchTemplateConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchTemplateConfig(apiObject))
	}

	return tfList
}

func flattenFleetLaunchTemplateSpecification(apiObject *ec2.FleetLaunchTemplateSpecification) map[string]interface{} {
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

func flattenLaunchTemplateOverrides(apiObject *ec2.LaunchTemplateOverrides) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap["availability_zone"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceRequirements; v != nil {
		tfMap["instance_requirements"] = []interface{}{flattenInstanceRequirements(v)}
	}

	if v := apiObject.InstanceType; v != nil {
		tfMap["instance_type"] = aws.StringValue(v)
	}

	if v := apiObject.Priority; v != nil {
		tfMap["priority"] = aws.Float64Value(v)
	}

	if v := apiObject.SpotPrice; v != nil {
		tfMap["spot_price"] = aws.StringValue(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap["subnet_id"] = aws.StringValue(v)
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.Float64Value(v)
	}

	return tfMap
}

func flattenLaunchTemplateOverrideses(apiObjects []*ec2.LaunchTemplateOverrides) []interface{} {
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

func flattenSpotMaintenanceStrategies(spotMaintenanceStrategies *ec2.SpotMaintenanceStrategies) []interface{} {
	if spotMaintenanceStrategies == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"capacity_rebalance": flattenSpotCapacityRebalance(spotMaintenanceStrategies.CapacityRebalance),
	}

	return []interface{}{m}
}

func flattenSpotCapacityRebalance(spotCapacityRebalance *ec2.SpotCapacityRebalance) []interface{} {
	if spotCapacityRebalance == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"replacement_strategy": aws.StringValue(spotCapacityRebalance.ReplacementStrategy),
	}

	return []interface{}{m}
}
