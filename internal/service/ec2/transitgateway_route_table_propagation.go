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
	id := TransitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID)
	input := &ec2.EnableTransitGatewayRouteTablePropagationInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	_, err := conn.EnableTransitGatewayRouteTablePropagation(input)

	if err != nil {
		return fmt.Errorf("creating EC2 Transit Gateway Route Table Propagation (%s): %w", id, err)
	}

	d.SetId(id)

	if _, err := WaitTransitGatewayRouteTablePropagationStateEnabled(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Propagation (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayRouteTablePropagationRead(d, meta)
}

func resourceTransitGatewayRouteTablePropagationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := TransitGatewayRouteTablePropagationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	transitGatewayPropagation, err := FindTransitGatewayRouteTablePropagation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
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

	transitGatewayRouteTableID, transitGatewayAttachmentID, err := TransitGatewayRouteTablePropagationParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route Table Propagation: %s", d.Id())
	_, err = conn.DisableTransitGatewayRouteTablePropagation(&ec2.DisableTransitGatewayRouteTablePropagationInput{
		TransitGatewayAttachmentId: aws.String(transitGatewayAttachmentID),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EC2 Transit Gateway Route Table Propagation (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayRouteTablePropagationStateDisabled(conn, transitGatewayRouteTableID, transitGatewayAttachmentID); err != nil {
		return fmt.Errorf("waiting for EC2 Transit Gateway Route Table Propagation (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const transitGatewayRouteTablePropagationIDSeparator = "_"

func TransitGatewayRouteTablePropagationCreateResourceID(transitGatewayRouteTableID, transitGatewayAttachmentID string) string {
	parts := []string{transitGatewayRouteTableID, transitGatewayAttachmentID}
	id := strings.Join(parts, transitGatewayRouteTablePropagationIDSeparator)

	return id
}

func TransitGatewayRouteTablePropagationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayRouteTablePropagationIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sTRANSIT-GATEWAY-ATTACHMENT-ID", id, transitGatewayRouteTablePropagationIDSeparator)
}
