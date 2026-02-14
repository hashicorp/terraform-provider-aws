// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lakeformation"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lakeformation/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

const (
	permissionsReadyTimeout       = 1 * time.Minute
	permissionsDeleteRetryTimeout = 30 * time.Second

	statusAvailable = "AVAILABLE"
	statusIAMDelay  = "IAM DELAY"
)

func waitPermissionsReady(ctx context.Context, conn *lakeformation.Client, input *lakeformation.ListPermissionsInput, filter PermissionsFilter) ([]awstypes.PrincipalResourcePermissions, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusIAMDelay},
		Target:  []string{statusAvailable},
		Refresh: statusPermissions(conn, input, filter),
		Timeout: permissionsReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.([]awstypes.PrincipalResourcePermissions); ok {
		return output, err
	}

	return nil, err
}
