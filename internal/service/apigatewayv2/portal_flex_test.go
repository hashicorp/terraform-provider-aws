// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestPortalEndpointConfigurationModelFlatten(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const (
		certificateARN = "arn:aws:acm:us-east-1:123456789012:certificate/00000000-0000-0000-0000-000000000000" //lintignore:AWSAT003,AWSAT005
		domainName     = "portal.example.com"
	)

	testCases := map[string]struct {
		input   any
		want    portalEndpointConfigurationModel
		wantErr bool
	}{
		"default domain": {
			input: awstypes.EndpointConfigurationResponse{
				PortalDefaultDomainName:  aws.String("p-abcdefghij.apigw-portal.us-east-1.on.aws"), //lintignore:AWSAT003
				PortalDomainHostedZoneId: aws.String("Z0000000000000000000"),
			},
			want: portalEndpointConfigurationModel{
				ACMManaged: fwtypes.NewListNestedObjectValueOfNull[portalACMManagedModel](ctx),
				None:       fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &portalNoneModel{}),
			},
		},
		"ACM managed": {
			input: &awstypes.EndpointConfigurationResponse{
				CertificateArn:           aws.String(certificateARN),
				DomainName:               aws.String(domainName),
				PortalDefaultDomainName:  aws.String("p-abcdefghij.apigw-portal.us-east-1.on.aws"), //lintignore:AWSAT003
				PortalDomainHostedZoneId: aws.String("Z0000000000000000000"),
			},
			want: portalEndpointConfigurationModel{
				ACMManaged: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &portalACMManagedModel{
					CertificateARN: fwtypes.ARNValue(certificateARN),
					DomainName:     types.StringValue(domainName),
				}),
				None: fwtypes.NewListNestedObjectValueOfNull[portalNoneModel](ctx),
			},
		},
		// Both fields are required for an ACM-managed endpoint. Half a response must
		// not be flattened into an empty, unparseable certificate ARN or domain name.
		"certificate without domain name": {
			input: &awstypes.EndpointConfigurationResponse{
				CertificateArn: aws.String(certificateARN),
			},
			wantErr: true,
		},
		"domain name without certificate": {
			input: &awstypes.EndpointConfigurationResponse{
				DomainName: aws.String(domainName),
			},
			wantErr: true,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got portalEndpointConfigurationModel
			diags := got.Flatten(ctx, testCase.input)

			if hasError := diags.HasError(); hasError != testCase.wantErr {
				t.Fatalf("HasError() = %t, want %t: %v", hasError, testCase.wantErr, diags)
			}
			if testCase.wantErr {
				return
			}

			if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("unexpected diff (+want, -got): %s", diff)
			}
		})
	}
}
