// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_fleet", name="Fleet")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFleetCreate,
		ReadWithoutTimeout:   resourceFleetRead,
		UpdateWithoutTimeout: resourceFleetUpdate,
		DeleteWithoutTimeout: resourceFleetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			resourceFleetCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"context": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"excess_capacity_termination_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.FleetExcessCapacityTerminationPolicyTermination,
				ValidateDiagFunc: enum.Validate[awstypes.FleetExcessCapacityTerminationPolicy](),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Get(names.AttrType) != string(awstypes.FleetTypeMaintain)
				},
				DiffSuppressOnRefresh: true,
			},
			"fleet_instance_set": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_ids": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						names.AttrInstanceType: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"lifecycle": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"platform": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"fleet_state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"fulfilled_capacity": {
				Type:     schema.TypeFloat,
				Optional: true,
				Computed: true,
			},
			"fulfilled_on_demand_capacity": {
				Type:     schema.TypeFloat,
				Optional: true,
				Computed: true,
			},
			"launch_template_config": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 0,
				MaxItems: 50,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"launch_template_specification": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"launch_template_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidLaunchTemplateName,
									},
									names.AttrVersion: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"override": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 300,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAvailabilityZone: {
										Type:     schema.TypeString,
										Optional: true,
									},
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
													Required: true,
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
																Required:     true,
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
													Required: true,
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
																Required:     true,
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
									// "image_id": {
									// 	Type:     schema.TypeString,
									// 	Optional: true,
									// },
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
									// Pending AWS to provide this attribute back in the `Describe` call
									// "placement": {
									// 	Type:     schema.TypeList,
									// 	Optional: true,
									// 	MaxItems: 1,
									// 	Elem: &schema.Resource{
									// 		Schema: map[string]*schema.Schema{
									// 			"group_name": {
									// 				Type:     schema.TypeString,
									// 				Optional: true,
									// 			},
									// 		},
									// 	},
									// },
									names.AttrPriority: {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									names.AttrSubnetID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"weighted_capacity": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"on_demand_options": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      fleetOnDemandAllocationStrategyLowestPrice,
							ValidateFunc: validation.StringInSlice(fleetOnDemandAllocationStrategy_Values(), false),
						},
						"capacity_reservation_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"usage_strategy": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.FleetCapacityReservationUsageStrategy](),
									},
								},
							},
						},
						"max_total_price": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"min_target_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"single_availability_zone": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"single_instance_type": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"replace_unhealthy_instances": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"spot_options": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      spotAllocationStrategyLowestPrice,
							ValidateFunc: validation.StringInSlice(spotAllocationStrategy_Values(), false),
						},
						"instance_interruption_behavior": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							Default:          awstypes.SpotInstanceInterruptionBehaviorTerminate,
							ValidateDiagFunc: enum.Validate[awstypes.SpotInstanceInterruptionBehavior](),
						},
						"instance_pools_to_use_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      1,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"maintenance_strategies": {
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
													ValidateDiagFunc: enum.Validate[awstypes.FleetReplacementStrategy](),
												},
												"termination_delay": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(120, 7200),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_capacity_specification": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_target_capacity_type": {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DefaultTargetCapacityType](),
						},
						"on_demand_target_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Show difference for new resources
								if d.Id() == "" {
									return false
								}
								// Show difference if value is configured
								if new != "0" {
									return false
								}
								// Show difference if existing state reflects different default type
								defaultTargetCapacityTypeO, _ := d.GetChange("target_capacity_specification.0.default_target_capacity_type")
								if defaultTargetCapacityTypeO.(string) != string(awstypes.DefaultTargetCapacityTypeOnDemand) {
									return false
								}
								// Show difference if existing state reflects different total capacity
								oldInt, err := strconv.Atoi(old)
								if err != nil {
									log.Printf("[WARN] %s DiffSuppressFunc error converting %s to integer: %s", k, old, err)
									return false
								}
								totalTargetCapacityO, _ := d.GetChange("target_capacity_specification.0.total_target_capacity")
								return oldInt == totalTargetCapacityO.(int)
							},
						},
						"spot_target_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Show difference for new resources
								if d.Id() == "" {
									return false
								}
								// Show difference if value is configured
								if new != "0" {
									return false
								}
								// Show difference if existing state reflects different default type
								defaultTargetCapacityTypeO, _ := d.GetChange("target_capacity_specification.0.default_target_capacity_type")
								if defaultTargetCapacityTypeO.(string) != string(awstypes.DefaultTargetCapacityTypeSpot) {
									return false
								}
								// Show difference if existing state reflects different total capacity
								oldInt, err := strconv.Atoi(old)
								if err != nil {
									log.Printf("[WARN] %s DiffSuppressFunc error converting %s to integer: %s", k, old, err)
									return false
								}
								totalTargetCapacityO, _ := d.GetChange("target_capacity_specification.0.total_target_capacity")
								return oldInt == totalTargetCapacityO.(int)
							},
						},
						"target_capacity_unit_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TargetCapacityUnitType](),
						},
						"total_target_capacity": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"terminate_instances": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"terminate_instances_with_expiration": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.FleetTypeMaintain,
				ValidateDiagFunc: enum.Validate[awstypes.FleetType](),
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
		},
	}
}

func resourceFleetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	fleetType := awstypes.FleetType(d.Get(names.AttrType).(string))
	input := &ec2.CreateFleetInput{
		ClientToken:                 aws.String(id.UniqueId()),
		LaunchTemplateConfigs:       expandFleetLaunchTemplateConfigRequests(d.Get("launch_template_config").([]interface{})),
		TargetCapacitySpecification: expandTargetCapacitySpecificationRequest(d.Get("target_capacity_specification").([]interface{})[0].(map[string]interface{})),
		TagSpecifications:           getTagSpecificationsIn(ctx, awstypes.ResourceTypeFleet),
		Type:                        fleetType,
	}

	if v, ok := d.GetOk("context"); ok {
		input.Context = aws.String(v.(string))
	}

	// This argument is only valid for fleet_type of `maintain`, but was defaulted in the schema above, hence the extra check.
	if v, ok := d.GetOk("excess_capacity_termination_policy"); ok && v != "" && fleetType == awstypes.FleetTypeMaintain {
		input.ExcessCapacityTerminationPolicy = awstypes.FleetExcessCapacityTerminationPolicy(v.(string))
	}

	if v, ok := d.GetOk("on_demand_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OnDemandOptions = expandOnDemandOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("replace_unhealthy_instances"); ok {
		input.ReplaceUnhealthyInstances = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("spot_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SpotOptions = expandSpotOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("terminate_instances_with_expiration"); ok {
		input.TerminateInstancesWithExpiration = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("valid_from"); ok {
		validFrom, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing valid_from: %s", err)
		}
		input.ValidFrom = aws.Time(validFrom)
	}

	if v, ok := d.GetOk("valid_until"); ok {
		validUntil, err := time.Parse(time.RFC3339, v.(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing valid_until: %s", err)
		}
		input.ValidUntil = aws.Time(validUntil)
	}

	output, err := conn.CreateFleet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Fleet: %s", err)
	}

	d.SetId(aws.ToString(output.FleetId))

	// If a request type is fulfilled immediately, we can miss the transition from active to deleted.
	// Instead of an error here, allow the Read function to trigger recreation.
	if input.ValidFrom == nil {
		targetStates := enum.Slice(awstypes.FleetStateCodeActive)
		if fleetType == awstypes.FleetTypeRequest {
			targetStates = append(targetStates, string(awstypes.FleetStateCodeDeleted), string(awstypes.FleetStateCodeDeletedRunning), string(awstypes.FleetStateCodeDeletedTerminatingInstances))
		}

		if err := waitFleet(ctx, conn, d.Id(), enum.Slice(awstypes.FleetStateCodeSubmitted), targetStates, d.Timeout(schema.TimeoutCreate), 0); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Fleet (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	fleet, err := findFleetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Fleet %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Fleet (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("fleet/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("context", fleet.Context)
	d.Set("excess_capacity_termination_policy", fleet.ExcessCapacityTerminationPolicy)
	if fleet.Instances != nil {
		if err := d.Set("fleet_instance_set", flattenFleetInstanceSet(fleet.Instances)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting fleet_instance_set: %s", err)
		}
	}
	d.Set("fleet_state", fleet.FleetState)
	d.Set("fulfilled_capacity", fleet.FulfilledCapacity)
	d.Set("fulfilled_on_demand_capacity", fleet.FulfilledOnDemandCapacity)
	if err := d.Set("launch_template_config", flattenFleetLaunchTemplateConfigs(fleet.LaunchTemplateConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting launch_template_config: %s", err)
	}
	if fleet.OnDemandOptions != nil {
		if err := d.Set("on_demand_options", []interface{}{flattenOnDemandOptions(fleet.OnDemandOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting on_demand_options: %s", err)
		}
	} else {
		d.Set("on_demand_options", nil)
	}
	d.Set("replace_unhealthy_instances", fleet.ReplaceUnhealthyInstances)
	if fleet.SpotOptions != nil {
		if err := d.Set("spot_options", []interface{}{flattenSpotOptions(fleet.SpotOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting spot_options: %s", err)
		}
	} else {
		d.Set("spot_options", nil)
	}
	if fleet.TargetCapacitySpecification != nil {
		if err := d.Set("target_capacity_specification", []interface{}{flattenTargetCapacitySpecification(fleet.TargetCapacitySpecification)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting target_capacity_specification: %s", err)
		}
	} else {
		d.Set("target_capacity_specification", nil)
	}
	d.Set("terminate_instances_with_expiration", fleet.TerminateInstancesWithExpiration)
	d.Set(names.AttrType, fleet.Type)
	if fleet.ValidFrom != nil && aws.ToTime(fleet.ValidFrom).Format(time.RFC3339) != "1970-01-01T00:00:00Z" {
		d.Set("valid_from", aws.ToTime(fleet.ValidFrom).Format(time.RFC3339))
	}
	if fleet.ValidUntil != nil && aws.ToTime(fleet.ValidUntil).Format(time.RFC3339) != "1970-01-01T00:00:00Z" {
		d.Set("valid_until", aws.ToTime(fleet.ValidUntil).Format(time.RFC3339))
	}

	setTagsOut(ctx, fleet.Tags)

	return diags
}

func resourceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &ec2.ModifyFleetInput{
			FleetId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("context"); ok {
			input.Context = aws.String(v.(string))
		}

		// This argument is only valid for fleet_type of `maintain`, but was defaulted in the schema above, hence the extra check.
		if v, ok := d.GetOk("excess_capacity_termination_policy"); ok && v != "" && awstypes.FleetType(d.Get(names.AttrType).(string)) == awstypes.FleetTypeMaintain {
			input.ExcessCapacityTerminationPolicy = awstypes.FleetExcessCapacityTerminationPolicy(v.(string))
		}

		input.LaunchTemplateConfigs = expandFleetLaunchTemplateConfigRequests(d.Get("launch_template_config").([]interface{}))

		// InvalidTargetCapacitySpecification: Currently we only support total target capacity modification.
		// TargetCapacitySpecification: expandEc2TargetCapacitySpecificationRequest(d.Get("target_capacity_specification").([]interface{})),
		input.TargetCapacitySpecification = &awstypes.TargetCapacitySpecificationRequest{
			TotalTargetCapacity: aws.Int32(int32(d.Get("target_capacity_specification.0.total_target_capacity").(int))),
		}

		_, err := conn.ModifyFleet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying EC2 Fleet (%s): %s", d.Id(), err)
		}

		if err := waitFleet(ctx, conn, d.Id(), enum.Slice(awstypes.FleetStateCodeModifying), enum.Slice(awstypes.FleetStateCodeActive), d.Timeout(schema.TimeoutUpdate), 0); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Fleet (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFleetRead(ctx, d, meta)...)
}

func resourceFleetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Fleet: %s", d.Id())
	output, err := conn.DeleteFleets(ctx, &ec2.DeleteFleetsInput{
		FleetIds:           []string{d.Id()},
		TerminateInstances: aws.Bool(d.Get("terminate_instances").(bool)),
	})

	if err == nil && output != nil {
		err = deleteFleetsError(output.UnsuccessfulFleetDeletions)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidFleetIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Fleet (%s): %s", d.Id(), err)
	}

	// Limiting waiter to non-instant fleet types.
	// `instant` fleet state is eventually consistent and can take 48 hours to update.
	if d.Get(names.AttrType) != "instant" {
		delay := 0 * time.Second
		pendingStates := enum.Slice(awstypes.FleetStateCodeActive)
		targetStates := enum.Slice(awstypes.FleetStateCodeDeleted)
		if d.Get("terminate_instances").(bool) {
			pendingStates = append(pendingStates, string(awstypes.FleetStateCodeDeletedTerminatingInstances))
			delay = 5 * time.Minute
		} else {
			targetStates = append(targetStates, string(awstypes.FleetStateCodeDeletedRunning))
		}

		if err := waitFleet(ctx, conn, d.Id(), pendingStates, targetStates, d.Timeout(schema.TimeoutDelete), delay); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Fleet (%s) delete: %s", d.Id(), err)
		}
	}

	return diags
}

func resourceFleetCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if diff.Id() == "" { // New resource.
		if diff.Get(names.AttrType).(string) != string(awstypes.FleetTypeMaintain) {
			if v, ok := diff.GetOk("spot_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				tfMap := v.([]interface{})[0].(map[string]interface{})
				if v, ok := tfMap["maintenance_strategies"].([]interface{}); ok && len(v) > 0 {
					return errors.New(`EC2 Fleet has an invalid configuration and can not be created. Capacity Rebalance maintenance strategies can only be specified for fleets of type maintain.`)
				}
			}
		}
	}

	return nil
}

func expandCapacityReservationOptionsRequest(tfMap map[string]interface{}) *awstypes.CapacityReservationOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.CapacityReservationOptionsRequest{}

	if v, ok := tfMap["usage_strategy"].(string); ok && v != "" {
		apiObject.UsageStrategy = awstypes.FleetCapacityReservationUsageStrategy(v)
	}

	return apiObject
}

func expandFleetLaunchTemplateConfigRequests(tfList []interface{}) []awstypes.FleetLaunchTemplateConfigRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FleetLaunchTemplateConfigRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandFleetLaunchTemplateConfigRequest(tfMap))
	}

	return apiObjects
}

func expandFleetLaunchTemplateConfigRequest(tfMap map[string]interface{}) awstypes.FleetLaunchTemplateConfigRequest {
	apiObject := awstypes.FleetLaunchTemplateConfigRequest{}

	if v, ok := tfMap["launch_template_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplateSpecification = expandFleetLaunchTemplateSpecificationRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["override"].([]interface{}); ok && len(v) > 0 {
		apiObject.Overrides = expandFleetLaunchTemplateOverridesRequests(v)
	}

	return apiObject
}

func expandFleetLaunchTemplateSpecificationRequest(tfMap map[string]interface{}) *awstypes.FleetLaunchTemplateSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FleetLaunchTemplateSpecificationRequest{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVersion].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandFleetLaunchTemplateOverridesRequests(tfList []interface{}) []awstypes.FleetLaunchTemplateOverridesRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.FleetLaunchTemplateOverridesRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandFleetLaunchTemplateOverridesRequest(tfMap))
	}

	return apiObjects
}

