// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindRealtimeLogConfigByARN(ctx context.Context, conn *cloudfront.CloudFront, arn string) (*cloudfront.RealtimeLogConfig, error) {
	input := &cloudfront.GetRealtimeLogConfigInput{
		ARN: aws.String(arn),
	}

	output, err := conn.GetRealtimeLogConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchRealtimeLogConfig) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RealtimeLogConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RealtimeLogConfig, nil
}

func FindRealtimeLogConfigByName(ctx context.Context, conn *cloudfront.CloudFront, name string) (*cloudfront.RealtimeLogConfig, error) {
	input := &cloudfront.GetRealtimeLogConfigInput{
		Name: aws.String(name),
	}

	output, err := conn.GetRealtimeLogConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchRealtimeLogConfig) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RealtimeLogConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RealtimeLogConfig, nil
}

func FindResponseHeadersPolicyByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetResponseHeadersPolicyOutput, error) {
	input := &cloudfront.GetResponseHeadersPolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetResponseHeadersPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchResponseHeadersPolicy) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResponseHeadersPolicy == nil || output.ResponseHeadersPolicy.ResponseHeadersPolicyConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
