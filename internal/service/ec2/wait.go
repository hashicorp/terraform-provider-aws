package ec2

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	InstanceReadyTimeout = 10 * time.Minute
	InstanceStartTimeout = 10 * time.Minute
	InstanceStopTimeout  = 10 * time.Minute

	// General timeout for EC2 resource creations to propagate.
	// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/query-api-troubleshooting.html#eventual-consistency.
	propagationTimeout = 2 * time.Minute

	RouteNotFoundChecks                        = 1000 // Should exceed any reasonable custom timeout value.
	RouteTableNotFoundChecks                   = 1000 // Should exceed any reasonable custom timeout value.
	RouteTableAssociationCreatedNotFoundChecks = 1000 // Should exceed any reasonable custom timeout value.
	SecurityGroupNotFoundChecks                = 1000 // Should exceed any reasonable custom timeout value.
	InternetGatewayNotFoundChecks              = 1000 // Should exceed any reasonable custom timeout value.
)

const (
	AvailabilityZoneGroupOptInStatusTimeout = 10 * time.Minute
)

func WaitAvailabilityZoneGroupOptedIn(conn *ec2.EC2, name string) (*ec2.AvailabilityZone, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AvailabilityZoneOptInStatusNotOptedIn},
		Target:  []string{ec2.AvailabilityZoneOptInStatusOptedIn},
		Refresh: StatusAvailabilityZoneGroupOptInStatus(conn, name),
		Timeout: AvailabilityZoneGroupOptInStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.AvailabilityZone); ok {
		return output, err
	}

	return nil, err
}

func WaitAvailabilityZoneGroupNotOptedIn(conn *ec2.EC2, name string) (*ec2.AvailabilityZone, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AvailabilityZoneOptInStatusOptedIn},
		Target:  []string{ec2.AvailabilityZoneOptInStatusNotOptedIn},
		Refresh: StatusAvailabilityZoneGroupOptInStatus(conn, name),
		Timeout: AvailabilityZoneGroupOptInStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.AvailabilityZone); ok {
		return output, err
	}

	return nil, err
}

const (
	CapacityReservationActiveTimeout  = 2 * time.Minute
	CapacityReservationDeletedTimeout = 2 * time.Minute
)

func WaitCapacityReservationActive(conn *ec2.EC2, id string) (*ec2.CapacityReservation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.CapacityReservationStatePending},
		Target:  []string{ec2.CapacityReservationStateActive},
		Refresh: StatusCapacityReservationState(conn, id),
		Timeout: CapacityReservationActiveTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.CapacityReservation); ok {
		return output, err
	}

	return nil, err
}

func WaitCapacityReservationDeleted(conn *ec2.EC2, id string) (*ec2.CapacityReservation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.CapacityReservationStateActive},
		Target:  []string{},
		Refresh: StatusCapacityReservationState(conn, id),
		Timeout: CapacityReservationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.CapacityReservation); ok {
		return output, err
	}

	return nil, err
}

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
	ClientVPNEndpointDeletedTimeout          = 5 * time.Minute
	ClientVPNEndpointAttributeUpdatedTimeout = 5 * time.Minute
)

func WaitClientVPNEndpointDeleted(conn *ec2.EC2, id string) (*ec2.ClientVpnEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnEndpointStatusCodeDeleting},
		Target:  []string{},
		Refresh: StatusClientVPNEndpointState(conn, id),
		Timeout: ClientVPNEndpointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ClientVpnEndpoint); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNEndpointClientConnectResponseOptionsUpdated(conn *ec2.EC2, id string) (*ec2.ClientConnectResponseOptions, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnEndpointAttributeStatusCodeApplying},
		Target:  []string{ec2.ClientVpnEndpointAttributeStatusCodeApplied},
		Refresh: StatusClientVPNEndpointClientConnectResponseOptionsState(conn, id),
		Timeout: ClientVPNEndpointAttributeUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ClientConnectResponseOptions); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

const (
	ClientVPNAuthorizationRuleCreatedTimeout = 10 * time.Minute
	ClientVPNAuthorizationRuleDeletedTimeout = 10 * time.Minute
)