func expandFleetLaunchTemplateOverridesRequest(tfMap map[string]interface{}) awstypes.FleetLaunchTemplateOverridesRequest {
	apiObject := awstypes.FleetLaunchTemplateOverridesRequest{}

	if v, ok := tfMap[names.AttrAvailabilityZone].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["instance_requirements"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.InstanceRequirements = expandInstanceRequirementsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrInstanceType].(string); ok && v != "" {
		apiObject.InstanceType = awstypes.InstanceType(v)
	}

	if v, ok := tfMap["image_id"].(string); ok && v != "" {
		apiObject.ImageId = aws.String(v)
	}

	if v, ok := tfMap["max_price"].(string); ok && v != "" {
		apiObject.MaxPrice = aws.String(v)
	}

	if v, ok := tfMap["placement"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Placement = expandPlacement(v.([]interface{})[0].(map[string]interface{}))
	}
	if v, ok := tfMap[names.AttrPriority].(float64); ok && v != 0 {
		apiObject.Priority = aws.Float64(v)
	}

	if v, ok := tfMap[names.AttrSubnetID].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	if v, ok := tfMap["weighted_capacity"].(float64); ok && v != 0 {
		apiObject.WeightedCapacity = aws.Float64(v)
	}

	return apiObject
}

func expandOnDemandOptionsRequest(tfMap map[string]interface{}) *awstypes.OnDemandOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.OnDemandOptionsRequest{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = awstypes.FleetOnDemandAllocationStrategy(v)
	}

	if v, ok := tfMap["capacity_reservation_options"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CapacityReservationOptions = expandCapacityReservationOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["max_total_price"].(string); ok && v != "" {
		apiObject.MaxTotalPrice = aws.String(v)
	}

	if v, ok := tfMap["min_target_capacity"].(int); ok {
		apiObject.MinTargetCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["single_availability_zone"].(bool); ok {
		apiObject.SingleAvailabilityZone = aws.Bool(v)
	}

	if v, ok := tfMap["single_instance_type"].(bool); ok {
		apiObject.SingleInstanceType = aws.Bool(v)
	}

	return apiObject
}

func expandSpotOptionsRequest(tfMap map[string]interface{}) *awstypes.SpotOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SpotOptionsRequest{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = awstypes.SpotAllocationStrategy(v)

		// InvalidFleetConfig: InstancePoolsToUseCount option is only available with the lowestPrice allocation strategy.
		if v == spotAllocationStrategyLowestPrice {
			if v, ok := tfMap["instance_pools_to_use_count"].(int); ok {
				apiObject.InstancePoolsToUseCount = aws.Int32(int32(v))
			}
		}
	}

	if v, ok := tfMap["instance_interruption_behavior"].(string); ok && v != "" {
		apiObject.InstanceInterruptionBehavior = awstypes.SpotInstanceInterruptionBehavior(v)
	}

	if v, ok := tfMap["maintenance_strategies"].([]interface{}); ok && len(v) > 0 {
		apiObject.MaintenanceStrategies = expandFleetSpotMaintenanceStrategiesRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandPlacement(tfMap map[string]interface{}) *awstypes.Placement {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Placement{}

	if v, ok := tfMap["affinity"].(string); ok && v != "" {
		apiObject.Affinity = aws.String(v)
	}

	if v, ok := tfMap[names.AttrAvailabilityZone].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["group_id"].(string); ok && v != "" {
		apiObject.GroupId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrGroupName].(string); ok && v != "" {
		apiObject.GroupName = aws.String(v)
	}

	if v, ok := tfMap["host_id"].(string); ok && v != "" {
		apiObject.HostId = aws.String(v)
	}

	if v, ok := tfMap["host_resource_group_arn"].(string); ok && v != "" {
		apiObject.HostResourceGroupArn = aws.String(v)
	}

	if v, ok := tfMap["partition_number"].(int); ok && v != 0 {
		apiObject.PartitionNumber = aws.Int32(int32(v))
	}

	if v, ok := tfMap["spread_domain"].(string); ok && v != "" {
		apiObject.SpreadDomain = aws.String(v)
	}

	if v, ok := tfMap["tenancy"].(string); ok && v != "" {
		apiObject.Tenancy = awstypes.Tenancy(v)
	}

	return apiObject
}

func expandFleetSpotMaintenanceStrategiesRequest(tfMap map[string]interface{}) *awstypes.FleetSpotMaintenanceStrategiesRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FleetSpotMaintenanceStrategiesRequest{}

	if v, ok := tfMap["capacity_rebalance"].([]interface{}); ok && len(v) > 0 {
		apiObject.CapacityRebalance = expandFleetSpotCapacityRebalanceRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFleetSpotCapacityRebalanceRequest(tfMap map[string]interface{}) *awstypes.FleetSpotCapacityRebalanceRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.FleetSpotCapacityRebalanceRequest{}

	if v, ok := tfMap["replacement_strategy"].(string); ok && v != "" {
		apiObject.ReplacementStrategy = awstypes.FleetReplacementStrategy(v)
	}

	if v, ok := tfMap["termination_delay"].(int); ok {
		apiObject.TerminationDelay = aws.Int32(int32(v))
	}

	return apiObject
}

func expandTargetCapacitySpecificationRequest(tfMap map[string]interface{}) *awstypes.TargetCapacitySpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.TargetCapacitySpecificationRequest{}

	if v, ok := tfMap["default_target_capacity_type"].(string); ok && v != "" {
		apiObject.DefaultTargetCapacityType = awstypes.DefaultTargetCapacityType(v)
	}

	if v, ok := tfMap["on_demand_target_capacity"].(int); ok && v != 0 {
		apiObject.OnDemandTargetCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["spot_target_capacity"].(int); ok && v != 0 {
		apiObject.SpotTargetCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["total_target_capacity"].(int); ok {
		apiObject.TotalTargetCapacity = aws.Int32(int32(v))
	}

	if v, ok := tfMap["target_capacity_unit_type"].(string); ok && v != "" {
		apiObject.TargetCapacityUnitType = awstypes.TargetCapacityUnitType(v)
	}

	return apiObject
}

func flattenCapacityReservationsOptions(apiObject *awstypes.CapacityReservationOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.UsageStrategy; v != "" {
		tfMap["usage_strategy"] = v
	}

	return tfMap
}

func flattenFleetInstances(apiObject awstypes.DescribeFleetsInstances) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.InstanceIds; v != nil {
		tfMap["instance_ids"] = v
	}

	if v := apiObject.InstanceType; v != "" {
		tfMap[names.AttrInstanceType] = v
	}

	if v := apiObject.Lifecycle; v != "" {
		tfMap["lifecycle"] = v
	}

	if v := apiObject.Platform; v != "" {
		tfMap["platform"] = v
	}

	return tfMap
}

func flattenFleetInstanceSet(apiObjects []awstypes.DescribeFleetsInstances) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenFleetInstances(apiObject))
	}

	return tfList
}

