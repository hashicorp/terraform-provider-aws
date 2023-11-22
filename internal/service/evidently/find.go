// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/evidently"
	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindFeatureWithProjectNameorARN(ctx context.Context, conn *evidently.Client, featureName, projectNameOrARN string) (*awstypes.Feature, error) {
	input := &evidently.GetFeatureInput{
		Feature: aws.String(featureName),
		Project: aws.String(projectNameOrARN),
	}

	output, err := conn.GetFeature(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Feature == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Feature, nil
}

func FindLaunchWithProjectNameorARN(ctx context.Context, conn *evidently.Client, launchName, projectNameOrARN string) (*awstypes.Launch, error) {
	input := &evidently.GetLaunchInput{
		Launch:  aws.String(launchName),
		Project: aws.String(projectNameOrARN),
	}

	output, err := conn.GetLaunch(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Launch == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Launch, nil
}

func FindProjectByNameOrARN(ctx context.Context, conn *evidently.Client, nameOrARN string) (*awstypes.Project, error) {
	input := &evidently.GetProjectInput{
		Project: aws.String(nameOrARN),
	}

	output, err := conn.GetProject(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Project == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Project, nil
}

func FindSegmentByNameOrARN(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string) (*cloudwatchevidently.Segment, error) {
	input := &cloudwatchevidently.GetSegmentInput{
		Segment: aws.String(nameOrARN),
	}

	output, err := conn.GetSegmentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Segment == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Segment, nil
}
