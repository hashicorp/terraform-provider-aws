package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	tfnet "github.com/terraform-providers/terraform-provider-aws/aws/internal/net"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
)

func ClientVpnAuthorizationRule(conn *ec2.EC2, endpointID, targetNetworkCidr, accessGroupID string) (*ec2.DescribeClientVpnAuthorizationRulesOutput, error) {
	filters := map[string]string{
		"destination-cidr": targetNetworkCidr,
	}
	if accessGroupID != "" {
		filters["group-id"] = accessGroupID
	}

	input := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             tfec2.BuildAttributeFilterList(filters),
	}

	return conn.DescribeClientVpnAuthorizationRules(input)

}

func ClientVpnAuthorizationRuleByID(conn *ec2.EC2, authorizationRuleID string) (*ec2.DescribeClientVpnAuthorizationRulesOutput, error) {
	endpointID, targetNetworkCidr, accessGroupID, err := tfec2.ClientVpnAuthorizationRuleParseID(authorizationRuleID)
	if err != nil {
		return nil, err
	}

	return ClientVpnAuthorizationRule(conn, endpointID, targetNetworkCidr, accessGroupID)
}

func ClientVpnRoute(conn *ec2.EC2, endpointID, targetSubnetID, destinationCidr string) (*ec2.DescribeClientVpnRoutesOutput, error) {
	filters := map[string]string{
		"target-subnet":    targetSubnetID,
		"destination-cidr": destinationCidr,
	}

	input := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             tfec2.BuildAttributeFilterList(filters),
	}

	return conn.DescribeClientVpnRoutes(input)
}

func ClientVpnRouteByID(conn *ec2.EC2, routeID string) (*ec2.DescribeClientVpnRoutesOutput, error) {
	endpointID, targetSubnetID, destinationCidr, err := tfec2.ClientVpnRouteParseID(routeID)
	if err != nil {
		return nil, err
	}

	return ClientVpnRoute(conn, endpointID, targetSubnetID, destinationCidr)
}

// RouteTableByID returns the route table corresponding to the specified identifier.
// Returns nil if no route table is found.
func RouteTableByID(conn *ec2.EC2, routeTableID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	output, err := conn.DescribeRouteTables(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.RouteTables) == 0 || output.RouteTables[0] == nil {
		return nil, nil
	}

	return output.RouteTables[0], nil
}

type RouteFinder func(*ec2.EC2, string, string) (*ec2.Route, error)

// RouteByIpv4Destination returns the route corresponding to the specified IPv4 destination.
// Returns nil if no route is found.
func RouteByIpv4Destination(conn *ec2.EC2, routeTableID, destinationCidr string) (*ec2.Route, error) {
	routeTable, err := RouteTableByID(conn, routeTableID)
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

// RouteByIpv6Destination returns the route corresponding to the specified IPv6 destination.
// Returns nil if no route is found.
func RouteByIpv6Destination(conn *ec2.EC2, routeTableID, destinationIpv6Cidr string) (*ec2.Route, error) {
	routeTable, err := RouteTableByID(conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if tfnet.CIDRBlocksEqual(aws.StringValue(route.DestinationIpv6CidrBlock), destinationIpv6Cidr) {
			return route, nil
		}
	}

	return nil, nil
}

// SecurityGroupByID looks up a security group by ID. When not found, returns nil and potentially an API error.
func SecurityGroupByID(conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	req := &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{id}),
	}
	result, err := conn.DescribeSecurityGroups(req)
	if err != nil {
		return nil, err
	}

	if result == nil || len(result.SecurityGroups) == 0 || result.SecurityGroups[0] == nil {
		return nil, nil
	}

	return result.SecurityGroups[0], nil
}

// VpnGatewayVpcAttachment returns the attachment between the specified VPN gateway and VPC.
// Returns nil and potentially an error if no attachment is found.
func VpnGatewayVpcAttachment(conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	vpnGateway, err := VpnGatewayByID(conn, vpnGatewayID)
	if err != nil {
		return nil, err
	}

	if vpnGateway == nil {
		return nil, nil
	}

	for _, vpcAttachment := range vpnGateway.VpcAttachments {
		if aws.StringValue(vpcAttachment.VpcId) == vpcID {
			return vpcAttachment, nil
		}
	}

	return nil, nil
}

// VpnGatewayByID returns the VPN gateway corresponding to the specified identifier.
// Returns nil and potentially an error if no VPN gateway is found.
func VpnGatewayByID(conn *ec2.EC2, id string) (*ec2.VpnGateway, error) {
	input := &ec2.DescribeVpnGatewaysInput{
		VpnGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeVpnGateways(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpnGateways) == 0 {
		return nil, nil
	}

	return output.VpnGateways[0], nil
}
