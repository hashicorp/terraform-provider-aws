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

// @SDKDataSource("aws_subnets")
func DataSourceSubnets() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSubnetsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSubnetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeSubnetsInput{}

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

	output, err := FindSubnets(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Subnets: %s", err)
	}

	var subnetIDs []string

	for _, v := range output {
		subnetIDs = append(subnetIDs, aws.StringValue(v.SubnetId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", subnetIDs)

	return diags
}
