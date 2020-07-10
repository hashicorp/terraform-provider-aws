package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

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

			"local_gateway_id": {
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

			"vpc_endpoint_id": {
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

	destination := d.Get(destinationAttr).(string)
	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.CreateRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder finder.RouteFinder

	switch destinationAttr {
	case "destination_cidr_block":
		input.DestinationCidrBlock = aws.String(destination)
		routeFinder = finder.RouteByIpv4Destination
	case "destination_ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = aws.String(destination)
		routeFinder = finder.RouteByIpv6Destination
	default:
		return fmt.Errorf("unexpected destination attribute: `%s`", destinationAttr)
	}

	target := aws.String(d.Get(targetAttr).(string))
	switch targetAttr {
	case "egress_only_gateway_id":
		input.EgressOnlyInternetGatewayId = target
	case "gateway_id":
		input.GatewayId = target
	case "instance_id":
		input.InstanceId = target
	case "local_gateway_id":
		input.LocalGatewayId = target
	case "nat_gateway_id":
		input.NatGatewayId = target
	case "network_interface_id":
		input.NetworkInterfaceId = target
	case "transit_gateway_id":
		input.TransitGatewayId = target
	case "vpc_endpoint_id":
		input.VpcEndpointId = target
	case "vpc_peering_connection_id":
		input.VpcPeeringConnectionId = target
	default:
		return fmt.Errorf("unexpected target attribute: `%s`", targetAttr)
	}

	log.Printf("[DEBUG] Creating Route: %s", input)
	err = createRoute(conn, input, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return err
	}

	var route *ec2.Route
	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		route, err = routeFinder(conn, routeTableID, destination)

		if err != nil {
			return resource.RetryableError(err)
		}

		if route == nil {
			return resource.RetryableError(fmt.Errorf("Route not found"))
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		route, err = routeFinder(conn, routeTableID, destination)
	}

	if err != nil {
		return fmt.Errorf("error reading Route for Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	if route == nil {
		return fmt.Errorf("Route in Route Table (%s) with destination (%s) not found", routeTableID, destination)
	}

	d.SetId(routeCreateID(routeTableID, destination))

	return resourceAwsRouteRead(d, meta)
}

func resourceAwsRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	destinationAttr := routeDestinationAttr(d)

	destination := d.Get(destinationAttr).(string)
	routeTableID := d.Get("route_table_id").(string)

	var routeFinder finder.RouteFinder

	switch destinationAttr {
	case "destination_cidr_block":
		routeFinder = finder.RouteByIpv4Destination
	case "destination_ipv6_cidr_block":
		routeFinder = finder.RouteByIpv6Destination
	default:
		return fmt.Errorf("unexpected destination attribute: `%s`", destinationAttr)
	}

	route, err := routeFinder(conn, routeTableID, destination)

	if isAWSErr(err, tfec2.ErrCodeRouteTableNotFound, "") {
		log.Printf("[WARN] Route Table (%s) not found, removing from state", routeTableID)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route for Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	if route == nil {
		log.Printf("[WARN] Route in Route Table (%s) with destination (%s) not found, removing from state", routeTableID, destination)
		d.SetId("")
		return nil
	}

	d.Set("destination_cidr_block", route.DestinationCidrBlock)
	d.Set("destination_ipv6_cidr_block", route.DestinationIpv6CidrBlock)
	d.Set("destination_prefix_list_id", route.DestinationPrefixListId)
	// VPC Endpoint ID is returned in Gateway ID field
	if strings.HasPrefix(aws.StringValue(route.GatewayId), "vpce-") {
		d.Set("vpc_endpoint_id", route.GatewayId)
	} else {
		d.Set("gateway_id", route.GatewayId)
	}
	d.Set("egress_only_gateway_id", route.EgressOnlyInternetGatewayId)
	d.Set("nat_gateway_id", route.NatGatewayId)
	d.Set("local_gateway_id", route.LocalGatewayId)
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

	destinationAttr, targetAttr, err := routeDestinationAndTargetAttributes(d)
	if err != nil {
		return err
	}

	destination := d.Get(destinationAttr).(string)
	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	switch destinationAttr {
	case "destination_cidr_block":
		input.DestinationCidrBlock = aws.String(destination)
	case "destination_ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = aws.String(destination)
	default:
		return fmt.Errorf("unexpected destination attribute: `%s`", destinationAttr)
	}

	target := aws.String(d.Get(targetAttr).(string))
	switch targetAttr {
	case "egress_only_gateway_id":
		input.EgressOnlyInternetGatewayId = target
	case "gateway_id":
		input.GatewayId = target
	case "instance_id":
		input.InstanceId = target
	case "local_gateway_id":
		input.LocalGatewayId = target
	case "nat_gateway_id":
		input.NatGatewayId = target
	case "network_interface_id":
		input.NetworkInterfaceId = target
	case "transit_gateway_id":
		input.TransitGatewayId = target
	case "vpc_endpoint_id":
		input.VpcEndpointId = target
	case "vpc_peering_connection_id":
		input.VpcPeeringConnectionId = target
	default:
		return fmt.Errorf("unexpected target attribute: `%s`", targetAttr)
	}

	log.Printf("[DEBUG] Updating Route: %s", input)
	_, err = conn.ReplaceRoute(input)

	if err != nil {
		return fmt.Errorf("error updating Route for Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	return resourceAwsRouteRead(d, meta)
}

func resourceAwsRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	destinationAttr := routeDestinationAttr(d)

	destination := d.Get(destinationAttr).(string)
	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.DeleteRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	switch destinationAttr {
	case "destination_cidr_block":
		input.DestinationCidrBlock = aws.String(destination)
	case "destination_ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = aws.String(destination)
	default:
		return fmt.Errorf("unexpected destination attribute: `%s`", destinationAttr)
	}

	err := resource.Retry(d.Timeout(schema.TimeoutDelete), func() *resource.RetryError {
		log.Printf("[DEBUG] Deleting Route (%s)", d.Id())
		_, err := conn.DeleteRoute(input)
		if err == nil {
			return nil
		}

		if isAWSErr(err, tfec2.ErrCodeRouteNotFound, "") {
			return nil
		}

		// Local routes (which may have been imported) cannot be deleted. Remove from state.
		if isAWSErr(err, tfec2.ErrCodeInvalidParameterValue, "cannot remove local route") {
			return nil
		}

		if isAWSErr(err, tfec2.ErrCodeInvalidParameterException, "") {
			return resource.RetryableError(err)
		}

		return resource.NonRetryableError(err)
	})

	if isResourceTimeoutError(err) {
		log.Printf("[DEBUG] Deleting Route (%s)", d.Id())
		_, err = conn.DeleteRoute(input)
	}

	if isAWSErr(err, tfec2.ErrCodeRouteNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route for Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	return nil
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
		"vpc_endpoint_id":           {true, false},
		"vpc_peering_connection_id": {true, true},
	}

	targetAttr := ""
	for attr, allowed := range targetAttrs {
		if v := d.Get(attr).(string); v != "" && d.HasChange(attr) {
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

func routeCreateID(routeTableID, destination string) string {
	return fmt.Sprintf("r-%s%d", routeTableID, hashcode.String(destination))
}

// createRoute attempts to create a route.
// The specified eventual consistency timeout is respected.
// Any error is returned.
func createRoute(conn *ec2.EC2, input *ec2.CreateRouteInput, timeout time.Duration) error {
	err := resource.Retry(timeout, func() *resource.RetryError {
		_, err := conn.CreateRoute(input)

		if isAWSErr(err, tfec2.ErrCodeInvalidParameterException, "") {
			return resource.RetryableError(err)
		}

		if isAWSErr(err, tfec2.ErrCodeTransitGatewayNotFound, "") {
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
		return fmt.Errorf("error creating Route: %s", err)
	}

	return nil
}
