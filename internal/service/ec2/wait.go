package ec2

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for EC2 Instance attribute modifications to propagate
	InstanceAttributePropagationTimeout = 2 * time.Minute

	InstanceStopTimeout = 10 * time.Minute

	// General timeout for EC2 resource creations to propagate
	PropagationTimeout = 2 * time.Minute

	RouteTableNotFoundChecks                   = 1000 // Should exceed any reasonable custom timeout value.
	RouteTableAssociationCreatedNotFoundChecks = 1000 // Should exceed any reasonable custom timeout value.
)

const (
	CarrierGatewayAvailableTimeout = 5 * time.Minute

	CarrierGatewayDeletedTimeout = 5 * time.Minute
)

func WaitCarrierGatewayAvailable(conn *ec2.EC2, carrierGatewayID string) (*ec2.CarrierGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.CarrierGatewayStatePending},
		Target:  []string{ec2.CarrierGatewayStateAvailable},
		Refresh: StatusCarrierGatewayState(conn, carrierGatewayID),
		Timeout: CarrierGatewayAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.CarrierGateway); ok {
		return output, err
	}

	return nil, err
}

func WaitCarrierGatewayDeleted(conn *ec2.EC2, carrierGatewayID string) (*ec2.CarrierGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.CarrierGatewayStateDeleting},
		Target:  []string{},
		Refresh: StatusCarrierGatewayState(conn, carrierGatewayID),
		Timeout: CarrierGatewayDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.CarrierGateway); ok {
		return output, err
	}

	return nil, err
}

const (
	// Maximum amount of time to wait for a LocalGatewayRouteTableVpcAssociation to return Associated
	LocalGatewayRouteTableVPCAssociationAssociatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a LocalGatewayRouteTableVpcAssociation to return Disassociated
	LocalGatewayRouteTableVPCAssociationDisassociatedTimeout = 5 * time.Minute
)

// WaitLocalGatewayRouteTableVPCAssociationAssociated waits for a LocalGatewayRouteTableVpcAssociation to return Associated
func WaitLocalGatewayRouteTableVPCAssociationAssociated(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.RouteTableAssociationStateCodeAssociating},
		Target:  []string{ec2.RouteTableAssociationStateCodeAssociated},
		Refresh: StatusLocalGatewayRouteTableVPCAssociationState(conn, localGatewayRouteTableVpcAssociationID),
		Timeout: LocalGatewayRouteTableVPCAssociationAssociatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.LocalGatewayRouteTableVpcAssociation); ok {
		return output, err
	}

	return nil, err
}

// WaitLocalGatewayRouteTableVPCAssociationDisassociated waits for a LocalGatewayRouteTableVpcAssociation to return Disassociated
func WaitLocalGatewayRouteTableVPCAssociationDisassociated(conn *ec2.EC2, localGatewayRouteTableVpcAssociationID string) (*ec2.LocalGatewayRouteTableVpcAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.RouteTableAssociationStateCodeDisassociating},
		Target:  []string{ec2.RouteTableAssociationStateCodeDisassociated},
		Refresh: StatusLocalGatewayRouteTableVPCAssociationState(conn, localGatewayRouteTableVpcAssociationID),
		Timeout: LocalGatewayRouteTableVPCAssociationAssociatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.LocalGatewayRouteTableVpcAssociation); ok {
		return output, err
	}

	return nil, err
}

const (
	ClientVPNEndpointDeletedTimout = 5 * time.Minute
)

func WaitClientVPNEndpointDeleted(conn *ec2.EC2, id string) (*ec2.ClientVpnEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnEndpointStatusCodeDeleting},
		Target:  []string{},
		Refresh: StatusClientVPNEndpoint(conn, id),
		Timeout: ClientVPNEndpointDeletedTimout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ClientVpnEndpoint); ok {
		return output, err
	}

	return nil, err
}

