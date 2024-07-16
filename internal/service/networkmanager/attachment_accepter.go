// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// AttachmentAccepter does not require AttachmentType. However, querying attachments for status updates requires knowing type
// To facilitate querying and waiters on specific attachment types, attachment_type set to required

// @SDKResource("aws_networkmanager_attachment_accepter")
func ResourceAttachmentAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAttachmentAccepterCreate,
		ReadWithoutTimeout:   resourceAttachmentAccepterRead,
		DeleteWithoutTimeout: resourceAttachmentAccepterDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(15 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"attachment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"attachment_policy_rule_number": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			// querying attachments requires knowing the type ahead of time
			// therefore type is required in provider, though not on the API
			"attachment_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(networkmanager.AttachmentType_Values(), false),
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"core_network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"edge_location": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerAccountID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"segment_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAttachmentAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	var state string
	attachmentID := d.Get("attachment_id").(string)
	attachmentType := d.Get("attachment_type").(string)

	switch attachmentType {
	case networkmanager.AttachmentTypeVpc:
		vpcAttachment, err := FindVPCAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s): %s", attachmentID, err)
		}

		state = aws.StringValue(vpcAttachment.Attachment.State)

		d.SetId(attachmentID)

	case networkmanager.AttachmentTypeSiteToSiteVpn:
		vpnAttachment, err := FindSiteToSiteVPNAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Site To Site VPN Attachment (%s): %s", attachmentID, err)
		}

		state = aws.StringValue(vpnAttachment.Attachment.State)

		d.SetId(attachmentID)

	case networkmanager.AttachmentTypeConnect:
		connectAttachment, err := FindConnectAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Attachment (%s): %s", attachmentID, err)
		}

		state = aws.StringValue(connectAttachment.Attachment.State)

		d.SetId(attachmentID)

	case networkmanager.AttachmentTypeTransitGatewayRouteTable:
		tgwAttachment, err := FindTransitGatewayRouteTableAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Route Table Attachment (%s): %s", attachmentID, err)
		}

		state = aws.StringValue(tgwAttachment.Attachment.State)

		d.SetId(attachmentID)

	default:
		return sdkdiag.AppendErrorf(diags, "unsupported Network Manager Attachment type: %s", attachmentType)
	}

	if state == networkmanager.AttachmentStatePendingAttachmentAcceptance || state == networkmanager.AttachmentStatePendingTagAcceptance {
		input := &networkmanager.AcceptAttachmentInput{
			AttachmentId: aws.String(attachmentID),
		}

		_, err := conn.AcceptAttachmentWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "accepting Network Manager Attachment (%s): %s", attachmentID, err)
		}

		switch attachmentType {
		case networkmanager.AttachmentTypeVpc:
			if _, err := waitVPCAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) to be attached: %s", attachmentID, err)
			}

		case networkmanager.AttachmentTypeSiteToSiteVpn:
			if _, err := waitSiteToSiteVPNAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPN Attachment (%s) create: %s", attachmentID, err)
			}

		case networkmanager.AttachmentTypeConnect:
			if _, err := waitConnectAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connect Attachment (%s) create: %s", attachmentID, err)
			}

		case networkmanager.AttachmentTypeTransitGatewayRouteTable:
			if _, err := waitTransitGatewayRouteTableAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Route Table Attachment (%s) create: %s", attachmentID, err)
			}
		}
	}

	return append(diags, resourceAttachmentAccepterRead(ctx, d, meta)...)
}

func resourceAttachmentAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	var a *networkmanager.Attachment

	switch aType := d.Get("attachment_type"); aType {
	case networkmanager.AttachmentTypeVpc:
		vpcAttachment, err := FindVPCAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager VPC Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}

		a = vpcAttachment.Attachment

	case networkmanager.AttachmentTypeSiteToSiteVpn:
		vpnAttachment, err := FindSiteToSiteVPNAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager Site To Site VPN Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Site To Site VPN Attachment (%s): %s", d.Id(), err)
		}

		a = vpnAttachment.Attachment

	case networkmanager.AttachmentTypeConnect:
		connectAttachment, err := FindConnectAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager Connect Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Attachment (%s): %s", d.Id(), err)
		}

		a = connectAttachment.Attachment

	case networkmanager.AttachmentTypeTransitGatewayRouteTable:
		tgwAttachment, err := FindTransitGatewayRouteTableAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager Transit Gateway Route Table Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Route Table Attachment (%s): %s", d.Id(), err)
		}

		a = tgwAttachment.Attachment
	}

	d.Set("attachment_policy_rule_number", a.AttachmentPolicyRuleNumber)
	d.Set("core_network_arn", a.CoreNetworkArn)
	d.Set("core_network_id", a.CoreNetworkId)
	d.Set("edge_location", a.EdgeLocation)
	d.Set(names.AttrOwnerAccountID, a.OwnerAccountId)
	d.Set(names.AttrResourceARN, a.ResourceArn)
	d.Set("segment_name", a.SegmentName)
	d.Set(names.AttrState, a.State)

	return diags
}

func resourceAttachmentAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	switch d.Get("attachment_type") {
	case networkmanager.AttachmentTypeVpc:
		_, err := conn.DeleteAttachmentWithContext(ctx, &networkmanager.DeleteAttachmentInput{
			AttachmentId: aws.String(d.Id()),
		})

		if tfawserr.ErrCodeEquals(err, networkmanager.ErrCodeResourceNotFoundException) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}

		if _, err := waitVPCAttachmentDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) delete: %s", d.Id(), err)
		}
	}

	return diags
}
