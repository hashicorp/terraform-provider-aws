// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// AddErrorPredicateRetrier returns a Retryer which runs the specified predicate on any error.
// If the predicate returns `false` then the specified retrier is called.
func AddErrorPredicateRetrier(r aws_sdkv2.RetryerV2, predicate tfslices.Predicate[error]) aws_sdkv2.RetryerV2 {
	return &withErrorPredicate{
		RetryerV2: r,
		predicate: predicate,
	}
}

type withErrorPredicate struct {
	aws_sdkv2.RetryerV2
	predicate tfslices.Predicate[error]
}

func (r *withErrorPredicate) IsErrorRetryable(err error) bool {
	if r.predicate(err) {
		return true
	}
	return r.RetryerV2.IsErrorRetryable(err)
}
