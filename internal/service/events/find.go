// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindTargetByThreePartKey(ctx context.Context, conn *eventbridge.EventBridge, busName, ruleName, targetID string) (*eventbridge.Target, error) {
	input := &eventbridge.ListTargetsByRuleInput{
		Rule:  aws.String(ruleName),
		Limit: aws.Int64(100), // Set limit to allowed maximum to prevent API throttling
	}

	if busName != "" {
		input.EventBusName = aws.String(busName)
	}

	var output *eventbridge.Target

	err := listTargetsByRulePages(ctx, conn, input, func(page *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Targets {
			if targetID == aws.StringValue(v.Id) {
				output = v
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, "ValidationException", eventbridge.ErrCodeResourceNotFoundException) || (err != nil && regexache.MustCompile(" not found$").MatchString(err.Error())) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}
