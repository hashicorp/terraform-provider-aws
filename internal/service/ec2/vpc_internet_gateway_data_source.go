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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_internet_gateway")
func DataSourceInternetGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInternetGatewayRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"attachments": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrState: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"filter": customFiltersSchema(),
			"internet_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceInternetGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	internetGatewayId, internetGatewayIdOk := d.GetOk("internet_gateway_id")
	tags, tagsOk := d.GetOk(names.AttrTags)
	filter, filterOk := d.GetOk("filter")

	if !internetGatewayIdOk && !filterOk && !tagsOk {
		return sdkdiag.AppendErrorf(diags, "One of internet_gateway_id or filter or tags must be assigned")
	}

	input := &ec2.DescribeInternetGatewaysInput{}
	input.Filters = newAttributeFilterList(map[string]string{
		"internet-gateway-id": internetGatewayId.(string),
	})
	input.Filters = append(input.Filters, newTagFilterList(
		Tags(tftags.New(ctx, tags.(map[string]interface{}))),
	)...)
	input.Filters = append(input.Filters, newCustomFilterList(
		filter.(*schema.Set),
	)...)

	igw, err := FindInternetGateway(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Internet Gateway", err))
	}

	d.SetId(aws.StringValue(igw.InternetGatewayId))

	ownerID := aws.StringValue(igw.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("internet-gateway/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)

	if err := d.Set("attachments", flattenInternetGatewayAttachments(igw.Attachments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting attachments: %s", err)
	}

	d.Set("internet_gateway_id", igw.InternetGatewayId)
	d.Set(names.AttrOwnerID, ownerID)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, igw.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func flattenInternetGatewayAttachments(igwAttachments []*ec2.InternetGatewayAttachment) []map[string]interface{} {
	attachments := make([]map[string]interface{}, 0, len(igwAttachments))
	for _, a := range igwAttachments {
		m := make(map[string]interface{})
		m[names.AttrState] = aws.StringValue(a.State)
		m[names.AttrVPCID] = aws.StringValue(a.VpcId)
		attachments = append(attachments, m)
	}

	return attachments
}
