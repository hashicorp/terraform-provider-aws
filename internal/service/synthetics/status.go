// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusCanaryState(conn *synthetics.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := FindCanaryByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.State), nil
	}
}