func flattenFleetLaunchTemplateConfigs(apiObjects []awstypes.FleetLaunchTemplateConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenFleetLaunchTemplateConfig(apiObject))
	}

	return tfList
}

func flattenFleetLaunchTemplateConfig(apiObject awstypes.FleetLaunchTemplateConfig) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateSpecification; v != nil {
		tfMap["launch_template_specification"] = []interface{}{flattenFleetLaunchTemplateSpecificationForFleet(v)}
	}

	if v := apiObject.Overrides; v != nil {
		tfMap["override"] = flattenFleetLaunchTemplateOverrideses(v)
	}

	return tfMap
}

func flattenFleetLaunchTemplateSpecificationForFleet(apiObject *awstypes.FleetLaunchTemplateSpecification) map[string]interface{} {
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

// Pending AWS to provide this attribute back in the `Describe` call.
// func flattenLaunchTemplatesAndOverridesResponse(apiObject *awstypes.LaunchTemplateAndOverridesResponse) map[string]interface{} {
// 	if apiObject == nil {
// 		return nil
// 	}

// 	tfMap := map[string]interface{}{}

// 	if v := apiObject.LaunchTemplateSpecification; v != nil {
// 		tfMap["launch_template_specification"] = []interface{}{flattenFleetLaunchTemplateSpecificationForFleet(v)}
// 	}

// 	if v := apiObject.Overrides; v != nil {
// 		tfMap["overrides"] = []interface{}{flattenFleetLaunchTemplateOverrides(v)}
// 	}

// 	return tfMap
// }

func flattenFleetLaunchTemplateOverrideses(apiObjects []awstypes.FleetLaunchTemplateOverrides) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenFleetLaunchTemplateOverrides(&apiObject))
	}

	return tfList
}

