// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appflow

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	FlowCreationTimeout = 2 * time.Minute
	FlowDeletionTimeout = 2 * time.Minute
)

func FlowDeleted(ctx context.Context, conn *appflow.Appflow, id string) error {
	stateConf := &retry.StateChangeConf{
		Target:  []string{},
		Refresh: FlowStatus(ctx, conn, id),
		Timeout: FlowDeletionTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
