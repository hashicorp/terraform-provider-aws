// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_eips")
func DataSourceEIPs() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceEIPsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"allocation_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": DataSourceFiltersSchema(),
			"public_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceEIPsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeAddressesInput{}

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(ctx, tags.(map[string]interface{}))),
		)...)
	}

	if filters, filtersOk := d.GetOk("filter"); filtersOk {
		input.Filters = append(input.Filters,
			BuildFiltersDataSource(filters.(*schema.Set))...)
	}

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindEIPs(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 EIPs: %s", err)
	}

	var allocationIDs []string
	var publicIPs []string

	for _, v := range output {
		publicIPs = append(publicIPs, aws.StringValue(v.PublicIp))

		if aws.StringValue(v.Domain) == ec2.DomainTypeVpc {
			allocationIDs = append(allocationIDs, aws.StringValue(v.AllocationId))
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("allocation_ids", allocationIDs)
	d.Set("public_ips", publicIPs)

	return diags
}
