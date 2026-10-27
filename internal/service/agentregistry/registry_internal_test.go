// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package agentregistry

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/agentregistrycontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestRegistryCustomJWTAuthorizerValidation(t *testing.T) {
	t.Parallel()

	for name, testCase := range map[string]struct {
		customClaim bool
		emptyClaims bool
		wantError   bool
	}{
		"custom claim only": {customClaim: true},
		"null claims":       {wantError: true},
		"empty claims":      {emptyClaims: true, wantError: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			r, err := newRegistryResource(ctx)
			if err != nil {
				t.Fatal(err)
			}
			var schemaResponse resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
			discoveryBlock := schemaResponse.Schema.Blocks["discovery_configuration"].(schema.ListNestedBlock)
			authorizerBlock := discoveryBlock.NestedObject.Blocks["authorizer_configuration"].(schema.ListNestedBlock)
			jwtBlock := authorizerBlock.NestedObject.Blocks["custom_jwt_authorizer"].(schema.ListNestedBlock)

			model, diags := fwtypes.Nullified[customJWTAuthorizerConfigurationModel](ctx)
			if diags.HasError() {
				t.Fatal(diags)
			}
			model.DiscoveryURL = types.StringValue("https://accounts.google.com/.well-known/openid-configuration")
			if testCase.emptyClaims {
				model.CustomClaims = fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []customClaimValidationTypeModel{})
			}
			if testCase.customClaim {
				model.CustomClaims = fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &customClaimValidationTypeModel{
					InboundTokenClaimName:      types.StringValue("claim_99"),
					InboundTokenClaimValueType: fwtypes.StringEnumValue(awstypes.InboundTokenClaimValueTypeString),
					AuthorizingClaimMatchValue: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &authorizingClaimMatchValueTypeModel{
						ClaimMatchOperator: fwtypes.StringEnumValue(awstypes.ClaimMatchOperatorTypeEquals),
						ClaimMatchValue: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &claimMatchValueTypeModel{
							MatchValueString:     types.StringValue("value01"),
							MatchValueStringList: fwtypes.NewSetValueOfNull[types.String](ctx),
						}),
					}),
				})
			}

			state := tfsdk.State{Schema: schemaResponse.Schema, Raw: tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(ctx), nil)}
			discovery := fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &discoveryConfigurationModel{
				AuthorizerType: fwtypes.StringEnumValue(awstypes.RegistryAuthorizerTypeCustomJwt),
				AuthorizerConfiguration: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &authorizerConfigurationModel{
					CustomJWTAuthorizer: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model),
				}),
			})
			if diags := state.SetAttribute(ctx, path.Root("discovery_configuration"), discovery); diags.HasError() {
				t.Fatal(diags)
			}
			request := validator.ObjectRequest{
				Config:         tfsdk.Config{Schema: schemaResponse.Schema, Raw: state.Raw},
				ConfigValue:    fwtypes.NewObjectValueOfMust(ctx, &model).ObjectValue,
				Path:           path.Root("discovery_configuration").AtListIndex(0).AtName("authorizer_configuration").AtListIndex(0).AtName("custom_jwt_authorizer").AtListIndex(0),
				PathExpression: path.MatchRoot("discovery_configuration").AtListIndex(0).AtName("authorizer_configuration").AtListIndex(0).AtName("custom_jwt_authorizer").AtListIndex(0),
			}
			var response validator.ObjectResponse
			for _, v := range jwtBlock.NestedObject.Validators {
				v.ValidateObject(ctx, request, &response)
			}
			if got := response.Diagnostics.HasError(); got != testCase.wantError {
				t.Fatalf("HasError = %t, want %t; diagnostics: %v", got, testCase.wantError, response.Diagnostics)
			}
			if testCase.wantError {
				if got := response.Diagnostics.Errors(); len(got) != 1 || got[0].Summary() != "Invalid Attribute Combination" {
					t.Fatalf("unexpected diagnostics: %v", got)
				}
			}
		})
	}
}
