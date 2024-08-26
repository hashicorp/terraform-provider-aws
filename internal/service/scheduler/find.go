// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findScheduleGroupByName(ctx context.Context, conn *scheduler.Client, name string) (*scheduler.GetScheduleGroupOutput, error) {
	in := &scheduler.GetScheduleGroupInput{
		Name: aws.String(name),
	}
	out, err := conn.GetScheduleGroup(ctx, in)
	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Arn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
