// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's Expand/Flatten of naming differences (plural, prefix, suffix, capitalization).

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// tfSingluarListOfNestedObjects testing for idiomatic singular on TF side but plural on AWS side
type tfSingluarListOfNestedObjects struct {
	Field fwtypes.ListNestedObjectValueOf[tfSingleStringField] `tfsdk:"field"`
}

type awsPluralSliceOfNestedObjectValues struct {
	Fields []awsSingleStringValue
}

// tfFieldNamePrefix has no prefix to test matching on prefix
type tfFieldNamePrefix struct {
	Name types.String `tfsdk:"name"`
}

// awsFieldNamePrefix has prefix to test matching on prefix
type awsFieldNamePrefix struct {
	IntentName *string
}

type tfFieldNamePrefixInsensitive struct {
	ID types.String `tfsdk:"id"`
}

type awsFieldNamePrefixInsensitive struct {
	ClientId *string
}

// tfFieldNameSuffix has no suffix to test matching on suffix
type tfFieldNameSuffix struct {
	Policy types.String `tfsdk:"policy"`
}

// awsFieldNameSuffix has suffix to test matching on suffix
type awsFieldNameSuffix struct {
	PolicyConfig *string
}

type tfPluralAndSingularFields struct {
	Value types.String `tfsdk:"Value"`
}

type awsPluralAndSingularFields struct {
	Value  string
	Values string
}

type tfSpecialPluralization struct {
	City      types.List `tfsdk:"city"`
	Coach     types.List `tfsdk:"coach"`
	Tomato    types.List `tfsdk:"tomato"`
	Vertex    types.List `tfsdk:"vertex"`
	Criterion types.List `tfsdk:"criterion"`
	Datum     types.List `tfsdk:"datum"`
	Hive      types.List `tfsdk:"hive"`
}

type awsSpecialPluralization struct {
	Cities   []*string
	Coaches  []*string
	Tomatoes []*string
	Vertices []*string
	Criteria []*string
	Data     []*string
	Hives    []*string
}

// tfCaptializationDiff testing for fields that only differ by capitalization
type tfCaptializationDiff struct {
	FieldURL types.String `tfsdk:"field_url"`
}

// awsCapitalizationDiff testing for fields that only differ by capitalization
type awsCapitalizationDiff struct {
	FieldUrl *string
}

