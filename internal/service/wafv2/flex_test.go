// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"encoding/base64"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func Test_expandWebACLRulesJSON(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		rawRules string
		want     []awstypes.Rule
		wantErr  bool
	}{
		"empty string": {
			rawRules: "",
			wantErr:  true,
		},
		"empty array": {
			rawRules: "[]",
			want:     []awstypes.Rule{},
		},
		"single empty object": {
			rawRules: "[{}]",
			wantErr:  true,
		},
		"single null object": {
			rawRules: "[null]",
			wantErr:  true,
		},
		"valid object": {
			rawRules: `[{"Action":{"Count":{}},"Name":"rule-1","Priority":1,"Statement":{"RateBasedStatement":{"AggregateKeyType":"IP","EvaluationWindowSec":600,"Limit":10000,"ScopeDownStatement":{"GeoMatchStatement":{"CountryCodes":["US","NL"]}}}},"VisibilityConfig":{"CloudwatchMetricsEnabled":false,"MetricName":"friendly-rule-metric-name","SampledRequestsEnabled":false}}]`,
			want: []awstypes.Rule{
				{
					Name:     aws.String("rule-1"),
					Priority: 1,
					Action: &awstypes.RuleAction{
						Count: &awstypes.CountAction{},
					},
					Statement: &awstypes.Statement{
						RateBasedStatement: &awstypes.RateBasedStatement{
							Limit:               aws.Int64(10000),
							AggregateKeyType:    awstypes.RateBasedStatementAggregateKeyType("IP"),
							EvaluationWindowSec: 600,
							ScopeDownStatement: &awstypes.Statement{
								GeoMatchStatement: &awstypes.GeoMatchStatement{
									CountryCodes: []awstypes.CountryCode{"US", "NL"},
								},
							},
						},
					},
					VisibilityConfig: &awstypes.VisibilityConfig{
						CloudWatchMetricsEnabled: false,
						MetricName:               aws.String("friendly-rule-metric-name"),
						SampledRequestsEnabled:   false,
					},
				},
			},
		},
		"valid and empty object": {
			rawRules: `[{"Action":{"Count":{}},"Name":"rule-1","Priority":1,"Statement":{"RateBasedStatement":{"AggregateKeyType":"IP","EvaluationWindowSec":600,"Limit":10000,"ScopeDownStatement":{"GeoMatchStatement":{"CountryCodes":["US","NL"]}}}},"VisibilityConfig":{"CloudwatchMetricsEnabled":false,"MetricName":"friendly-rule-metric-name","SampledRequestsEnabled":false}},{}]`,
			wantErr:  true,
		},
		"valid object SearchString": {
			rawRules: `[{"Name" : "test_rule0","Priority":0,"Statement":{"AndStatement":{"Statements":[{"ByteMatchStatement":{"SearchString":"test","FieldToMatch":{"SingleHeader":{"Name":"host"}},"TextTransformations":[{"Priority":0,"Type":"NONE"}],"PositionalConstraint":"EXACTLY"}}]},"ByteMatchStatement":{"SearchString":"test","FieldToMatch":{"SingleHeader":{"Name":"host"}},"TextTransformations":[{"Priority":0,"Type":"NONE"}],"PositionalConstraint":"EXACTLY"}},"Action":{"Block":{}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"test_rule0"}}]`,
			want: []awstypes.Rule{
				{
					Name:     aws.String("test_rule0"),
					Priority: 0,
					Action: &awstypes.RuleAction{
						Block: &awstypes.BlockAction{},
					},
					VisibilityConfig: &awstypes.VisibilityConfig{
						SampledRequestsEnabled:   true,
						CloudWatchMetricsEnabled: true,
						MetricName:               aws.String("test_rule0"),
					},
					Statement: &awstypes.Statement{
						AndStatement: &awstypes.AndStatement{
							Statements: []awstypes.Statement{
								{
									ByteMatchStatement: &awstypes.ByteMatchStatement{
										SearchString: []byte("test"),
										FieldToMatch: &awstypes.FieldToMatch{
											SingleHeader: &awstypes.SingleHeader{
												Name: aws.String("host"),
											},
										},
										TextTransformations: []awstypes.TextTransformation{
											{
												Priority: 0,
												Type:     awstypes.TextTransformationType("NONE"),
											},
										},
										PositionalConstraint: awstypes.PositionalConstraint("EXACTLY"),
									},
								},
							},
						},
						ByteMatchStatement: &awstypes.ByteMatchStatement{
							SearchString: []byte("test"),
							FieldToMatch: &awstypes.FieldToMatch{
								SingleHeader: &awstypes.SingleHeader{
									Name: aws.String("host"),
								},
							},
							TextTransformations: []awstypes.TextTransformation{
								{
									Priority: 0,
									Type:     awstypes.TextTransformationType("NONE"),
								},
							},
							PositionalConstraint: awstypes.PositionalConstraint("EXACTLY"),
						},
					},
				},
			},
		},
	}

	ignoreExportedOpts := cmpopts.IgnoreUnexported(
		awstypes.Rule{},
		awstypes.RuleAction{},
		awstypes.CountAction{},
		awstypes.Statement{},
		awstypes.RateBasedStatement{},
		awstypes.GeoMatchStatement{},
		awstypes.VisibilityConfig{},
		awstypes.SingleHeader{},
		awstypes.ByteMatchStatement{},
		awstypes.FieldToMatch{},
		awstypes.TextTransformation{},
		awstypes.BlockAction{},
		awstypes.AndStatement{},
	)

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := expandWebACLRulesJSON(tc.rawRules)
			if (err != nil) != tc.wantErr {
				t.Errorf("expandWebACLRulesJSON() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if diff := cmp.Diff(got, tc.want, ignoreExportedOpts); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestByteMatchStatementSearchString(t *testing.T) {
	t.Parallel()

	md5Digest := []byte{0xfc, 0x5e, 0x03, 0x8d, 0x38, 0xa5, 0x70, 0x32, 0x08, 0x54, 0x41, 0xe7, 0xfe, 0x70, 0x10, 0xb0} // md5("helloworld")

	testCases := map[string]struct {
		searchString       string
		searchStringBase64 string
		wantBytes          []byte
		wantFlattened      map[string]any
	}{
		"text": {
			searchString:  "BadBot",
			wantBytes:     []byte("BadBot"),
			wantFlattened: map[string]any{"search_string": "BadBot"},
		},
		"binary": {
			searchStringBase64: "/F4DjTilcDIIVEHn/nAQsA==",
			wantBytes:          md5Digest,
			wantFlattened:      map[string]any{"search_string_base64": "/F4DjTilcDIIVEHn/nAQsA=="},
		},
		// Valid UTF-8 set in base64 reads back as text unless prior state says otherwise.
		"text in base64": {
			searchStringBase64: "QmFkQm90",
			wantBytes:          []byte("BadBot"),
			wantFlattened:      map[string]any{"search_string": "BadBot"},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := expandSearchString(map[string]any{"search_string": tc.searchString, "search_string_base64": tc.searchStringBase64})
			if diff := cmp.Diff(got, tc.wantBytes); diff != "" {
				t.Errorf("expand (-got +want):\n%s", diff)
			}

			m := flattenByteMatchStatement(&awstypes.ByteMatchStatement{SearchString: got}).([]any)[0].(map[string]any)
			for _, k := range []string{"search_string", "search_string_base64"} {
				if diff := cmp.Diff(m[k], tc.wantFlattened[k]); diff != "" {
					t.Errorf("flatten %s (-got +want):\n%s", k, diff)
				}
			}
		})
	}
}

