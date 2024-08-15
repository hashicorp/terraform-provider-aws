// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"time"

	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	keyRotationUpdatedTimeout = 10 * time.Minute

	// General timeout for KMS resource changes to propagate.
	// See https://docs.aws.amazon.com/kms/latest/developerguide/programming-eventual-consistency.html
	propagationTimeout = 3 * time.Minute // nosemgrep:ci.kms-in-const-name, ci.kms-in-var-name

	// General timeout for IAM resource change to propagate.
	// See https://docs.aws.amazon.com/IAM/latest/UserGuide/troubleshoot_general.html#troubleshoot_general_eventual-consistency.
	// We have settled on 2 minutes as the best timeout value.
	iamPropagationTimeout = 2 * time.Minute
)

// waitIAMPropagation retries the specified function if the returned error indicates an IAM eventual consistency issue.
func waitIAMPropagation[T any](ctx context.Context, timeout time.Duration, f func() (T, error)) (T, error) {
	outputRaw, err := tfresource.RetryWhenIsA[*awstypes.MalformedPolicyDocumentException](ctx, timeout, func() (interface{}, error) {
		return f()
	})

	if err != nil {
		var zero T
		return zero, err
	}

	return outputRaw.(T), nil
}
