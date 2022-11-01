package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	localGatewayRouteEventualConsistencyTimeout = 1 * time.Minute
)

func ResourceLocalGatewayRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceLocalGatewayRouteCreate,
		Read:   resourceLocalGatewayRouteRead,
		Delete: resourceLocalGatewayRouteDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidCIDRNetworkAddress,
			},
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"local_gateway_virtual_interface_group_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLocalGatewayRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	destination := d.Get("destination_cidr_block").(string)
	localGatewayRouteTableID := d.Get("local_gateway_route_table_id").(string)

	input := &ec2.CreateLocalGatewayRouteInput{
		DestinationCidrBlock:                aws.String(destination),
		LocalGatewayRouteTableId:            aws.String(localGatewayRouteTableID),
		LocalGatewayVirtualInterfaceGroupId: aws.String(d.Get("local_gateway_virtual_interface_group_id").(string)),
	}

	_, err := conn.CreateLocalGatewayRoute(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Local Gateway Route: %s", err)
	}

	d.SetId(fmt.Sprintf("%s_%s", localGatewayRouteTableID, destination))

	return resourceLocalGatewayRouteRead(d, meta)
}

func resourceLocalGatewayRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	localGatewayRouteTableID, destination, err := DecodeLocalGatewayRouteID(d.Id())
	if err != nil {
		return err
	}

	var localGatewayRoute *ec2.LocalGatewayRoute
	err = resource.Retry(localGatewayRouteEventualConsistencyTimeout, func() *resource.RetryError {
		var err error
		localGatewayRoute, err = GetLocalGatewayRoute(conn, localGatewayRouteTableID, destination)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && localGatewayRoute == nil {
			return resource.RetryableError(&resource.NotFoundError{})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		localGatewayRoute, err = GetLocalGatewayRoute(conn, localGatewayRouteTableID, destination)
	}

	if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		log.Printf("[WARN] EC2 Local Gateway Route Table (%s) not found, removing from state", localGatewayRouteTableID)
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Local Gateway Route: %s", err)
	}

	if localGatewayRoute == nil {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	state := aws.StringValue(localGatewayRoute.State)
	if state == ec2.LocalGatewayRouteStateDeleted || state == ec2.LocalGatewayRouteStateDeleting {
		log.Printf("[WARN] EC2 Local Gateway Route (%s) deleted, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("destination_cidr_block", localGatewayRoute.DestinationCidrBlock)
	d.Set("local_gateway_virtual_interface_group_id", localGatewayRoute.LocalGatewayVirtualInterfaceGroupId)
	d.Set("local_gateway_route_table_id", localGatewayRoute.LocalGatewayRouteTableId)

	return nil
}

func resourceLocalGatewayRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	localGatewayRouteTableID, destination, err := DecodeLocalGatewayRouteID(d.Id())
	if err != nil {
		return err
	}

	input := &ec2.DeleteLocalGatewayRouteInput{
		DestinationCidrBlock:     aws.String(destination),
		LocalGatewayRouteTableId: aws.String(localGatewayRouteTableID),
	}

	log.Printf("[DEBUG] Deleting EC2 Local Gateway Route (%s): %s", d.Id(), input)
	_, err = conn.DeleteLocalGatewayRoute(input)

	if tfawserr.ErrCodeEquals(err, "InvalidRoute.NotFound") || tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Local Gateway Route: %s", err)
	}

	return nil
}

func DecodeLocalGatewayRouteID(id string) (string, string, error) {
	parts := strings.Split(id, "_")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected tgw-rtb-ID_DESTINATION", id)
	}

	return parts[0], parts[1], nil
}

func GetLocalGatewayRoute(conn *ec2.EC2, localGatewayRouteTableID, destination string) (*ec2.LocalGatewayRoute, error) {
	input := &ec2.SearchLocalGatewayRoutesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("type"),
				Values: aws.StringSlice([]string{"static"}),
			},
		},
		LocalGatewayRouteTableId: aws.String(localGatewayRouteTableID),
	}

	output, err := conn.SearchLocalGatewayRoutes(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Routes) == 0 {
		return nil, nil
	}

	for _, route := range output.Routes {
		if route == nil {
			continue
		}

		if aws.StringValue(route.DestinationCidrBlock) == destination {
			return route, nil
		}
	}

	return nil, nil
}