func WaitClientVPNAuthorizationRuleCreated(conn *ec2.EC2, endpointID, targetNetworkCIDR, accessGroupID string, timeout time.Duration) (*ec2.AuthorizationRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeAuthorizing},
		Target:  []string{ec2.ClientVpnAuthorizationRuleStatusCodeActive},
		Refresh: StatusClientVPNAuthorizationRule(conn, endpointID, targetNetworkCIDR, accessGroupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.AuthorizationRule); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNAuthorizationRuleDeleted(conn *ec2.EC2, endpointID, targetNetworkCIDR, accessGroupID string, timeout time.Duration) (*ec2.AuthorizationRule, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnAuthorizationRuleStatusCodeRevoking},
		Target:  []string{},
		Refresh: StatusClientVPNAuthorizationRule(conn, endpointID, targetNetworkCIDR, accessGroupID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.AuthorizationRule); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

const (
	ClientVPNNetworkAssociationCreatedTimeout     = 30 * time.Minute
	ClientVPNNetworkAssociationCreatedDelay       = 4 * time.Minute
	ClientVPNNetworkAssociationDeletedTimeout     = 30 * time.Minute
	ClientVPNNetworkAssociationDeletedDelay       = 4 * time.Minute
	ClientVPNNetworkAssociationStatusPollInterval = 10 * time.Second
)

func WaitClientVPNNetworkAssociationCreated(conn *ec2.EC2, associationID, endpointID string, timeout time.Duration) (*ec2.TargetNetwork, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{ec2.AssociationStatusCodeAssociating},
		Target:       []string{ec2.AssociationStatusCodeAssociated},
		Refresh:      StatusClientVPNNetworkAssociation(conn, associationID, endpointID),
		Timeout:      timeout,
		Delay:        ClientVPNNetworkAssociationCreatedDelay,
		PollInterval: ClientVPNNetworkAssociationStatusPollInterval,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TargetNetwork); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNNetworkAssociationDeleted(conn *ec2.EC2, associationID, endpointID string, timeout time.Duration) (*ec2.TargetNetwork, error) {
	stateConf := &resource.StateChangeConf{
		Pending:      []string{ec2.AssociationStatusCodeDisassociating},
		Target:       []string{},
		Refresh:      StatusClientVPNNetworkAssociation(conn, associationID, endpointID),
		Timeout:      timeout,
		Delay:        ClientVPNNetworkAssociationDeletedDelay,
		PollInterval: ClientVPNNetworkAssociationStatusPollInterval,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TargetNetwork); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

const (
	ClientVPNRouteCreatedTimeout = 1 * time.Minute
	ClientVPNRouteDeletedTimeout = 1 * time.Minute
)

func WaitClientVPNRouteCreated(conn *ec2.EC2, endpointID, targetSubnetID, destinationCIDR string, timeout time.Duration) (*ec2.ClientVpnRoute, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnRouteStatusCodeCreating},
		Target:  []string{ec2.ClientVpnRouteStatusCodeActive},
		Refresh: StatusClientVPNRoute(conn, endpointID, targetSubnetID, destinationCIDR),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ClientVpnRoute); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitClientVPNRouteDeleted(conn *ec2.EC2, endpointID, targetSubnetID, destinationCIDR string, timeout time.Duration) (*ec2.ClientVpnRoute, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.ClientVpnRouteStatusCodeActive, ec2.ClientVpnRouteStatusCodeDeleting},
		Target:  []string{},
		Refresh: StatusClientVPNRoute(conn, endpointID, targetSubnetID, destinationCIDR),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.ClientVpnRoute); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitFleet(conn *ec2.EC2, id string, pending, target []string, timeout time.Duration) (*ec2.FleetData, error) {
	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Refresh: StatusFleetState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.FleetData); ok {
		return output, err
	}

	return nil, err
}

