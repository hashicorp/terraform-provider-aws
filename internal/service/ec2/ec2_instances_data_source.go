// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_instances", name="Instances")
func dataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrFilter: customFiltersSchema(),
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"instance_tags": tftags.TagsSchemaComputed(),
			"instance_state_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.InstanceStateName](),
				},
			},
			"ipv6_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"private_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeInstancesInput{}

	if v, ok := d.GetOk("instance_state_names"); ok && v.(*schema.Set).Len() > 0 {
		input.Filters = append(input.Filters, awstypes.Filter{
			Name:   aws.String("instance-state-name"),
			Values: flex.ExpandStringValueSet(v.(*schema.Set)),
		})
	} else {
		input.Filters = append(input.Filters, awstypes.Filter{
			Name:   aws.String("instance-state-name"),
			Values: enum.Slice(awstypes.InstanceStateNameRunning),
		})
	}

	input.Filters = append(input.Filters, newTagFilterList(
		Tags(tftags.New(ctx, d.Get("instance_tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := findInstances(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instances: %s", err)
	}

	var instanceIDs, privateIPs, publicIPs, ipv6Addresses []string

	for _, v := range output {
		instanceIDs = append(instanceIDs, aws.ToString(v.InstanceId))
		if privateIP := aws.ToString(v.PrivateIpAddress); privateIP != "" {
			privateIPs = append(privateIPs, privateIP)
		}
		if publicIP := aws.ToString(v.PublicIpAddress); publicIP != "" {
			publicIPs = append(publicIPs, publicIP)
		}
		if ipv6Address := aws.ToString(v.Ipv6Address); ipv6Address != "" {
			ipv6Addresses = append(ipv6Addresses, ipv6Address)
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrIDs, instanceIDs)
	d.Set("ipv6_addresses", ipv6Addresses)
	d.Set("private_ips", privateIPs)
	d.Set("public_ips", publicIPs)

	return diags
}
