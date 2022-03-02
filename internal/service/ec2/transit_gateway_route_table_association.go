package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceTransitGatewayRouteTableAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayRouteTableAssociationCreate,
		Read:   resourceTransitGatewayRouteTableAssociationRead,
		Delete: resourceTransitGatewayRouteTableAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"resource_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"transit_gateway_route_table_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func resourceTransitGatewayRouteTableAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayAttachmentID := d.Get("transit_gateway_attachment_id").(string)
	transitGatewayRouteTableID := d.Get("transit_gateway_route_table_id").(string)

	input := &ec2.AssociateTransitGatewayRouteTableInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err := conn.AssociateTransitGatewayRouteTable(input)
	if err != nil {
		return fmt.Errorf("error associating EC2 Transit Gateway Route Table (%s) association (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", transitGatewayRouteTableID, transitGatewayAttachmentID))

	if err := waitForTransitGatewayRouteTableAssociationCreation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) association (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	return resourceTransitGatewayRouteTableAssociationRead(d, meta)
}

func resourceTransitGatewayRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := DecodeTransitGatewayRouteTableAssociationID(d.Id())
	if err != nil {
		return err
	}

	transitGatewayAssociation, err := DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

	if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) not found, removing from state", transitGatewayRouteTableID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route Table (%s) Association (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	if transitGatewayAssociation == nil {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) Association (%s) not found, removing from state", transitGatewayRouteTableID, transitGatewayAttachmentID)
		d.SetId("")
		return nil
	}

	if aws.StringValue(transitGatewayAssociation.State) == ec2.TransitGatewayAssociationStateDisassociating {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) Association (%s) in deleted state (%s), removing from state", transitGatewayRouteTableID, transitGatewayAttachmentID, aws.StringValue(transitGatewayAssociation.State))
		d.SetId("")
		return nil
	}

	d.Set("resource_id", transitGatewayAssociation.ResourceId)
	d.Set("resource_type", transitGatewayAssociation.ResourceType)
	d.Set("transit_gateway_attachment_id", transitGatewayAssociation.TransitGatewayAttachmentId)
	d.Set("transit_gateway_route_table_id", transitGatewayRouteTableID)

	return nil
}

func resourceTransitGatewayRouteTableAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := DecodeTransitGatewayRouteTableAssociationID(d.Id())
	if err != nil {
		return err
	}

	input := &ec2.DisassociateTransitGatewayRouteTableInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Disassociating EC2 Transit Gateway Route Table (%s) Association (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, input)
	_, err = conn.DisassociateTransitGatewayRouteTable(input)

	if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating EC2 Transit Gateway Route Table (%s) Association (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	if err := waitForTransitGatewayRouteTableAssociationDeletion(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) Association (%s) disassociation: %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	return nil
}
