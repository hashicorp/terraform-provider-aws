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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTransitGatewayRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceTransitGatewayRouteCreate,
		Read:   resourceTransitGatewayRouteRead,
		Delete: resourceTransitGatewayRouteDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"blackhole": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"destination_cidr_block": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     verify.ValidCIDRNetworkAddress,
				DiffSuppressFunc: suppressEqualCIDRBlockDiffs,
			},
			"transit_gateway_attachment_id": {
				Type:         schema.TypeString,
				Optional:     true,
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

func resourceTransitGatewayRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	destination := d.Get("destination_cidr_block").(string)
	transitGatewayRouteTableID := d.Get("transit_gateway_route_table_id").(string)
	id := TransitGatewayRouteCreateResourceID(transitGatewayRouteTableID, destination)
	input := &ec2.CreateTransitGatewayRouteInput{
		Blackhole:                  aws.Bool(d.Get("blackhole").(bool)),
		DestinationCidrBlock:       aws.String(destination),
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Route: %s", input)
	_, err := conn.CreateTransitGatewayRoute(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Route (%s): %w", id, err)
	}

	d.SetId(id)

	if _, err := WaitTransitGatewayRouteCreated(conn, transitGatewayRouteTableID, destination); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Route (%s) create: %w", d.Id(), err)
	}

	return resourceTransitGatewayRouteRead(d, meta)
}

func resourceTransitGatewayRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, destination, err := TransitGatewayRouteParseResourceID(d.Id())

	if err != nil {
		return err
	}

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(propagationTimeout, func() (interface{}, error) {
		return FindTransitGatewayRoute(conn, transitGatewayRouteTableID, destination)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Route %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route (%s): %w", d.Id(), err)
	}

	transitGatewayRoute := outputRaw.(*ec2.TransitGatewayRoute)

	d.Set("destination_cidr_block", transitGatewayRoute.DestinationCidrBlock)
	d.Set("transit_gateway_attachment_id", "")
	if len(transitGatewayRoute.TransitGatewayAttachments) > 0 && transitGatewayRoute.TransitGatewayAttachments[0] != nil {
		d.Set("transit_gateway_attachment_id", transitGatewayRoute.TransitGatewayAttachments[0].TransitGatewayAttachmentId)
		d.Set("blackhole", false)
	} else {
		d.Set("blackhole", true)
	}
	d.Set("transit_gateway_route_table_id", transitGatewayRouteTableID)

	return nil
}

func resourceTransitGatewayRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, destination, err := TransitGatewayRouteParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route: %s", d.Id())
	_, err = conn.DeleteTransitGatewayRoute(&ec2.DeleteTransitGatewayRouteInput{
		DestinationCidrBlock:       aws.String(destination),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteNotFound, ErrCodeInvalidRouteTableIDNotFound) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Route (%s): %w", d.Id(), err)
	}

	if _, err := WaitTransitGatewayRouteDeleted(conn, transitGatewayRouteTableID, destination); err != nil {
		return fmt.Errorf("error waiting for EC2 Transit Gateway Route (%s) delete: %w", d.Id(), err)
	}

	return nil
}

const transitGatewayRouteIDSeparator = "_"

func TransitGatewayRouteCreateResourceID(transitGatewayRouteTableID, destination string) string {
	parts := []string{transitGatewayRouteTableID, destination}
	id := strings.Join(parts, transitGatewayRouteIDSeparator)

	return id
}

func TransitGatewayRouteParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, transitGatewayRouteIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected TRANSIT-GATEWAY-ROUTE-TABLE-ID%[2]sDESTINATION", id, transitGatewayRouteIDSeparator)
}
