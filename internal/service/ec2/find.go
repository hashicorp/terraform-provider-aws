package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// FindCarrierGatewayByID returns the carrier gateway corresponding to the specified identifier.
// Returns nil and potentially an error if no carrier gateway is found.
func FindCarrierGatewayByID(conn *ec2.EC2, id string) (*ec2.CarrierGateway, error) {
	input := &ec2.DescribeCarrierGatewaysInput{
		CarrierGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeCarrierGateways(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.CarrierGateways) == 0 {
		return nil, nil
	}

	return output.CarrierGateways[0], nil
}

func FindClientVPNAuthorizationRule(conn *ec2.EC2, endpointID, targetNetworkCidr, accessGroupID string) (*ec2.DescribeClientVpnAuthorizationRulesOutput, error) {
	filters := map[string]string{
		"destination-cidr": targetNetworkCidr,
	}
	if accessGroupID != "" {
		filters["group-id"] = accessGroupID
	}

	input := &ec2.DescribeClientVpnAuthorizationRulesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             BuildAttributeFilterList(filters),
	}

	return conn.DescribeClientVpnAuthorizationRules(input)

}

func FindClientVPNAuthorizationRuleByID(conn *ec2.EC2, authorizationRuleID string) (*ec2.DescribeClientVpnAuthorizationRulesOutput, error) {
	endpointID, targetNetworkCidr, accessGroupID, err := ClientVPNAuthorizationRuleParseID(authorizationRuleID)
	if err != nil {
		return nil, err
	}

	return FindClientVPNAuthorizationRule(conn, endpointID, targetNetworkCidr, accessGroupID)
}

func FindClientVPNRoute(conn *ec2.EC2, endpointID, targetSubnetID, destinationCidr string) (*ec2.DescribeClientVpnRoutesOutput, error) {
	filters := map[string]string{
		"target-subnet":    targetSubnetID,
		"destination-cidr": destinationCidr,
	}

	input := &ec2.DescribeClientVpnRoutesInput{
		ClientVpnEndpointId: aws.String(endpointID),
		Filters:             BuildAttributeFilterList(filters),
	}

	return conn.DescribeClientVpnRoutes(input)
}

func FindClientVPNRouteByID(conn *ec2.EC2, routeID string) (*ec2.DescribeClientVpnRoutesOutput, error) {
	endpointID, targetSubnetID, destinationCidr, err := ClientVPNRouteParseID(routeID)
	if err != nil {
		return nil, err
	}

	return FindClientVPNRoute(conn, endpointID, targetSubnetID, destinationCidr)
}

func FindHostByID(conn *ec2.EC2, id string) (*ec2.Host, error) {
	input := &ec2.DescribeHostsInput{
		HostIds: aws.StringSlice([]string{id}),
	}

	return FindHost(conn, input)
}

func FindHostByIDAndFilters(conn *ec2.EC2, id string, filters []*ec2.Filter) (*ec2.Host, error) {
	input := &ec2.DescribeHostsInput{}

	if id != "" {
		input.HostIds = aws.StringSlice([]string{id})
	}

	if len(filters) > 0 {
		input.Filter = filters
	}

	return FindHost(conn, input)
}

func FindHost(conn *ec2.EC2, input *ec2.DescribeHostsInput) (*ec2.Host, error) {
	output, err := conn.DescribeHosts(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidHostIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Hosts) == 0 || output.Hosts[0] == nil || output.Hosts[0].HostProperties == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Hosts); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	host := output.Hosts[0]

	if state := aws.StringValue(host.State); state == ec2.AllocationStateReleased || state == ec2.AllocationStateReleasedPermanentFailure {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return host, nil
}

