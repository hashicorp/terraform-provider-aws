// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

//
// Move functions to waitv2.go as they are migrated to AWS SDK for Go v2.
//

const (
	InstanceReadyTimeout = 10 * time.Minute
	InstanceStartTimeout = 10 * time.Minute
	InstanceStopTimeout  = 10 * time.Minute

	// General timeout for IAM resource change to propagate.
	// See https://docs.aws.amazon.com/IAM/latest/UserGuide/troubleshoot_general.html#troubleshoot_general_eventual-consistency.
	// We have settled on 2 minutes as the best timeout value.
	iamPropagationTimeout = 2 * time.Minute

	// General timeout for EC2 resource changes to propagate.
	// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/query-api-troubleshooting.html#eventual-consistency.
	ec2PropagationTimeout = 5 * time.Minute // nosemgrep:ci.ec2-in-const-name, ci.ec2-in-var-name

	RouteNotFoundChecks                        = 1000 // Should exceed any reasonable custom timeout value.
	RouteTableNotFoundChecks                   = 1000 // Should exceed any reasonable custom timeout value.
	RouteTableAssociationCreatedNotFoundChecks = 1000 // Should exceed any reasonable custom timeout value.
	SecurityGroupNotFoundChecks                = 1000 // Should exceed any reasonable custom timeout value.
	InternetGatewayNotFoundChecks              = 1000 // Should exceed any reasonable custom timeout value.
)

const ManagedPrefixListEntryCreateTimeout = 5 * time.Minute

func WaitSecurityGroupCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.SecurityGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{SecurityGroupStatusCreated},
		Refresh:                   StatusSecurityGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            SecurityGroupNotFoundChecks,
		ContinuousTargetOccurence: 3,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SecurityGroup); ok {
		return output, err
	}

	return nil, err
}

const (
	SubnetIPv6CIDRBlockAssociationCreatedTimeout = 3 * time.Minute
	SubnetIPv6CIDRBlockAssociationDeletedTimeout = 3 * time.Minute
)

func WaitSubnetAvailable(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.SubnetStatePending},
		Target:  []string{ec2.SubnetStateAvailable},
		Refresh: StatusSubnetState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetIPv6CIDRBlockAssociationCreated(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SubnetCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.SubnetCidrBlockStateCodeAssociating, ec2.SubnetCidrBlockStateCodeDisassociated, ec2.SubnetCidrBlockStateCodeFailing},
		Target:  []string{ec2.SubnetCidrBlockStateCodeAssociated},
		Refresh: StatusSubnetIPv6CIDRBlockAssociationState(ctx, conn, id),
		Timeout: SubnetIPv6CIDRBlockAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SubnetCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.SubnetCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitSubnetIPv6CIDRBlockAssociationDeleted(ctx context.Context, conn *ec2.EC2, id string) (*ec2.SubnetCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.SubnetCidrBlockStateCodeAssociated, ec2.SubnetCidrBlockStateCodeDisassociating, ec2.SubnetCidrBlockStateCodeFailing},
		Target:  []string{},
		Refresh: StatusSubnetIPv6CIDRBlockAssociationState(ctx, conn, id),
		Timeout: SubnetIPv6CIDRBlockAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.SubnetCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.SubnetCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitSubnetAssignIPv6AddressOnCreationUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetAssignIPv6AddressOnCreation(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func waitSubnetEnableLniAtDeviceIndexUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue int64) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatInt(expectedValue, 10)},
		Refresh:    StatusSubnetEnableLniAtDeviceIndex(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func waitSubnetEnableDNS64Updated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableDNS64(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func waitSubnetEnableResourceNameDNSAAAARecordOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableResourceNameDNSAAAARecordOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func waitSubnetEnableResourceNameDNSARecordOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetEnableResourceNameDNSARecordOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetMapCustomerOwnedIPOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetMapCustomerOwnedIPOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetMapPublicIPOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue bool) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusSubnetMapPublicIPOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

func WaitSubnetPrivateDNSHostnameTypeOnLaunchUpdated(ctx context.Context, conn *ec2.EC2, subnetID string, expectedValue string) (*ec2.Subnet, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{expectedValue},
		Refresh:    StatusSubnetPrivateDNSHostnameTypeOnLaunch(ctx, conn, subnetID),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.Subnet); ok {
		return output, err
	}

	return nil, err
}

