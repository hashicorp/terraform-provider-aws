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

func ResourceTransitGatewayRouteTablePropagation() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayRouteTablePropagationCreate,
		Read:   resourceTransitGatewayRouteTablePropagationRead,
		Delete: resourceTransitGatewayRouteTablePropagationDelete,
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

func resourceTransitGatewayRouteTablePropagationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayAttachmentID := d.Get("transit_gateway_attachment_id").(string)
	transitGatewayRouteTableID := d.Get("transit_gateway_route_table_id").(string)

	input := &ec2.EnableTransitGatewayRouteTablePropagationInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err := conn.EnableTransitGatewayRouteTablePropagation(input)
	if err != nil {
		return fmt.Errorf("error enabling EC2 Transit Gateway Route Table (%s) propagation (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	d.SetId(fmt.Sprintf("%s_%s", transitGatewayRouteTableID, transitGatewayAttachmentID))

	if _, err := WaitTransitGatewayRouteTablePropagationStateEnabled(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) propagation (%s) to enable: %w", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	return resourceTransitGatewayRouteTablePropagationRead(d, meta)
}

func resourceTransitGatewayRouteTablePropagationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := DecodeTransitGatewayRouteTablePropagationID(d.Id())
	if err != nil {
		return err
	}

	transitGatewayPropagation, err := FindTransitGatewayRouteTablePropagation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) not found, removing from state", transitGatewayRouteTableID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route Table (%s) Propagation (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	if transitGatewayPropagation == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading EC2 Transit Gateway Route Table (%s) Propagation (%s): not found after creation", transitGatewayRouteTableID, transitGatewayAttachmentID)
		}

		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) Propagation (%s) not found, removing from state", transitGatewayRouteTableID, transitGatewayAttachmentID)
		d.SetId("")
		return nil
	}

	d.Set("resource_id", transitGatewayPropagation.ResourceId)
	d.Set("resource_type", transitGatewayPropagation.ResourceType)
	d.Set("transit_gateway_attachment_id", transitGatewayPropagation.TransitGatewayAttachmentId)
	d.Set("transit_gateway_route_table_id", transitGatewayRouteTableID)

	return nil
}

func resourceTransitGatewayRouteTablePropagationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := DecodeTransitGatewayRouteTablePropagationID(d.Id())
	if err != nil {
		return err
	}

	input := &ec2.DisableTransitGatewayRouteTablePropagationInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Disabling EC2 Transit Gateway Route Table (%s) Propagation (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, input)
	_, err = conn.DisableTransitGatewayRouteTablePropagation(input)

	if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disabling EC2 Transit Gateway Route Table (%s) Propagation (%s): %s", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	if _, err := WaitTransitGatewayRouteTablePropagationStateDisabled(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Route Table (%s) propagation (%s) to disable: %w", transitGatewayRouteTableID, transitGatewayAttachmentID, err)
	}

	return nil
}
