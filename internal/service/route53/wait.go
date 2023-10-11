// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	changeTimeout      = 30 * time.Minute
	changeMinTimeout   = 5 * time.Second
	changePollInterval = 15 * time.Second
	changeMinDelay     = 10
	changeMaxDelay     = 30

	hostedZoneDNSSECStatusTimeout = 5 * time.Minute

	keySigningKeyStatusTimeout = 5 * time.Minute

	trafficPolicyInstanceOperationTimeout = 4 * time.Minute
)

func waitChangeInfoStatusInsync(ctx context.Context, conn *route53.Route53, changeID string) (*route53.ChangeInfo, error) { //nolint:unparam
	// Route53 is vulnerable to throttling so longer delays, poll intervals helps significantly to avoid

	stateConf := &retry.StateChangeConf{
		Pending:      []string{route53.ChangeStatusPending},
		Target:       []string{route53.ChangeStatusInsync},
		Delay:        time.Duration(rand.Int63n(changeMaxDelay-changeMinDelay)+changeMinDelay) * time.Second,
		MinTimeout:   changeMinTimeout,
		PollInterval: changePollInterval,
		Refresh:      statusChangeInfo(ctx, conn, changeID),
		Timeout:      changeTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53.ChangeInfo); ok {
		return output, err
	}

	return nil, err
}

func waitHostedZoneDNSSECStatusUpdated(ctx context.Context, conn *route53.Route53, hostedZoneID string, status string) (*route53.DNSSECStatus, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:     []string{status},
		Refresh:    statusHostedZoneDNSSEC(ctx, conn, hostedZoneID),
		MinTimeout: 5 * time.Second,
		Timeout:    hostedZoneDNSSECStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53.DNSSECStatus); ok {
		if serveSignature := aws.StringValue(output.ServeSignature); serveSignature == ServeSignatureInternalFailure {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitKeySigningKeyStatusUpdated(ctx context.Context, conn *route53.Route53, hostedZoneID string, name string, status string) (*route53.KeySigningKey, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Target:     []string{status},
		Refresh:    statusKeySigningKey(ctx, conn, hostedZoneID, name),
		MinTimeout: 5 * time.Second,
		Timeout:    keySigningKeyStatusTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53.KeySigningKey); ok {
		if status := aws.StringValue(output.Status); status == KeySigningKeyStatusInternalFailure {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitTrafficPolicyInstanceStateCreated(ctx context.Context, conn *route53.Route53, id string) (*route53.TrafficPolicyInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{TrafficPolicyInstanceStateCreating},
		Target:  []string{TrafficPolicyInstanceStateApplied},
		Refresh: statusTrafficPolicyInstanceState(ctx, conn, id),
		Timeout: trafficPolicyInstanceOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53.TrafficPolicyInstance); ok {
		if state := aws.StringValue(output.State); state == TrafficPolicyInstanceStateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTrafficPolicyInstanceStateDeleted(ctx context.Context, conn *route53.Route53, id string) (*route53.TrafficPolicyInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{TrafficPolicyInstanceStateDeleting},
		Target:  []string{},
		Refresh: statusTrafficPolicyInstanceState(ctx, conn, id),
		Timeout: trafficPolicyInstanceOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53.TrafficPolicyInstance); ok {
		if state := aws.StringValue(output.State); state == TrafficPolicyInstanceStateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitTrafficPolicyInstanceStateUpdated(ctx context.Context, conn *route53.Route53, id string) (*route53.TrafficPolicyInstance, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{TrafficPolicyInstanceStateUpdating},
		Target:  []string{TrafficPolicyInstanceStateApplied},
		Refresh: statusTrafficPolicyInstanceState(ctx, conn, id),
		Timeout: trafficPolicyInstanceOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53.TrafficPolicyInstance); ok {
		if state := aws.StringValue(output.State); state == TrafficPolicyInstanceStateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Message)))
		}

		return output, err
	}

	return nil, err
}
