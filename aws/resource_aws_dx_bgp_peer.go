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
		Create:        resourceAwsDxBgpPeerCreate,
		Read:          resourceAwsDxBgpPeerRead,
		Delete:        resourceAwsDxBgpPeerDelete,
		CustomizeDiff: resourceAwsDxBgpPeerCustomizeDiff,

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

	// if err := dxPublicVirtualInterfaceWaitUntilAvailable(d, conn); err != nil {
	// 	return err
	// }

	return resourceAwsDxBgpPeerRead(d, meta)
}

func resourceAwsDxBgpPeerRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsDxBgpPeerDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceAwsDxBgpPeerCustomizeDiff(diff *schema.ResourceDiff, meta interface{}) error {
	if diff.Id() == "" {
		// New resource.
		if addressFamily := diff.Get("address_family").(string); addressFamily == directconnect.AddressFamilyIpv4 {
			if _, ok := diff.GetOk("customer_address"); !ok {
				return fmt.Errorf("'customer_address' must be set when 'address_family' is '%s'", addressFamily)
			}
			if _, ok := diff.GetOk("amazon_address"); !ok {
				return fmt.Errorf("'amazon_address' must be set when 'address_family' is '%s'", addressFamily)
			}
		}
	}

	return nil
}
