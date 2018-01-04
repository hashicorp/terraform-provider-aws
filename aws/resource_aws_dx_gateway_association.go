package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsDxGatewayAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsDxGatewayAssociationCreate,
		Read:   resourceAwsDxGatewayAssociationRead,
		Delete: resourceAwsDxGatewayAssociationDelete,

		Schema: map[string]*schema.Schema{
			"dx_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"virtual_gateway_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsDxGatewayAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxGatewayId := d.Get("dx_gateway_id").(string)
	vgwId := d.Get("virtual_gateway_id").(string)

	input := &directconnect.CreateDirectConnectGatewayAssociationInput{
		DirectConnectGatewayId: aws.String(dxGatewayId),
		VirtualGatewayId:       aws.String(vgwId),
	}
	_, err := conn.CreateDirectConnectGatewayAssociation(input)
	if err != nil {
		return err
	}

	d.SetId(dxGatewayIdVgwIdHash(dxGatewayId, vgwId))
	return nil
}

func resourceAwsDxGatewayAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxGatewayId := d.Get("dx_gateway_id").(string)
	vgwId := d.Get("virtual_gateway_id").(string)

	input := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
		DirectConnectGatewayId: aws.String(dxGatewayId),
		VirtualGatewayId:       aws.String(vgwId),
	}

	resp, err := conn.DescribeDirectConnectGatewayAssociations(input)
	if err != nil {
		return err
	}
	if len(resp.DirectConnectGatewayAssociations) < 1 {
		d.SetId("")
		return nil
	}
	if len(resp.DirectConnectGatewayAssociations) != 1 {
		return fmt.Errorf("Found %d Direct Connect Gateway associations for %s, expected 1", len(resp.DirectConnectGatewayAssociations), d.Id())
	}
	if *resp.DirectConnectGatewayAssociations[0].VirtualGatewayId != d.Get("virtual_gateway_id").(string) {
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAwsDxGatewayAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).dxconn

	dxGatewayId := d.Get("dx_gateway_id").(string)
	vgwId := d.Get("virtual_gateway_id").(string)

	input := &directconnect.DeleteDirectConnectGatewayAssociationInput{
		DirectConnectGatewayId: aws.String(dxGatewayId),
		VirtualGatewayId:       aws.String(vgwId),
	}

	_, err := conn.DeleteDirectConnectGatewayAssociation(input)
	if err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func dxGatewayIdVgwIdHash(gatewayId, vgwId string) string {
	return fmt.Sprintf("ga-%s%d", gatewayId, hashcode.String(vgwId))
}
