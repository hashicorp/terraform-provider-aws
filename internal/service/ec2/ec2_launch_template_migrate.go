// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"maps"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func launchTemplateSchemaV0() *schema.Resource {
	return &schema.Resource{
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
										Type:     nullable.TypeNullableBool,
										Optional: true,
									},
									names.AttrEncrypted: {
										Type:     nullable.TypeNullableBool,
										Optional: true,
									},
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Computed: true,
										Optional: true,
									},
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrSnapshotID: {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrThroughput: {
										Type:     schema.TypeInt,
										Computed: true,
										Optional: true,
									},
									"volume_initialization_rate": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									names.AttrVolumeSize: {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									names.AttrVolumeType: {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
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
							Type:     schema.TypeString,
							Optional: true,
						},
						"capacity_reservation_target": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_reservation_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"capacity_reservation_resource_group_arn": {
										Type:     schema.TypeString,
										Optional: true,
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
							Type:     schema.TypeString,
							Optional: true,
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
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"default_version": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
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
				Type:     nullable.TypeNullableBool,
				Optional: true,
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
							Type:     schema.TypeString,
							Optional: true,
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
				Type:     schema.TypeString,
				Optional: true,
			},
			"instance_market_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"market_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"spot_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_duration_minutes": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"instance_interruption_behavior": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"spot_instance_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"valid_until": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
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
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"accelerator_manufacturers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"accelerator_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"accelerator_total_memory_mib": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"accelerator_types": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"allowed_instance_types": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 400,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"bare_metal": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"baseline_ebs_bandwidth_mbps": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"burstable_performance": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"cpu_manufacturers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
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
								Type: schema.TypeString,
							},
						},
						"local_storage": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"local_storage_types": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"max_spot_price_as_percentage_of_optimal_on_demand_price": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"memory_gib_per_vcpu": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeFloat,
										Optional: true,
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
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Required: true,
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
										Type:     schema.TypeFloat,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeFloat,
										Optional: true,
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
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"on_demand_max_price_percentage_over_lowest_price": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"require_hibernate_support": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"spot_max_price_percentage_over_lowest_price": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"total_local_storage_gb": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeFloat,
										Optional: true,
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
										Type:     schema.TypeInt,
										Optional: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Required: true,
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
							Type:     schema.TypeString,
							Required: true,
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
							Type:     schema.TypeString,
							Optional: true,
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
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"http_protocol_ipv6": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"http_put_response_hop_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"http_tokens": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"instance_metadata_tags": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrNamePrefix: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network_interfaces": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"associate_carrier_ip_address": {
							Type:     nullable.TypeNullableBool,
							Optional: true,
						},
						"associate_public_ip_address": {
							Type:     nullable.TypeNullableBool,
							Optional: true,
						},
						"connection_tracking_specification": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tcp_established_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"udp_stream_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
									},
									"udp_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						names.AttrDeleteOnTermination: {
							Type:     nullable.TypeNullableBool,
							Optional: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ena_srd_specification": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ena_srd_enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									"ena_srd_udp_specification": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"ena_srd_udp_enabled": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"interface_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipv4_address_count": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"ipv4_addresses": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
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
								Type: schema.TypeString,
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
								Type: schema.TypeString,
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
								Type: schema.TypeString,
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
							Type:     nullable.TypeNullableBool,
							Optional: true,
						},
						"private_ip_address": {
							Type:     schema.TypeString,
							Optional: true,
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
							Type:     schema.TypeString,
							Optional: true,
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
							Type:     schema.TypeString,
							Optional: true,
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
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"ram_disk_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"security_group_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tag_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrResourceType: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrTags: tftags.TagsSchema(),
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"update_default_version": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func launchTemplateStateUpgradeV0(_ context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	if rawState == nil {
		rawState = map[string]any{}
	}

	maps.DeleteFunc(rawState, func(key string, _ any) bool {
		return strings.HasPrefix(key, "elastic_gpu_specifications.") || strings.HasPrefix(key, "elastic_inference_accelerator.")
	})

	return rawState, nil
}
