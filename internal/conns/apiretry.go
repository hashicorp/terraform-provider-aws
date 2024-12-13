// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package conns

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
)

// AddIsErrorRetryables returns a Retryer which runs the specified retryables on any error.
func AddIsErrorRetryables(r aws.RetryerV2, retryables ...retry.IsErrorRetryable) aws.RetryerV2 {
	return &withIsErrorRetryables{
		RetryerV2:  r,
		retryables: retryables,
	}
}

type withIsErrorRetryables struct {
	aws.RetryerV2
	retryables retry.IsErrorRetryables
}

func (r *withIsErrorRetryables) IsErrorRetryable(err error) bool {
	if v := r.retryables.IsErrorRetryable(err); v != aws.UnknownTernary {
		return v.Bool()
	}
	return r.RetryerV2.IsErrorRetryable(err)
}
