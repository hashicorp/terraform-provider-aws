// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/synthetics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/synthetics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

var errResourceNotFoundException = &awstypes.ResourceNotFoundException{}

func FindCanaryByName(ctx context.Context, conn *synthetics.Client, name string) (*awstypes.Canary, error) {
	input := &synthetics.GetCanaryInput{
		Name: aws.String(name),
	}

	output, err := conn.GetCanary(ctx, input)

	if err != nil {
		// error is not being asserted into type *awstypes.ResourceNotFoundException but has all the properties
		// of the error.
		if strings.Contains(err.Error(), errResourceNotFoundException.ErrorCode()) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if output == nil || output.Canary == nil || output.Canary.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Canary, nil
}

func FindGroupByName(ctx context.Context, conn *synthetics.Client, name string) (*awstypes.Group, error) {
	input := &synthetics.GetGroupInput{
		GroupIdentifier: aws.String(name),
	}
	output, err := conn.GetGroup(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func FindAssociatedGroup(ctx context.Context, conn *synthetics.Client, canaryArn string, groupName string) (*awstypes.GroupSummary, error) {
	input := &synthetics.ListAssociatedGroupsInput{
		ResourceArn: aws.String(canaryArn),
	}
	out, err := conn.ListAssociatedGroups(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

	var group awstypes.GroupSummary
	for _, groupSummary := range out.Groups {
		if aws.ToString(groupSummary.Name) == groupName {
			group = groupSummary
		}
	}

	if group == (awstypes.GroupSummary{}) {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return &group, nil
}
