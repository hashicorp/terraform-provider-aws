// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestFlatten(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var (
		typedNilSource *emptyStruct
		typedNilTarget *emptyStruct
	)

	testString := "test"

	testARN := "arn:aws:securityhub:us-west-2:1234567890:control/cis-aws-foundations-benchmark/v/1.2.0/1.1" //lintignore:AWSAT003,AWSAT005

	testTimeStr := "2013-09-25T09:34:01Z"
	testTimeTime := errs.Must(time.Parse(time.RFC3339, testTimeStr))
	var zeroTime time.Time

	testCases := autoFlexTestCases{
		"nil Source": {
			Target: &emptyStruct{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningSourceIsNil(nil),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(nil, reflect.TypeFor[*emptyStruct]()),
				errorSourceIsNil("", nil, "", reflect.TypeFor[*emptyStruct]()),
			},
		},
		"typed nil Source": {
			Source: typedNilSource,
			Target: &emptyStruct{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningSourceIsNil(reflect.TypeFor[*emptyStruct]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*emptyStruct](), reflect.TypeFor[*emptyStruct]()),
				errorSourceIsNil("", reflect.TypeFor[*emptyStruct](), "", reflect.TypeFor[*emptyStruct]()),
			},
		},
		"nil Target": {
			Source: emptyStruct{},
			expectedDiags: diag.Diagnostics{
				diagConvertingTargetIsNil(nil),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[emptyStruct](), nil),
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
				infoFlattening(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*emptyStruct]()),
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
				infoFlattening(reflect.TypeFor[emptyStruct](), reflect.TypeFor[int]()),
				errorTargetIsNotPointer("", reflect.TypeFor[emptyStruct](), "", reflect.TypeFor[int]()),
			},
		},
		"non-struct Source struct Target": {
			Source: testString,
			Target: &emptyStruct{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeFor[emptyStruct]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[string](), reflect.TypeFor[*emptyStruct]()),
				errorTargetDoesNotImplementAttrValue("", reflect.TypeFor[string](), "", reflect.TypeFor[emptyStruct]()),
			},
		},
		"struct Source non-struct Target": {
			Source: emptyStruct{},
			Target: &testString,
			expectedDiags: diag.Diagnostics{
				diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeFor[string]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*string]()),
				errorTargetDoesNotImplementAttrValue("", reflect.TypeFor[emptyStruct](), "", reflect.TypeFor[string]()),
			},
		},
		"empty struct Source and Target": {
			Source:     emptyStruct{},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*emptyStruct]()),
				infoConverting(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*emptyStruct]()),
			},
		},
		"empty struct pointer Source and Target": {
			Source:     &emptyStruct{},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*emptyStruct](), reflect.TypeFor[*emptyStruct]()),
				infoConverting(reflect.TypeFor[emptyStruct](), reflect.TypeFor[*emptyStruct]()),
			},
		},
		"single string struct pointer Source and empty Target": {
			Source:     &awsSingleStringValue{Field1: "a"},
			Target:     &emptyStruct{},
			WantTarget: &emptyStruct{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringValue](), reflect.TypeFor[*emptyStruct]()),
				infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*emptyStruct]()),
				debugNoCorrespondingField(reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*emptyStruct]()),
			},
		},
		"target field does not implement attr.Value Target": {
			Source: &awsSingleStringValue{Field1: "a"},
			Target: &awsSingleStringValue{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeFor[string]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringValue](), reflect.TypeFor[*awsSingleStringValue]()),
				infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*awsSingleStringValue]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*awsSingleStringValue]()),
				errorTargetDoesNotImplementAttrValue("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[string]()),
			},
		},
		"single empty string Source and single string Target": {
			Source:     &awsSingleStringValue{},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringValue("")},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringValue](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		"single string Source and single string Target": {
			Source:     &awsSingleStringValue{Field1: "a"},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringValue("a")},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringValue](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		"single byte slice Source and single string Target": {
			Source:     &awsSingleByteSliceValue{Field1: []byte("a")},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringValue("a")},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleByteSliceValue](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsSingleByteSliceValue](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleByteSliceValue](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]byte](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		"single nil *string Source and single string Target": {
			Source:     &awsSingleStringPointer{},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringNull()},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		"single *string Source and single string Target": {
			Source:     &awsSingleStringPointer{Field1: aws.String("a")},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{Field1: types.StringValue("a")},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
		"single string Source and single int64 Target": {
			Source:     &awsSingleStringValue{Field1: "a"},
			Target:     &tfSingleInt64Field{},
			WantTarget: &tfSingleInt64Field{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringValue](), reflect.TypeFor[*tfSingleInt64Field]()),
				infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleInt64Field]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.Int64]()),
				{
					"@level":             "error",
					"@module":            "provider.autoflex",
					"@message":           "AutoFlex Flatten; incompatible types",
					"from":               float64(reflect.String),
					"to":                 map[string]any{},
					logAttrKeySourcePath: "Field1",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[string]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[types.Int64]()),
				},
			},
		},
		"zero value primtive types Source and primtive types Target": {
			Source: &awsAllThePrimitiveFields{},
			Target: &tfAllThePrimitiveFields{},
			WantTarget: &tfAllThePrimitiveFields{
				Field1:  types.StringValue(""),
				Field2:  types.StringNull(),
				Field3:  types.Int64Value(0),
				Field4:  types.Int64Null(),
				Field5:  types.Int64Value(0),
				Field6:  types.Int64Null(),
				Field7:  types.Float64Value(0),
				Field8:  types.Float64Null(),
				Field9:  types.Float64Value(0),
				Field10: types.Float64Null(),
				Field11: types.BoolValue(false),
				Field12: types.BoolNull(),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsAllThePrimitiveFields](), reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConverting(reflect.TypeFor[awsAllThePrimitiveFields](), reflect.TypeFor[*tfAllThePrimitiveFields]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsAllThePrimitiveFields](), "Field1", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				traceMatchedFields("Field2", reflect.TypeFor[awsAllThePrimitiveFields](), "Field2", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*string](), "Field2", reflect.TypeFor[types.String]()),
				traceMatchedFields("Field3", reflect.TypeFor[awsAllThePrimitiveFields](), "Field3", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[int32](), "Field3", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field4", reflect.TypeFor[awsAllThePrimitiveFields](), "Field4", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[*int32](), "Field4", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field5", reflect.TypeFor[awsAllThePrimitiveFields](), "Field5", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[int64](), "Field5", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field6", reflect.TypeFor[awsAllThePrimitiveFields](), "Field6", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[*int64](), "Field6", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field7", reflect.TypeFor[awsAllThePrimitiveFields](), "Field7", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field7", reflect.TypeFor[float32](), "Field7", reflect.TypeFor[types.Float64]()),
				traceMatchedFields("Field8", reflect.TypeFor[awsAllThePrimitiveFields](), "Field8", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field8", reflect.TypeFor[*float32](), "Field8", reflect.TypeFor[types.Float64]()),
				traceMatchedFields("Field9", reflect.TypeFor[awsAllThePrimitiveFields](), "Field9", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field9", reflect.TypeFor[float64](), "Field9", reflect.TypeFor[types.Float64]()),
				traceMatchedFields("Field10", reflect.TypeFor[awsAllThePrimitiveFields](), "Field10", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field10", reflect.TypeFor[*float64](), "Field10", reflect.TypeFor[types.Float64]()),
				traceMatchedFields("Field11", reflect.TypeFor[awsAllThePrimitiveFields](), "Field11", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field11", reflect.TypeFor[bool](), "Field11", reflect.TypeFor[types.Bool]()),
				traceMatchedFields("Field12", reflect.TypeFor[awsAllThePrimitiveFields](), "Field12", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field12", reflect.TypeFor[*bool](), "Field12", reflect.TypeFor[types.Bool]()),
			},
		},
		"primtive types Source and primtive types Target": {
			Source: &awsAllThePrimitiveFields{
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
			Target: &tfAllThePrimitiveFields{},
			WantTarget: &tfAllThePrimitiveFields{
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsAllThePrimitiveFields](), reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConverting(reflect.TypeFor[awsAllThePrimitiveFields](), reflect.TypeFor[*tfAllThePrimitiveFields]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsAllThePrimitiveFields](), "Field1", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				traceMatchedFields("Field2", reflect.TypeFor[awsAllThePrimitiveFields](), "Field2", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*string](), "Field2", reflect.TypeFor[types.String]()),
				traceMatchedFields("Field3", reflect.TypeFor[awsAllThePrimitiveFields](), "Field3", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[int32](), "Field3", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field4", reflect.TypeFor[awsAllThePrimitiveFields](), "Field4", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[*int32](), "Field4", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field5", reflect.TypeFor[awsAllThePrimitiveFields](), "Field5", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[int64](), "Field5", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field6", reflect.TypeFor[awsAllThePrimitiveFields](), "Field6", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[*int64](), "Field6", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field7", reflect.TypeFor[awsAllThePrimitiveFields](), "Field7", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field7", reflect.TypeFor[float32](), "Field7", reflect.TypeFor[types.Float64]()),
				traceMatchedFields("Field8", reflect.TypeFor[awsAllThePrimitiveFields](), "Field8", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field8", reflect.TypeFor[*float32](), "Field8", reflect.TypeFor[types.Float64]()),
				traceMatchedFields("Field9", reflect.TypeFor[awsAllThePrimitiveFields](), "Field9", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field9", reflect.TypeFor[float64](), "Field9", reflect.TypeFor[types.Float64]()),
				traceMatchedFields("Field10", reflect.TypeFor[awsAllThePrimitiveFields](), "Field10", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field10", reflect.TypeFor[*float64](), "Field10", reflect.TypeFor[types.Float64]()),
				traceMatchedFields("Field11", reflect.TypeFor[awsAllThePrimitiveFields](), "Field11", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field11", reflect.TypeFor[bool](), "Field11", reflect.TypeFor[types.Bool]()),
				traceMatchedFields("Field12", reflect.TypeFor[awsAllThePrimitiveFields](), "Field12", reflect.TypeFor[*tfAllThePrimitiveFields]()),
				infoConvertingWithPath("Field12", reflect.TypeFor[*bool](), "Field12", reflect.TypeFor[types.Bool]()),
			},
		},
		"zero value slice or map of primitive types Source and Collection of primtive types Target": {
			Source: &awsCollectionsOfPrimitiveElements{},
			Target: &tfCollectionsOfPrimitiveElements{},
			WantTarget: &tfCollectionsOfPrimitiveElements{
				Field1: types.ListNull(types.StringType),
				Field2: types.ListNull(types.StringType),
				Field3: types.SetNull(types.StringType),
				Field4: types.SetNull(types.StringType),
				Field5: types.MapNull(types.StringType),
				Field6: types.MapNull(types.StringType),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsCollectionsOfPrimitiveElements](), reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConverting(reflect.TypeFor[awsCollectionsOfPrimitiveElements](), reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field1", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
				traceFlatteningWithListNull("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
				traceMatchedFields("Field2", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field2", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[types.List]()),
				traceFlatteningWithListNull("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[types.List]()),
				traceMatchedFields("Field3", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field3", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[types.Set]()),
				traceFlatteningWithSetNull("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[types.Set]()),
				traceMatchedFields("Field4", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field4", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[types.Set]()),
				traceFlatteningWithSetNull("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[types.Set]()),
				traceMatchedFields("Field5", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field5", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[types.Map]()),
				traceFlatteningWithMapNull("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[types.Map]()),
				traceMatchedFields("Field6", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field6", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[types.Map]()),
				traceFlatteningWithMapNull("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[types.Map]()),
			},
		},
		"slice or map of primitive types Source and Collection of primtive types Target": {
			Source: &awsCollectionsOfPrimitiveElements{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			Target: &tfCollectionsOfPrimitiveElements{},
			WantTarget: &tfCollectionsOfPrimitiveElements{
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsCollectionsOfPrimitiveElements](), reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConverting(reflect.TypeFor[awsCollectionsOfPrimitiveElements](), reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field1", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Field1", reflect.TypeFor[[]string](), 2, "Field1", reflect.TypeFor[types.List]()),
				traceMatchedFields("Field2", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field2", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Field2", reflect.TypeFor[[]*string](), 2, "Field2", reflect.TypeFor[types.List]()),
				traceMatchedFields("Field3", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field3", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[types.Set]()),
				traceFlatteningWithSetValue("Field3", reflect.TypeFor[[]string](), 2, "Field3", reflect.TypeFor[types.Set]()),
				traceMatchedFields("Field4", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field4", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[types.Set]()),
				traceFlatteningWithSetValue("Field4", reflect.TypeFor[[]*string](), 2, "Field4", reflect.TypeFor[types.Set]()),
				traceMatchedFields("Field5", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field5", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[types.Map]()),
				traceFlatteningWithMapValue("Field5", reflect.TypeFor[map[string]string](), 2, "Field5", reflect.TypeFor[types.Map]()),
				traceMatchedFields("Field6", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field6", reflect.TypeFor[*tfCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[types.Map]()),
				traceFlatteningWithMapValue("Field6", reflect.TypeFor[map[string]*string](), 2, "Field6", reflect.TypeFor[types.Map]()),
			},
		},
		"zero value slice or map of string type Source and Collection of string types Target": {
			Source: &awsCollectionsOfPrimitiveElements{},
			Target: &tfTypedCollectionsOfPrimitiveElements{},
			WantTarget: &tfTypedCollectionsOfPrimitiveElements{
				Field1: fwtypes.NewListValueOfNull[types.String](ctx),
				Field2: fwtypes.NewListValueOfNull[types.String](ctx),
				Field3: fwtypes.NewSetValueOfNull[types.String](ctx),
				Field4: fwtypes.NewSetValueOfNull[types.String](ctx),
				Field5: fwtypes.NewMapValueOfNull[types.String](ctx),
				Field6: fwtypes.NewMapValueOfNull[types.String](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsCollectionsOfPrimitiveElements](), reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConverting(reflect.TypeFor[awsCollectionsOfPrimitiveElements](), reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field1", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceFlatteningWithListNull("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceMatchedFields("Field2", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field2", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceFlatteningWithListNull("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceMatchedFields("Field3", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field3", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				traceFlatteningWithSetNull("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				traceMatchedFields("Field4", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field4", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				traceFlatteningWithSetNull("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				traceMatchedFields("Field5", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field5", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapNull("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceMatchedFields("Field6", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field6", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapNull("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		"slice or map of string types Source and Collection of string types Target": {
			Source: &awsCollectionsOfPrimitiveElements{
				Field1: []string{"a", "b"},
				Field2: aws.StringSlice([]string{"a", "b"}),
				Field3: []string{"a", "b"},
				Field4: aws.StringSlice([]string{"a", "b"}),
				Field5: map[string]string{"A": "a", "B": "b"},
				Field6: aws.StringMap(map[string]string{"A": "a", "B": "b"}),
			},
			Target: &tfTypedCollectionsOfPrimitiveElements{},
			WantTarget: &tfTypedCollectionsOfPrimitiveElements{
				Field1: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field2: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field3: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field4: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
				}),
				Field5: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
				Field6: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"A": types.StringValue("a"),
					"B": types.StringValue("b"),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsCollectionsOfPrimitiveElements](), reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConverting(reflect.TypeFor[awsCollectionsOfPrimitiveElements](), reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field1", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceFlatteningWithListValue("Field1", reflect.TypeFor[[]string](), 2, "Field1", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceMatchedFields("Field2", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field2", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[[]*string](), "Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceFlatteningWithListValue("Field2", reflect.TypeFor[[]*string](), 2, "Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceMatchedFields("Field3", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field3", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[[]string](), "Field3", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				traceFlatteningWithSetValue("Field3", reflect.TypeFor[[]string](), 2, "Field3", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				traceMatchedFields("Field4", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field4", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[[]*string](), "Field4", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				traceFlatteningWithSetValue("Field4", reflect.TypeFor[[]*string](), 2, "Field4", reflect.TypeFor[fwtypes.SetValueOf[types.String]]()),
				traceMatchedFields("Field5", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field5", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field5", reflect.TypeFor[map[string]string](), "Field5", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapValue("Field5", reflect.TypeFor[map[string]string](), 2, "Field5", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceMatchedFields("Field6", reflect.TypeFor[awsCollectionsOfPrimitiveElements](), "Field6", reflect.TypeFor[*tfTypedCollectionsOfPrimitiveElements]()),
				infoConvertingWithPath("Field6", reflect.TypeFor[map[string]*string](), "Field6", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapValue("Field6", reflect.TypeFor[map[string]*string](), 2, "Field6", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsPluralSliceOfNestedObjectValues](), reflect.TypeFor[*tfSingluarListOfNestedObjects]()),
				infoConverting(reflect.TypeFor[awsPluralSliceOfNestedObjectValues](), reflect.TypeFor[*tfSingluarListOfNestedObjects]()),
				traceMatchedFields("Fields", reflect.TypeFor[awsPluralSliceOfNestedObjectValues](), "Field", reflect.TypeFor[*tfSingluarListOfNestedObjects]()),
				infoConvertingWithPath("Fields", reflect.TypeFor[[]awsSingleStringValue](), "Field", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				traceFlatteningNestedObjectCollection("Fields", reflect.TypeFor[[]awsSingleStringValue](), 1, "Field", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				traceMatchedFieldsWithPath("Fields[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Fields[0].Field1", reflect.TypeFor[string](), "Field[0].Field1", reflect.TypeFor[types.String]()),
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSpecialPluralization](), reflect.TypeFor[*tfSpecialPluralization]()),
				infoConverting(reflect.TypeFor[awsSpecialPluralization](), reflect.TypeFor[*tfSpecialPluralization]()),
				traceMatchedFields("Cities", reflect.TypeFor[awsSpecialPluralization](), "City", reflect.TypeFor[*tfSpecialPluralization]()),
				infoConvertingWithPath("Cities", reflect.TypeFor[[]*string](), "City", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Cities", reflect.TypeFor[[]*string](), 2, "City", reflect.TypeFor[types.List]()),
				traceMatchedFields("Coaches", reflect.TypeFor[awsSpecialPluralization](), "Coach", reflect.TypeFor[*tfSpecialPluralization]()),
				infoConvertingWithPath("Coaches", reflect.TypeFor[[]*string](), "Coach", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Coaches", reflect.TypeFor[[]*string](), 2, "Coach", reflect.TypeFor[types.List]()),
				traceMatchedFields("Tomatoes", reflect.TypeFor[awsSpecialPluralization](), "Tomato", reflect.TypeFor[*tfSpecialPluralization]()),
				infoConvertingWithPath("Tomatoes", reflect.TypeFor[[]*string](), "Tomato", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Tomatoes", reflect.TypeFor[[]*string](), 2, "Tomato", reflect.TypeFor[types.List]()),
				traceMatchedFields("Vertices", reflect.TypeFor[awsSpecialPluralization](), "Vertex", reflect.TypeFor[*tfSpecialPluralization]()),
				infoConvertingWithPath("Vertices", reflect.TypeFor[[]*string](), "Vertex", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Vertices", reflect.TypeFor[[]*string](), 2, "Vertex", reflect.TypeFor[types.List]()),
				traceMatchedFields("Criteria", reflect.TypeFor[awsSpecialPluralization](), "Criterion", reflect.TypeFor[*tfSpecialPluralization]()),
				infoConvertingWithPath("Criteria", reflect.TypeFor[[]*string](), "Criterion", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Criteria", reflect.TypeFor[[]*string](), 2, "Criterion", reflect.TypeFor[types.List]()),
				traceMatchedFields("Data", reflect.TypeFor[awsSpecialPluralization](), "Datum", reflect.TypeFor[*tfSpecialPluralization]()),
				infoConvertingWithPath("Data", reflect.TypeFor[[]*string](), "Datum", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Data", reflect.TypeFor[[]*string](), 2, "Datum", reflect.TypeFor[types.List]()),
				traceMatchedFields("Hives", reflect.TypeFor[awsSpecialPluralization](), "Hive", reflect.TypeFor[*tfSpecialPluralization]()),
				infoConvertingWithPath("Hives", reflect.TypeFor[[]*string](), "Hive", reflect.TypeFor[types.List]()),
				traceFlatteningWithListValue("Hives", reflect.TypeFor[[]*string](), 2, "Hive", reflect.TypeFor[types.List]()),
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsPluralAndSingularFields](), reflect.TypeFor[*tfPluralAndSingularFields]()),
				infoConverting(reflect.TypeFor[awsPluralAndSingularFields](), reflect.TypeFor[*tfPluralAndSingularFields]()),
				traceMatchedFields("Value", reflect.TypeFor[awsPluralAndSingularFields](), "Value", reflect.TypeFor[*tfPluralAndSingularFields]()),
				infoConvertingWithPath("Value", reflect.TypeFor[string](), "Value", reflect.TypeFor[types.String]()),
				debugNoCorrespondingField(reflect.TypeFor[awsPluralAndSingularFields](), "Values", reflect.TypeFor[*tfPluralAndSingularFields]()),
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsCapitalizationDiff](), reflect.TypeFor[*tfCaptializationDiff]()),
				infoConverting(reflect.TypeFor[awsCapitalizationDiff](), reflect.TypeFor[*tfCaptializationDiff]()),
				traceMatchedFields("FieldUrl", reflect.TypeFor[awsCapitalizationDiff](), "FieldURL", reflect.TypeFor[*tfCaptializationDiff]()),
				infoConvertingWithPath("FieldUrl", reflect.TypeFor[*string](), "FieldURL", reflect.TypeFor[types.String]()),
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsFieldNamePrefix](), reflect.TypeFor[*tfFieldNamePrefix]()),
				infoConverting(reflect.TypeFor[awsFieldNamePrefix](), reflect.TypeFor[*tfFieldNamePrefix]()),
				traceMatchedFields("IntentName", reflect.TypeFor[awsFieldNamePrefix](), "Name", reflect.TypeFor[*tfFieldNamePrefix]()),
				infoConvertingWithPath("IntentName", reflect.TypeFor[*string](), "Name", reflect.TypeFor[types.String]()),
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsFieldNameSuffix](), reflect.TypeFor[*tfFieldNameSuffix]()),
				infoConverting(reflect.TypeFor[awsFieldNameSuffix](), reflect.TypeFor[*tfFieldNameSuffix]()),
				traceMatchedFields("PolicyConfig", reflect.TypeFor[awsFieldNameSuffix](), "Policy", reflect.TypeFor[*tfFieldNameSuffix]()),
				infoConvertingWithPath("PolicyConfig", reflect.TypeFor[*string](), "Policy", reflect.TypeFor[types.String]()),
			},
		},
		"single string Source and single ARN Target": {
			Source:     &awsSingleStringValue{Field1: testARN},
			Target:     &tfSingleARNField{},
			WantTarget: &tfSingleARNField{Field1: fwtypes.ARNValue(testARN)},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringValue](), reflect.TypeFor[*tfSingleARNField]()),
				infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleARNField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleARNField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[fwtypes.ARN]()),
			},
		},
		"single *string Source and single ARN Target": {
			Source:     &awsSingleStringPointer{Field1: aws.String(testARN)},
			Target:     &tfSingleARNField{},
			WantTarget: &tfSingleARNField{Field1: fwtypes.ARNValue(testARN)},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringPointer](), reflect.TypeFor[*tfSingleARNField]()),
				infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleARNField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleARNField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[fwtypes.ARN]()),
			},
		},
		"single nil *string Source and single ARN Target": {
			Source:     &awsSingleStringPointer{},
			Target:     &tfSingleARNField{},
			WantTarget: &tfSingleARNField{Field1: fwtypes.ARNNull()},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsSingleStringPointer](), reflect.TypeFor[*tfSingleARNField]()),
				infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleARNField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleARNField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[fwtypes.ARN]()),
			},
		},
		"timestamp": {
			Source: &awsRFC3339TimeValue{
				CreationDateTime: testTimeTime,
			},
			Target: &tfRFC3339Time{},
			WantTarget: &tfRFC3339Time{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsRFC3339TimeValue](), reflect.TypeFor[*tfRFC3339Time]()),
				infoConverting(reflect.TypeFor[awsRFC3339TimeValue](), reflect.TypeFor[*tfRFC3339Time]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[awsRFC3339TimeValue](), "CreationDateTime", reflect.TypeFor[*tfRFC3339Time]()),
				infoConvertingWithPath("CreationDateTime", reflect.TypeFor[time.Time](), "CreationDateTime", reflect.TypeFor[timetypes.RFC3339]()),
			},
		},
		"timestamp pointer": {
			Source: &awsRFC3339TimePointer{
				CreationDateTime: &testTimeTime,
			},
			Target: &tfRFC3339Time{},
			WantTarget: &tfRFC3339Time{
				CreationDateTime: timetypes.NewRFC3339ValueMust(testTimeStr),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsRFC3339TimePointer](), reflect.TypeFor[*tfRFC3339Time]()),
				infoConverting(reflect.TypeFor[awsRFC3339TimePointer](), reflect.TypeFor[*tfRFC3339Time]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[awsRFC3339TimePointer](), "CreationDateTime", reflect.TypeFor[*tfRFC3339Time]()),
				infoConvertingWithPath("CreationDateTime", reflect.TypeFor[*time.Time](), "CreationDateTime", reflect.TypeFor[timetypes.RFC3339]()),
			},
		},
		"timestamp nil": {
			Source: &awsRFC3339TimePointer{},
			Target: &tfRFC3339Time{},
			WantTarget: &tfRFC3339Time{
				CreationDateTime: timetypes.NewRFC3339Null(),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsRFC3339TimePointer](), reflect.TypeFor[*tfRFC3339Time]()),
				infoConverting(reflect.TypeFor[awsRFC3339TimePointer](), reflect.TypeFor[*tfRFC3339Time]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[awsRFC3339TimePointer](), "CreationDateTime", reflect.TypeFor[*tfRFC3339Time]()),
				infoConvertingWithPath("CreationDateTime", reflect.TypeFor[*time.Time](), "CreationDateTime", reflect.TypeFor[timetypes.RFC3339]()),
			},
		},
		"timestamp empty": {
			Source: &awsRFC3339TimeValue{},
			Target: &tfRFC3339Time{},
			WantTarget: &tfRFC3339Time{
				CreationDateTime: timetypes.NewRFC3339TimeValue(zeroTime),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsRFC3339TimeValue](), reflect.TypeFor[*tfRFC3339Time]()),
				infoConverting(reflect.TypeFor[awsRFC3339TimeValue](), reflect.TypeFor[*tfRFC3339Time]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[awsRFC3339TimeValue](), "CreationDateTime", reflect.TypeFor[*tfRFC3339Time]()),
				infoConvertingWithPath("CreationDateTime", reflect.TypeFor[time.Time](), "CreationDateTime", reflect.TypeFor[timetypes.RFC3339]()),
			},
		},

		"source struct field to non-attr.Value": {
			Source: &awsRFC3339TimeValue{},
			Target: &awsRFC3339TimeValue{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeFor[time.Time]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsRFC3339TimeValue](), reflect.TypeFor[*awsRFC3339TimeValue]()),
				infoConverting(reflect.TypeFor[awsRFC3339TimeValue](), reflect.TypeFor[*awsRFC3339TimeValue]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[awsRFC3339TimeValue](), "CreationDateTime", reflect.TypeFor[*awsRFC3339TimeValue]()),
				errorTargetDoesNotImplementAttrValue("CreationDateTime", reflect.TypeFor[time.Time](), "CreationDateTime", reflect.TypeFor[time.Time]()),
			},
		},
		"source struct ptr field to non-attr.Value": {
			Source: &awsRFC3339TimePointer{},
			Target: &awsRFC3339TimeValue{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeFor[time.Time]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsRFC3339TimePointer](), reflect.TypeFor[*awsRFC3339TimeValue]()),
				infoConverting(reflect.TypeFor[awsRFC3339TimePointer](), reflect.TypeFor[*awsRFC3339TimeValue]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[awsRFC3339TimePointer](), "CreationDateTime", reflect.TypeFor[*awsRFC3339TimeValue]()),
				errorTargetDoesNotImplementAttrValue("CreationDateTime", reflect.TypeFor[*time.Time](), "CreationDateTime", reflect.TypeFor[time.Time]()),
			},
		},
		"source struct field to non-attr.Value ptr": {
			Source: &awsRFC3339TimeValue{},
			Target: &awsRFC3339TimePointer{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeFor[*time.Time]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsRFC3339TimeValue](), reflect.TypeFor[*awsRFC3339TimePointer]()),
				infoConverting(reflect.TypeFor[awsRFC3339TimeValue](), reflect.TypeFor[*awsRFC3339TimePointer]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[awsRFC3339TimeValue](), "CreationDateTime", reflect.TypeFor[*awsRFC3339TimePointer]()),
				errorTargetDoesNotImplementAttrValue("CreationDateTime", reflect.TypeFor[time.Time](), "CreationDateTime", reflect.TypeFor[*time.Time]()),
			},
		},
		"source struct ptr field to non-attr.Value ptr": {
			Source: &awsRFC3339TimePointer{},
			Target: &awsRFC3339TimePointer{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeFor[*time.Time]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsRFC3339TimePointer](), reflect.TypeFor[*awsRFC3339TimePointer]()),
				infoConverting(reflect.TypeFor[awsRFC3339TimePointer](), reflect.TypeFor[*awsRFC3339TimePointer]()),
				traceMatchedFields("CreationDateTime", reflect.TypeFor[awsRFC3339TimePointer](), "CreationDateTime", reflect.TypeFor[*awsRFC3339TimePointer]()),
				errorTargetDoesNotImplementAttrValue("CreationDateTime", reflect.TypeFor[*time.Time](), "CreationDateTime", reflect.TypeFor[*time.Time]()),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenGeneric(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"complex Source and complex Target": {
			Source: &awsComplexValue{
				Field1: "m",
				Field2: &awsNestedObjectPointer{Field1: &awsSingleStringValue{Field1: "n"}},
				Field3: aws.StringMap(map[string]string{"X": "x", "Y": "y"}),
				Field4: []awsSingleInt64Value{{Field1: 100}, {Field1: 2000}, {Field1: 30000}},
			},
			Target: &tfComplexValue{},
			WantTarget: &tfComplexValue{
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsComplexValue](), reflect.TypeFor[*tfComplexValue]()),
				infoConverting(reflect.TypeFor[awsComplexValue](), reflect.TypeFor[*tfComplexValue]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsComplexValue](), "Field1", reflect.TypeFor[*tfComplexValue]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),

				traceMatchedFields("Field2", reflect.TypeFor[awsComplexValue](), "Field2", reflect.TypeFor[*tfComplexValue]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*awsNestedObjectPointer](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfListOfNestedObject]]()),
				traceMatchedFieldsWithPath("Field2", "Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field2", "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
				infoConvertingWithPath("Field2.Field1", reflect.TypeFor[*awsSingleStringValue](), "Field2.Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				traceMatchedFieldsWithPath("Field2.Field1", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field2.Field1", "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field2.Field1.Field1", reflect.TypeFor[string](), "Field2.Field1.Field1", reflect.TypeFor[types.String]()),

				traceMatchedFields("Field3", reflect.TypeFor[awsComplexValue](), "Field3", reflect.TypeFor[*tfComplexValue]()),
				infoConvertingWithPath("Field3", reflect.TypeFor[map[string]*string](), "Field3", reflect.TypeFor[types.Map]()),
				traceFlatteningWithMapValue("Field3", reflect.TypeFor[map[string]*string](), 2, "Field3", reflect.TypeFor[types.Map]()),

				traceMatchedFields("Field4", reflect.TypeFor[awsComplexValue](), "Field4", reflect.TypeFor[*tfComplexValue]()),
				infoConvertingWithPath("Field4", reflect.TypeFor[[]awsSingleInt64Value](), "Field4", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleInt64Field]]()),
				traceFlatteningNestedObjectCollection("Field4", reflect.TypeFor[[]awsSingleInt64Value](), 3, "Field4", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleInt64Field]]()),
				traceMatchedFieldsWithPath("Field4[0]", "Field1", reflect.TypeFor[awsSingleInt64Value](), "Field4[0]", "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
				infoConvertingWithPath("Field4[0].Field1", reflect.TypeFor[int64](), "Field4[0].Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFieldsWithPath("Field4[1]", "Field1", reflect.TypeFor[awsSingleInt64Value](), "Field4[1]", "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
				infoConvertingWithPath("Field4[1].Field1", reflect.TypeFor[int64](), "Field4[1].Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFieldsWithPath("Field4[2]", "Field1", reflect.TypeFor[awsSingleInt64Value](), "Field4[2]", "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
				infoConvertingWithPath("Field4[2].Field1", reflect.TypeFor[int64](), "Field4[2].Field1", reflect.TypeFor[types.Int64]()),
			},
		},
		"map of string": {
			Source: &awsMapOfString{
				FieldInner: map[string]string{
					"x": "y",
				},
			},
			Target: &tfMapOfString{},
			WantTarget: &tfMapOfString{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapOfString](), reflect.TypeFor[*tfMapOfString]()),
				infoConverting(reflect.TypeFor[awsMapOfString](), reflect.TypeFor[*tfMapOfString]()),
				traceMatchedFields("FieldInner", reflect.TypeFor[awsMapOfString](), "FieldInner", reflect.TypeFor[*tfMapOfString]()),
				infoConvertingWithPath("FieldInner", reflect.TypeFor[map[string]string](), "FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapValue("FieldInner", reflect.TypeFor[map[string]string](), 1, "FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		"map of string pointer": {
			Source: &awsMapOfStringPointer{
				FieldInner: map[string]*string{
					"x": aws.String("y"),
				},
			},
			Target: &tfMapOfString{},
			WantTarget: &tfMapOfString{
				FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
					"x": types.StringValue("y"),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapOfStringPointer](), reflect.TypeFor[*tfMapOfString]()),
				infoConverting(reflect.TypeFor[awsMapOfStringPointer](), reflect.TypeFor[*tfMapOfString]()),
				traceMatchedFields("FieldInner", reflect.TypeFor[awsMapOfStringPointer](), "FieldInner", reflect.TypeFor[*tfMapOfString]()),
				infoConvertingWithPath("FieldInner", reflect.TypeFor[map[string]*string](), "FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapValue("FieldInner", reflect.TypeFor[map[string]*string](), 1, "FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		"nested string map": {
			Source: &awsNestedMapOfString{
				FieldOuter: awsMapOfString{
					FieldInner: map[string]string{
						"x": "y",
					},
				},
			},
			Target: &tfNestedMapOfString{},
			WantTarget: &tfNestedMapOfString{
				FieldOuter: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfMapOfString{
					FieldInner: fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"x": types.StringValue("y"),
					}),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsNestedMapOfString](), reflect.TypeFor[*tfNestedMapOfString]()),
				infoConverting(reflect.TypeFor[awsNestedMapOfString](), reflect.TypeFor[*tfNestedMapOfString]()),
				traceMatchedFields("FieldOuter", reflect.TypeFor[awsNestedMapOfString](), "FieldOuter", reflect.TypeFor[*tfNestedMapOfString]()),
				infoConvertingWithPath("FieldOuter", reflect.TypeFor[awsMapOfString](), "FieldOuter", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapOfString]]()),
				traceMatchedFieldsWithPath("FieldOuter", "FieldInner", reflect.TypeFor[awsMapOfString](), "FieldOuter", "FieldInner", reflect.TypeFor[*tfMapOfString]()),
				infoConvertingWithPath("FieldOuter.FieldInner", reflect.TypeFor[map[string]string](), "FieldOuter.FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapValue("FieldOuter.FieldInner", reflect.TypeFor[map[string]string](), 1, "FieldOuter.FieldInner", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
		"map of map of string": {
			Source: &awsMapOfMapOfString{
				Field1: map[string]map[string]string{
					"x": {
						"y": "z",
					},
				},
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapOfMapOfString](), reflect.TypeFor[*tfMapOfMapOfString]()),
				infoConverting(reflect.TypeFor[awsMapOfMapOfString](), reflect.TypeFor[*tfMapOfMapOfString]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsMapOfMapOfString](), "Field1", reflect.TypeFor[*tfMapOfMapOfString]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[map[string]map[string]string](), "Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]]()),
				traceFlatteningMap("Field1", reflect.TypeFor[map[string]map[string]string](), 1, "Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]]()),
				traceFlatteningWithNewMapValueOf("Field1[\"x\"]", reflect.TypeFor[map[string]string](), 1, "Field1[\"x\"]", reflect.TypeFor[map[string]attr.Value]()),
			},
		},
		"map of map of string pointer": {
			Source: &awsMapOfMapOfStringPointer{
				Field1: map[string]map[string]*string{
					"x": {
						"y": aws.String("z"),
					},
				},
			},
			Target: &tfMapOfMapOfString{},
			WantTarget: &tfMapOfMapOfString{
				Field1: fwtypes.NewMapValueOfMust[fwtypes.MapValueOf[types.String]](ctx, map[string]attr.Value{
					"x": fwtypes.NewMapValueOfMust[types.String](ctx, map[string]attr.Value{
						"y": types.StringValue("z"),
					}),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapOfMapOfStringPointer](), reflect.TypeFor[*tfMapOfMapOfString]()),
				infoConverting(reflect.TypeFor[awsMapOfMapOfStringPointer](), reflect.TypeFor[*tfMapOfMapOfString]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsMapOfMapOfStringPointer](), "Field1", reflect.TypeFor[*tfMapOfMapOfString]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[map[string]map[string]*string](), "Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]]()),
				traceFlatteningMap("Field1", reflect.TypeFor[map[string]map[string]*string](), 1, "Field1", reflect.TypeFor[fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]]]()),
				traceFlatteningWithNewMapValueOf("Field1[\"x\"]", reflect.TypeFor[map[string]*string](), 1, "Field1[\"x\"]", reflect.TypeFor[map[string]attr.Value]()),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"bool to Bool": {
			"true": {
				Source: awsSingleBoolValue{
					Field1: true,
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolValue(true),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleBoolValue](), reflect.TypeFor[*tfSingleBoolField]()),
					infoConverting(reflect.TypeFor[awsSingleBoolValue](), reflect.TypeFor[*tfSingleBoolField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleBoolValue](), "Field1", reflect.TypeFor[*tfSingleBoolField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[bool](), "Field1", reflect.TypeFor[types.Bool]()),
				},
			},
			"false": {
				Source: awsSingleBoolValue{
					Field1: false,
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolValue(false),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleBoolValue](), reflect.TypeFor[*tfSingleBoolField]()),
					infoConverting(reflect.TypeFor[awsSingleBoolValue](), reflect.TypeFor[*tfSingleBoolField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleBoolValue](), "Field1", reflect.TypeFor[*tfSingleBoolField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[bool](), "Field1", reflect.TypeFor[types.Bool]()),
				},
			},
		},

		"*bool to Bool": {
			"true": {
				Source: awsSingleBoolPointer{
					Field1: aws.Bool(true),
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolValue(true),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolField]()),
					infoConverting(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleBoolPointer](), "Field1", reflect.TypeFor[*tfSingleBoolField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
				},
			},
			"false": {
				Source: awsSingleBoolPointer{
					Field1: aws.Bool(false),
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolValue(false),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolField]()),
					infoConverting(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleBoolPointer](), "Field1", reflect.TypeFor[*tfSingleBoolField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
				},
			},
			"null": {
				Source: awsSingleBoolPointer{
					Field1: nil,
				},
				Target: &tfSingleBoolField{},
				WantTarget: &tfSingleBoolField{
					Field1: types.BoolNull(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolField]()),
					infoConverting(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleBoolPointer](), "Field1", reflect.TypeFor[*tfSingleBoolField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
				},
			},
		},

		"legacy *bool to Bool": {
			"true": {
				Source: awsSingleBoolPointer{
					Field1: aws.Bool(true),
				},
				Target: &tfSingleBoolFieldLegacy{},
				WantTarget: &tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(true),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleBoolPointer](), "Field1", reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
				},
			},
			"false": {
				Source: awsSingleBoolPointer{
					Field1: aws.Bool(false),
				},
				Target: &tfSingleBoolFieldLegacy{},
				WantTarget: &tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(false),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleBoolPointer](), "Field1", reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
				},
			},
			"null": {
				Source: awsSingleBoolPointer{
					Field1: nil,
				},
				Target: &tfSingleBoolFieldLegacy{},
				WantTarget: &tfSingleBoolFieldLegacy{
					Field1: types.BoolValue(false),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleBoolPointer](), reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleBoolPointer](), "Field1", reflect.TypeFor[*tfSingleBoolFieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*bool](), "Field1", reflect.TypeFor[types.Bool]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"float64 to Float64": {
			"value": {
				Source: awsSingleFloat64Value{
					Field1: 42,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Value](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Value](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Value](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[float64](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"zero": {
				Source: awsSingleFloat64Value{
					Field1: 0,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Value](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Value](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Value](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[float64](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
		},

		"*float64 to Float64": {
			"value": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"zero": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(0),
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"null": {
				Source: awsSingleFloat64Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
		},

		"legacy *float64 to Float64": {
			"value": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"zero": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(0),
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"null": {
				Source: awsSingleFloat64Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
		},

		// For historical reasons, float32 can be flattened to Float64 values
		"float32 to Float64": {
			"value": {
				Source: awsSingleFloat32Value{
					Field1: 42,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Value](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Value](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Value](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[float32](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"zero": {
				Source: awsSingleFloat32Value{
					Field1: 0,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Value](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Value](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Value](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[float32](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
		},

		"*float32 to Float64": {
			"value": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"zero": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"null": {
				Source: awsSingleFloat32Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat64Field{},
				WantTarget: &tfSingleFloat64Field{
					Field1: types.Float64Null(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
		},

		"legacy *float32 to Float64": {
			"value": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"zero": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
			"null": {
				Source: awsSingleFloat32Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat64FieldLegacy{},
				WantTarget: &tfSingleFloat64FieldLegacy{
					Field1: types.Float64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float64]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenFloat32(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"float32 to Float32": {
			"value": {
				Source: awsSingleFloat32Value{
					Field1: 42,
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Value](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Value](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Value](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[float32](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
			"zero": {
				Source: awsSingleFloat32Value{
					Field1: 0,
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Value](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Value](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Value](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[float32](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
		},

		"*float32 to Float32": {
			"value": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
			"zero": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
			"null": {
				Source: awsSingleFloat32Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat32Field{},
				WantTarget: &tfSingleFloat32Field{
					Field1: types.Float32Null(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
		},

		"legacy *float32 to Float32": {
			"value": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(42),
				},
				Target: &tfSingleFloat32FieldLegacy{},
				WantTarget: &tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
			"zero": {
				Source: awsSingleFloat32Pointer{
					Field1: aws.Float32(0),
				},
				Target: &tfSingleFloat32FieldLegacy{},
				WantTarget: &tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
			"null": {
				Source: awsSingleFloat32Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat32FieldLegacy{},
				WantTarget: &tfSingleFloat32FieldLegacy{
					Field1: types.Float32Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleFloat32Pointer](), reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat32Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*float32](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
		},

		// float64 cannot be flattened to Float32
		"float64 to Float32": {
			"value": {
				Source: awsSingleFloat64Value{
					Field1: 42,
				},
				Target: &tfSingleFloat32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[float64](), reflect.TypeFor[types.Float32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Value](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Value](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Value](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[float64](), "Field1", reflect.TypeFor[types.Float32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[float64](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
			"zero": {
				Source: awsSingleFloat64Value{
					Field1: 0,
				},
				Target: &tfSingleFloat32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[float64](), reflect.TypeFor[types.Float32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Value](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Value](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Value](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[float64](), "Field1", reflect.TypeFor[types.Float32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[float64](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
		},

		"*float64 to Float32": {
			"value": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(42),
				},
				Target: &tfSingleFloat32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[*float64](), reflect.TypeFor[types.Float32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
			"zero": {
				Source: awsSingleFloat64Pointer{
					Field1: aws.Float64(0),
				},
				Target: &tfSingleFloat32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[*float64](), reflect.TypeFor[types.Float32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
			"null": {
				Source: awsSingleFloat64Pointer{
					Field1: nil,
				},
				Target: &tfSingleFloat32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[*float64](), reflect.TypeFor[types.Float32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConverting(reflect.TypeFor[awsSingleFloat64Pointer](), reflect.TypeFor[*tfSingleFloat32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleFloat64Pointer](), "Field1", reflect.TypeFor[*tfSingleFloat32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[*float64](), "Field1", reflect.TypeFor[types.Float32]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenInt64(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"int64 to Int64": {
			"value": {
				Source: awsSingleInt64Value{
					Field1: 42,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Value](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Value](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Value](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"zero": {
				Source: awsSingleInt64Value{
					Field1: 0,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Value](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Value](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Value](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
		},

		"*int64 to Int64": {
			"value": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"zero": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(0),
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"null": {
				Source: awsSingleInt64Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
		},

		"legacy *int64 to Int64": {
			"value": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"zero": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(0),
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"null": {
				Source: awsSingleInt64Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
		},

		// For historical reasons, int32 can be flattened to Int64 values
		"int32 to Int64": {
			"value": {
				Source: awsSingleInt32Value{
					Field1: 42,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Value](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Value](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Value](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[int32](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"zero": {
				Source: awsSingleInt32Value{
					Field1: 0,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Value](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Value](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Value](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[int32](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
		},

		"*int32 to Int64": {
			"value": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"zero": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"null": {
				Source: awsSingleInt32Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt64Field{},
				WantTarget: &tfSingleInt64Field{
					Field1: types.Int64Null(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
		},

		"legacy *int32 to Int64": {
			"value": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"zero": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
			"null": {
				Source: awsSingleInt32Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt64FieldLegacy{},
				WantTarget: &tfSingleInt64FieldLegacy{
					Field1: types.Int64Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt64FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int64]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenInt32(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"int32 to Int32": {
			"value": {
				Source: awsSingleInt32Value{
					Field1: 42,
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Value](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Value](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Value](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[int32](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
			"zero": {
				Source: awsSingleInt32Value{
					Field1: 0,
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Value](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Value](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Value](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[int32](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
		},

		"*int32 to Int32": {
			"value": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
			"zero": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
			"null": {
				Source: awsSingleInt32Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt32Field{},
				WantTarget: &tfSingleInt32Field{
					Field1: types.Int32Null(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
		},

		"legacy *int32 to Int32": {
			"value": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(42),
				},
				Target: &tfSingleInt32FieldLegacy{},
				WantTarget: &tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(42),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
			"zero": {
				Source: awsSingleInt32Pointer{
					Field1: aws.Int32(0),
				},
				Target: &tfSingleInt32FieldLegacy{},
				WantTarget: &tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
			"null": {
				Source: awsSingleInt32Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt32FieldLegacy{},
				WantTarget: &tfSingleInt32FieldLegacy{
					Field1: types.Int32Value(0),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleInt32Pointer](), reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt32Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32FieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*int32](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
		},

		// int64 cannot be flattened to Int32
		"int64 to Int32": {
			"value": {
				Source: awsSingleInt64Value{
					Field1: 42,
				},
				Target: &tfSingleInt32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[int64](), reflect.TypeFor[types.Int32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Value](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Value](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Value](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
			"zero": {
				Source: awsSingleInt64Value{
					Field1: 0,
				},
				Target: &tfSingleInt32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[int64](), reflect.TypeFor[types.Int32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Value](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Value](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Value](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
		},

		"*int64 to Int32": {
			"value": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(42),
				},
				Target: &tfSingleInt32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[*int64](), reflect.TypeFor[types.Int32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
			"zero": {
				Source: awsSingleInt64Pointer{
					Field1: aws.Int64(0),
				},
				Target: &tfSingleInt32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[*int64](), reflect.TypeFor[types.Int32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
			"null": {
				Source: awsSingleInt64Pointer{
					Field1: nil,
				},
				Target: &tfSingleInt32Field{},
				expectedDiags: diag.Diagnostics{
					DiagFlatteningIncompatibleTypes(reflect.TypeFor[*int64](), reflect.TypeFor[types.Int32]()),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					infoConverting(reflect.TypeFor[awsSingleInt64Pointer](), reflect.TypeFor[*tfSingleInt32Field]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleInt64Pointer](), "Field1", reflect.TypeFor[*tfSingleInt32Field]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int32]()),
					errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[*int64](), "Field1", reflect.TypeFor[types.Int32]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenString(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"string to String": {
			"value": {
				Source: awsSingleStringValue{
					Field1: "a",
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringValue("a"),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringField]()),
					infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
			"zero": {
				Source: awsSingleStringValue{
					Field1: "",
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringValue(""),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringField]()),
					infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"*string to String": {
			"value": {
				Source: awsSingleStringPointer{
					Field1: aws.String("a"),
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringValue("a"),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
			"zero": {
				Source: awsSingleStringPointer{
					Field1: aws.String(""),
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringValue(""),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
			"null": {
				Source: awsSingleStringPointer{
					Field1: nil,
				},
				Target: &tfSingleStringField{},
				WantTarget: &tfSingleStringField{
					Field1: types.StringNull(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringField]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"omitempty string to String": {
			"value": {
				Source: awsSingleStringValue{
					Field1: "a",
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringValue("a"),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
			"zero": {
				Source: awsSingleStringValue{
					Field1: "",
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringNull(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"omitempty *string to String": {
			"value": {
				Source: awsSingleStringPointer{
					Field1: aws.String("a"),
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringValue("a"),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
			"zero": {
				Source: awsSingleStringPointer{
					Field1: aws.String(""),
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringNull(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
			"null": {
				Source: awsSingleStringPointer{
					Field1: nil,
				},
				Target: &tfSingleStringFieldOmitEmpty{},
				WantTarget: &tfSingleStringFieldOmitEmpty{
					Field1: types.StringNull(),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringFieldOmitEmpty]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"legacy *string to String": {
			"value": {
				Source: awsSingleStringPointer{
					Field1: aws.String("a"),
				},
				Target: &tfSingleStringFieldLegacy{},
				WantTarget: &tfSingleStringFieldLegacy{
					Field1: types.StringValue("a"),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
			"zero": {
				Source: awsSingleStringPointer{
					Field1: aws.String(""),
				},
				Target: &tfSingleStringFieldLegacy{},
				WantTarget: &tfSingleStringFieldLegacy{
					Field1: types.StringValue(""),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
			"null": {
				Source: awsSingleStringPointer{
					Field1: nil,
				},
				Target: &tfSingleStringFieldLegacy{},
				WantTarget: &tfSingleStringFieldLegacy{
					Field1: types.StringValue(""),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringFieldLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*string](), "Field1", reflect.TypeFor[types.String]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenTopLevelStringPtr(t *testing.T) {
	t.Parallel()

	testCases := toplevelTestCases[*string, types.String]{
		"value": {
			source:        aws.String("value"),
			expectedValue: types.StringValue("value"),
			expectedDiags: diag.Diagnostics{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*string](), reflect.TypeFor[*types.String]()),
				infoConverting(reflect.TypeFor[*string](), reflect.TypeFor[types.String]()),
			},
		},

		"empty": {
			source:        aws.String(""),
			expectedValue: types.StringValue(""),
			expectedDiags: diag.Diagnostics{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*string](), reflect.TypeFor[*types.String]()),
				infoConverting(reflect.TypeFor[*string](), reflect.TypeFor[types.String]()),
			},
		},

		"nil": {
			source:        nil,
			expectedValue: types.StringNull(),
			expectedDiags: diag.Diagnostics{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*string](), reflect.TypeFor[*types.String]()),
				infoConverting(reflect.TypeFor[*string](), reflect.TypeFor[types.String]()),
			},
		},
	}

	runTopLevelTestCases(t, testCases)
}

func TestFlattenTopLevelInt64Ptr(t *testing.T) {
	t.Parallel()

	testCases := toplevelTestCases[*int64, types.Int64]{
		"value": {
			source:        aws.Int64(42),
			expectedValue: types.Int64Value(42),
			expectedDiags: diag.Diagnostics{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*int64](), reflect.TypeFor[*types.Int64]()),
				infoConverting(reflect.TypeFor[*int64](), reflect.TypeFor[types.Int64]()),
			},
		},

		"empty": {
			source:        aws.Int64(0),
			expectedValue: types.Int64Value(0),
			expectedDiags: diag.Diagnostics{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*int64](), reflect.TypeFor[*types.Int64]()),
				infoConverting(reflect.TypeFor[*int64](), reflect.TypeFor[types.Int64]()),
			},
		},

		"nil": {
			source:        nil,
			expectedValue: types.Int64Null(),
			expectedDiags: diag.Diagnostics{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*int64](), reflect.TypeFor[*types.Int64]()),
				infoConverting(reflect.TypeFor[*int64](), reflect.TypeFor[types.Int64]()),
			},
		},
	}

	runTopLevelTestCases(t, testCases)
}

func TestFlattenSimpleNestedBlockWithStringEnum(t *testing.T) {
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
			Source: &aws01{
				Field1: 1,
				Field2: testEnumList,
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumValue(testEnumList),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf01]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[testEnum](), "Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]]()),
			},
		},
		"single nested empty value": {
			Source: &aws01{
				Field1: 1,
				Field2: "",
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: fwtypes.StringEnumNull[testEnum](),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf01]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[testEnum](), "Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenComplexNestedBlockWithStringEnum(t *testing.T) {
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
	var zero fwtypes.StringEnum[testEnum]
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: testEnumList,
				},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumValue(testEnumList),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf02]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				traceMatchedFieldsWithPath("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2.Field2", reflect.TypeFor[testEnum](), "Field2.Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]]()),
			},
		},
		"single nested empty value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{Field2: ""},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: fwtypes.StringEnumNull[testEnum](),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf02]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				traceMatchedFieldsWithPath("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2.Field2", reflect.TypeFor[testEnum](), "Field2.Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]]()),
			},
		},
		"single nested zero value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field2: ""},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field2: zero,
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf02]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				traceMatchedFieldsWithPath("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2.Field2", reflect.TypeFor[testEnum](), "Field2.Field2", reflect.TypeFor[fwtypes.StringEnum[testEnum]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenSimpleSingleNestedBlock(t *testing.T) {
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
			Source: &aws02{
				Field1: &aws01{
					Field1: aws.String("a"),
					Field2: 1,
				},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{
					Field1: types.StringValue("a"),
					Field2: types.Int64Value(1),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws02](), reflect.TypeFor[*tf02]()),
				infoConverting(reflect.TypeFor[aws02](), reflect.TypeFor[*tf02]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws02](), "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[aws01](), "Field1", "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[*string](), "Field1.Field1", reflect.TypeFor[types.String]()),
				traceMatchedFieldsWithPath("Field1", "Field2", reflect.TypeFor[aws01](), "Field1", "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1.Field2", reflect.TypeFor[int64](), "Field1.Field2", reflect.TypeFor[types.Int64]()),
			},
		},
		"single nested block nil": {
			Source: &aws02{},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfNull[tf01](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws02](), reflect.TypeFor[*tf02]()),
				infoConverting(reflect.TypeFor[aws02](), reflect.TypeFor[*tf02]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws02](), "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*aws01](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]]()),
			},
		},
		"single nested block value": {
			Source: &aws03{
				Field1: aws01{
					Field1: aws.String("a"),
					Field2: 1},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{
					Field1: types.StringValue("a"),
					Field2: types.Int64Value(1),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws03](), reflect.TypeFor[*tf02]()),
				infoConverting(reflect.TypeFor[aws03](), reflect.TypeFor[*tf02]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws03](), "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[aws01](), "Field1", "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[*string](), "Field1.Field1", reflect.TypeFor[types.String]()),
				traceMatchedFieldsWithPath("Field1", "Field2", reflect.TypeFor[aws01](), "Field1", "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1.Field2", reflect.TypeFor[int64](), "Field1.Field2", reflect.TypeFor[types.Int64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenComplexSingleNestedBlock(t *testing.T) {
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
			Source: &aws03{
				Field1: &aws02{
					Field1: &aws01{
						Field1: true,
						Field2: []string{"a", "b"},
					},
				},
			},
			Target: &tf03{},
			WantTarget: &tf03{
				Field1: fwtypes.NewObjectValueOfMust[tf02](ctx, &tf02{
					Field1: fwtypes.NewObjectValueOfMust[tf01](ctx, &tf01{
						Field1: types.BoolValue(true),
						Field2: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
							types.StringValue("a"),
							types.StringValue("b"),
						}),
					}),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws03](), reflect.TypeFor[*tf03]()),
				infoConverting(reflect.TypeFor[aws03](), reflect.TypeFor[*tf03]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws03](), "Field1", reflect.TypeFor[*tf03]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*aws02](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf02]]()),
				traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[aws02](), "Field1", "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1.Field1", reflect.TypeFor[*aws01](), "Field1.Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tf01]]()),
				traceMatchedFieldsWithPath("Field1.Field1", "Field1", reflect.TypeFor[aws01](), "Field1.Field1", "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1.Field1.Field1", reflect.TypeFor[bool](), "Field1.Field1.Field1", reflect.TypeFor[types.Bool]()),
				traceMatchedFieldsWithPath("Field1.Field1", "Field2", reflect.TypeFor[aws01](), "Field1.Field1", "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1.Field1.Field2", reflect.TypeFor[[]string](), "Field1.Field1.Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
				traceFlatteningWithListValue("Field1.Field1.Field2", reflect.TypeFor[[]string](), 2, "Field1.Field1.Field2", reflect.TypeFor[fwtypes.ListValueOf[types.String]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenSimpleNestedBlockWithFloat32(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Int64   `tfsdk:"field1"`
		Field2 types.Float64 `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 int64
		Field2 *float32
	}

	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source:     &aws01{Field1: 1, Field2: aws.Float32(0.01)},
			Target:     &tf01{},
			WantTarget: &tf01{Field1: types.Int64Value(1), Field2: types.Float64Value(0.01)},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf01]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*float32](), "Field2", reflect.TypeFor[types.Float64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenComplexNestedBlockWithFloat32(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Float64 `tfsdk:"field1"`
		Field2 types.Float64 `tfsdk:"field2"`
	}
	type tf02 struct {
		Field1 types.Int64                           `tfsdk:"field1"`
		Field2 fwtypes.ListNestedObjectValueOf[tf01] `tfsdk:"field2"`
	}
	type aws02 struct {
		Field1 float32
		Field2 *float32
	}
	type aws01 struct {
		Field1 int64
		Field2 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field1: 1.11,
					Field2: aws.Float32(-2.22),
				},
			},
			Target: &tf02{},
			WantTarget: &tf02{
				Field1: types.Int64Value(1),
				Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{
					Field1: types.Float64Value(1.11),
					Field2: types.Float64Value(-2.22),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf02]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				traceMatchedFieldsWithPath("Field2", "Field1", reflect.TypeFor[aws02](), "Field2", "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2.Field1", reflect.TypeFor[float32](), "Field2.Field1", reflect.TypeFor[types.Float64]()),
				traceMatchedFieldsWithPath("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2.Field2", reflect.TypeFor[*float32](), "Field2.Field2", reflect.TypeFor[types.Float64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenSimpleNestedBlockWithFloat64(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Int64   `tfsdk:"field1"`
		Field2 types.Float64 `tfsdk:"field2"`
	}
	type aws01 struct {
		Field1 int64
		Field2 *float64
	}

	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: aws.Float64(0.01),
			},
			Target: &tf01{},
			WantTarget: &tf01{
				Field1: types.Int64Value(1),
				Field2: types.Float64Value(0.01),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf01]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*float64](), "Field2", reflect.TypeFor[types.Float64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenComplexNestedBlockWithFloat64(t *testing.T) {
	t.Parallel()

	type tf01 struct {
		Field1 types.Float64 `tfsdk:"field1"`
		Field2 types.Float64 `tfsdk:"field2"`
	}
	type tf02 struct {
		Field1 types.Int64                           `tfsdk:"field1"`
		Field2 fwtypes.ListNestedObjectValueOf[tf01] `tfsdk:"field2"`
	}
	type aws02 struct {
		Field1 float64
		Field2 *float64
	}
	type aws01 struct {
		Field1 int64
		Field2 *aws02
	}

	ctx := context.Background()
	testCases := autoFlexTestCases{
		"single nested valid value": {
			Source: &aws01{
				Field1: 1,
				Field2: &aws02{
					Field1: 1.11,
					Field2: aws.Float64(-2.22),
				},
			},
			Target:     &tf02{},
			WantTarget: &tf02{Field1: types.Int64Value(1), Field2: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tf01{Field1: types.Float64Value(1.11), Field2: types.Float64Value(-2.22)})},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf02]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf02]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[int64](), "Field1", reflect.TypeFor[types.Int64]()),
				traceMatchedFields("Field2", reflect.TypeFor[aws01](), "Field2", reflect.TypeFor[*tf02]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[*aws02](), "Field2", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tf01]]()),
				traceMatchedFieldsWithPath("Field2", "Field1", reflect.TypeFor[aws02](), "Field2", "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2.Field1", reflect.TypeFor[float64](), "Field2.Field1", reflect.TypeFor[types.Float64]()),
				traceMatchedFieldsWithPath("Field2", "Field2", reflect.TypeFor[aws02](), "Field2", "Field2", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field2.Field2", reflect.TypeFor[*float64](), "Field2.Field2", reflect.TypeFor[types.Float64]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenObjectValueField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"*struct to ObjectValue": {
			"nil": {
				Source: awsNestedObjectPointer{},
				Target: &tfObjectValue[tfSingleStringField]{},
				WantTarget: &tfObjectValue[tfSingleStringField]{
					Field1: fwtypes.NewObjectValueOfNull[tfSingleStringField](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfObjectValue[tfSingleStringField]]()),
					infoConverting(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfObjectValue[tfSingleStringField]]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field1", reflect.TypeFor[*tfObjectValue[tfSingleStringField]]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfSingleStringField]]()),
				},
			},
			"value": {
				Source: awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
				Target: &tfObjectValue[tfSingleStringField]{},
				WantTarget: &tfObjectValue[tfSingleStringField]{
					Field1: fwtypes.NewObjectValueOfMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfObjectValue[tfSingleStringField]]()),
					infoConverting(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfObjectValue[tfSingleStringField]]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field1", reflect.TypeFor[*tfObjectValue[tfSingleStringField]]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1.Field1", reflect.TypeFor[string](), "Field1.Field1", reflect.TypeFor[types.String]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenListOfNestedObjectField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"*struct to ListNestedObject": {
			"nil": {
				Source: awsNestedObjectPointer{},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfListOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfListOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"value": {
				Source: awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfListOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfListOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1.Field1", reflect.TypeFor[string](), "Field1.Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"legacy *struct to ListNestedObject": {
			"nil": {
				Source: awsNestedObjectPointer{},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field1", reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"value": {
				Source: awsNestedObjectPointer{
					Field1: &awsSingleStringValue{
						Field1: "a",
					},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field1", reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1.Field1", reflect.TypeFor[string](), "Field1.Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"[]struct to ListNestedObject": {
			"nil": {
				Source: awsSliceOfNestedObjectValues{},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningWithNullValue("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"empty": {
				Source: awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsSingleStringValue](), 0, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"values": {
				Source: awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsSingleStringValue](), 2, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"legacy []struct to ListNestedObject": {
			"nil": {
				Source: awsSliceOfNestedObjectValues{},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"empty": {
				Source: awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsSingleStringValue](), 0, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"values": {
				Source: awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsSingleStringValue](), 2, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"[]*struct to ListNestedObject": {
			"nil": {
				Source: awsSliceOfNestedObjectPointers{},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningWithNullValue("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"empty": {
				Source: awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsSingleStringValue](), 0, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"values": {
				Source: awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfListOfNestedObject{},
				WantTarget: &tfListOfNestedObject{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfListOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsSingleStringValue](), 2, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"legacy []*struct to ListNestedObject": {
			"nil": {
				Source: awsSliceOfNestedObjectPointers{},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"empty": {
				Source: awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsSingleStringValue](), 0, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"values": {
				Source: awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfListOfNestedObjectLegacy{},
				WantTarget: &tfListOfNestedObjectLegacy{
					Field1: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfListOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsSingleStringValue](), 2, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenTopLevelListOfNestedObject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]toplevelTestCase[[]awsSingleStringValue, fwtypes.ListNestedObjectValueOf[tfSingleStringField]]{
		"values": {
			source: []awsSingleStringValue{
				{
					Field1: "value1",
				},
				{
					Field1: "value2",
				},
			},
			expectedValue: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
				{
					Field1: types.StringValue("value1"),
				},
				{
					Field1: types.StringValue("value2"),
				},
			}),
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[[]awsSingleStringValue](), reflect.TypeFor[*fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				infoConverting(reflect.TypeFor[[]awsSingleStringValue](), reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				traceFlatteningNestedObjectCollection("", reflect.TypeFor[[]awsSingleStringValue](), 2, "", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				traceMatchedFieldsWithPath("[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("[0].Field1", reflect.TypeFor[string](), "[0].Field1", reflect.TypeFor[types.String]()),
				traceMatchedFieldsWithPath("[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("[1].Field1", reflect.TypeFor[string](), "[1].Field1", reflect.TypeFor[types.String]()),
			},
		},

		"empty": {
			source:        []awsSingleStringValue{},
			expectedValue: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[[]awsSingleStringValue](), reflect.TypeFor[*fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				infoConverting(reflect.TypeFor[[]awsSingleStringValue](), reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				traceFlatteningNestedObjectCollection("", reflect.TypeFor[[]awsSingleStringValue](), 0, "", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
			},
		},

		"null": {
			source:        nil,
			expectedValue: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[[]awsSingleStringValue](), reflect.TypeFor[*fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				infoConverting(reflect.TypeFor[[]awsSingleStringValue](), reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				traceFlatteningWithNullValue("", reflect.TypeFor[[]awsSingleStringValue](), "", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
			},
		},
	}

	runTopLevelTestCases(t, testCases)
}

func TestFlattenSetOfNestedObjectField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"*struct to SetNestedObject": {
			"nil": {
				Source: awsNestedObjectPointer{},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfSetOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field1", reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"value": {
				Source: awsNestedObjectPointer{
					Field1: &awsSingleStringValue{Field1: "a"},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &tfSingleStringField{
						Field1: types.StringValue("a"),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsNestedObjectPointer](), reflect.TypeFor[*tfSetOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsNestedObjectPointer](), "Field1", reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1.Field1", reflect.TypeFor[string](), "Field1.Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"[]struct to SetNestedObject": {
			"nil": {
				Source: &awsSliceOfNestedObjectValues{},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningWithNullValue("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"empty": {
				Source: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsSingleStringValue](), 0, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"values": {
				Source: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsSingleStringValue](), 2, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"legacy []struct to SetNestedObject": {
			"nil": {
				Source: &awsSliceOfNestedObjectValues{},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"empty": {
				Source: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{},
				},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsSingleStringValue](), 0, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"values": {
				Source: &awsSliceOfNestedObjectValues{
					Field1: []awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectValues](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectValues](), "Field1", reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsSingleStringValue](), 2, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"[]*struct to SetNestedObject": {
			"nil": {
				Source: &awsSliceOfNestedObjectPointers{},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfNull[tfSingleStringField](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningWithNullValue("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"empty": {
				Source: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsSingleStringValue](), 0, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"values": {
				Source: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfSetOfNestedObject{},
				WantTarget: &tfSetOfNestedObject{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObject]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfSetOfNestedObject]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsSingleStringValue](), 2, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
				},
			},
		},

		"legacy []*struct to SetNestedObject": {
			"nil": {
				Source: &awsSliceOfNestedObjectPointers{},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"empty": {
				Source: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{},
				},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsSingleStringValue](), 0, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
			"values": {
				Source: &awsSliceOfNestedObjectPointers{
					Field1: []*awsSingleStringValue{
						{Field1: "a"},
						{Field1: "b"},
					},
				},
				Target: &tfSetOfNestedObjectLegacy{},
				WantTarget: &tfSetOfNestedObjectLegacy{
					Field1: fwtypes.NewSetNestedObjectValueOfSliceMust(ctx, []*tfSingleStringField{
						{Field1: types.StringValue("a")},
						{Field1: types.StringValue("b")},
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[*awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConverting(reflect.TypeFor[awsSliceOfNestedObjectPointers](), reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfNestedObjectPointers](), "Field1", reflect.TypeFor[*tfSetOfNestedObjectLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]*awsSingleStringValue](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsSingleStringValue](), 2, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfSingleStringField]]()),
					traceMatchedFieldsWithPath("Field1[0]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[0]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[0].Field1", reflect.TypeFor[string](), "Field1[0].Field1", reflect.TypeFor[types.String]()),
					traceMatchedFieldsWithPath("Field1[1]", "Field1", reflect.TypeFor[awsSingleStringValue](), "Field1[1]", "Field1", reflect.TypeFor[*tfSingleStringField]()),
					infoConvertingWithPath("Field1[1].Field1", reflect.TypeFor[string](), "Field1[1].Field1", reflect.TypeFor[types.String]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenMapBlock(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"nil map block key": {
			Source: &awsMapBlockValues{
				MapBlock: nil,
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfNull[tfMapBlockElement](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapBlockValues](), reflect.TypeFor[*tfMapBlockList]()),
				infoConverting(reflect.TypeFor[awsMapBlockValues](), reflect.TypeFor[*tfMapBlockList]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[awsMapBlockValues](), "MapBlock", reflect.TypeFor[*tfMapBlockList]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[map[string]awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]]()),
				traceFlatteningNullValue("MapBlock", reflect.TypeFor[map[string]awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]]()),
			},
		},
		"map block key list": {
			Source: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[tfMapBlockElement](ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapBlockValues](), reflect.TypeFor[*tfMapBlockList]()),
				infoConverting(reflect.TypeFor[awsMapBlockValues](), reflect.TypeFor[*tfMapBlockList]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[awsMapBlockValues](), "MapBlock", reflect.TypeFor[*tfMapBlockList]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[map[string]awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]]()),
				traceMatchedFieldsWithPath("MapBlock[\"x\"]", "Attr1", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr1", reflect.TypeFor[*tfMapBlockElement]()),
				infoConvertingWithPath("MapBlock[\"x\"].Attr1", reflect.TypeFor[string](), "MapBlock[0].Attr1", reflect.TypeFor[types.String]()),
				traceMatchedFieldsWithPath("MapBlock[\"x\"]", "Attr2", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr2", reflect.TypeFor[*tfMapBlockElement]()),
				infoConvertingWithPath("MapBlock[\"x\"].Attr2", reflect.TypeFor[string](), "MapBlock[0].Attr2", reflect.TypeFor[types.String]()),
			},
		},
		"map block key set": {
			Source: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockSet{},
			WantTarget: &tfMapBlockSet{
				MapBlock: fwtypes.NewSetNestedObjectValueOfValueSliceMust[tfMapBlockElement](ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapBlockValues](), reflect.TypeFor[*tfMapBlockSet]()),
				infoConverting(reflect.TypeFor[awsMapBlockValues](), reflect.TypeFor[*tfMapBlockSet]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[awsMapBlockValues](), "MapBlock", reflect.TypeFor[*tfMapBlockSet]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[map[string]awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfMapBlockElement]]()),
				traceMatchedFieldsWithPath("MapBlock[\"x\"]", "Attr1", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr1", reflect.TypeFor[*tfMapBlockElement]()),
				infoConvertingWithPath("MapBlock[\"x\"].Attr1", reflect.TypeFor[string](), "MapBlock[0].Attr1", reflect.TypeFor[types.String]()),
				traceMatchedFieldsWithPath("MapBlock[\"x\"]", "Attr2", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr2", reflect.TypeFor[*tfMapBlockElement]()),
				infoConvertingWithPath("MapBlock[\"x\"].Attr2", reflect.TypeFor[string](), "MapBlock[0].Attr2", reflect.TypeFor[types.String]()),
			},
		},
		"nil map block key ptr": {
			Source: &awsMapBlockPointers{
				MapBlock: nil,
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfNull[tfMapBlockElement](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapBlockPointers](), reflect.TypeFor[*tfMapBlockList]()),
				infoConverting(reflect.TypeFor[awsMapBlockPointers](), reflect.TypeFor[*tfMapBlockList]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[awsMapBlockPointers](), "MapBlock", reflect.TypeFor[*tfMapBlockList]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[map[string]*awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]]()),
				traceFlatteningNullValue("MapBlock", reflect.TypeFor[map[string]*awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]]()),
			},
		},
		"map block key ptr source": {
			Source: &awsMapBlockPointers{
				MapBlock: map[string]*awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[tfMapBlockElement](ctx, []tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapBlockPointers](), reflect.TypeFor[*tfMapBlockList]()),
				infoConverting(reflect.TypeFor[awsMapBlockPointers](), reflect.TypeFor[*tfMapBlockList]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[awsMapBlockPointers](), "MapBlock", reflect.TypeFor[*tfMapBlockList]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[map[string]*awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]]()),
				traceMatchedFieldsWithPath("MapBlock[\"x\"]", "Attr1", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr1", reflect.TypeFor[*tfMapBlockElement]()),
				infoConvertingWithPath("MapBlock[\"x\"].Attr1", reflect.TypeFor[string](), "MapBlock[0].Attr1", reflect.TypeFor[types.String]()),
				traceMatchedFieldsWithPath("MapBlock[\"x\"]", "Attr2", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr2", reflect.TypeFor[*tfMapBlockElement]()),
				infoConvertingWithPath("MapBlock[\"x\"].Attr2", reflect.TypeFor[string](), "MapBlock[0].Attr2", reflect.TypeFor[types.String]()),
			},
		},
		"map block key ptr both": {
			Source: &awsMapBlockPointers{
				MapBlock: map[string]*awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockList{},
			WantTarget: &tfMapBlockList{
				MapBlock: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tfMapBlockElement{
					{
						MapBlockKey: types.StringValue("x"),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapBlockPointers](), reflect.TypeFor[*tfMapBlockList]()),
				infoConverting(reflect.TypeFor[awsMapBlockPointers](), reflect.TypeFor[*tfMapBlockList]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[awsMapBlockPointers](), "MapBlock", reflect.TypeFor[*tfMapBlockList]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[map[string]*awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElement]]()),
				traceMatchedFieldsWithPath("MapBlock[\"x\"]", "Attr1", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr1", reflect.TypeFor[*tfMapBlockElement]()),
				infoConvertingWithPath("MapBlock[\"x\"].Attr1", reflect.TypeFor[string](), "MapBlock[0].Attr1", reflect.TypeFor[types.String]()),
				traceMatchedFieldsWithPath("MapBlock[\"x\"]", "Attr2", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr2", reflect.TypeFor[*tfMapBlockElement]()),
				infoConvertingWithPath("MapBlock[\"x\"].Attr2", reflect.TypeFor[string](), "MapBlock[0].Attr2", reflect.TypeFor[types.String]()),
			},
		},
		"map block enum key": {
			Source: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					string(testEnumList): {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockListEnumKey{},
			WantTarget: &tfMapBlockListEnumKey{
				MapBlock: fwtypes.NewListNestedObjectValueOfValueSliceMust[tfMapBlockElementEnumKey](ctx, []tfMapBlockElementEnumKey{
					{
						MapBlockKey: fwtypes.StringEnumValue(testEnumList),
						Attr1:       types.StringValue("a"),
						Attr2:       types.StringValue("b"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapBlockValues](), reflect.TypeFor[*tfMapBlockListEnumKey]()),
				infoConverting(reflect.TypeFor[awsMapBlockValues](), reflect.TypeFor[*tfMapBlockListEnumKey]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[awsMapBlockValues](), "MapBlock", reflect.TypeFor[*tfMapBlockListEnumKey]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[map[string]awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElementEnumKey]]()),
				traceMatchedFieldsWithPath("MapBlock[\"List\"]", "Attr1", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr1", reflect.TypeFor[*tfMapBlockElementEnumKey]()),
				infoConvertingWithPath("MapBlock[\"List\"].Attr1", reflect.TypeFor[string](), "MapBlock[0].Attr1", reflect.TypeFor[types.String]()),
				traceMatchedFieldsWithPath("MapBlock[\"List\"]", "Attr2", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", "Attr2", reflect.TypeFor[*tfMapBlockElementEnumKey]()),
				infoConvertingWithPath("MapBlock[\"List\"].Attr2", reflect.TypeFor[string](), "MapBlock[0].Attr2", reflect.TypeFor[types.String]()),
			},
		},

		"map block list no key": {
			Source: &awsMapBlockValues{
				MapBlock: map[string]awsMapBlockElement{
					"x": {
						Attr1: "a",
						Attr2: "b",
					},
				},
			},
			Target: &tfMapBlockListNoKey{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningNoMapBlockKey(reflect.TypeFor[tfMapBlockElementNoKey]()),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsMapBlockValues](), reflect.TypeFor[*tfMapBlockListNoKey]()),
				infoConverting(reflect.TypeFor[awsMapBlockValues](), reflect.TypeFor[*tfMapBlockListNoKey]()),
				traceMatchedFields("MapBlock", reflect.TypeFor[awsMapBlockValues](), "MapBlock", reflect.TypeFor[*tfMapBlockListNoKey]()),
				infoConvertingWithPath("MapBlock", reflect.TypeFor[map[string]awsMapBlockElement](), "MapBlock", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfMapBlockElementNoKey]]()),
				errorTargetHasNoMapBlockKey("MapBlock[\"x\"]", reflect.TypeFor[awsMapBlockElement](), "MapBlock[0]", reflect.TypeFor[tfMapBlockElementNoKey]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenSimpleListOfPrimitiveValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"regular": {
			"values": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{"a", "b"},
				},
				Target: &tfSimpleList{},
				WantTarget: &tfSimpleList{
					Field1: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleList]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleList]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleList]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
					traceFlatteningWithListValue("Field1", reflect.TypeFor[[]string](), 2, "Field1", reflect.TypeFor[types.List]()),
				},
			},

			"empty": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{},
				},
				Target: &tfSimpleList{},
				WantTarget: &tfSimpleList{
					Field1: types.ListValueMust(types.StringType, []attr.Value{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleList]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleList]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleList]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
					traceFlatteningWithListValue("Field1", reflect.TypeFor[[]string](), 0, "Field1", reflect.TypeFor[types.List]()),
				},
			},

			"null": {
				Source: awsSimpleStringValueSlice{
					Field1: nil,
				},
				Target: &tfSimpleList{},
				WantTarget: &tfSimpleList{
					Field1: types.ListNull(types.StringType),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleList]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleList]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleList]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
					traceFlatteningWithListNull("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
				},
			},
		},

		"legacy": {
			"values": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{"a", "b"},
				},
				Target: &tfSimpleListLegacy{},
				WantTarget: &tfSimpleListLegacy{
					Field1: types.ListValueMust(types.StringType, []attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleListLegacy]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleListLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleListLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
					traceFlatteningWithListValue("Field1", reflect.TypeFor[[]string](), 2, "Field1", reflect.TypeFor[types.List]()),
				},
			},

			"empty": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{},
				},
				Target: &tfSimpleListLegacy{},
				WantTarget: &tfSimpleListLegacy{
					Field1: types.ListValueMust(types.StringType, []attr.Value{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleListLegacy]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleListLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleListLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
					traceFlatteningWithListValue("Field1", reflect.TypeFor[[]string](), 0, "Field1", reflect.TypeFor[types.List]()),
				},
			},

			"null": {
				Source: awsSimpleStringValueSlice{
					Field1: nil,
				},
				Target: &tfSimpleListLegacy{},
				WantTarget: &tfSimpleListLegacy{
					Field1: types.ListValueMust(types.StringType, []attr.Value{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleListLegacy]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleListLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleListLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.List]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenSimpleSetOfPrimitiveValues(t *testing.T) {
	t.Parallel()

	testCases := map[string]autoFlexTestCases{
		"regular": {
			"values": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{"a", "b"},
				},
				Target: &tfSimpleSet{},
				WantTarget: &tfSimpleSet{
					Field1: types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSet]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSet]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleSet]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
					traceFlatteningWithSetValue("Field1", reflect.TypeFor[[]string](), 2, "Field1", reflect.TypeFor[types.Set]()),
				},
			},

			"empty": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{},
				},
				Target: &tfSimpleSet{},
				WantTarget: &tfSimpleSet{
					Field1: types.SetValueMust(types.StringType, []attr.Value{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSet]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSet]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleSet]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
					traceFlatteningWithSetValue("Field1", reflect.TypeFor[[]string](), 0, "Field1", reflect.TypeFor[types.Set]()),
				},
			},

			"null": {
				Source: awsSimpleStringValueSlice{
					Field1: nil,
				},
				Target: &tfSimpleSet{},
				WantTarget: &tfSimpleSet{
					Field1: types.SetNull(types.StringType),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSet]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSet]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleSet]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
					traceFlatteningWithSetNull("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
				},
			},
		},

		"legacy": {
			"values": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{"a", "b"},
				},
				Target: &tfSimpleSetLegacy{},
				WantTarget: &tfSimpleSetLegacy{
					Field1: types.SetValueMust(types.StringType, []attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSetLegacy]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSetLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleSetLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
					traceFlatteningWithSetValue("Field1", reflect.TypeFor[[]string](), 2, "Field1", reflect.TypeFor[types.Set]()),
				},
			},

			"empty": {
				Source: awsSimpleStringValueSlice{
					Field1: []string{},
				},
				Target: &tfSimpleSetLegacy{},
				WantTarget: &tfSimpleSetLegacy{
					Field1: types.SetValueMust(types.StringType, []attr.Value{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSetLegacy]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSetLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleSetLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
					traceFlatteningWithSetValue("Field1", reflect.TypeFor[[]string](), 0, "Field1", reflect.TypeFor[types.Set]()),
				},
			},

			"null": {
				Source: awsSimpleStringValueSlice{
					Field1: nil,
				},
				Target: &tfSimpleSetLegacy{},
				WantTarget: &tfSimpleSetLegacy{
					Field1: types.SetValueMust(types.StringType, []attr.Value{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSetLegacy]()),
					infoConverting(reflect.TypeFor[awsSimpleStringValueSlice](), reflect.TypeFor[*tfSimpleSetLegacy]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSimpleStringValueSlice](), "Field1", reflect.TypeFor[*tfSimpleSetLegacy]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
					debugUsingLegacyFlattener("Field1", reflect.TypeFor[[]string](), "Field1", reflect.TypeFor[types.Set]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf01]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[bool](), "Field1", reflect.TypeFor[types.Bool]()),
				traceSkipIgnoredSourceField(reflect.TypeFor[aws01](), "Tags", reflect.TypeFor[*tf01]()),
			},
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf01]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[bool](), "Field1", reflect.TypeFor[types.Bool]()),
				traceSkipIgnoredSourceField(reflect.TypeFor[aws01](), "Tags", reflect.TypeFor[*tf01]()),
			},
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf01]()),
				traceMatchedFields("Field1", reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[bool](), "Field1", reflect.TypeFor[types.Bool]()),
				traceMatchedFields("Tags", reflect.TypeFor[aws01](), "Tags", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Tags", reflect.TypeFor[map[string]string](), "Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapValue("Tags", reflect.TypeFor[map[string]string](), 1, "Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
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
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*aws01](), reflect.TypeFor[*tf01]()),
				infoConverting(reflect.TypeFor[aws01](), reflect.TypeFor[*tf01]()),
				traceSkipIgnoredSourceField(reflect.TypeFor[aws01](), "Field1", reflect.TypeFor[*tf01]()),
				traceMatchedFields("Tags", reflect.TypeFor[aws01](), "Tags", reflect.TypeFor[*tf01]()),
				infoConvertingWithPath("Tags", reflect.TypeFor[map[string]string](), "Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
				traceFlatteningWithMapValue("Tags", reflect.TypeFor[map[string]string](), 1, "Tags", reflect.TypeFor[fwtypes.MapValueOf[types.String]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenIgnoreStructTag(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"from value": {
			Source: awsSingleStringValue{
				Field1: "value1",
			},
			Target:     &tfSingleStringFieldIgnore{},
			WantTarget: &tfSingleStringFieldIgnore{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringFieldIgnore]()),
				infoConverting(reflect.TypeFor[awsSingleStringValue](), reflect.TypeFor[*tfSingleStringFieldIgnore]()),
				traceSkipIgnoredTargetField(reflect.TypeFor[awsSingleStringValue](), "Field1", reflect.TypeFor[*tfSingleStringFieldIgnore](), "Field1"),
			},
		},
		"from pointer": {
			Source: awsSingleStringPointer{
				Field1: aws.String("value1"),
			},
			Target:     &tfSingleStringFieldIgnore{},
			WantTarget: &tfSingleStringFieldIgnore{},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldIgnore]()),
				infoConverting(reflect.TypeFor[awsSingleStringPointer](), reflect.TypeFor[*tfSingleStringFieldIgnore]()),
				traceSkipIgnoredTargetField(reflect.TypeFor[awsSingleStringPointer](), "Field1", reflect.TypeFor[*tfSingleStringFieldIgnore](), "Field1"),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenInterfaceToStringTypable(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"json interface Source string Target": {
			Source: &awsJSONStringer{
				Field1: &testJSONDocument{
					Value: &struct {
						Test string `json:"test"`
					}{
						Test: "a",
					},
				},
			},
			Target: &tfSingleStringField{},
			WantTarget: &tfSingleStringField{
				Field1: types.StringValue(`{"test":"a"}`),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsJSONStringer](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsJSONStringer](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsJSONStringer](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String]()),
				// infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[testJSONDocument](), "Field1", reflect.TypeFor[types.String]()),
				infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String]()), // TODO: fix source type
			},
		},
		"null json interface Source string Target": {
			Source: &awsJSONStringer{
				Field1: nil,
			},
			Target: &tfSingleStringField{},
			WantTarget: &tfSingleStringField{
				Field1: types.StringNull(),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsJSONStringer](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsJSONStringer](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsJSONStringer](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String]()),
				// infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[testJSONDocument](), "Field1", reflect.TypeFor[types.String]()),
				infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String]()), // TODO: fix source type
				traceFlatteningNullValue("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String]()),
			},
		},

		"json interface Source JSONValue Target": {
			Source: &awsJSONStringer{
				Field1: &testJSONDocument{
					Value: &struct {
						Test string `json:"test"`
					}{
						Test: "a",
					},
				},
			},
			Target: &tfJSONStringer{},
			WantTarget: &tfJSONStringer{
				Field1: fwtypes.SmithyJSONValue(`{"test":"a"}`, newTestJSONDocument),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsJSONStringer](), reflect.TypeFor[*tfJSONStringer]()),
				infoConverting(reflect.TypeFor[awsJSONStringer](), reflect.TypeFor[*tfJSONStringer]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsJSONStringer](), "Field1", reflect.TypeFor[*tfJSONStringer]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]]()),
				// infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[testJSONDocument](), "Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]]()),
				infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]]()), // TODO: fix source type
			},
		},
		"null json interface Source JSONValue Target": {
			Source: &awsJSONStringer{
				Field1: nil,
			},
			Target: &tfJSONStringer{},
			WantTarget: &tfJSONStringer{
				Field1: fwtypes.SmithyJSONNull[smithyjson.JSONStringer](),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsJSONStringer](), reflect.TypeFor[*tfJSONStringer]()),
				infoConverting(reflect.TypeFor[awsJSONStringer](), reflect.TypeFor[*tfJSONStringer]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsJSONStringer](), "Field1", reflect.TypeFor[*tfJSONStringer]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]]()),
				// infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[testJSONDocument](), "Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]]()),
				infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]]()), // TODO: fix source type
				traceFlatteningNullValue("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[fwtypes.SmithyJSON[smithyjson.JSONStringer]]()),
			},
		},

		"json interface Source marshal error": {
			Source: &awsJSONStringer{
				Field1: &testJSONDocumentError{},
			},
			Target: &tfSingleStringField{},
			expectedDiags: diag.Diagnostics{
				diagFlatteningMarshalSmithyDocument(reflect.TypeFor[*testJSONDocumentError](), errMarshallSmithyDocument),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsJSONStringer](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsJSONStringer](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsJSONStringer](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String]()),
				// infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[testJSONDocument](), "Field1", reflect.TypeFor[types.String]()),
				infoSourceImplementsJSONStringer("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String]()), // TODO: fix source type
				errorMarshallingJSONDocument("Field1", reflect.TypeFor[smithyjson.JSONStringer](), "Field1", reflect.TypeFor[types.String](), errMarshallSmithyDocument),
			},
		},

		"non-json interface Source string Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfSingleStringField{},
			WantTarget: &tfSingleStringField{
				Field1: types.StringNull(),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[types.String]()),
				errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[types.String]()),
			},
		},

		"null non-json interface Source string Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfSingleStringField{},
			WantTarget: &tfSingleStringField{
				Field1: types.StringNull(),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfSingleStringField]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfSingleStringField]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfSingleStringField]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[types.String]()),
				errorFlatteningIncompatibleTypes("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[types.String]()),
			},
		},
	}

	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenInterface(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"nil interface Source and list Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
		"single interface Source and single list Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
		"nil interface Source and non-Flattener list Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfListNestedObject[tfSingleStringField]{},
			WantTarget: &tfListNestedObject[tfSingleStringField]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfListNestedObject[tfSingleStringField]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfListNestedObject[tfSingleStringField]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfListNestedObject[tfSingleStringField]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
			},
		},
		"single interface Source and non-Flattener list Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfListNestedObject[tfSingleStringField]{},
			WantTarget: &tfListNestedObject[tfSingleStringField]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfSingleStringField](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfListNestedObject[tfSingleStringField]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfListNestedObject[tfSingleStringField]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfListNestedObject[tfSingleStringField]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				{
					"@level":   "error",
					"@module":  "provider.autoflex",
					"@message": "AutoFlex Flatten; incompatible types",
					"from":     float64(reflect.Interface),
					"to": map[string]any{
						"ElemType": map[string]any{
							"AttrTypes": map[string]any{
								"field1": map[string]any{},
							},
						},
					},
					logAttrKeySourcePath: "Field1",
					logAttrKeySourceType: fullTypeName(reflect.TypeFor[awsInterfaceInterface]()),
					logAttrKeyTargetPath: "Field1",
					logAttrKeyTargetType: fullTypeName(reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfSingleStringField]]()),
				},
			},
		},

		"nil interface Source and set Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
		"single interface Source and single set Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
			},
		},

		"nil interface list Source and empty list Target": {
			Source: awsInterfaceSlice{
				Field1: nil,
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSlice](), "Field1", reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
				traceFlatteningWithNullValue("Field1", reflect.TypeFor[[]awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
		"empty interface list Source and empty list Target": {
			Source: awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSlice](), "Field1", reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsInterfaceInterface](), 0, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
		"non-empty interface list Source and non-empty list Target": {
			Source: awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			Target: &tfListNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfListNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSlice](), "Field1", reflect.TypeFor[*tfListNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsInterfaceInterface](), 2, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfInterfaceFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1[0]", reflect.TypeFor[awsInterfaceInterfaceImpl](), "Field1[0]", reflect.TypeFor[*tfInterfaceFlexer]()),
				infoTargetImplementsFlexFlattener("Field1[1]", reflect.TypeFor[awsInterfaceInterfaceImpl](), "Field1[1]", reflect.TypeFor[*tfInterfaceFlexer]()),
			},
		},

		"nil interface list Source and empty set Target": {
			Source: awsInterfaceSlice{
				Field1: nil,
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSlice](), "Field1", reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
				traceFlatteningWithNullValue("Field1", reflect.TypeFor[[]awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
		"empty interface list Source and empty set Target": {
			Source: awsInterfaceSlice{
				Field1: []awsInterfaceInterface{},
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSlice](), "Field1", reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsInterfaceInterface](), 0, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
		"non-empty interface list Source and non-empty set Target": {
			Source: awsInterfaceSlice{
				Field1: []awsInterfaceInterface{
					&awsInterfaceInterfaceImpl{
						AWSField: "value1",
					},
					&awsInterfaceInterfaceImpl{
						AWSField: "value2",
					},
				},
			},
			Target: &tfSetNestedObject[tfInterfaceFlexer]{},
			WantTarget: &tfSetNestedObject[tfInterfaceFlexer]{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfInterfaceFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSlice](), reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSlice](), "Field1", reflect.TypeFor[*tfSetNestedObject[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsInterfaceInterface](), 2, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfInterfaceFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1[0]", reflect.TypeFor[awsInterfaceInterfaceImpl](), "Field1[0]", reflect.TypeFor[*tfInterfaceFlexer]()),
				infoTargetImplementsFlexFlattener("Field1[1]", reflect.TypeFor[awsInterfaceInterfaceImpl](), "Field1[1]", reflect.TypeFor[*tfInterfaceFlexer]()),
			},
		},
		"nil interface Source and nested object Target": {
			Source: awsInterfaceSingle{
				Field1: nil,
			},
			Target: &tfObjectValue[tfInterfaceFlexer]{},
			WantTarget: &tfObjectValue[tfInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfNull[tfInterfaceFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfObjectValue[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfObjectValue[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfObjectValue[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
		"interface Source and nested object Target": {
			Source: awsInterfaceSingle{
				Field1: &awsInterfaceInterfaceImpl{
					AWSField: "value1",
				},
			},
			Target: &tfObjectValue[tfInterfaceFlexer]{},
			WantTarget: &tfObjectValue[tfInterfaceFlexer]{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfInterfaceFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfObjectValue[tfInterfaceFlexer]]()),
				infoConverting(reflect.TypeFor[awsInterfaceSingle](), reflect.TypeFor[*tfObjectValue[tfInterfaceFlexer]]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsInterfaceSingle](), "Field1", reflect.TypeFor[*tfObjectValue[tfInterfaceFlexer]]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfInterfaceFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsInterfaceInterface](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfInterfaceFlexer]]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenFlattener(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := autoFlexTestCases{
		"top level struct Source": {
			Source: awsExpander{
				AWSField: "value1",
			},
			Target: &tfFlexer{},
			WantTarget: &tfFlexer{
				Field1: types.StringValue("value1"),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpander](), reflect.TypeFor[*tfFlexer]()),
				infoConverting(reflect.TypeFor[awsExpander](), reflect.TypeFor[*tfFlexer]()),
				infoTargetImplementsFlexFlattener("", reflect.TypeFor[awsExpander](), "", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"top level incompatible struct Target": {
			Source: awsExpanderIncompatible{
				Incompatible: 123,
			},
			Target: &tfFlexer{},
			WantTarget: &tfFlexer{
				Field1: types.StringNull(),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderIncompatible](), reflect.TypeFor[*tfFlexer]()),
				infoConverting(reflect.TypeFor[awsExpanderIncompatible](), reflect.TypeFor[*tfFlexer]()),
				infoTargetImplementsFlexFlattener("", reflect.TypeFor[awsExpanderIncompatible](), "", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"single struct Source and single list Target": {
			Source: awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderSingleStruct](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderSingleStruct](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderSingleStruct](), "Field1", reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"nil *struct Source and null list Target": {
			Source: awsExpanderSinglePtr{
				Field1: nil,
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfNull[tfFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderSinglePtr](), "Field1", reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*awsExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
			},
		},
		"single struct Source and single set Target": {
			Source: awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderSingleStruct](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderSingleStruct](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderSingleStruct](), "Field1", reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"single *struct Source and single list Target": {
			Source: awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderSinglePtr](), "Field1", reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*awsExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"single *struct Source and single set Target": {
			Source: awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderSinglePtr](), "Field1", reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*awsExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"nil *struct Source and null set Target": {
			Source: awsExpanderSinglePtr{
				Field1: nil,
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[tfFlexer](ctx),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderSinglePtr](), "Field1", reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*awsExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
			},
		},

		"empty struct list Source and empty list Target": {
			Source: &awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsExpanderStructSlice](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderStructSlice](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderStructSlice](), "Field1", reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsExpander](), 0, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
			},
		},
		"non-empty struct list Source and non-empty list Target": {
			Source: &awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsExpanderStructSlice](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderStructSlice](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderStructSlice](), "Field1", reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsExpander](), 2, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1[0]", reflect.TypeFor[awsExpander](), "Field1[0]", reflect.TypeFor[*tfFlexer]()),
				infoTargetImplementsFlexFlattener("Field1[1]", reflect.TypeFor[awsExpander](), "Field1[1]", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"empty *struct list Source and empty list Target": {
			Source: &awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsExpanderPtrSlice](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderPtrSlice](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderPtrSlice](), "Field1", reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsExpander](), 0, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
			},
		},
		"non-empty *struct list Source and non-empty list Target": {
			Source: &awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &tfExpanderListNestedObject{},
			WantTarget: &tfExpanderListNestedObject{
				Field1: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsExpanderPtrSlice](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderPtrSlice](), reflect.TypeFor[*tfExpanderListNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderPtrSlice](), "Field1", reflect.TypeFor[*tfExpanderListNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsExpander](), "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsExpander](), 2, "Field1", reflect.TypeFor[fwtypes.ListNestedObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1[0]", reflect.TypeFor[awsExpander](), "Field1[0]", reflect.TypeFor[*tfFlexer]()),
				infoTargetImplementsFlexFlattener("Field1[1]", reflect.TypeFor[awsExpander](), "Field1[1]", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"empty struct list Source and empty set Target": {
			Source: awsExpanderStructSlice{
				Field1: []awsExpander{},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderStructSlice](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderStructSlice](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderStructSlice](), "Field1", reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsExpander](), 0, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
			},
		},
		"non-empty struct list Source and set Target": {
			Source: awsExpanderStructSlice{
				Field1: []awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderStructSlice](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderStructSlice](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderStructSlice](), "Field1", reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]awsExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]awsExpander](), 2, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1[0]", reflect.TypeFor[awsExpander](), "Field1[0]", reflect.TypeFor[*tfFlexer]()),
				infoTargetImplementsFlexFlattener("Field1[1]", reflect.TypeFor[awsExpander](), "Field1[1]", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"empty *struct list Source and empty set Target": {
			Source: awsExpanderPtrSlice{
				Field1: []*awsExpander{},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderPtrSlice](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderPtrSlice](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderPtrSlice](), "Field1", reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsExpander](), 0, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
			},
		},
		"non-empty *struct list Source and non-empty set Target": {
			Source: awsExpanderPtrSlice{
				Field1: []*awsExpander{
					{
						AWSField: "value1",
					},
					{
						AWSField: "value2",
					},
				},
			},
			Target: &tfExpanderSetNestedObject{},
			WantTarget: &tfExpanderSetNestedObject{
				Field1: fwtypes.NewSetNestedObjectValueOfValueSliceMust(ctx, []tfFlexer{
					{
						Field1: types.StringValue("value1"),
					},
					{
						Field1: types.StringValue("value2"),
					},
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderPtrSlice](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConverting(reflect.TypeFor[awsExpanderPtrSlice](), reflect.TypeFor[*tfExpanderSetNestedObject]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderPtrSlice](), "Field1", reflect.TypeFor[*tfExpanderSetNestedObject]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[[]*awsExpander](), "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
				traceFlatteningNestedObjectCollection("Field1", reflect.TypeFor[[]*awsExpander](), 2, "Field1", reflect.TypeFor[fwtypes.SetNestedObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1[0]", reflect.TypeFor[awsExpander](), "Field1[0]", reflect.TypeFor[*tfFlexer]()),
				infoTargetImplementsFlexFlattener("Field1[1]", reflect.TypeFor[awsExpander](), "Field1[1]", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"struct Source and object value Target": {
			Source: awsExpanderSingleStruct{
				Field1: awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderObjectValue{},
			WantTarget: &tfExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderSingleStruct](), reflect.TypeFor[*tfExpanderObjectValue]()),
				infoConverting(reflect.TypeFor[awsExpanderSingleStruct](), reflect.TypeFor[*tfExpanderObjectValue]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderSingleStruct](), "Field1", reflect.TypeFor[*tfExpanderObjectValue]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[*tfFlexer]()),
			},
		},
		"*struct Source and object value Target": {
			Source: awsExpanderSinglePtr{
				Field1: &awsExpander{
					AWSField: "value1",
				},
			},
			Target: &tfExpanderObjectValue{},
			WantTarget: &tfExpanderObjectValue{
				Field1: fwtypes.NewObjectValueOfMust(ctx, &tfFlexer{
					Field1: types.StringValue("value1"),
				}),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderObjectValue]()),
				infoConverting(reflect.TypeFor[awsExpanderSinglePtr](), reflect.TypeFor[*tfExpanderObjectValue]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsExpanderSinglePtr](), "Field1", reflect.TypeFor[*tfExpanderObjectValue]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[*awsExpander](), "Field1", reflect.TypeFor[fwtypes.ObjectValueOf[tfFlexer]]()),
				infoTargetImplementsFlexFlattener("Field1", reflect.TypeFor[awsExpander](), "Field1", reflect.TypeFor[*tfFlexer]()),
			},
		},
	}
	runAutoFlattenTestCases(t, testCases)
}

func TestFlattenStructListOfStringEnum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"struct with list of string enum": {
			"valid value": {
				Source: awsSliceOfStringEnum{
					Field1: []testEnum{testEnumScalar, testEnumList},
				},
				Target: &tfListOfStringEnum{},
				WantTarget: &tfListOfStringEnum{
					Field1: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
						fwtypes.StringEnumValue(testEnumScalar),
						fwtypes.StringEnumValue(testEnumList),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfListOfStringEnum]()),
					infoConverting(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfListOfStringEnum]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfStringEnum](), "Field1", reflect.TypeFor[*tfListOfStringEnum]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]testEnum](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]]()),
					traceFlatteningWithListValue("Field1", reflect.TypeFor[[]testEnum](), 2, "Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]]()),
				},
			},
			"empty value": {
				Source: awsSliceOfStringEnum{
					Field1: []testEnum{},
				},
				Target: &tfListOfStringEnum{},
				WantTarget: &tfListOfStringEnum{
					Field1: fwtypes.NewListValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfListOfStringEnum]()),
					infoConverting(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfListOfStringEnum]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfStringEnum](), "Field1", reflect.TypeFor[*tfListOfStringEnum]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]testEnum](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]]()),
					traceFlatteningWithListValue("Field1", reflect.TypeFor[[]testEnum](), 0, "Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]]()),
				},
			},
			"null value": {
				Source: awsSliceOfStringEnum{},
				Target: &tfListOfStringEnum{},
				WantTarget: &tfListOfStringEnum{
					Field1: fwtypes.NewListValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfListOfStringEnum]()),
					infoConverting(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfListOfStringEnum]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfStringEnum](), "Field1", reflect.TypeFor[*tfListOfStringEnum]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]testEnum](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]]()),
					traceFlatteningWithListNull("Field1", reflect.TypeFor[[]testEnum](), "Field1", reflect.TypeFor[fwtypes.ListValueOf[fwtypes.StringEnum[testEnum]]]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenStructSetOfStringEnum(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	testCases := map[string]autoFlexTestCases{
		"struct with set of string enum": {
			"valid value": {
				Source: awsSliceOfStringEnum{
					Field1: []testEnum{testEnumScalar, testEnumList},
				},
				Target: &tfSetOfStringEnum{},
				WantTarget: &tfSetOfStringEnum{
					Field1: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{
						fwtypes.StringEnumValue(testEnumScalar),
						fwtypes.StringEnumValue(testEnumList),
					}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfSetOfStringEnum]()),
					infoConverting(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfSetOfStringEnum]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfStringEnum](), "Field1", reflect.TypeFor[*tfSetOfStringEnum]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]testEnum](), "Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]]()),
					traceFlatteningWithSetValue("Field1", reflect.TypeFor[[]testEnum](), 2, "Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]]()),
				},
			},
			"empty value": {
				Source: awsSliceOfStringEnum{
					Field1: []testEnum{},
				},
				Target: &tfSetOfStringEnum{},
				WantTarget: &tfSetOfStringEnum{
					Field1: fwtypes.NewSetValueOfMust[fwtypes.StringEnum[testEnum]](ctx, []attr.Value{}),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfSetOfStringEnum]()),
					infoConverting(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfSetOfStringEnum]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfStringEnum](), "Field1", reflect.TypeFor[*tfSetOfStringEnum]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]testEnum](), "Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]]()),
					traceFlatteningWithSetValue("Field1", reflect.TypeFor[[]testEnum](), 0, "Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]]()),
				},
			},
			"null value": {
				Source: awsSliceOfStringEnum{},
				Target: &tfSetOfStringEnum{},
				WantTarget: &tfSetOfStringEnum{
					Field1: fwtypes.NewSetValueOfNull[fwtypes.StringEnum[testEnum]](ctx),
				},
				expectedLogLines: []map[string]any{
					infoFlattening(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfSetOfStringEnum]()),
					infoConverting(reflect.TypeFor[awsSliceOfStringEnum](), reflect.TypeFor[*tfSetOfStringEnum]()),
					traceMatchedFields("Field1", reflect.TypeFor[awsSliceOfStringEnum](), "Field1", reflect.TypeFor[*tfSetOfStringEnum]()),
					infoConvertingWithPath("Field1", reflect.TypeFor[[]testEnum](), "Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]]()),
					traceFlatteningWithSetNull("Field1", reflect.TypeFor[[]testEnum](), "Field1", reflect.TypeFor[fwtypes.SetValueOf[fwtypes.StringEnum[testEnum]]]()),
				},
			},
		},
	}

	for testName, cases := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			runAutoFlattenTestCases(t, cases)
		})
	}
}

func TestFlattenEmbeddedStruct(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"exported": {
			Source: &awsEmbeddedStruct{
				Field1: "a",
				Field2: "b",
			},
			Target: &tfExportedEmbeddedStruct{},
			WantTarget: &tfExportedEmbeddedStruct{
				TFExportedStruct: TFExportedStruct{
					Field1: types.StringValue("a"),
				},
				Field2: types.StringValue("b"),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsEmbeddedStruct](), reflect.TypeFor[*tfExportedEmbeddedStruct]()),
				infoConverting(reflect.TypeFor[awsEmbeddedStruct](), reflect.TypeFor[*tfExportedEmbeddedStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsEmbeddedStruct](), "Field1", reflect.TypeFor[*tfExportedEmbeddedStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				traceMatchedFields("Field2", reflect.TypeFor[awsEmbeddedStruct](), "Field2", reflect.TypeFor[*tfExportedEmbeddedStruct]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[string](), "Field2", reflect.TypeFor[types.String]()),
			},
		},
		"unexported": {
			Source: &awsEmbeddedStruct{
				Field1: "a",
				Field2: "b",
			},
			Target: &tfUnexportedEmbeddedStruct{},
			WantTarget: &tfUnexportedEmbeddedStruct{
				tfSingleStringField: tfSingleStringField{
					Field1: types.StringValue("a"),
				},
				Field2: types.StringValue("b"),
			},
			expectedLogLines: []map[string]any{
				infoFlattening(reflect.TypeFor[*awsEmbeddedStruct](), reflect.TypeFor[*tfUnexportedEmbeddedStruct]()),
				infoConverting(reflect.TypeFor[awsEmbeddedStruct](), reflect.TypeFor[*tfUnexportedEmbeddedStruct]()),
				traceMatchedFields("Field1", reflect.TypeFor[awsEmbeddedStruct](), "Field1", reflect.TypeFor[*tfUnexportedEmbeddedStruct]()),
				infoConvertingWithPath("Field1", reflect.TypeFor[string](), "Field1", reflect.TypeFor[types.String]()),
				traceMatchedFields("Field2", reflect.TypeFor[awsEmbeddedStruct](), "Field2", reflect.TypeFor[*tfUnexportedEmbeddedStruct]()),
				infoConvertingWithPath("Field2", reflect.TypeFor[string](), "Field2", reflect.TypeFor[types.String]()),
			},
		},
	}
	// cmp.Diff cannot handle an unexported field
	runAutoFlattenTestCases(t, testCases, cmpopts.EquateComparable(tfUnexportedEmbeddedStruct{}))
}

func runAutoFlattenTestCases(t *testing.T, testCases autoFlexTestCases, opts ...cmp.Option) {
	t.Helper()

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var buf bytes.Buffer
			ctx = tflogtest.RootLogger(ctx, &buf)

			ctx = registerTestingLogger(ctx)

			diags := Flatten(ctx, testCase.Source, testCase.Target, testCase.Options...)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			lines, err := tflogtest.MultilineJSONDecode(&buf)
			if err != nil {
				t.Fatalf("Flatten: decoding log lines: %s", err)
			}
			if diff := cmp.Diff(lines, testCase.expectedLogLines); diff != "" {
				t.Errorf("unexpected log lines diff (+wanted, -got): %s", diff)
			}

			if !diags.HasError() {
				less := func(a, b any) bool { return fmt.Sprintf("%+v", a) < fmt.Sprintf("%+v", b) }
				if diff := cmp.Diff(testCase.Target, testCase.WantTarget, append(opts, cmpopts.SortSlices(less))...); diff != "" {
					if !testCase.WantDiff {
						t.Errorf("unexpected diff (+wanted, -got): %s", diff)
					}
				}
			}
		})
	}
}

// Top-level tests need a concrete target type for some reason when calling `cmp.Diff`
type toplevelTestCase[Tsource, Ttarget any] struct {
	source           Tsource
	expectedValue    Ttarget
	expectedDiags    diag.Diagnostics
	expectedLogLines []map[string]any
}

type toplevelTestCases[Tsource, Ttarget any] map[string]toplevelTestCase[Tsource, Ttarget]

func runTopLevelTestCases[Tsource, Ttarget any](t *testing.T, testCases toplevelTestCases[Tsource, Ttarget]) {
	t.Helper()

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			var buf bytes.Buffer
			ctx = tflogtest.RootLogger(ctx, &buf)

			ctx = registerTestingLogger(ctx)

			var target Ttarget
			diags := Flatten(ctx, testCase.source, &target)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}

			lines, err := tflogtest.MultilineJSONDecode(&buf)
			if err != nil {
				t.Fatalf("Flatten: decoding log lines: %s", err)
			}
			if diff := cmp.Diff(lines, testCase.expectedLogLines); diff != "" {
				t.Errorf("unexpected log lines diff (+wanted, -got): %s", diff)
			}

			if !diags.HasError() {
				less := func(a, b any) bool { return fmt.Sprintf("%+v", a) < fmt.Sprintf("%+v", b) }
				if diff := cmp.Diff(target, testCase.expectedValue, cmpopts.SortSlices(less)); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}

func TestFlattenPrePopulate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testCases := map[string]struct {
		target   any
		expected any
	}{
		"string": {
			target: &rootStringModel{},
			expected: &rootStringModel{
				Field1: types.StringNull(),
			},
		},

		"nested list": {
			target: &rootListNestedObjectModel{},
			expected: &rootListNestedObjectModel{
				Field1: fwtypes.NewListNestedObjectValueOfNull[nestedModel](ctx),
			},
		},

		"nested set": {
			target: &rootSetNestedObjectModel{},
			expected: &rootSetNestedObjectModel{
				Field1: fwtypes.NewSetNestedObjectValueOfNull[nestedModel](ctx),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			valTo := reflect.ValueOf(testCase.target)

			diags := flattenPrePopulate(ctx, valTo)

			if l := len(diags); l > 0 {
				t.Fatalf("expected 0 diags, got %s", fwdiag.DiagnosticsString(diags))
			}

			if diff := cmp.Diff(testCase.target, testCase.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

type rootStringModel struct {
	Field1 types.String `tfsdk:"field1"`
}

type rootListNestedObjectModel struct {
	Field1 fwtypes.ListNestedObjectValueOf[nestedModel] `tfsdk:"field1"`
}

type rootSetNestedObjectModel struct {
	Field1 fwtypes.SetNestedObjectValueOf[nestedModel] `tfsdk:"field1"`
}

type nestedModel struct {
	Field1 types.String `tfsdk:"field1"`
}
