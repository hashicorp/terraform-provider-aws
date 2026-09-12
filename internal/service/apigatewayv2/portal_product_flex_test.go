// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

func TestPortalProductAutoFlexExpand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]struct {
		model    portalProductResourceModel
		expected apigatewayv2.CreatePortalProductInput
	}{
		"display name only": {
			model: portalProductResourceModel{
				DisplayName: types.StringValue("AdoptAnimals"),
				Description: types.StringNull(),
			},
			expected: apigatewayv2.CreatePortalProductInput{
				DisplayName: aws.String("AdoptAnimals"),
			},
		},
		"display name and description": {
			model: portalProductResourceModel{
				DisplayName: types.StringValue("AdoptAnimals"),
				Description: types.StringValue("Shelter animal APIs"),
			},
			expected: apigatewayv2.CreatePortalProductInput{
				DisplayName: aws.String("AdoptAnimals"),
				Description: aws.String("Shelter animal APIs"),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got apigatewayv2.CreatePortalProductInput
			if diags := fwflex.Expand(ctx, testCase.model, &got); diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}

			if diff := cmp.Diff(testCase.expected, got, cmpopts.IgnoreUnexported(apigatewayv2.CreatePortalProductInput{})); diff != "" {
				t.Errorf("unexpected diff (+want, -got): %s", diff)
			}
		})
	}
}

func TestPortalProductAutoFlexFlatten(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	lastModified := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)

	// DisplayOrder has no counterpart in portalProductResourceModel. Flatten iterates
	// source fields, so it must be skipped rather than raising a diagnostic.
	out := apigatewayv2.GetPortalProductOutput{
		Description:      aws.String("Shelter animal APIs"),
		DisplayName:      aws.String("AdoptAnimals"),
		DisplayOrder:     &awstypes.DisplayOrder{OverviewPageArn: aws.String("arn:aws:apigateway:us-west-2:123456789012:/portalproducts/abcdef1234/productpages/zyxwvu9876")}, //lintignore:AWSAT003,AWSAT005
		LastModified:     aws.Time(lastModified),
		PortalProductArn: aws.String("arn:aws:apigateway:us-west-2:123456789012:/portalproducts/abcdef1234"), //lintignore:AWSAT003,AWSAT005
		PortalProductId:  aws.String("abcdef1234"),
	}

	var got portalProductResourceModel
	if diags := fwflex.Flatten(ctx, &out, &got); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if want, got := "AdoptAnimals", got.DisplayName.ValueString(); got != want {
		t.Errorf("DisplayName = %q, want %q", got, want)
	}
	if want, got := "Shelter animal APIs", got.Description.ValueString(); got != want {
		t.Errorf("Description = %q, want %q", got, want)
	}
	// PortalProductARN/ID bind to PortalProductArn/Id via case-insensitive matching.
	if want, got := "arn:aws:apigateway:us-west-2:123456789012:/portalproducts/abcdef1234", got.PortalProductARN.ValueString(); got != want { //lintignore:AWSAT003,AWSAT005
		t.Errorf("PortalProductARN = %q, want %q", got, want)
	}
	if want, got := "abcdef1234", got.PortalProductID.ValueString(); got != want {
		t.Errorf("PortalProductID = %q, want %q", got, want)
	}
	if want, got := timetypes.NewRFC3339TimeValue(lastModified), got.LastModified; !got.Equal(want) {
		t.Errorf("LastModified = %s, want %s", got, want)
	}
}
