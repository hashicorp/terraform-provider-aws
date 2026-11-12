// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"strings"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type testCustomClaim struct {
	name      string
	value     *string
	valueList []string
}

func testCustomJWTAuthorizer(ctx context.Context, t *testing.T, writeOnly fwtypes.MapOfString, version types.Int64, claims ...testCustomClaim) customJWTAuthorizerConfigurationModel {
	t.Helper()

	m, diags := fwtypes.Nullified[customJWTAuthorizerConfigurationModel](ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	m.DiscoveryURL = types.StringValue("https://accounts.google.com/.well-known/openid-configuration")
	m.CustomClaimMatchValuesWO = writeOnly
	m.CustomClaimMatchValuesWOVersion = version

	var customClaims []*customJWTAuthorizerCustomClaimModel
	for _, c := range claims {
		claimMatchValue := customJWTAuthorizerClaimMatchValueModel{
			MatchValueString:     types.StringPointerValue(c.value),
			MatchValueStringList: fwtypes.NewSetValueOfNull[types.String](ctx),
		}
		if c.valueList != nil {
			var elems []attr.Value
			for _, v := range c.valueList {
				elems = append(elems, types.StringValue(v))
			}
			claimMatchValue.MatchValueStringList = fwtypes.NewSetValueOfMust[types.String](ctx, elems)
		}
		customClaims = append(customClaims, &customJWTAuthorizerCustomClaimModel{
			InboundTokenClaimName:      types.StringValue(c.name),
			InboundTokenClaimValueType: fwtypes.StringEnumValue(awstypes.InboundTokenClaimValueTypeString),
			AuthorizingClaimMatchValue: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customJWTAuthorizerAuthorizingClaimMatchValueModel{
				ClaimMatchOperator: fwtypes.StringEnumValue(awstypes.ClaimMatchOperatorTypeEquals),
				ClaimMatchValue:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &claimMatchValue),
			}),
		})
	}
	if len(customClaims) > 0 {
		m.CustomClaim = fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, customClaims)
	}

	return m
}

func testWriteOnlyMap(ctx context.Context, t *testing.T, m map[string]string) fwtypes.MapOfString {
	t.Helper()

	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	v, diags := fwtypes.NewMapValueOf[types.String](ctx, elems)
	if diags.HasError() {
		t.Fatal(diags)
	}
	return v
}

func testClaimMatchValues(ctx context.Context, t *testing.T, auth fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel]) map[string]customJWTAuthorizerClaimMatchValueModel {
	t.Helper()

	a, diags := auth.ToPtr(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	jwt, diags := a.CustomJWTAuthorizer.ToPtr(ctx)
	if diags.HasError() {
		t.Fatal(diags)
	}
	got := make(map[string]customJWTAuthorizerClaimMatchValueModel)
	if diags := walkCustomClaimMatchValues(ctx, jwt, func(name string, v *customJWTAuthorizerClaimMatchValueModel) {
		got[name] = *v
	}); diags.HasError() {
		t.Fatal(diags)
	}
	return got
}

func testAuthorizerConfiguration(ctx context.Context, jwt customJWTAuthorizerConfigurationModel) fwtypes.ListNestedObjectValueOf[authorizerConfigurationModel] {
	return fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &authorizerConfigurationModel{
		CustomJWTAuthorizer: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &jwt),
	})
}

