// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloud9

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findEnvironment(ctx context.Context, conn *cloud9.Cloud9, input *cloud9.DescribeEnvironmentsInput) (*cloud9.Environment, error) {
	output, err := findEnvironments(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil || output[0].Lifecycle == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func findEnvironments(ctx context.Context, conn *cloud9.Cloud9, input *cloud9.DescribeEnvironmentsInput) ([]*cloud9.Environment, error) {
	output, err := conn.DescribeEnvironmentsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloud9.ErrCodeNotFoundException) {
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

	return output.Environments, nil
}

func FindEnvironmentByID(ctx context.Context, conn *cloud9.Cloud9, id string) (*cloud9.Environment, error) {
	input := &cloud9.DescribeEnvironmentsInput{
		EnvironmentIds: aws.StringSlice([]string{id}),
	}

	output, err := findEnvironment(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.Id) != id {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func FindEnvironmentMembershipByID(ctx context.Context, conn *cloud9.Cloud9, envId, userArn string) (*cloud9.EnvironmentMember, error) {
	input := &cloud9.DescribeEnvironmentMembershipsInput{
		EnvironmentId: aws.String(envId),
		UserArn:       aws.String(userArn),
	}
	out, err := conn.DescribeEnvironmentMembershipsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloud9.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	envs := out.Memberships

	if len(envs) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	env := envs[0]

	if env == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return env, nil
}
