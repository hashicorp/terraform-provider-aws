// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package efs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_efs_mount_target", name="Mount Target")
func dataSourceMountTarget() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceMountTargetRead,

		Schema: map[string]*schema.Schema{
			"access_point_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrIPAddress: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrIPAddressType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"mount_target_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"mount_target_dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMountTargetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.EFSClient(ctx)

	var inputDMT efs.DescribeMountTargetsInput

	if v, ok := d.GetOk("access_point_id"); ok {
		inputDMT.AccessPointId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrFileSystemID); ok {
		inputDMT.FileSystemId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("mount_target_id"); ok {
		inputDMT.MountTargetId = aws.String(v.(string))
	}

	mt, err := findMountTarget(ctx, conn, &inputDMT)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target: %s", err)
	}

	d.SetId(aws.ToString(mt.MountTargetId))
	fsID := aws.ToString(mt.FileSystemId)
	d.Set("availability_zone_id", mt.AvailabilityZoneId)
	d.Set("availability_zone_name", mt.AvailabilityZoneName)
	d.Set(names.AttrDNSName, c.RegionalHostname(ctx, fsID+".efs"))
	d.Set("file_system_arn", fileSystemARN(ctx, c, fsID))
	d.Set(names.AttrFileSystemID, fsID)
	d.Set(names.AttrIPAddress, mt.IpAddress)
	if mt.IpAddress != nil && mt.Ipv6Address != nil {
		d.Set(names.AttrIPAddressType, awstypes.IpAddressTypeDualStack)
	} else if mt.IpAddress != nil {
		d.Set(names.AttrIPAddressType, awstypes.IpAddressTypeIpv4Only)
	} else if mt.Ipv6Address != nil {
		d.Set(names.AttrIPAddressType, awstypes.IpAddressTypeIpv6Only)
	} else {
		d.Set(names.AttrIPAddressType, nil)
	}
	d.Set("ipv6_address", mt.Ipv6Address)
	d.Set("mount_target_dns_name", meta.(*conns.AWSClient).RegionalHostname(ctx, fmt.Sprintf("%s.%s.efs", aws.ToString(mt.AvailabilityZoneName), aws.ToString(mt.FileSystemId))))
	d.Set("mount_target_id", mt.MountTargetId)
	d.Set(names.AttrNetworkInterfaceID, mt.NetworkInterfaceId)
	d.Set(names.AttrOwnerID, mt.OwnerId)
	d.Set(names.AttrSubnetID, mt.SubnetId)

	inputDMTSG := efs.DescribeMountTargetSecurityGroupsInput{
		MountTargetId: aws.String(d.Id()),
	}
	output, err := conn.DescribeMountTargetSecurityGroups(ctx, &inputDMTSG)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Mount Target (%s) security groups: %s", d.Id(), err)
	}

	d.Set(names.AttrSecurityGroups, output.SecurityGroups)

	return diags
}
