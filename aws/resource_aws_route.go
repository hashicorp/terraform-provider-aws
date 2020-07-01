package aws

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

// How long to sleep if a limit-exceeded event happens
var routeTargetValidationError = errors.New("Error: more than 1 target specified. Only 1 of gateway_id, " +
	"egress_only_gateway_id, nat_gateway_id, instance_id, network_interface_id or " +
	"vpc_peering_connection_id is allowed.")

// AWS Route resource Schema declaration
func resourceAwsRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRouteCreate,
		Read:   resourceAwsRouteRead,
		Update: resourceAwsRouteUpdate,
		Delete: resourceAwsRouteDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "_")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected ROUTETABLEID_DESTINATION", d.Id())
				}
				routeTableID := idParts[0]
				destination := idParts[1]
				d.Set("route_table_id", routeTableID)
				if strings.Contains(destination, ":") {
					d.Set("destination_ipv6_cidr_block", destination)
				} else {
					d.Set("destination_cidr_block", destination)
				}
				d.SetId(fmt.Sprintf("r-%s%d", routeTableID, hashcode.String(destination)))
				return []*schema.ResourceData{d}, nil
			},
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"destination_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validateIpv4CIDRNetworkAddress,
				),
			},

			"destination_ipv6_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validateIpv6CIDRNetworkAddress,
				),
				DiffSuppressFunc: suppressEqualCIDRBlockDiffs,
			},

			"destination_prefix_list_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"egress_only_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"nat_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"instance_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"network_interface_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"transit_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	destinationAttr, targetAttr, err := routeDestinationAndTargetAttributes(d)
	if err != nil {
		return err
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.CreateRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var pDestination **string
	var routeReader func(*ec2.EC2, string, string) (*ec2.Route, error)
	switch destinationAttr {
	case "destination_cidr_block":
		pDestination = &input.DestinationCidrBlock
		routeReader = readIpv4Route
	case "destination_ipv6_cidr_block":
		pDestination = &input.DestinationIpv6CidrBlock
		routeReader = readIpv6Route
	default:
		return fmt.Errorf("unexpected destination attribute: `%s`", destinationAttr)
	}
	destination := d.Get(destinationAttr).(string)
	*pDestination = aws.String(destination)

	var pTarget **string
	switch targetAttr {
	case "egress_only_gateway_id":
		pTarget = &input.EgressOnlyInternetGatewayId
	case "gateway_id":
		pTarget = &input.GatewayId
	case "instance_id":
		pTarget = &input.InstanceId
	case "nat_gateway_id":
		pTarget = &input.NatGatewayId
	case "network_interface_id":
		pTarget = &input.NetworkInterfaceId
	case "transit_gateway_id":
		pTarget = &input.TransitGatewayId
	case "vpc_peering_connection_id":
		pTarget = &input.VpcPeeringConnectionId
	default:
		return fmt.Errorf("unexpected target attribute: `%s`", targetAttr)
	}
	*pTarget = aws.String(d.Get(targetAttr).(string))

	log.Printf("[DEBUG] Creating Route: %s", input)

	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err = conn.CreateRoute(input)

		if isAWSErr(err, "InvalidParameterException", "") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, "InvalidTransitGatewayID.NotFound", "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateRoute(input)
	}

	if err != nil {
		return fmt.Errorf("error creating route: %s", err)
	}

	var route *ec2.Route
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		route, err = routeReader(conn, routeTableID, destination)

		if err != nil {
			return resource.RetryableError(err)
		}

		if route == nil {
			return resource.RetryableError(fmt.Errorf("route not found"))
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		route, err = routeReader(conn, routeTableID, destination)
	}

	if err != nil {
		return fmt.Errorf("error reading route for Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	if route == nil {
		return fmt.Errorf("route in Route Table (%s) with destination (%s) not found", routeTableID, destination)
	}

	d.SetId(routeCreateID(routeTableID, destination))

	return resourceAwsRouteRead(d, meta)
}

func resourceAwsRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	destinationAttr := routeDestinationAttr(d)

	routeTableID := d.Get("route_table_id").(string)

	var routeReader func(*ec2.EC2, string, string) (*ec2.Route, error)
	switch destinationAttr {
	case "destination_cidr_block":
		routeReader = readIpv4Route
	case "destination_ipv6_cidr_block":
		routeReader = readIpv6Route
	default:
		return fmt.Errorf("unexpected destination attribute: `%s`", destinationAttr)
	}
	destination := d.Get(destinationAttr).(string)

	route, err := routeReader(conn, routeTableID, destination)

	if isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
		log.Printf("[WARN] Route Table (%s) not found, removing from state", routeTableID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading route for Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	if route == nil {
		log.Printf("[WARN] Route in Route Table (%s) with destination (%s) not found, removing from state", routeTableID, destination)
		d.SetId("")
		return nil
	}

	d.Set("destination_cidr_block", route.DestinationCidrBlock)
	d.Set("destination_ipv6_cidr_block", route.DestinationIpv6CidrBlock)
	d.Set("destination_prefix_list_id", route.DestinationPrefixListId)
	d.Set("gateway_id", route.GatewayId)
	d.Set("egress_only_gateway_id", route.EgressOnlyInternetGatewayId)
	d.Set("nat_gateway_id", route.NatGatewayId)
	d.Set("instance_id", route.InstanceId)
	d.Set("instance_owner_id", route.InstanceOwnerId)
	d.Set("network_interface_id", route.NetworkInterfaceId)
	d.Set("origin", route.Origin)
	d.Set("state", route.State)
	d.Set("transit_gateway_id", route.TransitGatewayId)
	d.Set("vpc_peering_connection_id", route.VpcPeeringConnectionId)

	return nil
}

func resourceAwsRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	var numTargets int
	var setTarget string

	allowedTargets := []string{
		"egress_only_gateway_id",
		"gateway_id",
		"nat_gateway_id",
		"network_interface_id",
		"instance_id",
		"transit_gateway_id",
		"vpc_peering_connection_id",
	}
	// Check if more than 1 target is specified
	for _, target := range allowedTargets {
		if len(d.Get(target).(string)) > 0 {
			numTargets++
			setTarget = target
		}
	}

	switch setTarget {
	//instance_id is a special case due to the fact that AWS will "discover" the network_interface_id
	//when it creates the route and return that data.  In the case of an update, we should ignore the
	//existing network_interface_id
	case "instance_id":
		if numTargets > 2 || (numTargets == 2 && len(d.Get("network_interface_id").(string)) == 0) {
			return routeTargetValidationError
		}
	default:
		if numTargets > 1 {
			return routeTargetValidationError
		}
	}

	var replaceOpts *ec2.ReplaceRouteInput
	// Formulate ReplaceRouteInput based on the target type
	switch setTarget {
	case "gateway_id":
		replaceOpts = &ec2.ReplaceRouteInput{
			RouteTableId:         aws.String(d.Get("route_table_id").(string)),
			DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
			GatewayId:            aws.String(d.Get("gateway_id").(string)),
		}
	case "egress_only_gateway_id":
		replaceOpts = &ec2.ReplaceRouteInput{
			RouteTableId:                aws.String(d.Get("route_table_id").(string)),
			DestinationIpv6CidrBlock:    aws.String(d.Get("destination_ipv6_cidr_block").(string)),
			EgressOnlyInternetGatewayId: aws.String(d.Get("egress_only_gateway_id").(string)),
		}
	case "nat_gateway_id":
		replaceOpts = &ec2.ReplaceRouteInput{
			RouteTableId:         aws.String(d.Get("route_table_id").(string)),
			DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
			NatGatewayId:         aws.String(d.Get("nat_gateway_id").(string)),
		}
	case "instance_id":
		replaceOpts = &ec2.ReplaceRouteInput{
			RouteTableId:         aws.String(d.Get("route_table_id").(string)),
			DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
			InstanceId:           aws.String(d.Get("instance_id").(string)),
		}
	case "network_interface_id":
		replaceOpts = &ec2.ReplaceRouteInput{
			RouteTableId:         aws.String(d.Get("route_table_id").(string)),
			DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
			NetworkInterfaceId:   aws.String(d.Get("network_interface_id").(string)),
		}
	case "transit_gateway_id":
		replaceOpts = &ec2.ReplaceRouteInput{
			RouteTableId:         aws.String(d.Get("route_table_id").(string)),
			DestinationCidrBlock: aws.String(d.Get("destination_cidr_block").(string)),
			TransitGatewayId:     aws.String(d.Get("transit_gateway_id").(string)),
		}
	case "vpc_peering_connection_id":
		replaceOpts = &ec2.ReplaceRouteInput{
			RouteTableId:           aws.String(d.Get("route_table_id").(string)),
			DestinationCidrBlock:   aws.String(d.Get("destination_cidr_block").(string)),
			VpcPeeringConnectionId: aws.String(d.Get("vpc_peering_connection_id").(string)),
		}
	default:
		return fmt.Errorf("An invalid target type specified: %s", setTarget)
	}
	log.Printf("[DEBUG] Route replace config: %s", replaceOpts)

	// Replace the route
	_, err := conn.ReplaceRoute(replaceOpts)
	return err
}

func resourceAwsRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	destinationAttr := routeDestinationAttr(d)

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.DeleteRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var pDestination **string
	switch destinationAttr {
	case "destination_cidr_block":
		pDestination = &input.DestinationCidrBlock
	case "destination_ipv6_cidr_block":
		pDestination = &input.DestinationIpv6CidrBlock
	default:
		return fmt.Errorf("unexpected destination attribute: `%s`", destinationAttr)
	}
	destination := d.Get(destinationAttr).(string)
	*pDestination = aws.String(destination)

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		log.Printf("[DEBUG] Deleting route (%s)", d.Id())
		_, err := conn.DeleteRoute(input)
		if err == nil {
			return nil
		}

		if isAWSErr(err, "InvalidRoute.NotFound", "") {
			return nil
		}

		if isAWSErr(err, "InvalidParameterException", "") {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})

	if isResourceTimeoutError(err) {
		log.Printf("[DEBUG] Deleting route (%s)", d.Id())
		_, err = conn.DeleteRoute(input)
	}

	if isAWSErr(err, "InvalidRoute.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting route for Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	return nil
}

// Helper: Create an ID for a route
func resourceAwsRouteID(d *schema.ResourceData, r *ec2.Route) string {

	if r.DestinationIpv6CidrBlock != nil && *r.DestinationIpv6CidrBlock != "" {
		return fmt.Sprintf("r-%s%d", d.Get("route_table_id").(string), hashcode.String(*r.DestinationIpv6CidrBlock))
	}

	return fmt.Sprintf("r-%s%d", d.Get("route_table_id").(string), hashcode.String(*r.DestinationCidrBlock))
}

// resourceAwsRouteFindRoute returns any route whose destination is the specified IPv4 or IPv6 CIDR block.
// Returns nil if the route table exists but no matching destination is found.
func resourceAwsRouteFindRoute(conn *ec2.EC2, rtbid string, cidr string, ipv6cidr string) (*ec2.Route, error) {
	routeTableID := rtbid

	findOpts := &ec2.DescribeRouteTablesInput{
		RouteTableIds: []*string{&routeTableID},
	}

	resp, err := conn.DescribeRouteTables(findOpts)
	if err != nil {
		return nil, err
	}

	if len(resp.RouteTables) < 1 || resp.RouteTables[0] == nil {
		return nil, nil
	}

	if cidr != "" {
		for _, route := range (*resp.RouteTables[0]).Routes {
			if route.DestinationCidrBlock != nil && *route.DestinationCidrBlock == cidr {
				return route, nil
			}
		}

		return nil, nil
	}

	if ipv6cidr != "" {
		for _, route := range (*resp.RouteTables[0]).Routes {
			if cidrBlocksEqual(aws.StringValue(route.DestinationIpv6CidrBlock), ipv6cidr) {
				return route, nil
			}
		}

		return nil, nil
	}

	return nil, nil
}

