// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findWithDelay[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	var resp T
	err := tfresource.Retry(ctx, budgetsPropagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		var err error
		resp, err = f(ctx)

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	}, tfresource.WithDelay(5*time.Second))

	return resp, err
}
