package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxPrivateVirtualInterface() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxPrivateVirtualInterfaceCreate,
		Read:   resourceAwsDxPrivateVirtualInterfaceRead,
		Update: resourceAwsDxPrivateVirtualInterfaceUpdate,
		Delete: resourceAwsDxPrivateVirtualInterfaceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: mergeSchemas(
			dxVirtualInterfaceSchemaWithTags,
			map[string]*schema.Schema{
				"vpn_gateway_id": {
					Type:          schema.TypeString,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"dx_gateway_id"},
				},
				"dx_gateway_id": {
					Type:          schema.TypeString,
					Optional:      true,
					ForceNew:      true,
					ConflictsWith: []string{"vpn_gateway_id"},
				},
			},
		),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxPrivateVirtualInterfaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vgwIdRaw, vgwOk := d.GetOk("vpn_gateway_id")
	dxgwIdRaw, dxgwOk := d.GetOk("dx_gateway_id")
	if vgwOk == dxgwOk {
		return fmt.Errorf(
			"One of ['vpn_gateway_id', 'dx_gateway_id'] must be set to create a Direct Connect private virtual interface")
	}

	req := &directconnect.CreatePrivateVirtualInterfaceInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		NewPrivateVirtualInterface: &directconnect.NewPrivateVirtualInterface{
			VirtualInterfaceName: aws.String(d.Get("name").(string)),
			Vlan:                 aws.Int64(int64(d.Get("vlan").(int))),
			Asn:                  aws.Int64(int64(d.Get("bgp_asn").(int))),
			AddressFamily:        aws.String(d.Get("address_family").(string)),
		},
	}
	if vgwOk {
		req.NewPrivateVirtualInterface.VirtualGatewayId = aws.String(vgwIdRaw.(string))
	}
	if dxgwOk {
		req.NewPrivateVirtualInterface.DirectConnectGatewayId = aws.String(dxgwIdRaw.(string))
	}
	if v, ok := d.GetOk("bgp_auth_key"); ok {
		req.NewPrivateVirtualInterface.AuthKey = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customer_address"); ok {
		req.NewPrivateVirtualInterface.CustomerAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("amazon_address"); ok {
		req.NewPrivateVirtualInterface.AmazonAddress = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Direct Connect private virtual interface: %#v", req)
	resp, err := conn.CreatePrivateVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("Error creating Direct Connect private virtual interface: %s", err.Error())
	}

	d.SetId(aws.StringValue(resp.VirtualInterfaceId))

	if err := dxPrivateVirtualInterfaceWaitUntilAvailable(d, conn); err != nil {
		return err
	}

	return resourceAwsDxPrivateVirtualInterfaceUpdate(d, meta)
}

func resourceAwsDxPrivateVirtualInterfaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d, meta)
	if err != nil {
		return err
	}
	if vif == nil {
		return nil
	}

	if err := dxPrivateVirtualInterfaceAttributes(d, meta, vif); err != nil {
		return err
	}
	d.Set("vpn_gateway_id", vif.VirtualGatewayId)
	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	if err := getTagsDX(conn, d, d.Get("arn").(string)); err != nil {
		return err
	}

	return nil
}

func resourceAwsDxPrivateVirtualInterfaceUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := dxVirtualInterfaceUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsDxPrivateVirtualInterfaceRead(d, meta)
}

func resourceAwsDxPrivateVirtualInterfaceDelete(d *schema.ResourceData, meta interface{}) error {
	return dxVirtualInterfaceDelete(d, meta)
}

func dxPrivateVirtualInterfaceWaitUntilAvailable(d *schema.ResourceData, conn *directconnect.DirectConnect) error {
	return dxVirtualInterfaceWaitUntilAvailable(
		d,
		conn,
		[]string{
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateDown,
		})
}