const (
	vpcCreatedTimeout = 10 * time.Minute
	vpcDeletedTimeout = 5 * time.Minute
)

func WaitVPCCIDRBlockAssociationCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociating, ec2.VpcCidrBlockStateCodeDisassociated, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh:    StatusVPCCIDRBlockAssociationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCCIDRBlockAssociationDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociated, ec2.VpcCidrBlockStateCodeDisassociating, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{},
		Refresh:    StatusVPCCIDRBlockAssociationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

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

func WaitVPCIPv6CIDRBlockAssociationCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociating, ec2.VpcCidrBlockStateCodeDisassociated, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{ec2.VpcCidrBlockStateCodeAssociated},
		Refresh:    StatusVPCIPv6CIDRBlockAssociationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCIPv6CIDRBlockAssociationDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.VpcCidrBlockStateCodeAssociated, ec2.VpcCidrBlockStateCodeDisassociating, ec2.VpcCidrBlockStateCodeFailing},
		Target:     []string{},
		Refresh:    StatusVPCIPv6CIDRBlockAssociationState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcCidrBlockState); ok {
		if state := aws.StringValue(output.State); state == ec2.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCPeeringConnectionActive(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcPeeringConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.VpcPeeringConnectionStateReasonCodeInitiatingRequest, ec2.VpcPeeringConnectionStateReasonCodeProvisioning},
		Target:  []string{ec2.VpcPeeringConnectionStateReasonCodeActive, ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance},
		Refresh: StatusVPCPeeringConnectionActive(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcPeeringConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitVPCPeeringConnectionDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.VpcPeeringConnection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			ec2.VpcPeeringConnectionStateReasonCodeActive,
			ec2.VpcPeeringConnectionStateReasonCodeDeleting,
			ec2.VpcPeeringConnectionStateReasonCodePendingAcceptance,
		},
		Target:  []string{},
		Refresh: StatusVPCPeeringConnectionDeleted(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.VpcPeeringConnection); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.Message)))

		return output, err
	}

	return nil, err
}

func WaitNATGatewayCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NatGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayStatePending},
		Target:  []string{ec2.NatGatewayStateAvailable},
		Refresh: StatusNATGatewayState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGateway); ok {
		if state := aws.StringValue(output.State); state == ec2.NatGatewayStateFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.FailureCode), aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayDeleted(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NatGateway, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.NatGatewayStateDeleting},
		Target:     []string{},
		Refresh:    StatusNATGatewayState(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGateway); ok {
		if state := aws.StringValue(output.State); state == ec2.NatGatewayStateFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.FailureCode), aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayAddressAssigned(ctx context.Context, conn *ec2.EC2, natGatewayID, privateIP string, timeout time.Duration) (*ec2.NatGatewayAddress, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayAddressStatusAssigning},
		Target:  []string{ec2.NatGatewayAddressStatusSucceeded},
		Refresh: StatusNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx, conn, natGatewayID, privateIP),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGatewayAddress); ok {
		if status := aws.StringValue(output.Status); status == ec2.NatGatewayAddressStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayAddressAssociated(ctx context.Context, conn *ec2.EC2, natGatewayID, allocationID string, timeout time.Duration) (*ec2.NatGatewayAddress, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayAddressStatusAssociating},
		Target:  []string{ec2.NatGatewayAddressStatusSucceeded},
		Refresh: StatusNATGatewayAddressByNATGatewayIDAndAllocationID(ctx, conn, natGatewayID, allocationID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGatewayAddress); ok {
		if status := aws.StringValue(output.Status); status == ec2.NatGatewayAddressStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayAddressDisassociated(ctx context.Context, conn *ec2.EC2, natGatewayID, allocationID string, timeout time.Duration) (*ec2.NatGatewayAddress, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayAddressStatusSucceeded, ec2.NatGatewayAddressStatusDisassociating},
		Target:  []string{},
		Refresh: StatusNATGatewayAddressByNATGatewayIDAndAllocationID(ctx, conn, natGatewayID, allocationID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGatewayAddress); ok {
		if status := aws.StringValue(output.Status); status == ec2.NatGatewayAddressStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitNATGatewayAddressUnassigned(ctx context.Context, conn *ec2.EC2, natGatewayID, privateIP string, timeout time.Duration) (*ec2.NatGatewayAddress, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.NatGatewayAddressStatusUnassigning},
		Target:  []string{},
		Refresh: StatusNATGatewayAddressByNATGatewayIDAndPrivateIP(ctx, conn, natGatewayID, privateIP),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NatGatewayAddress); ok {
		if status := aws.StringValue(output.Status); status == ec2.NatGatewayAddressStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureMessage)))
		}

		return output, err
	}

	return nil, err
}

