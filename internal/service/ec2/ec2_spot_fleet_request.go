// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_spot_fleet_request", name="Spot Fleet Request")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceSpotFleetRequest() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceSpotFleetRequestCreate,
		ReadWithoutTimeout:   resourceSpotFleetRequestRead,
		DeleteWithoutTimeout: resourceSpotFleetRequestDelete,
		UpdateWithoutTimeout: resourceSpotFleetRequestUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
		MigrateState:  spotFleetRequestMigrateState,

		Schema: map[string]*schema.Schema{
			"allocation_strategy": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.AllocationStrategyLowestPrice,
				ValidateDiagFunc: enum.Validate[awstypes.AllocationStrategy](),
			},
			"client_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"context": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.FleetTypeMaintain,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.FleetType](),
			},
			"iam_fleet_role": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"instance_interruption_behaviour": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.InstanceInterruptionBehaviorTerminate,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.InstanceInterruptionBehavior](),
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
						names.AttrAvailabilityZone: {
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
									names.AttrDeleteOnTermination: {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									names.AttrDeviceName: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									names.AttrEncrypted: {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrSnapshotID: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrThroughput: {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrVolumeSize: {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrVolumeType: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.VolumeType](),
									},
								},
							},
							Set: hashEBSBlockDevice,
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
									names.AttrDeviceName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrVirtualName: {
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
						names.AttrInstanceType: {
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
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Tenancy](),
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
									names.AttrDeleteOnTermination: {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										ForceNew: true,
									},
									names.AttrEncrypted: {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrThroughput: {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrVolumeSize: {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									names.AttrVolumeType: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.VolumeType](),
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
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						names.AttrTags: tftags.TagsSchemaForceNew(),
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
						names.AttrVPCSecurityGroupIDs: {
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
									names.AttrID: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidLaunchTemplateID,
									},
									names.AttrName: {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: verify.ValidLaunchTemplateName,
									},
									names.AttrVersion: {
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
									names.AttrAvailabilityZone: {
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
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(0),
															},
															names.AttrMin: {
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
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.AcceleratorManufacturer](),
													},
												},
												"accelerator_names": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.AcceleratorName](),
													},
												},
												"accelerator_total_memory_mib": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															names.AttrMin: {
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
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.AcceleratorType](),
													},
												},
												"allowed_instance_types": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													MaxItems: 400,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"bare_metal": {
													Type:             schema.TypeString,
													Optional:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.BareMetal](),
												},
												"baseline_ebs_bandwidth_mbps": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															names.AttrMin: {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"burstable_performance": {
													Type:             schema.TypeString,
													Optional:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.BurstablePerformance](),
												},
												"cpu_manufacturers": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.CpuManufacturer](),
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
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.InstanceGeneration](),
													},
												},
												"local_storage": {
													Type:             schema.TypeString,
													Optional:         true,
													ForceNew:         true,
													ValidateDiagFunc: enum.Validate[awstypes.LocalStorage](),
												},
												"local_storage_types": {
													Type:     schema.TypeSet,
													Optional: true,
													ForceNew: true,
													Elem: &schema.Schema{
														Type:             schema.TypeString,
														ValidateDiagFunc: enum.Validate[awstypes.LocalStorageType](),
													},
												},
												"memory_gib_per_vcpu": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: verify.FloatGreaterThan(0.0),
															},
															names.AttrMin: {
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
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															names.AttrMin: {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
												},
												"network_bandwidth_gbps": {
													Type:     schema.TypeList,
													Optional: true,
													ForceNew: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrMax: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: verify.FloatGreaterThan(0.0),
															},
															names.AttrMin: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: verify.FloatGreaterThan(0.0),
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
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															names.AttrMin: {
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
															names.AttrMax: {
																Type:         schema.TypeFloat,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: verify.FloatGreaterThan(0.0),
															},
															names.AttrMin: {
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
															names.AttrMax: {
																Type:         schema.TypeInt,
																Optional:     true,
																ForceNew:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
															names.AttrMin: {
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
									names.AttrInstanceType: {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									names.AttrPriority: {
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
									names.AttrSubnetID: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.OnDemandAllocationStrategyLowestPrice,
				ValidateDiagFunc: enum.Validate[awstypes.OnDemandAllocationStrategy](),
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
										Type:             schema.TypeString,
										Optional:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ReplacementStrategy](),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_capacity": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"target_capacity_unit_type": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TargetCapacityUnitType](),
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

func resourceSpotFleetRequestCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	_, launchSpecificationOk := d.GetOk("launch_specification")

	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2.html#type-SpotFleetRequestConfigData
	spotFleetConfig := &awstypes.SpotFleetRequestConfigData{
		ClientToken:                      aws.String(id.UniqueId()),
		IamFleetRole:                     aws.String(d.Get("iam_fleet_role").(string)),
		InstanceInterruptionBehavior:     awstypes.InstanceInterruptionBehavior(d.Get("instance_interruption_behaviour").(string)),
		ReplaceUnhealthyInstances:        aws.Bool(d.Get("replace_unhealthy_instances").(bool)),
		TagSpecifications:                getTagSpecificationsIn(ctx, awstypes.ResourceTypeSpotFleetRequest),
		TargetCapacity:                   aws.Int32(int32(d.Get("target_capacity").(int))),
		TerminateInstancesWithExpiration: aws.Bool(d.Get("terminate_instances_with_expiration").(bool)),
		Type:                             awstypes.FleetType(d.Get("fleet_type").(string)),
	}

	if v, ok := d.GetOk("context"); ok {
		spotFleetConfig.Context = aws.String(v.(string))
	}

	if launchSpecificationOk {
		launchSpecs, err := buildSpotFleetLaunchSpecifications(ctx, d, meta)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Spot Fleet Request: %s", err)
		}
		spotFleetConfig.LaunchSpecifications = launchSpecs
	}

	if v, ok := d.GetOk("launch_template_config"); ok && v.(*schema.Set).Len() > 0 {
		spotFleetConfig.LaunchTemplateConfigs = expandLaunchTemplateConfigs(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("excess_capacity_termination_policy"); ok {
		spotFleetConfig.ExcessCapacityTerminationPolicy = awstypes.ExcessCapacityTerminationPolicy(v.(string))
	}

	if v, ok := d.GetOk("allocation_strategy"); ok {
		spotFleetConfig.AllocationStrategy = awstypes.AllocationStrategy(v.(string))
	} else {
		spotFleetConfig.AllocationStrategy = awstypes.AllocationStrategyLowestPrice
	}

	if v, ok := d.GetOk("instance_pools_to_use_count"); ok && v.(int) != 1 {
		spotFleetConfig.InstancePoolsToUseCount = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("spot_maintenance_strategies"); ok {
		spotFleetConfig.SpotMaintenanceStrategies = expandSpotMaintenanceStrategies(v.([]interface{}))
	}

	// InvalidSpotFleetConfig: SpotMaintenanceStrategies option is only available with the spot fleet type maintain.
	if d.Get("fleet_type").(string) != string(awstypes.FleetTypeMaintain) {
		if spotFleetConfig.SpotMaintenanceStrategies != nil {
			log.Printf("[WARN] Spot Fleet (%s) has an invalid configuration and can not be requested. Capacity Rebalance maintenance strategies can only be specified for spot fleets of type maintain.", d.Id())
			return diags
		}
	}

	if v, ok := d.GetOk("spot_price"); ok {
		spotFleetConfig.SpotPrice = aws.String(v.(string))
	}

	spotFleetConfig.OnDemandTargetCapacity = aws.Int32(int32(d.Get("on_demand_target_capacity").(int)))

	if v, ok := d.GetOk("on_demand_allocation_strategy"); ok {
		spotFleetConfig.OnDemandAllocationStrategy = awstypes.OnDemandAllocationStrategy(v.(string))
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
		var elbNames []awstypes.ClassicLoadBalancer
		for _, v := range v.(*schema.Set).List() {
			elbNames = append(elbNames, awstypes.ClassicLoadBalancer{
				Name: aws.String(v.(string)),
			})
		}
		if spotFleetConfig.LoadBalancersConfig == nil {
			spotFleetConfig.LoadBalancersConfig = &awstypes.LoadBalancersConfig{}
		}
		spotFleetConfig.LoadBalancersConfig.ClassicLoadBalancersConfig = &awstypes.ClassicLoadBalancersConfig{
			ClassicLoadBalancers: elbNames,
		}
	}

	if v, ok := d.GetOk("target_group_arns"); ok && v.(*schema.Set).Len() > 0 {
		var targetGroups []awstypes.TargetGroup
		for _, v := range v.(*schema.Set).List() {
			targetGroups = append(targetGroups, awstypes.TargetGroup{
				Arn: aws.String(v.(string)),
			})
		}
		if spotFleetConfig.LoadBalancersConfig == nil {
			spotFleetConfig.LoadBalancersConfig = &awstypes.LoadBalancersConfig{}
		}
		spotFleetConfig.LoadBalancersConfig.TargetGroupsConfig = &awstypes.TargetGroupsConfig{
			TargetGroups: targetGroups,
		}
	}

	if v, ok := d.GetOk("target_capacity_unit_type"); ok {
		spotFleetConfig.TargetCapacityUnitType = awstypes.TargetCapacityUnitType(v.(string))
	}

	// http://docs.aws.amazon.com/sdk-for-go/api/service/ec2.html#type-RequestSpotFleetInput
	input := &ec2.RequestSpotFleetInput{
		SpotFleetRequestConfig: spotFleetConfig,
	}

	log.Printf("[DEBUG] Creating EC2 Spot Fleet Request: %s", d.Id())
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, iamPropagationTimeout,
		func() (interface{}, error) {
			return conn.RequestSpotFleet(ctx, input)
		},
		errCodeInvalidSpotFleetRequestConfig, "SpotFleetRequestConfig.IamFleetRole",
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Spot Fleet Request: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*ec2.RequestSpotFleetOutput).SpotFleetRequestId))

	if _, err := waitSpotFleetRequestCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Spot Fleet Request (%s) create: %s", d.Id(), err)
	}

	if d.Get("wait_for_fulfillment").(bool) {
		if _, err := waitSpotFleetRequestFulfilled(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Spot Fleet Request (%s) fulfillment: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSpotFleetRequestRead(ctx, d, meta)...)
}

func resourceSpotFleetRequestRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	output, err := findSpotFleetRequestByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Spot Fleet Request %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Spot Fleet Request (%s): %s", d.Id(), err)
	}

	d.Set("spot_request_state", output.SpotFleetRequestState)

	config := output.SpotFleetRequestConfig

	d.Set("allocation_strategy", config.AllocationStrategy)

	// The default of this argument does not get set in the create operation
	// Therefore if the API does not return a value, being a *int32 type it will result in 0 and always create a diff.
	if config.InstancePoolsToUseCount != nil { // nosemgrep:ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("instance_pools_to_use_count", config.InstancePoolsToUseCount)
	}

	d.Set("client_token", config.ClientToken)
	d.Set("context", config.Context)
	d.Set("excess_capacity_termination_policy", config.ExcessCapacityTerminationPolicy)
	d.Set("iam_fleet_role", config.IamFleetRole)
	d.Set("spot_maintenance_strategies", flattenSpotMaintenanceStrategies(config.SpotMaintenanceStrategies))
	d.Set("spot_price", config.SpotPrice)
	d.Set("target_capacity", config.TargetCapacity)
	d.Set("target_capacity_unit_type", config.TargetCapacityUnitType)
	d.Set("terminate_instances_with_expiration", config.TerminateInstancesWithExpiration)
	if config.ValidFrom != nil {
		d.Set("valid_from", aws.ToTime(config.ValidFrom).Format(time.RFC3339))
	}
	if config.ValidUntil != nil {
		d.Set("valid_until", aws.ToTime(config.ValidUntil).Format(time.RFC3339))
	}

	launchSpec, err := launchSpecsToSet(ctx, conn, config.LaunchSpecifications)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Spot Fleet Request (%s) launch specifications: %s", d.Id(), err)
	}

	d.Set("replace_unhealthy_instances", config.ReplaceUnhealthyInstances)
	d.Set("instance_interruption_behaviour", config.InstanceInterruptionBehavior)
	d.Set("fleet_type", config.Type)
	d.Set("launch_specification", launchSpec)

	setTagsOut(ctx, output.Tags)

	if err := d.Set("launch_template_config", flattenLaunchTemplateConfigs(config.LaunchTemplateConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting launch_template_config: %s", err)
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
				return sdkdiag.AppendErrorf(diags, "setting load_balancers: %s", err)
			}
		}

		if lbConf.TargetGroupsConfig != nil {
			flatTgs := make([]*string, 0)
			for _, tg := range lbConf.TargetGroupsConfig.TargetGroups {
				flatTgs = append(flatTgs, tg.Arn)
			}
			if err := d.Set("target_group_arns", flex.FlattenStringSet(flatTgs)); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting target_group_arns: %s", err)
			}
		}
	}

	return diags
}

func resourceSpotFleetRequestUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifySpotFleetRequestInput{
			SpotFleetRequestId: aws.String(d.Id()),
		}

		if d.HasChange("target_capacity") {
			input.TargetCapacity = aws.Int32(int32(d.Get("target_capacity").(int)))
		}

		if d.HasChange("on_demand_target_capacity") {
			input.OnDemandTargetCapacity = aws.Int32(int32(d.Get("on_demand_target_capacity").(int)))
		}

		if d.HasChange("excess_capacity_termination_policy") {
			if val, ok := d.GetOk("excess_capacity_termination_policy"); ok {
				input.ExcessCapacityTerminationPolicy = awstypes.ExcessCapacityTerminationPolicy(val.(string))
			}
		}

		log.Printf("[DEBUG] Modifying EC2 Spot Fleet Request: %s", d.Id())
		if _, err := conn.ModifySpotFleetRequest(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Spot Fleet Request (%s): %s", d.Id(), err)
		}

		if _, err := waitSpotFleetRequestUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Spot Fleet Request (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSpotFleetRequestRead(ctx, d, meta)...)
}

func resourceSpotFleetRequestDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	terminateInstances := d.Get("terminate_instances_with_expiration").(bool)
	// If terminate_instances_on_delete is not null, its value is used.
	if v, null, _ := nullable.Bool(d.Get("terminate_instances_on_delete").(string)).ValueBool(); !null {
		terminateInstances = v
	}

	log.Printf("[INFO] Deleting EC2 Spot Fleet Request: %s", d.Id())
	output, err := conn.CancelSpotFleetRequests(ctx, &ec2.CancelSpotFleetRequestsInput{
		SpotFleetRequestIds: []string{d.Id()},
		TerminateInstances:  aws.Bool(terminateInstances),
	})

	if err == nil && output != nil {
		err = cancelSpotFleetRequestsError(output.UnsuccessfulFleetRequests)
	}

	if tfawserr.ErrCodeEquals(err, string(awstypes.CancelBatchErrorCodeFleetRequestIdDoesNotExist)) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "cancelling EC2 Spot Fleet Request (%s): %s", d.Id(), err)
	}

	// Only wait for instance termination if requested.
	if !terminateInstances {
		return diags
	}

	_, err = tfresource.RetryUntilNotFound(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		input := &ec2.DescribeSpotFleetInstancesInput{
			SpotFleetRequestId: aws.String(d.Id()),
		}
		output, err := findSpotFleetInstances(ctx, conn, input)

		if err != nil {
			return nil, err
		}

		if len(output) == 0 {
			return nil, tfresource.NewEmptyResultError(input)
		}

		return output, nil
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Spot Fleet Request (%s) active instance count to reach 0: %s", d.Id(), err)
	}

	return diags
}

