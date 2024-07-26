// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusCanaryState(ctx context.Context, conn *synthetics.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCanaryByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.State), nil
	}
}
