// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_launch_configuration")
func DataSourceLaunchConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLaunchConfigurationRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"associate_public_ip_address": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ebs_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"encrypted": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"no_device": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"snapshot_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"throughput": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"enable_monitoring": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ephemeral_block_device": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"device_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"virtual_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"iam_instance_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_name": {
				Type:     schema.TypeString,
				Computed: true,
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
						"http_put_response_hop_limit": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"http_tokens": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"placement_tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"root_block_device": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"delete_on_termination": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"encrypted": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"iops": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"throughput": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"volume_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"spot_price": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLaunchConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	autoscalingconn := meta.(*conns.AWSClient).AutoScalingConn(ctx)
	ec2conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	name := d.Get("name").(string)
	lc, err := FindLaunchConfigurationByName(ctx, autoscalingconn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Launch Configuration (%s): %s", name, err)
	}

	d.SetId(name)

	d.Set("arn", lc.LaunchConfigurationARN)
	d.Set("associate_public_ip_address", lc.AssociatePublicIpAddress)
	d.Set("ebs_optimized", lc.EbsOptimized)
	if lc.InstanceMonitoring != nil {
		d.Set("enable_monitoring", lc.InstanceMonitoring.Enabled)
	} else {
		d.Set("enable_monitoring", false)
	}
	d.Set("iam_instance_profile", lc.IamInstanceProfile)
	d.Set("image_id", lc.ImageId)
	d.Set("instance_type", lc.InstanceType)
	d.Set("key_name", lc.KeyName)
	if lc.MetadataOptions != nil {
		if err := d.Set("metadata_options", []interface{}{flattenInstanceMetadataOptions(lc.MetadataOptions)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting metadata_options: %s", err)
		}
	} else {
		d.Set("metadata_options", nil)
	}
	d.Set("name", lc.LaunchConfigurationName)
	d.Set("placement_tenancy", lc.PlacementTenancy)
	d.Set("security_groups", aws.StringValueSlice(lc.SecurityGroups))
	d.Set("spot_price", lc.SpotPrice)
	d.Set("user_data", lc.UserData)

	rootDeviceName, err := findImageRootDeviceName(ctx, ec2conn, d.Get("image_id").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Auto Scaling Launch Configuration (%s): %s", name, err)
	}

	tfListEBSBlockDevice, tfListEphemeralBlockDevice, tfListRootBlockDevice := flattenBlockDeviceMappings(lc.BlockDeviceMappings, rootDeviceName, map[string]map[string]interface{}{})

	if err := d.Set("ebs_block_device", tfListEBSBlockDevice); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ebs_block_device: %s", err)
	}
	if err := d.Set("ephemeral_block_device", tfListEphemeralBlockDevice); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ephemeral_block_device: %s", err)
	}
	if err := d.Set("root_block_device", tfListRootBlockDevice); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_block_device: %s", err)
	}

	return diags
}