func WaitImageAvailable(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Image, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.ImageStatePending},
		Target:     []string{ec2.ImageStateAvailable},
		Refresh:    StatusImageState(conn, id),
		Timeout:    timeout,
		Delay:      amiRetryDelay,
		MinTimeout: amiRetryMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Image); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitImageDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Image, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.ImageStateAvailable, ec2.ImageStateFailed, ec2.ImageStatePending},
		Target:     []string{},
		Refresh:    StatusImageState(conn, id),
		Timeout:    timeout,
		Delay:      amiRetryDelay,
		MinTimeout: amiRetryMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Image); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitInstanceIAMInstanceProfileUpdated(conn *ec2.EC2, instanceID string, expectedValue string) (*ec2.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{expectedValue},
		Refresh:    StatusInstanceIAMInstanceProfile(conn, instanceID),
		Timeout:    propagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Instance); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.InstanceStateNamePending},
		Target:     []string{ec2.InstanceStateNameRunning},
		Refresh:    StatusInstanceState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitInstanceDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.InstanceStateNamePending,
			ec2.InstanceStateNameRunning,
			ec2.InstanceStateNameShuttingDown,
			ec2.InstanceStateNameStopping,
			ec2.InstanceStateNameStopped,
		},
		Target:     []string{ec2.InstanceStateNameTerminated},
		Refresh:    StatusInstanceState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitInstanceReady(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.InstanceStateNamePending, ec2.InstanceStateNameStopping},
		Target:     []string{ec2.InstanceStateNameRunning, ec2.InstanceStateNameStopped},
		Refresh:    StatusInstanceState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitInstanceStarted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.InstanceStateNamePending, ec2.InstanceStateNameStopped},
		Target:     []string{ec2.InstanceStateNameRunning},
		Refresh:    StatusInstanceState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitInstanceStopped(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.InstanceStateNamePending,
			ec2.InstanceStateNameRunning,
			ec2.InstanceStateNameShuttingDown,
			ec2.InstanceStateNameStopping,
		},
		Target:     []string{ec2.InstanceStateNameStopped},
		Refresh:    StatusInstanceState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Instance); ok {
		if stateReason := output.StateReason; stateReason != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(stateReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitInstanceCapacityReservationSpecificationUpdated(conn *ec2.EC2, instanceID string, expectedValue *ec2.CapacityReservationSpecification) (*ec2.Instance, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(true)},
		Refresh:    StatusInstanceCapacityReservationSpecificationEquals(conn, instanceID, expectedValue),
		Timeout:    propagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Instance); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceMaintenanceOptionsAutoRecoveryUpdated(conn *ec2.EC2, id, expectedValue string, timeout time.Duration) (*ec2.InstanceMaintenanceOptions, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{expectedValue},
		Refresh:    StatusInstanceMaintenanceOptionsAutoRecovery(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.InstanceMaintenanceOptions); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceMetadataOptionsApplied(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.InstanceMetadataOptionsResponse, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.InstanceMetadataOptionsStatePending},
		Target:     []string{ec2.InstanceMetadataOptionsStateApplied},
		Refresh:    StatusInstanceMetadataOptionsState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.InstanceMetadataOptionsResponse); ok {
		return output, err
	}

	return nil, err
}

func WaitInstanceRootBlockDeviceDeleteOnTerminationUpdated(conn *ec2.EC2, id string, expectedValue bool, timeout time.Duration) (*ec2.EbsInstanceBlockDevice, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusInstanceRootBlockDeviceDeleteOnTermination(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.EbsInstanceBlockDevice); ok {
		return output, err
	}

	return nil, err
}

const ManagedPrefixListEntryCreateTimeout = 5 * time.Minute

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
		NotFoundChecks:            RouteNotFoundChecks,
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
		Pending:                   []string{},
		Target:                    []string{SecurityGroupStatusCreated},
		Refresh:                   StatusSecurityGroup(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            SecurityGroupNotFoundChecks,
		ContinuousTargetOccurence: 3,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.SecurityGroup); ok {
		return output, err
	}

	return nil, err
}

const (
	SubnetPropagationTimeout                     = 2 * time.Minute
	SubnetAttributePropagationTimeout            = 5 * time.Minute
	SubnetIPv6CIDRBlockAssociationCreatedTimeout = 3 * time.Minute
	SubnetIPv6CIDRBlockAssociationDeletedTimeout = 3 * time.Minute
)