func flattenFleetLaunchTemplateOverrides(apiObject *awstypes.FleetLaunchTemplateOverrides) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap[names.AttrAvailabilityZone] = aws.ToString(v)
	}

	if v := apiObject.InstanceRequirements; v != nil {
		tfMap["instance_requirements"] = []interface{}{flattenInstanceRequirements(v)}
	}

	if v := apiObject.ImageId; v != nil {
		tfMap["image_id"] = aws.ToString(v)
	}

	if v := apiObject.InstanceType; v != "" {
		tfMap[names.AttrInstanceType] = v
	}

	if v := apiObject.MaxPrice; v != nil {
		tfMap["max_price"] = aws.ToString(v)
	}

	if v := apiObject.Placement; v != nil {
		tfMap["placement"] = []interface{}{flattenPlacement(v)}
	}

	if v := apiObject.Priority; v != nil {
		tfMap[names.AttrPriority] = aws.ToFloat64(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap[names.AttrSubnetID] = aws.ToString(v)
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.ToFloat64(v)
	}

	return tfMap
}

func flattenOnDemandOptions(apiObject *awstypes.OnDemandOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationStrategy; v != "" {
		tfMap["allocation_strategy"] = v
	}

	if v := apiObject.CapacityReservationOptions; v != nil {
		tfMap["capacity_reservation_options"] = []interface{}{flattenCapacityReservationsOptions(v)}
	}

	if v := apiObject.MaxTotalPrice; v != nil {
		tfMap["max_total_price"] = aws.ToString(v)
	}

	if v := apiObject.MinTargetCapacity; v != nil {
		tfMap["min_target_capacity"] = aws.ToInt32(v)
	}

	if v := apiObject.SingleAvailabilityZone; v != nil {
		tfMap["single_availability_zone"] = aws.ToBool(v)
	}

	if v := apiObject.SingleInstanceType; v != nil {
		tfMap["single_instance_type"] = aws.ToBool(v)
	}

	return tfMap
}

