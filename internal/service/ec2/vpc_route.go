// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	routeDestinationCIDRBlock     = "destination_cidr_block"
	routeDestinationIPv6CIDRBlock = "destination_ipv6_cidr_block"
	routeDestinationPrefixListID  = "destination_prefix_list_id"
)

var routeValidDestinations = []string{
	routeDestinationCIDRBlock,
	routeDestinationIPv6CIDRBlock,
	routeDestinationPrefixListID,
}

var routeValidTargets = []string{
	"carrier_gateway_id",
	"core_network_arn",
	"egress_only_gateway_id",
	"gateway_id",
	"local_gateway_id",
	"nat_gateway_id",
	names.AttrNetworkInterfaceID,
	names.AttrTransitGatewayID,
	names.AttrVPCEndpointID,
	"vpc_peering_connection_id",
}

// @SDKResource("aws_route", name="Route")
func resourceRoute() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteCreate,
		ReadWithoutTimeout:   resourceRouteRead,
		UpdateWithoutTimeout: resourceRouteUpdate,
		DeleteWithoutTimeout: resourceRouteDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRouteImport,
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
			routeDestinationCIDRBlock: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
				ExactlyOneOf: routeValidDestinations,
			},
			routeDestinationIPv6CIDRBlock: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     verify.ValidIPv6CIDRNetworkAddress,
				DiffSuppressFunc: suppressEqualCIDRBlockDiffs,
				ExactlyOneOf:     routeValidDestinations,
			},
			routeDestinationPrefixListID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: routeValidDestinations,
			},

			//
			// Targets.
			//
			"carrier_gateway_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ExactlyOneOf:  routeValidTargets,
				ConflictsWith: []string{routeDestinationIPv6CIDRBlock}, // IPv4 destinations only.
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
				ConflictsWith: []string{routeDestinationCIDRBlock}, // IPv6 destinations only.
			},
			"gateway_id": {
				Type:         schema.TypeString,
				Optional:     true,
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
			names.AttrNetworkInterfaceID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: routeValidTargets,
			},
			names.AttrTransitGatewayID: {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
			},
			names.AttrVPCEndpointID: {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: routeValidTargets,
				ConflictsWith: []string{
					routeDestinationPrefixListID, // "Cannot create or replace a prefix list route targeting a VPC Endpoint."
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
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"origin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	destinationAttributeKey, destination, err := routeDestinationAttribute(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	targetAttributeKey, target, err := routeTargetAttribute(d)
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.CreateRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder routeFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case routeDestinationCIDRBlock:
		input.DestinationCidrBlock = destination
		routeFinder = findRouteByIPv4Destination
	case routeDestinationIPv6CIDRBlock:
		input.DestinationIpv6CidrBlock = destination
		routeFinder = findRouteByIPv6Destination
	case routeDestinationPrefixListID:
		input.DestinationPrefixListId = destination
		routeFinder = findRouteByPrefixListIDDestination
	default:
		return sdkdiag.AppendErrorf(diags, "creating Route: unexpected route destination attribute: %q", destinationAttributeKey)
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
		return sdkdiag.AppendErrorf(diags, "creating Route: unexpected route target attribute: %q", targetAttributeKey)
	}

	route, err := routeFinder(ctx, conn, routeTableID, destination)

	switch {
	case err == nil:
		if route.Origin == awstypes.RouteOriginCreateRoute {
			return sdkdiag.AppendFromErr(diags, routeAlreadyExistsError(routeTableID, destination))
		}
	case tfresource.NotFound(err):
	default:
		return sdkdiag.AppendErrorf(diags, "reading Route: %s", err)
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate),
		func() (interface{}, error) {
			return conn.CreateRoute(ctx, input)
		},
		errCodeInvalidParameterException,
		errCodeInvalidTransitGatewayIDNotFound,
	)

	// Local routes cannot be created manually.
	if tfawserr.ErrMessageContains(err, errCodeInvalidGatewayIDNotFound, "The gateway ID 'local' does not exist") {
		return sdkdiag.AppendErrorf(diags, "cannot create local Route, use `terraform import` to manage existing local Routes")
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route in Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	d.SetId(routeCreateID(routeTableID, destination))

	if _, err := waitRouteReady(ctx, conn, routeFinder, routeTableID, destination, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route in Route Table (%s) with destination (%s) create: %s", routeTableID, destination, err)
	}

	return append(diags, resourceRouteRead(ctx, d, meta)...)
}

func resourceRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	destinationAttributeKey, destination, err := routeDestinationAttribute(d)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	var routeFinder routeFinder
	switch destinationAttributeKey {
	case routeDestinationCIDRBlock:
		routeFinder = findRouteByIPv4Destination
	case routeDestinationIPv6CIDRBlock:
		routeFinder = findRouteByIPv6Destination
	case routeDestinationPrefixListID:
		routeFinder = findRouteByPrefixListIDDestination
	default:
		return sdkdiag.AppendErrorf(diags, "reading Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	routeTableID := d.Get("route_table_id").(string)
	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return routeFinder(ctx, conn, routeTableID, destination)
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route in Route Table (%s) with destination (%s) not found, removing from state", routeTableID, destination)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route in Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	route := outputRaw.(*awstypes.Route)
	d.Set("carrier_gateway_id", route.CarrierGatewayId)
	d.Set("core_network_arn", route.CoreNetworkArn)
	d.Set(routeDestinationCIDRBlock, route.DestinationCidrBlock)
	d.Set(routeDestinationIPv6CIDRBlock, route.DestinationIpv6CidrBlock)
	d.Set(routeDestinationPrefixListID, route.DestinationPrefixListId)
	// VPC Endpoint ID is returned in Gateway ID field
	if strings.HasPrefix(aws.ToString(route.GatewayId), "vpce-") {
		d.Set("gateway_id", "")
		d.Set(names.AttrVPCEndpointID, route.GatewayId)
	} else {
		d.Set("gateway_id", route.GatewayId)
		d.Set(names.AttrVPCEndpointID, "")
	}
	d.Set("egress_only_gateway_id", route.EgressOnlyInternetGatewayId)
	d.Set("nat_gateway_id", route.NatGatewayId)
	d.Set("local_gateway_id", route.LocalGatewayId)
	d.Set(names.AttrInstanceID, route.InstanceId)
	d.Set("instance_owner_id", route.InstanceOwnerId)
	d.Set(names.AttrNetworkInterfaceID, route.NetworkInterfaceId)
	d.Set("origin", route.Origin)
	d.Set(names.AttrState, route.State)
	d.Set(names.AttrTransitGatewayID, route.TransitGatewayId)
	d.Set("vpc_peering_connection_id", route.VpcPeeringConnectionId)

	return diags
}

func resourceRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	destinationAttributeKey, destination, err := routeDestinationAttribute(d)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	targetAttributeKey, target, err := routeTargetAttribute(d)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.ReplaceRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder routeFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case routeDestinationCIDRBlock:
		input.DestinationCidrBlock = destination
		routeFinder = findRouteByIPv4Destination
	case routeDestinationIPv6CIDRBlock:
		input.DestinationIpv6CidrBlock = destination
		routeFinder = findRouteByIPv6Destination
	case routeDestinationPrefixListID:
		input.DestinationPrefixListId = destination
		routeFinder = findRouteByPrefixListIDDestination
	default:
		return sdkdiag.AppendErrorf(diags, "updating Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	localTarget := target == gatewayIDLocal
	switch target := aws.String(target); targetAttributeKey {
	case "carrier_gateway_id":
		input.CarrierGatewayId = target
	case "core_network_arn":
		input.CoreNetworkArn = target
	case "egress_only_gateway_id":
		input.EgressOnlyInternetGatewayId = target
	case "gateway_id":
		if localTarget {
			input.LocalTarget = aws.Bool(true)
		} else {
			input.GatewayId = target
		}
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
		return sdkdiag.AppendErrorf(diags, "updating Route: unexpected route target attribute: %q", targetAttributeKey)
	}

	log.Printf("[DEBUG] Updating Route: %v", input)
	_, err = conn.ReplaceRoute(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Route in Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	if _, err := waitRouteReady(ctx, conn, routeFinder, routeTableID, destination, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route in Route Table (%s) with destination (%s) update: %s", routeTableID, destination, err)
	}

	return append(diags, resourceRouteRead(ctx, d, meta)...)
}

func resourceRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	destinationAttributeKey, destination, err := routeDestinationAttribute(d)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	routeTableID := d.Get("route_table_id").(string)
	input := &ec2.DeleteRouteInput{
		RouteTableId: aws.String(routeTableID),
	}

	var routeFinder routeFinder

	switch destination := aws.String(destination); destinationAttributeKey {
	case routeDestinationCIDRBlock:
		input.DestinationCidrBlock = destination
		routeFinder = findRouteByIPv4Destination
	case routeDestinationIPv6CIDRBlock:
		input.DestinationIpv6CidrBlock = destination
		routeFinder = findRouteByIPv6Destination
	case routeDestinationPrefixListID:
		input.DestinationPrefixListId = destination
		routeFinder = findRouteByPrefixListIDDestination
	default:
		return sdkdiag.AppendErrorf(diags, "deleting Route: unexpected route destination attribute: %q", destinationAttributeKey)
	}

	log.Printf("[DEBUG] Deleting Route: %v", input)
	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete),
		func() (interface{}, error) {
			return conn.DeleteRoute(ctx, input)
		},
		errCodeInvalidParameterException,
	)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteNotFound) {
		return diags
	}

	// Local routes (which may have been imported) cannot be deleted. Remove from state.
	if tfawserr.ErrMessageContains(err, errCodeInvalidParameterValue, "cannot remove local route") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route in Route Table (%s) with destination (%s): %s", routeTableID, destination, err)
	}

	if _, err := waitRouteDeleted(ctx, conn, routeFinder, routeTableID, destination, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Route in Route Table (%s) with destination (%s) delete: %s", routeTableID, destination, err)
	}

	return diags
}

func resourceRouteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "_")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected ROUTETABLEID_DESTINATION", d.Id())
	}

	routeTableID := idParts[0]
	destination := idParts[1]
	d.Set("route_table_id", routeTableID)
	if strings.Contains(destination, ":") {
		d.Set(routeDestinationIPv6CIDRBlock, destination)
	} else if strings.Contains(destination, ".") {
		d.Set(routeDestinationCIDRBlock, destination)
	} else {
		d.Set(routeDestinationPrefixListID, destination)
	}

	d.SetId(routeCreateID(routeTableID, destination))

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
