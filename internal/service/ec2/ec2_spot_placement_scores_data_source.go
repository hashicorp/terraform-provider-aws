package ec2

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceSpotPricementScores() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSpotPlacementScoresRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"instance_requirements_with_metadata": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"architecture_types": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(ec2.ArchitectureType_Values(), false),
							},
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
						"virtualization_types": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(ec2.VirtualizationType_Values(), false),
							},
						},
					},
				},
			},
			"instance_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"max_results": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(10, 1000),
			},
			"region_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidRegionName,
				},
			},
			"single_availability_zone": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"target_capacity": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 2000000000),
			},
			"target_capacity_unit_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ec2.TargetCapacityUnitType_Values(), false),
			},
			"spot_placement_score_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"score": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceSpotPlacementScoresRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.GetSpotPlacementScoresInput{}

	if v, ok := d.GetOk("instance_types"); ok {
		instanceTypes := flex.ExpandStringSet(v.(*schema.Set))
		input.SetInstanceTypes(instanceTypes)
	}

	if v, ok := d.GetOk("max_results"); ok {
		input.SetMaxResults(int64(v.(int)))
	}

	if v, ok := d.GetOk("region_names"); ok {
		regionNames := flex.ExpandStringSet(v.(*schema.Set))
		input.SetRegionNames(regionNames)
	}

	if v, ok := d.GetOk("single_availability_zone"); ok {
		input.SetSingleAvailabilityZone(v.(bool))
	}

	if v, ok := d.GetOk("target_capacity"); ok {
		input.SetTargetCapacity(int64(v.(int)))
	}

	if v, ok := d.GetOk("target_capacity_unit_type"); ok {
		input.SetTargetCapacityUnitType(v.(string))
	}

	if v, ok := d.GetOk("instance_requirements_with_metadata"); ok {
		irwm := expandInstanceRequirementsWithMetadata(v.([]interface{})[0].(map[string]interface{}))
		input.SetInstanceRequirementsWithMetadata(irwm)
	}

	getSpotPlacementScoresOutput, err := conn.GetSpotPlacementScores(input)
	if err != nil {
		return fmt.Errorf("error reading EC2 Spot Placement Scores: %w", err)
	}

	var spotPlacementScores []map[string]interface{}
	for _, spotPlacementScore := range getSpotPlacementScoresOutput.SpotPlacementScores {
		spotPlacementScores = append(
			spotPlacementScores,
			map[string]interface{}{
				"availability_zone_id": spotPlacementScore.AvailabilityZoneId,
				"region":               spotPlacementScore.Region,
				"score":                spotPlacementScore.Score,
			},
		)
	}

	d.Set("spot_placement_score_sets", spotPlacementScores)
	d.SetId(meta.(*conns.AWSClient).Region)

	return nil
}

func expandInstanceRequirementsWithMetadata(tflist map[string]interface{}) *ec2.InstanceRequirementsWithMetadataRequest {
	input := &ec2.InstanceRequirementsWithMetadataRequest{}

	if v, ok := tflist["architecture_types"].(*schema.Set); ok && v.Len() > 0 {
		input.SetArchitectureTypes(flex.ExpandStringSet(v))
	}

	if v, ok := tflist["instance_requirements"]; ok {
		input.SetInstanceRequirements(expandInstanceRequirementsRequest(v.([]interface{})[0].(map[string]interface{})))
	}

	if v, ok := tflist["virtualization_types"].(*schema.Set); ok && v.Len() > 0 {
		input.SetVirtualizationTypes(flex.ExpandStringSet(v))
	}

	return input
}
