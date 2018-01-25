package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxHostedPrivateVirtualInterfaceAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxHostedPrivateVirtualInterfaceAccepterCreate,
		Read:   resourceAwsDxHostedPrivateVirtualInterfaceAccepterRead,
		Update: resourceAwsDxHostedPrivateVirtualInterfaceAccepterUpdate,
		Delete: resourceAwsDxHostedPrivateVirtualInterfaceAccepterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"virtual_interface_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
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
			"tags": tagsSchema(),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vgwIdRaw, vgwOk := d.GetOk("vpn_gateway_id")
	dxgwIdRaw, dxgwOk := d.GetOk("dx_gateway_id")
	if vgwOk == dxgwOk {
		return fmt.Errorf(
			"One of ['vpn_gateway_id', 'dx_gateway_id'] must be set to create a Direct Connect private virtual interface accepter")
	}

	vifId := d.Get("virtual_interface_id").(string)
	req := &directconnect.ConfirmPrivateVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(vifId),
	}
	if vgwOk {
		req.VirtualGatewayId = aws.String(vgwIdRaw.(string))
	}
	if dxgwOk {
		req.DirectConnectGatewayId = aws.String(dxgwIdRaw.(string))
	}

	log.Printf("[DEBUG] Accepting Direct Connect hosted private virtual interface: %#v", req)
	_, err := conn.ConfirmPrivateVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("Error accepting Direct Connect hosted private virtual interface: %s", err.Error())
	}

	d.SetId(vifId)

	if err := dxHostedPrivateVirtualInterfaceAccepterWaitUntilAvailable(d, conn); err != nil {
		return err
	}

	return resourceAwsDxHostedPrivateVirtualInterfaceAccepterUpdate(d, meta)
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vif, err := dxVirtualInterfaceRead(d, meta)
	if err != nil {
		return err
	}
	if vif == nil {
		return nil
	}

	if err := dxVirtualInterfaceArnAttribute(d, meta); err != nil {
		return err
	}
	d.Set("virtual_interface_id", vif.VirtualInterfaceId)
	d.Set("vpn_gateway_id", vif.VirtualGatewayId)
	d.Set("dx_gateway_id", vif.DirectConnectGatewayId)
	if err := getTagsDX(conn, d, d.Get("arn").(string)); err != nil {
		return err
	}

	return nil
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := dxVirtualInterfaceUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsDxHostedPrivateVirtualInterfaceAccepterRead(d, meta)
}

func resourceAwsDxHostedPrivateVirtualInterfaceAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	return dxVirtualInterfaceDelete(d, meta)
}

func dxHostedPrivateVirtualInterfaceAccepterWaitUntilAvailable(d *schema.ResourceData, conn *directconnect.DirectConnect) error {
	return dxVirtualInterfaceWaitUntilAvailable(
		d,
		conn,
		[]string{
			directconnect.VirtualInterfaceStateConfirming,
			directconnect.VirtualInterfaceStatePending,
		},
		[]string{
			directconnect.VirtualInterfaceStateAvailable,
			directconnect.VirtualInterfaceStateDown,
		})
}
