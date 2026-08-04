// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func acceleratedRecoveryStatus(conn *route53.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findHostedZoneByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.HostedZone.Features.AcceleratedRecoveryStatus), nil
	}
}

func waitUpdateAcceleratedRecoveryCompleted(ctx context.Context, conn *route53.Client, id string, timeout time.Duration) (*awstypes.AcceleratedRecoveryStatus, error) {
	// Route53 is vulnerable to throttling so a longer delay and poll interval helps to avoid it.
	const (
		delay        = 15 * time.Second
		minTimeout   = 5 * time.Second
		pollInterval = 15 * time.Second
	)
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.AcceleratedRecoveryStatusEnabling, awstypes.AcceleratedRecoveryStatusEnablingHostedZoneLocked, awstypes.AcceleratedRecoveryStatusDisabling, awstypes.AcceleratedRecoveryStatusDisablingHostedZoneLocked),
		Target:       enum.Slice(awstypes.AcceleratedRecoveryStatusEnabled, awstypes.AcceleratedRecoveryStatusDisabled),
		Refresh:      acceleratedRecoveryStatus(conn, id),
		Delay:        delay,
		MinTimeout:   minTimeout,
		PollInterval: pollInterval,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AcceleratedRecoveryStatus); ok {
		return output, err
	}

	return nil, err
}
