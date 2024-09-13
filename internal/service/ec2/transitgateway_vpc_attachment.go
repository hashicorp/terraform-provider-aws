// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_transit_gateway_vpc_attachment", name="Transit Gateway VPC Attachment")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTransitGatewayVPCAttachment() *schema.Resource {
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ApplianceModeSupportValueDisable,
				ValidateDiagFunc: enum.Validate[awstypes.ApplianceModeSupportValue](),
			},
			"dns_support": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.DnsSupportValueEnable,
				ValidateDiagFunc: enum.Validate[awstypes.DnsSupportValue](),
			},
			"ipv6_support": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.Ipv6SupportValueDisable,
				ValidateDiagFunc: enum.Validate[awstypes.Ipv6SupportValue](),
			},
			names.AttrSubnetIDs: {
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
				Computed: true,
			},
			"transit_gateway_default_route_table_propagation": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrTransitGatewayID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			names.AttrVPCID: {
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
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayID := d.Get(names.AttrTransitGatewayID).(string)
	input := &ec2.CreateTransitGatewayVpcAttachmentInput{
		Options: &awstypes.CreateTransitGatewayVpcAttachmentRequestOptions{
			ApplianceModeSupport: awstypes.ApplianceModeSupportValue(d.Get("appliance_mode_support").(string)),
			DnsSupport:           awstypes.DnsSupportValue(d.Get("dns_support").(string)),
			Ipv6Support:          awstypes.Ipv6SupportValue(d.Get("ipv6_support").(string)),
		},
		SubnetIds:         flex.ExpandStringValueSet(d.Get(names.AttrSubnetIDs).(*schema.Set)),
		TransitGatewayId:  aws.String(transitGatewayID),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeTransitGatewayAttachment),
		VpcId:             aws.String(d.Get(names.AttrVPCID).(string)),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway VPC Attachment: %+v", input)
	output, err := conn.CreateTransitGatewayVpcAttachment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway VPC Attachment: %s", err)
	}

	d.SetId(aws.ToString(output.TransitGatewayVpcAttachment.TransitGatewayAttachmentId))

	if _, err := waitTransitGatewayVPCAttachmentCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway VPC Attachment (%s) create: %s", d.Id(), err)
	}

	transitGateway, err := findTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	// We cannot modify Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways.
	if aws.ToString(transitGateway.OwnerId) == aws.ToString(output.TransitGatewayVpcAttachment.VpcOwnerId) {
		// Default values of transit_gateway_default_route_table_association and transit_gateway_default_route_table_propagation are both 'true'.
		transitGatewayDefaultRouteTableAssociation := true
		if v := d.GetRawConfig().GetAttr("transit_gateway_default_route_table_association"); v.IsKnown() && !v.IsNull() {
			transitGatewayDefaultRouteTableAssociation = v.True()
		}

		if err := transitGatewayRouteTableAssociationUpdate(ctx, conn, aws.ToString(transitGateway.Options.AssociationDefaultRouteTableId), d.Id(), transitGatewayDefaultRouteTableAssociation); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway VPC Attachment: %s", err)
		}

		transitGatewayDefaultRouteTablePropagation := true
		if v := d.GetRawConfig().GetAttr("transit_gateway_default_route_table_propagation"); v.IsKnown() && !v.IsNull() {
			transitGatewayDefaultRouteTablePropagation = v.True()
		}

		if err := transitGatewayRouteTablePropagationUpdate(ctx, conn, aws.ToString(transitGateway.Options.PropagationDefaultRouteTableId), d.Id(), transitGatewayDefaultRouteTablePropagation); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Transit Gateway VPC Attachment: %s", err)
		}
	}

	return append(diags, resourceTransitGatewayVPCAttachmentRead(ctx, d, meta)...)
}

func resourceTransitGatewayVPCAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	// We cannot read Transit Gateway Route Tables for Resource Access Manager shared Transit Gateways.
	transitGatewayDefaultRouteTableAssociation := true
	transitGatewayDefaultRouteTablePropagation := true

	if aws.ToString(transitGateway.OwnerId) == aws.ToString(transitGatewayVPCAttachment.VpcOwnerId) {
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
	}

	d.Set("appliance_mode_support", transitGatewayVPCAttachment.Options.ApplianceModeSupport)
	d.Set("dns_support", transitGatewayVPCAttachment.Options.DnsSupport)
	d.Set("ipv6_support", transitGatewayVPCAttachment.Options.Ipv6Support)
	d.Set(names.AttrSubnetIDs, transitGatewayVPCAttachment.SubnetIds)
	d.Set("transit_gateway_default_route_table_association", transitGatewayDefaultRouteTableAssociation)
	d.Set("transit_gateway_default_route_table_propagation", transitGatewayDefaultRouteTablePropagation)
	d.Set(names.AttrTransitGatewayID, transitGatewayVPCAttachment.TransitGatewayId)
	d.Set(names.AttrVPCID, transitGatewayVPCAttachment.VpcId)
	d.Set("vpc_owner_id", transitGatewayVPCAttachment.VpcOwnerId)

	setTagsOut(ctx, transitGatewayVPCAttachment.Tags)

	return diags
}

func resourceTransitGatewayVPCAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChanges("appliance_mode_support", "dns_support", "ipv6_support", names.AttrSubnetIDs) {
		input := &ec2.ModifyTransitGatewayVpcAttachmentInput{
			Options: &awstypes.ModifyTransitGatewayVpcAttachmentRequestOptions{
				ApplianceModeSupport: awstypes.ApplianceModeSupportValue(d.Get("appliance_mode_support").(string)),
				DnsSupport:           awstypes.DnsSupportValue(d.Get("dns_support").(string)),
				Ipv6Support:          awstypes.Ipv6SupportValue(d.Get("ipv6_support").(string)),
			},
			TransitGatewayAttachmentId: aws.String(d.Id()),
		}

		o, n := d.GetChange(names.AttrSubnetIDs)
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		if add := ns.Difference(os); add.Len() > 0 {
			input.AddSubnetIds = flex.ExpandStringValueSet(add)
		}

		if del := os.Difference(ns); del.Len() > 0 {
			input.RemoveSubnetIds = flex.ExpandStringValueSet(del)
		}

		if _, err := conn.ModifyTransitGatewayVpcAttachment(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Transit Gateway VPC Attachment (%s): %s", d.Id(), err)
		}

		if _, err := waitTransitGatewayVPCAttachmentUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway VPC Attachment (%s) update: %s", d.Id(), err)
		}
	}

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

func resourceTransitGatewayVPCAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway VPC Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}
