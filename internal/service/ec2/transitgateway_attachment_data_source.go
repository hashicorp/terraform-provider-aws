// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

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

// @SDKDataSource("aws_ec2_transit_gateway_attachment", name="Transit Gateway Attachment")
// @Tags
// @Testing(tagsTest=false)
func dataSourceTransitGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayAttachmentRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"association_transit_gateway_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFilter: customFiltersSchema(),
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTransitGatewayID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceTransitGatewayAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeTransitGatewayAttachmentsInput{}

	input.Filters = append(input.Filters, newCustomFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	if v, ok := d.GetOk(names.AttrTransitGatewayAttachmentID); ok {
		input.TransitGatewayAttachmentIds = []string{v.(string)}
	}

	transitGatewayAttachment, err := findTransitGatewayAttachment(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Transit Gateway Attachment", err))
	}

	transitGatewayAttachmentID := aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
	d.SetId(transitGatewayAttachmentID)

	resourceOwnerID := aws.ToString(transitGatewayAttachment.ResourceOwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: resourceOwnerID,
		Resource:  fmt.Sprintf("transit-gateway-attachment/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	if v := transitGatewayAttachment.Association; v != nil {
		d.Set("association_state", v.State)
		d.Set("association_transit_gateway_route_table_id", v.TransitGatewayRouteTableId)
	} else {
		d.Set("association_state", nil)
		d.Set("association_transit_gateway_route_table_id", nil)
	}
	d.Set(names.AttrResourceID, transitGatewayAttachment.ResourceId)
	d.Set("resource_owner_id", resourceOwnerID)
	d.Set(names.AttrResourceType, transitGatewayAttachment.ResourceType)
	d.Set(names.AttrState, transitGatewayAttachment.State)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayAttachmentID)
	d.Set(names.AttrTransitGatewayID, transitGatewayAttachment.TransitGatewayId)
	d.Set("transit_gateway_owner_id", transitGatewayAttachment.TransitGatewayOwnerId)

	setTagsOut(ctx, transitGatewayAttachment.Tags)

	return diags
}
