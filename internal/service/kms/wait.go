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
	// Maximum amount of time to wait for StatusKeyState to return PendingDeletion
	KeyStatePendingDeletionTimeout = 20 * time.Minute

	keyRotationUpdatedTimeout = 10 * time.Minute
	KeyTagsPropagationTimeout = 10 * time.Minute

	PropagationTimeout = 2 * time.Minute

	ReplicaExternalKeyCreatedTimeout = 2 * time.Minute
	ReplicaKeyCreatedTimeout         = 2 * time.Minute
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
