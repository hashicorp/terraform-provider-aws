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

func FindCachePolicyByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetCachePolicyOutput, error) {
	input := &cloudfront.GetCachePolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetCachePolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchCachePolicy) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.CachePolicy == nil || output.CachePolicy.CachePolicyConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindFieldLevelEncryptionConfigByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetFieldLevelEncryptionConfigOutput, error) {
	input := &cloudfront.GetFieldLevelEncryptionConfigInput{
		Id: aws.String(id),
	}

	output, err := conn.GetFieldLevelEncryptionConfigWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFieldLevelEncryptionConfig) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FieldLevelEncryptionConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindFieldLevelEncryptionProfileByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetFieldLevelEncryptionProfileOutput, error) {
	input := &cloudfront.GetFieldLevelEncryptionProfileInput{
		Id: aws.String(id),
	}

	output, err := conn.GetFieldLevelEncryptionProfileWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFieldLevelEncryptionProfile) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FieldLevelEncryptionProfile == nil || output.FieldLevelEncryptionProfile.FieldLevelEncryptionProfileConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindFunctionByNameAndStage(ctx context.Context, conn *cloudfront.CloudFront, name, stage string) (*cloudfront.DescribeFunctionOutput, error) {
	input := &cloudfront.DescribeFunctionInput{
		Name:  aws.String(name),
		Stage: aws.String(stage),
	}

	output, err := conn.DescribeFunctionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchFunctionExists) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FunctionSummary == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindMonitoringSubscriptionByDistributionID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetMonitoringSubscriptionOutput, error) {
	input := &cloudfront.GetMonitoringSubscriptionInput{
		DistributionId: aws.String(id),
	}

	output, err := conn.GetMonitoringSubscriptionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchDistribution) {
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

	return output, nil
}

func FindOriginRequestPolicyByID(ctx context.Context, conn *cloudfront.CloudFront, id string) (*cloudfront.GetOriginRequestPolicyOutput, error) {
	input := &cloudfront.GetOriginRequestPolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetOriginRequestPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchOriginRequestPolicy) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.OriginRequestPolicy == nil || output.OriginRequestPolicy.OriginRequestPolicyConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

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
