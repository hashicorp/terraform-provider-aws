// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitInboundIntegrationCreated(ctx context.Context, conn *glue.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{"CREATING", "MODIFYING"},
		Target:     []string{"ACTIVE"},
		Refresh:    statusInboundIntegration(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Integration); ok {
		// surface terminal state text if present
		tfresource.SetLastError(err, errors.New(string(output.Status)))
		return output, err
	}

	return nil, err
}

func waitInboundIntegrationDeleted(ctx context.Context, conn *glue.Client, arn string, timeout time.Duration) (*awstypes.Integration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{"DELETING", "ACTIVE"},
		Target:  []string{},
		Refresh: statusInboundIntegration(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Integration); ok {
		return output, err
	}

	return nil, err
}
