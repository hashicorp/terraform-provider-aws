package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxPublicVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxPublicVirtualInterfaceCreate,
		Read:   resourceAwsDxPublicVirtualInterfaceRead,
		Update: resourceAwsDxPublicVirtualInterfaceUpdate,
		Delete: resourceAwsDxPublicVirtualInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: mergeSchemas(
			dxVirtualInterfaceSchemaWithTags,
			map[string]*schema.Schema{
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

func resourceAwsDxPublicVirtualInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
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

	req := &directconnect.CreatePublicVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		NewPublicVirtualInterface: &directconnect.NewPublicVirtualInterface{
			VirtualInterfaceName: aws.String(d.Get("name").(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
			Asn:                  aws.Int64(int64(d.Get("bgp_asn").(int))),
			AddressFamily:        aws.String(addressFamily),
		},
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok {
		req.NewPublicVirtualInterface.AuthKey = aws.String(v.(string))
	}
	if caOk {
		req.NewPublicVirtualInterface.CustomerAddress = aws.String(caRaw.(string))
	}
	if aaOk {
		req.NewPublicVirtualInterface.AmazonAddress = aws.String(aaRaw.(string))
	}
	if v, ok := d.GetOk("route_filter_prefixes"); ok {
		req.NewPublicVirtualInterface.RouteFilterPrefixes = expandDxRouteFilterPrefixes(v.(*schema.Set).List())
	}

	log.Printf("[DEBUG] Creating Direct Connect public virtual interface: %#v", req)
	resp, err := conn.CreatePublicVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("Error creating Direct Connect public virtual interface: %s", err.Error())
	}

	d.SetId(aws.StringValue(resp.VirtualInterfaceId))

	if err := dxPublicVirtualInterfaceWaitUntilAvailable(d, conn); err != nil {
		return err
	}

	return resourceAwsDxPublicVirtualInterfaceUpdate(d, meta)
}

func resourceAwsDxPublicVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d.Id(), conn)
	if err != nil {
		return err
	}
	if vif == nil {
		log.Printf("[WARN] Direct Connect virtual interface (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err := dxPublicVirtualInterfaceAttributes(d, meta, vif); err != nil {
		return err
	}
	if err := getTagsDX(conn, d, d.Get("arn").(string)); err != nil {
		return err
	}

	return nil
}

func resourceAwsDxPublicVirtualInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := dxVirtualInterfaceUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsDxPublicVirtualInterfaceRead(d, meta)
}

func resourceAwsDxPublicVirtualInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	return dxVirtualInterfaceDelete(d, meta)
}

func dxPublicVirtualInterfaceWaitUntilAvailable(d *schema.ResourceData, conn *directconnect.DirectConnect) error {
	return dxVirtualInterfaceWaitUntilAvailable(
		d,
		conn,
		[]string{
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateDown,
			directconnect.VirtualInterfaceStateVerifying,
		})
}