func WaitSubnetAvailable(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.SubnetStatePending},
		Target:  []string{ec2.SubnetStateAvailable},
		Refresh: StatusSubnetState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetIPv6CIDRBlockAssociationCreated(conn *ec2.EC2, id string) (*ec2.SubnetCidrBlockState, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.SubnetCidrBlockStateCodeAssociating, ec2.SubnetCidrBlockStateCodeDisassociated, ec2.SubnetCidrBlockStateCodeFailing},
		Target:  []string{ec2.SubnetCidrBlockStateCodeAssociated},
		Refresh: StatusSubnetIPv6CIDRBlockAssociationState(conn, id),
		Timeout: SubnetIPv6CIDRBlockAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.SubnetCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.SubnetCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitSubnetIPv6CIDRBlockAssociationDeleted(conn *ec2.EC2, id string) (*ec2.SubnetCidrBlockState, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.SubnetCidrBlockStateCodeAssociated, ec2.SubnetCidrBlockStateCodeDisassociating, ec2.SubnetCidrBlockStateCodeFailing},
		Target:  []string{},
		Refresh: StatusSubnetIPv6CIDRBlockAssociationState(conn, id),
		Timeout: SubnetIPv6CIDRBlockAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.SubnetCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.SubnetCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitSubnetAssignIPv6AddressOnCreationUpdated(conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetAssignIPv6AddressOnCreation(conn, subnetID),
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

func waitSubnetEnableDNS64Updated(conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableDNS64(conn, subnetID),
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

func waitSubnetEnableResourceNameDNSAAAARecordOnLaunchUpdated(conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableResourceNameDNSAAAARecordOnLaunch(conn, subnetID),
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

func waitSubnetEnableResourceNameDNSARecordOnLaunchUpdated(conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableResourceNameDNSARecordOnLaunch(conn, subnetID),
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

func WaitSubnetPrivateDNSHostnameTypeOnLaunchUpdated(conn *ec2.EC2, subnetID string, expectedValue string) (*ec2.Subnet, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{expectedValue},
		Refresh:    StatusSubnetPrivateDNSHostnameTypeOnLaunch(conn, subnetID),
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
	TransitGatewayIncorrectStateTimeout = 5 * time.Minute
)

func WaitTransitGatewayCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayStatePending},
		Target:  []string{ec2.TransitGatewayStateAvailable},
		Refresh: StatusTransitGatewayState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGateway); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{ec2.TransitGatewayStateAvailable, ec2.TransitGatewayStateDeleting},
		Target:         []string{},
		Refresh:        StatusTransitGatewayState(conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGateway); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayUpdated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayStateModifying},
		Target:  []string{ec2.TransitGatewayStateAvailable},
		Refresh: StatusTransitGatewayState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGateway); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayConnectCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGatewayConnect, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayAttachmentStatePending},
		Target:  []string{ec2.TransitGatewayAttachmentStateAvailable},
		Refresh: StatusTransitGatewayConnectState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayConnect); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayConnectDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGatewayConnect, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{ec2.TransitGatewayAttachmentStateAvailable, ec2.TransitGatewayAttachmentStateDeleting},
		Target:         []string{},
		Refresh:        StatusTransitGatewayConnectState(conn, id),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayConnect); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayConnectPeerCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGatewayConnectPeer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayConnectPeerStatePending},
		Target:  []string{ec2.TransitGatewayConnectPeerStateAvailable},
		Refresh: StatusTransitGatewayConnectPeerState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayConnectPeer); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayConnectPeerDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGatewayConnectPeer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayConnectPeerStateAvailable, ec2.TransitGatewayConnectPeerStateDeleting},
		Target:  []string{},
		Refresh: StatusTransitGatewayConnectPeerState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayConnectPeer); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayMulticastDomainCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGatewayMulticastDomain, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayMulticastDomainStatePending},
		Target:  []string{ec2.TransitGatewayMulticastDomainStateAvailable},
		Refresh: StatusTransitGatewayMulticastDomainState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayMulticastDomain); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayMulticastDomainDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.TransitGatewayMulticastDomain, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayMulticastDomainStateAvailable, ec2.TransitGatewayMulticastDomainStateDeleting},
		Target:  []string{},
		Refresh: StatusTransitGatewayMulticastDomainState(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayMulticastDomain); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayMulticastDomainAssociationCreated(conn *ec2.EC2, multicastDomainID, attachmentID, subnetID string, timeout time.Duration) (*ec2.TransitGatewayMulticastDomainAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AssociationStatusCodeAssociating},
		Target:  []string{ec2.AssociationStatusCodeAssociated},
		Refresh: StatusTransitGatewayMulticastDomainAssociationState(conn, multicastDomainID, attachmentID, subnetID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayMulticastDomainAssociation); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayMulticastDomainAssociationDeleted(conn *ec2.EC2, multicastDomainID, attachmentID, subnetID string, timeout time.Duration) (*ec2.TransitGatewayMulticastDomainAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AssociationStatusCodeAssociated, ec2.AssociationStatusCodeDisassociating},
		Target:  []string{},
		Refresh: StatusTransitGatewayMulticastDomainAssociationState(conn, multicastDomainID, attachmentID, subnetID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayMulticastDomainAssociation); ok {
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
	TransitGatewayRouteCreatedTimeout = 2 * time.Minute
	TransitGatewayRouteDeletedTimeout = 2 * time.Minute
)

