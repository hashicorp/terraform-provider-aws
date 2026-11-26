// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"maps"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestUserPoolClientTermsModel(t *testing.T) {
	t.Parallel()

	const (
		poolID    = "example-region_EXAMPLE"
		clientID  = "1example23456789"
		termsID   = "a1b2c3d4-5678-4321-8765-abcdef123456"
		termsName = "terms-of-use"
	)

	testCases := map[string]map[string]string{
		"localized links": {
			"cognito:default": "https://example.com/terms",
			"cognito:dutch":   "https://example.com/nl/terms",
		},
		"links omitted in response": nil,
	}

	for name, links := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			var model userPoolClientTermsResourceModel
			diags := model.flatten(ctx, &awstypes.TermsType{
				ClientId:   aws.String(clientID),
				Links:      links,
				TermsId:    aws.String(termsID),
				TermsName:  aws.String(termsName),
				UserPoolId: aws.String(poolID),
			})
			if diags.HasError() {
				t.Fatalf("flattening terms: %v", diags)
			}
			if model.TermsID.ValueString() != termsID {
				t.Fatalf("terms ID = %q, want %q", model.TermsID.ValueString(), termsID)
			}
			if model.Links.IsNull() {
				t.Fatal("links are null, want a known map")
			}

			var input cognitoidentityprovider.CreateTermsInput
			diags = flex.Expand(ctx, model, &input)
			if diags.HasError() {
				t.Fatalf("expanding terms: %v", diags)
			}
			if aws.ToString(input.ClientId) != clientID || aws.ToString(input.UserPoolId) != poolID || aws.ToString(input.TermsName) != termsName {
				t.Errorf("wrong client, user pool, or terms name in create request: %+v", input)
			}
			if !maps.Equal(input.Links, links) {
				t.Errorf("links = %v, want %v", input.Links, links)
			}

			var update cognitoidentityprovider.UpdateTermsInput
			diags = flex.Expand(ctx, model, &update)
			if diags.HasError() {
				t.Fatalf("expanding terms update: %v", diags)
			}
			if aws.ToString(update.TermsId) != termsID || aws.ToString(update.UserPoolId) != poolID || aws.ToString(update.TermsName) != termsName {
				t.Errorf("wrong terms ID, user pool, or terms name in update request: %+v", update)
			}
			if !maps.Equal(update.Links, links) {
				t.Errorf("updated links = %v, want %v", update.Links, links)
			}

			if len(links) == 0 && !model.Links.Equal(fwtypes.NewMapValueOfEmpty[types.String](ctx)) {
				t.Errorf("links = %v, want an empty map", model.Links)
			}
		})
	}
}

func TestUserPoolClientTermsImportID(t *testing.T) {
	t.Parallel()

	const (
		poolID  = "example-region_EXAMPLE"
		termsID = "a1b2c3d4-5678-4321-8765-abcdef123456"
	)

	parser := userPoolClientTermsImportID{}
	_, attributes, err := parser.Parse(poolID + "," + termsID)
	if err != nil {
		t.Fatalf("parsing import ID: %v", err)
	}
	if attributes[names.AttrUserPoolID] != poolID || attributes["terms_id"] != termsID {
		t.Errorf("unexpected import attributes: %v", attributes)
	}

	for _, input := range []string{termsID, poolID + ",", "," + termsID, poolID + "," + termsID + ",extra"} {
		if _, _, err := parser.Parse(input); err == nil {
			t.Errorf("parsing %q succeeded, want an error", input)
		}
	}
}
