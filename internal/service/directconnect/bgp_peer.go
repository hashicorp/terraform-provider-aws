// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_dx_bgp_peer", name="BGP Peer")
func resourceBGPPeer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBGPPeerCreate,
		ReadWithoutTimeout:   resourceBGPPeerRead,
		DeleteWithoutTimeout: resourceBGPPeerDelete,

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AddressFamily](),
			},
			"amazon_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_asn": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"bgp_auth_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"bgp_peer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"customer_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
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

func resourceBGPPeerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vifID := d.Get("virtual_interface_id").(string)
	addrFamily := awstypes.AddressFamily(d.Get("address_family").(string))
	asn := int32(d.Get("bgp_asn").(int))
	input := &directconnect.CreateBGPPeerInput{
		NewBGPPeer: &awstypes.NewBGPPeer{
			AddressFamily: addrFamily,
			Asn:           asn,
		},
		VirtualInterfaceId: aws.String(vifID),
	}

	if v, ok := d.GetOk("amazon_address"); ok {
		input.NewBGPPeer.AmazonAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok {
		input.NewBGPPeer.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok {
		input.NewBGPPeer.CustomerAddress = aws.String(v.(string))
	}

	_, err := conn.CreateBGPPeer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect BGP Peer: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s-%d", vifID, addrFamily, asn))

	if _, err = waitBGPPeerCreated(ctx, conn, vifID, addrFamily, asn, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect BGP Peer (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceBGPPeerRead(ctx, d, meta)...)
}

func resourceBGPPeerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vifID := d.Get("virtual_interface_id").(string)
	addrFamily := awstypes.AddressFamily(d.Get("address_family").(string))
	asn := int32(d.Get("bgp_asn").(int))
	bgpPeer, err := findBGPPeerByThreePartKey(ctx, conn, vifID, addrFamily, asn)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Direct Connect BGP Peer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect BGP Peer (%s): %s", d.Id(), err)
	}

	d.Set("amazon_address", bgpPeer.AmazonAddress)
	d.Set("aws_device", bgpPeer.AwsDeviceV2)
	d.Set("bgp_auth_key", bgpPeer.AuthKey)
	d.Set("bgp_peer_id", bgpPeer.BgpPeerId)
	d.Set("bgp_status", bgpPeer.BgpStatus)
	d.Set("customer_address", bgpPeer.CustomerAddress)

	return diags
}

func resourceBGPPeerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vifID := d.Get("virtual_interface_id").(string)
	addrFamily := awstypes.AddressFamily(d.Get("address_family").(string))
	asn := int32(d.Get("bgp_asn").(int))

	log.Printf("[DEBUG] Deleting Direct Connect BGP peer: %s", d.Id())
	input := directconnect.DeleteBGPPeerInput{
		Asn:                asn,
		CustomerAddress:    aws.String(d.Get("customer_address").(string)),
		VirtualInterfaceId: aws.String(vifID),
	}
	_, err := conn.DeleteBGPPeer(ctx, &input)

	if errs.IsAErrorMessageContains[*awstypes.DirectConnectClientException](err, "The last BGP Peer on a Virtual Interface cannot be deleted") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect BGP Peer (%s): %s", d.Id(), err)
	}

	if _, err = waitBGPPeerDeleted(ctx, conn, vifID, addrFamily, asn, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect BGP Peer (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findBGPPeerByThreePartKey(ctx context.Context, conn *directconnect.Client, vifID string, addrFamily awstypes.AddressFamily, asn int32) (*awstypes.BGPPeer, error) {
	vif, err := findVirtualInterfaceByID(ctx, conn, vifID)

	if err != nil {
		return nil, err
	}

	output, err := tfresource.AssertSingleValueResult(tfslices.Filter(vif.BgpPeers, func(v awstypes.BGPPeer) bool {
		return v.AddressFamily == addrFamily && v.Asn == asn
	}))

	if err != nil {
		return nil, err
	}

	if state := output.BgpPeerState; state == awstypes.BGPPeerStateDeleted {
		return nil, &retry.NotFoundError{
			Message: string(state),
		}
	}

	return output, nil
}

func statusBGPPeer(ctx context.Context, conn *directconnect.Client, vifID string, addrFamily awstypes.AddressFamily, asn int32) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findBGPPeerByThreePartKey(ctx, conn, vifID, addrFamily, asn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.BgpPeerState), nil
	}
}

func waitBGPPeerCreated(ctx context.Context, conn *directconnect.Client, vifID string, addrFamily awstypes.AddressFamily, asn int32, timeout time.Duration) (*awstypes.BGPPeer, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.BGPPeerStatePending),
		Target:     enum.Slice(awstypes.BGPPeerStateAvailable, awstypes.BGPPeerStateVerifying),
		Refresh:    statusBGPPeer(ctx, conn, vifID, addrFamily, asn),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BGPPeer); ok {
		return output, err
	}

	return nil, err
}

func waitBGPPeerDeleted(ctx context.Context, conn *directconnect.Client, vifID string, addrFamily awstypes.AddressFamily, asn int32, timeout time.Duration) (*awstypes.BGPPeer, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.BGPPeerStateAvailable, awstypes.BGPPeerStateDeleting, awstypes.BGPPeerStatePending, awstypes.BGPPeerStateVerifying),
		Target:     []string{},
		Refresh:    statusBGPPeer(ctx, conn, vifID, addrFamily, asn),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BGPPeer); ok {
		return output, err
	}

	return nil, err
}
