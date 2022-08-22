package cloudfront

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
)

// Custom CloudFront listing functions using similar formatting as other service generated code.

func ListCachePoliciesPages(conn *cloudfront.CloudFront, input *cloudfront.ListCachePoliciesInput, fn func(*cloudfront.ListCachePoliciesOutput, bool) bool) error {
	for {
		output, err := conn.ListCachePolicies(input)
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

func ListFieldLevelEncryptionConfigsPages(conn *cloudfront.CloudFront, input *cloudfront.ListFieldLevelEncryptionConfigsInput, fn func(*cloudfront.ListFieldLevelEncryptionConfigsOutput, bool) bool) error {
	for {
		output, err := conn.ListFieldLevelEncryptionConfigs(input)
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

func ListFieldLevelEncryptionProfilesPages(conn *cloudfront.CloudFront, input *cloudfront.ListFieldLevelEncryptionProfilesInput, fn func(*cloudfront.ListFieldLevelEncryptionProfilesOutput, bool) bool) error {
	for {
		output, err := conn.ListFieldLevelEncryptionProfiles(input)
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

func ListFunctionsPages(conn *cloudfront.CloudFront, input *cloudfront.ListFunctionsInput, fn func(*cloudfront.ListFunctionsOutput, bool) bool) error {
	for {
		output, err := conn.ListFunctions(input)
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

func ListOriginRequestPoliciesPages(conn *cloudfront.CloudFront, input *cloudfront.ListOriginRequestPoliciesInput, fn func(*cloudfront.ListOriginRequestPoliciesOutput, bool) bool) error {
	for {
		output, err := conn.ListOriginRequestPolicies(input)
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

func ListResponseHeadersPoliciesPages(conn *cloudfront.CloudFront, input *cloudfront.ListResponseHeadersPoliciesInput, fn func(*cloudfront.ListResponseHeadersPoliciesOutput, bool) bool) error {
	for {
		output, err := conn.ListResponseHeadersPolicies(input)
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
