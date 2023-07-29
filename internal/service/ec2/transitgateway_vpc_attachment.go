// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_vpc_attachment", name="Transit Gateway VPC Attachment")
// @Tags(identifierAttribute="id")
func ResourceTransitGatewayVPCAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayVPCAttachmentCreate,
		ReadWithoutTimeout:   resourceTransitGatewayVPCAttachmentRead,
		UpdateWithoutTimeout: resourceTransitGatewayVPCAttachmentUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayVPCAttachmentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"appliance_mode_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.ApplianceModeSupportValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.ApplianceModeSupportValue_Values(), false),
			},
			"dns_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.DnsSupportValueEnable,
				ValidateFunc: validation.StringInSlice(ec2.DnsSupportValue_Values(), false),
			},
			"ipv6_support": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      ec2.Ipv6SupportValueDisable,
				ValidateFunc: validation.StringInSlice(ec2.Ipv6SupportValue_Values(), false),
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
			"transit_gateway_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"vpc_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"vpc_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTransitGatewayVPCAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	transitGatewayID := d.Get("transit_gateway_id").(string)
	input := &ec2.CreateTransitGatewayVpcAttachmentInput{
		Options: &ec2.CreateTransitGatewayVpcAttachmentRequestOptions{
			ApplianceModeSupport: aws.String(d.Get("appliance_mode_support").(string)),
			DnsSupport:           aws.String(d.Get("dns_support").(string)),
			Ipv6Support:          aws.String(d.Get("ipv6_support").(string)),
		},
		SubnetIds:         flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
		TransitGatewayId:  aws.String(transitGatewayID),
		TagSpecifications: getTagSpecificationsIn(ctx, ec2.ResourceTypeTransitGatewayAttachment),
		VpcId:             aws.String(d.Get("vpc_id").(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway VPC Attachment: %s", input)
	output, err := conn.CreateTransitGatewayVpcAttachmentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway VPC Attachment: %s", err)
	}

	d.SetId(aws.StringValue(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))

	if _, err := WaitTransitGatewayVPCAttachmentCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway VPC Attachment (%s) create: %s", d.Id(), err)
	}

	transitGateway, err := FindTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	// We cannot modify Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways.
	if aws.StringValue(transitGateway.OwnerId) == aws.StringValue(output.TransitGatewayVpcAttachment.VpcOwnerId) {
		if err := transitGatewayRouteTableAssociationUpdate(ctx, conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway VPC Attachment: %s", err)
		}

		if err := transitGatewayRouteTablePropagationUpdate(ctx, conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway VPC Attachment: %s", err)
		}
	}

	return append(diags, resourceTransitGatewayVPCAttachmentRead(ctx, d, meta)...)
}

func resourceTransitGatewayVPCAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	transitGatewayVPCAttachment, err := FindTransitGatewayVPCAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway VPC Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
	}

	transitGatewayID := aws.StringValue(transitGatewayVPCAttachment.TransitGatewayId)
	transitGateway, err := FindTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	// We cannot read Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways.
	transitGatewayDefaultRouteTableAssociation := true
	transitGatewayDefaultRouteTablePropagation := true

	if aws.StringValue(transitGateway.OwnerId) == aws.StringValue(transitGatewayVPCAttachment.VpcOwnerId) {
		if transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId); transitGatewayRouteTableID != "" {
			_, err := FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, d.Id())

			if tfresource.NotFound(err) {
				transitGatewayDefaultRouteTableAssociation = false
			} else if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Association (%s): %s", TransitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, d.Id()), err)
			}
		} else {
			transitGatewayDefaultRouteTableAssociation = false
		}

		if transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId); transitGatewayRouteTableID != "" {
			_, err := FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, d.Id())

			if tfresource.NotFound(err) {
				transitGatewayDefaultRouteTablePropagation = false
			} else if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table Propagation (%s): %s", TransitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, d.Id()), err)
			}
		} else {
			transitGatewayDefaultRouteTablePropagation = false
		}
	}

	d.Set("appliance_mode_support", transitGatewayVPCAttachment.Options.ApplianceModeSupport)
	d.Set("dns_support", transitGatewayVPCAttachment.Options.DnsSupport)
	d.Set("ipv6_support", transitGatewayVPCAttachment.Options.Ipv6Support)
	d.Set("subnet_ids", aws.StringValueSlice(transitGatewayVPCAttachment.SubnetIds))
	d.Set("transit_gateway_default_route_table_association", transitGatewayDefaultRouteTableAssociation)
	d.Set("transit_gateway_default_route_table_propagation", transitGatewayDefaultRouteTablePropagation)
	d.Set("transit_gateway_id", transitGatewayVPCAttachment.TransitGatewayId)
	d.Set("vpc_id", transitGatewayVPCAttachment.VpcId)
	d.Set("vpc_owner_id", transitGatewayVPCAttachment.VpcOwnerId)

	setTagsOut(ctx, transitGatewayVPCAttachment.Tags)

	return diags
}

func resourceTransitGatewayVPCAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	if d.HasChanges("appliance_mode_support", "dns_support", "ipv6_support", "subnet_ids") {
		input := &ec2.ModifyTransitGatewayVpcAttachmentInput{
			Options: &ec2.ModifyTransitGatewayVpcAttachmentRequestOptions{
				ApplianceModeSupport: aws.String(d.Get("appliance_mode_support").(string)),
				DnsSupport:           aws.String(d.Get("dns_support").(string)),
				Ipv6Support:          aws.String(d.Get("ipv6_support").(string)),
			},
			TransitGatewayAttachmentId: aws.String(d.Id()),
		}

		o, n := d.GetChange("subnet_ids")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		if add := ns.Difference(os); add.Len() > 0 {
			input.AddSubnetIds = flex.ExpandStringSet(add)
		}

		if del := os.Difference(ns); del.Len() > 0 {
			input.RemoveSubnetIds = flex.ExpandStringSet(del)
		}

		if _, err := conn.ModifyTransitGatewayVpcAttachmentWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
		}

		if _, err := WaitTransitGatewayVPCAttachmentUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway VPC Attachment (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChanges("transit_gateway_default_route_table_association", "transit_gateway_default_route_table_propagation") {
		transitGatewayID := d.Get("transit_gateway_id").(string)
		transitGateway, err := FindTransitGatewayByID(ctx, conn, transitGatewayID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
		}

		if d.HasChange("transit_gateway_default_route_table_association") {
			if err := transitGatewayRouteTableAssociationUpdate(ctx, conn, aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_association").(bool)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
			}
		}

		if d.HasChange("transit_gateway_default_route_table_propagation") {
			if err := transitGatewayRouteTablePropagationUpdate(ctx, conn, aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), d.Get("transit_gateway_default_route_table_propagation").(bool)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
			}
		}
	}

	return diags
}

func resourceTransitGatewayVPCAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway VPC Attachment: %s", d.Id())
	_, err := conn.DeleteTransitGatewayVpcAttachmentWithContext(ctx, &ec2.DeleteTransitGatewayVpcAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
	}

	if _, err := WaitTransitGatewayVPCAttachmentDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway VPC Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}