const (
	ClientVPNAuthorizationRuleActiveTimeout = 10 * time.Minute

	ClientVPNAuthorizationRuleRevokedTimeout = 10 * time.Minute
)

func WaitClientVPNAuthorizationRuleAuthorized(conn *ec2.EC2, authorizationRuleID string) (*ec2.AuthorizationRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeAuthorizing},
		Target:  []string{ec2.ClientVpnAuthorizationRuleStatusCodeActive},
		Refresh: StatusClientVPNAuthorizationRule(conn, authorizationRuleID),
		Timeout: ClientVPNAuthorizationRuleActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.AuthorizationRule); ok {
		return output, err
	}

	return nil, err
}

func WaitClientVPNAuthorizationRuleRevoked(conn *ec2.EC2, authorizationRuleID string) (*ec2.AuthorizationRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeRevoking},
		Target:  []string{},
		Refresh: StatusClientVPNAuthorizationRule(conn, authorizationRuleID),
		Timeout: ClientVPNAuthorizationRuleRevokedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.AuthorizationRule); ok {
		return output, err
	}

	return nil, err
}

const (
	ClientVPNNetworkAssociationAssociatedTimeout = 30 * time.Minute

	ClientVPNNetworkAssociationAssociatedDelay = 4 * time.Minute

	ClientVPNNetworkAssociationDisassociatedTimeout = 30 * time.Minute

	ClientVPNNetworkAssociationDisassociatedDelay = 4 * time.Minute

	ClientVPNNetworkAssociationStatusPollInterval = 10 * time.Second
)

func WaitClientVPNNetworkAssociationAssociated(conn *ec2.EC2, networkAssociationID, clientVpnEndpointID string) (*ec2.TargetNetwork, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{ec2.AssociationStatusCodeAssociating},
		Target:       []string{ec2.AssociationStatusCodeAssociated},
		Refresh:      StatusClientVPNNetworkAssociation(conn, networkAssociationID, clientVpnEndpointID),
		Timeout:      ClientVPNNetworkAssociationAssociatedTimeout,
		Delay:        ClientVPNNetworkAssociationAssociatedDelay,
		PollInterval: ClientVPNNetworkAssociationStatusPollInterval,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TargetNetwork); ok {
		return output, err
	}

	return nil, err
}

func WaitClientVPNNetworkAssociationDisassociated(conn *ec2.EC2, networkAssociationID, clientVpnEndpointID string) (*ec2.TargetNetwork, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{ec2.AssociationStatusCodeDisassociating},
		Target:       []string{},
		Refresh:      StatusClientVPNNetworkAssociation(conn, networkAssociationID, clientVpnEndpointID),
		Timeout:      ClientVPNNetworkAssociationDisassociatedTimeout,
		Delay:        ClientVPNNetworkAssociationDisassociatedDelay,
		PollInterval: ClientVPNNetworkAssociationStatusPollInterval,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TargetNetwork); ok {
		return output, err
	}

	return nil, err
}

const (
	ClientVPNRouteDeletedTimeout = 1 * time.Minute
)

func WaitClientVPNRouteDeleted(conn *ec2.EC2, routeID string) (*ec2.ClientVpnRoute, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnRouteStatusCodeActive, ec2.ClientVpnRouteStatusCodeDeleting},
		Target:  []string{},
		Refresh: StatusClientVPNRoute(conn, routeID),
		Timeout: ClientVPNRouteDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ClientVpnRoute); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceIAMInstanceProfileUpdated(conn *ec2.EC2, instanceID string, expectedValue string) (*ec2.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{expectedValue},
		Refresh:    StatusInstanceIAMInstanceProfile(conn, instanceID),
		Timeout:    InstanceAttributePropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Instance); ok {
		return output, err
	}

	return nil, err
}

const ManagedPrefixListEntryCreateTimeout = 5 * time.Minute

const (
	NetworkACLPropagationTimeout      = 2 * time.Minute
	NetworkACLEntryPropagationTimeout = 5 * time.Minute
)

