// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCanaryByName(ctx context.Context, conn *synthetics.Synthetics, name string) (*synthetics.Canary, error) {
	input := &synthetics.GetCanaryInput{
		Name: aws.String(name),
	}

	output, err := conn.GetCanaryWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, synthetics.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Canary == nil || output.Canary.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Canary, nil
}

func FindGroupByName(ctx context.Context, conn *synthetics.Synthetics, name string) (*synthetics.Group, error) {
	input := &synthetics.GetGroupInput{
		GroupIdentifier: aws.String(name),
	}
	output, err := conn.GetGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, synthetics.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Group == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Group, nil
}

func FindAssociatedGroup(ctx context.Context, conn *synthetics.Synthetics, canaryArn string, groupName string) (*synthetics.GroupSummary, error) {
	input := &synthetics.ListAssociatedGroupsInput{
		ResourceArn: aws.String(canaryArn),
	}
	out, err := conn.ListAssociatedGroupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, synthetics.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Groups == nil || len(out.Groups) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	var group *synthetics.GroupSummary
	for _, groupSummary := range out.Groups {
		if aws.StringValue(groupSummary.Name) == groupName {
			group = groupSummary
		}
	}

	if group == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return group, nil
}
