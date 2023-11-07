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

// @SDKDataSource("aws_ec2_transit_gateway_attachments")
func DataSourceTransitGatewayAttachments() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayAttachmentsRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": CustomFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceTransitGatewayAttachmentsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeTransitGatewayAttachmentsInput{}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if v, ok := d.GetOk("tags"); ok {
		input.Filters = append(input.Filters, BuildTagFilterList(
			Tags(tftags.New(ctx, v.(map[string]interface{}))),
		)...)
	}

	transitGatewayAttachments, err := FindTransitGatewayAttachments(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Attachments: %s", err)
	}

	var attachmentIDs []string

	for _, v := range transitGatewayAttachments {
		attachmentIDs = append(attachmentIDs, aws.StringValue(v.TransitGatewayAttachmentId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", attachmentIDs)

	return diags
}