func WaitRouteDeleted(conn *ec2.EC2, routeFinder RouteFinder, routeTableID, destination string, timeout time.Duration) (*ec2.Route, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{RouteStatusReady},
		Target:                    []string{},
		Refresh:                   StatusRoute(conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Route); ok {
		return output, err
	}

	return nil, err
}

func WaitRouteReady(conn *ec2.EC2, routeFinder RouteFinder, routeTableID, destination string, timeout time.Duration) (*ec2.Route, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{RouteStatusReady},
		Refresh:                   StatusRoute(conn, routeFinder, routeTableID, destination),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Route); ok {
		return output, err
	}

	return nil, err
}

const (
	RouteTableAssociationPropagationTimeout = 5 * time.Minute

	RouteTableAssociationCreatedTimeout = 5 * time.Minute
	RouteTableAssociationUpdatedTimeout = 5 * time.Minute
	RouteTableAssociationDeletedTimeout = 5 * time.Minute
)

func WaitRouteTableReady(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.RouteTable, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{},
		Target:         []string{RouteTableStatusReady},
		Refresh:        StatusRouteTable(conn, id),
		Timeout:        timeout,
		NotFoundChecks: RouteTableNotFoundChecks,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func WaitRouteTableDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.RouteTable, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{RouteTableStatusReady},
		Target:  []string{},
		Refresh: StatusRouteTable(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.RouteTable); ok {
		return output, err
	}

	return nil, err
}

func WaitRouteTableAssociationCreated(conn *ec2.EC2, id string) (*ec2.RouteTableAssociationState, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{ec2.RouteTableAssociationStateCodeAssociating},
		Target:         []string{ec2.RouteTableAssociationStateCodeAssociated},
		Refresh:        StatusRouteTableAssociationState(conn, id),
		Timeout:        RouteTableAssociationCreatedTimeout,
		NotFoundChecks: RouteTableAssociationCreatedNotFoundChecks,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.RouteTableAssociationState); ok {
		if state := aws.StringValue(output.State); state == ec2.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitRouteTableAssociationDeleted(conn *ec2.EC2, id string) (*ec2.RouteTableAssociationState, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.RouteTableAssociationStateCodeDisassociating, ec2.RouteTableAssociationStateCodeAssociated},
		Target:  []string{},
		Refresh: StatusRouteTableAssociationState(conn, id),
		Timeout: RouteTableAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.RouteTableAssociationState); ok {
		if state := aws.StringValue(output.State); state == ec2.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitRouteTableAssociationUpdated(conn *ec2.EC2, id string) (*ec2.RouteTableAssociationState, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.RouteTableAssociationStateCodeAssociating},
		Target:  []string{ec2.RouteTableAssociationStateCodeAssociated},
		Refresh: StatusRouteTableAssociationState(conn, id),
		Timeout: RouteTableAssociationUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.RouteTableAssociationState); ok {
		if state := aws.StringValue(output.State); state == ec2.RouteTableAssociationStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitSecurityGroupCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SecurityGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{SecurityGroupStatusNotFound},
		Target:  []string{SecurityGroupStatusCreated},
		Refresh: StatusSecurityGroup(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.SecurityGroup); ok {
		return output, err
	}

	return nil, err
}

const (
	SubnetPropagationTimeout          = 2 * time.Minute
	SubnetAttributePropagationTimeout = 5 * time.Minute
)

