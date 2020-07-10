package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
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