func TestPreferSearchStringBase64(t *testing.T) {
	t.Parallel()

	rule := func(name string, byteMatches ...map[string]any) map[string]any {
		statements := make([]any, len(byteMatches))
		for i, m := range byteMatches {
			statements[i] = map[string]any{"byte_match_statement": []any{m}}
		}
		return map[string]any{names.AttrName: name, "statement": []any{map[string]any{
			"and_statement": []any{map[string]any{"statement": statements}},
		}}}
	}
	leaf := func(rule map[string]any, i int) map[string]any {
		return rule["statement"].([]any)[0].(map[string]any)["and_statement"].([]any)[0].(map[string]any)["statement"].([]any)[i].(map[string]any)["byte_match_statement"].([]any)[0].(map[string]any)
	}
	text := func(v string) map[string]any { return map[string]any{"search_string": v} }
	base64 := func(v string) map[string]any { return map[string]any{"search_string_base64": v} }

	// Read back from the API, all values are valid UTF-8, so all flatten to search_string.
	flattened := []map[string]any{
		rule("rule-1", text("BadBot"), text("BadBot")),
		rule("rule-2", text("BadBot")),
	}
	// The configuration set the same bytes in base64 in one place only.
	prior := []any{
		rule("rule-1", base64("QmFkQm90"), text("BadBot")),
		rule("rule-2", text("BadBot")),
	}

	preferSearchStringBase64(flattened, prior)

	for _, tc := range []struct {
		name string
		got  map[string]any
		want map[string]any
	}{
		{"rule-1 statement 0, configured in base64", leaf(flattened[0], 0), base64("QmFkQm90")},
		{"rule-1 statement 1, configured as text", leaf(flattened[0], 1), text("BadBot")},
		{"rule-2, configured as text", leaf(flattened[1], 0), text("BadBot")},
	} {
		if diff := cmp.Diff(tc.got, tc.want); diff != "" {
			t.Errorf("%s (-got +want):\n%s", tc.name, diff)
		}
	}
}

