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

// InstanceTypeInfo returns the details of the specified instance type.
// Returns nil if no instance type details are found.
func InstanceTypeInfo(conn *ec2.EC2, instanceType string) (*ec2.InstanceTypeInfo, error) {
	input := &ec2.DescribeInstanceTypesInput{
		InstanceTypes: aws.StringSlice([]string{instanceType}),
	}

	output, err := conn.DescribeInstanceTypes(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.InstanceTypes) == 0 {
		return nil, nil
	}

	return output.InstanceTypes[0], nil
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