func WaitTransitGatewayRouteCreated(conn *ec2.EC2, transitGatewayRouteTableID, destination string) (*ec2.TransitGatewayRoute, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayRouteStatePending},
		Target:  []string{ec2.TransitGatewayRouteStateActive, ec2.TransitGatewayRouteStateBlackhole},
		Timeout: TransitGatewayRouteCreatedTimeout,
		Refresh: StatusTransitGatewayRouteState(conn, transitGatewayRouteTableID, destination),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayRoute); ok {
		return output, err
	}

	return nil, err
}

func WaitTransitGatewayRouteDeleted(conn *ec2.EC2, transitGatewayRouteTableID, destination string) (*ec2.TransitGatewayRoute, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.TransitGatewayRouteStateActive, ec2.TransitGatewayRouteStateBlackhole, ec2.TransitGatewayRouteStateDeleting},
		Target:  []string{},
		Timeout: TransitGatewayRouteDeletedTimeout,
		Refresh: StatusTransitGatewayRouteState(conn, transitGatewayRouteTableID, destination),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.TransitGatewayRoute); ok {
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

	if tfawserr.ErrCodeEquals(err, errCodeInvalidRouteTableIDNotFound) {
		return nil, nil
	}

	if output, ok := outputRaw.(*ec2.TransitGatewayRouteTablePropagation); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Volume, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VolumeStateCreating},
		Target:     []string{ec2.VolumeStateAvailable},
		Refresh:    StatusVolumeState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Volume); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Volume, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VolumeStateDeleting},
		Target:     []string{},
		Refresh:    StatusVolumeState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Volume); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeUpdated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Volume, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VolumeStateCreating, ec2.VolumeModificationStateModifying},
		Target:     []string{ec2.VolumeStateAvailable, ec2.VolumeStateInUse},
		Refresh:    StatusVolumeState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Volume); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeAttachmentCreated(conn *ec2.EC2, volumeID, instanceID, deviceName string, timeout time.Duration) (*ec2.VolumeAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VolumeAttachmentStateAttaching},
		Target:     []string{ec2.VolumeAttachmentStateAttached},
		Refresh:    StatusVolumeAttachmentState(conn, volumeID, instanceID, deviceName),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VolumeAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeAttachmentDeleted(conn *ec2.EC2, volumeID, instanceID, deviceName string, timeout time.Duration) (*ec2.VolumeAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VolumeAttachmentStateDetaching},
		Target:     []string{},
		Refresh:    StatusVolumeAttachmentState(conn, volumeID, instanceID, deviceName),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VolumeAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitVolumeModificationComplete(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VolumeModification, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.VolumeModificationStateModifying},
		// The volume is useable once the state is "optimizing", but will not be at full performance.
		// Optimization can take hours. e.g. a full 1 TiB drive takes approximately 6 hours to optimize,
		// according to https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/monitoring-volume-modifications.html.
		Target:     []string{ec2.VolumeModificationStateCompleted, ec2.VolumeModificationStateOptimizing},
		Refresh:    StatusVolumeModificationState(conn, id),
		Timeout:    timeout,
		Delay:      30 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VolumeModification); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

