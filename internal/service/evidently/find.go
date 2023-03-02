package evidently

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindFeatureWithProjectNameorARN(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, featureName, projectNameOrARN string) (*cloudwatchevidently.Feature, error) {
	input := &cloudwatchevidently.GetFeatureInput{
		Feature: aws.String(featureName),
		Project: aws.String(projectNameOrARN),
	}

	output, err := conn.GetFeatureWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindLaunchWithProjectNameorARN(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, launchName, projectNameOrARN string) (*cloudwatchevidently.Launch, error) {
	input := &cloudwatchevidently.GetLaunchInput{
		Launch:  aws.String(launchName),
		Project: aws.String(projectNameOrARN),
	}

	output, err := conn.GetLaunchWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func FindProjectByNameOrARN(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string) (*cloudwatchevidently.Project, error) {
	input := &cloudwatchevidently.GetProjectInput{
		Project: aws.String(nameOrARN),
	}

	output, err := conn.GetProjectWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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
