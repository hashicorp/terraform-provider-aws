// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
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

// @SDKResource("aws_launch_template", name="Launch Template")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLaunchTemplateCreate,
		ReadWithoutTimeout:   resourceLaunchTemplateRead,
		UpdateWithoutTimeout: resourceLaunchTemplateUpdate,
		DeleteWithoutTimeout: resourceLaunchTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ebs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDeleteOnTermination: {
										Type:             nullable.TypeNullableBool,
										Optional:         true,
										DiffSuppressFunc: nullable.DiffSuppressNullableBool,
										ValidateFunc:     nullable.ValidateTypeStringNullableBool,
									},
									names.AttrEncrypted: {
										Type:             nullable.TypeNullableBool,
										Optional:         true,
										DiffSuppressFunc: nullable.DiffSuppressNullableBool,
										ValidateFunc:     nullable.ValidateTypeStringNullableBool,
									},
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Computed: true,
										Optional: true,
									},
									names.AttrKMSKeyID: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrSnapshotID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrThroughput: {
										Type:         schema.TypeInt,
										Computed:     true,
										Optional:     true,
										ValidateFunc: validation.IntBetween(125, 1000),
									},
									names.AttrVolumeSize: {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									names.AttrVolumeType: {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[awstypes.VolumeType](),
									},
								},
							},
						},
						"no_device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrVirtualName: {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.CapacityReservationPreference](),
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
						"amd_sev_snp": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.AmdSevSnpSpecification](),
						},
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
							ValidateFunc: validation.StringInSlice(cpuCredits_Values(), false),
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
			names.AttrDescription: {
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
				Type:             nullable.TypeNullableBool,
				Optional:         true,
				DiffSuppressFunc: nullable.DiffSuppressNullableBool,
				ValidateFunc:     nullable.ValidateTypeStringNullableBool,
			},
			"elastic_gpu_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
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
						names.AttrType: {
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
						names.AttrEnabled: {
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
						names.AttrARN: {
							Type:          schema.TypeString,
							Optional:      true,
							ConflictsWith: []string{"iam_instance_profile.0.name"},
							ValidateFunc:  verify.ValidARN,
						},
						names.AttrName: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ShutdownBehavior](),
			},
			"instance_market_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"market_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MarketType](),
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
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.InstanceInterruptionBehavior](),
									},
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"spot_instance_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.SpotInstanceType](),
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
							Type:          schema.TypeSet,
							Optional:      true,
							MaxItems:      400,
							Elem:          &schema.Schema{Type: schema.TypeString},
							ConflictsWith: []string{"instance_requirements.0.excluded_instance_types"},
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
							Type:          schema.TypeSet,
							Optional:      true,
							MaxItems:      400,
							Elem:          &schema.Schema{Type: schema.TypeString},
							ConflictsWith: []string{"instance_requirements.0.allowed_instance_types"},
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
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntAtLeast(1),
							ConflictsWith: []string{"instance_requirements.0.spot_max_price_percentage_over_lowest_price"},
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
							Type:          schema.TypeInt,
							Optional:      true,
							ValidateFunc:  validation.IntAtLeast(1),
							ConflictsWith: []string{"instance_requirements.0.max_spot_price_as_percentage_of_optimal_on_demand_price"},
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
				ConflictsWith: []string{names.AttrInstanceType},
			},
			names.AttrInstanceType: {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LaunchTemplateAutoRecoveryState](),
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
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LaunchTemplateInstanceMetadataEndpointState](),
						},
						"http_protocol_ipv6": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LaunchTemplateInstanceMetadataProtocolIpv6](),
						},
						"http_put_response_hop_limit": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(1, 64),
						},
						"http_tokens": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LaunchTemplateHttpTokensState](),
						},
						"instance_metadata_tags": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LaunchTemplateInstanceMetadataTagsState](),
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
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  verify.ValidLaunchTemplateName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  verify.ValidLaunchTemplateName,
			},
			"network_interfaces": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"associate_carrier_ip_address": {
							Type:             nullable.TypeNullableBool,
							Optional:         true,
							DiffSuppressFunc: nullable.DiffSuppressNullableBool,
							ValidateFunc:     nullable.ValidateTypeStringNullableBool,
						},
						"associate_public_ip_address": {
							Type:             nullable.TypeNullableBool,
							Optional:         true,
							DiffSuppressFunc: nullable.DiffSuppressNullableBool,
							ValidateFunc:     nullable.ValidateTypeStringNullableBool,
						},
						names.AttrDeleteOnTermination: {
							Type:             nullable.TypeNullableBool,
							Optional:         true,
							DiffSuppressFunc: nullable.DiffSuppressNullableBool,
							ValidateFunc:     nullable.ValidateTypeStringNullableBool,
						},
						names.AttrDescription: {
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
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"primary_ipv6": {
							Type:             nullable.TypeNullableBool,
							Optional:         true,
							DiffSuppressFunc: nullable.DiffSuppressNullableBool,
							ValidateFunc:     nullable.ValidateTypeStringNullableBool,
						},
						"private_ip_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv4Address,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetID: {
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
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrGroupName: {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.Tenancy](),
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.HostnameType](),
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
				ConflictsWith: []string{names.AttrVPCSecurityGroupIDs},
			},
			"tag_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrResourceType: {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ResourceType](),
						},
						names.AttrTags: tftags.TagsSchema(),
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"update_default_version": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"default_version"},
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrVPCSecurityGroupIDs: {
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

func resourceLaunchTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	name := create.Name(d.Get(names.AttrName).(string), d.Get(names.AttrNamePrefix).(string))
	input := &ec2.CreateLaunchTemplateInput{
		ClientToken:        aws.String(id.UniqueId()),
		LaunchTemplateName: aws.String(name),
		TagSpecifications:  getTagSpecificationsIn(ctx, awstypes.ResourceTypeLaunchTemplate),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.VersionDescription = aws.String(v.(string))
	}

	if v, err := expandRequestLaunchTemplateData(ctx, conn, d); err == nil {
		input.LaunchTemplateData = v
	} else {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := conn.CreateLaunchTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Launch Template (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.LaunchTemplate.LaunchTemplateId))

	return append(diags, resourceLaunchTemplateRead(ctx, d, meta)...)
}

func resourceLaunchTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	lt, err := findLaunchTemplateByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Launch Template %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Launch Template (%s): %s", d.Id(), err)
	}

	version := flex.Int64ToStringValue(lt.LatestVersionNumber)
	ltv, err := findLaunchTemplateVersionByTwoPartKey(ctx, conn, d.Id(), version)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Launch Template (%s) Version (%s): %s", d.Id(), version, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("launch-template/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("default_version", lt.DefaultVersionNumber)
	d.Set(names.AttrDescription, ltv.VersionDescription)
	d.Set("latest_version", lt.LatestVersionNumber)
	d.Set(names.AttrName, lt.LaunchTemplateName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(lt.LaunchTemplateName)))

	if err := flattenResponseLaunchTemplateData(ctx, conn, d, ltv.LaunchTemplateData); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	setTagsOut(ctx, lt.Tags)

	return diags
}

func resourceLaunchTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	updateKeys := []string{
		"block_device_mappings",
		"capacity_reservation_specification",
		"cpu_options",
		"credit_specification",
		names.AttrDescription,
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
		names.AttrInstanceType,
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
		names.AttrVPCSecurityGroupIDs,
	}
	latestVersion := int64(d.Get("latest_version").(int))

	if d.HasChanges(updateKeys...) {
		input := &ec2.CreateLaunchTemplateVersionInput{
			ClientToken:      aws.String(id.UniqueId()),
			LaunchTemplateId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.VersionDescription = aws.String(v.(string))
		}

		if v, err := expandRequestLaunchTemplateData(ctx, conn, d); err == nil {
			input.LaunchTemplateData = v
		} else {
			return sdkdiag.AppendFromErr(diags, err)
		}

		output, err := conn.CreateLaunchTemplateVersion(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Launch Template (%s) Version: %s", d.Id(), err)
		}

		latestVersion = aws.ToInt64(output.LaunchTemplateVersion.VersionNumber)
	}

	if d.Get("update_default_version").(bool) || d.HasChange("default_version") {
		input := &ec2.ModifyLaunchTemplateInput{
			LaunchTemplateId: aws.String(d.Id()),
		}

		if d.Get("update_default_version").(bool) {
			input.DefaultVersion = flex.Int64ValueToString(latestVersion)
		} else if d.HasChange("default_version") {
			input.DefaultVersion = flex.IntValueToString(d.Get("default_version").(int))
		}

		_, err := conn.ModifyLaunchTemplate(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Launch Template (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceLaunchTemplateRead(ctx, d, meta)...)
}

func resourceLaunchTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Launch Template: %s", d.Id())
	_, err := conn.DeleteLaunchTemplate(ctx, &ec2.DeleteLaunchTemplateInput{
		LaunchTemplateId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidLaunchTemplateIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Launch Template (%s): %s", d.Id(), err)
	}

	return diags
}

func expandRequestLaunchTemplateData(ctx context.Context, conn *ec2.Client, d *schema.ResourceData) (*awstypes.RequestLaunchTemplateData, error) {
	apiObject := &awstypes.RequestLaunchTemplateData{
		// Always set at least one field.
		UserData: aws.String(d.Get("user_data").(string)),
	}

	var instanceType string
	if v, ok := d.GetOk(names.AttrInstanceType); ok {
		v := v.(string)

		instanceType = v
		apiObject.InstanceType = awstypes.InstanceType(v)
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
			instanceTypeInfo, err := findInstanceTypeByName(ctx, conn, instanceType)

			if err != nil {
				return nil, fmt.Errorf("reading EC2 Instance Type (%s): %w", instanceType, err)
			}

			if aws.ToBool(instanceTypeInfo.BurstablePerformanceSupported) {
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

	if v, null, _ := nullable.Bool(d.Get("ebs_optimized").(string)).ValueBool(); !null {
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

		apiObject.EnclaveOptions = &awstypes.LaunchTemplateEnclaveOptionsRequest{
			Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
		}
	}

	if v, ok := d.GetOk("hibernation_options"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		apiObject.HibernationOptions = &awstypes.LaunchTemplateHibernationOptionsRequest{
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
		apiObject.InstanceInitiatedShutdownBehavior = awstypes.ShutdownBehavior(v.(string))
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

		apiObject.Monitoring = &awstypes.LaunchTemplatesMonitoringRequest{
			Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
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
		apiObject.SecurityGroups = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("tag_specifications"); ok && len(v.([]interface{})) > 0 {
		apiObject.TagSpecifications = expandLaunchTemplateTagSpecificationRequests(ctx, v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrVPCSecurityGroupIDs); ok && v.(*schema.Set).Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	return apiObject, nil
}

func expandLaunchTemplateBlockDeviceMappingRequest(tfMap map[string]interface{}) awstypes.LaunchTemplateBlockDeviceMappingRequest {
	apiObject := awstypes.LaunchTemplateBlockDeviceMappingRequest{}

	if v, ok := tfMap["ebs"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.Ebs = expandLaunchTemplateEBSBlockDeviceRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap[names.AttrDeviceName].(string); ok && v != "" {
		apiObject.DeviceName = aws.String(v)
	}

	if v, ok := tfMap["no_device"].(string); ok && v != "" {
		apiObject.NoDevice = aws.String(v)
	}

	if v, ok := tfMap[names.AttrVirtualName].(string); ok && v != "" {
		apiObject.VirtualName = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateBlockDeviceMappingRequests(tfList []interface{}) []awstypes.LaunchTemplateBlockDeviceMappingRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateBlockDeviceMappingRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateBlockDeviceMappingRequest(tfMap))
	}

	return apiObjects
}

func expandLaunchTemplateEBSBlockDeviceRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplateEbsBlockDeviceRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateEbsBlockDeviceRequest{}

	if v, null, _ := nullable.Bool(tfMap[names.AttrDeleteOnTermination].(string)).ValueBool(); !null {
		apiObject.DeleteOnTermination = aws.Bool(v)
	}

	if v, null, _ := nullable.Bool(tfMap[names.AttrEncrypted].(string)).ValueBool(); !null {
		apiObject.Encrypted = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrIOPS].(int); ok && v != 0 {
		apiObject.Iops = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrKMSKeyID].(string); ok && v != "" {
		apiObject.KmsKeyId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrSnapshotID].(string); ok && v != "" {
		apiObject.SnapshotId = aws.String(v)
	}

	if v, ok := tfMap[names.AttrThroughput].(int); ok && v != 0 {
		apiObject.Throughput = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeSize].(int); ok && v != 0 {
		apiObject.VolumeSize = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrVolumeType].(string); ok && v != "" {
		apiObject.VolumeType = awstypes.VolumeType(v)
	}

	return apiObject
}

func expandLaunchTemplateCapacityReservationSpecificationRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplateCapacityReservationSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateCapacityReservationSpecificationRequest{}

	if v, ok := tfMap["capacity_reservation_preference"].(string); ok && v != "" {
		apiObject.CapacityReservationPreference = awstypes.CapacityReservationPreference(v)
	}

	if v, ok := tfMap["capacity_reservation_target"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.CapacityReservationTarget = expandCapacityReservationTarget(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandLaunchTemplateCPUOptionsRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplateCpuOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateCpuOptionsRequest{}

	if v, ok := tfMap["amd_sev_snp"].(string); ok && v != "" {
		apiObject.AmdSevSnp = awstypes.AmdSevSnpSpecification(v)
	}

	if v, ok := tfMap["core_count"].(int); ok && v != 0 {
		apiObject.CoreCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["threads_per_core"].(int); ok && v != 0 {
		apiObject.ThreadsPerCore = aws.Int32(int32(v))
	}

	return apiObject
}

func expandElasticGpuSpecification(tfMap map[string]interface{}) awstypes.ElasticGpuSpecification {
	apiObject := awstypes.ElasticGpuSpecification{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandElasticGpuSpecifications(tfList []interface{}) []awstypes.ElasticGpuSpecification {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ElasticGpuSpecification

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandElasticGpuSpecification(tfMap))
	}

	return apiObjects
}

func expandLaunchTemplateElasticInferenceAccelerator(tfMap map[string]interface{}) awstypes.LaunchTemplateElasticInferenceAccelerator {
	apiObject := awstypes.LaunchTemplateElasticInferenceAccelerator{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateElasticInferenceAccelerators(tfList []interface{}) []awstypes.LaunchTemplateElasticInferenceAccelerator {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateElasticInferenceAccelerator

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateElasticInferenceAccelerator(tfMap))
	}

	return apiObjects
}

func expandLaunchTemplateIAMInstanceProfileSpecificationRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplateIamInstanceProfileSpecificationRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateIamInstanceProfileSpecificationRequest{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceMarketOptionsRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplateInstanceMarketOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateInstanceMarketOptionsRequest{}

	if v, ok := tfMap["market_type"].(string); ok && v != "" {
		apiObject.MarketType = awstypes.MarketType(v)
	}

	if v, ok := tfMap["spot_options"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.SpotOptions = expandLaunchTemplateSpotMarketOptionsRequest(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandInstanceRequirementsRequest(tfMap map[string]interface{}) *awstypes.InstanceRequirementsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.InstanceRequirementsRequest{}

	if v, ok := tfMap["accelerator_count"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.AcceleratorCount = expandAcceleratorCountRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["accelerator_manufacturers"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorManufacturers = flex.ExpandStringyValueSet[awstypes.AcceleratorManufacturer](v)
	}

	if v, ok := tfMap["accelerator_names"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AcceleratorNames = flex.ExpandStringyValueSet[awstypes.AcceleratorName](v)
	}

	if v, ok := tfMap["accelerator_total_memory_mib"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
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

	if v, ok := tfMap["baseline_ebs_bandwidth_mbps"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
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

	if v, ok := tfMap["memory_gib_per_vcpu"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.MemoryGiBPerVCpu = expandMemoryGiBPerVCPURequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["memory_mib"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.MemoryMiB = expandMemoryMiBRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["network_bandwidth_gbps"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.NetworkBandwidthGbps = expandNetworkBandwidthGbpsRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["network_interface_count"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
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

	if v, ok := tfMap["total_local_storage_gb"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.TotalLocalStorageGB = expandTotalLocalStorageGBRequest(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["vcpu_count"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.VCpuCount = expandVCPUCountRangeRequest(v[0].(map[string]interface{}))
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

func expandVCPUCountRangeRequest(tfMap map[string]interface{}) *awstypes.VCpuCountRangeRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.VCpuCountRangeRequest{}

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

func expandLaunchTemplateSpotMarketOptionsRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplateSpotMarketOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateSpotMarketOptionsRequest{}

	if v, ok := tfMap["block_duration_minutes"].(int); ok && v != 0 {
		apiObject.BlockDurationMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["instance_interruption_behavior"].(string); ok && v != "" {
		apiObject.InstanceInterruptionBehavior = awstypes.InstanceInterruptionBehavior(v)
	}

	if v, ok := tfMap["max_price"].(string); ok && v != "" {
		apiObject.MaxPrice = aws.String(v)
	}

	if v, ok := tfMap["spot_instance_type"].(string); ok && v != "" {
		apiObject.SpotInstanceType = awstypes.SpotInstanceType(v)
	}

	if v, ok := tfMap["valid_until"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.ValidUntil = aws.Time(v)
	}

	return apiObject
}

func expandLaunchTemplateLicenseConfigurationRequest(tfMap map[string]interface{}) awstypes.LaunchTemplateLicenseConfigurationRequest {
	apiObject := awstypes.LaunchTemplateLicenseConfigurationRequest{}

	if v, ok := tfMap["license_configuration_arn"].(string); ok && v != "" {
		apiObject.LicenseConfigurationArn = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateLicenseConfigurationRequests(tfList []interface{}) []awstypes.LaunchTemplateLicenseConfigurationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateLicenseConfigurationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateLicenseConfigurationRequest(tfMap))
	}

	return apiObjects
}

func expandLaunchTemplateInstanceMetadataOptionsRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplateInstanceMetadataOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateInstanceMetadataOptionsRequest{}

	if v, ok := tfMap["http_endpoint"].(string); ok && v != "" {
		apiObject.HttpEndpoint = awstypes.LaunchTemplateInstanceMetadataEndpointState(v)
	}

	if v, ok := tfMap["http_tokens"].(string); ok && v != "" {
		apiObject.HttpTokens = awstypes.LaunchTemplateHttpTokensState(v)
	}

	if v, ok := tfMap["http_put_response_hop_limit"].(int); ok && v != 0 {
		apiObject.HttpPutResponseHopLimit = aws.Int32(int32(v))
	}

	if v, ok := tfMap["instance_metadata_tags"].(string); ok && v != "" {
		apiObject.InstanceMetadataTags = awstypes.LaunchTemplateInstanceMetadataTagsState(v)
	}

	if v, ok := tfMap["http_protocol_ipv6"].(string); ok && v != "" {
		apiObject.HttpProtocolIpv6 = awstypes.LaunchTemplateInstanceMetadataProtocolIpv6(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceMaintenanceOptionsRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplateInstanceMaintenanceOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplateInstanceMaintenanceOptionsRequest{}

	if v, ok := tfMap["auto_recovery"].(string); ok && v != "" {
		apiObject.AutoRecovery = awstypes.LaunchTemplateAutoRecoveryState(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequest(tfMap map[string]interface{}) awstypes.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest {
	apiObject := awstypes.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{}

	if v, null, _ := nullable.Bool(tfMap["associate_carrier_ip_address"].(string)).ValueBool(); !null {
		apiObject.AssociateCarrierIpAddress = aws.Bool(v)
	}

	if v, null, _ := nullable.Bool(tfMap["associate_public_ip_address"].(string)).ValueBool(); !null {
		apiObject.AssociatePublicIpAddress = aws.Bool(v)
	}

	if v, null, _ := nullable.Bool(tfMap[names.AttrDeleteOnTermination].(string)).ValueBool(); !null {
		apiObject.DeleteOnTermination = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["device_index"].(int); ok {
		apiObject.DeviceIndex = aws.Int32(int32(v))
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
		apiObject.SecondaryPrivateIpAddressCount = aws.Int32(int32(v))
	} else if v, ok := tfMap["ipv4_addresses"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			v := v.(string)

			apiObject.PrivateIpAddresses = append(apiObject.PrivateIpAddresses, awstypes.PrivateIpAddressSpecification{
				Primary:          aws.Bool(v == privateIPAddress),
				PrivateIpAddress: aws.String(v),
			})
		}
	}

	if v, ok := tfMap["ipv4_prefix_count"].(int); ok && v != 0 {
		apiObject.Ipv4PrefixCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["ipv4_prefixes"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Ipv4Prefixes = expandLaunchTemplateIPv4PrefixSpecificationRequests(v.List())
	}

	if v, ok := tfMap["ipv6_address_count"].(int); ok && v != 0 {
		apiObject.Ipv6AddressCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["ipv6_addresses"].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.Ipv6Addresses = append(apiObject.Ipv6Addresses, awstypes.InstanceIpv6AddressRequest{
				Ipv6Address: aws.String(v.(string)),
			})
		}
	}

	if v, ok := tfMap["ipv6_prefix_count"].(int); ok && v != 0 {
		apiObject.Ipv6PrefixCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["ipv6_prefixes"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Ipv6Prefixes = expandLaunchTemplateIPv6PrefixSpecificationRequests(v.List())
	}

	if v, ok := tfMap["network_card_index"].(int); ok {
		apiObject.NetworkCardIndex = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrNetworkInterfaceID].(string); ok && v != "" {
		apiObject.NetworkInterfaceId = aws.String(v)
	}

	if v, null, _ := nullable.Bool(tfMap["primary_ipv6"].(string)).ValueBool(); !null {
		apiObject.PrimaryIpv6 = aws.Bool(v)
	}

	if v, ok := tfMap[names.AttrSecurityGroups].(*schema.Set); ok && v.Len() > 0 {
		for _, v := range v.List() {
			apiObject.Groups = append(apiObject.Groups, v.(string))
		}
	}

	if v, ok := tfMap[names.AttrSubnetID].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	return apiObject
}

func expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequests(tfList []interface{}) []awstypes.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateInstanceNetworkInterfaceSpecificationRequest(tfMap))
	}

	return apiObjects
}

func expandLaunchTemplatePlacementRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplatePlacementRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplatePlacementRequest{}

	if v, ok := tfMap["affinity"].(string); ok && v != "" {
		apiObject.Affinity = aws.String(v)
	}

	if v, ok := tfMap[names.AttrAvailabilityZone].(string); ok && v != "" {
		apiObject.AvailabilityZone = aws.String(v)
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

func expandLaunchTemplatePrivateDNSNameOptionsRequest(tfMap map[string]interface{}) *awstypes.LaunchTemplatePrivateDnsNameOptionsRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.LaunchTemplatePrivateDnsNameOptionsRequest{
		EnableResourceNameDnsAAAARecord: aws.Bool(tfMap["enable_resource_name_dns_aaaa_record"].(bool)),
		EnableResourceNameDnsARecord:    aws.Bool(tfMap["enable_resource_name_dns_a_record"].(bool)),
	}

	if v, ok := tfMap["hostname_type"].(string); ok && v != "" {
		apiObject.HostnameType = awstypes.HostnameType(v)
	}

	return apiObject
}

func expandLaunchTemplateTagSpecificationRequest(ctx context.Context, tfMap map[string]interface{}) awstypes.LaunchTemplateTagSpecificationRequest {
	apiObject := awstypes.LaunchTemplateTagSpecificationRequest{}

	if v, ok := tfMap[names.AttrResourceType].(string); ok && v != "" {
		apiObject.ResourceType = awstypes.ResourceType(v)
	}

	if v, ok := tfMap[names.AttrTags].(map[string]interface{}); ok && len(v) > 0 {
		if v := tftags.New(ctx, v).IgnoreAWS(); len(v) > 0 {
			apiObject.Tags = Tags(v)
		}
	}

	return apiObject
}

func expandLaunchTemplateTagSpecificationRequests(ctx context.Context, tfList []interface{}) []awstypes.LaunchTemplateTagSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.LaunchTemplateTagSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateTagSpecificationRequest(ctx, tfMap))
	}

	return apiObjects
}

func flattenResponseLaunchTemplateData(ctx context.Context, conn *ec2.Client, d *schema.ResourceData, apiObject *awstypes.ResponseLaunchTemplateData) error {
	instanceType := string(apiObject.InstanceType)

	if err := d.Set("block_device_mappings", flattenLaunchTemplateBlockDeviceMappings(apiObject.BlockDeviceMappings)); err != nil {
		return fmt.Errorf("setting block_device_mappings: %w", err)
	}
	if apiObject.CapacityReservationSpecification != nil {
		if err := d.Set("capacity_reservation_specification", []interface{}{flattenLaunchTemplateCapacityReservationSpecificationResponse(apiObject.CapacityReservationSpecification)}); err != nil {
			return fmt.Errorf("setting capacity_reservation_specification: %w", err)
		}
	} else {
		d.Set("capacity_reservation_specification", nil)
	}
	if apiObject.CpuOptions != nil {
		if err := d.Set("cpu_options", []interface{}{flattenLaunchTemplateCPUOptions(apiObject.CpuOptions)}); err != nil {
			return fmt.Errorf("setting cpu_options: %w", err)
		}
	} else {
		d.Set("cpu_options", nil)
	}
	if apiObject.CreditSpecification != nil && instanceType != "" {
		instanceTypeInfo, err := findInstanceTypeByName(ctx, conn, instanceType)

		if err != nil {
			return fmt.Errorf("reading EC2 Instance Type (%s): %w", instanceType, err)
		}

		if aws.ToBool(instanceTypeInfo.BurstablePerformanceSupported) {
			if err := d.Set("credit_specification", []interface{}{flattenCreditSpecification(apiObject.CreditSpecification)}); err != nil {
				return fmt.Errorf("setting credit_specification: %w", err)
			}
		}
	} // Don't overwrite any configured value.
	d.Set("disable_api_stop", apiObject.DisableApiStop)
	d.Set("disable_api_termination", apiObject.DisableApiTermination)
	if apiObject.EbsOptimized != nil {
		d.Set("ebs_optimized", flex.BoolToStringValue(apiObject.EbsOptimized))
	} else {
		d.Set("ebs_optimized", "")
	}
	if err := d.Set("elastic_gpu_specifications", flattenElasticGpuSpecificationResponses(apiObject.ElasticGpuSpecifications)); err != nil {
		return fmt.Errorf("setting elastic_gpu_specifications: %w", err)
	}
	if err := d.Set("elastic_inference_accelerator", flattenLaunchTemplateElasticInferenceAcceleratorResponses(apiObject.ElasticInferenceAccelerators)); err != nil {
		return fmt.Errorf("setting elastic_inference_accelerator: %w", err)
	}
	if apiObject.EnclaveOptions != nil {
		tfMap := map[string]interface{}{
			names.AttrEnabled: aws.ToBool(apiObject.EnclaveOptions.Enabled),
		}

		if err := d.Set("enclave_options", []interface{}{tfMap}); err != nil {
			return fmt.Errorf("setting enclave_options: %w", err)
		}
	} else {
		d.Set("enclave_options", nil)
	}
	if apiObject.HibernationOptions != nil {
		tfMap := map[string]interface{}{
			"configured": aws.ToBool(apiObject.HibernationOptions.Configured),
		}

		if err := d.Set("hibernation_options", []interface{}{tfMap}); err != nil {
			return fmt.Errorf("setting hibernation_options: %w", err)
		}
	} else {
		d.Set("hibernation_options", nil)
	}
	if apiObject.IamInstanceProfile != nil {
		if err := d.Set("iam_instance_profile", []interface{}{flattenLaunchTemplateIAMInstanceProfileSpecification(apiObject.IamInstanceProfile)}); err != nil {
			return fmt.Errorf("setting iam_instance_profile: %w", err)
		}
	} else {
		d.Set("iam_instance_profile", nil)
	}
	d.Set("image_id", apiObject.ImageId)
	d.Set("instance_initiated_shutdown_behavior", apiObject.InstanceInitiatedShutdownBehavior)
	if apiObject.InstanceMarketOptions != nil {
		if err := d.Set("instance_market_options", []interface{}{flattenLaunchTemplateInstanceMarketOptions(apiObject.InstanceMarketOptions)}); err != nil {
			return fmt.Errorf("setting instance_market_options: %w", err)
		}
	} else {
		d.Set("instance_market_options", nil)
	}
	if apiObject.InstanceRequirements != nil {
		if err := d.Set("instance_requirements", []interface{}{flattenInstanceRequirements(apiObject.InstanceRequirements)}); err != nil {
			return fmt.Errorf("setting instance_requirements: %w", err)
		}
	} else {
		d.Set("instance_requirements", nil)
	}
	d.Set(names.AttrInstanceType, instanceType)
	d.Set("kernel_id", apiObject.KernelId)
	d.Set("key_name", apiObject.KeyName)
	if err := d.Set("license_specification", flattenLaunchTemplateLicenseConfigurations(apiObject.LicenseSpecifications)); err != nil {
		return fmt.Errorf("setting license_specification: %w", err)
	}
	if apiObject.MaintenanceOptions != nil {
		if err := d.Set("maintenance_options", []interface{}{flattenLaunchTemplateInstanceMaintenanceOptions(apiObject.MaintenanceOptions)}); err != nil {
			return fmt.Errorf("setting maintenance_options: %w", err)
		}
	} else {
		d.Set("maintenance_options", nil)
	}
	if apiObject.MetadataOptions != nil {
		if err := d.Set("metadata_options", []interface{}{flattenLaunchTemplateInstanceMetadataOptions(apiObject.MetadataOptions)}); err != nil {
			return fmt.Errorf("setting metadata_options: %w", err)
		}
	} else {
		d.Set("metadata_options", nil)
	}
	if apiObject.Monitoring != nil {
		tfMap := map[string]interface{}{
			names.AttrEnabled: aws.ToBool(apiObject.Monitoring.Enabled),
		}

		if err := d.Set("monitoring", []interface{}{tfMap}); err != nil {
			return fmt.Errorf("setting monitoring: %w", err)
		}
	} else {
		d.Set("monitoring", nil)
	}
	if err := d.Set("network_interfaces", flattenLaunchTemplateInstanceNetworkInterfaceSpecifications(apiObject.NetworkInterfaces)); err != nil {
		return fmt.Errorf("setting network_interfaces: %w", err)
	}
	if apiObject.Placement != nil {
		if err := d.Set("placement", []interface{}{flattenLaunchTemplatePlacement(apiObject.Placement)}); err != nil {
			return fmt.Errorf("setting placement: %w", err)
		}
	} else {
		d.Set("placement", nil)
	}
	if apiObject.PrivateDnsNameOptions != nil {
		if err := d.Set("private_dns_name_options", []interface{}{flattenLaunchTemplatePrivateDNSNameOptions(apiObject.PrivateDnsNameOptions)}); err != nil {
			return fmt.Errorf("setting private_dns_name_options: %w", err)
		}
	} else {
		d.Set("private_dns_name_options", nil)
	}
	d.Set("ram_disk_id", apiObject.RamDiskId)
	d.Set("security_group_names", apiObject.SecurityGroups)
	if err := d.Set("tag_specifications", flattenLaunchTemplateTagSpecifications(ctx, apiObject.TagSpecifications)); err != nil {
		return fmt.Errorf("setting tag_specifications: %w", err)
	}
	d.Set("user_data", apiObject.UserData)
	d.Set(names.AttrVPCSecurityGroupIDs, apiObject.SecurityGroupIds)

	return nil
}

func flattenLaunchTemplateBlockDeviceMapping(apiObject awstypes.LaunchTemplateBlockDeviceMapping) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.DeviceName; v != nil {
		tfMap[names.AttrDeviceName] = aws.ToString(v)
	}

	if v := apiObject.Ebs; v != nil {
		tfMap["ebs"] = []interface{}{flattenLaunchTemplateEBSBlockDevice(v)}
	}

	if v := apiObject.NoDevice; v != nil {
		tfMap["no_device"] = aws.ToString(v)
	}

	if v := apiObject.VirtualName; v != nil {
		tfMap[names.AttrVirtualName] = aws.ToString(v)
	}

	return tfMap
}

func flattenLaunchTemplateBlockDeviceMappings(apiObjects []awstypes.LaunchTemplateBlockDeviceMapping) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateBlockDeviceMapping(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateEBSBlockDevice(apiObject *awstypes.LaunchTemplateEbsBlockDevice) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DeleteOnTermination; v != nil {
		tfMap[names.AttrDeleteOnTermination] = flex.BoolToStringValue(v)
	}

	if v := apiObject.Encrypted; v != nil {
		tfMap[names.AttrEncrypted] = flex.BoolToStringValue(v)
	}

	if v := apiObject.Iops; v != nil {
		tfMap[names.AttrIOPS] = aws.ToInt32(v)
	}

	if v := apiObject.KmsKeyId; v != nil {
		tfMap[names.AttrKMSKeyID] = aws.ToString(v)
	}

	if v := apiObject.SnapshotId; v != nil {
		tfMap[names.AttrSnapshotID] = aws.ToString(v)
	}

	if v := apiObject.Throughput; v != nil {
		tfMap[names.AttrThroughput] = aws.ToInt32(v)
	}

	if v := apiObject.VolumeSize; v != nil {
		tfMap[names.AttrVolumeSize] = aws.ToInt32(v)
	}

	if v := apiObject.VolumeType; v != "" {
		tfMap[names.AttrVolumeType] = v
	}

	return tfMap
}

func flattenLaunchTemplateCapacityReservationSpecificationResponse(apiObject *awstypes.LaunchTemplateCapacityReservationSpecificationResponse) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CapacityReservationPreference; v != "" {
		tfMap["capacity_reservation_preference"] = v
	}

	if v := apiObject.CapacityReservationTarget; v != nil {
		tfMap["capacity_reservation_target"] = []interface{}{flattenCapacityReservationTargetResponse(v)}
	}

	return tfMap
}

func flattenLaunchTemplateCPUOptions(apiObject *awstypes.LaunchTemplateCpuOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AmdSevSnp; v != "" {
		tfMap["amd_sev_snp"] = v
	}

	if v := apiObject.CoreCount; v != nil {
		tfMap["core_count"] = aws.ToInt32(v)
	}

	if v := apiObject.ThreadsPerCore; v != nil {
		tfMap["threads_per_core"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenCreditSpecification(apiObject *awstypes.CreditSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CpuCredits; v != nil {
		tfMap["cpu_credits"] = aws.ToString(v)
	}

	return tfMap
}

func flattenElasticGpuSpecificationResponse(apiObject awstypes.ElasticGpuSpecificationResponse) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	return tfMap
}

func flattenElasticGpuSpecificationResponses(apiObjects []awstypes.ElasticGpuSpecificationResponse) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenElasticGpuSpecificationResponse(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateElasticInferenceAcceleratorResponse(apiObject awstypes.LaunchTemplateElasticInferenceAcceleratorResponse) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.Type; v != nil {
		tfMap[names.AttrType] = aws.ToString(v)
	}

	return tfMap
}

func flattenLaunchTemplateElasticInferenceAcceleratorResponses(apiObjects []awstypes.LaunchTemplateElasticInferenceAcceleratorResponse) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateElasticInferenceAcceleratorResponse(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateIAMInstanceProfileSpecification(apiObject *awstypes.LaunchTemplateIamInstanceProfileSpecification) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		tfMap[names.AttrARN] = aws.ToString(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	return tfMap
}

func flattenLaunchTemplateInstanceMarketOptions(apiObject *awstypes.LaunchTemplateInstanceMarketOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MarketType; v != "" {
		tfMap["market_type"] = v
	}

	if v := apiObject.SpotOptions; v != nil {
		tfMap["spot_options"] = []interface{}{flattenLaunchTemplateSpotMarketOptions(v)}
	}

	return tfMap
}

func flattenInstanceRequirements(apiObject *awstypes.InstanceRequirements) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AcceleratorCount; v != nil {
		tfMap["accelerator_count"] = []interface{}{flattenAcceleratorCount(v)}
	}

	if v := apiObject.AcceleratorManufacturers; v != nil {
		tfMap["accelerator_manufacturers"] = flex.FlattenStringyValueSet[awstypes.AcceleratorManufacturer](v)
	}

	if v := apiObject.AcceleratorNames; v != nil {
		tfMap["accelerator_names"] = flex.FlattenStringyValueSet[awstypes.AcceleratorName](v)
	}

	if v := apiObject.AcceleratorTotalMemoryMiB; v != nil {
		tfMap["accelerator_total_memory_mib"] = []interface{}{flattenAcceleratorTotalMemoryMiB(v)}
	}

	if v := apiObject.AcceleratorTypes; v != nil {
		tfMap["accelerator_types"] = flex.FlattenStringyValueSet[awstypes.AcceleratorType](v)
	}

	if v := apiObject.AllowedInstanceTypes; v != nil {
		tfMap["allowed_instance_types"] = v
	}

	if v := apiObject.BareMetal; v != "" {
		tfMap["bare_metal"] = v
	}

	if v := apiObject.BaselineEbsBandwidthMbps; v != nil {
		tfMap["baseline_ebs_bandwidth_mbps"] = []interface{}{flattenBaselineEBSBandwidthMbps(v)}
	}

	if v := apiObject.BurstablePerformance; v != "" {
		tfMap["burstable_performance"] = v
	}

	if v := apiObject.CpuManufacturers; v != nil {
		tfMap["cpu_manufacturers"] = flex.FlattenStringyValueSet[awstypes.CpuManufacturer](v)
	}

	if v := apiObject.ExcludedInstanceTypes; v != nil {
		tfMap["excluded_instance_types"] = v
	}

	if v := apiObject.InstanceGenerations; v != nil {
		tfMap["instance_generations"] = flex.FlattenStringyValueSet[awstypes.InstanceGeneration](v)
	}

	if v := apiObject.LocalStorage; v != "" {
		tfMap["local_storage"] = v
	}

	if v := apiObject.LocalStorageTypes; v != nil {
		tfMap["local_storage_types"] = flex.FlattenStringyValueSet[awstypes.LocalStorageType](v)
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
		tfMap["total_local_storage_gb"] = []interface{}{flattenTotalLocalStorageGB(v)}
	}

	if v := apiObject.VCpuCount; v != nil {
		tfMap["vcpu_count"] = []interface{}{flattenVCPUCountRange(v)}
	}

	return tfMap
}

func flattenAcceleratorCount(apiObject *awstypes.AcceleratorCount) map[string]interface{} {
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

func flattenAcceleratorTotalMemoryMiB(apiObject *awstypes.AcceleratorTotalMemoryMiB) map[string]interface{} {
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

func flattenBaselineEBSBandwidthMbps(apiObject *awstypes.BaselineEbsBandwidthMbps) map[string]interface{} {
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

func flattenMemoryGiBPerVCPU(apiObject *awstypes.MemoryGiBPerVCpu) map[string]interface{} {
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

func flattenMemoryMiB(apiObject *awstypes.MemoryMiB) map[string]interface{} {
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

func flattenNetworkBandwidthGbps(apiObject *awstypes.NetworkBandwidthGbps) map[string]interface{} {
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

func flattenNetworkInterfaceCount(apiObject *awstypes.NetworkInterfaceCount) map[string]interface{} {
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

func flattenTotalLocalStorageGB(apiObject *awstypes.TotalLocalStorageGB) map[string]interface{} {
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

func flattenVCPUCountRange(apiObject *awstypes.VCpuCountRange) map[string]interface{} {
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

func flattenLaunchTemplateSpotMarketOptions(apiObject *awstypes.LaunchTemplateSpotMarketOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.BlockDurationMinutes; v != nil {
		tfMap["block_duration_minutes"] = aws.ToInt32(v)
	}

	if v := apiObject.InstanceInterruptionBehavior; v != "" {
		tfMap["instance_interruption_behavior"] = v
	}

	if v := apiObject.MaxPrice; v != nil {
		tfMap["max_price"] = aws.ToString(v)
	}

	if v := apiObject.SpotInstanceType; v != "" {
		tfMap["spot_instance_type"] = v
	}

	if v := apiObject.ValidUntil; v != nil {
		tfMap["valid_until"] = aws.ToTime(v).Format(time.RFC3339)
	}

	return tfMap
}

func flattenLaunchTemplateLicenseConfiguration(apiObject awstypes.LaunchTemplateLicenseConfiguration) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.LicenseConfigurationArn; v != nil {
		tfMap["license_configuration_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenLaunchTemplateLicenseConfigurations(apiObjects []awstypes.LaunchTemplateLicenseConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateLicenseConfiguration(apiObject))
	}

	return tfList
}

func flattenLaunchTemplateInstanceMaintenanceOptions(apiObject *awstypes.LaunchTemplateInstanceMaintenanceOptions) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AutoRecovery; v != "" {
		tfMap["auto_recovery"] = v
	}

	return tfMap
}

func flattenLaunchTemplateInstanceMetadataOptions(apiObject *awstypes.LaunchTemplateInstanceMetadataOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.HttpEndpoint; v != "" {
		tfMap["http_endpoint"] = v
	}

	if v := apiObject.HttpProtocolIpv6; v != "" {
		tfMap["http_protocol_ipv6"] = v
	}

	if v := apiObject.HttpPutResponseHopLimit; v != nil {
		tfMap["http_put_response_hop_limit"] = aws.ToInt32(v)
	}

	if v := apiObject.HttpTokens; v != "" {
		tfMap["http_tokens"] = v
	}

	if v := apiObject.InstanceMetadataTags; v != "" {
		tfMap["instance_metadata_tags"] = v
	}

	return tfMap
}

func flattenLaunchTemplateInstanceNetworkInterfaceSpecification(apiObject awstypes.LaunchTemplateInstanceNetworkInterfaceSpecification) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AssociateCarrierIpAddress; v != nil {
		tfMap["associate_carrier_ip_address"] = flex.BoolToStringValue(v)
	}

	if v := apiObject.AssociatePublicIpAddress; v != nil {
		tfMap["associate_public_ip_address"] = flex.BoolToStringValue(v)
	}

	if v := apiObject.DeleteOnTermination; v != nil {
		tfMap[names.AttrDeleteOnTermination] = flex.BoolToStringValue(v)
	}

	if v := apiObject.Description; v != nil {
		tfMap[names.AttrDescription] = aws.ToString(v)
	}

	if v := apiObject.DeviceIndex; v != nil {
		tfMap["device_index"] = aws.ToInt32(v)
	}

	if v := apiObject.InterfaceType; v != nil {
		tfMap["interface_type"] = aws.ToString(v)
	}

	if v := apiObject.SecondaryPrivateIpAddressCount; v != nil {
		tfMap["ipv4_address_count"] = aws.ToInt32(v)
	}

	if v := apiObject.PrivateIpAddresses; len(v) > 0 {
		var ipv4Addresses []string

		for _, v := range v {
			ipv4Addresses = append(ipv4Addresses, aws.ToString(v.PrivateIpAddress))
		}

		tfMap["ipv4_addresses"] = ipv4Addresses
	}

	if v := apiObject.Ipv4PrefixCount; v != nil {
		tfMap["ipv4_prefix_count"] = aws.ToInt32(v)
	}

	if v := apiObject.Ipv4Prefixes; v != nil {
		var ipv4Prefixes []string

		for _, v := range v {
			ipv4Prefixes = append(ipv4Prefixes, aws.ToString(v.Ipv4Prefix))
		}

		tfMap["ipv4_prefixes"] = ipv4Prefixes
	}

	if v := apiObject.Ipv6AddressCount; v != nil {
		tfMap["ipv6_address_count"] = aws.ToInt32(v)
	}

	if v := apiObject.Ipv6Addresses; len(v) > 0 {
		var ipv6Addresses []string

		for _, v := range v {
			ipv6Addresses = append(ipv6Addresses, aws.ToString(v.Ipv6Address))
		}

		tfMap["ipv6_addresses"] = ipv6Addresses
	}

	if v := apiObject.Ipv6PrefixCount; v != nil {
		tfMap["ipv6_prefix_count"] = aws.ToInt32(v)
	}

	if v := apiObject.Ipv6Prefixes; v != nil {
		var ipv6Prefixes []string

		for _, v := range v {
			ipv6Prefixes = append(ipv6Prefixes, aws.ToString(v.Ipv6Prefix))
		}

		tfMap["ipv6_prefixes"] = ipv6Prefixes
	}

	if v := apiObject.NetworkCardIndex; v != nil {
		tfMap["network_card_index"] = aws.ToInt32(v)
	}

	if v := apiObject.NetworkInterfaceId; v != nil {
		tfMap[names.AttrNetworkInterfaceID] = aws.ToString(v)
	}

	if v := apiObject.PrimaryIpv6; v != nil {
		tfMap["primary_ipv6"] = flex.BoolToStringValue(v)
	}

	if v := apiObject.PrivateIpAddress; v != nil {
		tfMap["private_ip_address"] = aws.ToString(v)
	}

	if v := apiObject.Groups; v != nil {
		tfMap[names.AttrSecurityGroups] = v
	}

	if v := apiObject.SubnetId; v != nil {
		tfMap[names.AttrSubnetID] = aws.ToString(v)
	}

	return tfMap
}

func flattenLaunchTemplateInstanceNetworkInterfaceSpecifications(apiObjects []awstypes.LaunchTemplateInstanceNetworkInterfaceSpecification) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateInstanceNetworkInterfaceSpecification(apiObject))
	}

	return tfList
}

func flattenLaunchTemplatePlacement(apiObject *awstypes.LaunchTemplatePlacement) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Affinity; v != nil {
		tfMap["affinity"] = aws.ToString(v)
	}

	if v := apiObject.AvailabilityZone; v != nil {
		tfMap[names.AttrAvailabilityZone] = aws.ToString(v)
	}

	if v := apiObject.GroupName; v != nil {
		tfMap[names.AttrGroupName] = aws.ToString(v)
	}

	if v := apiObject.HostId; v != nil {
		tfMap["host_id"] = aws.ToString(v)
	}

	if v := apiObject.HostResourceGroupArn; v != nil {
		tfMap["host_resource_group_arn"] = aws.ToString(v)
	}

	if v := apiObject.PartitionNumber; v != nil {
		tfMap["partition_number"] = aws.ToInt32(v)
	}

	if v := apiObject.SpreadDomain; v != nil {
		tfMap["spread_domain"] = aws.ToString(v)
	}

	if v := apiObject.Tenancy; v != "" {
		tfMap["tenancy"] = v
	}

	return tfMap
}

func flattenLaunchTemplatePrivateDNSNameOptions(apiObject *awstypes.LaunchTemplatePrivateDnsNameOptions) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"enable_resource_name_dns_aaaa_record": aws.ToBool(apiObject.EnableResourceNameDnsAAAARecord),
		"enable_resource_name_dns_a_record":    aws.ToBool(apiObject.EnableResourceNameDnsARecord),
	}

	if v := apiObject.HostnameType; v != "" {
		tfMap["hostname_type"] = v
	}

	return tfMap
}

func flattenLaunchTemplateTagSpecification(ctx context.Context, apiObject awstypes.LaunchTemplateTagSpecification) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.ResourceType; v != "" {
		tfMap[names.AttrResourceType] = v
	}

	if v := apiObject.Tags; len(v) > 0 {
		tfMap[names.AttrTags] = keyValueTags(ctx, v).IgnoreAWS().Map()
	}

	return tfMap
}

func flattenLaunchTemplateTagSpecifications(ctx context.Context, apiObjects []awstypes.LaunchTemplateTagSpecification) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenLaunchTemplateTagSpecification(ctx, apiObject))
	}

	return tfList
}

func expandLaunchTemplateIPv4PrefixSpecificationRequest(tfString string) awstypes.Ipv4PrefixSpecificationRequest {
	if tfString == "" {
		return awstypes.Ipv4PrefixSpecificationRequest{}
	}

	apiObject := awstypes.Ipv4PrefixSpecificationRequest{
		Ipv4Prefix: aws.String(tfString),
	}

	return apiObject
}

func expandLaunchTemplateIPv4PrefixSpecificationRequests(tfList []interface{}) []awstypes.Ipv4PrefixSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Ipv4PrefixSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateIPv4PrefixSpecificationRequest(tfString))
	}

	return apiObjects
}

func expandLaunchTemplateIPv6PrefixSpecificationRequest(tfString string) awstypes.Ipv6PrefixSpecificationRequest {
	if tfString == "" {
		return awstypes.Ipv6PrefixSpecificationRequest{}
	}

	apiObject := awstypes.Ipv6PrefixSpecificationRequest{
		Ipv6Prefix: aws.String(tfString),
	}

	return apiObject
}

func expandLaunchTemplateIPv6PrefixSpecificationRequests(tfList []interface{}) []awstypes.Ipv6PrefixSpecificationRequest {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Ipv6PrefixSpecificationRequest

	for _, tfMapRaw := range tfList {
		tfString, ok := tfMapRaw.(string)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandLaunchTemplateIPv6PrefixSpecificationRequest(tfString))
	}

	return apiObjects
}