func TestExpandNaming(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"plural ordinary field names": {
			Source: &tfSingluarListOfNestedObjects{
				Field: fwtypes.NewListNestedObjectValueOfPtrMust(context.Background(), &tfSingleStringField{
					Field1: types.StringValue("a"),
				}),
			},
			Target: &awsPluralSliceOfNestedObjectValues{},
			WantTarget: &awsPluralSliceOfNestedObjectValues{
				Fields: []awsSingleStringValue{{Field1: "a"}},
			},
		},
		"plural field names": {
			Source: &tfSpecialPluralization{
				City: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("paris"),
					types.StringValue("london"),
				}),
				Coach: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("guardiola"),
					types.StringValue("mourinho"),
				}),
				Tomato: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("brandywine"),
					types.StringValue("roma"),
				}),
				Vertex: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("ab"),
					types.StringValue("bc"),
				}),
				Criterion: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("votes"),
					types.StringValue("editors"),
				}),
				Datum: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("d1282f78-fa99-5d9d-bd51-e6f0173eb74a"),
					types.StringValue("0f10cb10-2076-5254-bd21-d3f62fe66303"),
				}),
				Hive: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("Cegieme"),
					types.StringValue("Fahumvid"),
				}),
			},
			Target: &awsSpecialPluralization{},
			WantTarget: &awsSpecialPluralization{
				Cities: []*string{
					aws.String("paris"),
					aws.String("london"),
				},
				Coaches: []*string{
					aws.String("guardiola"),
					aws.String("mourinho"),
				},
				Tomatoes: []*string{
					aws.String("brandywine"),
					aws.String("roma"),
				},
				Vertices: []*string{
					aws.String("ab"),
					aws.String("bc"),
				},
				Criteria: []*string{
					aws.String("votes"),
					aws.String("editors"),
				},
				Data: []*string{
					aws.String("d1282f78-fa99-5d9d-bd51-e6f0173eb74a"),
					aws.String("0f10cb10-2076-5254-bd21-d3f62fe66303"),
				},
				Hives: []*string{
					aws.String("Cegieme"),
					aws.String("Fahumvid"),
				},
			},
		},
		"capitalization field names": {
			Source: &tfCaptializationDiff{
				FieldURL: types.StringValue("h"),
			},
			Target: &awsCapitalizationDiff{},
			WantTarget: &awsCapitalizationDiff{
				FieldUrl: aws.String("h"),
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestExpandOptions(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Bool                       `tfsdk:"field1"`
		Tags   fwtypes.MapValueOf[types.String] `tfsdk:"tags"`
	}
	type aws01 struct {
		Field1 bool
		Tags   map[string]string
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"empty source with tags": {
			Source:     &tf01{},
			Target:     &aws01{},
			WantTarget: &aws01{},
		},
		"ignore tags by default": {
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				},
				),
			},
			Target:     &aws01{},
			WantTarget: &aws01{Field1: true},
		},
		"include tags with option override": {
			Options: []AutoFlexOptionsFunc{WithNoIgnoredFieldNames()},
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				},
				),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: true,
				Tags:   map[string]string{"foo": "bar"},
			},
		},
		"ignore custom field": {
			Options: []AutoFlexOptionsFunc{WithIgnoredFieldNames([]string{"Field1"})},
			Source: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				},
				),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Tags: map[string]string{"foo": "bar"},
			},
		},
		"resource name suffix": {
			Options: []AutoFlexOptionsFunc{WithFieldNameSuffix("Config")},
			Source: &tfFieldNameSuffix{
				Policy: types.StringValue("foo"),
			},
			Target: &awsFieldNameSuffix{},
			WantTarget: &awsFieldNameSuffix{
				PolicyConfig: aws.String("foo"),
			},
		},
	}
	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFindFieldFuzzy_Combinations(t *testing.T) {
	t.Parallel()

	type builder func() (typeFrom reflect.Type, typeTo reflect.Type, fieldNameFrom string, expectedFieldName string)

	cases := map[string]struct {
		prefix string
		suffix string
		build  builder
	}{
		// 1) suffix-only on target; source has neither
		"suffix on target only (prefix configured but not applied)": {
			prefix: "Cluster",
			suffix: "Input",
			build: func() (reflect.Type, reflect.Type, string, string) {
				type source struct{ ExecutionConfig string }
				type target struct{ ExecutionConfigInput string }
				return reflect.TypeFor[source](), reflect.TypeFor[target](), "ExecutionConfig", "ExecutionConfigInput"
			},
		},
		// 2) trim prefix on source, then add suffix
		"trim prefix on source then add suffix": {
			prefix: "Cluster",
			suffix: "Input",
			build: func() (reflect.Type, reflect.Type, string, string) {
				type source struct{ ClusterExecutionConfig string }
				type target struct{ ExecutionConfigInput string }
				return reflect.TypeFor[source](), reflect.TypeFor[target](), "ClusterExecutionConfig", "ExecutionConfigInput"
			},
		},
		// 3) add prefix and suffix on target (source has neither)
		"add prefix and suffix on target": {
			prefix: "Cluster",
			suffix: "Input",
			build: func() (reflect.Type, reflect.Type, string, string) {
				type source struct{ ExecutionConfig string }
				type target struct{ ClusterExecutionConfigInput string }
				return reflect.TypeFor[source](), reflect.TypeFor[target](), "ExecutionConfig", "ClusterExecutionConfigInput"
			},
		},
		// 4) trim suffix on source (target has neither)
		"trim suffix on source": {
			prefix: "Cluster",
			suffix: "Input",
			build: func() (reflect.Type, reflect.Type, string, string) {
				type source struct{ ExecutionConfigInput string }
				type target struct{ ExecutionConfig string }
				return reflect.TypeFor[source](), reflect.TypeFor[target](), "ExecutionConfigInput", "ExecutionConfig"
			},
		},
		// 5) trim both on source (target has neither)
		"trim both prefix and suffix on source": {
			prefix: "Cluster",
			suffix: "Input",
			build: func() (reflect.Type, reflect.Type, string, string) {
				type source struct{ ClusterExecutionConfigInput string }
				type target struct{ ExecutionConfig string }
				return reflect.TypeFor[source](), reflect.TypeFor[target](), "ClusterExecutionConfigInput", "ExecutionConfig"
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			typeFrom, typeTo, fieldNameFrom, expected := tc.build()
			ctx := context.Background()
			opts := []AutoFlexOptionsFunc{
				WithFieldNamePrefix(tc.prefix),
				WithFieldNameSuffix(tc.suffix),
			}
			flexer := newAutoExpander(opts)

			field, found := (&fuzzyFieldFinder{}).findField(ctx, fieldNameFrom, typeFrom, typeTo, flexer)
			if !found {
				t.Fatalf("expected to find field, but found==false")
			}
			if field.Name != expected {
				t.Fatalf("expected field name %q, got %q", expected, field.Name)
			}
		})
	}
}