const (
	vpcAttributePropagationTimeout = 5 * time.Minute
	vpcCreatedTimeout              = 10 * time.Minute
	vpcDeletedTimeout              = 5 * time.Minute
)

func WaitVPCCreated(conn *ec2.EC2, id string) (*ec2.Vpc, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.VpcStatePending},
		Target:  []string{ec2.VpcStateAvailable},
		Refresh: StatusVPCState(conn, id),
		Timeout: vpcCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Vpc); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCAttributeUpdated(conn *ec2.EC2, vpcID string, attribute string, expectedValue bool) (*ec2.Vpc, error) {
	stateConf := &resource.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusVPCAttributeValue(conn, vpcID, attribute),
		Timeout:    vpcAttributePropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.Vpc); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCCIDRBlockAssociationCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociating, ec2.VpcCidrBlockStateCodeDisassociated, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh:    StatusVPCCIDRBlockAssociationState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCCIDRBlockAssociationDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociated, ec2.VpcCidrBlockStateCodeDisassociating, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{},
		Refresh:    StatusVPCCIDRBlockAssociationState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

const (
	vpcIPv6CIDRBlockAssociationCreatedTimeout = 10 * time.Minute
	vpcIPv6CIDRBlockAssociationDeletedTimeout = 5 * time.Minute
)

func WaitVPCIPv6CIDRBlockAssociationCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociating, ec2.VpcCidrBlockStateCodeDisassociated, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh:    StatusVPCIPv6CIDRBlockAssociationState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCIPv6CIDRBlockAssociationDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociated, ec2.VpcCidrBlockStateCodeDisassociating, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{},
		Refresh:    StatusVPCIPv6CIDRBlockAssociationState(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

const (
	VPCPeeringConnectionOptionsPropagationTimeout = 3 * time.Minute
)

func WaitVPCPeeringConnectionActive(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcPeeringConnection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.VpcPeeringConnectionStateReasonCodeInitiatingRequest, ec2.VpcPeeringConnectionStateReasonCodeProvisioning},
		Target:  []string{ec2.VpcPeeringConnectionStateReasonCodeActive, ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance},
		Refresh: StatusVPCPeeringConnectionActive(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcPeeringConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitVPCPeeringConnectionDeleted(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcPeeringConnection, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ec2.VpcPeeringConnectionStateReasonCodeActive,
			ec2.VpcPeeringConnectionStateReasonCodeDeleting,
			ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance,
		},
		Target:  []string{},
		Refresh: StatusVPCPeeringConnectionDeleted(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcPeeringConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

const (
	VPNGatewayDeletedTimeout = 5 * time.Minute

	VPNGatewayVPCAttachmentAttachedTimeout = 15 * time.Minute
	VPNGatewayVPCAttachmentDetachedTimeout = 30 * time.Minute
)

func WaitVPNGatewayVPCAttachmentAttached(conn *ec2.EC2, vpnGatewayID, vpcID string) (*ec2.VpcAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusAttaching},
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
		Target:  []string{},
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
	customerGatewayCreatedTimeout = 10 * time.Minute
	customerGatewayDeletedTimeout = 5 * time.Minute
)

func WaitCustomerGatewayCreated(conn *ec2.EC2, id string) (*ec2.CustomerGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{CustomerGatewayStatePending},
		Target:     []string{CustomerGatewayStateAvailable},
		Refresh:    StatusCustomerGatewayState(conn, id),
		Timeout:    customerGatewayCreatedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.CustomerGateway); ok {
		return output, err
	}

	return nil, err
}

func WaitCustomerGatewayDeleted(conn *ec2.EC2, id string) (*ec2.CustomerGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{CustomerGatewayStateAvailable, CustomerGatewayStateDeleting},
		Target:  []string{},
		Refresh: StatusCustomerGatewayState(conn, id),
		Timeout: customerGatewayDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.CustomerGateway); ok {
		return output, err
	}

	return nil, err
}

