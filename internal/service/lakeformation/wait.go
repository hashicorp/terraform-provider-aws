// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	permissionsReadyTimeout       = 1 * time.Minute
	permissionsDeleteRetryTimeout = 30 * time.Second

	statusAvailable = "AVAILABLE"
	statusNotFound  = "NOT FOUND"
	statusFailed    = "FAILED"
	statusIAMDelay  = "IAM DELAY"
)

func waitPermissionsReady(ctx context.Context, conn *lakeformation.LakeFormation, input *lakeformation.ListPermissionsInput, tableType string, columnNames []*string, excludedColumnNames []*string, columnWildcard bool) ([]*lakeformation.PrincipalResourcePermissions, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusNotFound, statusIAMDelay},
		Target:  []string{statusAvailable},
		Refresh: statusPermissions(ctx, conn, input, tableType, columnNames, excludedColumnNames, columnWildcard),
		Timeout: permissionsReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]*lakeformation.PrincipalResourcePermissions); ok {
		return output, err
	}

	return nil, err
}
