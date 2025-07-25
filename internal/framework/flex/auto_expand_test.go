// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestExpand(t *testing.T) {
	t.Parallel()

	testString := "test"
	testStringResult := "a"

	testByteSlice := []byte("test")
	testByteSliceResult := []byte("a")

	var (
		typedNilSource *emptyStruct
		typedNilTarget *emptyStruct
	)

	testARN := "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1" //lintignore:AWSAT003,AWSAT005

	testTimeStr := "2013-09-25T09:34:01Z"
	testTimeTime := errs.Must(time.Parse(time.RFC3339, testTimeStr))

	testCases := autoFlexTestCases{
		"nil Source": {
			Target: &emptyStruct{},
			expectedDiags: diag.Diagnostics{
				diagExpandingSourceIsNil(nil),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(nil, reflect.TypeFor[*emptyStruct]()),
				errorSourceIsNil("", nil, "", reflect.TypeFor[emptyStruct]()),
			},
		},
		"typed nil Source": {
			Source: typedNilSource,
			Target: &emptyStruct{},
			expectedDiags: diag.Diagnostics{
				// diagExpandingSourceIsNil(reflect.TypeFor[*emptyStruct]()),
				diagExpandingSourceIsNil(nil), // FIXME: Should give the actual type
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*emptyStruct](), reflect.TypeFor[*emptyStruct]()),
				// errorSourceIsNil("", reflect.TypeFor[*emptyStruct](), "", reflect.TypeFor[emptyStruct]()),
				errorSourceIsNil("", nil, "", reflect.TypeFor[emptyStruct]()), // FIXME: Should give the actual type
			},
		},
		"nil Target": {
			Source: emptyStruct{},
			expectedDiags: diag.Diagnostics{
				diagConvertingTargetIsNil(nil),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[emptyStruct](), nil),
				errorTargetIsNil("", reflect.TypeFor[emptyStruct](), "", nil),
			},
		},
		"typed nil Target": {
			Source: emptyStruct{},
			Target: typedNilTarget,
			expectedDiags: diag.Diagnostics{
				diagConvertingTargetIsNil(reflect.TypeFor[*emptyStruct]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*emptyStruct]()),
				errorTargetIsNil("", reflect.TypeFor[emptyStruct](), "", reflect.TypeFor[*emptyStruct]()),
			},
		},
		"non-pointer Target": {
			Source: emptyStruct{},
			Target: 0,
			expectedDiags: diag.Diagnostics{
				diagConvertingTargetIsNotPointer(reflect.TypeFor[int]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[emptyStruct](), reflect.TypeFor[int]()),
				errorTargetIsNotPointer("", reflect.TypeFor[emptyStruct](), "", reflect.TypeFor[int]()),
			},
		},
		"non-struct Source struct Target": {
			Source: testString,
			Target: &emptyStruct{},
			expectedDiags: diag.Diagnostics{
				diagExpandingSourceDoesNotImplementAttrValue(reflect.TypeFor[string]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[string](), reflect.TypeFor[*emptyStruct]()),
				infoConverting(reflect.TypeFor[string](), reflect.TypeFor[emptyStruct]()),
				errorSourceDoesNotImplementAttrValue("", reflect.TypeFor[string](), "", reflect.TypeFor[emptyStruct]()),
			},
		},
		"struct Source non-struct Target": {
			Source: emptyStruct{},
			Target: &testString,
			expectedDiags: diag.Diagnostics{
				diagExpandingSourceDoesNotImplementAttrValue(reflect.TypeFor[emptyStruct]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*string]()),
				infoConverting(reflect.TypeFor[emptyStruct](), reflect.TypeFor[string]()),
				errorSourceDoesNotImplementAttrValue("", reflect.TypeFor[emptyStruct](), "", reflect.TypeFor[string]()),
			},
		},
		"types.String to string": {
			Source:     types.StringValue("a"),
			Target:     &testString,
			WantTarget: &testStringResult,
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.String](), reflect.TypeFor[*string]()),
				infoConverting(reflect.TypeFor[types.String](), reflect.TypeFor[string]()),
			},
		},
		"types.String to byte slice": {
			Source:     types.StringValue("a"),
			Target:     &testByteSlice,
			WantTarget: &testByteSliceResult,
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.String](), reflect.TypeFor[*[]byte]()),
				infoConverting(reflect.TypeFor[types.String](), reflect.TypeFor[[]byte]()),
			},
		},
		"empty struct Source and Target": {
			Source:     emptyStruct{},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*emptyStruct]()),
				infoConverting(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*emptyStruct]()),
			},
		},
		"empty struct pointer Source and Target": {
			Source:     &emptyStruct{},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*emptyStruct](), reflect.TypeFor[*emptyStruct]()),
				infoConverting(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*emptyStruct]()),
			},
		},
		"single string struct pointer Source and empty Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSingleStringField](), reflect.TypeFor[*emptyStruct]()),
				infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*emptyStruct]()),
				debugNoCorrespondingField(reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*emptyStruct]()),
			},
		},
		"source field does not implement attr.Value Source": {
			Source: &awsSingleStringValue{Field1: "a"},
			Target: &awsSingleStringValue{},
			expectedDiags: diag.Diagnostics{
				diagExpandingSourceDoesNotImplementAttrValue(reflect.TypeFor[string]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*awsSingleStringValue](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*awsSingleStringValue]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[string]()),
				errorSourceDoesNotImplementAttrValue("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"single string Source and single string Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{Field1: "a"},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSingleStringField](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringValue]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"single string Source and byte slice Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &awsSingleByteSliceValue{},
			WantTarget: &awsSingleByteSliceValue{Field1: []byte("a")},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSingleStringField](), reflect.TypeFor[*awsSingleByteSliceValue]()),
				infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleByteSliceValue]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleByteSliceValue]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[[]byte]()),
			},
		},
		"single string Source and single *string Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &awsSingleStringPointer{},
			WantTarget: &awsSingleStringPointer{Field1: aws.String("a")},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSingleStringField](), reflect.TypeFor[*awsSingleStringPointer]()),
				infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringPointer]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
			},
		},
		"single string Source and single int64 Target": {
			Source:     &tfSingleStringField{Field1: types.StringValue("a")},
			Target:     &awsSingleInt64Value{},
			WantTarget: &awsSingleInt64Value{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSingleStringField](), reflect.TypeFor[*awsSingleInt64Value]()),
				infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleInt64Value]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[int64]()),
				{
					"@level":             "error",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Int64),
					logAttrKeySourcePath: "Field1",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[types.String]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[int64]()),
				},
			},
		},
		"primitive types Source and primtive types Target": {
			Source: &tfAllThePrimitiveFields{
				Field1:  types.StringValue("field1"),
				Field2:  types.StringValue("field2"),
				Field3:  types.Int64Value(3),
				Field4:  types.Int64Value(-4),
				Field5:  types.Int64Value(5),
				Field6:  types.Int64Value(-6),
				Field7:  types.Float64Value(7.7),
				Field8:  types.Float64Value(-8.8),
				Field9:  types.Float64Value(9.99),
				Field10: types.Float64Value(-10.101),
				Field11: types.BoolValue(true),
				Field12: types.BoolValue(false),
			},
			Target: &awsAllThePrimitiveFields{},
			WantTarget: &awsAllThePrimitiveFields{
				Field1:  "field1",
				Field2:  aws.String("field2"),
				Field3:  3,
				Field4:  aws.Int32(-4),
				Field5:  5,
				Field6:  aws.Int64(-6),
				Field7:  7.7,
				Field8:  aws.Float32(-8.8),
				Field9:  9.99,
				Field10: aws.Float64(-10.101),
				Field11: true,
				Field12: aws.Bool(false),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfAllThePrimitiveFields](), reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConverting(reflect.TypeFor[tfAllThePrimitiveFields](), reflect.TypeFor[*awsAllThePrimitiveFields]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfAllThePrimitiveFields](), "Field1", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				traceMatchedFields("Field2", reflect.TypeFor[tfAllThePrimitiveFields](), "Field2", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[types.String](), "Field2", reflect.TypeFor[*string]()),
				traceMatchedFields("Field3", reflect.TypeFor[tfAllThePrimitiveFields](), "Field3", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[types.Int64](), "Field3", reflect.TypeFor[int32]()),
				traceMatchedFields("Field4", reflect.TypeFor[tfAllThePrimitiveFields](), "Field4", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[types.Int64](), "Field4", reflect.TypeFor[*int32]()),
				traceMatchedFields("Field5", reflect.TypeFor[tfAllThePrimitiveFields](), "Field5", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[types.Int64](), "Field5", reflect.TypeFor[int64]()),
				traceMatchedFields("Field6", reflect.TypeFor[tfAllThePrimitiveFields](), "Field6", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[types.Int64](), "Field6", reflect.TypeFor[*int64]()),
				traceMatchedFields("Field7", reflect.TypeFor[tfAllThePrimitiveFields](), "Field7", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field7", reflect.TypeFor[types.Float64](), "Field7", reflect.TypeFor[float32]()),
				traceMatchedFields("Field8", reflect.TypeFor[tfAllThePrimitiveFields](), "Field8", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field8", reflect.TypeFor[types.Float64](), "Field8", reflect.TypeFor[*float32]()),
				traceMatchedFields("Field9", reflect.TypeFor[tfAllThePrimitiveFields](), "Field9", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field9", reflect.TypeFor[types.Float64](), "Field9", reflect.TypeFor[float64]()),
				traceMatchedFields("Field10", reflect.TypeFor[tfAllThePrimitiveFields](), "Field10", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field10", reflect.TypeFor[types.Float64](), "Field10", reflect.TypeFor[*float64]()),
				traceMatchedFields("Field11", reflect.TypeFor[tfAllThePrimitiveFields](), "Field11", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field11", reflect.TypeFor[types.Bool](), "Field11", reflect.TypeFor[bool]()),
				traceMatchedFields("Field12", reflect.TypeFor[tfAllThePrimitiveFields](), "Field12", reflect.TypeFor[*awsAllThePrimitiveFields]()),
				infoConvertingWithPath("Field12", reflect.TypeFor[types.Bool](), "Field12", reflect.TypeFor[*bool]()),
			},
		},
		"Collection of primitive types Source and slice or map of primtive types Target": {
			Source: &tfCollectionsOfPrimitiveElements{
				Field1: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field2: types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field3: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field4: types.SetValueMust(types.StringType, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field5: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
				Field6: types.MapValueMust(types.StringType, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
			},
			Target: &awsCollectionsOfPrimitiveElements{},
			WantTarget: &awsCollectionsOfPrimitiveElements{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfCollectionsOfPrimitiveElements](), reflect.TypeFor[*awsCollectionsOfPrimitiveElements]()),
				infoConverting(reflect.TypeFor[tfCollectionsOfPrimitiveElements](), reflect.TypeFor[*awsCollectionsOfPrimitiveElements]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfCollectionsOfPrimitiveElements](), "Field1", reflect.TypeFor[*awsCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.List](), "Field1", reflect.TypeFor[[]string]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[types.List](), 2, "Field1", reflect.TypeFor[[]string]()),
				traceMatchedFields("Field2", reflect.TypeFor[tfCollectionsOfPrimitiveElements](), "Field2", reflect.TypeFor[*awsCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[types.List](), "Field2", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Field2", reflect.TypeFor[types.List](), 2, "Field2", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Field3", reflect.TypeFor[tfCollectionsOfPrimitiveElements](), "Field3", reflect.TypeFor[*awsCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[types.Set](), "Field3", reflect.TypeFor[[]string]()),
				traceExpandingWithElementsAs("Field3", reflect.TypeFor[types.Set](), 2, "Field3", reflect.TypeFor[[]string]()),
				traceMatchedFields("Field4", reflect.TypeFor[tfCollectionsOfPrimitiveElements](), "Field4", reflect.TypeFor[*awsCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[types.Set](), "Field4", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Field4", reflect.TypeFor[types.Set](), 2, "Field4", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Field5", reflect.TypeFor[tfCollectionsOfPrimitiveElements](), "Field5", reflect.TypeFor[*awsCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[types.Map](), "Field5", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("Field5", reflect.TypeFor[types.Map](), 2, "Field5", reflect.TypeFor[map[string]string]()),
				traceMatchedFields("Field6", reflect.TypeFor[tfCollectionsOfPrimitiveElements](), "Field6", reflect.TypeFor[*awsCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[types.Map](), "Field6", reflect.TypeFor[map[string]*string]()),
				traceExpandingWithElementsAs("Field6", reflect.TypeFor[types.Map](), 2, "Field6", reflect.TypeFor[map[string]*string]()),
			},
		},
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSingluarListOfNestedObjects](), reflect.TypeFor[*awsPluralSliceOfNestedObjectValues]()),
				infoConverting(reflect.TypeFor[tfSingluarListOfNestedObjects](), reflect.TypeFor[*awsPluralSliceOfNestedObjectValues]()),
				traceMatchedFields("Field", reflect.TypeFor[tfSingluarListOfNestedObjects](), "Fields", reflect.TypeFor[*awsPluralSliceOfNestedObjectValues]()),
				infoConvertingWithPath("Field", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Fields", reflect.TypeFor[[]awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("Field", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 1, "Fields", reflect.TypeFor[[]awsSingleStringValue]()),
				traceMatchedFieldsWithPath("Field[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "Fields[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("Field[0].Field1", reflect.TypeFor[types.String](), "Fields[0].Field1", reflect.TypeFor[string]()),
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSpecialPluralization](), reflect.TypeFor[*awsSpecialPluralization]()),
				infoConverting(reflect.TypeFor[tfSpecialPluralization](), reflect.TypeFor[*awsSpecialPluralization]()),
				traceMatchedFields("City", reflect.TypeFor[tfSpecialPluralization](), "Cities", reflect.TypeFor[*awsSpecialPluralization]()),
				infoConvertingWithPath("City", reflect.TypeFor[types.List](), "Cities", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("City", reflect.TypeFor[types.List](), 2, "Cities", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Coach", reflect.TypeFor[tfSpecialPluralization](), "Coaches", reflect.TypeFor[*awsSpecialPluralization]()),
				infoConvertingWithPath("Coach", reflect.TypeFor[types.List](), "Coaches", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Coach", reflect.TypeFor[types.List](), 2, "Coaches", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Tomato", reflect.TypeFor[tfSpecialPluralization](), "Tomatoes", reflect.TypeFor[*awsSpecialPluralization]()),
				infoConvertingWithPath("Tomato", reflect.TypeFor[types.List](), "Tomatoes", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Tomato", reflect.TypeFor[types.List](), 2, "Tomatoes", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Vertex", reflect.TypeFor[tfSpecialPluralization](), "Vertices", reflect.TypeFor[*awsSpecialPluralization]()),
				infoConvertingWithPath("Vertex", reflect.TypeFor[types.List](), "Vertices", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Vertex", reflect.TypeFor[types.List](), 2, "Vertices", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Criterion", reflect.TypeFor[tfSpecialPluralization](), "Criteria", reflect.TypeFor[*awsSpecialPluralization]()),
				infoConvertingWithPath("Criterion", reflect.TypeFor[types.List](), "Criteria", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Criterion", reflect.TypeFor[types.List](), 2, "Criteria", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Datum", reflect.TypeFor[tfSpecialPluralization](), "Data", reflect.TypeFor[*awsSpecialPluralization]()),
				infoConvertingWithPath("Datum", reflect.TypeFor[types.List](), "Data", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Datum", reflect.TypeFor[types.List](), 2, "Data", reflect.TypeFor[[]*string]()),
				traceMatchedFields("Hive", reflect.TypeFor[tfSpecialPluralization](), "Hives", reflect.TypeFor[*awsSpecialPluralization]()),
				infoConvertingWithPath("Hive", reflect.TypeFor[types.List](), "Hives", reflect.TypeFor[[]*string]()),
				traceExpandingWithElementsAs("Hive", reflect.TypeFor[types.List](), 2, "Hives", reflect.TypeFor[[]*string]()),
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfCaptializationDiff](), reflect.TypeFor[*awsCapitalizationDiff]()),
				infoConverting(reflect.TypeFor[tfCaptializationDiff](), reflect.TypeFor[*awsCapitalizationDiff]()),
				traceMatchedFields("FieldURL", reflect.TypeFor[tfCaptializationDiff](), "FieldUrl", reflect.TypeFor[*awsCapitalizationDiff]()),
				infoConvertingWithPath("FieldURL", reflect.TypeFor[types.String](), "FieldUrl", reflect.TypeFor[*string]()),
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfFieldNameSuffix](), reflect.TypeFor[*awsFieldNameSuffix]()),
				infoConverting(reflect.TypeFor[tfFieldNameSuffix](), reflect.TypeFor[*awsFieldNameSuffix]()),
				traceMatchedFields("Policy", reflect.TypeFor[tfFieldNameSuffix](), "PolicyConfig", reflect.TypeFor[*awsFieldNameSuffix]()),
				infoConvertingWithPath("Policy", reflect.TypeFor[types.String](), "PolicyConfig", reflect.TypeFor[*string]()),
			},
		},
		"single ARN Source and single string Target": {
			Source:     &tfSingleARNField{Field1: fwtypes.ARNValue(testARN)},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{Field1: testARN},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSingleARNField](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[tfSingleARNField](), reflect.TypeFor[*awsSingleStringValue]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSingleARNField](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ARN](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"single ARN Source and single *string Target": {
			Source:     &tfSingleARNField{Field1: fwtypes.ARNValue(testARN)},
			Target:     &awsSingleStringPointer{},
			WantTarget: &awsSingleStringPointer{Field1: aws.String(testARN)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSingleARNField](), reflect.TypeFor[*awsSingleStringPointer]()),
				infoConverting(reflect.TypeFor[tfSingleARNField](), reflect.TypeFor[*awsSingleStringPointer]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSingleARNField](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ARN](), "Field1", reflect.TypeFor[*string]()),
			},
		},
		"timestamp pointer": {
			Source: &tfRFC3339Time{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			Target: &awsRFC3339TimePointer{},
			WantTarget: &awsRFC3339TimePointer{
				CreationDateTime: &testTimeTime,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfRFC3339Time](), reflect.TypeFor[*awsRFC3339TimePointer]()),
				infoConverting(reflect.TypeFor[tfRFC3339Time](), reflect.TypeFor[*awsRFC3339TimePointer]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[tfRFC3339Time](), "CreationDateTime", reflect.TypeFor[*awsRFC3339TimePointer]()),
				infoConvertingWithPath("CreationDateTime", reflect.TypeFor[timetypes.RFC3339](), "CreationDateTime", reflect.TypeFor[*time.Time]()),
			},
		},
		"timestamp": {
			Source: &tfRFC3339Time{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			Target: &awsRFC3339TimeValue{},
			WantTarget: &awsRFC3339TimeValue{
				CreationDateTime: testTimeTime,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfRFC3339Time](), reflect.TypeFor[*awsRFC3339TimeValue]()),
				infoConverting(reflect.TypeFor[tfRFC3339Time](), reflect.TypeFor[*awsRFC3339TimeValue]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[tfRFC3339Time](), "CreationDateTime", reflect.TypeFor[*awsRFC3339TimeValue]()),
				infoConvertingWithPath("CreationDateTime", reflect.TypeFor[timetypes.RFC3339](), "CreationDateTime", reflect.TypeFor[time.Time]()),
			},
		},
		"JSONValue Source to json interface Target": {
			Source: &tfJSONStringer{Field1: fwtypes.SmithyJSONValue(`{"field1": "a"}`, newTestJSONDocument)},
			Target: &awsJSONStringer{},
			WantTarget: &awsJSONStringer{
				Field1: &testJSONDocument{
					Value: map[string]any{
						"field1": "a",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfJSONStringer](), reflect.TypeFor[*awsJSONStringer]()),
				infoConverting(reflect.TypeFor[tfJSONStringer](), reflect.TypeFor[*awsJSONStringer]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfJSONStringer](), "Field1", reflect.TypeFor[*awsJSONStringer]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]](), "Field1", reflect.TypeFor[smithyjson.JSONStringer]()),
			},
		},
	}

	runAutoExpandTestCases(t, testCases)
}