const (
	natGatewayCreatedTimeout = 10 * time.Minute
	natGatewayDeletedTimeout = 30 * time.Minute
)

func WaitNATGatewayCreated(conn *ec2.EC2, id string) (*ec2.NatGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.NatGatewayStatePending},
		Target:  []string{ec2.NatGatewayStateAvailable},
		Refresh: StatusNATGatewayState(conn, id),
		Timeout: natGatewayCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.NatGateway); ok {
		if state := aws.StringValue(output.State); state == ec2.NatGatewayStateFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.FailureCode), aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayDeleted(conn *ec2.EC2, id string) (*ec2.NatGateway, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.NatGatewayStateDeleting},
		Target:     []string{},
		Refresh:    StatusNATGatewayState(conn, id),
		Timeout:    natGatewayDeletedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.NatGateway); ok {
		if state := aws.StringValue(output.State); state == ec2.NatGatewayStateFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.FailureCode), aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

const (
	vpnConnectionCreatedTimeout = 40 * time.Minute
	vpnConnectionDeletedTimeout = 30 * time.Minute
	vpnConnectionUpdatedTimeout = 30 * time.Minute
)

func WaitVPNConnectionCreated(conn *ec2.EC2, id string) (*ec2.VpnConnection, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpnStatePending},
		Target:     []string{ec2.VpnStateAvailable},
		Refresh:    StatusVPNConnectionState(conn, id),
		Timeout:    vpnConnectionCreatedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNConnectionDeleted(conn *ec2.EC2, id string) (*ec2.VpnConnection, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VpnStateDeleting},
		Target:     []string{},
		Refresh:    StatusVPNConnectionState(conn, id),
		Timeout:    vpnConnectionDeletedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNConnectionUpdated(conn *ec2.EC2, id string) (*ec2.VpnConnection, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{vpnStateModifying},
		Target:     []string{ec2.VpnStateAvailable},
		Refresh:    StatusVPNConnectionState(conn, id),
		Timeout:    vpnConnectionUpdatedTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpnConnection); ok {
		return output, err
	}

	return nil, err
}

const (
	vpnConnectionRouteCreatedTimeout = 15 * time.Second
	vpnConnectionRouteDeletedTimeout = 15 * time.Second
)

func WaitVPNConnectionRouteCreated(conn *ec2.EC2, vpnConnectionID, cidrBlock string) (*ec2.VpnStaticRoute, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.VpnStatePending},
		Target:  []string{ec2.VpnStateAvailable},
		Refresh: StatusVPNConnectionRouteState(conn, vpnConnectionID, cidrBlock),
		Timeout: vpnConnectionRouteCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpnStaticRoute); ok {
		return output, err
	}

	return nil, err
}

func WaitVPNConnectionRouteDeleted(conn *ec2.EC2, vpnConnectionID, cidrBlock string) (*ec2.VpnStaticRoute, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.VpnStatePending, ec2.VpnStateAvailable, ec2.VpnStateDeleting},
		Target:  []string{},
		Refresh: StatusVPNConnectionRouteState(conn, vpnConnectionID, cidrBlock),
		Timeout: vpnConnectionRouteDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpnStaticRoute); ok {
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
	dhcpOptionSetDeletedTimeout = 3 * time.Minute
)

const (
	internetGatewayAttachedTimeout = 4 * time.Minute
	internetGatewayDeletedTimeout  = 10 * time.Minute
	internetGatewayDetachedTimeout = 15 * time.Minute
)