func WaitSubnetMapCustomerOwnedIPOnLaunchUpdated(conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetMapCustomerOwnedIPOnLaunch(conn, subnetID),
		Timeout:    SubnetAttributePropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetMapPublicIPOnLaunchUpdated(conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetMapPublicIPOnLaunch(conn, subnetID),
		Timeout:    SubnetAttributePropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

const (
	TransitGatewayPrefixListReferenceTimeout = 5 * time.Minute
)

func WaitTransitGatewayPrefixListReferenceStateCreated(conn *ec2.EC2, transitGatewayRouteTableID string, prefixListID string) (*ec2.TransitGatewayPrefixListReference, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayPrefixListReferenceStatePending},
		Target:  []string{ec2.TransitGatewayPrefixListReferenceStateAvailable},
		Timeout: TransitGatewayPrefixListReferenceTimeout,
		Refresh: StatusTransitGatewayPrefixListReferenceState(conn, transitGatewayRouteTableID, prefixListID),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayPrefixListReference); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayPrefixListReferenceStateDeleted(conn *ec2.EC2, transitGatewayRouteTableID string, prefixListID string) (*ec2.TransitGatewayPrefixListReference, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayPrefixListReferenceStateDeleting},
		Target:  []string{},
		Timeout: TransitGatewayPrefixListReferenceTimeout,
		Refresh: StatusTransitGatewayPrefixListReferenceState(conn, transitGatewayRouteTableID, prefixListID),
	}

	outputRaw, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		return nil, nil
	}

	if output, ok := outputRaw.(*ec2.TransitGatewayPrefixListReference); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayPrefixListReferenceStateUpdated(conn *ec2.EC2, transitGatewayRouteTableID string, prefixListID string) (*ec2.TransitGatewayPrefixListReference, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayPrefixListReferenceStateModifying},
		Target:  []string{ec2.TransitGatewayPrefixListReferenceStateAvailable},
		Timeout: TransitGatewayPrefixListReferenceTimeout,
		Refresh: StatusTransitGatewayPrefixListReferenceState(conn, transitGatewayRouteTableID, prefixListID),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayPrefixListReference); ok {
		return output, err
	}

	return nil, err
}

const (
	TransitGatewayRouteTablePropagationTimeout = 5 * time.Minute
)

func WaitTransitGatewayRouteTablePropagationStateEnabled(conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTablePropagation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayPropagationStateEnabling},
		Target:  []string{ec2.TransitGatewayPropagationStateEnabled},
		Timeout: TransitGatewayRouteTablePropagationTimeout,
		Refresh: StatusTransitGatewayRouteTablePropagationState(conn, transitGatewayRouteTableID, transitGatewayAttachmentID),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayRouteTablePropagation); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayRouteTablePropagationStateDisabled(conn *ec2.EC2, transitGatewayRouteTableID string, transitGatewayAttachmentID string) (*ec2.TransitGatewayRouteTablePropagation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayPropagationStateDisabling},
		Target:  []string{},
		Timeout: TransitGatewayRouteTablePropagationTimeout,
		Refresh: StatusTransitGatewayRouteTablePropagationState(conn, transitGatewayRouteTableID, transitGatewayAttachmentID),
	}

	outputRaw, err := stateConf.WaitForState()

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidRouteTableIDNotFound) {
		return nil, nil
	}

	if output, ok := outputRaw.(*ec2.TransitGatewayRouteTablePropagation); ok {
		return output, err
	}

	return nil, err
}

const (
	VPCPropagationTimeout          = 2 * time.Minute
	VPCAttributePropagationTimeout = 5 * time.Minute
)

func WaitVPCAttributeUpdated(conn *ec2.EC2, vpcID string, attribute string, expectedValue bool) (*ec2.Vpc, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusVPCAttribute(conn, vpcID, attribute),
		Timeout:    VPCAttributePropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Vpc); ok {
		return output, err
	}

	return nil, err
}

const (
	VPNGatewayVPCAttachmentAttachedTimeout = 15 * time.Minute

	VPNGatewayVPCAttachmentDetachedTimeout = 30 * time.Minute
)