func TestExpandFieldNamePrefix(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"exact match": {
			Options: []AutoFlexOptionsFunc{
				WithFieldNamePrefix("Intent"),
			},
			Source: &tfFieldNamePrefix{
				Name: types.StringValue("Ovodoghen"),
			},
			Target: &awsFieldNamePrefix{},
			WantTarget: &awsFieldNamePrefix{
				IntentName: aws.String("Ovodoghen"),
			},
		},

		"case-insensitive": {
			Options: []AutoFlexOptionsFunc{
				WithFieldNamePrefix("Client"),
			},
			Source: &tfFieldNamePrefixInsensitive{
				ID: types.StringValue("abc123"),
			},
			Target: &awsFieldNamePrefixInsensitive{},
			WantTarget: &awsFieldNamePrefixInsensitive{
				ClientId: aws.String("abc123"),
			},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenNaming(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"plural ordinary field names": {
			Source: &awsPluralSliceOfNestedObjectValues{
				Fields: []awsSingleStringValue{{Field1: "a"}},
			},
			Target: &tfSingluarListOfNestedObjects{},
			WantTarget: &tfSingluarListOfNestedObjects{
				Field: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
					Field1: types.StringValue("a"),
				}),
			},
		},
		"plural field names": {
			Source: &awsSpecialPluralization{
				Cities: []*string{
					aws.String("paris"),
					aws.String("london"),
				},
				Coaches: []*string{
					aws.String("guardiola"),
					aws.String("mourinho"),
				},
				Tomatoes: []*string{
					aws.String("brandywine"),
					aws.String("roma"),
				},
				Vertices: []*string{
					aws.String("ab"),
					aws.String("bc"),
				},
				Criteria: []*string{
					aws.String("votes"),
					aws.String("editors"),
				},
				Data: []*string{
					aws.String("d1282f78-fa99-5d9d-bd51-e6f0173eb74a"),
					aws.String("0f10cb10-2076-5254-bd21-d3f62fe66303"),
				},
				Hives: []*string{
					aws.String("Cegieme"),
					aws.String("Fahumvid"),
				},
			},
			Target: &tfSpecialPluralization{},
			WantTarget: &tfSpecialPluralization{
				City: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("paris"),
					types.StringValue("london"),
				}),
				Coach: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("guardiola"),
					types.StringValue("mourinho"),
				}),
				Tomato: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("brandywine"),
					types.StringValue("roma"),
				}),
				Vertex: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("ab"),
					types.StringValue("bc"),
				}),
				Criterion: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("votes"),
					types.StringValue("editors"),
				}),
				Datum: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("d1282f78-fa99-5d9d-bd51-e6f0173eb74a"),
					types.StringValue("0f10cb10-2076-5254-bd21-d3f62fe66303"),
				}),
				Hive: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("Cegieme"),
					types.StringValue("Fahumvid"),
				}),
			},
		},
		"strange plurality": {
			Source: &awsPluralAndSingularFields{
				Value:  "a",
				Values: "b",
			},
			Target: &tfPluralAndSingularFields{},
			WantTarget: &tfPluralAndSingularFields{
				Value: types.StringValue("a"),
			},
		},
		"capitalization field names": {
			Source: &awsCapitalizationDiff{
				FieldUrl: aws.String("h"),
			},
			Target: &tfCaptializationDiff{},
			WantTarget: &tfCaptializationDiff{
				FieldURL: types.StringValue("h"),
			},
		},
		"resource name prefix": {
			Options: []AutoFlexOptionsFunc{
				WithFieldNamePrefix("Intent"),
			},
			Source: &awsFieldNamePrefix{
				IntentName: aws.String("Ovodoghen"),
			},
			Target: &tfFieldNamePrefix{},
			WantTarget: &tfFieldNamePrefix{
				Name: types.StringValue("Ovodoghen"),
			},
		},
		"resource name suffix": {
			Options: []AutoFlexOptionsFunc{WithFieldNameSuffix("Config")},
			Source: &awsFieldNameSuffix{
				PolicyConfig: aws.String("foo"),
			},
			Target: &tfFieldNameSuffix{},
			WantTarget: &tfFieldNameSuffix{
				Policy: types.StringValue("foo"),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}

func TestFlattenOptions(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Bool                       `tfsdk:"field1"`
		Tags   fwtypes.MapValueOf[types.String] `tfsdk:"tags"`
	}
	type aws01 struct {
		Field1 bool
		Tags   map[string]string
	}

	// For test cases below where a field of `MapValue` type is ignored, the
	// result of `cmp.Diff` is intentionally not checked.
	//
	// When a target contains an ignored field of a `MapValue` type, the resulting
	// target will contain a zero value, which, because the `elementType` is nil, will
	// always return `false` from the `Equal` method, even when compared with another
	// zero value. In practice, this zeroed `MapValue` would be overwritten
	// by a subsequent step (ie. transparent tagging), and the temporary invalid
	// state of the zeroed `MapValue` will not appear in the final state.
	//
	// Example expected diff:
	// 	    unexpected diff (+wanted, -got):   &flex.tf01{
	//                 Field1: s"false",
	//         -       Tags:   types.MapValueOf[github.com/hashicorp/terraform-plugin-framework/types/types.String]{},
	//         +       Tags:   types.MapValueOf[github.com/hashicorp/terraform-plugin-framework/types/types.String]{MapValue: types.Map{elementType: basetypes.StringType{}}},
	//           }
	ctx := context.Background()
	testCases := autoFlexTestCases{
		"empty source with tags": {
			Source: &aws01{},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.BoolValue(false),
				Tags:   fwtypes.NewMapValueOfNull[types.String](ctx),
			},
			WantDiff: true, // Ignored MapValue type, expect diff
		},
		"ignore tags by default": {
			Source: &aws01{
				Field1: true,
				Tags:   map[string]string{"foo": "bar"},
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.BoolValue(true),
				Tags:   fwtypes.NewMapValueOfNull[types.String](ctx),
			},
			WantDiff: true, // Ignored MapValue type, expect diff
		},
		"include tags with option override": {
			Options: []AutoFlexOptionsFunc{WithNoIgnoredFieldNames()},
			Source: &aws01{
				Field1: true,
				Tags:   map[string]string{"foo": "bar"},
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.BoolValue(true),
				Tags: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"foo": types.StringValue("bar"),
				}),
			},
		},
		"ignore custom field": {
			Options: []AutoFlexOptionsFunc{WithIgnoredFieldNames([]string{"Field1"})},
			Source: &aws01{
				Field1: true,
				Tags:   map[string]string{"foo": "bar"},
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.BoolNull(),
				Tags: fwtypes.NewMapValueOfMust[types.String](
					ctx,
					map[string]attr.Value{
						"foo": types.StringValue("bar"),
					},
				),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases, runChecks{CompareDiags: true, CompareTarget: true, SkipGoldenLogs: true})
}
