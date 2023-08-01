// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_ec2_host")
func DataSourceHost() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceHostRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"filter": DataSourceFiltersSchema(),
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
			"instance_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"total_vcpus": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeHostsInput{
		Filter: BuildFiltersDataSource(d.Get("filter").(*schema.Set)),
	}

	if v, ok := d.GetOk("host_id"); ok {
		input.HostIds = aws.StringSlice([]string{v.(string)})
	}

	if len(input.Filter) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filter = nil
	}

	host, err := FindHost(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Host", err))
	}

	d.SetId(aws.StringValue(host.HostId))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(host.OwnerId),
		Resource:  fmt.Sprintf("dedicated-host/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("asset_id", host.AssetId)
	d.Set("auto_placement", host.AutoPlacement)
	d.Set("availability_zone", host.AvailabilityZone)
	d.Set("cores", host.HostProperties.Cores)
	d.Set("host_id", host.HostId)
	d.Set("host_recovery", host.HostRecovery)
	d.Set("instance_family", host.HostProperties.InstanceFamily)
	d.Set("instance_type", host.HostProperties.InstanceType)
	d.Set("outpost_arn", host.OutpostArn)
	d.Set("owner_id", host.OwnerId)
	d.Set("sockets", host.HostProperties.Sockets)
	d.Set("total_vcpus", host.HostProperties.TotalVCpus)

	if err := d.Set("tags", KeyValueTags(ctx, host.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