func TestValidateByteMatchStatementSearchString(t *testing.T) {
	t.Parallel()

	byteMatchType := cty.Object(map[string]cty.Type{"search_string": cty.String, "search_string_base64": cty.String})
	config := func(byteMatch cty.Value) cty.Value {
		// rule (set) > statement > not_statement > statement > byte_match_statement
		return cty.ObjectVal(map[string]cty.Value{
			names.AttrRule: cty.SetVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
				"statement": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
					"not_statement": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
						"statement": cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
							"byte_match_statement": cty.ListVal([]cty.Value{byteMatch}),
						})}),
					})}),
				})}),
			})}),
		})
	}
	byteMatch := func(searchString, searchStringBase64 cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{"search_string": searchString, "search_string_base64": searchStringBase64})
	}

	testCases := map[string]struct {
		config    cty.Value
		wantError bool
	}{
		"search_string":           {config: config(byteMatch(cty.StringVal("a"), cty.NullVal(cty.String)))},
		"search_string_base64":    {config: config(byteMatch(cty.NullVal(cty.String), cty.StringVal("YQ==")))},
		"neither":                 {config: config(byteMatch(cty.NullVal(cty.String), cty.NullVal(cty.String))), wantError: true},
		"both":                    {config: config(byteMatch(cty.StringVal("a"), cty.StringVal("YQ=="))), wantError: true},
		"unknown":                 {config: config(byteMatch(cty.UnknownVal(cty.String), cty.UnknownVal(cty.String)))},
		"unknown block":           {config: config(cty.UnknownVal(byteMatchType))},
		"no byte_match_statement": {config: cty.ObjectVal(map[string]cty.Value{names.AttrRule: cty.SetValEmpty(cty.EmptyObject)})},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var resp schema.ValidateResourceConfigFuncResponse
			validateByteMatchStatementSearchString(t.Context(), schema.ValidateResourceConfigFuncRequest{RawConfig: tc.config}, &resp)

			if got := resp.Diagnostics.HasError(); got != tc.wantError {
				t.Errorf("HasError() = %t, want %t: %v", got, tc.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestWebACLRuleByteMatchStatementSearchString(t *testing.T) {
	t.Parallel()

	md5Digest := []byte{0xfc, 0x5e, 0x03, 0x8d, 0x38, 0xa5, 0x70, 0x32, 0x08, 0x54, 0x41, 0xe7, 0xfe, 0x70, 0x10, 0xb0} // md5("helloworld")

	testCases := map[string]struct {
		searchString           []byte
		wantSearchString       types.String
		wantSearchStringBase64 types.String
	}{
		"text": {
			searchString:           []byte("BadBot"),
			wantSearchString:       types.StringValue("BadBot"),
			wantSearchStringBase64: types.StringNull(),
		},
		"binary": {
			searchString:           md5Digest,
			wantSearchString:       types.StringNull(),
			wantSearchStringBase64: types.StringValue("/F4DjTilcDIIVEHn/nAQsA=="),
		},
	}

	ignore := cmpopts.IgnoreUnexported(awstypes.Statement{}, awstypes.NotStatement{}, awstypes.ByteMatchStatement{}, awstypes.FieldToMatch{}, awstypes.SingleHeader{}, awstypes.TextTransformation{})

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			byteMatch := awstypes.ByteMatchStatement{
				FieldToMatch:         &awstypes.FieldToMatch{SingleHeader: &awstypes.SingleHeader{Name: aws.String("authorization")}},
				PositionalConstraint: awstypes.PositionalConstraintExactly,
				SearchString:         tc.searchString,
				TextTransformations:  []awstypes.TextTransformation{{Priority: 0, Type: awstypes.TextTransformationTypeMd5}},
			}

			var m webACLRuleByteMatchStatementModel
			if diags := fwflex.Flatten(ctx, &byteMatch, &m); diags.HasError() {
				t.Fatalf("Flatten: %v", diags)
			}
			if !m.SearchString.Equal(tc.wantSearchString) || !m.SearchStringBase64.Equal(tc.wantSearchStringBase64) {
				t.Errorf("Flatten: search_string = %s, search_string_base64 = %s; want %s, %s", m.SearchString, m.SearchStringBase64, tc.wantSearchString, tc.wantSearchStringBase64)
			}

			var got awstypes.ByteMatchStatement
			if diags := fwflex.Expand(ctx, m, &got); diags.HasError() {
				t.Fatalf("Expand: %v", diags)
			}
			if diff := cmp.Diff(got, byteMatch, ignore); diff != "" {
				t.Errorf("round trip (-got +want):\n%s", diff)
			}

			// Nested: the custom methods also run below a not_statement.
			statement := awstypes.Statement{NotStatement: &awstypes.NotStatement{Statement: &awstypes.Statement{ByteMatchStatement: &byteMatch}}}
			var sm webACLRuleStatementModel
			if diags := fwflex.Flatten(ctx, &statement, &sm); diags.HasError() {
				t.Fatalf("Flatten statement: %v", diags)
			}
			var gotStatement awstypes.Statement
			if diags := fwflex.Expand(ctx, sm, &gotStatement); diags.HasError() {
				t.Fatalf("Expand statement: %v", diags)
			}
			if diff := cmp.Diff(gotStatement, statement, ignore); diff != "" {
				t.Errorf("nested round trip (-got +want):\n%s", diff)
			}
		})
	}
}

