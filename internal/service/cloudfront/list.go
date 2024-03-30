// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// Custom CloudFront listing functions using similar formatting as other service generated code.

func ListCachePoliciesPages(ctx context.Context, input *cloudfront.ListCachePoliciesInput, fn func(*cloudfront.ListCachePoliciesOutput, bool) bool, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	for {
		output, err := conn.ListCachePolicies(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.CachePolicyList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.CachePolicyList.NextMarker
	}
	return nil
}

func ListFieldLevelEncryptionConfigsPages(ctx context.Context, input *cloudfront.ListFieldLevelEncryptionConfigsInput, fn func(*cloudfront.ListFieldLevelEncryptionConfigsOutput, bool) bool, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	for {
		output, err := conn.ListFieldLevelEncryptionConfigs(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.FieldLevelEncryptionList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.FieldLevelEncryptionList.NextMarker
	}
	return nil
}

func ListFieldLevelEncryptionProfilesPages(ctx context.Context, input *cloudfront.ListFieldLevelEncryptionProfilesInput, fn func(*cloudfront.ListFieldLevelEncryptionProfilesOutput, bool) bool, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	for {
		output, err := conn.ListFieldLevelEncryptionProfiles(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.FieldLevelEncryptionProfileList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.FieldLevelEncryptionProfileList.NextMarker
	}
	return nil
}

func ListFunctionsPages(ctx context.Context, input *cloudfront.ListFunctionsInput, fn func(*cloudfront.ListFunctionsOutput, bool) bool, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	for {
		output, err := conn.ListFunctions(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.FunctionList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.FunctionList.NextMarker
	}
	return nil
}

func ListOriginRequestPoliciesPages(ctx context.Context, input *cloudfront.ListOriginRequestPoliciesInput, fn func(*cloudfront.ListOriginRequestPoliciesOutput, bool) bool, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	for {
		output, err := conn.ListOriginRequestPolicies(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.OriginRequestPolicyList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.OriginRequestPolicyList.NextMarker
	}
	return nil
}

func ListResponseHeadersPoliciesPages(ctx context.Context, input *cloudfront.ListResponseHeadersPoliciesInput, fn func(*cloudfront.ListResponseHeadersPoliciesOutput, bool) bool, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	for {
		output, err := conn.ListResponseHeadersPolicies(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.ResponseHeadersPolicyList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.ResponseHeadersPolicyList.NextMarker
	}
	return nil
}

func ListOriginAccessControlsPages(ctx context.Context, input *cloudfront.ListOriginAccessControlsInput, fn func(*cloudfront.ListOriginAccessControlsOutput, bool) bool, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)
	for {
		output, err := conn.ListOriginAccessControls(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.OriginAccessControlList.NextMarker) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.Marker = output.OriginAccessControlList.NextMarker
	}
	return nil
}
