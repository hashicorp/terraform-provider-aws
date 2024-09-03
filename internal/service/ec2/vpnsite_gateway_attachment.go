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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpn_gateway_attachment", name="VPN Gateway Attachment")
func resourceVPNGatewayAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPNGatewayAttachmentCreate,
		ReadWithoutTimeout:   resourceVPNGatewayAttachmentRead,
		DeleteWithoutTimeout: resourceVPNGatewayAttachmentDelete,

		Schema: map[string]*schema.Schema{
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPNGatewayAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcID := d.Get(names.AttrVPCID).(string)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)
	input := &ec2.AttachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	}

	log.Printf("[DEBUG] Creating EC2 VPN Gateway Attachment: %s", vpnGatewayID)
	_, err := conn.AttachVpnGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPN Gateway (%s) Attachment (%s): %s", vpnGatewayID, vpcID, err)
	}

	d.SetId(vpnGatewayVPCAttachmentCreateID(vpnGatewayID, vpcID))

	_, err = waitVPNGatewayVPCAttachmentAttached(ctx, conn, vpnGatewayID, vpcID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Gateway (%s) Attachment (%s) to become attached: %s", vpnGatewayID, vpcID, err)
	}

	return append(diags, resourceVPNGatewayAttachmentRead(ctx, d, meta)...)
}

func resourceVPNGatewayAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcID := d.Get(names.AttrVPCID).(string)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)

	_, err := findVPNGatewayVPCAttachmentByTwoPartKey(ctx, conn, vpnGatewayID, vpcID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPN Gateway (%s) Attachment (%s) not found, removing from state", vpnGatewayID, vpcID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPN Gateway (%s) Attachment (%s): %s", vpnGatewayID, vpcID, err)
	}

	return diags
}

func resourceVPNGatewayAttachmentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vpcID := d.Get(names.AttrVPCID).(string)
	vpnGatewayID := d.Get("vpn_gateway_id").(string)

	log.Printf("[INFO] Deleting EC2 VPN Gateway (%s) Attachment (%s)", vpnGatewayID, vpcID)
	_, err := conn.DetachVpnGateway(ctx, &ec2.DetachVpnGatewayInput{
		VpcId:        aws.String(vpcID),
		VpnGatewayId: aws.String(vpnGatewayID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVPNGatewayAttachmentNotFound, errCodeInvalidVPNGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPN Gateway (%s) Attachment (%s): %s", vpnGatewayID, vpcID, err)
	}

	_, err = waitVPNGatewayVPCAttachmentDetached(ctx, conn, vpnGatewayID, vpcID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EC2 VPN Gateway (%s) Attachment (%s) to become detached: %s", vpnGatewayID, vpcID, err)
	}

	return diags
}