func TestExpandGeneric(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"complex Source and complex Target": {
			Source: &tfComplexValue{
				Field1: types.StringValue("m"),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("n"),
					}),
				}),
				Field3: types.MapValueMust(types.StringType, map[string]attr.Value{
					"X": types.StringValue("x"),
					"Y": types.StringValue("y"),
				}),
				Field4: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleInt64Field{
					{Field1: types.Int64Value(100)},
					{Field1: types.Int64Value(2000)},
					{Field1: types.Int64Value(30000)},
				}),
			},
			Target: &awsComplexValue{},
			WantTarget: &awsComplexValue{
				Field1: "m",
				Field2: &awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "n",
					},
				},
				Field3: aws.StringMap(map[string]string{
					"X": "x",
					"Y": "y",
				}),
				Field4: []awsSingleInt64Value{
					{Field1: 100},
					{Field1: 2000},
					{Field1: 30000},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfComplexValue](), reflect.TypeFor[*awsComplexValue]()),
				infoConverting(reflect.TypeFor[tfComplexValue](), reflect.TypeFor[*awsComplexValue]()),

				traceMatchedFields("Field1", reflect.TypeFor[tfComplexValue](), "Field1", reflect.TypeFor[*awsComplexValue]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				traceMatchedFields("Field2", reflect.TypeFor[tfComplexValue](), "Field2", reflect.TypeFor[*awsComplexValue]()),

				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfListOfNestedObject]](), "Field2", reflect.TypeFor[*awsNestedObjectPointer]()),
				traceMatchedFieldsWithPath("Field2[0]", "Field1", reflect.TypeFor[tfListOfNestedObject](), "Field2", "Field1", reflect.TypeFor[*awsNestedObjectPointer]()),
				infoConvertingWithPath("Field2[0].Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Field2.Field1", reflect.TypeFor[*awsSingleStringValue]()),
				traceMatchedFieldsWithPath("Field2[0].Field1[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field2.Field1", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("Field2[0].Field1[0].Field1", reflect.TypeFor[types.String](), "Field2.Field1.Field1", reflect.TypeFor[string]()),

				traceMatchedFields("Field3", reflect.TypeFor[tfComplexValue](), "Field3", reflect.TypeFor[*awsComplexValue]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[types.Map](), "Field3", reflect.TypeFor[map[string]*string]()),
				traceExpandingWithElementsAs("Field3", reflect.TypeFor[types.Map](), 2, "Field3", reflect.TypeFor[map[string]*string]()),

				traceMatchedFields("Field4", reflect.TypeFor[tfComplexValue](), "Field4", reflect.TypeFor[*awsComplexValue]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleInt64Field]](), "Field4", reflect.TypeFor[[]awsSingleInt64Value]()),
				traceExpandingNestedObjectCollection("Field4", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleInt64Field]](), 3, "Field4", reflect.TypeFor[[]awsSingleInt64Value]()),
				traceMatchedFieldsWithPath("Field4[0]", "Field1", reflect.TypeFor[tfSingleInt64Field](), "Field4[0]", "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
				infoConvertingWithPath("Field4[0].Field1", reflect.TypeFor[types.Int64](), "Field4[0].Field1", reflect.TypeFor[int64]()),
				traceMatchedFieldsWithPath("Field4[1]", "Field1", reflect.TypeFor[tfSingleInt64Field](), "Field4[1]", "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
				infoConvertingWithPath("Field4[1].Field1", reflect.TypeFor[types.Int64](), "Field4[1].Field1", reflect.TypeFor[int64]()),
				traceMatchedFieldsWithPath("Field4[2]", "Field1", reflect.TypeFor[tfSingleInt64Field](), "Field4[2]", "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
				infoConvertingWithPath("Field4[2].Field1", reflect.TypeFor[types.Int64](), "Field4[2].Field1", reflect.TypeFor[int64]()),
			},
		},
		"map of string": {
			Source: &tfMapOfString{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target: &awsMapOfString{},
			WantTarget: &awsMapOfString{
				FieldInner: map[string]string{
					"x": "y",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapOfString](), reflect.TypeFor[*awsMapOfString]()),
				infoConverting(reflect.TypeFor[tfMapOfString](), reflect.TypeFor[*awsMapOfString]()),
				traceMatchedFields("FieldInner", reflect.TypeFor[tfMapOfString](), "FieldInner", reflect.TypeFor[*awsMapOfString]()),
				infoConvertingWithPath("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "FieldInner", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "FieldInner", reflect.TypeFor[map[string]string]()),
			},
		},
		"map of string pointer": {
			Source: &tfMapOfString{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			Target: &awsMapOfStringPointer{},
			WantTarget: &awsMapOfStringPointer{
				FieldInner: map[string]*string{
					"x": aws.String("y"),
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapOfString](), reflect.TypeFor[*awsMapOfStringPointer]()),
				infoConverting(reflect.TypeFor[tfMapOfString](), reflect.TypeFor[*awsMapOfStringPointer]()),
				traceMatchedFields("FieldInner", reflect.TypeFor[tfMapOfString](), "FieldInner", reflect.TypeFor[*awsMapOfStringPointer]()),
				infoConvertingWithPath("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "FieldInner", reflect.TypeFor[map[string]*string]()),
				traceExpandingWithElementsAs("FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "FieldInner", reflect.TypeFor[map[string]*string]()),
			},
		},
		"map of map of string": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target: &awsMapOfMapOfString{},
			WantTarget: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapOfMapOfString](), reflect.TypeFor[*awsMapOfMapOfString]()),
				infoConverting(reflect.TypeFor[tfMapOfMapOfString](), reflect.TypeFor[*awsMapOfMapOfString]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfMapOfMapOfString](), "Field1", reflect.TypeFor[*awsMapOfMapOfString]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]](), "Field1", reflect.TypeFor[map[string]map[string]string]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]](), 1, "Field1", reflect.TypeFor[map[string]map[string]string]()),
			},
		},
		"map of map of string pointer": {
			Source: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			Target: &awsMapOfMapOfStringPointer{},
			WantTarget: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapOfMapOfString](), reflect.TypeFor[*awsMapOfMapOfStringPointer]()),
				infoConverting(reflect.TypeFor[tfMapOfMapOfString](), reflect.TypeFor[*awsMapOfMapOfStringPointer]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfMapOfMapOfString](), "Field1", reflect.TypeFor[*awsMapOfMapOfStringPointer]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]](), "Field1", reflect.TypeFor[map[string]map[string]*string]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]](), 1, "Field1", reflect.TypeFor[map[string]map[string]*string]()),
			},
		},
		"nested string map": {
			Source: &tfNestedMapOfString{
				FieldOuter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfMapOfString{
					FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"x": types.StringValue("y"),
					}),
				}),
			},
			Target: &awsNestedMapOfString{},
			WantTarget: &awsNestedMapOfString{
				FieldOuter: awsMapOfString{
					FieldInner: map[string]string{
						"x": "y",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfNestedMapOfString](), reflect.TypeFor[*awsNestedMapOfString]()),
				infoConverting(reflect.TypeFor[tfNestedMapOfString](), reflect.TypeFor[*awsNestedMapOfString]()),
				traceMatchedFields("FieldOuter", reflect.TypeFor[tfNestedMapOfString](), "FieldOuter", reflect.TypeFor[*awsNestedMapOfString]()),
				infoConvertingWithPath("FieldOuter", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapOfString]](), "FieldOuter", reflect.TypeFor[awsMapOfString]()),
				traceMatchedFieldsWithPath("FieldOuter[0]", "FieldInner", reflect.TypeFor[tfMapOfString](), "FieldOuter", "FieldInner", reflect.TypeFor[*awsMapOfString]()),
				infoConvertingWithPath("FieldOuter[0].FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "FieldOuter.FieldInner", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("FieldOuter[0].FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "FieldOuter.FieldInner", reflect.TypeFor[map[string]string]()),
			},
		},
	}

	runAutoExpandTestCases(t, testCases)
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfFieldNamePrefix](), reflect.TypeFor[*awsFieldNamePrefix]()),
				infoConverting(reflect.TypeFor[tfFieldNamePrefix](), reflect.TypeFor[*awsFieldNamePrefix]()),
				traceMatchedFields("Name", reflect.TypeFor[tfFieldNamePrefix](), "IntentName", reflect.TypeFor[*awsFieldNamePrefix]()),
				infoConvertingWithPath("Name", reflect.TypeFor[types.String](), "IntentName", reflect.TypeFor[*string]()),
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfFieldNamePrefixInsensitive](), reflect.TypeFor[*awsFieldNamePrefixInsensitive]()),
				infoConverting(reflect.TypeFor[tfFieldNamePrefixInsensitive](), reflect.TypeFor[*awsFieldNamePrefixInsensitive]()),
				traceMatchedFields("ID", reflect.TypeFor[tfFieldNamePrefixInsensitive](), "ClientId", reflect.TypeFor[*awsFieldNamePrefixInsensitive]()),
				infoConvertingWithPath("ID", reflect.TypeFor[types.String](), "ClientId", reflect.TypeFor[*string]()),
			},
		},
	}

	runAutoExpandTestCases(t, testCases)
}

func TestExpandBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Bool to bool": {
			"true": {
				Source: tfSingleBoolField{
					Field1: types.BoolValue(true),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: true,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolValue]()),
					infoConverting(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolField](), "Field1", reflect.TypeFor[*awsSingleBoolValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				},
			},
			"false": {
				Source: tfSingleBoolField{
					Field1: types.BoolValue(false),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: false,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolValue]()),
					infoConverting(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolField](), "Field1", reflect.TypeFor[*awsSingleBoolValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				},
			},
			"null": {
				Source: tfSingleBoolField{
					Field1: types.BoolNull(),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: false,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolValue]()),
					infoConverting(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolField](), "Field1", reflect.TypeFor[*awsSingleBoolValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				},
			},
		},

		"legacy Bool to bool": {
			"true": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(true),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: true,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolValue]()),
					infoConverting(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleBoolValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				},
			},
			"false": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(false),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: false,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolValue]()),
					infoConverting(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleBoolValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				},
			},
			"null": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolNull(),
				},
				Target: &awsSingleBoolValue{},
				WantTarget: &awsSingleBoolValue{
					Field1: false,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolValue]()),
					infoConverting(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleBoolValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				},
			},
		},

		"Bool to *bool": {
			"true": {
				Source: tfSingleBoolField{
					Field1: types.BoolValue(true),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: aws.Bool(true),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConverting(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolField](), "Field1", reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
				},
			},
			"false": {
				Source: tfSingleBoolField{
					Field1: types.BoolValue(false),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: aws.Bool(false),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConverting(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolField](), "Field1", reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
				},
			},
			"null": {
				Source: tfSingleBoolField{
					Field1: types.BoolNull(),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConverting(reflect.TypeFor[tfSingleBoolField](), reflect.TypeFor[*awsSingleBoolPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolField](), "Field1", reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
				},
			},
		},

		"legacy Bool to *bool": {
			"true": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(true),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: aws.Bool(true),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConverting(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
				},
			},
			"false": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(false),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConverting(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
				},
			},
			"null": {
				Source: tfSingleBoolFieldLegacy{
					Field1: types.BoolNull(),
				},
				Target: &awsSingleBoolPointer{},
				WantTarget: &awsSingleBoolPointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConverting(reflect.TypeFor[tfSingleBoolFieldLegacy](), reflect.TypeFor[*awsSingleBoolPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleBoolFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleBoolPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[*bool]()),
					// TODO: should log about legacy expander
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases)
		})
	}
}

func TestExpandFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Float64 to float64": {
			"value": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float64]()),
				},
			},
			"zero": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float64]()),
				},
			},
			"null": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float64]()),
				},
			},
		},

		"legacy Float64 to float64": {
			"value": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float64]()),
				},
			},
			"zero": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float64]()),
				},
			},
			"null": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float64]()),
				},
			},
		},

		"Float64 to *float64": {
			"value": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
			"zero": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: aws.Float64(0),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
			"null": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
		},

		"legacy Float64 to *float64": {
			"value": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
			"zero": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
			"null": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float64]()),
					// TODO: should log about legacy expander
				},
			},
		},

		// For historical reasons, Float64 can be expanded to float32 values
		"Float64 to float32": {
			"value": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float32]()),
				},
			},
			"zero": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float32]()),
				},
			},
			"null": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float32]()),
				},
			},
		},

		"legacy Float64 to float32": {
			"value": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float32]()),
				},
			},
			"zero": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float32]()),
				},
			},
			"null": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[float32]()),
				},
			},
		},

		"Float64 to *float32": {
			"value": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
			"zero": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
			"null": {
				Source: tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
		},

		"legacy Float64 to *float32": {
			"value": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
			"zero": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
			"null": {
				Source: tfSingleFloat64FieldLegacy{
					Field1: types.Float64Null(),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat64FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float64](), "Field1", reflect.TypeFor[*float32]()),
					// TODO: should log about legacy expander
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases)
		})
	}
}

func TestExpandFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Float32 to float32": {
			"value": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float32]()),
				},
			},
			"zero": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float32]()),
				},
			},
			"null": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float32]()),
				},
			},
		},

		"legacy Float32 to float32": {
			"value": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float32]()),
				},
			},
			"zero": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float32]()),
				},
			},
			"null": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat32Value{},
				WantTarget: &awsSingleFloat32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float32]()),
				},
			},
		},

		"Float32 to *float32": {
			"value": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
			"zero": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
			"null": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
		},

		"legacy Float32 to *float32": {
			"value": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
			"zero": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
				},
			},
			"null": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat32Pointer{},
				WantTarget: &awsSingleFloat32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float32]()),
					// TODO: should log about legacy expander
				},
			},
		},

		// Float32 cannot be expanded to float64
		"Float32 to float64": {
			"value": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat64Value{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[float64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
				},
			},
			"zero": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat64Value{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[float64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
				},
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
				},
			},
		},

		"legacy Float32 to float64": {
			"value": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat64Value{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[float64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
				},
			},
			"zero": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat64Value{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[float64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
				},
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat64Value{},
				WantTarget: &awsSingleFloat64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[float64]()),
				},
			},
		},

		"Float32 to *float64": {
			"value": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat64Pointer{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[*float64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
			"zero": {
				Source: tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat64Pointer{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[*float64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32Field](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32Field](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
		},

		"legacy Float32 to *float64": {
			"value": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				Target: &awsSingleFloat64Pointer{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[*float64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
			"zero": {
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				Target: &awsSingleFloat64Pointer{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[*float64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleFloat32FieldLegacy{
					Field1: types.Float32Null(),
				},
				Target: &awsSingleFloat64Pointer{},
				WantTarget: &awsSingleFloat64Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleFloat32FieldLegacy](), reflect.TypeFor[*awsSingleFloat64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleFloat32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleFloat64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Float32](), "Field1", reflect.TypeFor[*float64]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases)
		})
	}
}

func TestExpandInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Int64 to int64": {
			"value": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				},
			},
			"zero": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				},
			},
			"null": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				},
			},
		},

		"legacy Int64 to int64": {
			"value": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				},
			},
			"zero": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				},
			},
			"null": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				},
			},
		},

		"Int64 to *int64": {
			"value": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
			"zero": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: aws.Int64(0),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
			"null": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
		},

		"legacy Int64 to *int64": {
			"value": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
			"zero": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
			"null": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int64]()),
					// TODO: should log about legacy expander
				},
			},
		},

		// For historical reasons, Int64 can be expanded to int32 values
		"Int64 to int32": {
			"value": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int32]()),
				},
			},
			"zero": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int32]()),
				},
			},
			"null": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int32]()),
				},
			},
		},

		"legacy Int64 to int32": {
			"value": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int32]()),
				},
			},
			"zero": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int32]()),
				},
			},
			"null": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int32]()),
				},
			},
		},

		"Int64 to *int32": {
			"value": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
			"zero": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
			"null": {
				Source: tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64Field](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
		},

		"legacy Int64 to *int32": {
			"value": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
			"zero": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
			"null": {
				Source: tfSingleInt64FieldLegacy{
					Field1: types.Int64Null(),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt64FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt64FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[*int32]()),
					// TODO: should log about legacy expander
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases)
		})
	}
}

func TestExpandInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"Int32 to int32": {
			"value": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int32]()),
				},
			},
			"zero": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int32]()),
				},
			},
			"null": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int32]()),
				},
			},
		},

		"legacy Int32 to int32": {
			"value": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 42,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int32]()),
				},
			},
			"zero": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int32]()),
				},
			},
			"null": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Null(),
				},
				Target: &awsSingleInt32Value{},
				WantTarget: &awsSingleInt32Value{
					Field1: 0,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int32]()),
				},
			},
		},

		"Int32 to *int32": {
			"value": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
			"zero": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
			"null": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
		},

		"legacy Int32 to *int32": {
			"value": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
			"zero": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
				},
			},
			"null": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Null(),
				},
				Target: &awsSingleInt32Pointer{},
				WantTarget: &awsSingleInt32Pointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt32Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt32Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int32]()),
					// TODO: should log about legacy expander
				},
			},
		},

		// Int32 cannot be expanded to int64
		"Int32 to int64": {
			"value": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt64Value{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[int64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
				},
			},
			"zero": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt64Value{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[int64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
				},
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				Target:     &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
				},
			},
		},

		"legacy Int32 to int64": {
			"value": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt64Value{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[int64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
				},
			},
			"zero": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt64Value{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[int64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
				},
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Null(),
				},
				Target:     &awsSingleInt64Value{},
				WantTarget: &awsSingleInt64Value{},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Value]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Value]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[int64]()),
				},
			},
		},

		"Int32 to *int64": {
			"value": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt64Pointer{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[*int64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
			"zero": {
				Source: tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt64Pointer{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[*int64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				Target:     &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32Field](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32Field](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
		},

		"legacy Int32 to *int64": {
			"value": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				Target: &awsSingleInt64Pointer{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[*int64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
			"zero": {
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				Target: &awsSingleInt64Pointer{},
				expectedDiags: diag.Diagnostics{
					diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[*int64]()),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
					errorExpandingIncompatibleTypes("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
			"null": {
				// TODO: The test for a null value happens before type checking
				Source: tfSingleInt32FieldLegacy{
					Field1: types.Int32Null(),
				},
				Target:     &awsSingleInt64Pointer{},
				WantTarget: &awsSingleInt64Pointer{},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConverting(reflect.TypeFor[tfSingleInt32FieldLegacy](), reflect.TypeFor[*awsSingleInt64Pointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleInt32FieldLegacy](), "Field1", reflect.TypeFor[*awsSingleInt64Pointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.Int32](), "Field1", reflect.TypeFor[*int64]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases)
		})
	}
}

func TestExpandString(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"String to string": {
			"value": {
				Source: tfSingleStringField{
					Field1: types.StringValue("value"),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "value",
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringValue]()),
					infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				},
			},
			"empty": {
				Source: tfSingleStringField{
					Field1: types.StringValue(""),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "",
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringValue]()),
					infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				},
			},
			"null": {
				Source: tfSingleStringField{
					Field1: types.StringNull(),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "",
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringValue]()),
					infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				},
			},
		},

		"legacy String to string": {
			"value": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringValue("value"),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "value",
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringValue]()),
					infoConverting(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				},
			},
			"empty": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringValue(""),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "",
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringValue]()),
					infoConverting(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				},
			},
			"null": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringNull(),
				},
				Target: &awsSingleStringValue{},
				WantTarget: &awsSingleStringValue{
					Field1: "",
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringValue]()),
					infoConverting(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringValue]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				},
			},
		},

		"String to *string": {
			"value": {
				Source: tfSingleStringField{
					Field1: types.StringValue("value"),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: aws.String("value"),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringPointer]()),
					infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				},
			},
			"empty": {
				Source: tfSingleStringField{
					Field1: types.StringValue(""),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: aws.String(""),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringPointer]()),
					infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				},
			},
			"null": {
				Source: tfSingleStringField{
					Field1: types.StringNull(),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringPointer]()),
					infoConverting(reflect.TypeFor[tfSingleStringField](), reflect.TypeFor[*awsSingleStringPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringField](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				},
			},
		},

		"legacy String to *string": {
			"value": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringValue("value"),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: aws.String("value"),
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringPointer]()),
					infoConverting(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				},
			},
			"empty": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringValue(""),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringPointer]()),
					infoConverting(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
					debugUsingLegacyExpander("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
				},
			},
			"null": {
				Source: tfSingleStringFieldLegacy{
					Field1: types.StringNull(),
				},
				Target: &awsSingleStringPointer{},
				WantTarget: &awsSingleStringPointer{
					Field1: nil,
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringPointer]()),
					infoConverting(reflect.TypeFor[tfSingleStringFieldLegacy](), reflect.TypeFor[*awsSingleStringPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSingleStringFieldLegacy](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
					traceExpandingNullValue("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[*string]()),
					// TODO: should log about legacy expander
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases)
		})
	}
}

