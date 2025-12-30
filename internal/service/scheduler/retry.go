// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	iamPropagationTimeout = 2 * time.Minute
)

func retryWhenIAMNotPropagated[T any](ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	return tfresource.RetryWhenIsAErrorMessageContains[T, *awstypes.ValidationException](ctx, iamPropagationTimeout, f, "The execution role you provide must allow AWS EventBridge Scheduler to assume the role.")
}
