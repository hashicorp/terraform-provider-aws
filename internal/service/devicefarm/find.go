// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/devicefarm"
	awstypes "github.com/aws/aws-sdk-go-v2/service/devicefarm/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDevicePoolByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.DevicePool, error) {
	input := &devicefarm.GetDevicePoolInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetDevicePool(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DevicePool == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DevicePool, nil
}

func FindProjectByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.Project, error) {
	input := &devicefarm.GetProjectInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetProject(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func FindUploadByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.Upload, error) {
	input := &devicefarm.GetUploadInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetUpload(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Upload == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Upload, nil
}

func FindNetworkProfileByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.NetworkProfile, error) {
	input := &devicefarm.GetNetworkProfileInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetNetworkProfile(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.NetworkProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.NetworkProfile, nil
}

func FindInstanceProfileByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.InstanceProfile, error) {
	input := &devicefarm.GetInstanceProfileInput{
		Arn: aws.String(arn),
	}
	output, err := conn.GetInstanceProfile(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.InstanceProfile == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.InstanceProfile, nil
}

func FindTestGridProjectByARN(ctx context.Context, conn *devicefarm.Client, arn string) (*awstypes.TestGridProject, error) {
	input := &devicefarm.GetTestGridProjectInput{
		ProjectArn: aws.String(arn),
	}
	output, err := conn.GetTestGridProject(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TestGridProject == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TestGridProject, nil
}
