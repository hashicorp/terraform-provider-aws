// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func waitNamespaceDeleted(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Namespace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			redshiftserverless.NamespaceStatusDeleting,
		},
		Target:  []string{},
		Refresh: statusNamespace(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftserverless.Namespace); ok {
		return output, err
	}

	return nil, err
}

func waitNamespaceUpdated(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Namespace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			redshiftserverless.NamespaceStatusModifying,
		},
		Target: []string{
			redshiftserverless.NamespaceStatusAvailable,
		},
		Refresh: statusNamespace(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftserverless.Namespace); ok {
		return output, err
	}

	return nil, err
}

func waitEndpointAccessActive(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.EndpointAccess, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			"CREATING",
			"MODIFYING",
		},
		Target: []string{
			"ACTIVE",
		},
		Refresh: statusEndpointAccess(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftserverless.EndpointAccess); ok {
		return output, err
	}

	return nil, err
}

func waitEndpointAccessDeleted(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.EndpointAccess, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			"DELETING",
		},
		Target:  []string{},
		Refresh: statusEndpointAccess(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftserverless.EndpointAccess); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotAvailable(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			redshiftserverless.SnapshotStatusCreating,
		},
		Target: []string{
			redshiftserverless.SnapshotStatusAvailable,
		},
		Refresh: statusSnapshot(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftserverless.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			redshiftserverless.SnapshotStatusAvailable,
		},
		Target:  []string{},
		Refresh: statusSnapshot(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*redshiftserverless.Snapshot); ok {
		return output, err
	}

	return nil, err
}
