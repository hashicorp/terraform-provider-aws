package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
			"destination_cidr_block": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateFunc:     verify.ValidCIDRNetworkAddress,
				DiffSuppressFunc: suppressEqualCIDRBlockDiffs,
			},
			"blackhole": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
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

	input := &ec2.CreateTransitGatewayRouteInput{
		DestinationCidrBlock:       aws.String(destination),
		Blackhole:                  aws.Bool(d.Get("blackhole").(bool)),
		TransitGatewayAttachmentId: aws.String(d.Get("transit_gateway_attachment_id").(string)),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Creating EC2 Transit Gateway Route: %s", input)
	_, err := conn.CreateTransitGatewayRoute(input)
	if err != nil {
		return fmt.Errorf("error creating EC2 Transit Gateway Route: %s", err)
	}

	d.SetId(fmt.Sprintf("%s_%s", transitGatewayRouteTableID, destination))

	return resourceTransitGatewayRouteRead(d, meta)
}

func resourceTransitGatewayRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	transitGatewayRouteTableID, destination, err := decodeTransitGatewayRouteID(d.Id())
	if err != nil {
		return err
	}

	// Handle EC2 eventual consistency
	var transitGatewayRoute *ec2.TransitGatewayRoute
	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		transitGatewayRoute, err = ec2DescribeTransitGatewayRoute(conn, transitGatewayRouteTableID, destination)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && transitGatewayRoute == nil {
			return resource.RetryableError(&resource.NotFoundError{})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		transitGatewayRoute, err = ec2DescribeTransitGatewayRoute(conn, transitGatewayRouteTableID, destination)
	}

	if tfawserr.ErrMessageContains(err, "InvalidRouteTableID.NotFound", "") {
		log.Printf("[WARN] EC2 Transit Gateway Route Table (%s) not found, removing from state", transitGatewayRouteTableID)
		d.SetId("")
		return nil
	}

	if tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Transit Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Transit Gateway Route: %s", err)
	}

	if transitGatewayRoute == nil {
		log.Printf("[WARN] EC2 Transit Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	state := aws.StringValue(transitGatewayRoute.State)
	if state == ec2.TransitGatewayRouteStateDeleted || state == ec2.TransitGatewayRouteStateDeleting {
		log.Printf("[WARN] EC2 Transit Gateway Route (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

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

	transitGatewayRouteTableID, destination, err := decodeTransitGatewayRouteID(d.Id())
	if err != nil {
		return err
	}

	input := &ec2.DeleteTransitGatewayRouteInput{
		DestinationCidrBlock:       aws.String(destination),
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Deleting EC2 Transit Gateway Route (%s): %s", d.Id(), input)
	_, err = conn.DeleteTransitGatewayRoute(input)

	if tfawserr.ErrMessageContains(err, "InvalidRoute.NotFound", "") || tfawserr.ErrMessageContains(err, "InvalidRouteTableID.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Transit Gateway Route: %s", err)
	}

	return nil
}