func TestExpandSimpleSingleNestedBlock(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.String `tfsdk:"field1"`
		Field2 types.Int64  `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 *string
		Field2 int64
	}

	type tf02 struct {
		Field1 fwtypes.ObjectValueOf[tf01] `tfsdk:"field1"`
	}
	type aws02 struct {
		Field1 *aws01
	}
	type aws03 struct {
		Field1 aws01
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested block pointer": {
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws02{},
			WantTarget: &aws02{Field1: &aws01{Field1: aws.String("a"), Field2: 1}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws02]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[*aws02]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf02](), "Field1", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[*aws01]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[tf01](), "Field1", "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[*string]()),
				traceMatchedFieldsWithPath("Field1", "Field2", reflect.TypeFor[tf01](), "Field1", "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field2", reflect.TypeFor[types.Int64](), "Field1.Field2", reflect.TypeFor[int64]()),
			},
		},
		"single nested block nil": {
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfNull[tf01](ctx)},
			Target:     &aws02{},
			WantTarget: &aws02{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws02]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[*aws02]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf02](), "Field1", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[*aws01]()),
				traceExpandingNullValue("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[*aws01]()),
			},
		},
		"single nested block value": {
			Source:     &tf02{Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{Field1: types.StringValue("a"), Field2: types.Int64Value(1)})},
			Target:     &aws03{},
			WantTarget: &aws03{Field1: aws01{Field1: aws.String("a"), Field2: 1}},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws03]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[*aws03]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf02](), "Field1", reflect.TypeFor[*aws03]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1", reflect.TypeFor[aws01]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[tf01](), "Field1", "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[*string]()),
				traceMatchedFieldsWithPath("Field1", "Field2", reflect.TypeFor[tf01](), "Field1", "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field2", reflect.TypeFor[types.Int64](), "Field1.Field2", reflect.TypeFor[int64]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandComplexSingleNestedBlock(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Bool                        `tfsdk:"field1"`
		Field2 fwtypes.ListValueOf[types.String] `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 bool
		Field2 []string
	}

	type tf02 struct {
		Field1 fwtypes.ObjectValueOf[tf01] `tfsdk:"field1"`
	}
	type aws02 struct {
		Field1 *aws01
	}

	type tf03 struct {
		Field1 fwtypes.ObjectValueOf[tf02] `tfsdk:"field1"`
	}
	type aws03 struct {
		Field1 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested block pointer": {
			Source: &tf03{
				Field1: fwtypes.NewObjectValueOfMust[tf02](
					ctx,
					&tf02{
						Field1: fwtypes.NewObjectValueOfMust[tf01](
							ctx,
							&tf01{
								Field1: types.BoolValue(true),
								Field2: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{types.StringValue("a"), types.StringValue("b")}),
							},
						),
					},
				),
			},
			Target: &aws03{},
			WantTarget: &aws03{
				Field1: &aws02{
					Field1: &aws01{
						Field1: true,
						Field2: []string{"a", "b"},
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf03](), reflect.TypeFor[*aws03]()),
				infoConverting(reflect.TypeFor[tf03](), reflect.TypeFor[*aws03]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf03](), "Field1", reflect.TypeFor[*aws03]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf02]](), "Field1", reflect.TypeFor[*aws02]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[tf02](), "Field1", "Field1", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]](), "Field1.Field1", reflect.TypeFor[*aws01]()),
				traceMatchedFieldsWithPath("Field1.Field1", "Field1", reflect.TypeFor[tf01](), "Field1.Field1", "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field1.Field1", reflect.TypeFor[types.Bool](), "Field1.Field1.Field1", reflect.TypeFor[bool]()),
				traceMatchedFieldsWithPath("Field1.Field1", "Field2", reflect.TypeFor[tf01](), "Field1.Field1", "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1.Field1.Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]](), "Field1.Field1.Field2", reflect.TypeFor[[]string]()),
				traceExpandingWithElementsAs("Field1.Field1.Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]](), 2, "Field1.Field1.Field2", reflect.TypeFor[[]string]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandStringEnum(t *testing.T) {
	t.Parallel()

	var enum testEnum
	enumList := testEnumList

	testCases := autoFlexTestCases{
		"valid value": {
			Source:     fwtypes.StringEnumValue(testEnumList),
			Target:     &enum,
			WantTarget: &enumList,
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.StringEnum[testEnum]](), reflect.TypeFor[*testEnum]()),
				infoConverting(reflect.TypeFor[fwtypes.StringEnum[testEnum]](), reflect.TypeFor[testEnum]()),
			},
		},
		"empty value": {
			Source:     fwtypes.StringEnumNull[testEnum](),
			Target:     &enum,
			WantTarget: &enum,
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.StringEnum[testEnum]](), reflect.TypeFor[*testEnum]()),
				infoConverting(reflect.TypeFor[fwtypes.StringEnum[testEnum]](), reflect.TypeFor[testEnum]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.StringEnum[testEnum]](), "", reflect.TypeFor[testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandListOfInt64(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"valid value []int64": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int64{},
			WantTarget: &[]int64{1, -1},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]int64]()),
			},
		},
		"empty value []int64": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]int64]()),
			},
		},
		"null value []int64": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int64]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]int64]()),
			},
		},
		"valid value []*int64": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]*int64]()),
			},
		},
		"empty value []*int64": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]*int64]()),
			},
		},
		"null value []*int64": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int64]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]*int64]()),
			},
		},
		"valid value []int32": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]int32]()),
			},
		},
		"empty value []int32": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]int32]()),
			},
		},
		"null value []int32": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]int32]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]int32]()),
			},
		},
		"valid value []*int32": {
			Source: types.ListValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]*int32]()),
			},
		},
		"empty value []*int32": {
			Source:     types.ListValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]*int32]()),
			},
		},
		"null value []*int32": {
			Source:     types.ListNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]*int32]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]*int32]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSetOfInt64(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"valid value []int64": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int64{},
			WantTarget: &[]int64{1, -1},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]int64]()),
			},
		},
		"empty value []int64": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]int64]()),
			},
		},
		"null value []int64": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int64{},
			WantTarget: &[]int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int64]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]int64]()),
			},
		},
		"valid value []*int64": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{aws.Int64(1), aws.Int64(-1)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]*int64]()),
			},
		},
		"empty value []*int64": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int64]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]*int64]()),
			},
		},
		"null value []*int64": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int64{},
			WantTarget: &[]*int64{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int64]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int64]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]*int64]()),
			},
		},
		"valid value []int32": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]int32{},
			WantTarget: &[]int32{1, -1},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]int32]()),
			},
		},
		"empty value []int32": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]int32]()),
			},
		},
		"null value []int32": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]int32{},
			WantTarget: &[]int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]int32]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]int32]()),
			},
		},
		"valid value []*int32": {
			Source: types.SetValueMust(types.Int64Type, []attr.Value{
				types.Int64Value(1),
				types.Int64Value(-1),
			}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{aws.Int32(1), aws.Int32(-1)},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]*int32]()),
			},
		},
		"empty value []*int32": {
			Source:     types.SetValueMust(types.Int64Type, []attr.Value{}),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int32]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]*int32]()),
			},
		},
		"null value []*int32": {
			Source:     types.SetNull(types.Int64Type),
			Target:     &[]*int32{},
			WantTarget: &[]*int32{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]*int32]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]*int32]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]*int32]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandListOfStringEnum(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"valid value": {
			Source: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testEnumScalar)),
				types.StringValue(string(testEnumList)),
			}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{testEnumScalar, testEnumList},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 2, "", reflect.TypeFor[[]testEnum]()),
			},
		},
		"empty value": {
			Source:     types.ListValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.List](), 0, "", reflect.TypeFor[[]testEnum]()),
			},
		},
		"null value": {
			Source:     types.ListNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.List](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.List](), reflect.TypeFor[[]testEnum]()),
				traceExpandingNullValue("", reflect.TypeFor[types.List](), "", reflect.TypeFor[[]testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSetOfStringEnum(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"valid value": {
			Source: types.SetValueMust(types.StringType, []attr.Value{
				types.StringValue(string(testEnumScalar)),
				types.StringValue(string(testEnumList)),
			}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{testEnumScalar, testEnumList},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 2, "", reflect.TypeFor[[]testEnum]()),
			},
		},
		"empty value": {
			Source:     types.SetValueMust(types.StringType, []attr.Value{}),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("", reflect.TypeFor[types.Set](), 0, "", reflect.TypeFor[[]testEnum]()),
			},
		},
		"null value": {
			Source:     types.SetNull(types.StringType),
			Target:     &[]testEnum{},
			WantTarget: &[]testEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[types.Set](), reflect.TypeFor[*[]testEnum]()),
				infoConverting(reflect.TypeFor[types.Set](), reflect.TypeFor[[]testEnum]()),
				traceExpandingNullValue("", reflect.TypeFor[types.Set](), "", reflect.TypeFor[[]testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandStructListOfStringEnum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value": {
			Source: &tfListOfStringEnum{
				Field1: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{
				Field1: []testEnum{testEnumScalar, testEnumList},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfListOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConverting(reflect.TypeFor[tfListOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListOfStringEnum](), "Field1", reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]](), "Field1", reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]](), 2, "Field1", reflect.TypeFor[[]testEnum]()),
			},
		},
		"empty value": {
			Source: &tfListOfStringEnum{
				Field1: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
			Target: &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{
				Field1: []testEnum{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfListOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConverting(reflect.TypeFor[tfListOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListOfStringEnum](), "Field1", reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]](), "Field1", reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]](), 0, "Field1", reflect.TypeFor[([]testEnum)]()),
			},
		},
		"null value": {
			Source: &tfListOfStringEnum{
				Field1: fwtypes.NewListValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
			Target:     &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfListOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConverting(reflect.TypeFor[tfListOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListOfStringEnum](), "Field1", reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]](), "Field1", reflect.TypeFor[[]testEnum]()),
				traceExpandingNullValue("Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]](), "Field1", reflect.TypeFor[[]testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandStructSetOfStringEnum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value": {
			Source: &tfSetOfStringEnum{
				Field1: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
					fwtypes.StringEnumValue(testEnumScalar),
					fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{
				Field1: []testEnum{testEnumScalar, testEnumList},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSetOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConverting(reflect.TypeFor[tfSetOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetOfStringEnum](), "Field1", reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]](), "Field1", reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]](), 2, "Field1", reflect.TypeFor[[]testEnum]()),
			},
		},
		"empty value": {
			Source: &tfSetOfStringEnum{
				Field1: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
			},
			Target: &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{
				Field1: []testEnum{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSetOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConverting(reflect.TypeFor[tfSetOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetOfStringEnum](), "Field1", reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]](), "Field1", reflect.TypeFor[[]testEnum]()),
				traceExpandingWithElementsAs("Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]](), 0, "Field1", reflect.TypeFor[([]testEnum)]()),
			},
		},
		"null value": {
			Source: &tfSetOfStringEnum{
				Field1: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
			},
			Target:     &awsSliceOfStringEnum{},
			WantTarget: &awsSliceOfStringEnum{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfSetOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConverting(reflect.TypeFor[tfSetOfStringEnum](), reflect.TypeFor[*awsSliceOfStringEnum]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetOfStringEnum](), "Field1", reflect.TypeFor[*awsSliceOfStringEnum]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]](), "Field1", reflect.TypeFor[[]testEnum]()),
				traceExpandingNullValue("Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]](), "Field1", reflect.TypeFor[[]testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandTopLevelListOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value to []struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 2, "", reflect.TypeFor[[]awsSingleStringValue]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[tfSingleStringField](), "[1]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty value to []struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 0, "", reflect.TypeFor[[]awsSingleStringValue]()),
			},
		},
		"null value to []struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]awsSingleStringValue]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "", reflect.TypeFor[[]awsSingleStringValue]()),
			},
		},

		"valid value to []*struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]*awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 2, "", reflect.TypeFor[[]*awsSingleStringValue]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[tfSingleStringField](), "[1]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty value to []*struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]*awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 0, "", reflect.TypeFor[[]*awsSingleStringValue]()),
			},
		},
		"null value to []*struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]*awsSingleStringValue]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "", reflect.TypeFor[[]*awsSingleStringValue]()),
			},
		},

		"single list value to single struct": {
			Source: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
			}),
			Target: &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{
				Field1: "value1",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[awsSingleStringValue]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"empty list value to single struct": {
			Source:     fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[awsSingleStringValue]()),
			},
		},
		"null value to single struct": {
			Source:     fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[awsSingleStringValue]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "", reflect.TypeFor[awsSingleStringValue]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSetOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"valid value to []struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), 2, "", reflect.TypeFor[[]awsSingleStringValue]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[tfSingleStringField](), "[1]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty value to []struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), 0, "", reflect.TypeFor[[]awsSingleStringValue]()),
			},
		},
		"null value to []struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &[]awsSingleStringValue{},
			WantTarget: &[]awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]awsSingleStringValue]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), "", reflect.TypeFor[[]awsSingleStringValue]()),
			},
		},

		"valid value to []*struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			Target: &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]*awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), 2, "", reflect.TypeFor[[]*awsSingleStringValue]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "[0].Field1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[tfSingleStringField](), "[1]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[types.String](), "[1].Field1", reflect.TypeFor[string]()),
			},
		},
		"empty value to []*struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]*awsSingleStringValue]()),
				traceExpandingNestedObjectCollection("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), 0, "", reflect.TypeFor[[]*awsSingleStringValue]()),
			},
		},
		"null value to []*struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &[]*awsSingleStringValue{},
			WantTarget: &[]*awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*[]*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[[]*awsSingleStringValue]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), "", reflect.TypeFor[[]*awsSingleStringValue]()),
			},
		},

		"single set value to single struct": {
			Source: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
			}),
			Target: &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{
				Field1: "value1",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[awsSingleStringValue]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"empty set value to single struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[awsSingleStringValue]()),
			},
		},
		"null value to single struct": {
			Source:     fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), reflect.TypeFor[awsSingleStringValue]()),
				traceExpandingNullValue("", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), "", reflect.TypeFor[awsSingleStringValue]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandSimpleNestedBlockWithStringEnum(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Int64                  `tfsdk:"field1"`
		Field2 fwtypes.StringEnum[testEnum] `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 int64
		Field2 testEnum
	}

	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumValue(testEnumList),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: testEnumList,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[*aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[tf01](), "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]](), "Field2", reflect.TypeFor[testEnum]()),
			},
		},
		"single nested null value": {
			Source: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumNull[testEnum](),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: "",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[*aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[tf01](), "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]](), "Field2", reflect.TypeFor[testEnum]()),
				traceExpandingNullValue("Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]](), "Field2", reflect.TypeFor[testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandComplexNestedBlockWithStringEnum(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field2 fwtypes.StringEnum[testEnum] `tfsdk:"field2"`
	}
	type tf02 struct {
		Field1 types.Int64                           `tfsdk:"field1"`
		Field2 fwtypes.ListNestedObjectValueOf[tf01] `tfsdk:"field2"`
	}
	type aws02 struct {
		Field2 testEnum
	}
	type aws01 struct {
		Field1 int64
		Field2 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumValue(testEnumList),
				}),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: testEnumList,
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[*aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf02](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[tf02](), "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]](), "Field2", reflect.TypeFor[*aws02]()),
				traceMatchedFieldsWithPath("Field2[0]", "Field2", reflect.TypeFor[tf01](), "Field2", "Field2", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field2[0].Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]](), "Field2.Field2", reflect.TypeFor[testEnum]()),
			},
		},
		"single nested null value": {
			Source: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumNull[testEnum](),
				}),
			},
			Target: &aws01{},
			WantTarget: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: "",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf02](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf02](), reflect.TypeFor[*aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf02](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Int64](), "Field1", reflect.TypeFor[int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[tf02](), "Field2", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]](), "Field2", reflect.TypeFor[*aws02]()),
				traceMatchedFieldsWithPath("Field2[0]", "Field2", reflect.TypeFor[tf01](), "Field2", "Field2", reflect.TypeFor[*aws02]()),
				infoConvertingWithPath("Field2[0].Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]](), "Field2.Field2", reflect.TypeFor[testEnum]()),
				traceExpandingNullValue("Field2[0].Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]](), "Field2.Field2", reflect.TypeFor[testEnum]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandListOfNestedObjectField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"ListNestedObject to *struct": {
			"value": {
				Source: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
				Target: &awsNestedObjectPointer{},
				WantTarget: &awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfListOfNestedObject](), reflect.TypeFor[*awsNestedObjectPointer]()),
					infoConverting(reflect.TypeFor[tfListOfNestedObject](), reflect.TypeFor[*awsNestedObjectPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfListOfNestedObject](), "Field1", reflect.TypeFor[*awsNestedObjectPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[string]()),
				},
			},
		},

		"ListNestedObject to []struct": {
			"empty": {
				Source: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
				Target: &awsSliceOfNestedObjectValues{},
				WantTarget: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfListOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					infoConverting(reflect.TypeFor[tfListOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfListOfNestedObject](), "Field1", reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[[]awsSingleStringValue]()),
					traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 0, "Field1", reflect.TypeFor[[]awsSingleStringValue]()),
				},
			},
			"values": {
				Source: &tfListOfNestedObject{Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
					{Field1: types.StringValue("a")},
					{Field1: types.StringValue("b")},
				})},
				Target: &awsSliceOfNestedObjectValues{},
				WantTarget: &awsSliceOfNestedObjectValues{Field1: []awsSingleStringValue{
					{Field1: "a"},
					{Field1: "b"},
				}},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfListOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					infoConverting(reflect.TypeFor[tfListOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfListOfNestedObject](), "Field1", reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[[]awsSingleStringValue]()),
					traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 2, "Field1", reflect.TypeFor[[]awsSingleStringValue]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1[1]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
				},
			},
		},

		"ListNestedObject to []*struct": {
			"empty": {
				Source: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
				Target: &awsSliceOfNestedObjectPointers{},
				WantTarget: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfListOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					infoConverting(reflect.TypeFor[tfListOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfListOfNestedObject](), "Field1", reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[[]*awsSingleStringValue]()),
					traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 0, "Field1", reflect.TypeFor[[]*awsSingleStringValue]()),
				},
			},
			"values": {
				Source: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				Target: &awsSliceOfNestedObjectPointers{},
				WantTarget: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfListOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					infoConverting(reflect.TypeFor[tfListOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfListOfNestedObject](), "Field1", reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[[]*awsSingleStringValue]()),
					traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), 2, "Field1", reflect.TypeFor[[]*awsSingleStringValue]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1[1]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases)
		})
	}
}

func TestExpandSetOfNestedObjectField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"SetNestedObject to *struct": {
			"value": {
				Source: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
				Target: &awsNestedObjectPointer{},
				WantTarget: &awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfSetOfNestedObject](), reflect.TypeFor[*awsNestedObjectPointer]()),
					infoConverting(reflect.TypeFor[tfSetOfNestedObject](), reflect.TypeFor[*awsNestedObjectPointer]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSetOfNestedObject](), "Field1", reflect.TypeFor[*awsNestedObjectPointer]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1.Field1", reflect.TypeFor[string]()),
				},
			},
		},

		"SetNestedObject to []*struct": {
			"empty": {
				Source:     &tfSetOfNestedObject{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{})},
				Target:     &awsSliceOfNestedObjectPointers{},
				WantTarget: &awsSliceOfNestedObjectPointers{Field1: []*awsSingleStringValue{}},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfSetOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					infoConverting(reflect.TypeFor[tfSetOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSetOfNestedObject](), "Field1", reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[[]*awsSingleStringValue]()),
					traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), 0, "Field1", reflect.TypeFor[[]*awsSingleStringValue]()),
				},
			},
			"values": {
				Source: &tfSetOfNestedObject{Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
					{Field1: types.StringValue("a")},
					{Field1: types.StringValue("b")},
				})},
				Target: &awsSliceOfNestedObjectPointers{},
				WantTarget: &awsSliceOfNestedObjectPointers{Field1: []*awsSingleStringValue{
					{Field1: "a"},
					{Field1: "b"},
				}},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfSetOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					infoConverting(reflect.TypeFor[tfSetOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSetOfNestedObject](), "Field1", reflect.TypeFor[*awsSliceOfNestedObjectPointers]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[[]*awsSingleStringValue]()),
					traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), 2, "Field1", reflect.TypeFor[[]*awsSingleStringValue]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1[1]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
				},
			},
		},

		"SetNestedObject to []struct": {
			"values": {
				Source: &tfSetOfNestedObject{Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
					{Field1: types.StringValue("a")},
					{Field1: types.StringValue("b")},
				})},
				Target: &awsSliceOfNestedObjectValues{},
				WantTarget: &awsSliceOfNestedObjectValues{Field1: []awsSingleStringValue{
					{Field1: "a"},
					{Field1: "b"},
				}},
				expectedLogLines: []map[string]any{
					infoExpanding(reflect.TypeFor[*tfSetOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					infoConverting(reflect.TypeFor[tfSetOfNestedObject](), reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					traceMatchedFields("Field1", reflect.TypeFor[tfSetOfNestedObject](), "Field1", reflect.TypeFor[*awsSliceOfNestedObjectValues]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[[]awsSingleStringValue]()),
					traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]](), 2, "Field1", reflect.TypeFor[[]awsSingleStringValue]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1[0]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[types.String](), "Field1[0].Field1", reflect.TypeFor[string]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[tfSingleStringField](), "Field1[1]", "Field1", reflect.TypeFor[*awsSingleStringValue]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[types.String](), "Field1[1].Field1", reflect.TypeFor[string]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoExpandTestCases(t, cases)
		})
	}
}
func TestExpandMapBlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"nil map block key": {
			Source: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfNull[tfMapBlockElement](ctx),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: nil,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapBlockList](), reflect.TypeFor[*awsMapBlockValues]()),
				infoConverting(reflect.TypeFor[tfMapBlockList](), reflect.TypeFor[*awsMapBlockValues]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[tfMapBlockList](), "MapBlock", reflect.TypeFor[*awsMapBlockValues]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]](), "MapBlock", reflect.TypeFor[map[string]awsMapBlockElement]()),
				traceExpandingNullValue("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]](), "MapBlock", reflect.TypeFor[map[string]awsMapBlockElement]()),
			},
		},
		"map block key list": {
			Source: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[tfMapBlockElement](ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: types.StringValue("y"),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
					"y": {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapBlockList](), reflect.TypeFor[*awsMapBlockValues]()),
				infoConverting(reflect.TypeFor[tfMapBlockList](), reflect.TypeFor[*awsMapBlockValues]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[tfMapBlockList](), "MapBlock", reflect.TypeFor[*awsMapBlockValues]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]](), "MapBlock", reflect.TypeFor[map[string]awsMapBlockElement]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr2", reflect.TypeFor[string]()),
			},
		},
		"map block key set": {
			Source: &tfMapBlockSet{
				MapBlock: fwtypes.NewSetNestedObjectValueOfValueSliceMust[tfMapBlockElement](ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: types.StringValue("y"),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
					"y": {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapBlockSet](), reflect.TypeFor[*awsMapBlockValues]()),
				infoConverting(reflect.TypeFor[tfMapBlockSet](), reflect.TypeFor[*awsMapBlockValues]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[tfMapBlockSet](), "MapBlock", reflect.TypeFor[*awsMapBlockValues]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfMapBlockElement]](), "MapBlock", reflect.TypeFor[map[string]awsMapBlockElement]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr2", reflect.TypeFor[string]()),
			},
		},
		"map block key ptr source": {
			Source: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: types.StringValue("y"),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
					"y": {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapBlockList](), reflect.TypeFor[*awsMapBlockValues]()),
				infoConverting(reflect.TypeFor[tfMapBlockList](), reflect.TypeFor[*awsMapBlockValues]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[tfMapBlockList](), "MapBlock", reflect.TypeFor[*awsMapBlockValues]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]](), "MapBlock", reflect.TypeFor[map[string]awsMapBlockElement]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr2", reflect.TypeFor[string]()),
			},
		},
		"map block key ptr both": {
			Source: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: types.StringValue("y"),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockPointers{},
			WantTarget: &awsMapBlockPointers{
				MapBlock: map[string]*awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
					"y": {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapBlockList](), reflect.TypeFor[*awsMapBlockPointers]()),
				infoConverting(reflect.TypeFor[tfMapBlockList](), reflect.TypeFor[*awsMapBlockPointers]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[tfMapBlockList](), "MapBlock", reflect.TypeFor[*awsMapBlockPointers]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]](), "MapBlock", reflect.TypeFor[map[string]*awsMapBlockElement]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"x\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"x\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[tfMapBlockElement](), "MapBlock[\"y\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"y\"].Attr2", reflect.TypeFor[string]()),
			},
		},
		"map block enum key": {
			Source: &tfMapBlockListEnumKey{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[tfMapBlockElementEnumKey](ctx, []tfMapBlockElementEnumKey{
					{
						MapBlockKey: fwtypes.StringEnumValue(testEnumList),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
					{
						MapBlockKey: fwtypes.StringEnumValue(testEnumScalar),
						Attr1:       types.StringValue("c"),
						Attr2:       types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			WantTarget: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					string(testEnumList): {
						Attr1: "a",
						Attr2: "b",
					},
					string(testEnumScalar): {
						Attr1: "c",
						Attr2: "d",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapBlockListEnumKey](), reflect.TypeFor[*awsMapBlockValues]()),
				infoConverting(reflect.TypeFor[tfMapBlockListEnumKey](), reflect.TypeFor[*awsMapBlockValues]()),

				traceMatchedFields("MapBlock", reflect.TypeFor[tfMapBlockListEnumKey](), "MapBlock", reflect.TypeFor[*awsMapBlockValues]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElementEnumKey]](), "MapBlock", reflect.TypeFor[map[string]awsMapBlockElement]()),

				traceSkipMapBlockKey("MapBlock[0]", reflect.TypeFor[tfMapBlockElementEnumKey](), "MapBlock[\"List\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr1", reflect.TypeFor[tfMapBlockElementEnumKey](), "MapBlock[\"List\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"List\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[0]", "Attr2", reflect.TypeFor[tfMapBlockElementEnumKey](), "MapBlock[\"List\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[0].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"List\"].Attr2", reflect.TypeFor[string]()),

				traceSkipMapBlockKey("MapBlock[1]", reflect.TypeFor[tfMapBlockElementEnumKey](), "MapBlock[\"Scalar\"]", reflect.TypeFor[*awsMapBlockElement]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr1", reflect.TypeFor[tfMapBlockElementEnumKey](), "MapBlock[\"Scalar\"]", "Attr1", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr1", reflect.TypeFor[types.String](), "MapBlock[\"Scalar\"].Attr1", reflect.TypeFor[string]()),
				traceMatchedFieldsWithPath("MapBlock[1]", "Attr2", reflect.TypeFor[tfMapBlockElementEnumKey](), "MapBlock[\"Scalar\"]", "Attr2", reflect.TypeFor[*awsMapBlockElement]()),
				infoConvertingWithPath("MapBlock[1].Attr2", reflect.TypeFor[types.String](), "MapBlock[\"Scalar\"].Attr2", reflect.TypeFor[string]()),
			},
		},

		"map block list no key": {
			Source: &tfMapBlockListNoKey{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[tfMapBlockElementNoKey](ctx, []tfMapBlockElementNoKey{
					{
						Attr1: types.StringValue("a"),
						Attr2: types.StringValue("b"),
					},
					{
						Attr1: types.StringValue("c"),
						Attr2: types.StringValue("d"),
					},
				}),
			},
			Target: &awsMapBlockValues{},
			expectedDiags: diag.Diagnostics{
				diagExpandingNoMapBlockKey(reflect.TypeFor[tfMapBlockElementNoKey]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfMapBlockListNoKey](), reflect.TypeFor[*awsMapBlockValues]()),
				infoConverting(reflect.TypeFor[tfMapBlockListNoKey](), reflect.TypeFor[*awsMapBlockValues]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[tfMapBlockListNoKey](), "MapBlock", reflect.TypeFor[*awsMapBlockValues]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElementNoKey]](), "MapBlock", reflect.TypeFor[map[string]awsMapBlockElement]()),
				errorSourceHasNoMapBlockKey("MapBlock[0]", reflect.TypeFor[tfMapBlockElementNoKey](), "MapBlock", reflect.TypeFor[awsMapBlockElement]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[*aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				traceExpandingNullValue("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				traceSkipIgnoredSourceField(reflect.TypeFor[tf01](), "Tags", reflect.TypeFor[*aws01]()),
			},
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[*aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				traceSkipIgnoredSourceField(reflect.TypeFor[tf01](), "Tags", reflect.TypeFor[*aws01]()),
			},
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[*aws01]()),
				traceMatchedFields("Field1", reflect.TypeFor[tf01](), "Field1", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.Bool](), "Field1", reflect.TypeFor[bool]()),
				traceMatchedFields("Tags", reflect.TypeFor[tf01](), "Tags", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "Tags", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "Tags", reflect.TypeFor[map[string]string]()),
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
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tf01](), reflect.TypeFor[*aws01]()),
				infoConverting(reflect.TypeFor[tf01](), reflect.TypeFor[*aws01]()),
				traceSkipIgnoredSourceField(reflect.TypeFor[tf01](), "Field1", reflect.TypeFor[*aws01]()),
				traceMatchedFields("Tags", reflect.TypeFor[tf01](), "Tags", reflect.TypeFor[*aws01]()),
				infoConvertingWithPath("Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), "Tags", reflect.TypeFor[map[string]string]()),
				traceExpandingWithElementsAs("Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]](), 1, "Tags", reflect.TypeFor[map[string]string]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandIgnoreStructTag(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"to value": {
			Source: tfSingleStringFieldIgnore{
				Field1: types.StringValue("value1"),
			},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSingleStringFieldIgnore](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[tfSingleStringFieldIgnore](), reflect.TypeFor[*awsSingleStringValue]()),
				traceSkipIgnoredSourceField(reflect.TypeFor[tfSingleStringFieldIgnore](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
			},
		},
		"to pointer": {
			Source: tfSingleStringFieldIgnore{
				Field1: types.StringValue("value1"),
			},
			Target:     &awsSingleStringPointer{},
			WantTarget: &awsSingleStringPointer{},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSingleStringFieldIgnore](), reflect.TypeFor[*awsSingleStringPointer]()),
				infoConverting(reflect.TypeFor[tfSingleStringFieldIgnore](), reflect.TypeFor[*awsSingleStringPointer]()),
				traceSkipIgnoredSourceField(reflect.TypeFor[tfSingleStringFieldIgnore](), "Field1", reflect.TypeFor[*awsSingleStringPointer]()),
			},
		},
	}

	runAutoExpandTestCases(t, testCases)
}

func TestExpandInterface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var targetInterface awsInterfaceInterface

	testCases := autoFlexTestCases{
		"top level": {
			Source: tfInterfaceFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			WantTarget: testFlexAWSInterfaceInterfacePtr(&awsInterfaceInterfaceImpl{
				AWSField: "value1",
			}),
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfInterfaceFlexer](), reflect.TypeFor[*awsInterfaceInterface]()),
				infoConverting(reflect.TypeFor[tfInterfaceFlexer](), reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[tfInterfaceFlexer](), "", reflect.TypeFor[awsInterfaceInterface]()),
			},
		},
		"top level return value does not implement target interface": {
			Source: tfInterfaceIncompatibleExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			expectedDiags: diag.Diagnostics{
				diagExpandedTypeDoesNotImplement(reflect.TypeFor[*awsInterfaceIncompatibleImpl](), reflect.TypeFor[awsInterfaceInterface]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfInterfaceIncompatibleExpander](), reflect.TypeFor[*awsInterfaceInterface]()),
				infoConverting(reflect.TypeFor[tfInterfaceIncompatibleExpander](), reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[tfInterfaceIncompatibleExpander](), "", reflect.TypeFor[awsInterfaceInterface]()),
			},
		},
		"single list Source and single interface Target": {
			Source: tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSingle]()),
				infoConverting(reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[*awsInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfInterfaceFlexer](), "Field1", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
		"single list non-Expander Source and single interface Target": {
			Source: tfListNestedObject[tfSingleStringField]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: nil,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfListNestedObject[tfSingleStringField]](), reflect.TypeFor[*awsInterfaceSingle]()),
				infoConverting(reflect.TypeFor[tfListNestedObject[tfSingleStringField]](), reflect.TypeFor[*awsInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListNestedObject[tfSingleStringField]](), "Field1", reflect.TypeFor[*awsInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[awsInterfaceInterface]()),
				{
					"@level":             "error",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Interface),
					logAttrKeySourcePath: "Field1[0]",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[tfSingleStringField]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[*awsInterfaceInterface]()),
				},
			},
		},
		"single set Source and single interface Target": {
			Source: tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSingle]()),
				infoConverting(reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[*awsInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfInterfaceFlexer](), "Field1", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
		"empty list Source and empty interface Target": {
			Source: tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSlice]()),
				infoConverting(reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[*awsInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]](), 0, "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
			},
		},
		"non-empty list Source and non-empty interface Target": {
			Source: tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSlice]()),
				infoConverting(reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListNestedObject[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[*awsInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]](), 2, "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfInterfaceFlexer](), "Field1[0]", reflect.TypeFor[*awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[tfInterfaceFlexer](), "Field1[1]", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
		"empty set Source and empty interface Target": {
			Source: tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSlice]()),
				infoConverting(reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[*awsInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]](), 0, "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
			},
		},
		"non-empty set Source and non-empty interface Target": {
			Source: tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSlice]()),
				infoConverting(reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetNestedObject[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[*awsInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]](), 2, "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfInterfaceFlexer](), "Field1[0]", reflect.TypeFor[*awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[tfInterfaceFlexer](), "Field1[1]", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
		"object value Source and struct Target": {
			Source: tfObjectValue[tfInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfInterfaceFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfObjectValue[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSingle]()),
				infoConverting(reflect.TypeFor[tfObjectValue[tfInterfaceFlexer]](), reflect.TypeFor[*awsInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfObjectValue[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[*awsInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfInterfaceFlexer]](), "Field1", reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexExpander("Field1", reflect.TypeFor[tfInterfaceFlexer](), "Field1", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func testFlexAWSInterfaceInterfacePtr(v awsInterfaceInterface) *awsInterfaceInterface { // nosemgrep:ci.aws-in-func-name
	return &v
}

func TestExpandExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"top level struct Target": {
			Source: tfFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &awsExpander{},
			WantTarget: &awsExpander{
				AWSField: "value1",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfFlexer](), reflect.TypeFor[*awsExpander]()),
				infoConverting(reflect.TypeFor[tfFlexer](), reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[tfFlexer](), "", reflect.TypeFor[*awsExpander]()),
			},
		},
		"top level string Target": {
			Source: tfExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target:     aws.String(""),
			WantTarget: aws.String("value1"),
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderToString](), reflect.TypeFor[*string]()),
				infoConverting(reflect.TypeFor[tfExpanderToString](), reflect.TypeFor[string]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[tfExpanderToString](), "", reflect.TypeFor[string]()),
			},
		},
		"top level incompatible struct Target": {
			Source: tfFlexer{
				Field1: types.StringValue("value1"),
			},
			Target: &awsExpanderIncompatible{},
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[awsExpander](), reflect.TypeFor[awsExpanderIncompatible]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfFlexer](), reflect.TypeFor[*awsExpanderIncompatible]()),
				infoConverting(reflect.TypeFor[tfFlexer](), reflect.TypeFor[*awsExpanderIncompatible]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[tfFlexer](), "", reflect.TypeFor[*awsExpanderIncompatible]()),
			},
		},
		"top level expands to nil": {
			Source: tfExpanderToNil{
				Field1: types.StringValue("value1"),
			},
			Target: &awsExpander{},
			expectedDiags: diag.Diagnostics{
				diagExpandsToNil(reflect.TypeFor[tfExpanderToNil]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderToNil](), reflect.TypeFor[*awsExpander]()),
				infoConverting(reflect.TypeFor[tfExpanderToNil](), reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[tfExpanderToNil](), "", reflect.TypeFor[*awsExpander]()),
			},
		},
		"top level incompatible non-struct Target": {
			Source: tfExpanderToString{
				Field1: types.StringValue("value1"),
			},
			Target: aws.Int64(0),
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[string](), reflect.TypeFor[int64]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderToString](), reflect.TypeFor[*int64]()),
				infoConverting(reflect.TypeFor[tfExpanderToString](), reflect.TypeFor[int64]()),
				infoSourceImplementsFlexExpander("", reflect.TypeFor[tfExpanderToString](), "", reflect.TypeFor[int64]()),
			},
		},
		"single list Source and single struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfFlexer](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"single set Source and single struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfFlexer](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"single list Source and single *struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfFlexer](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"single set Source and single *struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfFlexer](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"empty list Source and empty struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[[]awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), 0, "Field1", reflect.TypeFor[[]awsExpander]()),
			},
		},
		"non-empty list Source and non-empty struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[[]awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), 2, "Field1", reflect.TypeFor[[]awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfFlexer](), "Field1[0]", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[tfFlexer](), "Field1[1]", reflect.TypeFor[*awsExpander]()),
			},
		},
		"empty list Source and empty *struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[[]*awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), 0, "Field1", reflect.TypeFor[[]*awsExpander]()),
			},
		},
		"non-empty list Source and non-empty *struct Target": {
			Source: tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[tfExpanderListNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[[]*awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]](), 2, "Field1", reflect.TypeFor[[]*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfFlexer](), "Field1[0]", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[tfFlexer](), "Field1[1]", reflect.TypeFor[*awsExpander]()),
			},
		},
		"empty set Source and empty struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[[]awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), 0, "Field1", reflect.TypeFor[[]awsExpander]()),
			},
		},
		"non-empty set Source and non-empty struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[[]awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), 2, "Field1", reflect.TypeFor[[]awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfFlexer](), "Field1[0]", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[tfFlexer](), "Field1[1]", reflect.TypeFor[*awsExpander]()),
			},
		},
		"empty set Source and empty *struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[[]*awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), 0, "Field1", reflect.TypeFor[[]*awsExpander]()),
			},
		},
		"non-empty set Source and non-empty *struct Target": {
			Source: tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[tfExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[[]*awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]](), 2, "Field1", reflect.TypeFor[[]*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[0]", reflect.TypeFor[tfFlexer](), "Field1[0]", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1[1]", reflect.TypeFor[tfFlexer](), "Field1[1]", reflect.TypeFor[*awsExpander]()),
			},
		},
		"object value Source and struct Target": {
			Source: tfExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderObjectValue](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[tfExpanderObjectValue](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderObjectValue](), "Field1", reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[awsExpander]()),
				infoSourceImplementsFlexExpander("Field1", reflect.TypeFor[tfFlexer](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"object value Source and *struct Target": {
			Source: tfExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfExpanderObjectValue](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[tfExpanderObjectValue](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExpanderObjectValue](), "Field1", reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfFlexer]](), "Field1", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexExpander("Field1", reflect.TypeFor[tfFlexer](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandInterfaceTypedExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var targetInterface awsInterfaceInterface

	testCases := autoFlexTestCases{
		"top level": {
			Source: tfInterfaceTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			WantTarget: testFlexAWSInterfaceInterfacePtr(&awsInterfaceInterfaceImpl{
				AWSField: "value1",
			}),
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfInterfaceTypedExpander](), reflect.TypeFor[*awsInterfaceInterface]()),
				infoConverting(reflect.TypeFor[tfInterfaceTypedExpander](), reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[tfInterfaceTypedExpander](), "", reflect.TypeFor[awsInterfaceInterface]()),
			},
		},
		"top level return value does not implement target interface": {
			Source: tfInterfaceIncompatibleTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &targetInterface,
			expectedDiags: diag.Diagnostics{
				diagExpandedTypeDoesNotImplement(reflect.TypeFor[*awsInterfaceIncompatibleImpl](), reflect.TypeFor[awsInterfaceInterface]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfInterfaceIncompatibleTypedExpander](), reflect.TypeFor[*awsInterfaceInterface]()),
				infoConverting(reflect.TypeFor[tfInterfaceIncompatibleTypedExpander](), reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[tfInterfaceIncompatibleTypedExpander](), "", reflect.TypeFor[awsInterfaceInterface]()),
			},
		},
		"single list Source and single interface Target": {
			Source: tfListNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSingle]()),
				infoConverting(reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*awsInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfInterfaceTypedExpander](), "Field1", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
		"single list non-Expander Source and single interface Target": {
			Source: tfListNestedObject[tfSingleStringField]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: nil,
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfListNestedObject[tfSingleStringField]](), reflect.TypeFor[*awsInterfaceSingle]()),
				infoConverting(reflect.TypeFor[tfListNestedObject[tfSingleStringField]](), reflect.TypeFor[*awsInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListNestedObject[tfSingleStringField]](), "Field1", reflect.TypeFor[*awsInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]](), "Field1", reflect.TypeFor[awsInterfaceInterface]()),
				{
					"@level":             "error",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Expand; incompatible types",
					"from":               map[string]any{},
					"to":                 float64(reflect.Interface),
					logAttrKeySourcePath: "Field1[0]",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[tfSingleStringField]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[*awsInterfaceInterface]()),
				},
			},
		},
		"single set Source and single interface Target": {
			Source: tfSetNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSingle]()),
				infoConverting(reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*awsInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfInterfaceTypedExpander](), "Field1", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
		"empty list Source and empty interface Target": {
			Source: tfListNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSlice]()),
				infoConverting(reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*awsInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceTypedExpander]](), 0, "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
			},
		},
		"non-empty list Source and non-empty interface Target": {
			Source: tfListNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSlice]()),
				infoConverting(reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfListNestedObject[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*awsInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceTypedExpander]](), 2, "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfInterfaceTypedExpander](), "Field1[0]", reflect.TypeFor[*awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[tfInterfaceTypedExpander](), "Field1[1]", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
		"empty set Source and empty interface Target": {
			Source: tfSetNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSlice]()),
				infoConverting(reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*awsInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceTypedExpander]](), 0, "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
			},
		},
		"non-empty set Source and non-empty interface Target": {
			Source: tfSetNestedObject[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsInterfaceSlice{},
			WantTarget: &awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSlice]()),
				infoConverting(reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetNestedObject[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*awsInterfaceSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceTypedExpander]](), 2, "Field1", reflect.TypeFor[[]awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfInterfaceTypedExpander](), "Field1[0]", reflect.TypeFor[*awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[tfInterfaceTypedExpander](), "Field1[1]", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
		"object value Source and struct Target": {
			Source: tfObjectValue[tfInterfaceTypedExpander]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfInterfaceTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsInterfaceSingle{},
			WantTarget: &awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfObjectValue[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSingle]()),
				infoConverting(reflect.TypeFor[tfObjectValue[tfInterfaceTypedExpander]](), reflect.TypeFor[*awsInterfaceSingle]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfObjectValue[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[*awsInterfaceSingle]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfInterfaceTypedExpander]](), "Field1", reflect.TypeFor[awsInterfaceInterface]()),
				infoSourceImplementsFlexTypedExpander("Field1", reflect.TypeFor[tfInterfaceTypedExpander](), "Field1", reflect.TypeFor[*awsInterfaceInterface]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

func TestExpandTypedExpander(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"top level struct Target": {
			Source: tfTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &awsExpander{},
			WantTarget: &awsExpander{
				AWSField: "value1",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpander](), reflect.TypeFor[*awsExpander]()),
				infoConverting(reflect.TypeFor[tfTypedExpander](), reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[tfTypedExpander](), "", reflect.TypeFor[*awsExpander]()),
			},
		},
		"top level incompatible struct Target": {
			Source: tfTypedExpander{
				Field1: types.StringValue("value1"),
			},
			Target: &awsExpanderIncompatible{},
			expectedDiags: diag.Diagnostics{
				diagCannotBeAssigned(reflect.TypeFor[awsExpander](), reflect.TypeFor[awsExpanderIncompatible]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpander](), reflect.TypeFor[*awsExpanderIncompatible]()),
				infoConverting(reflect.TypeFor[tfTypedExpander](), reflect.TypeFor[*awsExpanderIncompatible]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[tfTypedExpander](), "", reflect.TypeFor[*awsExpanderIncompatible]()),
			},
		},
		"top level expands to nil": {
			Source: tfTypedExpanderToNil{
				Field1: types.StringValue("value1"),
			},
			Target: &awsExpander{},
			expectedDiags: diag.Diagnostics{
				diagExpandsToNil(reflect.TypeFor[tfTypedExpanderToNil]()),
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderToNil](), reflect.TypeFor[*awsExpander]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderToNil](), reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("", reflect.TypeFor[tfTypedExpanderToNil](), "", reflect.TypeFor[*awsExpander]()),
			},
		},
		"single list Source and single struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfTypedExpander](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"single set Source and single struct Target": {
			Source: tfSetNestedObject[tfTypedExpander]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfSetNestedObject[tfTypedExpander]](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[tfSetNestedObject[tfTypedExpander]](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfSetNestedObject[tfTypedExpander]](), "Field1", reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfTypedExpander](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"single list Source and single *struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfTypedExpander](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"single set Source and single *struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfTypedExpander](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"empty list Source and empty struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[[]awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), 0, "Field1", reflect.TypeFor[[]awsExpander]()),
			},
		},
		"non-empty list Source and non-empty struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[[]awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), 2, "Field1", reflect.TypeFor[[]awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfTypedExpander](), "Field1[0]", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[tfTypedExpander](), "Field1[1]", reflect.TypeFor[*awsExpander]()),
			},
		},
		"empty list Source and empty *struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[[]*awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), 0, "Field1", reflect.TypeFor[[]*awsExpander]()),
			},
		},
		"non-empty list Source and non-empty *struct Target": {
			Source: tfTypedExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderListNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderListNestedObject](), "Field1", reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[[]*awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfTypedExpander]](), 2, "Field1", reflect.TypeFor[[]*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfTypedExpander](), "Field1[0]", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[tfTypedExpander](), "Field1[1]", reflect.TypeFor[*awsExpander]()),
			},
		},
		"empty set Source and empty struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[[]awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), 0, "Field1", reflect.TypeFor[[]awsExpander]()),
			},
		},
		"non-empty set Source and non-empty struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderStructSlice{},
			WantTarget: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderStructSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderStructSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[[]awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), 2, "Field1", reflect.TypeFor[[]awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfTypedExpander](), "Field1[0]", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[tfTypedExpander](), "Field1[1]", reflect.TypeFor[*awsExpander]()),
			},
		},
		"empty set Source and empty *struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[[]*awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), 0, "Field1", reflect.TypeFor[[]*awsExpander]()),
			},
		},
		"non-empty set Source and non-empty *struct Target": {
			Source: tfTypedExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfTypedExpander{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			Target: &awsExpanderPtrSlice{},
			WantTarget: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderSetNestedObject](), reflect.TypeFor[*awsExpanderPtrSlice]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderSetNestedObject](), "Field1", reflect.TypeFor[*awsExpanderPtrSlice]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[[]*awsExpander]()),
				traceExpandingNestedObjectCollection("Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfTypedExpander]](), 2, "Field1", reflect.TypeFor[[]*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[0]", reflect.TypeFor[tfTypedExpander](), "Field1[0]", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1[1]", reflect.TypeFor[tfTypedExpander](), "Field1[1]", reflect.TypeFor[*awsExpander]()),
			},
		},
		"object value Source and struct Target": {
			Source: tfTypedExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsExpanderSingleStruct{},
			WantTarget: &awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderObjectValue](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderObjectValue](), reflect.TypeFor[*awsExpanderSingleStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderObjectValue](), "Field1", reflect.TypeFor[*awsExpanderSingleStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1", reflect.TypeFor[tfTypedExpander](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
		"object value Source and *struct Target": {
			Source: tfTypedExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfTypedExpander{
					Field1: types.StringValue("value1"),
				}),
			},
			Target: &awsExpanderSinglePtr{},
			WantTarget: &awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[tfTypedExpanderObjectValue](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConverting(reflect.TypeFor[tfTypedExpanderObjectValue](), reflect.TypeFor[*awsExpanderSinglePtr]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfTypedExpanderObjectValue](), "Field1", reflect.TypeFor[*awsExpanderSinglePtr]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfTypedExpander]](), "Field1", reflect.TypeFor[*awsExpander]()),
				infoSourceImplementsFlexTypedExpander("Field1", reflect.TypeFor[tfTypedExpander](), "Field1", reflect.TypeFor[*awsExpander]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

type TFExportedStruct struct {
	Field1 types.String `tfsdk:"field1"`
}

type tfExportedEmbeddedStruct struct {
	TFExportedStruct
	Field2 types.String `tfsdk:"field2"`
}

type tfUnexportedEmbeddedStruct struct {
	tfSingleStringField
	Field2 types.String `tfsdk:"field2"`
}

type awsEmbeddedStruct struct {
	Field1 string
	Field2 string
}

func TestExpandEmbeddedStruct(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"exported": {
			Source: &tfExportedEmbeddedStruct{
				TFExportedStruct: TFExportedStruct{
					Field1: types.StringValue("a"),
				},
				Field2: types.StringValue("b"),
			},
			Target: &awsEmbeddedStruct{},
			WantTarget: &awsEmbeddedStruct{
				Field1: "a",
				Field2: "b",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfExportedEmbeddedStruct](), reflect.TypeFor[*awsEmbeddedStruct]()),
				infoConverting(reflect.TypeFor[tfExportedEmbeddedStruct](), reflect.TypeFor[*awsEmbeddedStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfExportedEmbeddedStruct](), "Field1", reflect.TypeFor[*awsEmbeddedStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				traceMatchedFields("Field2", reflect.TypeFor[tfExportedEmbeddedStruct](), "Field2", reflect.TypeFor[*awsEmbeddedStruct]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[types.String](), "Field2", reflect.TypeFor[string]()),
			},
		},
		"unexported": {
			Source: &tfUnexportedEmbeddedStruct{
				tfSingleStringField: tfSingleStringField{
					Field1: types.StringValue("a"),
				},
				Field2: types.StringValue("b"),
			},
			Target: &awsEmbeddedStruct{},
			WantTarget: &awsEmbeddedStruct{
				Field1: "a",
				Field2: "b",
			},
			expectedLogLines: []map[string]any{
				infoExpanding(reflect.TypeFor[*tfUnexportedEmbeddedStruct](), reflect.TypeFor[*awsEmbeddedStruct]()),
				infoConverting(reflect.TypeFor[tfUnexportedEmbeddedStruct](), reflect.TypeFor[*awsEmbeddedStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[tfUnexportedEmbeddedStruct](), "Field1", reflect.TypeFor[*awsEmbeddedStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[types.String](), "Field1", reflect.TypeFor[string]()),
				traceMatchedFields("Field2", reflect.TypeFor[tfUnexportedEmbeddedStruct](), "Field2", reflect.TypeFor[*awsEmbeddedStruct]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[types.String](), "Field2", reflect.TypeFor[string]()),
			},
		},
	}
	runAutoExpandTestCases(t, testCases)
}

type autoFlexTestCase struct {
	Options          []AutoFlexOptionsFunc
	Source           any
	Target           any
	expectedDiags    diag.Diagnostics
	expectedLogLines []map[string]any
	WantTarget       any
	WantDiff         bool
}

type autoFlexTestCases map[string]autoFlexTestCase

func runAutoExpandTestCases(t *testing.T, testCases autoFlexTestCases) {
	t.Helper()

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var buf bytes.Buffer
			ctx = tflogtest.RootLogger(ctx, &buf)

			ctx = registerTestingLogger(ctx)

			diags := Expand(ctx, testCase.Source, testCase.Target, testCase.Options...)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			lines, err := tflogtest.MultilineJSONDecode(&buf)
			if err != nil {
				t.Fatalf("Expand: decoding log lines: %s", err)
			}
			if diff := cmp.Diff(lines, testCase.expectedLogLines); diff != "" {
				t.Errorf("unexpected log lines diff (+wanted, -got): %s", diff)
			}

			if !diags.HasError() {
				if diff := cmp.Diff(testCase.Target, testCase.WantTarget); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}
