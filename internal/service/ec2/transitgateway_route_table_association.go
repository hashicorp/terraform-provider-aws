package ec2

import (
	"fmt"
	"log"
	"strings"

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
	id := TransitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)
	input := &ec2.AssociateTransitGatewayRouteTableInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err := conn.AssociateTransitGatewayRouteTable(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Transit Gateway Route Table Association (%s): %w", id, err)
	}

	d.SetId(id)

	if err := waitForTransitGatewayRouteTableAssociationCreation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Association (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayRouteTableAssociationRead(d, meta)
}

func resourceTransitGatewayRouteTableAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := TransitGatewayRouteTableAssociationParseResourceID(d.Id())

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

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := TransitGatewayRouteTableAssociationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route Table Association: %s", d.Id())
	_, err = conn.DisassociateTransitGatewayRouteTable(&ec2.DisassociateTransitGatewayRouteTableInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Route Table Association (%s): %w", d.Id(), err)
	}

	if err := waitForTransitGatewayRouteTableAssociationDeletion(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Association (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const transitGatewayRouteTableAssociationIDSeparator = "_"

func TransitGatewayRouteTableAssociationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID string) string {
	parts := []string{transitGatewayRouteTableID, transitGatewayAttachmentID}
	id := strings.Join(parts, transitGatewayRouteTableAssociationIDSeparator)

	return id
}

func TransitGatewayRouteTableAssociationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayRouteTableAssociationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sTRANSIT-GATEWAY-ATTACHMENT-ID", id, transitGatewayRouteTableAssociationIDSeparator)
}
