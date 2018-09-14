package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsDxBgpPeer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxBgpPeerCreate,
		Read:   resourceAwsDxBgpPeerRead,
		Delete: resourceAwsDxBgpPeerDelete,

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
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxBgpPeerCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	req := &directconnect.CreateBGPPeerInput{
		VirtualInterfaceId: aws.String(d.Get("virtual_interface_id").(string)),
		NewBGPPeer: &directconnect.NewBGPPeer{
			AddressFamily: aws.String(d.Get("address_family").(string)),
			Asn:           aws.Int64(int64(d.Get("bgp_asn").(int))),
		},
	}
	if v, ok := d.GetOk("amazon_address"); ok && v.(string) != "" {
		req.NewBGPPeer.AmazonAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok && v.(string) != "" {
		req.NewBGPPeer.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok && v.(string) != "" {
		req.NewBGPPeer.CustomerAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Direct Connect BGP peer: %#v", req)
	_, err := conn.CreateBGPPeer(req)
	if err != nil {
		return fmt.Errorf("Error creating Direct Connect BGP peer: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s-%d", d.Get("virtual_interface_id"), d.Get("address_family"), d.Get("bgp_asn")))

	return resourceAwsDxBgpPeerRead(d, meta)
}

func resourceAwsDxBgpPeerRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vifId := d.Get("virtual_interface_id").(string)
	vif, err := dxVirtualInterfaceRead(vifId, conn)
	if err != nil {
		return err
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect virtual interface (%s) not found, removing BGP peer from state", vifId)
		d.SetId("")
		return nil
	}

	var bgpPeer *directconnect.BGPPeer
	for _, peer := range vif.BgpPeers {
		if aws.StringValue(peer.AddressFamily) == d.Get("address_family").(string) &&
			aws.Int64Value(peer.Asn) == int64(d.Get("bgp_asn").(int)) {
			bgpPeer = peer
			break
		}
	}
	if bgpPeer == nil {
		log.Printf("[WARN] Direct Connect BGP peer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("address_family", bgpPeer.AddressFamily)
	d.Set("bgp_asn", bgpPeer.Asn)
	d.Set("virtual_interface_id", vif.VirtualInterfaceId)
	d.Set("amazon_address", bgpPeer.AmazonAddress)
	d.Set("bgp_auth_key", bgpPeer.AuthKey)
	d.Set("customer_address", bgpPeer.CustomerAddress)
	d.Set("bgp_status", bgpPeer.BgpStatus)

	return nil
}

func resourceAwsDxBgpPeerDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	log.Printf("[DEBUG] Deleting Direct Connect BGP peer: %s", d.Id())
	_, err := conn.DeleteBGPPeer(&directconnect.DeleteBGPPeerInput{
		Asn:                aws.Int64(int64(d.Get("bgp_asn").(int))),
		CustomerAddress:    aws.String(d.Get("customer_address").(string)),
		VirtualInterfaceId: aws.String(d.Get("virtual_interface_id").(string)),
	})
	if err != nil {
		if isAWSErr(err, "DirectConnectClientException", "XXX") {
			return nil
		}
		return err
	}

	return nil
}
