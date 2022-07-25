package ec2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceFleetCreate,
		Read:   resourceFleetRead,
		Update: resourceFleetUpdate,
		Delete: resourceFleetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		CustomizeDiff: customdiff.All(
			func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
				if diff.Id() == "" {
					if diff.Get("type").(string) != ec2.FleetTypeMaintain {
						if v, ok := diff.GetOk("spot_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
							tfMap := v.([]interface{})[0].(map[string]interface{})
							if v, ok := tfMap["maintenance_strategies"].([]interface{}); ok && len(v) > 0 {
								return errors.New(`EC2 Fleet has an invalid configuration and can not be created. Capacity Rebalance maintenance strategies can only be specified for fleets of type maintain.`)
							}
						}
					}
				}

				return nil
			},
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"context": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"excess_capacity_termination_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.FleetExcessCapacityTerminationPolicyTermination,
				ValidateFunc: validation.StringInSlice(ec2.FleetExcessCapacityTerminationPolicy_Values(), false),
			},
			"launch_template_config": {
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
									},
									"launch_template_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"version": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"override": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 50,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
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
														ValidateFunc: validation.StringInSlice(ec2.AcceleratorManufacturer_Values(), false),
													},
												},
												"accelerator_names": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.AcceleratorName_Values(), false),
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
														ValidateFunc: validation.StringInSlice(ec2.AcceleratorType_Values(), false),
													},
												},
												"bare_metal": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(ec2.BareMetal_Values(), false),
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
													ValidateFunc: validation.StringInSlice(ec2.BurstablePerformance_Values(), false),
												},
												"cpu_manufacturers": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.CpuManufacturer_Values(), false),
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
														ValidateFunc: validation.StringInSlice(ec2.InstanceGeneration_Values(), false),
													},
												},
												"local_storage": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(ec2.LocalStorage_Values(), false),
												},
												"local_storage_types": {
													Type:     schema.TypeSet,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringInSlice(ec2.LocalStorageType_Values(), false),
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
													Required: true,
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
																Required:     true,
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
													Required: true,
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
																Required:     true,
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
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"priority": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"subnet_id": {
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
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      FleetOnDemandAllocationStrategyLowestPrice,
							ValidateFunc: validation.StringInSlice(FleetOnDemandAllocationStrategy_Values(), false),
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
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      SpotAllocationStrategyLowestPrice,
							ValidateFunc: validation.StringInSlice(SpotAllocationStrategy_Values(), false),
						},
						"instance_interruption_behavior": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							Default:      ec2.SpotInstanceInterruptionBehaviorTerminate,
							ValidateFunc: validation.StringInSlice(ec2.SpotInstanceInterruptionBehavior_Values(), false),
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
													Type:         schema.TypeString,
													Optional:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringInSlice(ec2.FleetReplacementStrategy_Values(), false),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"target_capacity_specification": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_target_capacity_type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(ec2.DefaultTargetCapacityType_Values(), false),
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
								if defaultTargetCapacityTypeO.(string) != ec2.DefaultTargetCapacityTypeOnDemand {
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
								if defaultTargetCapacityTypeO.(string) != ec2.DefaultTargetCapacityTypeSpot {
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
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.FleetTypeMaintain,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.FleetTypeMaintain,
					ec2.FleetTypeRequest,
				}, false),
			},
		},
	}
}

func resourceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	fleetType := d.Get("type").(string)
	input := &ec2.CreateFleetInput{
		ExcessCapacityTerminationPolicy:  aws.String(d.Get("excess_capacity_termination_policy").(string)),
		ReplaceUnhealthyInstances:        aws.Bool(d.Get("replace_unhealthy_instances").(bool)),
		TagSpecifications:                tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeFleet),
		TerminateInstancesWithExpiration: aws.Bool(d.Get("terminate_instances_with_expiration").(bool)),
		Type:                             aws.String(fleetType),
	}

	if v, ok := d.GetOk("context"); ok {
		input.Context = aws.String(v.(string))
	}

	if v, ok := d.GetOk("launch_template_config"); ok && len(v.([]interface{})) > 0 {
		input.LaunchTemplateConfigs = expandFleetLaunchTemplateConfigRequests(v.([]interface{}))
	}

	if v, ok := d.GetOk("on_demand_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OnDemandOptions = expandOnDemandOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("spot_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.SpotOptions = expandSpotOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("target_capacity_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.TargetCapacitySpecification = expandTargetCapacitySpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("context"); ok {
		input.Context = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Fleet: %s", input)
	output, err := conn.CreateFleet(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Fleet: %w", err)
	}

	d.SetId(aws.StringValue(output.FleetId))

	// If a request type is fulfilled immediately, we can miss the transition from active to deleted.
	// Instead of an error here, allow the Read function to trigger recreation.
	targetStates := []string{ec2.FleetStateCodeActive}
	if fleetType == ec2.FleetTypeRequest {
		targetStates = append(targetStates, ec2.FleetStateCodeDeleted)
		targetStates = append(targetStates, ec2.FleetStateCodeDeletedRunning)
		targetStates = append(targetStates, ec2.FleetStateCodeDeletedTerminating)
	}

	if _, err := WaitFleet(conn, d.Id(), []string{ec2.FleetStateCodeSubmitted}, targetStates, d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("waiting for EC2 Fleet (%s) create: %w", d.Id(), err)
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	fleet, err := FindFleetByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Fleet %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EC2 Fleet (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("fleet/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("context", fleet.Context)
	d.Set("excess_capacity_termination_policy", fleet.ExcessCapacityTerminationPolicy)
	if err := d.Set("launch_template_config", flattenFleetLaunchTemplateConfigs(fleet.LaunchTemplateConfigs)); err != nil {
		return fmt.Errorf("setting launch_template_config: %w", err)
	}
	if fleet.OnDemandOptions != nil {
		if err := d.Set("on_demand_options", []interface{}{flattenOnDemandOptions(fleet.OnDemandOptions)}); err != nil {
			return fmt.Errorf("setting on_demand_options: %w", err)
		}
	} else {
		d.Set("on_demand_options", nil)
	}
	d.Set("replace_unhealthy_instances", fleet.ReplaceUnhealthyInstances)
	if fleet.SpotOptions != nil {
		if err := d.Set("spot_options", []interface{}{flattenSpotOptions(fleet.SpotOptions)}); err != nil {
			return fmt.Errorf("setting spot_options: %w", err)
		}
	} else {
		d.Set("spot_options", nil)
	}
	if fleet.TargetCapacitySpecification != nil {
		if err := d.Set("target_capacity_specification", []interface{}{flattenTargetCapacitySpecification(fleet.TargetCapacitySpecification)}); err != nil {
			return fmt.Errorf("setting target_capacity_specification: %w", err)
		}
	} else {
		d.Set("target_capacity_specification", nil)
	}
	d.Set("terminate_instances_with_expiration", fleet.TerminateInstancesWithExpiration)
	d.Set("type", fleet.Type)

	tags := KeyValueTags(fleet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &ec2.ModifyFleetInput{
			Context:                         aws.String(d.Get("context").(string)),
			ExcessCapacityTerminationPolicy: aws.String(d.Get("excess_capacity_termination_policy").(string)),
			LaunchTemplateConfigs:           expandFleetLaunchTemplateConfigRequests(d.Get("launch_template_config").([]interface{})),
			FleetId:                         aws.String(d.Id()),
			// InvalidTargetCapacitySpecification: Currently we only support total target capacity modification.
			// TargetCapacitySpecification: expandEc2TargetCapacitySpecificationRequest(d.Get("target_capacity_specification").([]interface{})),
			TargetCapacitySpecification: &ec2.TargetCapacitySpecificationRequest{
				TotalTargetCapacity: aws.Int64(int64(d.Get("target_capacity_specification.0.total_target_capacity").(int))),
			},
		}

		log.Printf("[DEBUG] Modifying EC2 Fleet: %s", input)
		_, err := conn.ModifyFleet(input)

		if err != nil {
			return fmt.Errorf("modifying EC2 Fleet (%s): %w", d.Id(), err)
		}

		if _, err := WaitFleet(conn, d.Id(), []string{ec2.FleetStateCodeModifying}, []string{ec2.FleetStateCodeActive}, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("waiting for EC2 Fleet (%s) update: %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("updating EC2 Fleet (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Fleet: %s", d.Id())
	output, err := conn.DeleteFleets(&ec2.DeleteFleetsInput{
		FleetIds:           aws.StringSlice([]string{d.Id()}),
		TerminateInstances: aws.Bool(d.Get("terminate_instances").(bool)),
	})

	if err == nil && output != nil {
		err = DeleteFleetsError(output.UnsuccessfulFleetDeletions)
	}

	if tfawserr.ErrCodeEquals(err, errCodeInvalidFleetIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Fleet (%s): %w", d.Id(), err)
	}

	pendingStates := []string{ec2.FleetStateCodeActive}
	targetStates := []string{ec2.FleetStateCodeDeleted}
	if d.Get("terminate_instances").(bool) {
		pendingStates = append(pendingStates, ec2.FleetStateCodeDeletedTerminating)
	} else {
		targetStates = append(targetStates, ec2.FleetStateCodeDeletedRunning)
	}

	if _, err := WaitFleet(conn, d.Id(), pendingStates, targetStates, d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("waiting for EC2 Fleet (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandFleetLaunchTemplateConfigRequests(tfList []interface{}) []*ec2.FleetLaunchTemplateConfigRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.FleetLaunchTemplateConfigRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandFleetLaunchTemplateConfigRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandFleetLaunchTemplateConfigRequest(tfMap map[string]interface{}) *ec2.FleetLaunchTemplateConfigRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.FleetLaunchTemplateConfigRequest{}

	if v, ok := tfMap["launch_template_specification"].([]interface{}); ok && len(v) > 0 {
		apiObject.LaunchTemplateSpecification = expandFleetLaunchTemplateSpecificationRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["override"].([]interface{}); ok && len(v) > 0 {
		apiObject.Overrides = expandFleetLaunchTemplateOverridesRequests(v)
	}

	return apiObject
}

func expandFleetLaunchTemplateSpecificationRequest(tfMap map[string]interface{}) *ec2.FleetLaunchTemplateSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.FleetLaunchTemplateSpecificationRequest{}

	if v, ok := tfMap["launch_template_id"].(string); ok && v != "" {
		apiObject.LaunchTemplateId = aws.String(v)
	}

	if v, ok := tfMap["launch_template_name"].(string); ok && v != "" {
		apiObject.LaunchTemplateName = aws.String(v)
	}

	if v, ok := tfMap["version"].(string); ok && v != "" {
		apiObject.Version = aws.String(v)
	}

	return apiObject
}

func expandFleetLaunchTemplateOverridesRequests(tfList []interface{}) []*ec2.FleetLaunchTemplateOverridesRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.FleetLaunchTemplateOverridesRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandFleetLaunchTemplateOverridesRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandFleetLaunchTemplateOverridesRequest(tfMap map[string]interface{}) *ec2.FleetLaunchTemplateOverridesRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.FleetLaunchTemplateOverridesRequest{}

	if v, ok := tfMap["availability_zone"].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["instance_requirements"]; ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.InstanceRequirements = expandInstanceRequirementsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["instance_type"].(string); ok && v != "" {
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := tfMap["max_price"].(string); ok && v != "" {
		apiObject.MaxPrice = aws.String(v)
	}

	if v, ok := tfMap["priority"].(float64); ok && v != 0 {
		apiObject.Priority = aws.Float64(v)
	}

	if v, ok := tfMap["subnet_id"].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	if v, ok := tfMap["weighted_capacity"].(float64); ok && v != 0 {
		apiObject.WeightedCapacity = aws.Float64(v)
	}

	return apiObject
}

func expandOnDemandOptionsRequest(tfMap map[string]interface{}) *ec2.OnDemandOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.OnDemandOptionsRequest{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = aws.String(v)
	}

	return apiObject
}

func expandSpotOptionsRequest(tfMap map[string]interface{}) *ec2.SpotOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.SpotOptionsRequest{}

	if v, ok := tfMap["allocation_strategy"].(string); ok && v != "" {
		apiObject.AllocationStrategy = aws.String(v)

		// InvalidFleetConfig: InstancePoolsToUseCount option is only available with the lowestPrice allocation strategy.
		if v == SpotAllocationStrategyLowestPrice {
			if v, ok := tfMap["instance_pools_to_use_count"].(int); ok {
				apiObject.InstancePoolsToUseCount = aws.Int64(int64(v))
			}
		}
	}

	if v, ok := tfMap["instance_interruption_behavior"].(string); ok && v != "" {
		apiObject.InstanceInterruptionBehavior = aws.String(v)
	}

	if v, ok := tfMap["maintenance_strategies"].([]interface{}); ok && len(v) > 0 {
		apiObject.MaintenanceStrategies = expandFleetSpotMaintenanceStrategiesRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFleetSpotMaintenanceStrategiesRequest(tfMap map[string]interface{}) *ec2.FleetSpotMaintenanceStrategiesRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.FleetSpotMaintenanceStrategiesRequest{}

	if v, ok := tfMap["capacity_rebalance"].([]interface{}); ok && len(v) > 0 {
		apiObject.CapacityRebalance = expandFleetSpotCapacityRebalanceRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandFleetSpotCapacityRebalanceRequest(tfMap map[string]interface{}) *ec2.FleetSpotCapacityRebalanceRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.FleetSpotCapacityRebalanceRequest{}

	if v, ok := tfMap["replacement_strategy"].(string); ok && v != "" {
		apiObject.ReplacementStrategy = aws.String(v)
	}

	return apiObject
}

func expandTargetCapacitySpecificationRequest(tfMap map[string]interface{}) *ec2.TargetCapacitySpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.TargetCapacitySpecificationRequest{}

	if v, ok := tfMap["default_target_capacity_type"].(string); ok && v != "" {
		apiObject.DefaultTargetCapacityType = aws.String(v)
	}

	if v, ok := tfMap["on_demand_target_capacity"].(int); ok && v != 0 {
		apiObject.OnDemandTargetCapacity = aws.Int64(int64(v))
	}

	if v, ok := tfMap["spot_target_capacity"].(int); ok && v != 0 {
		apiObject.SpotTargetCapacity = aws.Int64(int64(v))
	}

	if v, ok := tfMap["total_target_capacity"].(int); ok {
		apiObject.TotalTargetCapacity = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenFleetLaunchTemplateConfigs(apiObjects []*ec2.FleetLaunchTemplateConfig) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenFleetLaunchTemplateConfig(apiObject))
	}

	return tfList
}

func flattenFleetLaunchTemplateConfig(apiObject *ec2.FleetLaunchTemplateConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LaunchTemplateSpecification; v != nil {
		tfMap["launch_template_specification"] = []interface{}{flattenFleetLaunchTemplateSpecificationForFleet(v)}
	}

	if v := apiObject.Overrides; v != nil {
		tfMap["override"] = flattenFleetLaunchTemplateOverrideses(v)
	}

	return tfMap
}

func flattenFleetLaunchTemplateSpecificationForFleet(apiObject *ec2.FleetLaunchTemplateSpecification) map[string]interface{} {
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

func flattenFleetLaunchTemplateOverrideses(apiObjects []*ec2.FleetLaunchTemplateOverrides) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenFleetLaunchTemplateOverrides(apiObject))
	}

	return tfList
}

func flattenFleetLaunchTemplateOverrides(apiObject *ec2.FleetLaunchTemplateOverrides) map[string]interface{} {
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

	if v := apiObject.MaxPrice; v != nil {
		tfMap["max_price"] = aws.StringValue(v)
	}

	if v := apiObject.Priority; v != nil {
		tfMap["priority"] = aws.Float64Value(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap["subnet_id"] = aws.StringValue(v)
	}

	if v := apiObject.WeightedCapacity; v != nil {
		tfMap["weighted_capacity"] = aws.Float64Value(v)
	}

	return tfMap
}

func flattenOnDemandOptions(apiObject *ec2.OnDemandOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationStrategy; v != nil {
		tfMap["allocation_strategy"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenSpotOptions(apiObject *ec2.SpotOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AllocationStrategy; v != nil {
		tfMap["allocation_strategy"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceInterruptionBehavior; v != nil {
		tfMap["instance_interruption_behavior"] = aws.StringValue(v)
	}

	if v := apiObject.InstancePoolsToUseCount; v != nil {
		tfMap["instance_pools_to_use_count"] = aws.Int64Value(v)
	} else if aws.StringValue(apiObject.AllocationStrategy) == ec2.SpotAllocationStrategyDiversified {
		// API will omit InstancePoolsToUseCount if AllocationStrategy is diversified, which breaks our Default: 1
		// Here we just reset it to 1 to prevent removing the Default and setting up a special DiffSuppressFunc.
		tfMap["instance_pools_to_use_count"] = 1
	}

	if v := apiObject.MaintenanceStrategies; v != nil {
		tfMap["maintenance_strategies"] = []interface{}{flattenFleetSpotMaintenanceStrategies(v)}
	}

	return tfMap
}

func flattenFleetSpotMaintenanceStrategies(apiObject *ec2.FleetSpotMaintenanceStrategies) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CapacityRebalance; v != nil {
		tfMap["capacity_rebalance"] = []interface{}{flattenFleetSpotCapacityRebalance(v)}
	}

	return tfMap
}

func flattenFleetSpotCapacityRebalance(apiObject *ec2.FleetSpotCapacityRebalance) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ReplacementStrategy; v != nil {
		tfMap["replacement_strategy"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenTargetCapacitySpecification(apiObject *ec2.TargetCapacitySpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DefaultTargetCapacityType; v != nil {
		tfMap["default_target_capacity_type"] = aws.StringValue(v)
	}

	if v := apiObject.OnDemandTargetCapacity; v != nil {
		tfMap["on_demand_target_capacity"] = aws.Int64Value(v)
	}

	if v := apiObject.SpotTargetCapacity; v != nil {
		tfMap["spot_target_capacity"] = aws.Int64Value(v)
	}

	if v := apiObject.TotalTargetCapacity; v != nil {
		tfMap["total_target_capacity"] = aws.Int64Value(v)
	}

	return tfMap
}
