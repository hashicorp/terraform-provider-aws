// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_local_gateway_route_table_vpc_association", name="Local Gateway Route Table VPC Association")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceLocalGatewayRouteTableVPCAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocalGatewayRouteTableVPCAssociationCreate,
		ReadWithoutTimeout:   resourceLocalGatewayRouteTableVPCAssociationRead,
		UpdateWithoutTimeout: resourceLocalGatewayRouteTableVPCAssociationUpdate,
		DeleteWithoutTimeout: resourceLocalGatewayRouteTableVPCAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"local_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLocalGatewayRouteTableVPCAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	req := &ec2.CreateLocalGatewayRouteTableVpcAssociationInput{
		LocalGatewayRouteTableId: aws.String(d.Get("local_gateway_route_table_id").(string)),
		TagSpecifications:        getTagSpecificationsInV2(ctx, awstypes.ResourceTypeLocalGatewayRouteTableVpcAssociation),
		VpcId:                    aws.String(d.Get(names.AttrVPCID).(string)),
	}

	output, err := conn.CreateLocalGatewayRouteTableVpcAssociation(ctx, req)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Local Gateway Route Table VPC Association: %s", err)
	}

	d.SetId(aws.ToString(output.LocalGatewayRouteTableVpcAssociation.LocalGatewayRouteTableVpcAssociationId))

	if _, err := waitLocalGatewayRouteTableVPCAssociationAssociated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Local Gateway Route Table VPC Association (%s) to associate: %s", d.Id(), err)
	}

	return append(diags, resourceLocalGatewayRouteTableVPCAssociationRead(ctx, d, meta)...)
}

func resourceLocalGatewayRouteTableVPCAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	association, err := GetLocalGatewayRouteTableVPCAssociation(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Local Gateway Route Table VPC Association (%s): %s", d.Id(), err)
	}

	if association == nil {
		log.Printf("[WARN] EC2 Local Gateway Route Table VPC Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if aws.ToString(association.State) != string(awstypes.RouteTableAssociationStateCodeAssociated) {
		log.Printf("[WARN] EC2 Local Gateway Route Table VPC Association (%s) status (%s), removing from state", d.Id(), aws.ToString(association.State))
		d.SetId("")
		return diags
	}

	d.Set("local_gateway_id", association.LocalGatewayId)
	d.Set("local_gateway_route_table_id", association.LocalGatewayRouteTableId)

	setTagsOutV2(ctx, association.Tags)

	d.Set(names.AttrVPCID, association.VpcId)

	return diags
}

func resourceLocalGatewayRouteTableVPCAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceLocalGatewayRouteTableVPCAssociationRead(ctx, d, meta)...)
}

func resourceLocalGatewayRouteTableVPCAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DeleteLocalGatewayRouteTableVpcAssociationInput{
		LocalGatewayRouteTableVpcAssociationId: aws.String(d.Id()),
	}

	_, err := conn.DeleteLocalGatewayRouteTableVpcAssociation(ctx, input)

	if tfawserr.ErrCodeEquals(err, "InvalidLocalGatewayRouteTableVpcAssociationID.NotFound") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Local Gateway Route Table VPC Association (%s): %s", d.Id(), err)
	}

	if _, err := waitLocalGatewayRouteTableVPCAssociationDisassociated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Local Gateway Route Table VPC Association (%s) to disassociate: %s", d.Id(), err)
	}

	return diags
}

func GetLocalGatewayRouteTableVPCAssociation(ctx context.Context, conn *ec2.Client, localGatewayRouteTableVpcAssociationID string) (*awstypes.LocalGatewayRouteTableVpcAssociation, error) {
	input := &ec2.DescribeLocalGatewayRouteTableVpcAssociationsInput{
		LocalGatewayRouteTableVpcAssociationIds: []string{localGatewayRouteTableVpcAssociationID},
	}

	output, err := conn.DescribeLocalGatewayRouteTableVpcAssociations(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, fmt.Errorf("empty response")
	}

	var association awstypes.LocalGatewayRouteTableVpcAssociation

	for _, outputAssociation := range output.LocalGatewayRouteTableVpcAssociations {
		if aws.ToString(outputAssociation.LocalGatewayRouteTableVpcAssociationId) == localGatewayRouteTableVpcAssociationID {
			association = outputAssociation
			break
		}
	}

	return &association, nil
}
