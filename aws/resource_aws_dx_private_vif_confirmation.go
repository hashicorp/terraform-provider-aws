package aws

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/directconnect"
)

func resourceAwsDxPrivateVifConfirmation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxPrivateVifConfirmationCreate,
		Read:   resourceAwsDxPrivateVifConfirmationRead,
		Update: resourceAwsDxPrivateVifConfirmationUpdate,
		Delete: resourceAwsDxPrivateVifConfirmationDelete,

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

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsDxPrivateVifConfirmationCreate(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	id := d.Get("virtual_interface_id").(string)

	vif, err := resourceAwsDxPrivateVifGet(dxconn, id)

	if err != nil {
		return err
	}

	if vif == nil {
		return fmt.Errorf("Private virtual interface %q not found", id)
	}

	state := *vif.VirtualInterfaceState

	if state == "confirming" {
		_, err = dxconn.ConfirmPrivateVirtualInterface(&directconnect.ConfirmPrivateVirtualInterfaceInput{
			VirtualInterfaceId: aws.String(id),
			VirtualGatewayId:   aws.String(d.Get("virtual_gateway_id").(string)),
		})

		if err != nil {
			return err
		}

		_, err = waitForAwsDxPrivateVif(dxconn, id, []string{"available", "down"})

		if err != nil {
			return err
		}
	} else if state != "available" && state != "down" {
		return fmt.Errorf("Invalid virtual interface state %q", *vif.VirtualInterfaceState)
	}

	d.SetId(id)

	return resourceAwsDxPrivateVifConfirmationUpdate(d, meta)
}

func resourceAwsDxPrivateVifConfirmationUpdate(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()

	if err := setTagsDX(dxconn, d, arn); err != nil {
		return err
	}

	return nil
}

func resourceAwsDxPrivateVifConfirmationRead(d *schema.ResourceData, meta interface{}) error {
	dxconn := meta.(*AWSClient).dxconn

	vif, err := resourceAwsDxPrivateVifGet(dxconn, d.Id())

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

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Region:    meta.(*AWSClient).region,
		Service:   "directconnect",
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("dxvif/%s", d.Id()),
	}.String()

	if err := getTagsDX(dxconn, d, arn); err != nil {
		return err
	}

	return nil
}

func resourceAwsDxPrivateVifConfirmationDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Will not delete private VIF. Terraform will remove this resource from the state file, however resources may remain.")
	d.SetId("")
	return nil
}
