// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshiftserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findNamespaceByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Namespace, error) {
	input := &redshiftserverless.GetNamespaceInput{
		NamespaceName: aws.String(name),
	}

	output, err := conn.GetNamespaceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Namespace, nil
}

func findEndpointAccessByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.EndpointAccess, error) {
	input := &redshiftserverless.GetEndpointAccessInput{
		EndpointName: aws.String(name),
	}

	output, err := conn.GetEndpointAccessWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, redshiftserverless.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Endpoint, nil
}

func findUsageLimitByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, id string) (*redshiftserverless.UsageLimit, error) {
	input := &redshiftserverless.GetUsageLimitInput{
		UsageLimitId: aws.String(id),
	}

	output, err := conn.GetUsageLimitWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, redshiftserverless.ErrCodeValidationException, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.UsageLimit, nil
}

func findSnapshotByName(ctx context.Context, conn *redshiftserverless.RedshiftServerless, name string) (*redshiftserverless.Snapshot, error) {
	input := &redshiftserverless.GetSnapshotInput{
		SnapshotName: aws.String(name),
	}

	output, err := conn.GetSnapshotWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, redshiftserverless.ErrCodeResourceNotFoundException, "snapshot") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Snapshot, nil
}

func findResourcePolicyByARN(ctx context.Context, conn *redshiftserverless.RedshiftServerless, arn string) (*redshiftserverless.ResourcePolicy, error) {
	input := &redshiftserverless.GetResourcePolicyInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetResourcePolicyWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, redshiftserverless.ErrCodeResourceNotFoundException, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResourcePolicy, nil
}
