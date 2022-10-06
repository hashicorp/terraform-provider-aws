package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

var routeValidDestinations = []string{
	"destination_cidr_block",
	"destination_ipv6_cidr_block",
	"destination_prefix_list_id",
}

var routeValidTargets = []string{
	"carrier_gateway_id",
	"core_network_arn",
	"egress_only_gateway_id",
	"gateway_id",
	"instance_id",
	"local_gateway_id",
	"nat_gateway_id",
	"network_interface_id",
	"transit_gateway_id",
	"vpc_endpoint_id",
	"vpc_peering_connection_id",
}

func ResourceRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceRouteCreate,
		Read:   resourceRouteRead,
		Update: resourceRouteUpdate,
		Delete: resourceRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceRouteImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			///
			// Destinations.
			///
			"destination_cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
				ExactlyOneOf: []string{"destination_cidr_block", "destination_ipv6_cidr_block", "destination_prefix_list_id"},
			},
			"destination_ipv6_cidr_block": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     verify.ValidIPv6CIDRNetworkAddress,
				DiffSuppressFunc: suppressEqualCIDRBlockDiffs,
				ExactlyOneOf:     []string{"destination_cidr_block", "destination_ipv6_cidr_block", "destination_prefix_list_id"},
			},
			"destination_prefix_list_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"destination_cidr_block", "destination_ipv6_cidr_block", "destination_prefix_list_id"},
			},

			//
			// Targets.
			//
			"carrier_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ExactlyOneOf:  routeValidTargets,
				ConflictsWith: []string{"destination_ipv6_cidr_block"}, // IPv4 destinations only.
			},
			"core_network_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
			},
			"egress_only_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ExactlyOneOf:  routeValidTargets,
				ConflictsWith: []string{"destination_cidr_block"}, // IPv6 destinations only.
			},
			"gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
			},
			"instance_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Deprecated:   "Use network_interface_id instead",
				ExactlyOneOf: routeValidTargets,
			},
			"local_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
			},
			"nat_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
			},
			"network_interface_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: routeValidTargets,
			},
			"transit_gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
			},
			"vpc_endpoint_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
				ConflictsWith: []string{
					"destination_ipv6_cidr_block", // IPv4 destinations only.
					"destination_prefix_list_id",  // "Cannot create or replace a prefix list route targeting a VPC Endpoint."
				},
			},
			"vpc_peering_connection_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
			},

			//
			// Computed attributes.
			//
			"instance_owner_id": {
				Type:     schema.TypeString,
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
		},
	}
}

func resourceRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	destinationAttributeKey, destination, err := routeDestinationAttribute(d)

	if err != nil {
		return fmt.Errorf("error creating Route: %w", err)
	}

	targetAttributeKey, target, err := routeTargetAttribute(d)

	if err != nil {
		return fmt.Errorf("error creating Route: %w", err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.CreateRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder RouteFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case "destination_cidr_block":
		input.DestinationCidrBlock = destination
		routeFinder = FindRouteByIPv4Destination
	case "destination_ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = destination
		routeFinder = FindRouteByIPv6Destination
	case "destination_prefix_list_id":
		input.DestinationPrefixListId = destination
		routeFinder = FindRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("error creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	switch target := aws.String(target); targetAttributeKey {
	case "carrier_gateway_id":
		input.CarrierGatewayId = target
	case "core_network_arn":
		input.CoreNetworkArn = target
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
		return fmt.Errorf("error creating Route: unexpected route target attribute: %q", targetAttributeKey)
	}

	log.Printf("[DEBUG] Creating Route: %s", input)
	_, err = tfresource.RetryWhenAWSErrCodeEquals(
		d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateRoute(input)
		},
		errCodeInvalidParameterException,
		errCodeInvalidTransitGatewayIDNotFound,
	)

	if err != nil {
		return fmt.Errorf("error creating Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	d.SetId(RouteCreateID(routeTableID, destination))

	_, err = WaitRouteReady(conn, routeFinder, routeTableID, destination, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return fmt.Errorf("error waiting for Route in Route Table (%s) with destination (%s) to become available: %w", routeTableID, destination, err)
	}

	return resourceRouteRead(d, meta)
}

func resourceRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	destinationAttributeKey, destination, err := routeDestinationAttribute(d)

	if err != nil {
		return fmt.Errorf("error reading Route: %w", err)
	}

	var routeFinder RouteFinder

	switch destinationAttributeKey {
	case "destination_cidr_block":
		routeFinder = FindRouteByIPv4Destination
	case "destination_ipv6_cidr_block":
		routeFinder = FindRouteByIPv6Destination
	case "destination_prefix_list_id":
		routeFinder = FindRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("error reading Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	routeTableID := d.Get("route_table_id").(string)

	route, err := routeFinder(conn, routeTableID, destination)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route in Route Table (%s) with destination (%s) not found, removing from state", routeTableID, destination)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	d.Set("carrier_gateway_id", route.CarrierGatewayId)
	d.Set("core_network_arn", route.CoreNetworkArn)
	d.Set("destination_cidr_block", route.DestinationCidrBlock)
	d.Set("destination_ipv6_cidr_block", route.DestinationIpv6CidrBlock)
	d.Set("destination_prefix_list_id", route.DestinationPrefixListId)
	// VPC Endpoint ID is returned in Gateway ID field
	if strings.HasPrefix(aws.StringValue(route.GatewayId), "vpce-") {
		d.Set("gateway_id", "")
		d.Set("vpc_endpoint_id", route.GatewayId)
	} else {
		d.Set("gateway_id", route.GatewayId)
		d.Set("vpc_endpoint_id", "")
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

func resourceRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	destinationAttributeKey, destination, err := routeDestinationAttribute(d)

	if err != nil {
		return fmt.Errorf("error updating Route: %w", err)
	}

	targetAttributeKey, target, err := routeTargetAttribute(d)

	if err != nil {
		return fmt.Errorf("error updating Route: %w", err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder RouteFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case "destination_cidr_block":
		input.DestinationCidrBlock = destination
		routeFinder = FindRouteByIPv4Destination
	case "destination_ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = destination
		routeFinder = FindRouteByIPv6Destination
	case "destination_prefix_list_id":
		input.DestinationPrefixListId = destination
		routeFinder = FindRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("error updating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	switch target := aws.String(target); targetAttributeKey {
	case "carrier_gateway_id":
		input.CarrierGatewayId = target
	case "core_network_arn":
		input.CoreNetworkArn = target
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
		return fmt.Errorf("error updating Route: unexpected route target attribute: %q", targetAttributeKey)
	}

	log.Printf("[DEBUG] Updating Route: %s", input)
	_, err = conn.ReplaceRoute(input)

	if err != nil {
		return fmt.Errorf("error updating Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	_, err = WaitRouteReady(conn, routeFinder, routeTableID, destination, d.Timeout(schema.TimeoutUpdate))

	if err != nil {
		return fmt.Errorf("error waiting for Route in Route Table (%s) with destination (%s) to become available: %w", routeTableID, destination, err)
	}

	return resourceRouteRead(d, meta)
}

func resourceRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	destinationAttributeKey, destination, err := routeDestinationAttribute(d)

	if err != nil {
		return fmt.Errorf("error deleting Route: %w", err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.DeleteRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder RouteFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case "destination_cidr_block":
		input.DestinationCidrBlock = destination
		routeFinder = FindRouteByIPv4Destination
	case "destination_ipv6_cidr_block":
		input.DestinationIpv6CidrBlock = destination
		routeFinder = FindRouteByIPv6Destination
	case "destination_prefix_list_id":
		input.DestinationPrefixListId = destination
		routeFinder = FindRouteByPrefixListIDDestination
	default:
		return fmt.Errorf("error deleting Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	log.Printf("[DEBUG] Deleting Route: %s", input)
	_, err = tfresource.RetryWhenAWSErrCodeEquals(
		d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteRoute(input)
		},
		errCodeInvalidParameterException,
	)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteNotFound) {
		return nil
	}

	// Local routes (which may have been imported) cannot be deleted. Remove from state.
	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "cannot remove local route") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route in Route Table (%s) with destination (%s): %w", routeTableID, destination, err)
	}

	_, err = WaitRouteDeleted(conn, routeFinder, routeTableID, destination, d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return fmt.Errorf("error waiting for Route in Route Table (%s) with destination (%s) to delete: %w", routeTableID, destination, err)
	}

	return nil
}

func resourceRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected ROUTETABLEID_DESTINATION", d.Id())
	}

	routeTableID := idParts[0]
	destination := idParts[1]
	d.Set("route_table_id", routeTableID)
	if strings.Contains(destination, ":") {
		d.Set("destination_ipv6_cidr_block", destination)
	} else if strings.Contains(destination, ".") {
		d.Set("destination_cidr_block", destination)
	} else {
		d.Set("destination_prefix_list_id", destination)
	}

	d.SetId(RouteCreateID(routeTableID, destination))

	return []*schema.ResourceData{d}, nil
}

// routeDestinationAttribute returns the attribute key and value of the route's destination.
func routeDestinationAttribute(d *schema.ResourceData) (string, string, error) {
	for _, key := range routeValidDestinations {
		if v, ok := d.Get(key).(string); ok && v != "" {
			return key, v, nil
		}
	}

	return "", "", fmt.Errorf("route destination attribute not specified")
}

// routeTargetAttribute returns the attribute key and value of the route's target.
func routeTargetAttribute(d *schema.ResourceData) (string, string, error) {
	for _, key := range routeValidTargets {
		// The HasChange check is necessary to handle Computed attributes that will be cleared once they are read back after update.
		if v, ok := d.Get(key).(string); ok && v != "" && d.HasChange(key) {
			return key, v, nil
		}
	}

	return "", "", fmt.Errorf("route target attribute not specified")
}
