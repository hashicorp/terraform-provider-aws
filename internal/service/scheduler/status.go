// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/scheduler"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	scheduleGroupStatusActive   = "ACTIVE"
	scheduleGroupStatusDeleting = "DELETING"
)

func statusScheduleGroup(ctx context.Context, conn *scheduler.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findScheduleGroupByName(ctx, conn, name)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}
