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

// @SDKResource("aws_dx_hosted_public_virtual_interface_accepter", name="Hosted Public Virtual Interface Accepter")
// @Tags(identifierAttribute="arn")
func resourceHostedPublicVirtualInterfaceAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHostedPublicVirtualInterfaceAccepterCreate,
		ReadWithoutTimeout:   resourceHostedPublicVirtualInterfaceAccepterRead,
		UpdateWithoutTimeout: resourceHostedPublicVirtualInterfaceAccepterUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: resourceHostedPublicVirtualInterfaceAccepterImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceHostedPublicVirtualInterfaceAccepterCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vifID := d.Get("virtual_interface_id").(string)
	input := &directconnect.ConfirmPublicVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(vifID),
	}

	_, err := conn.ConfirmPublicVirtualInterface(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting Direct Connect Hosted Public Virtual Interface (%s): %s", vifID, err)
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

	if _, err := waitHostedPublicVirtualInterfaceAccepterAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect Hosted Public Virtual Interface Accepter (%s) create: %s", d.Id(), err)
	}

	if err := createTags(ctx, conn, arn, getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Direct Connect Hosted Public Virtual Interface (%s) tags: %s", arn, err)
	}

	return append(diags, resourceHostedPublicVirtualInterfaceAccepterUpdate(ctx, d, meta)...)
}

func resourceHostedPublicVirtualInterfaceAccepterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect Hosted Public Virtual Interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect Hosted Public Virtual Interface (%s): %s", d.Id(), err)
	}

	if state := vif.VirtualInterfaceState; state != awstypes.VirtualInterfaceStateAvailable && state != awstypes.VirtualInterfaceStateDown && state != awstypes.VirtualInterfaceStateVerifying {
		log.Printf("[WARN] Direct Connect Hosted Public Virtual Interface (%s) is '%s', removing from state", d.Id(), state)
		d.SetId("")
		return diags
	}

	d.Set("virtual_interface_id", vif.VirtualInterfaceId)

	return diags
}

func resourceHostedPublicVirtualInterfaceAccepterUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	diags = append(diags, virtualInterfaceUpdate(ctx, d, meta)...)
	if diags.HasError() {
		return diags
	}

	return append(diags, resourceHostedPublicVirtualInterfaceAccepterRead(ctx, d, meta)...)
}

func resourceHostedPublicVirtualInterfaceAccepterImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vif, err := findVirtualInterfaceByID(ctx, conn, d.Id())

	if err != nil {
		return nil, err
	}

	if vifType := aws.ToString(vif.VirtualInterfaceType); vifType != "public" {
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

func waitHostedPublicVirtualInterfaceAccepterAvailable(ctx context.Context, conn *directconnect.Client, vifId string, timeout time.Duration) (*awstypes.VirtualInterface, error) {
	return waitVirtualInterfaceAvailable(
		ctx,
		conn,
		vifId,
		enum.Slice(awstypes.VirtualInterfaceStateConfirming, awstypes.VirtualInterfaceStatePending),
		enum.Slice(awstypes.VirtualInterfaceStateAvailable, awstypes.VirtualInterfaceStateDown, awstypes.VirtualInterfaceStateVerifying),
		timeout,
	)
}