func TestCustomClaimMatchValueSourcesValidator(t *testing.T) {
	t.Parallel()

	const secret = "s3cr3t-value"
	null := fwtypes.NewMapValueOfNull[types.String]

	for name, testCase := range map[string]struct {
		writeOnly     func(context.Context, *testing.T) fwtypes.MapOfString
		claims        []testCustomClaim
		wantSummaries []string
	}{
		"match_value_string": {
			claims: []testCustomClaim{{name: "tenant_id", value: new("admin")}},
		},
		"match_value_string_list": {
			claims: []testCustomClaim{{name: "groups", valueList: []string{"admin", "user"}}},
		},
		"write-only": {
			writeOnly: func(ctx context.Context, t *testing.T) fwtypes.MapOfString {
				return testWriteOnlyMap(ctx, t, map[string]string{"tenant_id": secret})
			},
			claims: []testCustomClaim{{name: "tenant_id"}},
		},
		"write-only and match_value_string on different claims": {
			writeOnly: func(ctx context.Context, t *testing.T) fwtypes.MapOfString {
				return testWriteOnlyMap(ctx, t, map[string]string{"tenant_id": secret})
			},
			claims: []testCustomClaim{{name: "tenant_id"}, {name: names.AttrRole, value: new("admin")}},
		},
		"write-only unknown": {
			writeOnly: func(ctx context.Context, t *testing.T) fwtypes.MapOfString {
				return fwtypes.NewMapValueOfUnknown[types.String](ctx)
			},
			claims: []testCustomClaim{{name: "tenant_id"}},
		},
		"no source": {
			claims:        []testCustomClaim{{name: "tenant_id"}},
			wantSummaries: []string{"Invalid Attribute Combination"},
		},
		"match_value_string and match_value_string_list": {
			claims:        []testCustomClaim{{name: "tenant_id", value: new("admin"), valueList: []string{"user"}}},
			wantSummaries: []string{"Invalid Attribute Combination"},
		},
		"match_value_string and write-only": {
			writeOnly: func(ctx context.Context, t *testing.T) fwtypes.MapOfString {
				return testWriteOnlyMap(ctx, t, map[string]string{"tenant_id": secret})
			},
			claims:        []testCustomClaim{{name: "tenant_id", value: new("admin")}},
			wantSummaries: []string{"Invalid Attribute Combination"},
		},
		"write-only key without claim": {
			writeOnly: func(ctx context.Context, t *testing.T) fwtypes.MapOfString {
				return testWriteOnlyMap(ctx, t, map[string]string{"tenant_id": secret, "other": secret})
			},
			claims:        []testCustomClaim{{name: "tenant_id"}},
			wantSummaries: []string{"Invalid Attribute Value"},
		},
		"write-only invalid characters": {
			writeOnly: func(ctx context.Context, t *testing.T) fwtypes.MapOfString {
				return testWriteOnlyMap(ctx, t, map[string]string{"tenant_id": secret + "!"})
			},
			claims:        []testCustomClaim{{name: "tenant_id"}},
			wantSummaries: []string{"Invalid Attribute Value Match"},
		},
		"write-only too long": {
			writeOnly: func(ctx context.Context, t *testing.T) fwtypes.MapOfString {
				return testWriteOnlyMap(ctx, t, map[string]string{"tenant_id": strings.Repeat("a", 256)})
			},
			claims:        []testCustomClaim{{name: "tenant_id"}},
			wantSummaries: []string{"Invalid Attribute Value Length"},
		},
		"write-only empty": {
			writeOnly: func(ctx context.Context, t *testing.T) fwtypes.MapOfString {
				return testWriteOnlyMap(ctx, t, map[string]string{"tenant_id": ""})
			},
			claims:        []testCustomClaim{{name: "tenant_id"}},
			wantSummaries: []string{"Invalid Attribute Value Length", "Invalid Attribute Value Match"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			writeOnly := null(ctx)
			version := types.Int64Null()
			if testCase.writeOnly != nil {
				writeOnly = testCase.writeOnly(ctx, t)
				version = types.Int64Value(1)
			}
			model := testCustomJWTAuthorizer(ctx, t, writeOnly, version, testCase.claims...)

			request := validator.ObjectRequest{
				ConfigValue: fwtypes.NewObjectValueOfMust(ctx, &model).ObjectValue,
				Path:        path.Root("authorizer_configuration").AtListIndex(0).AtName("custom_jwt_authorizer").AtListIndex(0),
			}
			var response validator.ObjectResponse
			customClaimMatchValueSourcesValidator{}.ValidateObject(ctx, request, &response)

			var got []string
			for _, d := range response.Diagnostics.Errors() {
				got = append(got, d.Summary())
				if strings.Contains(d.Detail(), secret) {
					t.Errorf("diagnostic detail contains the write-only value: %q", d.Detail())
				}
			}
			if strings.Join(got, ",") != strings.Join(testCase.wantSummaries, ",") {
				t.Fatalf("error summaries = %v, want %v; diagnostics: %v", got, testCase.wantSummaries, response.Diagnostics)
			}
		})
	}
}

