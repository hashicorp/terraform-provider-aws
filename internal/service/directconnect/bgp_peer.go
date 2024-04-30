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
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_dx_bgp_peer")
func ResourceBGPPeer() *schema.Resource {
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
			"bgp_asn": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"amazon_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"bgp_auth_key": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"customer_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"bgp_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bgp_peer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"aws_device": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceBGPPeerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vifId := d.Get("virtual_interface_id").(string)
	addrFamily := d.Get("address_family").(string)
	asn := int32(d.Get("bgp_asn").(int))

	req := &directconnect.CreateBGPPeerInput{
		VirtualInterfaceId: aws.String(vifId),
		NewBGPPeer: &awstypes.NewBGPPeer{
			AddressFamily: awstypes.AddressFamily(addrFamily),
			Asn:           asn,
		},
	}
	if v, ok := d.GetOk("amazon_address"); ok {
		req.NewBGPPeer.AmazonAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok {
		req.NewBGPPeer.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok {
		req.NewBGPPeer.CustomerAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Direct Connect BGP peer: %#v", req)
	_, err := conn.CreateBGPPeer(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Direct Connect BGP peer: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s-%d", vifId, addrFamily, asn))

	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.BGPPeerStatePending),
		Target:     enum.Slice(awstypes.BGPPeerStateAvailable, awstypes.BGPPeerStateVerifying),
		Refresh:    bgpPeerStateRefresh(ctx, conn, vifId, addrFamily, asn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Direct Connect BGP peer (%s) to be available: %s", d.Id(), err)
	}

	return append(diags, resourceBGPPeerRead(ctx, d, meta)...)
}

func resourceBGPPeerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vifId := d.Get("virtual_interface_id").(string)
	addrFamily := d.Get("address_family").(string)
	asn := int32(d.Get("bgp_asn").(int))

	bgpPeerRaw, state, err := bgpPeerStateRefresh(ctx, conn, vifId, addrFamily, asn)()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Direct Connect BGP peer: %s", err)
	}
	if state == string(awstypes.BGPPeerStateDeleted) {
		log.Printf("[WARN] Direct Connect BGP peer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	bgpPeer := bgpPeerRaw.(*awstypes.BGPPeer)
	d.Set("amazon_address", bgpPeer.AmazonAddress)
	d.Set("bgp_auth_key", bgpPeer.AuthKey)
	d.Set("customer_address", bgpPeer.CustomerAddress)
	d.Set("bgp_status", bgpPeer.BgpStatus)
	d.Set("bgp_peer_id", bgpPeer.BgpPeerId)
	d.Set("aws_device", bgpPeer.AwsDeviceV2)

	return diags
}

func resourceBGPPeerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectClient(ctx)

	vifId := d.Get("virtual_interface_id").(string)
	addrFamily := d.Get("address_family").(string)
	asn := int32(d.Get("bgp_asn").(int))

	log.Printf("[DEBUG] Deleting Direct Connect BGP peer: %s", d.Id())
	_, err := conn.DeleteBGPPeer(ctx, &directconnect.DeleteBGPPeerInput{
		Asn:                asn,
		CustomerAddress:    aws.String(d.Get("customer_address").(string)),
		VirtualInterfaceId: aws.String(vifId),
	})
	if err != nil {
		// This is the error returned if the BGP peering has already gone.
		if tfawserr.ErrMessageContains(err, "DirectConnectClientException", "The last BGP Peer on a Virtual Interface cannot be deleted") {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect BGP Peer (%s): %s", d.Id(), err)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.BGPPeerStateAvailable, awstypes.BGPPeerStateDeleting, awstypes.BGPPeerStatePending, awstypes.BGPPeerStateVerifying),
		Target:     enum.Slice(awstypes.BGPPeerStateDeleted),
		Refresh:    bgpPeerStateRefresh(ctx, conn, vifId, addrFamily, asn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Direct Connect BGP Peer (%s): waiting for completion: %s", d.Id(), err)
	}

	return diags
}

func bgpPeerStateRefresh(ctx context.Context, conn *directconnect.Client, vifId, addrFamily string, asn int32) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vif, err := virtualInterfaceRead(ctx, vifId, conn)
		if err != nil {
			return nil, "", err
		}
		if vif == nil {
			return "", string(awstypes.BGPPeerStateDeleted), nil
		}

		for _, bgpPeer := range vif.BgpPeers {
			if string(bgpPeer.AddressFamily) == addrFamily && bgpPeer.Asn == asn {
				return bgpPeer, string(bgpPeer.BgpPeerState), nil
			}
		}

		return "", string(awstypes.BGPPeerStateDeleted), nil
	}
}
