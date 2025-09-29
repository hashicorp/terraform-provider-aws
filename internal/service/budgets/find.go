// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package budgets

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBudgetWithDelay[T any](ctx context.Context, f func() (T, error)) (T, error) {
	var resp T
	err := tfresource.Retry(ctx, 30*time.Second, func(ctx context.Context) *tfresource.RetryError {
		var err error
		resp, err = f()

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	}, tfresource.WithDelay(5*time.Second))

	return resp, err
}
