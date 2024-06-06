// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findChangeByID(ctx context.Context, conn *route53.Client, id string) (*awstypes.ChangeInfo, error) {
	input := &route53.GetChangeInput{
		Id: aws.String(id),
	}

	output, err := conn.GetChange(ctx, input)

	if errs.IsA[*awstypes.NoSuchChange](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ChangeInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ChangeInfo, nil
}

func statusChange(ctx context.Context, conn *route53.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findChangeByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitChangeInsync(ctx context.Context, conn *route53.Client, id string) (*awstypes.ChangeInfo, error) {
	// Route53 is vulnerable to throttling so longer delays, poll intervals helps significantly to avoid.
	const (
		timeout      = 30 * time.Minute
		minTimeout   = 5 * time.Second
		pollInterval = 15 * time.Second
		minDelay     = 10
		maxDelay     = 30
	)
	stateConf := &retry.StateChangeConf{
		Pending:      enum.Slice(awstypes.ChangeStatusPending),
		Target:       enum.Slice(awstypes.ChangeStatusInsync),
		Refresh:      statusChange(ctx, conn, id),
		Delay:        time.Duration(rand.Int63n(maxDelay-minDelay)+minDelay) * time.Second,
		MinTimeout:   minTimeout,
		PollInterval: pollInterval,
		Timeout:      timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ChangeInfo); ok {
		return output, err
	}

	return nil, err
}
