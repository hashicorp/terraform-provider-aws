// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for StatusKeyState to return PendingDeletion
	KeyStatePendingDeletionTimeout = 20 * time.Minute

	KeyDeletedTimeout                = 20 * time.Minute
	KeyDescriptionPropagationTimeout = 10 * time.Minute
	KeyMaterialImportedTimeout       = 10 * time.Minute
	KeyPolicyPropagationTimeout      = 10 * time.Minute
	KeyRotationUpdatedTimeout        = 10 * time.Minute
	KeyStatePropagationTimeout       = 20 * time.Minute
	KeyTagsPropagationTimeout        = 10 * time.Minute
	KeyValidToPropagationTimeout     = 5 * time.Minute

	PropagationTimeout = 2 * time.Minute

	ReplicaExternalKeyCreatedTimeout = 2 * time.Minute
	ReplicaKeyCreatedTimeout         = 2 * time.Minute
)

// WaitIAMPropagation retries the specified function if the returned error indicates an IAM eventual consistency issue.
// If the retries time out the specified function is called one last time.
func WaitIAMPropagation[T any](ctx context.Context, f func() (T, error)) (T, error) {
	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return f()
	},
		kms.ErrCodeMalformedPolicyDocumentException)

	if err != nil {
		var zero T
		return zero, err
	}

	return outputRaw.(T), nil
}

func WaitKeyDeleted(ctx context.Context, conn *kms.KMS, id string) (*kms.KeyMetadata, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kms.KeyStateDisabled, kms.KeyStateEnabled},
		Target:  []string{},
		Refresh: StatusKeyState(ctx, conn, id),
		Timeout: KeyDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kms.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}

func WaitKeyDescriptionPropagated(ctx context.Context, conn *kms.KMS, id string, description string) error {
	checkFunc := func() (bool, error) {
		output, err := FindKeyByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.StringValue(output.Description) == description, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                2 * time.Second,
	}

	return tfresource.WaitUntil(ctx, KeyDescriptionPropagationTimeout, checkFunc, opts)
}

func WaitKeyMaterialImported(ctx context.Context, conn *kms.KMS, id string) (*kms.KeyMetadata, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kms.KeyStatePendingImport},
		Target:  []string{kms.KeyStateDisabled, kms.KeyStateEnabled},
		Refresh: StatusKeyState(ctx, conn, id),
		Timeout: KeyMaterialImportedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kms.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}

func WaitKeyPolicyPropagated(ctx context.Context, conn *kms.KMS, id, policy string) error {
	checkFunc := func() (bool, error) {
		output, err := FindKeyPolicyByKeyIDAndPolicyName(ctx, conn, id, PolicyNameDefault)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		equivalent, err := awspolicy.PoliciesAreEquivalent(aws.StringValue(output), policy)

		if err != nil {
			return false, err
		}

		return equivalent, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}

	return tfresource.WaitUntil(ctx, KeyPolicyPropagationTimeout, checkFunc, opts)
}

func WaitKeyRotationEnabledPropagated(ctx context.Context, conn *kms.KMS, id string, enabled bool) error {
	checkFunc := func() (bool, error) {
		output, err := FindKeyRotationEnabledByKeyID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.BoolValue(output) == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                1 * time.Second,
	}

	return tfresource.WaitUntil(ctx, KeyRotationUpdatedTimeout, checkFunc, opts)
}

func WaitKeyStatePropagated(ctx context.Context, conn *kms.KMS, id string, enabled bool) error {
	checkFunc := func() (bool, error) {
		output, err := FindKeyByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		return aws.BoolValue(output.Enabled) == enabled, nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 15,
		MinTimeout:                2 * time.Second,
	}

	return tfresource.WaitUntil(ctx, KeyStatePropagationTimeout, checkFunc, opts)
}

func WaitKeyValidToPropagated(ctx context.Context, conn *kms.KMS, id string, validTo string) error {
	checkFunc := func() (bool, error) {
		output, err := FindKeyByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return false, nil
		}

		if err != nil {
			return false, err
		}

		if output.ValidTo != nil {
			return aws.TimeValue(output.ValidTo).Format(time.RFC3339) == validTo, nil
		}

		return validTo == "", nil
	}
	opts := tfresource.WaitOpts{
		ContinuousTargetOccurence: 5,
		MinTimeout:                2 * time.Second,
	}

	return tfresource.WaitUntil(ctx, KeyValidToPropagationTimeout, checkFunc, opts)
}

func WaitReplicaExternalKeyCreated(ctx context.Context, conn *kms.KMS, id string) (*kms.KeyMetadata, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kms.KeyStateCreating},
		Target:  []string{kms.KeyStatePendingImport},
		Refresh: StatusKeyState(ctx, conn, id),
		Timeout: ReplicaExternalKeyCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kms.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}

func WaitReplicaKeyCreated(ctx context.Context, conn *kms.KMS, id string) (*kms.KeyMetadata, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kms.KeyStateCreating},
		Target:  []string{kms.KeyStateEnabled},
		Refresh: StatusKeyState(ctx, conn, id),
		Timeout: ReplicaKeyCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kms.KeyMetadata); ok {
		return output, err
	}

	return nil, err
}
