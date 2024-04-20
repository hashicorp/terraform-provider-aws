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
)

// @SDKDataSource("aws_ec2_transit_gateway_peering_attachments")
func DataSourceTransitGatewayPeeringAttachments() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayPeeringAttachmentsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": customFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceTransitGatewayPeeringAttachmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindTransitGatewayPeeringAttachments(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Peering Attachments: %s", err)
	}

	var attachmentIDs []string

	for _, v := range output {
		attachmentIDs = append(attachmentIDs, aws.StringValue(v.TransitGatewayAttachmentId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", attachmentIDs)

	return diags
}
