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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_host", name="Host")
// @Tags
// @Testing(tagsTest=false)
func dataSourceHost() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHostRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"asset_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_placement": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			"host_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"host_recovery": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"total_vcpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeHostsInput{
		Filter: newCustomFilterList(d.Get(names.AttrFilter).(*schema.Set)),
	}

	if v, ok := d.GetOk("host_id"); ok {
		input.HostIds = []string{v.(string)}
	}

	if len(input.Filter) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filter = nil
	}

	host, err := findHost(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Host", err))
	}

	d.SetId(aws.ToString(host.HostId))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.ToString(host.OwnerId),
		Resource:  fmt.Sprintf("dedicated-host/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("asset_id", host.AssetId)
	d.Set("auto_placement", host.AutoPlacement)
	d.Set(names.AttrAvailabilityZone, host.AvailabilityZone)
	d.Set("cores", host.HostProperties.Cores)
	d.Set("host_id", host.HostId)
	d.Set("host_recovery", host.HostRecovery)
	d.Set("instance_family", host.HostProperties.InstanceFamily)
	d.Set(names.AttrInstanceType, host.HostProperties.InstanceType)
	d.Set("outpost_arn", host.OutpostArn)
	d.Set(names.AttrOwnerID, host.OwnerId)
	d.Set("sockets", host.HostProperties.Sockets)
	d.Set("total_vcpus", host.HostProperties.TotalVCpus)

	setTagsOut(ctx, host.Tags)

	return diags
}