func buildSpotFleetLaunchSpecification(ctx context.Context, d map[string]interface{}, meta interface{}) (awstypes.SpotFleetLaunchSpecification, error) {
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	opts := awstypes.SpotFleetLaunchSpecification{
		ImageId:      aws.String(d["ami"].(string)),
		InstanceType: awstypes.InstanceType(d[names.AttrInstanceType].(string)),
		SpotPrice:    aws.String(d["spot_price"].(string)),
	}

	placement := new(awstypes.SpotPlacement)
	if v, ok := d[names.AttrAvailabilityZone]; ok {
		placement.AvailabilityZone = aws.String(v.(string))
		opts.Placement = placement
	}

	if v, ok := d["placement_tenancy"]; ok {
		placement.Tenancy = awstypes.Tenancy(v.(string))
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
		opts.Monitoring = &awstypes.SpotFleetMonitoring{
			Enabled: aws.Bool(v.(bool)),
		}
	}

	if v, ok := d["iam_instance_profile"]; ok {
		opts.IamInstanceProfile = &awstypes.IamInstanceProfileSpecification{
			Name: aws.String(v.(string)),
		}
	}

	if v, ok := d["iam_instance_profile_arn"]; ok && v.(string) != "" {
		opts.IamInstanceProfile = &awstypes.IamInstanceProfileSpecification{
			Arn: aws.String(v.(string)),
		}
	}

	if v, ok := d["user_data"]; ok {
		opts.UserData = flex.StringValueToBase64String(v.(string))
	}

	if v, ok := d["key_name"]; ok && v != "" {
		opts.KeyName = aws.String(v.(string))
	}

	if v, ok := d["weighted_capacity"]; ok && v != "" {
		wc, _ := strconv.ParseFloat(v.(string), 64)
		opts.WeightedCapacity = aws.Float64(wc)
	}

	var securityGroupIds []string
	if v, ok := d[names.AttrVPCSecurityGroupIDs]; ok {
		if s := v.(*schema.Set); s.Len() > 0 {
			for _, v := range s.List() {
				securityGroupIds = append(securityGroupIds, v.(string))
			}
		}
	}

	if m, ok := d[names.AttrTags].(map[string]interface{}); ok && len(m) > 0 {
		tagsSpec := make([]awstypes.SpotFleetTagSpecification, 0)

		tags := Tags(tftags.New(ctx, m).IgnoreAWS())

		spec := awstypes.SpotFleetTagSpecification{
			ResourceType: awstypes.ResourceTypeInstance,
			Tags:         tags,
		}

		tagsSpec = append(tagsSpec, spec)

		opts.TagSpecifications = tagsSpec
	}

	subnetId, hasSubnetId := d[names.AttrSubnetID]
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
		ni := awstypes.InstanceNetworkInterfaceSpecification{
			AssociatePublicIpAddress: aws.Bool(true),
			DeleteOnTermination:      aws.Bool(true),
			DeviceIndex:              aws.Int32(0),
			SubnetId:                 aws.String(subnetId.(string)),
			Groups:                   securityGroupIds,
		}

		opts.NetworkInterfaces = []awstypes.InstanceNetworkInterfaceSpecification{ni}
		opts.SubnetId = aws.String("")
	} else {
		for _, id := range securityGroupIds {
			opts.SecurityGroups = append(opts.SecurityGroups, awstypes.GroupIdentifier{GroupId: aws.String(id)})
		}
	}

	blockDevices, err := readSpotFleetBlockDeviceMappingsFromConfig(ctx, d, conn)
	if err != nil {
		return awstypes.SpotFleetLaunchSpecification{}, err
	}
	if len(blockDevices) > 0 {
		opts.BlockDeviceMappings = blockDevices
	}

	return opts, nil
}

