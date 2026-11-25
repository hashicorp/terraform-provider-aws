// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockcrud

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/mockcrud"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusFoo(conn *mockcrud.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFooByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
