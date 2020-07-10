package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const (
	// Maximum amount of time to wait for a LocalGatewayRouteTableVpcAssociation to return Associated
	LocalGatewayRouteTableVpcAssociationAssociatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a LocalGatewayRouteTableVpcAssociation to return Disassociated
	LocalGatewayRouteTableVpcAssociationDisassociatedTimeout = 5 * time.Minute
)

// LocalGatewayRouteTableVpcAssociationAssociated waits for a LocalGatewayRouteTableVpcAssociation to return Associated
func LocalGatewayRouteTableVpcAssociationAssociated(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.RouteTableAssociationStateCodeAssociating},
		Target:  []string{ec2.RouteTableAssociationStateCodeAssociated},
		Refresh: LocalGatewayRouteTableVpcAssociationState(conn, localGatewayRouteTableVpcAssociationID),
		Timeout: LocalGatewayRouteTableVpcAssociationAssociatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.LocalGatewayRouteTableVpcAssociation); ok {
		return output, err
	}

	return nil, err
}

// LocalGatewayRouteTableVpcAssociationDisassociated waits for a LocalGatewayRouteTableVpcAssociation to return Disassociated
func LocalGatewayRouteTableVpcAssociationDisassociated(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.RouteTableAssociationStateCodeDisassociating},
		Target:  []string{ec2.RouteTableAssociationStateCodeDisassociated},
		Refresh: LocalGatewayRouteTableVpcAssociationState(conn, localGatewayRouteTableVpcAssociationID),
		Timeout: LocalGatewayRouteTableVpcAssociationAssociatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.LocalGatewayRouteTableVpcAssociation); ok {
		return output, err
	}

	return nil, err
}

const (
	ClientVpnEndpointDeletedTimout = 5 * time.Minute
)

func ClientVpnEndpointDeleted(conn *ec2.EC2, id string) (*ec2.ClientVpnEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnEndpointStatusCodeDeleting},
		Target:  []string{},
		Refresh: ClientVpnEndpointStatus(conn, id),
		Timeout: ClientVpnEndpointDeletedTimout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ClientVpnEndpoint); ok {
		return output, err
	}

	return nil, err
}

const (
	ClientVpnAuthorizationRuleActiveTimeout = 1 * time.Minute

	ClientVpnAuthorizationRuleRevokedTimeout = 1 * time.Minute
)

func ClientVpnAuthorizationRuleAuthorized(conn *ec2.EC2, authorizationRuleID string) (*ec2.AuthorizationRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeAuthorizing},
		Target:  []string{ec2.ClientVpnAuthorizationRuleStatusCodeActive},
		Refresh: ClientVpnAuthorizationRuleStatus(conn, authorizationRuleID),
		Timeout: ClientVpnAuthorizationRuleActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.AuthorizationRule); ok {
		return output, err
	}

	return nil, err
}

func ClientVpnAuthorizationRuleRevoked(conn *ec2.EC2, authorizationRuleID string) (*ec2.AuthorizationRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeRevoking},
		Target:  []string{},
		Refresh: ClientVpnAuthorizationRuleStatus(conn, authorizationRuleID),
		Timeout: ClientVpnAuthorizationRuleRevokedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.AuthorizationRule); ok {
		return output, err
	}

	return nil, err
}

const (
	ClientVpnRouteDeletedTimeout = 1 * time.Minute
)

func ClientVpnRouteDeleted(conn *ec2.EC2, routeID string) (*ec2.ClientVpnRoute, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnRouteStatusCodeActive, ec2.ClientVpnRouteStatusCodeDeleting},
		Target:  []string{},
		Refresh: ClientVpnRouteStatus(conn, routeID),
		Timeout: ClientVpnRouteDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ClientVpnRoute); ok {
		return output, err
	}

	return nil, err
}

const (
	RouteTableCreatedTimeout = 10 * time.Minute

	RouteTableDeletedTimeout = 5 * time.Minute
)

func RouteTableCreated(conn *ec2.EC2, routeTableID string) (*ec2.RouteTable, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{},
		Target:         []string{RouteTableStateReady},
		Refresh:        RouteTableState(conn, routeTableID),
		Timeout:        RouteTableCreatedTimeout,
		NotFoundChecks: 40,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func RouteTableDeleted(conn *ec2.EC2, routeTableID string) (*ec2.RouteTable, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{RouteTableStateReady},
		Target:  []string{},
		Refresh: RouteTableState(conn, routeTableID),
		Timeout: RouteTableDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.RouteTable); ok {
		return output, err
	}

	return nil, err
}