func WaitVPNGatewayVPCAttachmentAttached(conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusDetached, ec2.AttachmentStatusAttaching},
		Target:  []string{ec2.AttachmentStatusAttached},
		Refresh: StatusVPNGatewayVPCAttachmentState(conn, vpnGatewayID, vpcID),
		Timeout: VPNGatewayVPCAttachmentAttachedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNGatewayVPCAttachmentDetached(conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusAttached, ec2.AttachmentStatusDetaching},
		Target:  []string{ec2.AttachmentStatusDetached},
		Refresh: StatusVPNGatewayVPCAttachmentState(conn, vpnGatewayID, vpcID),
		Timeout: VPNGatewayVPCAttachmentDetachedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcAttachment); ok {
		return output, err
	}

	return nil, err
}

const (
	HostCreatedTimeout = 10 * time.Minute
	HostUpdatedTimeout = 10 * time.Minute
	HostDeletedTimeout = 20 * time.Minute
)

func WaitHostCreated(conn *ec2.EC2, id string) (*ec2.Host, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AllocationStatePending},
		Target:  []string{ec2.AllocationStateAvailable},
		Timeout: HostCreatedTimeout,
		Refresh: StatusHostState(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Host); ok {
		return output, err
	}

	return nil, err
}

func WaitHostUpdated(conn *ec2.EC2, id string) (*ec2.Host, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AllocationStatePending},
		Target:  []string{ec2.AllocationStateAvailable},
		Timeout: HostUpdatedTimeout,
		Refresh: StatusHostState(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Host); ok {
		return output, err
	}

	return nil, err
}

func WaitHostDeleted(conn *ec2.EC2, id string) (*ec2.Host, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AllocationStateAvailable},
		Target:  []string{},
		Timeout: HostDeletedTimeout,
		Refresh: StatusHostState(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Host); ok {
		return output, err
	}

	return nil, err
}

const (
	internetGatewayAttachedTimeout = 4 * time.Minute
	internetGatewayDeletedTimeout  = 10 * time.Minute
	internetGatewayDetachedTimeout = 15 * time.Minute
)

func WaitInternetGatewayAttached(conn *ec2.EC2, internetGatewayID, vpcID string, timeout time.Duration) (*ec2.InternetGatewayAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusAttaching},
		Target:  []string{InternetGatewayAttachmentStateAvailable},
		Timeout: timeout,
		Refresh: StatusInternetGatewayAttachmentState(conn, internetGatewayID, vpcID),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.InternetGatewayAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitInternetGatewayDetached(conn *ec2.EC2, internetGatewayID, vpcID string, timeout time.Duration) (*ec2.InternetGatewayAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusDetaching},
		Target:  []string{},
		Timeout: timeout,
		Refresh: StatusInternetGatewayAttachmentState(conn, internetGatewayID, vpcID),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.InternetGatewayAttachment); ok {
		return output, err
	}

	return nil, err
}

const (
	ManagedPrefixListTimeout = 15 * time.Minute
)

