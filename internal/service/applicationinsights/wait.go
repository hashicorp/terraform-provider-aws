// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationinsights

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/applicationinsights"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	ApplicationCreatedTimeout = 2 * time.Minute

	ApplicationDeletedTimeout = 2 * time.Minute
)

func waitApplicationCreated(ctx context.Context, conn *applicationinsights.ApplicationInsights, name string) (*applicationinsights.ApplicationInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"CREATING"},
		Target:  []string{"NOT_CONFIGURED", "ACTIVE"},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: ApplicationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*applicationinsights.ApplicationInfo); ok {
		return output, err
	}

	return nil, err
}

func waitApplicationTerminated(ctx context.Context, conn *applicationinsights.ApplicationInsights, name string) (*applicationinsights.ApplicationInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"ACTIVE", "NOT_CONFIGURED", "DELETING"},
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, name),
		Timeout: ApplicationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*applicationinsights.ApplicationInfo); ok {
		return output, err
	}

	return nil, err
}
