package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxHostedPublicVirtualInterfaceAccepter() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxHostedPublicVirtualInterfaceAccepterCreate,
		Read:   resourceAwsDxHostedPublicVirtualInterfaceAccepterRead,
		Update: resourceAwsDxHostedPublicVirtualInterfaceAccepterUpdate,
		Delete: resourceAwsDxHostedPublicVirtualInterfaceAccepterDelete,
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
			"tags": tagsSchema(),
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},
	}
}

func resourceAwsDxHostedPublicVirtualInterfaceAccepterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	vifId := d.Get("virtual_interface_id").(string)
	req := &directconnect.ConfirmPublicVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(vifId),
	}

	log.Printf("[DEBUG] Accepting Direct Connect hosted public virtual interface: %#v", req)
	_, err := conn.ConfirmPublicVirtualInterface(req)
	if err != nil {
		return fmt.Errorf("Error accepting Direct Connect hosted public virtual interface: %s", err.Error())
	}

	d.SetId(vifId)

	if err := dxHostedPublicVirtualInterfaceAccepterWaitUntilAvailable(d, conn); err != nil {
		return err
	}

	return resourceAwsDxHostedPublicVirtualInterfaceAccepterUpdate(d, meta)
}

func resourceAwsDxHostedPublicVirtualInterfaceAccepterRead(d *schema.ResourceData, meta interface{}) error {
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
	if err := getTagsDX(conn, d, d.Get("arn").(string)); err != nil {
		return err
	}

	return nil
}

func resourceAwsDxHostedPublicVirtualInterfaceAccepterUpdate(d *schema.ResourceData, meta interface{}) error {
	if err := dxVirtualInterfaceUpdate(d, meta); err != nil {
		return err
	}

	return resourceAwsDxHostedPublicVirtualInterfaceAccepterRead(d, meta)
}

func resourceAwsDxHostedPublicVirtualInterfaceAccepterDelete(d *schema.ResourceData, meta interface{}) error {
	return dxVirtualInterfaceDelete(d, meta)
}

func dxHostedPublicVirtualInterfaceAccepterWaitUntilAvailable(d *schema.ResourceData, conn *directconnect.DirectConnect) error {
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
			directconnect.VirtualInterfaceStateVerifying,
		})
}
