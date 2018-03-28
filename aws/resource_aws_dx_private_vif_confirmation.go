package aws

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
)

func resourceAwsDxPrivateVirtualInterfaceConfirmation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxPrivateVirtualInterfaceConfirmationCreate,
		Read:   resourceAwsDxPrivateVirtualInterfaceConfirmationRead,
		Delete: resourceAwsDxPrivateVirtualInterfaceConfirmationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"virtual_gateway_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"virtual_interface_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDxPrivateVirtualInterfaceConfirmationCreate(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	id := d.Get("virtual_interface_id").(string)

	vif, err := resourceAwsDxPrivateVirtualInterfaceGet(dxconn, id)

	if err != nil {
		return err
	}

	if vif == nil {
		return fmt.Errorf("Private virtual interface %q not found", id)
	}

	state := *vif.VirtualInterfaceState

	if state == "available" || state == "down" {
		return nil
	} else if state != "confirming" {
		return fmt.Errorf("Invalid virtual interface state %q", *vif.VirtualInterfaceState)
	}

	_, err = dxconn.ConfirmPrivateVirtualInterface(&directconnect.ConfirmPrivateVirtualInterfaceInput{
		VirtualInterfaceId: aws.String(id),
		VirtualGatewayId:   aws.String(d.Get("virtual_gateway_id").(string)),
	})

	if err != nil {
		return err
	}

	_, err = waitForAwsDxPrivateVirtualInterface(dxconn, id, []string{"available", "down"})

	if err != nil {
		return err
	}

	d.SetId(id)

	return nil
}

func resourceAwsDxPrivateVirtualInterfaceConfirmationRead(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	vif, err := resourceAwsDxPrivateVirtualInterfaceGet(dxconn, d.Id())

	if err != nil {
		return err
	}

	if vif == nil {
		d.SetId("")
		return nil
	}

	d.Set("virtual_interface_id", *vif.VirtualInterfaceId)

	if vif.VirtualGatewayId != nil {
		d.Set("virtual_gateway_id", *vif.VirtualGatewayId)
	}

	return nil
}

func resourceAwsDxPrivateVirtualInterfaceConfirmationDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete private VIF. Terraform will remove this resource from the state file, however resources may remain.")
	d.SetId("")
	return nil
}
