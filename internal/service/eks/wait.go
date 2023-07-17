// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	clusterDeleteRetryTimeout = 60 * time.Minute
)

func waitAddonCreated(ctx context.Context, conn *eks.EKS, clusterName, addonName string, timeout time.Duration) (*eks.Addon, error) {
	stateConf := retry.StateChangeConf{
		Pending: []string{eks.AddonStatusCreating, eks.AddonStatusDegraded},
		Target:  []string{eks.AddonStatusActive},
		Refresh: statusAddon(ctx, conn, clusterName, addonName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Addon); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.AddonStatusCreateFailed && health != nil {
			tfresource.SetLastError(err, AddonIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitAddonDeleted(ctx context.Context, conn *eks.EKS, clusterName, addonName string, timeout time.Duration) (*eks.Addon, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eks.AddonStatusActive, eks.AddonStatusDeleting},
		Target:  []string{},
		Refresh: statusAddon(ctx, conn, clusterName, addonName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Addon); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.AddonStatusDeleteFailed && health != nil {
			tfresource.SetLastError(err, AddonIssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitAddonUpdateSuccessful(ctx context.Context, conn *eks.EKS, clusterName, addonName, id string, timeout time.Duration) (*eks.Update, error) {
	stateConf := retry.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target:  []string{eks.UpdateStatusSuccessful},
		Refresh: statusAddonUpdate(ctx, conn, clusterName, addonName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Update); ok {
		if status := aws.StringValue(output.Status); status == eks.UpdateStatusCancelled || status == eks.UpdateStatusFailed {
			tfresource.SetLastError(err, ErrorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func waitFargateProfileCreated(ctx context.Context, conn *eks.EKS, clusterName, fargateProfileName string, timeout time.Duration) (*eks.FargateProfile, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eks.FargateProfileStatusCreating},
		Target:  []string{eks.FargateProfileStatusActive},
		Refresh: statusFargateProfile(ctx, conn, clusterName, fargateProfileName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.FargateProfile); ok {
		return output, err
	}

	return nil, err
}

func waitFargateProfileDeleted(ctx context.Context, conn *eks.EKS, clusterName, fargateProfileName string, timeout time.Duration) (*eks.FargateProfile, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eks.FargateProfileStatusActive, eks.FargateProfileStatusDeleting},
		Target:  []string{},
		Refresh: statusFargateProfile(ctx, conn, clusterName, fargateProfileName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.FargateProfile); ok {
		return output, err
	}

	return nil, err
}

func waitNodegroupCreated(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName string, timeout time.Duration) (*eks.Nodegroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eks.NodegroupStatusCreating},
		Target:  []string{eks.NodegroupStatusActive},
		Refresh: statusNodegroup(ctx, conn, clusterName, nodeGroupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Nodegroup); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.NodegroupStatusCreateFailed && health != nil {
			tfresource.SetLastError(err, IssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitNodegroupDeleted(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName string, timeout time.Duration) (*eks.Nodegroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{eks.NodegroupStatusActive, eks.NodegroupStatusDeleting},
		Target:  []string{},
		Refresh: statusNodegroup(ctx, conn, clusterName, nodeGroupName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Nodegroup); ok {
		if status, health := aws.StringValue(output.Status), output.Health; status == eks.NodegroupStatusDeleteFailed && health != nil {
			tfresource.SetLastError(err, IssuesError(health.Issues))
		}

		return output, err
	}

	return nil, err
}

func waitNodegroupUpdateSuccessful(ctx context.Context, conn *eks.EKS, clusterName, nodeGroupName, id string, timeout time.Duration) (*eks.Update, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{eks.UpdateStatusInProgress},
		Target:  []string{eks.UpdateStatusSuccessful},
		Refresh: statusNodegroupUpdate(ctx, conn, clusterName, nodeGroupName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.Update); ok {
		if status := aws.StringValue(output.Status); status == eks.UpdateStatusCancelled || status == eks.UpdateStatusFailed {
			tfresource.SetLastError(err, ErrorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func waitOIDCIdentityProviderConfigCreated(ctx context.Context, conn *eks.EKS, clusterName, configName string, timeout time.Duration) (*eks.OidcIdentityProviderConfig, error) {
	stateConf := retry.StateChangeConf{
		Pending: []string{eks.ConfigStatusCreating},
		Target:  []string{eks.ConfigStatusActive},
		Refresh: statusOIDCIdentityProviderConfig(ctx, conn, clusterName, configName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.OidcIdentityProviderConfig); ok {
		return output, err
	}

	return nil, err
}

func waitOIDCIdentityProviderConfigDeleted(ctx context.Context, conn *eks.EKS, clusterName, configName string, timeout time.Duration) (*eks.OidcIdentityProviderConfig, error) {
	stateConf := retry.StateChangeConf{
		Pending: []string{eks.ConfigStatusActive, eks.ConfigStatusDeleting},
		Target:  []string{},
		Refresh: statusOIDCIdentityProviderConfig(ctx, conn, clusterName, configName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*eks.OidcIdentityProviderConfig); ok {
		return output, err
	}

	return nil, err
}