func TestWithCustomClaimMatchValuesWO(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	r, err := newAgentRuntimeResource(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var schemaResponse resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)

	claims := []testCustomClaim{{name: "tenant_id"}, {name: names.AttrRole, value: new("admin")}}
	configured := testAuthorizerConfiguration(ctx, testCustomJWTAuthorizer(ctx, t, testWriteOnlyMap(ctx, t, map[string]string{"tenant_id": "secret"}), types.Int64Value(1), claims...))
	planned := testAuthorizerConfiguration(ctx, testCustomJWTAuthorizer(ctx, t, fwtypes.NewMapValueOfNull[types.String](ctx), types.Int64Value(1), claims...))

	state := tfsdk.State{Schema: schemaResponse.Schema, Raw: tftypes.NewValue(schemaResponse.Schema.Type().TerraformType(ctx), nil)}
	if diags := state.SetAttribute(ctx, path.Root("authorizer_configuration"), configured); diags.HasError() {
		t.Fatal(diags)
	}
	config := tfsdk.Config{Schema: schemaResponse.Schema, Raw: state.Raw}

	got, diags := withCustomClaimMatchValuesWO(ctx, config, planned)
	if diags.HasError() {
		t.Fatal(diags)
	}

	values := testClaimMatchValues(ctx, t, got)
	if v := values["tenant_id"].MatchValueString; v.ValueString() != "secret" {
		t.Errorf("tenant_id match_value_string = %s, want %q", v, "secret")
	}
	if v := values[names.AttrRole].MatchValueString; v.ValueString() != "admin" {
		t.Errorf("group match_value_string = %s, want %q", v, "admin")
	}

	if v := testClaimMatchValues(ctx, t, planned)["tenant_id"].MatchValueString; !v.IsNull() {
		t.Errorf("planned tenant_id match_value_string = %s, want null", v)
	}
}

func TestPreserveAuthorizerConfiguration(t *testing.T) {
	t.Parallel()
	ctx := t.Context()

	null := fwtypes.NewMapValueOfNull[types.String](ctx)
	src := testAuthorizerConfiguration(ctx, testCustomJWTAuthorizer(ctx, t, null, types.Int64Value(2),
		testCustomClaim{name: "tenant_id"},
		testCustomClaim{name: names.AttrRole, value: new("admin")},
	))
	dst := testAuthorizerConfiguration(ctx, testCustomJWTAuthorizer(ctx, t, null, types.Int64Null(),
		testCustomClaim{name: "tenant_id", value: new("secret")},
		testCustomClaim{name: names.AttrRole, value: new("admin")},
	))

	got, diags := preserveAuthorizerConfiguration(ctx, dst, src)
	if diags.HasError() {
		t.Fatal(diags)
	}
	if !got.Equal(src) {
		t.Errorf("preserveAuthorizerConfiguration() = %s, want %s", got, src)
	}
}

func TestAuthorizerConfigurationSchemaImplementation(t *testing.T) {
	t.Parallel()

	for name, f := range map[string]func(context.Context) (resource.ResourceWithConfigure, error){
		"agent_runtime": newAgentRuntimeResource,
		"gateway":       newGatewayResource,
		"harness":       newHarnessResource,
		"registry":      newRegistryResource,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			r, err := f(ctx)
			if err != nil {
				t.Fatal(err)
			}
			var response resource.SchemaResponse
			r.Schema(ctx, resource.SchemaRequest{}, &response)
			if response.Diagnostics.HasError() {
				t.Fatal(response.Diagnostics)
			}
			if diags := response.Schema.ValidateImplementation(ctx); diags.HasError() {
				t.Fatal(diags)
			}
		})
	}
}