const (
	dhcpOptionSetDeletedTimeout = 3 * time.Minute
)

func WaitInternetGatewayAttached(ctx context.Context, conn *ec2.EC2, internetGatewayID, vpcID string, timeout time.Duration) (*ec2.InternetGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{ec2.AttachmentStatusAttaching},
		Target:         []string{InternetGatewayAttachmentStateAvailable},
		Timeout:        timeout,
		NotFoundChecks: InternetGatewayNotFoundChecks,
		Refresh:        StatusInternetGatewayAttachmentState(ctx, conn, internetGatewayID, vpcID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.InternetGatewayAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitInternetGatewayDetached(ctx context.Context, conn *ec2.EC2, internetGatewayID, vpcID string, timeout time.Duration) (*ec2.InternetGatewayAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{InternetGatewayAttachmentStateAvailable, ec2.AttachmentStatusDetaching},
		Target:  []string{},
		Timeout: timeout,
		Refresh: StatusInternetGatewayAttachmentState(ctx, conn, internetGatewayID, vpcID),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.InternetGatewayAttachment); ok {
		return output, err
	}

	return nil, err
}

const (
	ManagedPrefixListTimeout = 15 * time.Minute
)

func WaitManagedPrefixListCreated(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.PrefixListStateCreateInProgress},
		Target:  []string{ec2.PrefixListStateCreateComplete},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.ManagedPrefixList); ok {
		if state := aws.StringValue(output.State); state == ec2.PrefixListStateCreateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitManagedPrefixListModified(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.PrefixListStateModifyInProgress},
		Target:  []string{ec2.PrefixListStateModifyComplete},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.ManagedPrefixList); ok {
		if state := aws.StringValue(output.State); state == ec2.PrefixListStateModifyFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StateMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitManagedPrefixListDeleted(ctx context.Context, conn *ec2.EC2, id string) (*ec2.ManagedPrefixList, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.PrefixListStateDeleteInProgress},
		Target:  []string{},
		Timeout: ManagedPrefixListTimeout,
		Refresh: StatusManagedPrefixListState(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

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

func WaitNetworkInterfaceAttached(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusAttaching},
		Target:  []string{ec2.AttachmentStatusAttached},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceAttachmentStatus(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceAvailableAfterUse(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterface, error) {
	// Hyperplane attached ENI.
	// Wait for it to be moved into a removable state.
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ec2.NetworkInterfaceStatusInUse},
		Target:     []string{ec2.NetworkInterfaceStatusAvailable},
		Timeout:    timeout,
		Refresh:    StatusNetworkInterfaceStatus(ctx, conn, id),
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
		// Handle EC2 ENI eventual consistency. It can take up to 3 minutes.
		ContinuousTargetOccurence: 18,
		NotFoundChecks:            1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceCreated(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterface, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{NetworkInterfaceStatusPending},
		Target:  []string{ec2.NetworkInterfaceStatusAvailable},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceStatus(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func WaitNetworkInterfaceDetached(ctx context.Context, conn *ec2.EC2, id string, timeout time.Duration) (*ec2.NetworkInterfaceAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ec2.AttachmentStatusAttached, ec2.AttachmentStatusDetaching},
		Target:  []string{ec2.AttachmentStatusDetached},
		Timeout: timeout,
		Refresh: StatusNetworkInterfaceAttachmentStatus(ctx, conn, id),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterfaceAttachment); ok {
		return output, err
	}

	return nil, err
}
