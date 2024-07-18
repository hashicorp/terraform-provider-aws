// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"errors"

	guardduty_v2 "github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindMalwareProtectionPlanByID(ctx context.Context, conn *guardduty_v2.Client, id string) (*guardduty_v2.GetMalwareProtectionPlanOutput, error) {
	input := &guardduty_v2.GetMalwareProtectionPlanInput{
		MalwareProtectionPlanId: aws.String(id),
	}

	return FindMalwareProtectionPlan(ctx, conn, input)
}

func FindMalwareProtectionPlan(ctx context.Context, conn *guardduty_v2.Client, input *guardduty_v2.GetMalwareProtectionPlanInput) (*guardduty_v2.GetMalwareProtectionPlanOutput, error) {
	result, err := conn.GetMalwareProtectionPlan(ctx, input)

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if result == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return result, nil
}
