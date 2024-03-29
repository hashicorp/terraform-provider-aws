// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

// Custom CloudFront listing functions using similar formatting as other service generated code.

func ListCachePoliciesPages(ctx context.Context, conn *cloudfront.CloudFront, input *cloudfront.ListCachePoliciesInput, fn func(*cloudfront.ListCachePoliciesOutput, bool) bool) error {
	for {
		output, err := conn.ListCachePoliciesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.CachePolicyList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.CachePolicyList.NextMarker
	}
	return nil
}

func ListFieldLevelEncryptionConfigsPages(ctx context.Context, conn *cloudfront.CloudFront, input *cloudfront.ListFieldLevelEncryptionConfigsInput, fn func(*cloudfront.ListFieldLevelEncryptionConfigsOutput, bool) bool) error {
	for {
		output, err := conn.ListFieldLevelEncryptionConfigsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.FieldLevelEncryptionList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.FieldLevelEncryptionList.NextMarker
	}
	return nil
}

func ListFieldLevelEncryptionProfilesPages(ctx context.Context, conn *cloudfront.CloudFront, input *cloudfront.ListFieldLevelEncryptionProfilesInput, fn func(*cloudfront.ListFieldLevelEncryptionProfilesOutput, bool) bool) error {
	for {
		output, err := conn.ListFieldLevelEncryptionProfilesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.FieldLevelEncryptionProfileList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.FieldLevelEncryptionProfileList.NextMarker
	}
	return nil
}

func ListFunctionsPages(ctx context.Context, conn *cloudfront.CloudFront, input *cloudfront.ListFunctionsInput, fn func(*cloudfront.ListFunctionsOutput, bool) bool) error {
	for {
		output, err := conn.ListFunctionsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.FunctionList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.FunctionList.NextMarker
	}
	return nil
}

func ListOriginRequestPoliciesPages(ctx context.Context, conn *cloudfront.CloudFront, input *cloudfront.ListOriginRequestPoliciesInput, fn func(*cloudfront.ListOriginRequestPoliciesOutput, bool) bool) error {
	for {
		output, err := conn.ListOriginRequestPoliciesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.OriginRequestPolicyList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.OriginRequestPolicyList.NextMarker
	}
	return nil
}

func ListResponseHeadersPoliciesPages(ctx context.Context, conn *cloudfront.CloudFront, input *cloudfront.ListResponseHeadersPoliciesInput, fn func(*cloudfront.ListResponseHeadersPoliciesOutput, bool) bool) error {
	for {
		output, err := conn.ListResponseHeadersPoliciesWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.ResponseHeadersPolicyList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.ResponseHeadersPolicyList.NextMarker
	}
	return nil
}

func ListOriginAccessControlsPages(ctx context.Context, conn *cloudfront.CloudFront, input *cloudfront.ListOriginAccessControlsInput, fn func(*cloudfront.ListOriginAccessControlsOutput, bool) bool) error {
	for {
		output, err := conn.ListOriginAccessControlsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.OriginAccessControlList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.OriginAccessControlList.NextMarker
	}
	return nil
}
