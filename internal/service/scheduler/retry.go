// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	iamPropagationTimeout = 2 * time.Minute
)

func retryWhenIAMNotPropagated[T any](ctx context.Context, f func() (T, error)) (T, error) {
	v, err := tfresource.RetryWhen(
		ctx,
		iamPropagationTimeout,
		func() (interface{}, error) {
			return f()
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*types.ValidationException](err, "The execution role you provide must allow AWS EventBridge Scheduler to assume the role.") {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		var zero T
		return zero, err
	}

	return v.(T), nil
}
