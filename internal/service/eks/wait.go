// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	clusterDeleteRetryTimeout = 60 * time.Minute
)

func waitAddonUpdateSuccessful(ctx context.Context, client *eks.Client, clusterName, addonName, id string, timeout time.Duration) (*types.Update, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(types.UpdateStatusInProgress),
		Target:  enum.Slice(types.UpdateStatusSuccessful),
		Refresh: statusAddonUpdate(ctx, client, clusterName, addonName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Update); ok {
		if status := output.Status; status == types.UpdateStatusCancelled || status == types.UpdateStatusFailed {
			tfresource.SetLastError(err, ErrorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func waitNodegroupUpdateSuccessful(ctx context.Context, client *eks.Client, clusterName, nodeGroupName, id string, timeout time.Duration) (*types.Update, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.UpdateStatusInProgress),
		Target:  enum.Slice(types.UpdateStatusSuccessful),
		Refresh: statusNodegroupUpdate(ctx, client, clusterName, nodeGroupName, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Update); ok {
		if status := output.Status; status == types.UpdateStatusCancelled || status == types.UpdateStatusFailed {
			tfresource.SetLastError(err, ErrorDetailsError(output.Errors))
		}

		return output, err
	}

	return nil, err
}

func waitOIDCIdentityProviderConfigCreated(ctx context.Context, client *eks.Client, clusterName, configName string, timeout time.Duration) (*types.OidcIdentityProviderConfig, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(types.ConfigStatusCreating),
		Target:  enum.Slice(types.ConfigStatusActive),
		Refresh: statusOIDCIdentityProviderConfig(ctx, client, clusterName, configName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OidcIdentityProviderConfig); ok {
		return output, err
	}

	return nil, err
}

func waitOIDCIdentityProviderConfigDeleted(ctx context.Context, client *eks.Client, clusterName, configName string, timeout time.Duration) (*types.OidcIdentityProviderConfig, error) {
	stateConf := retry.StateChangeConf{
		Pending: enum.Slice(types.ConfigStatusActive, types.ConfigStatusDeleting),
		Target:  []string{},
		Refresh: statusOIDCIdentityProviderConfig(ctx, client, clusterName, configName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.OidcIdentityProviderConfig); ok {
		return output, err
	}

	return nil, err
}
