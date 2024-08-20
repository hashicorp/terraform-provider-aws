// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_launch_template", name="Launch Template")
// @Tags
// @Testing(tagsTest=false)
func dataSourceLaunchTemplate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLaunchTemplateRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"block_device_mappings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ebs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDeleteOnTermination: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrEncrypted: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrSnapshotID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrThroughput: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrVolumeSize: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrVolumeType: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"no_device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVirtualName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"capacity_reservation_specification": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity_reservation_preference": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"capacity_reservation_target": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_reservation_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"capacity_reservation_resource_group_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"cpu_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amd_sev_snp": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"core_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"threads_per_core": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"credit_specification": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu_credits": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"default_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disable_api_stop": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"disable_api_termination": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ebs_optimized": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"elastic_gpu_specifications": {
				Type:     schema.TypeList,
				Computed: true,
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
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"enclave_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrFilter: customFiltersSchema(),
			"hibernation_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"configured": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"iam_instance_profile": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_initiated_shutdown_behavior": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_market_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"market_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"spot_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"block_duration_minutes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"instance_interruption_behavior": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"max_price": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"spot_instance_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"valid_until": {
										Type:     schema.TypeString,
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
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"accelerator_count": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"accelerator_manufacturers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"accelerator_names": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"accelerator_total_memory_mib": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"accelerator_types": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_instance_types": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"bare_metal": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"baseline_ebs_bandwidth_mbps": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"burstable_performance": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cpu_manufacturers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"excluded_instance_types": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"instance_generations": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"local_storage": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"local_storage_types": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_spot_price_as_percentage_of_optimal_on_demand_price": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"memory_gib_per_vcpu": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
						},
						"memory_mib": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"network_bandwidth_gbps": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
						},
						"network_interface_count": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"on_demand_max_price_percentage_over_lowest_price": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"require_hibernate_support": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"spot_max_price_percentage_over_lowest_price": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"total_local_storage_gb": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
						},
						"vcpu_count": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrMax: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrMin: {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kernel_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"license_specification": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"license_configuration_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"maintenance_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_recovery": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"metadata_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"http_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_protocol_ipv6": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"http_put_response_hop_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"http_tokens": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_metadata_tags": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"monitoring": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"network_interfaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"associate_carrier_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"associate_public_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDeleteOnTermination: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"device_index": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"interface_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipv4_address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ipv4_addresses": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ipv4_prefix_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ipv4_prefixes": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ipv6_address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ipv6_addresses": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ipv6_prefix_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ipv6_prefixes": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"network_card_index": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrNetworkInterfaceID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary_ipv6": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSecurityGroups: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"placement": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"affinity": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrAvailabilityZone: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrGroupName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"host_resource_group_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"partition_number": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"spread_domain": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tenancy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"private_dns_name_options": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_resource_name_dns_aaaa_record": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"enable_resource_name_dns_a_record": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"hostname_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ram_disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tag_specifications": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrResourceType: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrTags: tftags.TagsSchemaComputed(),
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCSecurityGroupIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceLaunchTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeLaunchTemplatesInput{}

	if v, ok := d.GetOk(names.AttrID); ok {
		input.LaunchTemplateIds = []string{v.(string)}
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.LaunchTemplateNames = []string{v.(string)}
	}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	input.Filters = append(input.Filters, newTagFilterList(
		Tags(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	lt, err := findLaunchTemplate(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Launch Template", err))
	}

	d.SetId(aws.ToString(lt.LaunchTemplateId))

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

	if err := flattenResponseLaunchTemplateData(ctx, conn, d, ltv.LaunchTemplateData); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	setTagsOut(ctx, lt.Tags)

	return diags
}
