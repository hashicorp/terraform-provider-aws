// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package examplecrud

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/examplecrud"
	awstypes "github.com/aws/aws-sdk-go-v2/service/examplecrud/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func waitWidgetCreated(ctx context.Context, conn *examplecrud.Client, id string, timeout time.Duration) (*awstypes.Widget, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WidgetStatusCreating),
		Target:  enum.Slice(awstypes.WidgetStatusAvailable),
		Refresh: statusWidget(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Widget); ok {
		return output, err
	}

	return nil, err
}

func waitWidgetDeleted(ctx context.Context, conn *examplecrud.Client, id string, timeout time.Duration) (*awstypes.Widget, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WidgetStatusAvailable, awstypes.WidgetStatusDeleting),
		Target:  []string{},
		Refresh: statusWidget(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Widget); ok {
		return output, err
	}

	return nil, err
}
