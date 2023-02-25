package directconnect

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceBGPPeer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBGPPeerCreate,
		ReadWithoutTimeout:   resourceBGPPeerRead,
		DeleteWithoutTimeout: resourceBGPPeerDelete,

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{directconnect.AddressFamilyIpv4, directconnect.AddressFamilyIpv6}, false),
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
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	vifId := d.Get("virtual_interface_id").(string)
	addrFamily := d.Get("address_family").(string)
	asn := int64(d.Get("bgp_asn").(int))

	req := &directconnect.CreateBGPPeerInput{
		VirtualInterfaceId: aws.String(vifId),
		NewBGPPeer: &directconnect.NewBGPPeer{
			AddressFamily: aws.String(addrFamily),
			Asn:           aws.Int64(asn),
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
	_, err := conn.CreateBGPPeerWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error creating Direct Connect BGP peer: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s-%d", vifId, addrFamily, asn))

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			directconnect.BGPPeerStatePending,
		},
		Target: []string{
			directconnect.BGPPeerStateAvailable,
			directconnect.BGPPeerStateVerifying,
		},
		Refresh:    bgpPeerStateRefresh(ctx, conn, vifId, addrFamily, asn),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error waiting for Direct Connect BGP peer (%s) to be available: %s", d.Id(), err)
	}

	return append(diags, resourceBGPPeerRead(ctx, d, meta)...)
}

func resourceBGPPeerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	vifId := d.Get("virtual_interface_id").(string)
	addrFamily := d.Get("address_family").(string)
	asn := int64(d.Get("bgp_asn").(int))

	bgpPeerRaw, state, err := bgpPeerStateRefresh(ctx, conn, vifId, addrFamily, asn)()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "Error reading Direct Connect BGP peer: %s", err)
	}
	if state == directconnect.BGPPeerStateDeleted {
		log.Printf("[WARN] Direct Connect BGP peer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	bgpPeer := bgpPeerRaw.(*directconnect.BGPPeer)
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
	conn := meta.(*conns.AWSClient).DirectConnectConn()

	vifId := d.Get("virtual_interface_id").(string)
	addrFamily := d.Get("address_family").(string)
	asn := int64(d.Get("bgp_asn").(int))

	log.Printf("[DEBUG] Deleting Direct Connect BGP peer: %s", d.Id())
	_, err := conn.DeleteBGPPeerWithContext(ctx, &directconnect.DeleteBGPPeerInput{
		Asn:                aws.Int64(asn),
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

	stateConf := &resource.StateChangeConf{
		Pending: []string{
			directconnect.BGPPeerStateAvailable,
			directconnect.BGPPeerStateDeleting,
			directconnect.BGPPeerStatePending,
			directconnect.BGPPeerStateVerifying,
		},
		Target: []string{
			directconnect.BGPPeerStateDeleted,
		},
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

func bgpPeerStateRefresh(ctx context.Context, conn *directconnect.DirectConnect, vifId, addrFamily string, asn int64) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vif, err := virtualInterfaceRead(ctx, vifId, conn)
		if err != nil {
			return nil, "", err
		}
		if vif == nil {
			return "", directconnect.BGPPeerStateDeleted, nil
		}

		for _, bgpPeer := range vif.BgpPeers {
			if aws.StringValue(bgpPeer.AddressFamily) == addrFamily && aws.Int64Value(bgpPeer.Asn) == asn {
				return bgpPeer, aws.StringValue(bgpPeer.BgpPeerState), nil
			}
		}

		return "", directconnect.BGPPeerStateDeleted, nil
	}
}
