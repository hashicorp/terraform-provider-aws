// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func WaitVPCCreatedV2(ctx context.Context, conn *ec2.Client, id string) (*awstypes.Vpc, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.VpcStatePending)},
		Target:  []string{string(awstypes.VpcStateAvailable)},
		Refresh: StatusVPCStateV2(ctx, conn, id),
		Timeout: vpcCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Vpc); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCIPv6CIDRBlockAssociationCreatedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(awstypes.VpcCidrBlockStateCodeAssociating), string(awstypes.VpcCidrBlockStateCodeDisassociated), string(awstypes.VpcCidrBlockStateCodeFailing)},
		Target:     []string{string(awstypes.VpcCidrBlockStateCodeAssociated)},
		Refresh:    StatusVPCIPv6CIDRBlockAssociationStateV2(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcCidrBlockState); ok {
		if state := output.State; state == awstypes.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func WaitVPCAttributeUpdatedV2(ctx context.Context, conn *ec2.Client, vpcID string, attribute string, expectedValue bool) (*awstypes.Vpc, error) {
	stateConf := &retry.StateChangeConf{
		Target:     []string{strconv.FormatBool(expectedValue)},
		Refresh:    StatusVPCAttributeValueV2(ctx, conn, vpcID, attribute),
		Timeout:    ec2PropagationTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Vpc); ok {
		return output, err
	}

	return nil, err
}

func WaitVPCIPv6CIDRBlockAssociationDeletedV2(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.VpcCidrBlockState, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{string(awstypes.VpcCidrBlockStateCodeAssociated), string(awstypes.VpcCidrBlockStateCodeDisassociating), string(awstypes.VpcCidrBlockStateCodeFailing)},
		Target:     []string{},
		Refresh:    StatusVPCIPv6CIDRBlockAssociationStateV2(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.VpcCidrBlockState); ok {
		if state := output.State; state == awstypes.VpcCidrBlockStateCodeFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}