func WaitManagedPrefixListCreated(conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.PrefixListStateCreateInProgress},
		Target:  []string{ec2.PrefixListStateCreateComplete},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ManagedPrefixList); ok {
		if state := aws.StringValue(output.State); state == ec2.PrefixListStateCreateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitManagedPrefixListModified(conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.PrefixListStateModifyInProgress},
		Target:  []string{ec2.PrefixListStateModifyComplete},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ManagedPrefixList); ok {
		if state := aws.StringValue(output.State); state == ec2.PrefixListStateModifyFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitManagedPrefixListDeleted(conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.PrefixListStateDeleteInProgress},
		Target:  []string{},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ManagedPrefixList); ok {
		if state := aws.StringValue(output.State); state == ec2.PrefixListStateDeleteFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

const (
	networkInterfaceAttachedTimeout = 5 * time.Minute
	NetworkInterfaceDetachedTimeout = 10 * time.Minute
)

func WaitNetworkInterfaceAttached(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterfaceAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusAttaching},
		Target:  []string{ec2.AttachmentStatusAttached},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceAttachmentStatus(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceAvailableAfterUse(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterface, error) {
	// Hyperplane attached ENI.
	// Wait for it to be moved into a removable state.
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.NetworkInterfaceStatusInUse},
		Target:     []string{ec2.NetworkInterfaceStatusAvailable},
		Timeout:    timeout,
		Refresh:    StatusNetworkInterfaceStatus(conn, id),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		// Handle EC2 ENI eventual consistency. It can take up to 3 minutes.
		ContinuousTargetOccurence: 18,
		NotFoundChecks:            1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterface, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{NetworkInterfaceStatusPending},
		Target:  []string{ec2.NetworkInterfaceStatusAvailable},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceStatus(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceDetached(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterfaceAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusDetaching},
		Target:  []string{ec2.AttachmentStatusDetached},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceAttachmentStatus(conn, id),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

const (
	PlacementGroupCreatedTimeout = 5 * time.Minute
	PlacementGroupDeletedTimeout = 5 * time.Minute
)

func WaitPlacementGroupCreated(conn *ec2.EC2, name string) (*ec2.PlacementGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.PlacementGroupStatePending},
		Target:  []string{ec2.PlacementGroupStateAvailable},
		Timeout: PlacementGroupCreatedTimeout,
		Refresh: StatusPlacementGroupState(conn, name),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.PlacementGroup); ok {
		return output, err
	}

	return nil, err
}

func WaitPlacementGroupDeleted(conn *ec2.EC2, name string) (*ec2.PlacementGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.PlacementGroupStateDeleting},
		Target:  []string{},
		Timeout: PlacementGroupDeletedTimeout,
		Refresh: StatusPlacementGroupState(conn, name),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.PlacementGroup); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCEndpointAccepted(conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{VPCEndpointStatePendingAcceptance},
		Target:     []string{VPCEndpointStateAvailable},
		Timeout:    timeout,
		Refresh:    StatusVPCEndpointState(conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcEndpoint); ok {
		if state, lastError := aws.StringValue(output.State), output.LastError; state == VPCEndpointStateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(lastError.Code), aws.StringValue(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCEndpointAvailable(conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{VPCEndpointStatePending},
		Target:     []string{VPCEndpointStateAvailable, VPCEndpointStatePendingAcceptance},
		Timeout:    timeout,
		Refresh:    StatusVPCEndpointState(conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcEndpoint); ok {
		if state, lastError := aws.StringValue(output.State), output.LastError; state == VPCEndpointStateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(lastError.Code), aws.StringValue(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCEndpointDeleted(conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{VPCEndpointStateDeleting},
		Target:     []string{},
		Timeout:    timeout,
		Refresh:    StatusVPCEndpointState(conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcEndpoint); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCEndpointRouteTableAssociationDeleted(conn *ec2.EC2, vpcEndpointID, routeTableID string) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{VPCEndpointRouteTableAssociationStatusReady},
		Target:                    []string{},
		Refresh:                   StatusVPCEndpointRouteTableAssociation(conn, vpcEndpointID, routeTableID),
		Timeout:                   PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitVPCEndpointRouteTableAssociationReady(conn *ec2.EC2, vpcEndpointID, routeTableID string) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{VPCEndpointRouteTableAssociationStatusReady},
		Refresh:                   StatusVPCEndpointRouteTableAssociation(conn, vpcEndpointID, routeTableID),
		Timeout:                   PropagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitEBSSnapshotImportComplete(conn *ec2.EC2, importTaskID string) (*ec2.SnapshotTaskDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{EBSSnapshotImportStateActive,
			EBSSnapshotImportStateUpdating,
			EBSSnapshotImportStateValidating,
			EBSSnapshotImportStateValidd,
			EBSSnapshotImportStateConverting,
		},
		Target:  []string{EBSSnapshotImportStateCompleted},
		Refresh: StatusEBSSnapshotImport(conn, importTaskID),
		Timeout: 60 * time.Minute,
		Delay:   10 * time.Second,
	}

	detail, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	} else {
		return detail.(*ec2.SnapshotTaskDetail), nil
	}
}