func flattenPlacement(apiObject *awstypes.PlacementResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.GroupName; v != nil {
		tfMap[names.AttrGroupName] = aws.ToString(v)
	}

	return tfMap
}

func flattenSpotOptions(apiObject *awstypes.SpotOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationStrategy; v != "" {
		tfMap["allocation_strategy"] = v
	}

	if v := apiObject.InstanceInterruptionBehavior; v != "" {
		tfMap["instance_interruption_behavior"] = v
	}

	if v := apiObject.InstancePoolsToUseCount; v != nil {
		tfMap["instance_pools_to_use_count"] = aws.ToInt32(v)
	} else if apiObject.AllocationStrategy == awstypes.SpotAllocationStrategyDiversified {
		// API will omit InstancePoolsToUseCount if AllocationStrategy is diversified, which breaks our Default: 1
		// Here we just reset it to 1 to prevent removing the Default and setting up a special DiffSuppressFunc.
		tfMap["instance_pools_to_use_count"] = 1
	}

	if v := apiObject.MaintenanceStrategies; v != nil {
		tfMap["maintenance_strategies"] = []interface{}{flattenFleetSpotMaintenanceStrategies(v)}
	}

	return tfMap
}

func flattenFleetSpotMaintenanceStrategies(apiObject *awstypes.FleetSpotMaintenanceStrategies) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CapacityRebalance; v != nil {
		tfMap["capacity_rebalance"] = []interface{}{flattenFleetSpotCapacityRebalance(v)}
	}

	return tfMap
}

func flattenFleetSpotCapacityRebalance(apiObject *awstypes.FleetSpotCapacityRebalance) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ReplacementStrategy; v != "" {
		tfMap["replacement_strategy"] = v
	}

	if v := apiObject.TerminationDelay; v != nil {
		tfMap["termination_delay"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenTargetCapacitySpecification(apiObject *awstypes.TargetCapacitySpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DefaultTargetCapacityType; v != "" {
		tfMap["default_target_capacity_type"] = v
	}

	if v := apiObject.OnDemandTargetCapacity; v != nil {
		tfMap["on_demand_target_capacity"] = aws.ToInt32(v)
	}

	if v := apiObject.SpotTargetCapacity; v != nil {
		tfMap["spot_target_capacity"] = aws.ToInt32(v)
	}

	if v := apiObject.TotalTargetCapacity; v != nil {
		tfMap["total_target_capacity"] = aws.ToInt32(v)
	}

	if v := apiObject.TargetCapacityUnitType; v != "" {
		tfMap["target_capacity_unit_type"] = v
	}

	return tfMap
}
