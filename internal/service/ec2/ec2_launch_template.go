package ec2

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceLaunchTemplateCreate,
		Read:   resourceLaunchTemplateRead,
		Update: resourceLaunchTemplateUpdate,
		Delete: resourceLaunchTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ebs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delete_on_termination": {
										// Use TypeString to allow an "unspecified" value,
										// since TypeBool only has true/false with false default.
										// The conversion from bare true/false values in
										// configurations to TypeString value is currently safe.
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
										ValidateFunc:     verify.ValidTypeStringNullableBoolean,
									},
									"encrypted": {
										// Use TypeString to allow an "unspecified" value,
										// since TypeBool only has true/false with false default.
										// The conversion from bare true/false values in
										// configurations to TypeString value is currently safe.
										Type:             schema.TypeString,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
										ValidateFunc:     verify.ValidTypeStringNullableBoolean,
									},
									"iops": {
										Type:     schema.TypeInt,
										Computed: true,
										Optional: true,
									},
									"kms_key_id": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"snapshot_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"throughput": {
										Type:         schema.TypeInt,
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.IntBetween(125, 1000),
									},
									"volume_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"volume_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(ec2.VolumeType_Values(), false),
									},
								},
							},
						},
						"no_device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"virtual_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"capacity_reservation_specification": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity_reservation_preference": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.CapacityReservationPreference_Values(), false),
						},
						"capacity_reservation_target": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_reservation_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_resource_group_arn"},
									},
									"capacity_reservation_resource_group_arn": {
										Type:          schema.TypeString,
										Optional:      true,
										ValidateFunc:  verify.ValidARN,
										ConflictsWith: []string{"capacity_reservation_specification.0.capacity_reservation_target.0.capacity_reservation_id"},
									},
								},
							},
						},
					},
				},
			},
			"cpu_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"core_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"threads_per_core": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"credit_specification": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_credits": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(CPUCredits_Values(), false),
						},
					},
				},
			},
			"default_version": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"update_default_version"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"disable_api_stop": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"disable_api_termination": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ebs_optimized": {
				// Use TypeString to allow an "unspecified" value,
				// since TypeBool only has true/false with false default.
				// The conversion from bare true/false values in
				// configurations to TypeString value is currently safe.
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
				ValidateFunc:     verify.ValidTypeStringNullableBoolean,
			},
			"elastic_gpu_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"elastic_inference_accelerator": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"enclave_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"hibernation_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configured": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"iam_instance_profile": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arn": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"iam_instance_profile.0.name"},
							ValidateFunc:  verify.ValidARN,
						},
						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"image_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"instance_initiated_shutdown_behavior": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ec2.ShutdownBehavior_Values(), false),
			},
			"instance_market_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"market_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.MarketType_Values(), false),
						},
						"spot_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_duration_minutes": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntDivisibleBy(60),
									},
									"instance_interruption_behavior": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ec2.InstanceInterruptionBehavior_Values(), false),
									},
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"spot_instance_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(ec2.SpotInstanceType_Values(), false),
									},
									"valid_until": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IsRFC3339Time,
									},
								},
							},
						},
					},
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
				ConflictsWith: []string{"instance_type"},
			},
			"instance_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"instance_requirements"},
			},
			"kernel_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"latest_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"license_specification": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"license_configuration_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"maintenance_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_recovery": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateAutoRecoveryState_Values(), false),
						},
					},
				},
			},
			"metadata_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_endpoint": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateInstanceMetadataEndpointState_Values(), false),
						},
						"http_protocol_ipv6": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ec2.LaunchTemplateInstanceMetadataProtocolIpv6Disabled,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateInstanceMetadataProtocolIpv6_Values(), false),
						},
						"http_put_response_hop_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 64),
						},
						"http_tokens": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateHttpTokensState_Values(), false),
						},
						"instance_metadata_tags": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      ec2.LaunchTemplateInstanceMetadataTagsStateDisabled,
							ValidateFunc: validation.StringInSlice(ec2.LaunchTemplateInstanceMetadataTagsState_Values(), false),
						},
					},
				},
			},
			"monitoring": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  verify.ValidLaunchTemplateName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  verify.ValidLaunchTemplateName,
			},
			"network_interfaces": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"associate_carrier_ip_address": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
							ValidateFunc:     verify.ValidTypeStringNullableBoolean,
						},
						"associate_public_ip_address": {
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
							ValidateFunc:     verify.ValidTypeStringNullableBoolean,
						},
						"delete_on_termination": {
							// Use TypeString to allow an "unspecified" value,
							// since TypeBool only has true/false with false default.
							// The conversion from bare true/false values in
							// configurations to TypeString value is currently safe.
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressEquivalentTypeStringBoolean,
							ValidateFunc:     verify.ValidTypeStringNullableBoolean,
						},
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"interface_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"efa", "interface"}, false),
						},
						"ipv4_address_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ipv4_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPv4Address,
							},
						},
						"ipv4_prefix_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ipv4_prefixes": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
							},
						},
						"ipv6_address_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ipv6_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.IsIPv6Address,
							},
						},
						"ipv6_prefix_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ipv6_prefixes": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
							},
						},
						"network_card_index": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"network_interface_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"private_ip_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv4Address,
						},
						"security_groups": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"placement": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"affinity": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"availability_zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"group_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"host_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"host_resource_group_arn": {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"placement.0.host_id"},
							ValidateFunc:  verify.ValidARN,
						},
						"partition_number": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"spread_domain": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tenancy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.Tenancy_Values(), false),
						},
					},
				},
			},
			"private_dns_name_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_resource_name_dns_aaaa_record": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_resource_name_dns_a_record": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"hostname_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.HostnameType_Values(), false),
						},
					},
				},
			},
			"ram_disk_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_group_names": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"vpc_security_group_ids"},
			},
			"tag_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(ec2.ResourceType_Values(), false),
						},
						"tags": tftags.TagsSchema(),
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"update_default_version": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"default_version"},
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"vpc_security_group_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				ConflictsWith: []string{"security_group_names"},
			},
		},

		// Enable downstream updates for resources referencing schema attributes
		// to prevent non-empty plans after "terraform apply"
		CustomizeDiff: customdiff.Sequence(
			customdiff.ComputedIf("default_version", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				for _, changedKey := range diff.GetChangedKeysPrefix("") {
					switch changedKey {
					case "name", "name_prefix", "description":
						continue
					default:
						return diff.Get("update_default_version").(bool)
					}
				}
				return false
			}),
			customdiff.ComputedIf("latest_version", func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) bool {
				for _, changedKey := range diff.GetChangedKeysPrefix("") {
					switch changedKey {
					case "name", "name_prefix", "description", "default_version", "update_default_version":
						continue
					default:
						return true
					}
				}
				return false
			}),
			verify.SetTagsDiff,
		),
	}
}

func resourceLaunchTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &ec2.CreateLaunchTemplateInput{
		ClientToken:        aws.String(resource.UniqueId()),
		LaunchTemplateName: aws.String(name),
		TagSpecifications:  tagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeLaunchTemplate),
	}

	if v, ok := d.GetOk("description"); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	if v, err := expandRequestLaunchTemplateData(conn, d); err == nil {
		input.LaunchTemplateData = v
	} else {
		return err
	}

	log.Printf("[DEBUG] Creating EC2 Launch Template: %s", input)
	output, err := conn.CreateLaunchTemplate(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Launch Template: %w", err)
	}

	d.SetId(aws.StringValue(output.LaunchTemplate.LaunchTemplateId))

	return resourceLaunchTemplateRead(d, meta)
}

func resourceLaunchTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	lt, err := FindLaunchTemplateByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Launch Template %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Launch Template (%s): %w", d.Id(), err)
	}

	version := strconv.FormatInt(aws.Int64Value(lt.LatestVersionNumber), 10)
	ltv, err := FindLaunchTemplateVersionByTwoPartKey(conn, d.Id(), version)

	if err != nil {
		return fmt.Errorf("error reading EC2 Launch Template (%s) Version (%s): %w", d.Id(), version, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("launch-template/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("default_version", lt.DefaultVersionNumber)
	d.Set("description", ltv.VersionDescription)
	d.Set("latest_version", lt.LatestVersionNumber)
	d.Set("name", lt.LaunchTemplateName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(lt.LaunchTemplateName)))

	if err := flattenResponseLaunchTemplateData(conn, d, ltv.LaunchTemplateData); err != nil {
		return err
	}

	tags := KeyValueTags(lt.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceLaunchTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	updateKeys := []string{
		"block_device_mappings",
		"capacity_reservation_specification",
		"cpu_options",
		"credit_specification",
		"description",
		"disable_api_stop",
		"disable_api_termination",
		"ebs_optimized",
		"elastic_gpu_specifications",
		"elastic_inference_accelerator",
		"enclave_options",
		"hibernation_options",
		"iam_instance_profile",
		"image_id",
		"instance_initiated_shutdown_behavior",
		"instance_market_options",
		"instance_requirements",
		"instance_type",
		"kernel_id",
		"key_name",
		"license_specification",
		"metadata_options",
		"monitoring",
		"network_interfaces",
		"placement",
		"private_dns_name_options",
		"ram_disk_id",
		"security_group_names",
		"tag_specifications",
		"user_data",
		"vpc_security_group_ids",
	}
	latestVersion := int64(d.Get("latest_version").(int))

	if d.HasChanges(updateKeys...) {
		input := &ec2.CreateLaunchTemplateVersionInput{
			ClientToken:      aws.String(resource.UniqueId()),
			LaunchTemplateId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("description"); ok {
			input.VersionDescription = aws.String(v.(string))
		}

		if v, err := expandRequestLaunchTemplateData(conn, d); err == nil {
			input.LaunchTemplateData = v
		} else {
			return err
		}

		output, err := conn.CreateLaunchTemplateVersion(input)

		if err != nil {
			return fmt.Errorf("error creating EC2 Launch Template (%s) Version: %w", d.Id(), err)
		}

		latestVersion = aws.Int64Value(output.LaunchTemplateVersion.VersionNumber)

	}

	if d.Get("update_default_version").(bool) || d.HasChange("default_version") {
		input := &ec2.ModifyLaunchTemplateInput{
			LaunchTemplateId: aws.String(d.Id()),
		}

		if d.Get("update_default_version").(bool) {
			input.DefaultVersion = aws.String(strconv.FormatInt(latestVersion, 10))
		} else if d.HasChange("default_version") {
			input.DefaultVersion = aws.String(strconv.Itoa(d.Get("default_version").(int)))
		}

		_, err := conn.ModifyLaunchTemplate(input)

		if err != nil {
			return fmt.Errorf("error updating EC2 Launch Template (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Launch Template (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceLaunchTemplateRead(d, meta)
}

func resourceLaunchTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[DEBUG] Deleting EC2 Launch Template: %s", d.Id())
	_, err := conn.DeleteLaunchTemplate(&ec2.DeleteLaunchTemplateInput{
		LaunchTemplateId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidLaunchTemplateIdNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Launch Template (%s): %w", d.Id(), err)
	}

	return nil
}

func expandRequestLaunchTemplateData(conn *ec2.EC2, d *schema.ResourceData) (*ec2.RequestLaunchTemplateData, error) {
	apiObject := &ec2.RequestLaunchTemplateData{
		// Always set at least one field.
		UserData: aws.String(d.Get("user_data").(string)),
	}

	var instanceType string
	if v, ok := d.GetOk("instance_type"); ok {
		v := v.(string)

		instanceType = v
		apiObject.InstanceType = aws.String(v)
	}

	if v, ok := d.GetOk("block_device_mappings"); ok && len(v.([]interface{})) > 0 {
		apiObject.BlockDeviceMappings = expandLaunchTemplateBlockDeviceMappingRequests(v.([]interface{}))
	}

	if v, ok := d.GetOk("capacity_reservation_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CapacityReservationSpecification = expandLaunchTemplateCapacityReservationSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("cpu_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.CpuOptions = expandLaunchTemplateCPUOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("credit_specification"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		if instanceType != "" {
			instanceTypeInfo, err := FindInstanceTypeByName(conn, instanceType)

			if err != nil {
				return nil, fmt.Errorf("reading EC2 Instance Type (%s): %w", instanceType, err)
			}

			if aws.BoolValue(instanceTypeInfo.BurstablePerformanceSupported) {
				apiObject.CreditSpecification = expandCreditSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
			}
		}
	}

	if v, ok := d.GetOk("disable_api_stop"); ok {
		apiObject.DisableApiStop = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("disable_api_termination"); ok {
		apiObject.DisableApiTermination = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("ebs_optimized"); ok {
		v, _ := strconv.ParseBool(v.(string))

		apiObject.EbsOptimized = aws.Bool(v)
	}

	if v, ok := d.GetOk("elastic_gpu_specifications"); ok && len(v.([]interface{})) > 0 {
		apiObject.ElasticGpuSpecifications = expandElasticGpuSpecifications(v.([]interface{}))
	}

	if v, ok := d.GetOk("elastic_inference_accelerator"); ok && len(v.([]interface{})) > 0 {
		apiObject.ElasticInferenceAccelerators = expandLaunchTemplateElasticInferenceAccelerators(v.([]interface{}))
	}

	if v, ok := d.GetOk("enclave_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		apiObject.EnclaveOptions = &ec2.LaunchTemplateEnclaveOptionsRequest{
			Enabled: aws.Bool(tfMap["enabled"].(bool)),
		}
	}

	if v, ok := d.GetOk("hibernation_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		apiObject.HibernationOptions = &ec2.LaunchTemplateHibernationOptionsRequest{
			Configured: aws.Bool(tfMap["configured"].(bool)),
		}
	}

	if v, ok := d.GetOk("iam_instance_profile"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.IamInstanceProfile = expandLaunchTemplateIAMInstanceProfileSpecificationRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("image_id"); ok {
		apiObject.ImageId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_initiated_shutdown_behavior"); ok {
		apiObject.InstanceInitiatedShutdownBehavior = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_market_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.InstanceMarketOptions = expandLaunchTemplateInstanceMarketOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("instance_requirements"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.InstanceRequirements = expandInstanceRequirementsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("kernel_id"); ok {
		apiObject.KernelId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("key_name"); ok {
		apiObject.KeyName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("license_specification"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.LicenseSpecifications = expandLaunchTemplateLicenseConfigurationRequests(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("maintenance_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.MaintenanceOptions = expandLaunchTemplateInstanceMaintenanceOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("metadata_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.MetadataOptions = expandLaunchTemplateInstanceMetadataOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("monitoring"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		apiObject.Monitoring = &ec2.LaunchTemplatesMonitoringRequest{
			Enabled: aws.Bool(tfMap["enabled"].(bool)),
		}
	}

	if v, ok := d.GetOk("network_interfaces"); ok && len(v.([]interface{})) > 0 {
		apiObject.NetworkInterfaces = expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequests(v.([]interface{}))
	}

	if v, ok := d.GetOk("placement"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.Placement = expandLaunchTemplatePlacementRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("private_dns_name_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		apiObject.PrivateDnsNameOptions = expandLaunchTemplatePrivateDNSNameOptionsRequest(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("ram_disk_id"); ok {
		apiObject.RamDiskId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_group_names"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_specifications"); ok && len(v.([]interface{})) > 0 {
		apiObject.TagSpecifications = expandLaunchTemplateTagSpecificationRequests(v.([]interface{}))
	}

	if v, ok := d.GetOk("vpc_security_group_ids"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	return apiObject, nil
}

func expandLaunchTemplateBlockDeviceMappingRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateBlockDeviceMappingRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateBlockDeviceMappingRequest{}

	if v, ok := tfMap["ebs"].([]interface{}); ok && len(v) > 0 {
		apiObject.Ebs = expandLaunchTemplateEBSBlockDeviceRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["device_name"].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["no_device"].(string); ok && v != "" {
		apiObject.NoDevice = aws.String(v)
	}

	if v, ok := tfMap["virtual_name"].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateBlockDeviceMappingRequests(tfList []interface{}) []*ec2.LaunchTemplateBlockDeviceMappingRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateBlockDeviceMappingRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateBlockDeviceMappingRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateEBSBlockDeviceRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateEbsBlockDeviceRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateEbsBlockDeviceRequest{}

	if v, ok := tfMap["delete_on_termination"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap["encrypted"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap["iops"].(int); ok && v != 0 {
		apiObject.Iops = aws.Int64(int64(v))
	}

	if v, ok := tfMap["kms_key_id"].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap["snapshot_id"].(string); ok && v != "" {
		apiObject.SnapshotId = aws.String(v)
	}

	if v, ok := tfMap["throughput"].(int); ok && v != 0 {
		apiObject.Throughput = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_size"].(int); ok && v != 0 {
		apiObject.VolumeSize = aws.Int64(int64(v))
	}

	if v, ok := tfMap["volume_type"].(string); ok && v != "" {
		apiObject.VolumeType = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateCapacityReservationSpecificationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateCapacityReservationSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateCapacityReservationSpecificationRequest{}

	if v, ok := tfMap["capacity_reservation_preference"].(string); ok && v != "" {
		apiObject.CapacityReservationPreference = aws.String(v)
	}

	if v, ok := tfMap["capacity_reservation_target"].([]interface{}); ok && len(v) > 0 {
		apiObject.CapacityReservationTarget = expandCapacityReservationTarget(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandLaunchTemplateCPUOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateCpuOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateCpuOptionsRequest{}

	if v, ok := tfMap["core_count"].(int); ok && v != 0 {
		apiObject.CoreCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["threads_per_core"].(int); ok && v != 0 {
		apiObject.ThreadsPerCore = aws.Int64(int64(v))
	}

	return apiObject
}

func expandElasticGpuSpecification(tfMap map[string]interface{}) *ec2.ElasticGpuSpecification {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.ElasticGpuSpecification{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandElasticGpuSpecifications(tfList []interface{}) []*ec2.ElasticGpuSpecification {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.ElasticGpuSpecification

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandElasticGpuSpecification(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateElasticInferenceAccelerator(tfMap map[string]interface{}) *ec2.LaunchTemplateElasticInferenceAccelerator {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateElasticInferenceAccelerator{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateElasticInferenceAccelerators(tfList []interface{}) []*ec2.LaunchTemplateElasticInferenceAccelerator {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateElasticInferenceAccelerator

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateElasticInferenceAccelerator(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateIAMInstanceProfileSpecificationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateIamInstanceProfileSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateIamInstanceProfileSpecificationRequest{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceMarketOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateInstanceMarketOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateInstanceMarketOptionsRequest{}

	if v, ok := tfMap["market_type"].(string); ok && v != "" {
		apiObject.MarketType = aws.String(v)
	}

	if v, ok := tfMap["spot_options"].([]interface{}); ok && len(v) > 0 {
		apiObject.SpotOptions = expandLaunchTemplateSpotMarketOptionsRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandInstanceRequirementsRequest(tfMap map[string]interface{}) *ec2.InstanceRequirementsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.InstanceRequirementsRequest{}

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
		apiObject.VCpuCount = expandVCPUCountRangeRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandAcceleratorCountRequest(tfMap map[string]interface{}) *ec2.AcceleratorCountRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.AcceleratorCountRequest{}

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

func expandAcceleratorTotalMemoryMiBRequest(tfMap map[string]interface{}) *ec2.AcceleratorTotalMemoryMiBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.AcceleratorTotalMemoryMiBRequest{}

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

func expandBaselineEBSBandwidthMbpsRequest(tfMap map[string]interface{}) *ec2.BaselineEbsBandwidthMbpsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.BaselineEbsBandwidthMbpsRequest{}

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

func expandMemoryGiBPerVCPURequest(tfMap map[string]interface{}) *ec2.MemoryGiBPerVCpuRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.MemoryGiBPerVCpuRequest{}

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

func expandMemoryMiBRequest(tfMap map[string]interface{}) *ec2.MemoryMiBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.MemoryMiBRequest{}

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

func expandNetworkInterfaceCountRequest(tfMap map[string]interface{}) *ec2.NetworkInterfaceCountRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.NetworkInterfaceCountRequest{}

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

func expandTotalLocalStorageGBRequest(tfMap map[string]interface{}) *ec2.TotalLocalStorageGBRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.TotalLocalStorageGBRequest{}

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

func expandVCPUCountRangeRequest(tfMap map[string]interface{}) *ec2.VCpuCountRangeRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.VCpuCountRangeRequest{}

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

func expandLaunchTemplateSpotMarketOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateSpotMarketOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateSpotMarketOptionsRequest{}

	if v, ok := tfMap["block_duration_minutes"].(int); ok && v != 0 {
		apiObject.BlockDurationMinutes = aws.Int64(int64(v))
	}

	if v, ok := tfMap["instance_interruption_behavior"].(string); ok && v != "" {
		apiObject.InstanceInterruptionBehavior = aws.String(v)
	}

	if v, ok := tfMap["max_price"].(string); ok && v != "" {
		apiObject.MaxPrice = aws.String(v)
	}

	if v, ok := tfMap["spot_instance_type"].(string); ok && v != "" {
		apiObject.SpotInstanceType = aws.String(v)
	}

	if v, ok := tfMap["valid_until"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.ValidUntil = aws.Time(v)
	}

	return apiObject
}

func expandLaunchTemplateLicenseConfigurationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateLicenseConfigurationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateLicenseConfigurationRequest{}

	if v, ok := tfMap["license_configuration_arn"].(string); ok && v != "" {
		apiObject.LicenseConfigurationArn = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateLicenseConfigurationRequests(tfList []interface{}) []*ec2.LaunchTemplateLicenseConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateLicenseConfigurationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateLicenseConfigurationRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplateInstanceMetadataOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateInstanceMetadataOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateInstanceMetadataOptionsRequest{}

	if v, ok := tfMap["http_endpoint"].(string); ok && v != "" {
		apiObject.HttpEndpoint = aws.String(v)

		if v == ec2.LaunchTemplateInstanceMetadataEndpointStateEnabled {
			// These parameters are not allowed unless HttpEndpoint is enabled.
			if v, ok := tfMap["http_tokens"].(string); ok && v != "" {
				apiObject.HttpTokens = aws.String(v)
			}

			if v, ok := tfMap["http_put_response_hop_limit"].(int); ok && v != 0 {
				apiObject.HttpPutResponseHopLimit = aws.Int64(int64(v))
			}

			if v, ok := tfMap["instance_metadata_tags"].(string); ok && v != "" {
				apiObject.InstanceMetadataTags = aws.String(v)
			}
		}
	}

	if v, ok := tfMap["http_protocol_ipv6"].(string); ok && v != "" {
		apiObject.HttpProtocolIpv6 = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceMaintenanceOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateInstanceMaintenanceOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateInstanceMaintenanceOptionsRequest{}

	if v, ok := tfMap["auto_recovery"].(string); ok && v != "" {
		apiObject.AutoRecovery = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}

	if v, ok := tfMap["associate_carrier_ip_address"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.AssociateCarrierIpAddress = aws.Bool(v)
	}

	if v, ok := tfMap["associate_public_ip_address"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.AssociatePublicIpAddress = aws.Bool(v)
	}

	if v, ok := tfMap["delete_on_termination"].(string); ok && v != "" {
		v, _ := strconv.ParseBool(v)

		apiObject.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["device_index"].(int); ok {
		apiObject.DeviceIndex = aws.Int64(int64(v))
	}

	if v, ok := tfMap["interface_type"].(string); ok && v != "" {
		apiObject.InterfaceType = aws.String(v)
	}

	var privateIPAddress string

	if v, ok := tfMap["private_ip_address"].(string); ok && v != "" {
		privateIPAddress = v
		apiObject.PrivateIpAddress = aws.String(v)
	}

	if v, ok := tfMap["ipv4_address_count"].(int); ok && v != 0 {
		apiObject.SecondaryPrivateIpAddressCount = aws.Int64(int64(v))
	} else if v, ok := tfMap["ipv4_addresses"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			v := v.(string)

			apiObject.PrivateIpAddresses = append(apiObject.PrivateIpAddresses, &ec2.PrivateIpAddressSpecification{
				Primary:          aws.Bool(v == privateIPAddress),
				PrivateIpAddress: aws.String(v),
			})
		}
	}

	if v, ok := tfMap["ipv4_prefix_count"].(int); ok && v != 0 {
		apiObject.Ipv4PrefixCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["ipv4_prefixes"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Ipv4Prefixes = expandIPv4PrefixSpecificationRequests(v.List())
	}

	if v, ok := tfMap["ipv6_address_count"].(int); ok && v != 0 {
		apiObject.Ipv6AddressCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["ipv6_addresses"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.Ipv6Addresses = append(apiObject.Ipv6Addresses, &ec2.InstanceIpv6AddressRequest{
				Ipv6Address: aws.String(v.(string)),
			})
		}
	}

	if v, ok := tfMap["ipv6_prefix_count"].(int); ok && v != 0 {
		apiObject.Ipv6PrefixCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["ipv6_prefixes"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Ipv6Prefixes = expandIPv6PrefixSpecificationRequests(v.List())
	}

	if v, ok := tfMap["network_card_index"].(int); ok {
		apiObject.NetworkCardIndex = aws.Int64(int64(v))
	}

	if v, ok := tfMap["network_interface_id"].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, ok := tfMap["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.Groups = append(apiObject.Groups, aws.String(v.(string)))
		}
	}

	if v, ok := tfMap["subnet_id"].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequests(tfList []interface{}) []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandLaunchTemplatePlacementRequest(tfMap map[string]interface{}) *ec2.LaunchTemplatePlacementRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplatePlacementRequest{}

	if v, ok := tfMap["affinity"].(string); ok && v != "" {
		apiObject.Affinity = aws.String(v)
	}

	if v, ok := tfMap["availability_zone"].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
	}

	if v, ok := tfMap["group_name"].(string); ok && v != "" {
		apiObject.GroupName = aws.String(v)
	}

	if v, ok := tfMap["host_id"].(string); ok && v != "" {
		apiObject.HostId = aws.String(v)
	}

	if v, ok := tfMap["host_resource_group_arn"].(string); ok && v != "" {
		apiObject.HostResourceGroupArn = aws.String(v)
	}

	if v, ok := tfMap["partition_number"].(int); ok && v != 0 {
		apiObject.PartitionNumber = aws.Int64(int64(v))
	}

	if v, ok := tfMap["spread_domain"].(string); ok && v != "" {
		apiObject.SpreadDomain = aws.String(v)
	}

	if v, ok := tfMap["tenancy"].(string); ok && v != "" {
		apiObject.Tenancy = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplatePrivateDNSNameOptionsRequest(tfMap map[string]interface{}) *ec2.LaunchTemplatePrivateDnsNameOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplatePrivateDnsNameOptionsRequest{
		EnableResourceNameDnsAAAARecord: aws.Bool(tfMap["enable_resource_name_dns_aaaa_record"].(bool)),
		EnableResourceNameDnsARecord:    aws.Bool(tfMap["enable_resource_name_dns_a_record"].(bool)),
	}

	if v, ok := tfMap["hostname_type"].(string); ok && v != "" {
		apiObject.HostnameType = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateTagSpecificationRequest(tfMap map[string]interface{}) *ec2.LaunchTemplateTagSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &ec2.LaunchTemplateTagSpecificationRequest{}

	if v, ok := tfMap["resource_type"].(string); ok && v != "" {
		apiObject.ResourceType = aws.String(v)
	}

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && len(v) > 0 {
		if v := tftags.New(v).IgnoreAWS(); len(v) > 0 {
			apiObject.Tags = Tags(v)
		}
	}

	return apiObject
}

func expandLaunchTemplateTagSpecificationRequests(tfList []interface{}) []*ec2.LaunchTemplateTagSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*ec2.LaunchTemplateTagSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandLaunchTemplateTagSpecificationRequest(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenResponseLaunchTemplateData(conn *ec2.EC2, d *schema.ResourceData, apiObject *ec2.ResponseLaunchTemplateData) error {
	instanceType := aws.StringValue(apiObject.InstanceType)

	if err := d.Set("block_device_mappings", flattenLaunchTemplateBlockDeviceMappings(apiObject.BlockDeviceMappings)); err != nil {
		return fmt.Errorf("error setting block_device_mappings: %w", err)
	}
	if apiObject.CapacityReservationSpecification != nil {
		if err := d.Set("capacity_reservation_specification", []interface{}{flattenLaunchTemplateCapacityReservationSpecificationResponse(apiObject.CapacityReservationSpecification)}); err != nil {
			return fmt.Errorf("error setting capacity_reservation_specification: %w", err)
		}
	} else {
		d.Set("capacity_reservation_specification", nil)
	}
	if apiObject.CpuOptions != nil {
		if err := d.Set("cpu_options", []interface{}{flattenLaunchTemplateCPUOptions(apiObject.CpuOptions)}); err != nil {
			return fmt.Errorf("error setting cpu_options: %w", err)
		}
	} else {
		d.Set("cpu_options", nil)
	}
	if apiObject.CreditSpecification != nil && instanceType != "" {
		instanceTypeInfo, err := FindInstanceTypeByName(conn, instanceType)

		if err != nil {
			return fmt.Errorf("reading EC2 Instance Type (%s): %w", instanceType, err)
		}

		if aws.BoolValue(instanceTypeInfo.BurstablePerformanceSupported) {
			if err := d.Set("credit_specification", []interface{}{flattenCreditSpecification(apiObject.CreditSpecification)}); err != nil {
				return fmt.Errorf("error setting credit_specification: %w", err)
			}
		}
	} // Don't overwrite any configured value.
	d.Set("disable_api_stop", apiObject.DisableApiStop)
	d.Set("disable_api_termination", apiObject.DisableApiTermination)
	if apiObject.EbsOptimized != nil {
		d.Set("ebs_optimized", strconv.FormatBool(aws.BoolValue(apiObject.EbsOptimized)))
	} else {
		d.Set("ebs_optimized", "")
	}
	if err := d.Set("elastic_gpu_specifications", flattenElasticGpuSpecificationResponses(apiObject.ElasticGpuSpecifications)); err != nil {
		return fmt.Errorf("error setting elastic_gpu_specifications: %w", err)
	}
	if err := d.Set("elastic_inference_accelerator", flattenLaunchTemplateElasticInferenceAcceleratorResponses(apiObject.ElasticInferenceAccelerators)); err != nil {
		return fmt.Errorf("error setting elastic_inference_accelerator: %w", err)
	}
	if apiObject.EnclaveOptions != nil {
		tfMap := map[string]interface{}{
			"enabled": aws.BoolValue(apiObject.EnclaveOptions.Enabled),
		}

		if err := d.Set("enclave_options", []interface{}{tfMap}); err != nil {
			return fmt.Errorf("error setting enclave_options: %w", err)
		}
	} else {
		d.Set("enclave_options", nil)
	}
	if apiObject.HibernationOptions != nil {
		tfMap := map[string]interface{}{
			"configured": aws.BoolValue(apiObject.HibernationOptions.Configured),
		}

		if err := d.Set("hibernation_options", []interface{}{tfMap}); err != nil {
			return fmt.Errorf("error setting hibernation_options: %w", err)
		}
	} else {
		d.Set("hibernation_options", nil)
	}
	if apiObject.IamInstanceProfile != nil {
		if err := d.Set("iam_instance_profile", []interface{}{flattenLaunchTemplateIAMInstanceProfileSpecification(apiObject.IamInstanceProfile)}); err != nil {
			return fmt.Errorf("error setting iam_instance_profile: %w", err)
		}
	} else {
		d.Set("iam_instance_profile", nil)
	}
	d.Set("image_id", apiObject.ImageId)
	d.Set("instance_initiated_shutdown_behavior", apiObject.InstanceInitiatedShutdownBehavior)
	if apiObject.InstanceMarketOptions != nil {
		if err := d.Set("instance_market_options", []interface{}{flattenLaunchTemplateInstanceMarketOptions(apiObject.InstanceMarketOptions)}); err != nil {
			return fmt.Errorf("error setting instance_market_options: %w", err)
		}
	} else {
		d.Set("instance_market_options", nil)
	}
	if apiObject.InstanceRequirements != nil {
		if err := d.Set("instance_requirements", []interface{}{flattenInstanceRequirements(apiObject.InstanceRequirements)}); err != nil {
			return fmt.Errorf("error setting instance_requirements: %w", err)
		}
	} else {
		d.Set("instance_requirements", nil)
	}
	d.Set("instance_type", instanceType)
	d.Set("kernel_id", apiObject.KernelId)
	d.Set("key_name", apiObject.KeyName)
	if err := d.Set("license_specification", flattenLaunchTemplateLicenseConfigurations(apiObject.LicenseSpecifications)); err != nil {
		return fmt.Errorf("error setting license_specification: %w", err)
	}
	if apiObject.MaintenanceOptions != nil {
		if err := d.Set("maintenance_options", []interface{}{flattenLaunchTemplateInstanceMaintenanceOptions(apiObject.MaintenanceOptions)}); err != nil {
			return fmt.Errorf("error setting maintenance_options: %w", err)
		}
	} else {
		d.Set("maintenance_options", nil)
	}
	if apiObject.MetadataOptions != nil {
		if err := d.Set("metadata_options", []interface{}{flattenLaunchTemplateInstanceMetadataOptions(apiObject.MetadataOptions)}); err != nil {
			return fmt.Errorf("error setting metadata_options: %w", err)
		}
	} else {
		d.Set("metadata_options", nil)
	}
	if apiObject.Monitoring != nil {
		tfMap := map[string]interface{}{
			"enabled": aws.BoolValue(apiObject.Monitoring.Enabled),
		}

		if err := d.Set("monitoring", []interface{}{tfMap}); err != nil {
			return fmt.Errorf("error setting monitoring: %w", err)
		}
	} else {
		d.Set("monitoring", nil)
	}
	if err := d.Set("network_interfaces", flattenLaunchTemplateInstanceNetworkInterfaceSpecifications(apiObject.NetworkInterfaces)); err != nil {
		return fmt.Errorf("error setting network_interfaces: %w", err)
	}
	if apiObject.Placement != nil {
		if err := d.Set("placement", []interface{}{flattenLaunchTemplatePlacement(apiObject.Placement)}); err != nil {
			return fmt.Errorf("error setting placement: %w", err)
		}
	} else {
		d.Set("placement", nil)
	}
	if apiObject.PrivateDnsNameOptions != nil {
		if err := d.Set("private_dns_name_options", []interface{}{flattenLaunchTemplatePrivateDNSNameOptions(apiObject.PrivateDnsNameOptions)}); err != nil {
			return fmt.Errorf("error setting private_dns_name_options: %w", err)
		}
	} else {
		d.Set("private_dns_name_options", nil)
	}
	d.Set("ram_disk_id", apiObject.RamDiskId)
	d.Set("security_group_names", aws.StringValueSlice(apiObject.SecurityGroups))
	if err := d.Set("tag_specifications", flattenLaunchTemplateTagSpecifications(apiObject.TagSpecifications)); err != nil {
		return fmt.Errorf("error setting tag_specifications: %w", err)
	}
	d.Set("user_data", apiObject.UserData)
	d.Set("vpc_security_group_ids", aws.StringValueSlice(apiObject.SecurityGroupIds))

	return nil
}

func flattenLaunchTemplateBlockDeviceMapping(apiObject *ec2.LaunchTemplateBlockDeviceMapping) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeviceName; v != nil {
		tfMap["device_name"] = aws.StringValue(v)
	}

	if v := apiObject.Ebs; v != nil {
		tfMap["ebs"] = []interface{}{flattenLaunchTemplateEBSBlockDevice(v)}
	}

	if v := apiObject.NoDevice; v != nil {
		tfMap["no_device"] = aws.StringValue(v)
	}

	if v := apiObject.VirtualName; v != nil {
		tfMap["virtual_name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateBlockDeviceMappings(apiObjects []*ec2.LaunchTemplateBlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchTemplateBlockDeviceMapping(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateEBSBlockDevice(apiObject *ec2.LaunchTemplateEbsBlockDevice) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeleteOnTermination; v != nil {
		tfMap["delete_on_termination"] = strconv.FormatBool(aws.BoolValue(v))
	}

	if v := apiObject.Encrypted; v != nil {
		tfMap["encrypted"] = strconv.FormatBool(aws.BoolValue(v))
	}

	if v := apiObject.Iops; v != nil {
		tfMap["iops"] = aws.Int64Value(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap["kms_key_id"] = aws.StringValue(v)
	}

	if v := apiObject.SnapshotId; v != nil {
		tfMap["snapshot_id"] = aws.StringValue(v)
	}

	if v := apiObject.Throughput; v != nil {
		tfMap["throughput"] = aws.Int64Value(v)
	}

	if v := apiObject.VolumeSize; v != nil {
		tfMap["volume_size"] = aws.Int64Value(v)
	}

	if v := apiObject.VolumeType; v != nil {
		tfMap["volume_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateCapacityReservationSpecificationResponse(apiObject *ec2.LaunchTemplateCapacityReservationSpecificationResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CapacityReservationPreference; v != nil {
		tfMap["capacity_reservation_preference"] = aws.StringValue(v)
	}

	if v := apiObject.CapacityReservationTarget; v != nil {
		tfMap["capacity_reservation_target"] = []interface{}{flattenCapacityReservationTargetResponse(v)}
	}

	return tfMap
}

func flattenLaunchTemplateCPUOptions(apiObject *ec2.LaunchTemplateCpuOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CoreCount; v != nil {
		tfMap["core_count"] = aws.Int64Value(v)
	}

	if v := apiObject.ThreadsPerCore; v != nil {
		tfMap["threads_per_core"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenCreditSpecification(apiObject *ec2.CreditSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CpuCredits; v != nil {
		tfMap["cpu_credits"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenElasticGpuSpecificationResponse(apiObject *ec2.ElasticGpuSpecificationResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenElasticGpuSpecificationResponses(apiObjects []*ec2.ElasticGpuSpecificationResponse) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenElasticGpuSpecificationResponse(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateElasticInferenceAcceleratorResponse(apiObject *ec2.LaunchTemplateElasticInferenceAcceleratorResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateElasticInferenceAcceleratorResponses(apiObjects []*ec2.LaunchTemplateElasticInferenceAcceleratorResponse) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchTemplateElasticInferenceAcceleratorResponse(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateIAMInstanceProfileSpecification(apiObject *ec2.LaunchTemplateIamInstanceProfileSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		tfMap["arn"] = aws.StringValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateInstanceMarketOptions(apiObject *ec2.LaunchTemplateInstanceMarketOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MarketType; v != nil {
		tfMap["market_type"] = aws.StringValue(v)
	}

	if v := apiObject.SpotOptions; v != nil {
		tfMap["spot_options"] = []interface{}{flattenLaunchTemplateSpotMarketOptions(v)}
	}

	return tfMap
}

func flattenInstanceRequirements(apiObject *ec2.InstanceRequirements) map[string]interface{} {
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
		tfMap["total_local_storage_gb"] = []interface{}{flattenTotalLocalStorageGB(v)}
	}

	if v := apiObject.VCpuCount; v != nil {
		tfMap["vcpu_count"] = []interface{}{flattenVCPUCountRange(v)}
	}

	return tfMap
}

func flattenAcceleratorCount(apiObject *ec2.AcceleratorCount) map[string]interface{} {
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

func flattenAcceleratorTotalMemoryMiB(apiObject *ec2.AcceleratorTotalMemoryMiB) map[string]interface{} {
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

func flattenBaselineEBSBandwidthMbps(apiObject *ec2.BaselineEbsBandwidthMbps) map[string]interface{} {
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

func flattenMemoryGiBPerVCPU(apiObject *ec2.MemoryGiBPerVCpu) map[string]interface{} {
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

func flattenMemoryMiB(apiObject *ec2.MemoryMiB) map[string]interface{} {
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

func flattenNetworkInterfaceCount(apiObject *ec2.NetworkInterfaceCount) map[string]interface{} {
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

func flattenTotalLocalStorageGB(apiObject *ec2.TotalLocalStorageGB) map[string]interface{} {
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

func flattenVCPUCountRange(apiObject *ec2.VCpuCountRange) map[string]interface{} {
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

func flattenLaunchTemplateSpotMarketOptions(apiObject *ec2.LaunchTemplateSpotMarketOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BlockDurationMinutes; v != nil {
		tfMap["block_duration_minutes"] = aws.Int64Value(v)
	}

	if v := apiObject.InstanceInterruptionBehavior; v != nil {
		tfMap["instance_interruption_behavior"] = aws.StringValue(v)
	}

	if v := apiObject.MaxPrice; v != nil {
		tfMap["max_price"] = aws.StringValue(v)
	}

	if v := apiObject.SpotInstanceType; v != nil {
		tfMap["spot_instance_type"] = aws.StringValue(v)
	}

	if v := apiObject.ValidUntil; v != nil {
		tfMap["valid_until"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	return tfMap
}

func flattenLaunchTemplateLicenseConfiguration(apiObject *ec2.LaunchTemplateLicenseConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LicenseConfigurationArn; v != nil {
		tfMap["license_configuration_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateLicenseConfigurations(apiObjects []*ec2.LaunchTemplateLicenseConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchTemplateLicenseConfiguration(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateInstanceMaintenanceOptions(apiObject *ec2.LaunchTemplateInstanceMaintenanceOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AutoRecovery; v != nil {
		tfMap["auto_recovery"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateInstanceMetadataOptions(apiObject *ec2.LaunchTemplateInstanceMetadataOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HttpEndpoint; v != nil {
		tfMap["http_endpoint"] = aws.StringValue(v)
	}

	if v := apiObject.HttpProtocolIpv6; v != nil {
		tfMap["http_protocol_ipv6"] = aws.StringValue(v)
	}

	if v := apiObject.HttpPutResponseHopLimit; v != nil {
		tfMap["http_put_response_hop_limit"] = aws.Int64Value(v)
	}

	if v := apiObject.HttpTokens; v != nil {
		tfMap["http_tokens"] = aws.StringValue(v)
	}

	if v := apiObject.InstanceMetadataTags; v != nil {
		tfMap["instance_metadata_tags"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateInstanceNetworkInterfaceSpecification(apiObject *ec2.LaunchTemplateInstanceNetworkInterfaceSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AssociateCarrierIpAddress; v != nil {
		tfMap["associate_carrier_ip_address"] = strconv.FormatBool(aws.BoolValue(v))
	}

	if v := apiObject.AssociatePublicIpAddress; v != nil {
		tfMap["associate_public_ip_address"] = strconv.FormatBool(aws.BoolValue(v))
	}

	if v := apiObject.DeleteOnTermination; v != nil {
		tfMap["delete_on_termination"] = strconv.FormatBool(aws.BoolValue(v))
	}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	if v := apiObject.DeviceIndex; v != nil {
		tfMap["device_index"] = aws.Int64Value(v)
	}

	if v := apiObject.InterfaceType; v != nil {
		tfMap["interface_type"] = aws.StringValue(v)
	}

	if v := apiObject.SecondaryPrivateIpAddressCount; v != nil {
		tfMap["ipv4_address_count"] = aws.Int64Value(v)
	}

	if v := apiObject.PrivateIpAddresses; len(v) > 0 {
		var ipv4Addresses []string

		for _, v := range v {
			ipv4Addresses = append(ipv4Addresses, aws.StringValue(v.PrivateIpAddress))
		}

		tfMap["ipv4_addresses"] = ipv4Addresses
	}

	if v := apiObject.Ipv4PrefixCount; v != nil {
		tfMap["ipv4_prefix_count"] = aws.Int64Value(v)
	}

	if v := apiObject.Ipv4Prefixes; v != nil {
		var ipv4Prefixes []string

		for _, v := range v {
			ipv4Prefixes = append(ipv4Prefixes, aws.StringValue(v.Ipv4Prefix))
		}

		tfMap["ipv4_prefixes"] = ipv4Prefixes
	}

	if v := apiObject.Ipv6AddressCount; v != nil {
		tfMap["ipv6_address_count"] = aws.Int64Value(v)
	}

	if v := apiObject.Ipv6Addresses; len(v) > 0 {
		var ipv6Addresses []string

		for _, v := range v {
			ipv6Addresses = append(ipv6Addresses, aws.StringValue(v.Ipv6Address))
		}

		tfMap["ipv6_addresses"] = ipv6Addresses
	}

	if v := apiObject.Ipv6PrefixCount; v != nil {
		tfMap["ipv6_prefix_count"] = aws.Int64Value(v)
	}

	if v := apiObject.Ipv6Prefixes; v != nil {
		var ipv6Prefixes []string

		for _, v := range v {
			ipv6Prefixes = append(ipv6Prefixes, aws.StringValue(v.Ipv6Prefix))
		}

		tfMap["ipv6_prefixes"] = ipv6Prefixes
	}

	if v := apiObject.NetworkCardIndex; v != nil {
		tfMap["network_card_index"] = aws.Int64Value(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap["network_interface_id"] = aws.StringValue(v)
	}

	if v := apiObject.PrivateIpAddress; v != nil {
		tfMap["private_ip_address"] = aws.StringValue(v)
	}

	if v := apiObject.Groups; v != nil {
		tfMap["security_groups"] = aws.StringValueSlice(v)
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap["subnet_id"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateInstanceNetworkInterfaceSpecifications(apiObjects []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecification) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchTemplateInstanceNetworkInterfaceSpecification(apiObject))
	}

	return tfList
}

func flattenLaunchTemplatePlacement(apiObject *ec2.LaunchTemplatePlacement) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Affinity; v != nil {
		tfMap["affinity"] = aws.StringValue(v)
	}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap["availability_zone"] = aws.StringValue(v)
	}

	if v := apiObject.GroupName; v != nil {
		tfMap["group_name"] = aws.StringValue(v)
	}

	if v := apiObject.HostId; v != nil {
		tfMap["host_id"] = aws.StringValue(v)
	}

	if v := apiObject.HostResourceGroupArn; v != nil {
		tfMap["host_resource_group_arn"] = aws.StringValue(v)
	}

	if v := apiObject.PartitionNumber; v != nil {
		tfMap["partition_number"] = aws.Int64Value(v)
	}

	if v := apiObject.SpreadDomain; v != nil {
		tfMap["spread_domain"] = aws.StringValue(v)
	}

	if v := apiObject.Tenancy; v != nil {
		tfMap["tenancy"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplatePrivateDNSNameOptions(apiObject *ec2.LaunchTemplatePrivateDnsNameOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"enable_resource_name_dns_aaaa_record": aws.BoolValue(apiObject.EnableResourceNameDnsAAAARecord),
		"enable_resource_name_dns_a_record":    aws.BoolValue(apiObject.EnableResourceNameDnsARecord),
	}

	if v := apiObject.HostnameType; v != nil {
		tfMap["hostname_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenLaunchTemplateTagSpecification(apiObject *ec2.LaunchTemplateTagSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ResourceType; v != nil {
		tfMap["resource_type"] = aws.StringValue(v)
	}

	if v := apiObject.Tags; len(v) > 0 {
		tfMap["tags"] = KeyValueTags(v).IgnoreAWS().Map()
	}

	return tfMap
}

func flattenLaunchTemplateTagSpecifications(apiObjects []*ec2.LaunchTemplateTagSpecification) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenLaunchTemplateTagSpecification(apiObject))
	}

	return tfList
}