func WaitInternetGatewayAttached(conn *ec2.EC2, internetGatewayID, vpcID string, timeout time.Duration) (*ec2.InternetGatewayAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{ec2.AttachmentStatusAttaching},
		Target:         []string{InternetGatewayAttachmentStateAvailable},
		Timeout:        timeout,
		NotFoundChecks: InternetGatewayNotFoundChecks,
		Refresh:        StatusInternetGatewayAttachmentState(conn, internetGatewayID, vpcID),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.InternetGatewayAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitInternetGatewayDetached(conn *ec2.EC2, internetGatewayID, vpcID string, timeout time.Duration) (*ec2.InternetGatewayAttachment, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{InternetGatewayAttachmentStateAvailable, ec2.AttachmentStatusDetaching},
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

func WaitSpotFleetRequestCreated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SpotFleetRequestConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.BatchStateSubmitted},
		Target:     []string{ec2.BatchStateActive},
		Refresh:    StatusSpotFleetRequestState(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func WaitSpotFleetRequestFulfilled(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SpotFleetRequestConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.ActivityStatusPendingFulfillment},
		Target:     []string{ec2.ActivityStatusFulfilled},
		Refresh:    StatusSpotFleetActivityStatus(conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func WaitSpotFleetRequestUpdated(conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SpotFleetRequestConfig, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.BatchStateModifying},
		Target:     []string{ec2.BatchStateActive},
		Refresh:    StatusSpotFleetRequestState(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.SpotFleetRequestConfig); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCEndpointAccepted(conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{vpcEndpointStatePendingAcceptance},
		Target:     []string{vpcEndpointStateAvailable},
		Timeout:    timeout,
		Refresh:    StatusVPCEndpointState(conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcEndpoint); ok {
		if state, lastError := aws.StringValue(output.State), output.LastError; state == vpcEndpointStateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(lastError.Code), aws.StringValue(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCEndpointAvailable(conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{vpcEndpointStatePending},
		Target:     []string{vpcEndpointStateAvailable, vpcEndpointStatePendingAcceptance},
		Timeout:    timeout,
		Refresh:    StatusVPCEndpointState(conn, vpcEndpointID),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcEndpoint); ok {
		if state, lastError := aws.StringValue(output.State), output.LastError; state == vpcEndpointStateFailed && lastError != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(lastError.Code), aws.StringValue(lastError.Message)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCEndpointDeleted(conn *ec2.EC2, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{vpcEndpointStateDeleting},
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
		Timeout:                   propagationTimeout,
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
		Timeout:                   propagationTimeout,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitEBSSnapshotImportComplete(conn *ec2.EC2, importTaskID string) (*ec2.SnapshotTaskDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			EBSSnapshotImportStateActive,
			EBSSnapshotImportStateUpdating,
			EBSSnapshotImportStateValidating,
			EBSSnapshotImportStateValidated,
			EBSSnapshotImportStateConverting,
		},
		Target:  []string{EBSSnapshotImportStateCompleted},
		Refresh: StatusEBSSnapshotImport(conn, importTaskID),
		Timeout: 60 * time.Minute,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.SnapshotTaskDetail); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitVPCEndpointConnectionAccepted(conn *ec2.EC2, serviceID, vpcEndpointID string, timeout time.Duration) (*ec2.VpcEndpointConnection, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{vpcEndpointStatePendingAcceptance, vpcEndpointStatePending},
		Target:     []string{vpcEndpointStateAvailable},
		Refresh:    statusVPCEndpointConnectionVPCEndpointState(conn, serviceID, vpcEndpointID),
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*ec2.VpcEndpointConnection); ok {
		return output, err
	}

	return nil, err
}

func WaitEBSSnapshotTierArchive(conn *ec2.EC2, id string) (*ec2.SnapshotTierStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"standard"},
		Target:  []string{ec2.TargetStorageTierArchive},
		Refresh: StatusSnapshotTierStatus(conn, id),
		Timeout: 60 * time.Minute,
		Delay:   10 * time.Second,
	}

	detail, err := stateConf.WaitForState()
	if err != nil {
		return nil, err
	} else {
		return detail.(*ec2.SnapshotTierStatus), nil
	}
}

// WaitVolumeAttachmentAttached waits for a VolumeAttachment to return Attached
func WaitVolumeAttachmentAttached(conn *ec2.EC2, name, volumeID, instanceID string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ec2.VolumeAttachmentStateAttaching},
		Target:     []string{ec2.VolumeAttachmentStateAttached},
		Refresh:    volumeAttachmentStateRefreshFunc(conn, name, volumeID, instanceID),
		Timeout:    5 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
