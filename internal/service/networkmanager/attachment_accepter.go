// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkmanager/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// AttachmentAccepter does not require AttachmentType. However, querying attachments for status updates requires knowing type
// To facilitate querying and waiters on specific attachment types, attachment_type set to required

// @SDKResource("aws_networkmanager_attachment_accepter", name="Attachment Accepter")
func resourceAttachmentAccepter() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AttachmentType](),
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
			"edge_locations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
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

func resourceAttachmentAccepterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	var state awstypes.AttachmentState
	attachmentID := d.Get("attachment_id").(string)
	attachmentType := awstypes.AttachmentType(d.Get("attachment_type").(string))

	switch attachmentType {
	case awstypes.AttachmentTypeVpc:
		vpcAttachment, err := findVPCAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s): %s", attachmentID, err)
		}

		state = vpcAttachment.Attachment.State

		d.SetId(attachmentID)

	case awstypes.AttachmentTypeSiteToSiteVpn:
		vpnAttachment, err := findSiteToSiteVPNAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Site To Site VPN Attachment (%s): %s", attachmentID, err)
		}

		state = vpnAttachment.Attachment.State

		d.SetId(attachmentID)

	case awstypes.AttachmentTypeConnect:
		connectAttachment, err := findConnectAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Attachment (%s): %s", attachmentID, err)
		}

		state = connectAttachment.Attachment.State

		d.SetId(attachmentID)

	case awstypes.AttachmentTypeTransitGatewayRouteTable:
		tgwAttachment, err := findTransitGatewayRouteTableAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Route Table Attachment (%s): %s", attachmentID, err)
		}

		state = tgwAttachment.Attachment.State

		d.SetId(attachmentID)

	case awstypes.AttachmentTypeDirectConnectGateway:
		dxgwAttachment, err := findDirectConnectGatewayAttachmentByID(ctx, conn, attachmentID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Direct Connect Gateway Attachment (%s): %s", attachmentID, err)
		}

		state = dxgwAttachment.Attachment.State

		d.SetId(attachmentID)

	default:
		return sdkdiag.AppendErrorf(diags, "unsupported Network Manager Attachment type: %s", attachmentType)
	}

	if state == awstypes.AttachmentStatePendingAttachmentAcceptance || state == awstypes.AttachmentStatePendingTagAcceptance {
		input := &networkmanager.AcceptAttachmentInput{
			AttachmentId: aws.String(attachmentID),
		}

		_, err := conn.AcceptAttachment(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "accepting Network Manager Attachment (%s): %s", attachmentID, err)
		}

		switch attachmentType {
		case awstypes.AttachmentTypeVpc:
			if _, err := waitVPCAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager VPC Attachment (%s) to be attached: %s", attachmentID, err)
			}

		case awstypes.AttachmentTypeSiteToSiteVpn:
			if _, err := waitSiteToSiteVPNAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Site To Site VPN Attachment (%s) create: %s", attachmentID, err)
			}

		case awstypes.AttachmentTypeConnect:
			if _, err := waitConnectAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Connect Attachment (%s) create: %s", attachmentID, err)
			}

		case awstypes.AttachmentTypeTransitGatewayRouteTable:
			if _, err := waitTransitGatewayRouteTableAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Transit Gateway Route Table Attachment (%s) create: %s", attachmentID, err)
			}

		case awstypes.AttachmentTypeDirectConnectGateway:
			if _, err := waitDirectConnectGatewayAttachmentAvailable(ctx, conn, attachmentID, d.Timeout(schema.TimeoutCreate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Direct Connect Gateway Attachment (%s) create: %s", attachmentID, err)
			}
		}
	}

	return append(diags, resourceAttachmentAccepterRead(ctx, d, meta)...)
}

func resourceAttachmentAccepterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	var attachment *awstypes.Attachment
	switch attachmentType := awstypes.AttachmentType(d.Get("attachment_type").(string)); attachmentType {
	case awstypes.AttachmentTypeVpc:
		vpcAttachment, err := findVPCAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager VPC Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager VPC Attachment (%s): %s", d.Id(), err)
		}

		attachment = vpcAttachment.Attachment
		d.Set("edge_location", attachment.EdgeLocation)
		d.Set("edge_locations", nil)

	case awstypes.AttachmentTypeSiteToSiteVpn:
		vpnAttachment, err := findSiteToSiteVPNAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager Site To Site VPN Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Site To Site VPN Attachment (%s): %s", d.Id(), err)
		}

		attachment = vpnAttachment.Attachment
		d.Set("edge_location", attachment.EdgeLocation)
		d.Set("edge_locations", nil)

	case awstypes.AttachmentTypeConnect:
		connectAttachment, err := findConnectAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager Connect Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Connect Attachment (%s): %s", d.Id(), err)
		}

		attachment = connectAttachment.Attachment
		d.Set("edge_location", attachment.EdgeLocation)
		d.Set("edge_locations", nil)

	case awstypes.AttachmentTypeTransitGatewayRouteTable:
		tgwAttachment, err := findTransitGatewayRouteTableAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager Transit Gateway Route Table Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Transit Gateway Route Table Attachment (%s): %s", d.Id(), err)
		}

		attachment = tgwAttachment.Attachment
		d.Set("edge_location", attachment.EdgeLocation)
		d.Set("edge_locations", nil)

	case awstypes.AttachmentTypeDirectConnectGateway:
		dxgwAttachment, err := findDirectConnectGatewayAttachmentByID(ctx, conn, d.Id())

		if !d.IsNewResource() && tfresource.NotFound(err) {
			log.Printf("[WARN] Network Manager Direct Connect Gateway Attachment %s not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Network Manager Direct Connect Gateway Attachment (%s): %s", d.Id(), err)
		}

		attachment = dxgwAttachment.Attachment
		d.Set("edge_location", nil)
		d.Set("edge_locations", attachment.EdgeLocations)
	}

	d.Set("attachment_policy_rule_number", attachment.AttachmentPolicyRuleNumber)
	d.Set("core_network_arn", attachment.CoreNetworkArn)
	d.Set("core_network_id", attachment.CoreNetworkId)
	d.Set(names.AttrOwnerAccountID, attachment.OwnerAccountId)
	d.Set(names.AttrResourceARN, attachment.ResourceArn)
	d.Set("segment_name", attachment.SegmentName)
	d.Set(names.AttrState, attachment.State)

	return diags
}

func resourceAttachmentAccepterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkManagerClient(ctx)

	switch attachmentType := awstypes.AttachmentType(d.Get("attachment_type").(string)); attachmentType {
	case awstypes.AttachmentTypeVpc:
		_, err := conn.DeleteAttachment(ctx, &networkmanager.DeleteAttachmentInput{
			AttachmentId: aws.String(d.Id()),
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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
