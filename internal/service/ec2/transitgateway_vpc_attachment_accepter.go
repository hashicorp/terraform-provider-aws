// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_vpc_attachment_accepter", name="Transit Gateway VPC Attachment Accepter")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTransitGatewayVPCAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayVPCAttachmentAccepterCreate,
		ReadWithoutTimeout:   resourceTransitGatewayVPCAttachmentAccepterRead,
		UpdateWithoutTimeout: resourceTransitGatewayVPCAttachmentAccepterUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayVPCAttachmentAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"appliance_mode_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ipv6_support": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnetIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"transit_gateway_default_route_table_association": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"transit_gateway_default_route_table_propagation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			names.AttrTransitGatewayID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTransitGatewayVPCAttachmentAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayAttachmentID := d.Get(names.AttrTransitGatewayAttachmentID).(string)
	input := &ec2.AcceptTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway VPC Attachment: %s", transitGatewayAttachmentID)
	output, err := conn.AcceptTransitGatewayVpcAttachment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	d.SetId(aws.ToString(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))
	transitGatewayID := aws.ToString(output.TransitGatewayVpcAttachment.TransitGatewayId)

	if _, err := waitTransitGatewayVPCAttachmentAccepted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): waiting for completion: %s", transitGatewayAttachmentID, err)
	}

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EC2 Transit Gateway VPC Attachment (%s) tags: %s", d.Id(), err)
	}

	transitGateway, err := findTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	if err := transitGatewayRouteTableAssociationUpdate(ctx, conn, aws.ToString(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	if err := transitGatewayRouteTablePropagationUpdate(ctx, conn, aws.ToString(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway VPC Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	return append(diags, resourceTransitGatewayVPCAttachmentAccepterRead(ctx, d, meta)...)
}

func resourceTransitGatewayVPCAttachmentAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayVPCAttachment, err := findTransitGatewayVPCAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
	}

	transitGatewayID := aws.ToString(transitGatewayVPCAttachment.TransitGatewayId)
	transitGateway, err := findTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	transitGatewayDefaultRouteTableAssociation := true
	transitGatewayDefaultRouteTablePropagation := true

	if transitGatewayRouteTableID := aws.ToString(transitGateway.Options.AssociationDefaultRouteTableId); transitGatewayRouteTableID != "" {
		_, err := findTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, d.Id())

		if tfresource.NotFound(err) {
			transitGatewayDefaultRouteTableAssociation = false
		} else if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Association (%s): %s", transitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, d.Id()), err)
		}
	} else {
		transitGatewayDefaultRouteTableAssociation = false
	}

	if transitGatewayRouteTableID := aws.ToString(transitGateway.Options.PropagationDefaultRouteTableId); transitGatewayRouteTableID != "" {
		_, err := findTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, d.Id())

		if tfresource.NotFound(err) {
			transitGatewayDefaultRouteTablePropagation = false
		} else if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Propagation (%s): %s", transitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, d.Id()), err)
		}
	} else {
		transitGatewayDefaultRouteTablePropagation = false
	}

	d.Set("appliance_mode_support", transitGatewayVPCAttachment.Options.ApplianceModeSupport)
	d.Set("dns_support", transitGatewayVPCAttachment.Options.DnsSupport)
	d.Set("ipv6_support", transitGatewayVPCAttachment.Options.Ipv6Support)
	d.Set(names.AttrSubnetIDs, transitGatewayVPCAttachment.SubnetIds)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayVPCAttachment.TransitGatewayAttachmentId)
	d.Set("transit_gateway_default_route_table_association", transitGatewayDefaultRouteTableAssociation)
	d.Set("transit_gateway_default_route_table_propagation", transitGatewayDefaultRouteTablePropagation)
	d.Set(names.AttrTransitGatewayID, transitGatewayVPCAttachment.TransitGatewayId)
	d.Set(names.AttrVPCID, transitGatewayVPCAttachment.VpcId)
	d.Set("vpc_owner_id", transitGatewayVPCAttachment.VpcOwnerId)

	setTagsOut(ctx, transitGatewayVPCAttachment.Tags)

	return diags
}

func resourceTransitGatewayVPCAttachmentAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transitGatewayID := d.Get(names.AttrTransitGatewayID).(string)
		transitGateway, err := findTransitGatewayByID(ctx, conn, transitGatewayID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := transitGatewayRouteTableAssociationUpdate(ctx, conn, aws.ToString(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
			}
		}

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := transitGatewayRouteTablePropagationUpdate(ctx, conn, aws.ToString(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
			}
		}
	}

	return diags
}

func resourceTransitGatewayVPCAttachmentAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway VPC Attachment: %s", d.Id())
	_, err := conn.DeleteTransitGatewayVpcAttachment(ctx, &ec2.DeleteTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
	}

	if err := waitTransitGatewayVPCAttachmentDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway VPC Attachment (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}
