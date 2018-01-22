package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxHostedPublicVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxHostedPublicVirtualInterfaceCreate,
		Read:   resourceAwsDxHostedPublicVirtualInterfaceRead,
		Delete: resourceAwsDxHostedPublicVirtualInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: mergeSchemas(
			dxVirtualInterfaceSchema,
			map[string]*schema.Schema{
				"owner_account_id": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validateAwsAccountId,
				},
				"route_filter_prefixes": &schema.Schema{
					Type:     schema.TypeSet,
					Required: true,
					ForceNew: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
					MinItems: 1,
				},
			},
		),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxHostedPublicVirtualInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	addressFamily := d.Get("address_family").(string)
	caRaw, caOk := d.GetOk("customer_address")
	aaRaw, aaOk := d.GetOk("amazon_address")
	if addressFamily == directconnect.AddressFamilyIpv4 {
		if !caOk {
			return fmt.Errorf("'customer_address' must be set when 'address_family' is '%s'", addressFamily)
		}
		if !aaOk {
			return fmt.Errorf("'amazon_address' must be set when 'address_family' is '%s'", addressFamily)
		}
	}

	req := &directconnect.AllocatePublicVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		OwnerAccount: aws.String(d.Get("owner_account_id").(string)),
		NewPublicVirtualInterfaceAllocation: &directconnect.NewPublicVirtualInterfaceAllocation{
			VirtualInterfaceName: aws.String(d.Get("name").(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
			Asn:                  aws.Int64(int64(d.Get("bgp_asn").(int))),
			AddressFamily:        aws.String(addressFamily),
		},
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok {
		req.NewPublicVirtualInterfaceAllocation.AuthKey = aws.String(v.(string))
	}
	if caOk {
		req.NewPublicVirtualInterfaceAllocation.CustomerAddress = aws.String(caRaw.(string))
	}
	if aaOk {
		req.NewPublicVirtualInterfaceAllocation.AmazonAddress = aws.String(aaRaw.(string))
	}
	if v, ok := d.GetOk("route_filter_prefixes"); ok {
		req.NewPublicVirtualInterfaceAllocation.RouteFilterPrefixes = expandDxRouteFilterPrefixes(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Allocating Direct Connect hosted public virtual interface: %#v", req)
	resp, err := conn.AllocatePublicVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("Error allocating Direct Connect hosted public virtual interface: %s", err.Error())
	}

	d.SetId(aws.StringValue(resp.VirtualInterfaceId))

	if err := dxHostedPublicVirtualInterfaceWaitUntilAvailable(d, conn); err != nil {
		return err
	}

	return resourceAwsDxHostedPublicVirtualInterfaceRead(d, meta)
}

func resourceAwsDxHostedPublicVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	vif, err := dxVirtualInterfaceRead(d, meta)
	if err != nil {
		return err
	}
	if vif == nil {
		return nil
	}

	d.Set("owner_account_id", vif.OwnerAccount)
	return dxPublicVirtualInterfaceAttributes(d, meta, vif)
}

func resourceAwsDxHostedPublicVirtualInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	return dxVirtualInterfaceDelete(d, meta)
}

func dxHostedPublicVirtualInterfaceWaitUntilAvailable(d *schema.ResourceData, conn *directconnect.DirectConnect) error {
	return dxVirtualInterfaceWaitUntilAvailable(
		d,
		conn,
		[]string{
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateConfirming,
			directconnect.VirtualInterfaceStateDown,
			directconnect.VirtualInterfaceStateVerifying,
		})
}
