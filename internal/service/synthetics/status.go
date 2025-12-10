// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusCanaryState(ctx context.Context, conn *synthetics.Client, name string) sdkretry.StateRefreshFunc {
	return func() (any, string, error) {
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