// FindInstanceByID looks up a Instance by ID. When not found, returns nil and potentially an API error.
func FindInstanceByID(conn *ec2.EC2, id string) (*ec2.Instance, error) {
	input := &ec2.DescribeInstancesInput{
		InstanceIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeInstances(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Reservations) == 0 || output.Reservations[0] == nil || len(output.Reservations[0].Instances) == 0 || output.Reservations[0].Instances[0] == nil {
		return nil, nil
	}

	return output.Reservations[0].Instances[0], nil
}

// FindNetworkACLByID looks up a NetworkAcl by ID. When not found, returns nil and potentially an API error.
func FindNetworkACLByID(conn *ec2.EC2, id string) (*ec2.NetworkAcl, error) {
	input := &ec2.DescribeNetworkAclsInput{
		NetworkAclIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeNetworkAcls(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, networkAcl := range output.NetworkAcls {
		if networkAcl == nil {
			continue
		}

		if aws.StringValue(networkAcl.NetworkAclId) != id {
			continue
		}

		return networkAcl, nil
	}

	return nil, nil
}

// FindNetworkACLEntry looks up a FindNetworkACLEntry by Network ACL ID, Egress, and Rule Number. When not found, returns nil and potentially an API error.
func FindNetworkACLEntry(conn *ec2.EC2, networkAclID string, egress bool, ruleNumber int) (*ec2.NetworkAclEntry, error) {
	input := &ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("entry.egress"),
				Values: aws.StringSlice([]string{fmt.Sprintf("%t", egress)}),
			},
			{
				Name:   aws.String("entry.rule-number"),
				Values: aws.StringSlice([]string{fmt.Sprintf("%d", ruleNumber)}),
			},
		},
		NetworkAclIds: aws.StringSlice([]string{networkAclID}),
	}

	output, err := conn.DescribeNetworkAcls(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, networkAcl := range output.NetworkAcls {
		if networkAcl == nil {
			continue
		}

		if aws.StringValue(networkAcl.NetworkAclId) != networkAclID {
			continue
		}

		for _, entry := range output.NetworkAcls[0].Entries {
			if entry == nil {
				continue
			}

			if aws.BoolValue(entry.Egress) != egress || aws.Int64Value(entry.RuleNumber) != int64(ruleNumber) {
				continue
			}

			return entry, nil
		}
	}

	return nil, nil
}

func FindNetworkInterface(conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) (*ec2.NetworkInterface, error) {
	output, err := FindNetworkInterfaces(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindNetworkInterfaces(conn *ec2.EC2, input *ec2.DescribeNetworkInterfacesInput) ([]*ec2.NetworkInterface, error) {
	var output []*ec2.NetworkInterface

	err := conn.DescribeNetworkInterfacesPages(input, func(page *ec2.DescribeNetworkInterfacesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.NetworkInterfaces {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidNetworkInterfaceIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindNetworkInterfaceByID(conn *ec2.EC2, id string) (*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: aws.StringSlice([]string{id}),
	}

	output, err := FindNetworkInterface(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.NetworkInterfaceId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(conn *ec2.EC2, attachmentInstanceOwnerID, description string) ([]*ec2.NetworkInterface, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.instance-owner-id": attachmentInstanceOwnerID,
			"description":                  description,
		}),
	}

	return FindNetworkInterfaces(conn, input)
}

func FindNetworkInterfaceAttachmentByID(conn *ec2.EC2, id string) (*ec2.NetworkInterfaceAttachment, error) {
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"attachment.attachment-id": id,
		}),
	}

	networkInterface, err := FindNetworkInterface(conn, input)

	if err != nil {
		return nil, err
	}

	if networkInterface.Attachment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return networkInterface.Attachment, nil
}

func FindNetworkInterfaceSecurityGroup(conn *ec2.EC2, networkInterfaceID string, securityGroupID string) (*ec2.GroupIdentifier, error) {
	networkInterface, err := FindNetworkInterfaceByID(conn, networkInterfaceID)

	if err != nil {
		return nil, err
	}

	for _, groupIdentifier := range networkInterface.Groups {
		if aws.StringValue(groupIdentifier.GroupId) == securityGroupID {
			return groupIdentifier, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Network Interface (%s) Security Group (%s) not found", networkInterfaceID, securityGroupID),
	}
}

// FindMainRouteTableAssociationByID returns the main route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func FindMainRouteTableAssociationByID(conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	association, err := FindRouteTableAssociationByID(conn, associationID)

	if err != nil {
		return nil, err
	}

	if !aws.BoolValue(association.Main) {
		return nil, &resource.NotFoundError{
			Message: fmt.Sprintf("%s is not the association with the main route table", associationID),
		}
	}

	return association, err
}

// FindMainRouteTableAssociationByVPCID returns the main route table association for the specified VPC.
// Returns NotFoundError if no route table association is found.
func FindMainRouteTableAssociationByVPCID(conn *ec2.EC2, vpcID string) (*ec2.RouteTableAssociation, error) {
	routeTable, err := FindMainRouteTableByVPCID(conn, vpcID)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.BoolValue(association.Main) {
			if state := aws.StringValue(association.AssociationState.State); state == ec2.RouteTableAssociationStateCodeDisassociated {
				continue
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// FindRouteTableAssociationByID returns the route table association corresponding to the specified identifier.
// Returns NotFoundError if no route table association is found.
func FindRouteTableAssociationByID(conn *ec2.EC2, associationID string) (*ec2.RouteTableAssociation, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.route-table-association-id": associationID,
		}),
	}

	routeTable, err := FindRouteTable(conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range routeTable.Associations {
		if aws.StringValue(association.RouteTableAssociationId) == associationID {
			if state := aws.StringValue(association.AssociationState.State); state == ec2.RouteTableAssociationStateCodeDisassociated {
				return nil, &resource.NotFoundError{Message: state}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

// FindMainRouteTableByVPCID returns the main route table for the specified VPC.
// Returns NotFoundError if no route table is found.
func FindMainRouteTableByVPCID(conn *ec2.EC2, vpcID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"association.main": "true",
			"vpc-id":           vpcID,
		}),
	}

	return FindRouteTable(conn, input)
}

// FindRouteTableByID returns the route table corresponding to the specified identifier.
// Returns NotFoundError if no route table is found.
func FindRouteTableByID(conn *ec2.EC2, routeTableID string) (*ec2.RouteTable, error) {
	input := &ec2.DescribeRouteTablesInput{
		RouteTableIds: aws.StringSlice([]string{routeTableID}),
	}

	return FindRouteTable(conn, input)
}

// FindRouteTable returns the route table corresponding to the specified input.
// Returns EmptyResultError if no route table is found or TooManyResultsError if more than 1
// matching route table is found.
func FindRouteTable(conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) (*ec2.RouteTable, error) {
	output, err := FindRouteTables(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

// FindRouteTables returns an array of route tables for the specified input.
// Returns NotFoundError if no route table is found for a specified route table ID.
func FindRouteTables(conn *ec2.EC2, input *ec2.DescribeRouteTablesInput) ([]*ec2.RouteTable, error) {
	var output []*ec2.RouteTable

	err := conn.DescribeRouteTablesPages(input, func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, table := range page.RouteTables {
			if table == nil {
				continue
			}

			output = append(output, table)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// RouteFinder returns the route corresponding to the specified destination.
// Returns NotFoundError if no route is found.
type RouteFinder func(*ec2.EC2, string, string) (*ec2.Route, error)

// FindRouteByIPv4Destination returns the route corresponding to the specified IPv4 destination.
// Returns NotFoundError if no route is found.
func FindRouteByIPv4Destination(conn *ec2.EC2, routeTableID, destinationCidr string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if verify.CIDRBlocksEqual(aws.StringValue(route.DestinationCidrBlock), destinationCidr) {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv4 destination (%s) not found", routeTableID, destinationCidr),
	}
}

// FindRouteByIPv6Destination returns the route corresponding to the specified IPv6 destination.
// Returns NotFoundError if no route is found.
func FindRouteByIPv6Destination(conn *ec2.EC2, routeTableID, destinationIpv6Cidr string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if verify.CIDRBlocksEqual(aws.StringValue(route.DestinationIpv6CidrBlock), destinationIpv6Cidr) {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with IPv6 destination (%s) not found", routeTableID, destinationIpv6Cidr),
	}
}

// FindRouteByPrefixListIDDestination returns the route corresponding to the specified prefix list destination.
// Returns NotFoundError if no route is found.
func FindRouteByPrefixListIDDestination(conn *ec2.EC2, routeTableID, prefixListID string) (*ec2.Route, error) {
	routeTable, err := FindRouteTableByID(conn, routeTableID)
	if err != nil {
		return nil, err
	}

	for _, route := range routeTable.Routes {
		if aws.StringValue(route.DestinationPrefixListId) == prefixListID {
			return route, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError: fmt.Errorf("Route in Route Table (%s) with Prefix List ID destination (%s) not found", routeTableID, prefixListID),
	}
}

// FindSecurityGroupByID looks up a security group by ID. Returns a resource.NotFoundError if not found.
func FindSecurityGroupByID(conn *ec2.EC2, id string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{id}),
	}
	return FindSecurityGroup(conn, input)
}

// FindSecurityGroupByNameAndVPCID looks up a security group by name and VPC ID. Returns a resource.NotFoundError if not found.
func FindSecurityGroupByNameAndVPCID(conn *ec2.EC2, name, vpcID string) (*ec2.SecurityGroup, error) {
	input := &ec2.DescribeSecurityGroupsInput{
		Filters: BuildAttributeFilterList(
			map[string]string{
				"group-name": name,
				"vpc-id":     vpcID,
			},
		),
	}
	return FindSecurityGroup(conn, input)
}

// FindSecurityGroup looks up a security group using an ec2.DescribeSecurityGroupsInput. Returns a resource.NotFoundError if not found.
func FindSecurityGroup(conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) (*ec2.SecurityGroup, error) {
	output, err := FindSecurityGroups(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

// FindSecurityGroups returns an array of security groups that match an ec2.DescribeSecurityGroupsInput.
// Returns a resource.NotFoundError if no group is found for a specified SecurityGroup or SecurityGroupId.
func FindSecurityGroups(conn *ec2.EC2, input *ec2.DescribeSecurityGroupsInput) ([]*ec2.SecurityGroup, error) {
	var output []*ec2.SecurityGroup

	err := conn.DescribeSecurityGroupsPages(input, func(page *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, sg := range page.SecurityGroups {
			if sg == nil {
				continue
			}

			output = append(output, sg)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSecurityGroupIDNotFound) ||
		tfawserr.ErrCodeEquals(err, ErrCodeInvalidGroupNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

// FindSpotInstanceRequestByID looks up a SpotInstanceRequest by ID. When not found, returns nil and potentially an API error.
func FindSpotInstanceRequestByID(conn *ec2.EC2, id string) (*ec2.SpotInstanceRequest, error) {
	input := &ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeSpotInstanceRequests(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, spotInstanceRequest := range output.SpotInstanceRequests {
		if spotInstanceRequest == nil {
			continue
		}

		if aws.StringValue(spotInstanceRequest.SpotInstanceRequestId) != id {
			continue
		}

		return spotInstanceRequest, nil
	}

	return nil, nil
}

func FindSubnetByID(conn *ec2.EC2, id string) (*ec2.Subnet, error) {
	input := &ec2.DescribeSubnetsInput{
		SubnetIds: aws.StringSlice([]string{id}),
	}

	output, err := FindSubnet(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.SubnetId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindSubnet(conn *ec2.EC2, input *ec2.DescribeSubnetsInput) (*ec2.Subnet, error) {
	output, err := FindSubnets(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindSubnets(conn *ec2.EC2, input *ec2.DescribeSubnetsInput) ([]*ec2.Subnet, error) {
	var output []*ec2.Subnet

	err := conn.DescribeSubnetsPages(input, func(page *ec2.DescribeSubnetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Subnets {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSubnetIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindSubnetCidrReservationBySubnetIDAndReservationID(conn *ec2.EC2, subnetID, reservationID string) (*ec2.SubnetCidrReservation, error) {
	input := &ec2.GetSubnetCidrReservationsInput{
		SubnetId: aws.String(subnetID),
	}

	output, err := conn.GetSubnetCidrReservations(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidSubnetIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || (len(output.SubnetIpv4CidrReservations) == 0 && len(output.SubnetIpv6CidrReservations) == 0) {
		return nil, tfresource.NewEmptyResultError(input)
	}

	for _, r := range output.SubnetIpv4CidrReservations {
		if aws.StringValue(r.SubnetCidrReservationId) == reservationID {
			return r, nil
		}
	}
	for _, r := range output.SubnetIpv6CidrReservations {
		if aws.StringValue(r.SubnetCidrReservationId) == reservationID {
			return r, nil
		}
	}

	return nil, &resource.NotFoundError{
		LastError:   err,
		LastRequest: input,
	}
}

func FindSubnetIPv6CIDRBlockAssociationByID(conn *ec2.EC2, associationID string) (*ec2.SubnetIpv6CidrBlockAssociation, error) {
	input := &ec2.DescribeSubnetsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"ipv6-cidr-block-association.association-id": associationID,
		}),
	}

	output, err := FindSubnet(conn, input)

	if err != nil {
		return nil, err
	}

	for _, association := range output.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(association.AssociationId) == associationID {
			if state := aws.StringValue(association.Ipv6CidrBlockState.State); state == ec2.SubnetCidrBlockStateCodeDisassociated {
				return nil, &resource.NotFoundError{Message: state}
			}

			return association, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindTransitGatewayPrefixListReference(conn *ec2.EC2, transitGatewayRouteTableID string, prefixListID string) (*ec2.TransitGatewayPrefixListReference, error) {
	filters := map[string]string{
		"prefix-list-id": prefixListID,
	}

	input := &ec2.GetTransitGatewayPrefixListReferencesInput{
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
		Filters:                    BuildAttributeFilterList(filters),
	}

	var result *ec2.TransitGatewayPrefixListReference

	err := conn.GetTransitGatewayPrefixListReferencesPages(input, func(page *ec2.GetTransitGatewayPrefixListReferencesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, transitGatewayPrefixListReference := range page.TransitGatewayPrefixListReferences {
			if transitGatewayPrefixListReference == nil {
				continue
			}

			if aws.StringValue(transitGatewayPrefixListReference.PrefixListId) == prefixListID {
				result = transitGatewayPrefixListReference
				return false
			}
		}

		return !lastPage
	})

	return result, err
}

func FindTransitGatewayPrefixListReferenceByID(conn *ec2.EC2, resourceID string) (*ec2.TransitGatewayPrefixListReference, error) {
	transitGatewayRouteTableID, prefixListID, err := TransitGatewayPrefixListReferenceParseID(resourceID)

	if err != nil {
		return nil, fmt.Errorf("error parsing EC2 Transit Gateway Prefix List Reference (%s) identifier: %w", resourceID, err)
	}

	return FindTransitGatewayPrefixListReference(conn, transitGatewayRouteTableID, prefixListID)
}

func FindTransitGatewayRouteTablePropagation(conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTablePropagation, error) {
	if transitGatewayRouteTableID == "" {
		return nil, nil
	}

	input := &ec2.GetTransitGatewayRouteTablePropagationsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("transit-gateway-attachment-id"),
				Values: aws.StringSlice([]string{transitGatewayAttachmentID}),
			},
		},
		TransitGatewayRouteTableId: aws.String(transitGatewayRouteTableID),
	}

	var result *ec2.TransitGatewayRouteTablePropagation

	err := conn.GetTransitGatewayRouteTablePropagationsPages(input, func(page *ec2.GetTransitGatewayRouteTablePropagationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, transitGatewayRouteTablePropagation := range page.TransitGatewayRouteTablePropagations {
			if transitGatewayRouteTablePropagation == nil {
				continue
			}

			if aws.StringValue(transitGatewayRouteTablePropagation.TransitGatewayAttachmentId) == transitGatewayAttachmentID {
				result = transitGatewayRouteTablePropagation
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// FindVPCAttribute looks up a VPC attribute.
func FindVPCAttribute(conn *ec2.EC2, vpcID string, attribute string) (*bool, error) {
	input := &ec2.DescribeVpcAttributeInput{
		Attribute: aws.String(attribute),
		VpcId:     aws.String(vpcID),
	}

	output, err := conn.DescribeVpcAttribute(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	switch attribute {
	case ec2.VpcAttributeNameEnableDnsHostnames:
		if output.EnableDnsHostnames == nil {
			return nil, nil
		}

		return output.EnableDnsHostnames.Value, nil
	case ec2.VpcAttributeNameEnableDnsSupport:
		if output.EnableDnsSupport == nil {
			return nil, nil
		}

		return output.EnableDnsSupport.Value, nil
	}

	return nil, fmt.Errorf("unimplemented VPC attribute: %s", attribute)
}

// FindVPCByID looks up a Vpc by ID. When not found, returns nil and potentially an API error.
func FindVPCByID(conn *ec2.EC2, id string) (*ec2.Vpc, error) {
	input := &ec2.DescribeVpcsInput{
		VpcIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeVpcs(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, vpc := range output.Vpcs {
		if vpc == nil {
			continue
		}

		if aws.StringValue(vpc.VpcId) != id {
			continue
		}

		return vpc, nil
	}

	return nil, nil
}

// FindVPCEndpointByID returns the VPC endpoint corresponding to the specified identifier.
// Returns NotFoundError if no VPC endpoint is found.
func FindVPCEndpointByID(conn *ec2.EC2, vpcEndpointID string) (*ec2.VpcEndpoint, error) {
	input := &ec2.DescribeVpcEndpointsInput{
		VpcEndpointIds: aws.StringSlice([]string{vpcEndpointID}),
	}

	output, err := FindVPCEndpoint(conn, input)

	if err != nil {
		return nil, err
	}

	if state := aws.StringValue(output.State); state == VpcEndpointStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	// Eventual consistency check.
	if aws.StringValue(output.VpcEndpointId) != vpcEndpointID {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindVPCEndpoint(conn *ec2.EC2, input *ec2.DescribeVpcEndpointsInput) (*ec2.VpcEndpoint, error) {
	output, err := conn.DescribeVpcEndpoints(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidVpcEndpointIdNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpcEndpoints) == 0 || output.VpcEndpoints[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VpcEndpoints[0], nil
}

// FindVPCEndpointRouteTableAssociationExists returns NotFoundError if no association for the specified VPC endpoint and route table IDs is found.
func FindVPCEndpointRouteTableAssociationExists(conn *ec2.EC2, vpcEndpointID string, routeTableID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointRouteTableID := range vpcEndpoint.RouteTableIds {
		if aws.StringValue(vpcEndpointRouteTableID) == routeTableID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint Route Table Association (%s/%s) not found", vpcEndpointID, routeTableID),
	}
}

// FindVPCEndpointSubnetAssociationExists returns NotFoundError if no association for the specified VPC endpoint and subnet IDs is found.
func FindVPCEndpointSubnetAssociationExists(conn *ec2.EC2, vpcEndpointID string, subnetID string) error {
	vpcEndpoint, err := FindVPCEndpointByID(conn, vpcEndpointID)

	if err != nil {
		return err
	}

	for _, vpcEndpointSubnetID := range vpcEndpoint.SubnetIds {
		if aws.StringValue(vpcEndpointSubnetID) == subnetID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("VPC Endpoint (%s) Subnet (%s) Association not found", vpcEndpointID, subnetID),
	}
}

// FindVPCPeeringConnectionByID returns the VPC peering connection corresponding to the specified identifier.
// Returns nil and potentially an error if no VPC peering connection is found.
func FindVPCPeeringConnectionByID(conn *ec2.EC2, id string) (*ec2.VpcPeeringConnection, error) {
	input := &ec2.DescribeVpcPeeringConnectionsInput{
		VpcPeeringConnectionIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeVpcPeeringConnections(input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.VpcPeeringConnections) == 0 {
		return nil, nil
	}

	return output.VpcPeeringConnections[0], nil
}

// FindVPNGatewayRoutePropagationExists returns NotFoundError if no route propagation for the specified VPN gateway is found.
func FindVPNGatewayRoutePropagationExists(conn *ec2.EC2, routeTableID, gatewayID string) error {
	routeTable, err := FindRouteTableByID(conn, routeTableID)

	if err != nil {
		return err
	}

	for _, v := range routeTable.PropagatingVgws {
		if aws.StringValue(v.GatewayId) == gatewayID {
			return nil
		}
	}

	return &resource.NotFoundError{
		LastError: fmt.Errorf("Route Table (%s) VPN Gateway (%s) route propagation not found", routeTableID, gatewayID),
	}
}

// FindVPNGatewayVPCAttachment returns the attachment between the specified VPN gateway and VPC.
// Returns nil and potentially an error if no attachment is found.
func FindVPNGatewayVPCAttachment(conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	vpnGateway, err := FindVPNGatewayByID(conn, vpnGatewayID)
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

// FindVPNGatewayByID returns the VPN gateway corresponding to the specified identifier.
// Returns nil and potentially an error if no VPN gateway is found.
func FindVPNGatewayByID(conn *ec2.EC2, id string) (*ec2.VpnGateway, error) {
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

func FindFlowLogByID(conn *ec2.EC2, id string) (*ec2.FlowLog, error) {
	input := &ec2.DescribeFlowLogsInput{
		FlowLogIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeFlowLogs(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.FlowLogs) == 0 || output.FlowLogs[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.FlowLogs); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.FlowLogs[0], nil
}

func FindInternetGateway(conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) (*ec2.InternetGateway, error) {
	output, err := FindInternetGateways(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindInternetGateways(conn *ec2.EC2, input *ec2.DescribeInternetGatewaysInput) ([]*ec2.InternetGateway, error) {
	var output []*ec2.InternetGateway

	err := conn.DescribeInternetGatewaysPages(input, func(page *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InternetGateways {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidInternetGatewayIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindInternetGatewayByID(conn *ec2.EC2, id string) (*ec2.InternetGateway, error) {
	input := &ec2.DescribeInternetGatewaysInput{
		InternetGatewayIds: aws.StringSlice([]string{id}),
	}

	output, err := FindInternetGateway(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.InternetGatewayId) != id {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindInternetGatewayAttachment(conn *ec2.EC2, internetGatewayID, vpcID string) (*ec2.InternetGatewayAttachment, error) {
	internetGateway, err := FindInternetGatewayByID(conn, internetGatewayID)

	if err != nil {
		return nil, err
	}

	if len(internetGateway.Attachments) == 0 || internetGateway.Attachments[0] == nil {
		return nil, tfresource.NewEmptyResultError(internetGatewayID)
	}

	if count := len(internetGateway.Attachments); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, internetGatewayID)
	}

	attachment := internetGateway.Attachments[0]

	if aws.StringValue(attachment.VpcId) != vpcID {
		return nil, tfresource.NewEmptyResultError(vpcID)
	}

	return attachment, nil
}

func FindKeyPair(conn *ec2.EC2, input *ec2.DescribeKeyPairsInput) (*ec2.KeyPairInfo, error) {
	output, err := FindKeyPairs(conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func FindKeyPairs(conn *ec2.EC2, input *ec2.DescribeKeyPairsInput) ([]*ec2.KeyPairInfo, error) {
	output, err := conn.DescribeKeyPairs(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidKeyPairNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output.KeyPairs, nil
}

func FindKeyPairByName(conn *ec2.EC2, name string) (*ec2.KeyPairInfo, error) {
	input := &ec2.DescribeKeyPairsInput{
		KeyNames: aws.StringSlice([]string{name}),
	}

	output, err := FindKeyPair(conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.KeyName) != name {
		return nil, &resource.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindManagedPrefixListByID(conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	input := &ec2.DescribeManagedPrefixListsInput{
		PrefixListIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeManagedPrefixLists(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPrefixListIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PrefixLists) == 0 || output.PrefixLists[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PrefixLists); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	prefixList := output.PrefixLists[0]

	if state := aws.StringValue(prefixList.State); state == ec2.PrefixListStateDeleteComplete {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return prefixList, nil
}

func FindManagedPrefixListEntriesByID(conn *ec2.EC2, id string) ([]*ec2.PrefixListEntry, error) {
	input := &ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: aws.String(id),
	}

	var prefixListEntries []*ec2.PrefixListEntry

	err := conn.GetManagedPrefixListEntriesPages(input, func(page *ec2.GetManagedPrefixListEntriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, entry := range page.Entries {
			if entry == nil {
				continue
			}

			prefixListEntries = append(prefixListEntries, entry)
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPrefixListIDNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return prefixListEntries, nil
}

func FindManagedPrefixListEntryByIDAndCIDR(conn *ec2.EC2, id, cidr string) (*ec2.PrefixListEntry, error) {
	prefixListEntries, err := FindManagedPrefixListEntriesByID(conn, id)

	if err != nil {
		return nil, err
	}

	for _, entry := range prefixListEntries {
		if aws.StringValue(entry.Cidr) == cidr {
			return entry, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func FindPlacementGroupByName(conn *ec2.EC2, name string) (*ec2.PlacementGroup, error) {
	input := &ec2.DescribePlacementGroupsInput{
		GroupNames: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribePlacementGroups(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidPlacementGroupUnknown) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.PlacementGroups) == 0 || output.PlacementGroups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.PlacementGroups); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	placementGroup := output.PlacementGroups[0]

	if state := aws.StringValue(placementGroup.State); state == ec2.PlacementGroupStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return placementGroup, nil
}

func FindVPCEndpointConnectionByServiceIDAndVPCEndpointID(conn *ec2.EC2, serviceID, vpcEndpointID string) (*ec2.VpcEndpointConnection, error) {
	input := &ec2.DescribeVpcEndpointConnectionsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"service-id": serviceID,
			// "InvalidFilter: The filter vpc-endpoint-id  is invalid"
			// "vpc-endpoint-id ": vpcEndpointID,
		}),
	}

	var output *ec2.VpcEndpointConnection

	err := conn.DescribeVpcEndpointConnectionsPages(input, func(page *ec2.DescribeVpcEndpointConnectionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.VpcEndpointConnections {
			if aws.StringValue(v.VpcEndpointId) == vpcEndpointID {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if vpcEndpointState := aws.StringValue(output.VpcEndpointState); vpcEndpointState == VpcEndpointStateDeleted {
		return nil, &resource.NotFoundError{
			Message:     vpcEndpointState,
			LastRequest: input,
		}
	}

	return output, nil
}
