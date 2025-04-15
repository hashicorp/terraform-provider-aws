// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dx_hosted_private_virtual_interface_accepter", name="Hosted Private Virtual Interface Accepter")
// @Tags(identifierAttribute="arn")
func resourceHostedPrivateVirtualInterfaceAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedPrivateVirtualInterfaceAccepterCreate,
		ReadWithoutTimeout:   resourceHostedPrivateVirtualInterfaceAccepterRead,
		UpdateWithoutTimeout: resourceHostedPrivateVirtualInterfaceAccepterUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHostedPrivateVirtualInterfaceAccepterImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dx_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"dx_gateway_id", "vpn_gateway_id"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"vpn_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"dx_gateway_id", "vpn_gateway_id"},
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceHostedPrivateVirtualInterfaceAccepterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vifID := d.Get("virtual_interface_id").(string)
	input := &directconnect.ConfirmPrivateVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(vifID),
	}

	if v, ok := d.GetOk("dx_gateway_id"); ok {
		input.DirectConnectGatewayId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpn_gateway_id"); ok {
		input.VirtualGatewayId = aws.String(v.(string))
	}

	_, err := conn.ConfirmPrivateVirtualInterface(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Direct Connect Hosted Private Virtual Interface (%s): %s", vifID, err)
	}

	d.SetId(vifID)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Service:   "directconnect",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)

	if _, err := waitHostedPrivateVirtualInterfaceAccepterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Hosted Private Virtual Interface Accepter (%s) create: %s", d.Id(), err)
	}

	if err := createTags(ctx, conn, arn, getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Direct Connect Hosted Private Virtual Interface (%s) tags: %s", arn, err)
	}

	return append(diags, resourceHostedPrivateVirtualInterfaceAccepterUpdate(ctx, d, meta)...)
}

func resourceHostedPrivateVirtualInterfaceAccepterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Hosted Private Virtual Interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Hosted Private Virtual Interface (%s): %s", d.Id(), err)
	}

	if state := vif.VirtualInterfaceState; state != awstypes.VirtualInterfaceStateAvailable && state != awstypes.VirtualInterfaceStateDown {
		log.Printf("[WARN] Direct Connect Hosted Private Virtual Interface (%s) is '%s', removing from state", d.Id(), state)
		d.SetId("")
		return diags
	}

	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	d.Set("virtual_interface_id", vif.VirtualInterfaceId)
	d.Set("vpn_gateway_id", vif.VirtualGatewayId)

	return diags
}

func resourceHostedPrivateVirtualInterfaceAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, virtualInterfaceUpdate(ctx, d, meta)...)
	if diags.HasError() {
		return diags
	}

	return append(diags, resourceHostedPrivateVirtualInterfaceAccepterRead(ctx, d, meta)...)
}

func resourceHostedPrivateVirtualInterfaceAccepterImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if err != nil {
		return nil, err
	}

	if vifType := aws.ToString(vif.VirtualInterfaceType); vifType != "private" {
		return nil, fmt.Errorf("virtual interface (%s) has incorrect type: %s", d.Id(), vifType)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition(ctx),
		Region:    meta.(*conns.AWSClient).Region(ctx),
		Service:   "directconnect",
		AccountID: meta.(*conns.AWSClient).AccountID(ctx),
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)

	return []*schema.ResourceData{d}, nil
}

func waitHostedPrivateVirtualInterfaceAccepterAvailable(ctx context.Context, conn *directconnect.Client, id string, timeout time.Duration) (*awstypes.VirtualInterface, error) {
	return waitVirtualInterfaceAvailable(
		ctx,
		conn,
		id,
		enum.Slice(awstypes.VirtualInterfaceStateConfirming, awstypes.VirtualInterfaceStatePending),
		enum.Slice(awstypes.VirtualInterfaceStateAvailable, awstypes.VirtualInterfaceStateDown),
		timeout,
	)
}
