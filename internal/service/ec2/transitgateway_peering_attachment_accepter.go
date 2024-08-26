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

// @SDKResource("aws_ec2_transit_gateway_peering_attachment_accepter", name="Transit Gateway Peering Attachment Accepter")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceTransitGatewayPeeringAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransitGatewayPeeringAttachmentAccepterCreate,
		ReadWithoutTimeout:   resourceTransitGatewayPeeringAttachmentAccepterRead,
		UpdateWithoutTimeout: resourceTransitGatewayPeeringAttachmentAccepterUpdate,
		DeleteWithoutTimeout: resourceTransitGatewayPeeringAttachmentAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"peer_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"peer_transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTransitGatewayAttachmentID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTransitGatewayID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTransitGatewayPeeringAttachmentAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayAttachmentID := d.Get(names.AttrTransitGatewayAttachmentID).(string)
	input := &ec2.AcceptTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
	}

	log.Printf("[DEBUG] Accepting EC2 Transit Gateway Peering Attachment: %s", transitGatewayAttachmentID)
	output, err := conn.AcceptTransitGatewayPeeringAttachment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting EC2 Transit Gateway Peering Attachment (%s): %s", transitGatewayAttachmentID, err)
	}

	d.SetId(aws.ToString(output.TransitGatewayPeeringAttachment.TransitGatewayAttachmentId))

	if _, err := waitTransitGatewayPeeringAttachmentAccepted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Peering Attachment (%s) update: %s", d.Id(), err)
	}

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EC2 Transit Gateway Peering Attachment (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceTransitGatewayPeeringAttachmentAccepterRead(ctx, d, meta)...)
}

func resourceTransitGatewayPeeringAttachmentAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	transitGatewayPeeringAttachment, err := findTransitGatewayPeeringAttachmentByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Peering Attachment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Peering Attachment (%s): %s", d.Id(), err)
	}

	transitGatewayID := aws.ToString(transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)
	_, err = findTransitGatewayByID(ctx, conn, transitGatewayID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway (%s): %s", transitGatewayID, err)
	}

	d.Set("peer_account_id", transitGatewayPeeringAttachment.RequesterTgwInfo.OwnerId)
	d.Set("peer_region", transitGatewayPeeringAttachment.RequesterTgwInfo.Region)
	d.Set("peer_transit_gateway_id", transitGatewayPeeringAttachment.RequesterTgwInfo.TransitGatewayId)
	d.Set(names.AttrTransitGatewayAttachmentID, transitGatewayPeeringAttachment.TransitGatewayAttachmentId)
	d.Set(names.AttrTransitGatewayID, transitGatewayPeeringAttachment.AccepterTgwInfo.TransitGatewayId)

	setTagsOut(ctx, transitGatewayPeeringAttachment.Tags)

	return diags
}

func resourceTransitGatewayPeeringAttachmentAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceTransitGatewayPeeringAttachmentAccepterRead(ctx, d, meta)...)
}

func resourceTransitGatewayPeeringAttachmentAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Peering Attachment: %s", d.Id())
	_, err := conn.DeleteTransitGatewayPeeringAttachment(ctx, &ec2.DeleteTransitGatewayPeeringAttachmentInput{
		TransitGatewayAttachmentId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidTransitGatewayAttachmentIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Transit Gateway Peering Attachment (%s): %s", d.Id(), err)
	}

	if err := waitTransitGatewayPeeringAttachmentDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 Transit Gateway Peering Attachment (%s) delete: %s", d.Id(), err)
	}

	return diags
}