// routeDestinationAttr return the route's destination attribute name.
// No validation is done.
func routeDestinationAttr(d *schema.ResourceData) string {
	destinationAttrs := []string{"destination_cidr_block", "destination_ipv6_cidr_block"}

	for _, attr := range destinationAttrs {
		if v := d.Get(attr).(string); v != "" {
			return attr
		}
	}

	return ""
}

// routeDestinationAndTargetAttributes return the route's destination and target attribute names.
// It also validates the resource, returning any validation error.
func routeDestinationAndTargetAttributes(d *schema.ResourceData) (string, string, error) {
	destinationAttrs := map[string]struct {
		ipv4Destination bool
		ipv6Destination bool
	}{
		"destination_cidr_block":      {true, false},
		"destination_ipv6_cidr_block": {false, true},
	}

	destinationAttr := ""
	ipv4Destination := false
	ipv6Destination := false
	for attr, kind := range destinationAttrs {
		if v := d.Get(attr).(string); v != "" {
			if destinationAttr != "" {
				return "", "", fmt.Errorf("`%s` conflicts with `%s`", attr, destinationAttr)
			}

			destinationAttr = attr
			ipv4Destination = kind.ipv4Destination
			ipv6Destination = kind.ipv6Destination
		}
	}

	if destinationAttr == "" {
		keys := []string{}
		for k := range destinationAttrs {
			keys = append(keys, k)
		}

		return "", "", fmt.Errorf("one of `%s` must be specified", strings.Join(keys, "`, `"))
	}

	targetAttrs := map[string]struct {
		ipv4Destination bool
		ipv6Destination bool
	}{
		"egress_only_gateway_id":    {false, true},
		"gateway_id":                {true, true},
		"instance_id":               {true, true},
		"nat_gateway_id":            {true, false},
		"network_interface_id":      {true, true},
		"transit_gateway_id":        {true, false},
		"vpc_peering_connection_id": {true, true},
	}

	targetAttr := ""
	for attr, allowed := range targetAttrs {
		if v := d.Get(attr).(string); v != "" {
			if targetAttr != "" {
				return "", "", fmt.Errorf("`%s` conflicts with `%s`", attr, targetAttr)
			}

			if (ipv4Destination && !allowed.ipv4Destination) || (ipv6Destination && !allowed.ipv6Destination) {
				return "", "", fmt.Errorf("`%s` not allowed for `%s` target", destinationAttr, attr)
			}

			targetAttr = attr
		}
	}

	if targetAttr == "" {
		keys := []string{}
		for k := range targetAttrs {
			keys = append(keys, k)
		}

		return "", "", fmt.Errorf("one of `%s` must be specified", strings.Join(keys, "`, `"))
	}

	return destinationAttr, targetAttr, nil
}

// TODO
// TODO Move these to a per-service internal package and auto-generate where possible.
// TODO

// readRouteTable returns the route table corresponding to the specified identifier.
// Returns nil if no route table is found.
func readRouteTable(conn *ec2.EC2, identifier string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: aws.StringSlice([]string{identifier}),
	}

	output, err := conn.DescribeRouteTables(input)
	if err != nil {
		return nil, err
	}

	if len(output.RouteTables) == 0 || output.RouteTables[0] == nil {
		return nil, nil
	}

	return output.RouteTables[0], nil
}

// readIpv4Route returns the route corresponding to the specified destination.
// Returns nil if no route is found.
func readIpv4Route(conn *ec2.EC2, routeTableID, destinationCidr string) (*ec2.Route, error) {
	routeTable, err := readRouteTable(conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if aws.StringValue(route.DestinationCidrBlock) == destinationCidr {
			return route, nil
		}
	}

	return nil, nil
}

// readIpv6Route returns the route corresponding to the specified destination.
// Returns nil if no route is found.
func readIpv6Route(conn *ec2.EC2, routeTableID, destinationIpv6Cidr string) (*ec2.Route, error) {
	routeTable, err := readRouteTable(conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if cidrBlocksEqual(aws.StringValue(route.DestinationIpv6CidrBlock), destinationIpv6Cidr) {
			return route, nil
		}
	}

	return nil, nil
}

func routeCreateID(routeTableID, destination string) string {
	return fmt.Sprintf("r-%s%d", routeTableID, hashcode.String(destination))
}
