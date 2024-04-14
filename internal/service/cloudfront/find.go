// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindCachePolicyByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetCachePolicyOutput, error) {
	input := &cloudfront.GetCachePolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetCachePolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchCachePolicy](err) {
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

func FindFieldLevelEncryptionConfigByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetFieldLevelEncryptionConfigOutput, error) {
	input := &cloudfront.GetFieldLevelEncryptionConfigInput{
		Id: aws.String(id),
	}

	output, err := conn.GetFieldLevelEncryptionConfig(ctx, input)

	if errs.IsA[*awstypes.NoSuchCachePolicy](err) {
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

func FindFieldLevelEncryptionProfileByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetFieldLevelEncryptionProfileOutput, error) {
	input := &cloudfront.GetFieldLevelEncryptionProfileInput{
		Id: aws.String(id),
	}

	output, err := conn.GetFieldLevelEncryptionProfile(ctx, input)

	if errs.IsA[*awstypes.NoSuchFieldLevelEncryptionProfile](err) {
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

func FindMonitoringSubscriptionByDistributionID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetMonitoringSubscriptionOutput, error) {
	input := &cloudfront.GetMonitoringSubscriptionInput{
		DistributionId: aws.String(id),
	}

	output, err := conn.GetMonitoringSubscription(ctx, input)

	if errs.IsA[*awstypes.NoSuchDistribution](err) || errs.IsA[*awstypes.NoSuchMonitoringSubscription](err) {
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

func FindOriginRequestPolicyByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetOriginRequestPolicyOutput, error) {
	input := &cloudfront.GetOriginRequestPolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetOriginRequestPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchOriginRequestPolicy](err) {
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

func FindRealtimeLogConfigByARN(ctx context.Context, conn *cloudfront.Client, arn string) (*awstypes.RealtimeLogConfig, error) {
	input := &cloudfront.GetRealtimeLogConfigInput{
		ARN: aws.String(arn),
	}

	output, err := conn.GetRealtimeLogConfig(ctx, input)

	if errs.IsA[*awstypes.NoSuchRealtimeLogConfig](err) {
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

func FindRealtimeLogConfigByName(ctx context.Context, conn *cloudfront.Client, name string) (*awstypes.RealtimeLogConfig, error) {
	input := &cloudfront.GetRealtimeLogConfigInput{
		Name: aws.String(name),
	}

	output, err := conn.GetRealtimeLogConfig(ctx, input)

	if errs.IsA[*awstypes.NoSuchRealtimeLogConfig](err) {
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

func FindResponseHeadersPolicyByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetResponseHeadersPolicyOutput, error) {
	input := &cloudfront.GetResponseHeadersPolicyInput{
		Id: aws.String(id),
	}

	output, err := conn.GetResponseHeadersPolicy(ctx, input)

	if errs.IsA[*awstypes.NoSuchResponseHeadersPolicy](err) {
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