func TestKeepSearchStringBase64(t *testing.T) {
	t.Parallel()

	byteMatchType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"search_string":        tftypes.String,
		"search_string_base64": tftypes.String,
	}}
	byteMatch := func(searchString, searchStringBase64 any) tftypes.Value {
		return tftypes.NewValue(byteMatchType, map[string]tftypes.Value{
			"search_string":        tftypes.NewValue(tftypes.String, searchString),
			"search_string_base64": tftypes.NewValue(tftypes.String, searchStringBase64),
		})
	}
	// statement > byte_match_statement (a list), two statements with the same bytes
	ruleType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"statement": tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
			"byte_match_statement": tftypes.List{ElementType: byteMatchType},
		}}},
	}}
	rule := func(byteMatches ...tftypes.Value) tftypes.Value {
		statementType := ruleType.AttributeTypes["statement"].(tftypes.List).ElementType
		statements := make([]tftypes.Value, len(byteMatches))
		for i, v := range byteMatches {
			statements[i] = tftypes.NewValue(statementType, map[string]tftypes.Value{
				"byte_match_statement": tftypes.NewValue(tftypes.List{ElementType: byteMatchType}, []tftypes.Value{v}),
			})
		}
		return tftypes.NewValue(ruleType, map[string]tftypes.Value{
			"statement": tftypes.NewValue(ruleType.AttributeTypes["statement"], statements),
		})
	}

	// The plan set the bytes in base64 in the first statement and as text in the second.
	prior := rule(byteMatch(nil, "QmFkQm90"), byteMatch("BadBot", nil))
	// Read back from the API, both flatten to search_string.
	state := tfsdk.State{Raw: rule(byteMatch("BadBot", nil), byteMatch("BadBot", nil))}

	var diags diag.Diagnostics
	keepSearchStringBase64(prior, &state, &diags)
	if diags.HasError() {
		t.Fatalf("keepSearchStringBase64: %v", diags)
	}

	if want := rule(byteMatch(nil, "QmFkQm90"), byteMatch("BadBot", nil)); !state.Raw.Equal(want) {
		t.Errorf("state = %s, want %s", state.Raw, want)
	}

	// No prior value (import): state is unchanged.
	imported := tfsdk.State{Raw: rule(byteMatch("BadBot", nil))}
	keepSearchStringBase64(tftypes.NewValue(ruleType, nil), &imported, &diags)
	if want := rule(byteMatch("BadBot", nil)); diags.HasError() || !imported.Raw.Equal(want) {
		t.Errorf("imported state = %s, want %s (%v)", imported.Raw, want, diags)
	}
}

func TestValidSearchStringBase64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value     string
		wantError bool
	}{
		"1 byte":          {value: "YQ=="},
		"200 bytes":       {value: base64.StdEncoding.EncodeToString(make([]byte, 200))},
		"201 bytes":       {value: base64.StdEncoding.EncodeToString(make([]byte, 201)), wantError: true},
		"empty":           {value: "", wantError: true},
		"not base64":      {value: "not base64!", wantError: true},
		"missing padding": {value: "YQ", wantError: true},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, errs := validSearchStringBase64(tc.value, "search_string_base64")
			if got := len(errs) > 0; got != tc.wantError {
				t.Errorf("validSearchStringBase64(%q) errors = %v, want error %t", tc.value, errs, tc.wantError)
			}
		})
	}
}
