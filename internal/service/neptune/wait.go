// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// Maximum amount of time to wait for an DBClusterEndpoint to return Available
	DBClusterEndpointAvailableTimeout = 10 * time.Minute

	// Maximum amount of time to wait for an DBClusterEndpoint to return Deleted
	DBClusterEndpointDeletedTimeout = 10 * time.Minute
)

// WaitDBClusterEndpointAvailable waits for a DBClusterEndpoint to return Available
func WaitDBClusterEndpointAvailable(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"creating", "modifying"},
		Target:  []string{"available"},
		Refresh: StatusDBClusterEndpoint(ctx, conn, id),
		Timeout: DBClusterEndpointAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*neptune.DBClusterEndpoint); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterEndpointDeleted waits for a DBClusterEndpoint to return Deleted
func WaitDBClusterEndpointDeleted(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{},
		Refresh: StatusDBClusterEndpoint(ctx, conn, id),
		Timeout: DBClusterEndpointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*neptune.DBClusterEndpoint); ok {
		return v, err
	}

	return nil, err
}