func readSpotFleetBlockDeviceMappingsFromConfig(ctx context.Context, d map[string]interface{}, conn *ec2.Client) ([]awstypes.BlockDeviceMapping, error) {
	blockDevices := make([]awstypes.BlockDeviceMapping, 0)

	if v, ok := d["ebs_block_device"]; ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			ebs := &awstypes.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd[names.AttrDeleteOnTermination].(bool)),
			}

			if v, ok := bd[names.AttrSnapshotID].(string); ok && v != "" {
				ebs.SnapshotId = aws.String(v)
			}

			if v, ok := bd[names.AttrEncrypted].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd[names.AttrKMSKeyID].(string); ok && v != "" {
				ebs.KmsKeyId = aws.String(v)
			}

			if v, ok := bd[names.AttrVolumeSize].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int32(int32(v))
			}

			if v, ok := bd[names.AttrVolumeType].(string); ok && v != "" {
				ebs.VolumeType = awstypes.VolumeType(v)
			}

			if v, ok := bd[names.AttrIOPS].(int); ok && v > 0 {
				ebs.Iops = aws.Int32(int32(v))
			}

			if v, ok := bd[names.AttrThroughput].(int); ok && v > 0 {
				ebs.Throughput = aws.Int32(int32(v))
			}

			blockDevices = append(blockDevices, awstypes.BlockDeviceMapping{
				DeviceName: aws.String(bd[names.AttrDeviceName].(string)),
				Ebs:        ebs,
			})
		}
	}

	if v, ok := d["ephemeral_block_device"]; ok {
		vL := v.(*schema.Set).List()
		for _, v := range vL {
			bd := v.(map[string]interface{})
			blockDevices = append(blockDevices, awstypes.BlockDeviceMapping{
				DeviceName:  aws.String(bd[names.AttrDeviceName].(string)),
				VirtualName: aws.String(bd[names.AttrVirtualName].(string)),
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
			ebs := &awstypes.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(bd[names.AttrDeleteOnTermination].(bool)),
			}

			if v, ok := bd[names.AttrEncrypted].(bool); ok && v {
				ebs.Encrypted = aws.Bool(v)
			}

			if v, ok := bd[names.AttrKMSKeyID].(string); ok && v != "" {
				ebs.KmsKeyId = aws.String(v)
			}

			if v, ok := bd[names.AttrVolumeSize].(int); ok && v != 0 {
				ebs.VolumeSize = aws.Int32(int32(v))
			}

			if v, ok := bd[names.AttrVolumeType].(string); ok && v != "" {
				ebs.VolumeType = awstypes.VolumeType(v)
			}

			if v, ok := bd[names.AttrIOPS].(int); ok && v > 0 {
				ebs.Iops = aws.Int32(int32(v))
			}

			if v, ok := bd[names.AttrThroughput].(int); ok && v > 0 {
				ebs.Throughput = aws.Int32(int32(v))
			}

			if dn, err := findRootDeviceName(ctx, conn, d["ami"].(string)); err == nil {
				if dn == nil {
					return nil, fmt.Errorf(
						"Expected 1 AMI for ID: %s, got none",
						d["ami"].(string))
				}

				blockDevices = append(blockDevices, awstypes.BlockDeviceMapping{
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

func buildSpotFleetLaunchSpecifications(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]awstypes.SpotFleetLaunchSpecification, error) {
	userSpecs := d.Get("launch_specification").(*schema.Set).List()
	specs := make([]awstypes.SpotFleetLaunchSpecification, len(userSpecs))
	for i, userSpec := range userSpecs {
		userSpecMap := userSpec.(map[string]interface{})
		// panic: interface conversion: interface {} is map[string]interface {}, not *schema.ResourceData
		opts, err := buildSpotFleetLaunchSpecification(ctx, userSpecMap, meta)
		if err != nil {
			return nil, err
		}
		specs[i] = opts
	}

	return specs, nil
}

func expandLaunchTemplateConfig(tfMap map[string]interface{}) awstypes.LaunchTemplateConfig {
	apiObject := awstypes.LaunchTemplateConfig{}

	if v, ok := tfMap["launch_template_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplateSpecification = expandFleetLaunchTemplateSpecification(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["overrides"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Overrides = expandLaunchTemplateOverrideses(v.List())
	}

	return apiObject
}

func expandLaunchTemplateConfigs(tfList []interface{}) []awstypes.LaunchTemplateConfig {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateConfig

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateConfig(tfMap))
	}

	return apiObjects
}

func expandFleetLaunchTemplateSpecification(tfMap map[string]interface{}) *awstypes.FleetLaunchTemplateSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FleetLaunchTemplateSpecification{}

	if v, ok := tfMap[names.AttrID].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateOverrides(tfMap map[string]interface{}) awstypes.LaunchTemplateOverrides {
	apiObject := awstypes.LaunchTemplateOverrides{}

	if v, ok := tfMap[names.AttrAvailabilityZone].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["instance_requirements"].([]interface{}); ok && len(v) > 0 {
		apiObject.InstanceRequirements = expandInstanceRequirements(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrInstanceType].(string); ok && v != "" {
		apiObject.InstanceType = awstypes.InstanceType(v)
	}

	if v, ok := tfMap[names.AttrPriority].(float64); ok && v != 0.0 {
		apiObject.Priority = aws.Float64(v)
	}

	if v, ok := tfMap["spot_price"].(string); ok && v != "" {
		apiObject.SpotPrice = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSubnetID].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	if v, ok := tfMap["weighted_capacity"].(float64); ok && v != 0.0 {
		apiObject.WeightedCapacity = aws.Float64(v)
	}

	return apiObject
}

func expandLaunchTemplateOverrideses(tfList []interface{}) []awstypes.LaunchTemplateOverrides {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateOverrides

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateOverrides(tfMap))
	}

	return apiObjects
}

func expandInstanceRequirements(tfMap map[string]interface{}) *awstypes.InstanceRequirements {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstanceRequirements{}

	if v, ok := tfMap["accelerator_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.AcceleratorCount = expandAcceleratorCount(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["accelerator_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorManufacturers = flex.ExpandStringyValueSet[awstypes.AcceleratorManufacturer](v)
	}

	if v, ok := tfMap["accelerator_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorNames = flex.ExpandStringyValueSet[awstypes.AcceleratorName](v)
	}

	if v, ok := tfMap["accelerator_total_memory_mib"].([]interface{}); ok && len(v) > 0 {
		apiObject.AcceleratorTotalMemoryMiB = expandAcceleratorTotalMemoryMiB(v[0].(map[string]interface{}))
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
		apiObject.BaselineEbsBandwidthMbps = expandBaselineEBSBandwidthMbps(v[0].(map[string]interface{}))
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

	if v, ok := tfMap["memory_gib_per_vcpu"].([]interface{}); ok && len(v) > 0 {
		apiObject.MemoryGiBPerVCpu = expandMemoryGiBPerVCPU(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["memory_mib"].([]interface{}); ok && len(v) > 0 {
		apiObject.MemoryMiB = expandMemoryMiB(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["network_interface_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.NetworkInterfaceCount = expandNetworkInterfaceCount(v[0].(map[string]interface{}))
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
		apiObject.TotalLocalStorageGB = expandTotalLocalStorageGB(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["vcpu_count"].([]interface{}); ok && len(v) > 0 {
		apiObject.VCpuCount = expandVCPUCountRange(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAcceleratorCount(tfMap map[string]interface{}) *awstypes.AcceleratorCount {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AcceleratorCount{}

	if v, ok := tfMap[names.AttrMax].(int); ok {
		apiObject.Max = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMin].(int); ok {
		apiObject.Min = aws.Int32(int32(v))
	}

	return apiObject
}

func expandAcceleratorTotalMemoryMiB(tfMap map[string]interface{}) *awstypes.AcceleratorTotalMemoryMiB {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.AcceleratorTotalMemoryMiB{}

	if v, ok := tfMap[names.AttrMax].(int); ok {
		apiObject.Max = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMin].(int); ok {
		apiObject.Min = aws.Int32(int32(v))
	}

	return apiObject
}

func expandBaselineEBSBandwidthMbps(tfMap map[string]interface{}) *awstypes.BaselineEbsBandwidthMbps {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.BaselineEbsBandwidthMbps{}

	if v, ok := tfMap[names.AttrMax].(int); ok {
		apiObject.Max = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMin].(int); ok {
		apiObject.Min = aws.Int32(int32(v))
	}

	return apiObject
}

func expandMemoryGiBPerVCPU(tfMap map[string]interface{}) *awstypes.MemoryGiBPerVCpu {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MemoryGiBPerVCpu{}

	if v, ok := tfMap[names.AttrMax].(float64); ok {
		apiObject.Max = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrMin].(float64); ok {
		apiObject.Min = aws.Float64(v)
	}

	return apiObject
}

func expandMemoryMiB(tfMap map[string]interface{}) *awstypes.MemoryMiB {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MemoryMiB{}

	if v, ok := tfMap[names.AttrMax].(int); ok {
		apiObject.Max = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMin].(int); ok {
		apiObject.Min = aws.Int32(int32(v))
	}

	return apiObject
}

func expandNetworkInterfaceCount(tfMap map[string]interface{}) *awstypes.NetworkInterfaceCount {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.NetworkInterfaceCount{}

	if v, ok := tfMap[names.AttrMax].(int); ok {
		apiObject.Max = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMin].(int); ok {
		apiObject.Min = aws.Int32(int32(v))
	}

	return apiObject
}

func expandTotalLocalStorageGB(tfMap map[string]interface{}) *awstypes.TotalLocalStorageGB {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TotalLocalStorageGB{}

	if v, ok := tfMap[names.AttrMax].(float64); ok {
		apiObject.Max = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrMin].(float64); ok {
		apiObject.Min = aws.Float64(v)
	}

	return apiObject
}

func expandVCPUCountRange(tfMap map[string]interface{}) *awstypes.VCpuCountRange {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VCpuCountRange{}

	if v, ok := tfMap[names.AttrMax].(int); ok {
		apiObject.Max = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrMin].(int); ok {
		apiObject.Min = aws.Int32(int32(v))
	}

	return apiObject
}

func expandSpotMaintenanceStrategies(l []interface{}) *awstypes.SpotMaintenanceStrategies {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	fleetSpotMaintenanceStrategies := &awstypes.SpotMaintenanceStrategies{
		CapacityRebalance: expandSpotCapacityRebalance(m["capacity_rebalance"].([]interface{})),
	}

	return fleetSpotMaintenanceStrategies
}

func expandSpotCapacityRebalance(l []interface{}) *awstypes.SpotCapacityRebalance {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	capacityRebalance := &awstypes.SpotCapacityRebalance{}

	if v, ok := m["replacement_strategy"]; ok && v.(string) != "" {
		capacityRebalance.ReplacementStrategy = awstypes.ReplacementStrategy(v.(string))
	}

	return capacityRebalance
}

func launchSpecsToSet(ctx context.Context, conn *ec2.Client, launchSpecs []awstypes.SpotFleetLaunchSpecification) (*schema.Set, error) {
	specSet := &schema.Set{F: hashLaunchSpecification}
	for _, spec := range launchSpecs {
		rootDeviceName, err := findRootDeviceName(ctx, conn, aws.ToString(spec.ImageId))
		if err != nil {
			return nil, err
		}

		specSet.Add(launchSpecToMap(ctx, spec, rootDeviceName))
	}
	return specSet, nil
}

func launchSpecToMap(ctx context.Context, l awstypes.SpotFleetLaunchSpecification, rootDevName *string) map[string]interface{} {
	m := make(map[string]interface{})

	m["root_block_device"] = rootBlockDeviceToSet(l.BlockDeviceMappings, rootDevName)
	m["ebs_block_device"] = ebsBlockDevicesToSet(l.BlockDeviceMappings, rootDevName)
	m["ephemeral_block_device"] = ephemeralBlockDevicesToSet(l.BlockDeviceMappings)

	if l.ImageId != nil {
		m["ami"] = aws.ToString(l.ImageId)
	}

	if l.InstanceType != "" {
		m[names.AttrInstanceType] = l.InstanceType
	}

	if l.SpotPrice != nil {
		m["spot_price"] = aws.ToString(l.SpotPrice)
	}

	if l.EbsOptimized != nil {
		m["ebs_optimized"] = aws.ToBool(l.EbsOptimized)
	}

	if l.Monitoring != nil && l.Monitoring.Enabled != nil {
		m["monitoring"] = aws.ToBool(l.Monitoring.Enabled)
	}

	if l.IamInstanceProfile != nil && l.IamInstanceProfile.Name != nil {
		m["iam_instance_profile"] = aws.ToString(l.IamInstanceProfile.Name)
	}

	if l.IamInstanceProfile != nil && l.IamInstanceProfile.Arn != nil {
		m["iam_instance_profile_arn"] = aws.ToString(l.IamInstanceProfile.Arn)
	}

	if l.UserData != nil {
		m["user_data"] = userDataHashSum(aws.ToString(l.UserData))
	}

	if l.KeyName != nil {
		m["key_name"] = aws.ToString(l.KeyName)
	}

	if l.Placement != nil {
		m[names.AttrAvailabilityZone] = aws.ToString(l.Placement.AvailabilityZone)
	}

	if l.SubnetId != nil {
		m[names.AttrSubnetID] = aws.ToString(l.SubnetId)
	}

	securityGroupIds := &schema.Set{F: schema.HashString}
	if len(l.NetworkInterfaces) > 0 {
		m["associate_public_ip_address"] = aws.ToBool(l.NetworkInterfaces[0].AssociatePublicIpAddress)
		m[names.AttrSubnetID] = aws.ToString(l.NetworkInterfaces[0].SubnetId)

		for _, group := range l.NetworkInterfaces[0].Groups {
			securityGroupIds.Add(group)
		}
	} else {
		for _, group := range l.SecurityGroups {
			securityGroupIds.Add(aws.ToString(group.GroupId))
		}
	}
	m[names.AttrVPCSecurityGroupIDs] = securityGroupIds

	if l.WeightedCapacity != nil {
		m["weighted_capacity"] = strconv.FormatFloat(*l.WeightedCapacity, 'f', 0, 64)
	}

	if l.TagSpecifications != nil {
		for _, tagSpecs := range l.TagSpecifications {
			// only "instance" tags are currently supported: http://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_SpotFleetTagSpecification.html
			if tagSpecs.ResourceType == awstypes.ResourceTypeInstance {
				m[names.AttrTags] = keyValueTags(ctx, tagSpecs.Tags).IgnoreAWS().Map()
			}
		}
	}

	return m
}

func ebsBlockDevicesToSet(bdm []awstypes.BlockDeviceMapping, rootDevName *string) *schema.Set {
	set := &schema.Set{F: hashEBSBlockDevice}

	for _, val := range bdm {
		if val.Ebs != nil {
			m := make(map[string]interface{})

			ebs := val.Ebs

			if val.DeviceName != nil {
				if aws.ToString(rootDevName) == aws.ToString(val.DeviceName) {
					continue
				}

				m[names.AttrDeviceName] = aws.ToString(val.DeviceName)
			}

			if ebs.DeleteOnTermination != nil {
				m[names.AttrDeleteOnTermination] = aws.ToBool(ebs.DeleteOnTermination)
			}

			if ebs.SnapshotId != nil {
				m[names.AttrSnapshotID] = aws.ToString(ebs.SnapshotId)
			}

			if ebs.Encrypted != nil {
				m[names.AttrEncrypted] = aws.ToBool(ebs.Encrypted)
			}

			if ebs.KmsKeyId != nil {
				m[names.AttrKMSKeyID] = aws.ToString(ebs.KmsKeyId)
			}

			if ebs.VolumeSize != nil {
				m[names.AttrVolumeSize] = aws.ToInt32(ebs.VolumeSize)
			}

			if ebs.VolumeType != "" {
				m[names.AttrVolumeType] = ebs.VolumeType
			}

			if ebs.Iops != nil {
				m[names.AttrIOPS] = aws.ToInt32(ebs.Iops)
			}

			if ebs.Throughput != nil {
				m[names.AttrThroughput] = aws.ToInt32(ebs.Throughput)
			}

			set.Add(m)
		}
	}

	return set
}

func ephemeralBlockDevicesToSet(bdm []awstypes.BlockDeviceMapping) *schema.Set {
	set := &schema.Set{F: hashEphemeralBlockDevice}

	for _, val := range bdm {
		if val.VirtualName != nil {
			m := make(map[string]interface{})
			m[names.AttrVirtualName] = aws.ToString(val.VirtualName)

			if val.DeviceName != nil {
				m[names.AttrDeviceName] = aws.ToString(val.DeviceName)
			}

			set.Add(m)
		}
	}

	return set
}

func rootBlockDeviceToSet(bdm []awstypes.BlockDeviceMapping, rootDevName *string) *schema.Set {
	set := &schema.Set{F: hashRootBlockDevice}

	if rootDevName != nil {
		for _, val := range bdm {
			if aws.ToString(val.DeviceName) == aws.ToString(rootDevName) {
				m := make(map[string]interface{})
				if val.Ebs.DeleteOnTermination != nil {
					m[names.AttrDeleteOnTermination] = aws.ToBool(val.Ebs.DeleteOnTermination)
				}

				if val.Ebs.Encrypted != nil {
					m[names.AttrEncrypted] = aws.ToBool(val.Ebs.Encrypted)
				}

				if val.Ebs.KmsKeyId != nil {
					m[names.AttrKMSKeyID] = aws.ToString(val.Ebs.KmsKeyId)
				}

				if val.Ebs.VolumeSize != nil {
					m[names.AttrVolumeSize] = aws.ToInt32(val.Ebs.VolumeSize)
				}

				if val.Ebs.VolumeType != "" {
					m[names.AttrVolumeType] = val.Ebs.VolumeType
				}

				if val.Ebs.Iops != nil {
					m[names.AttrIOPS] = aws.ToInt32(val.Ebs.Iops)
				}

				if val.Ebs.Throughput != nil {
					m[names.AttrThroughput] = aws.ToInt32(val.Ebs.Throughput)
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
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrDeviceName].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrVirtualName].(string)))
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
	if v, ok := m[names.AttrAvailabilityZone].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := m[names.AttrSubnetID].(string); ok && v != "" {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	buf.WriteString(fmt.Sprintf("%s-", m[names.AttrInstanceType]))
	buf.WriteString(fmt.Sprintf("%s-", m["spot_price"].(string)))
	return create.StringHashcode(buf.String())
}

func hashEBSBlockDevice(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if name, ok := m[names.AttrDeviceName]; ok {
		buf.WriteString(fmt.Sprintf("%s-", name.(string)))
	}
	if id, ok := m[names.AttrSnapshotID]; ok {
		buf.WriteString(fmt.Sprintf("%s-", id.(string)))
	}
	return create.StringHashcode(buf.String())
}

func flattenLaunchTemplateConfig(apiObject awstypes.LaunchTemplateConfig) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateSpecification; v != nil {
		tfMap["launch_template_specification"] = []interface{}{flattenFleetLaunchTemplateSpecificationForSpotFleetRequest(v)}
	}

	if v := apiObject.Overrides; v != nil {
		tfMap["overrides"] = flattenLaunchTemplateOverrideses(v)
	}

	return tfMap
}

func flattenLaunchTemplateConfigs(apiObjects []awstypes.LaunchTemplateConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateConfig(apiObject))
	}

	return tfList
}

func flattenFleetLaunchTemplateSpecificationForSpotFleetRequest(apiObject *awstypes.FleetLaunchTemplateSpecification) map[string]interface{} {
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

func flattenLaunchTemplateOverrides(apiObject awstypes.LaunchTemplateOverrides) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap[names.AttrAvailabilityZone] = aws.ToString(v)
	}

	if v := apiObject.InstanceRequirements; v != nil {
		tfMap["instance_requirements"] = []interface{}{flattenInstanceRequirements(v)}
	}

	if v := apiObject.InstanceType; v != "" {
		tfMap[names.AttrInstanceType] = v
	}

	if v := apiObject.Priority; v != nil {
		tfMap[names.AttrPriority] = aws.ToFloat64(v)
	}

	if v := apiObject.SpotPrice; v != nil {
		tfMap["spot_price"] = aws.ToString(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap[names.AttrSubnetID] = aws.ToString(v)
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.ToFloat64(v)
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

func flattenSpotMaintenanceStrategies(spotMaintenanceStrategies *awstypes.SpotMaintenanceStrategies) []interface{} {
	if spotMaintenanceStrategies == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"capacity_rebalance": flattenSpotCapacityRebalance(spotMaintenanceStrategies.CapacityRebalance),
	}

	return []interface{}{m}
}

func flattenSpotCapacityRebalance(spotCapacityRebalance *awstypes.SpotCapacityRebalance) []interface{} {
	if spotCapacityRebalance == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"replacement_strategy": spotCapacityRebalance.ReplacementStrategy,
	}

	return []interface{}{m}
}
