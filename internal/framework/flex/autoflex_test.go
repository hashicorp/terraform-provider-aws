// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	smithydocument "github.com/aws/smithy-go/document"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

type emptyStruct struct{}

type tfSingleStringField struct {
	Field1 types.String `tfsdk:"field1"`
}

type tfSingleFloat64Field struct {
	Field1 types.Float64 `tfsdk:"field1"`
}

type tfSingleInt64Field struct {
	Field1 types.Int64 `tfsdk:"field1"`
}

type tfSingleInt32Field struct {
	Field1 types.Int32 `tfsdk:"field1"`
}

// All primitive types.
type tfAllThePrimitiveFields struct {
	Field1  types.String  `tfsdk:"field1"`
	Field2  types.String  `tfsdk:"field2"`
	Field3  types.Int64   `tfsdk:"field3"`
	Field4  types.Int64   `tfsdk:"field4"`
	Field5  types.Int64   `tfsdk:"field5"`
	Field6  types.Int64   `tfsdk:"field6"`
	Field7  types.Float64 `tfsdk:"field7"`
	Field8  types.Float64 `tfsdk:"field8"`
	Field9  types.Float64 `tfsdk:"field9"`
	Field10 types.Float64 `tfsdk:"field10"`
	Field11 types.Bool    `tfsdk:"field11"`
	Field12 types.Bool    `tfsdk:"field12"`
}

type awsAllThePrimitiveFields struct {
	Field1  string
	Field2  *string
	Field3  int32
	Field4  *int32
	Field5  int64
	Field6  *int64
	Field7  float32
	Field8  *float32
	Field9  float64
	Field10 *float64
	Field11 bool
	Field12 *bool
}

// List/Set/Map of primitive types.
type tfCollectionsOfPrimitiveElements struct {
	Field1 types.List `tfsdk:"field1"`
	Field2 types.List `tfsdk:"field2"`
	Field3 types.Set  `tfsdk:"field3"`
	Field4 types.Set  `tfsdk:"field4"`
	Field5 types.Map  `tfsdk:"field5"`
	Field6 types.Map  `tfsdk:"field6"`
}

// List/Set/Map of string types.
type tfTypedCollectionsOfPrimitiveElements struct {
	Field1 fwtypes.ListValueOf[types.String] `tfsdk:"field1"`
	Field2 fwtypes.ListValueOf[types.String] `tfsdk:"field2"`
	Field3 fwtypes.SetValueOf[types.String]  `tfsdk:"field3"`
	Field4 fwtypes.SetValueOf[types.String]  `tfsdk:"field4"`
	Field5 fwtypes.MapValueOf[types.String]  `tfsdk:"field5"`
	Field6 fwtypes.MapValueOf[types.String]  `tfsdk:"field6"`
}

type awsCollectionsOfPrimitiveElements struct {
	Field1 []string
	Field2 []*string
	Field3 []string
	Field4 []*string
	Field5 map[string]string
	Field6 map[string]*string
}

type tfListOfNestedObject struct {
	Field1 fwtypes.ListNestedObjectValueOf[tfSingleStringField] `tfsdk:"field1"`
}

type tfSetOfNestedObject struct {
	Field1 fwtypes.SetNestedObjectValueOf[tfSingleStringField] `tfsdk:"field1"`
}

type tfComplexValue struct {
	Field1 types.String                                          `tfsdk:"field1"`
	Field2 fwtypes.ListNestedObjectValueOf[tfListOfNestedObject] `tfsdk:"field2"`
	Field3 types.Map                                             `tfsdk:"field3"`
	Field4 fwtypes.SetNestedObjectValueOf[tfSingleInt64Field]    `tfsdk:"field4"`
}

type awsComplexValue struct {
	Field1 string
	Field2 *awsNestedObjectPointer
	Field3 map[string]*string
	Field4 []awsSingleInt64Value
}

// tfSingluarListOfNestedObjects testing for idiomatic singular on TF side but plural on AWS side
type tfSingluarListOfNestedObjects struct {
	Field fwtypes.ListNestedObjectValueOf[tfSingleStringField] `tfsdk:"field"`
}

type awsPluralSliceOfNestedObjectValues struct {
	Fields []awsSingleStringValue
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

type awsSingleStringValue struct {
	Field1 string
}

type awsSingleStringPointer struct {
	Field1 *string
}

type awsSingleFloat64Value struct {
	Field1 float64
}

type awsSingleFloat64Pointer struct {
	Field1 *float64
}

type awsSingleFloat32Value struct {
	Field1 float32
}

type awsSingleFloat32Pointer struct {
	Field1 *float32
}

type awsSingleInt64Value struct {
	Field1 int64
}

type awsSingleInt64Pointer struct {
	Field1 *int64
}

type awsSingleInt32Value struct {
	Field1 int32
}

type awsSingleInt32Pointer struct {
	Field1 *int32
}

type awsNestedObjectPointer struct {
	Field1 *awsSingleStringValue
}

type awsSliceOfNestedObjectPointers struct {
	Field1 []*awsSingleStringValue
}

type awsSliceOfNestedObjectValues struct {
	Field1 []awsSingleStringValue
}

// tfFieldNamePrefix has no prefix to test matching on prefix
type tfFieldNamePrefix struct {
	Name types.String `tfsdk:"name"`
}

// awsFieldNamePrefix has prefix to test matching on prefix
type awsFieldNamePrefix struct {
	IntentName *string
}

// tfFieldNameSuffix has no suffix to test matching on suffix
type tfFieldNameSuffix struct {
	Policy types.String `tfsdk:"policy"`
}

// awsFieldNameSuffix has suffix to test matching on suffix
type awsFieldNameSuffix struct {
	PolicyConfig *string
}

type tfRFC3339Time struct {
	CreationDateTime timetypes.RFC3339 `tfsdk:"creation_date_time"`
}

type awsRFC3339TimePointer struct {
	CreationDateTime *time.Time
}

type awsRFC3339TimeValue struct {
	CreationDateTime time.Time
}

type tfMapOfString struct {
	FieldInner fwtypes.MapValueOf[basetypes.StringValue] `tfsdk:"field_inner"`
}

type tfNestedMapOfString struct {
	FieldOuter fwtypes.ListNestedObjectValueOf[tfMapOfString] `tfsdk:"field_outer"`
}

type awsNestedMapOfString struct {
	FieldOuter awsMapOfString
}

type tfMapOfMapOfString struct {
	Field1 fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]] `tfsdk:"field1"`
}

type awsMapOfString struct {
	FieldInner map[string]string
}

type awsMapOfMapOfString struct {
	Field1 map[string]map[string]string
}

type awsMapOfMapOfStringPointer struct {
	Field1 map[string]map[string]*string
}

type awsMapOfStringPointer struct {
	FieldInner map[string]*string
}

type testEnum string

// Enum values for SlotShape
const (
	testEnumScalar testEnum = "Scalar"
	testEnumList   testEnum = "List"
)

func (testEnum) Values() []testEnum {
	return []testEnum{
		testEnumScalar,
		testEnumList,
	}
}

type tfPluralAndSingularFields struct {
	Value types.String `tfsdk:"Value"`
}

type awsPluralAndSingularFields struct {
	Value  string
	Values string
}

type tfSingleARNField struct {
	Field1 fwtypes.ARN `tfsdk:"field1"`
}

type tfMapBlockList struct {
	MapBlock fwtypes.ListNestedObjectValueOf[tfMapBlockElement] `tfsdk:"map_block"`
}

type tfMapBlockSet struct {
	MapBlock fwtypes.SetNestedObjectValueOf[tfMapBlockElement] `tfsdk:"map_block"`
}

type awsMapBlockValues struct {
	MapBlock map[string]awsMapBlockElement
}

type awsMapBlockPointers struct {
	MapBlock map[string]*awsMapBlockElement
}

type tfMapBlockElement struct {
	MapBlockKey types.String `tfsdk:"map_block_key"`
	Attr1       types.String `tfsdk:"attr1"`
	Attr2       types.String `tfsdk:"attr2"`
}

type awsMapBlockElement struct {
	Attr1 string
	Attr2 string
}

type tfMapBlockListEnumKey struct {
	MapBlock fwtypes.ListNestedObjectValueOf[tfMapBlockElementEnumKey] `tfsdk:"map_block"`
}

type tfMapBlockElementEnumKey struct {
	MapBlockKey fwtypes.StringEnum[testEnum] `tfsdk:"map_block_key"`
	Attr1       types.String                 `tfsdk:"attr1"`
	Attr2       types.String                 `tfsdk:"attr2"`
}

type tfMapBlockListNoKey struct {
	MapBlock fwtypes.ListNestedObjectValueOf[tfMapBlockElementNoKey] `tfsdk:"map_block"`
}

type tfMapBlockElementNoKey struct {
	Attr1 types.String `tfsdk:"attr1"`
	Attr2 types.String `tfsdk:"attr2"`
}

var _ smithyjson.JSONStringer = (*testJSONDocument)(nil)
var _ smithydocument.Marshaler = (*testJSONDocument)(nil)

type testJSONDocument struct {
	Value any
}

func newTestJSONDocument(v any) smithyjson.JSONStringer {
	return &testJSONDocument{Value: v}
}

func (m *testJSONDocument) UnmarshalSmithyDocument(v interface{}) error {
	data, err := json.Marshal(m.Value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (m *testJSONDocument) MarshalSmithyDocument() ([]byte, error) {
	return json.Marshal(m.Value)
}

var _ smithyjson.JSONStringer = &testJSONDocumentError{}

type testJSONDocumentError struct{}

func (m *testJSONDocumentError) UnmarshalSmithyDocument(v interface{}) error {
	return errUnmarshallSmithyDocument
}

func (m *testJSONDocumentError) MarshalSmithyDocument() ([]byte, error) {
	return nil, errMarshallSmithyDocument
}

var (
	errUnmarshallSmithyDocument = errors.New("test unmarshal error")
	errMarshallSmithyDocument   = errors.New("test marshal error")
)

type awsJSONStringer struct {
	Field1 smithyjson.JSONStringer `json:"field1"`
}

type tfJSONStringer struct {
	Field1 fwtypes.SmithyJSON[smithyjson.JSONStringer] `tfsdk:"field1"`
}

type testFlexTFListNestedObject[T any] struct {
	Field1 fwtypes.ListNestedObjectValueOf[T] `tfsdk:"field1"`
}

type testFlexTFSetNestedObject[T any] struct {
	Field1 fwtypes.SetNestedObjectValueOf[T] `tfsdk:"field1"`
}

type testFlexTFObjectValue[T any] struct {
	Field1 fwtypes.ObjectValueOf[T] `tfsdk:"field1"`
}

type testFlexTFInterfaceFlexer struct {
	Field1 types.String `tfsdk:"field1"`
}

var (
	_ Expander  = testFlexTFInterfaceFlexer{}
	_ Flattener = &testFlexTFInterfaceFlexer{}
)

func (t testFlexTFInterfaceFlexer) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &testFlexAWSInterfaceInterfaceImpl{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

func (t *testFlexTFInterfaceFlexer) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch val := v.(type) {
	case testFlexAWSInterfaceInterfaceImpl:
		t.Field1 = StringValueToFramework(ctx, val.AWSField)
		return diags

	default:
		return diags
	}
}

type testFlexTFInterfaceIncompatibleExpander struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ Expander = testFlexTFInterfaceIncompatibleExpander{}

func (t testFlexTFInterfaceIncompatibleExpander) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &testFlexAWSInterfaceIncompatibleImpl{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

type testFlexAWSInterfaceIncompatibleImpl struct {
	AWSField string
}

type testFlexAWSInterfaceSingle struct {
	Field1 testFlexAWSInterfaceInterface
}

type testFlexAWSInterfaceSlice struct {
	Field1 []testFlexAWSInterfaceInterface
}

type testFlexAWSInterfaceInterface interface {
	isTestFlexAWSInterfaceInterface()
}

type testFlexAWSInterfaceInterfaceImpl struct {
	AWSField string
}

var _ testFlexAWSInterfaceInterface = &testFlexAWSInterfaceInterfaceImpl{}

func (t *testFlexAWSInterfaceInterfaceImpl) isTestFlexAWSInterfaceInterface() {} // nosemgrep:ci.aws-in-func-name

type testFlexTFFlexer struct {
	Field1 types.String `tfsdk:"field1"`
}

var (
	_ Expander  = testFlexTFFlexer{}
	_ Flattener = &testFlexTFFlexer{}
)

func (t testFlexTFFlexer) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return &testFlexAWSExpander{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

func (t *testFlexTFFlexer) Flatten(ctx context.Context, v any) (diags diag.Diagnostics) {
	switch val := v.(type) {
	case testFlexAWSExpander:
		t.Field1 = StringValueToFramework(ctx, val.AWSField)
		return diags

	default:
		return diags
	}
}

type testFlexTFExpanderListNestedObject testFlexTFListNestedObject[testFlexTFFlexer]

type testFlexTFExpanderSetNestedObject testFlexTFSetNestedObject[testFlexTFFlexer]

type testFlexTFExpanderObjectValue testFlexTFObjectValue[testFlexTFFlexer]

type testFlexTFTypedExpanderListNestedObject testFlexTFListNestedObject[testFlexTFTypedExpander]

type testFlexTFTypedExpanderSetNestedObject testFlexTFSetNestedObject[testFlexTFTypedExpander]

type testFlexTFTypedExpanderObjectValue testFlexTFObjectValue[testFlexTFTypedExpander]

type testFlexTFExpanderToString struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ Expander = testFlexTFExpanderToString{}

func (t testFlexTFExpanderToString) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return StringValueFromFramework(ctx, t.Field1), nil
}

type testFlexTFExpanderToNil struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ Expander = testFlexTFExpanderToNil{}

func (t testFlexTFExpanderToNil) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return nil, nil
}

type testFlexTFTypedExpander struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ TypedExpander = testFlexTFTypedExpander{}

func (t testFlexTFTypedExpander) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	return &testFlexAWSExpander{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

type testFlexTFTypedExpanderToNil struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ TypedExpander = testFlexTFTypedExpanderToNil{}

func (t testFlexTFTypedExpanderToNil) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	return nil, nil
}

type testFlexTFInterfaceTypedExpander struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ TypedExpander = testFlexTFInterfaceTypedExpander{}

func (t testFlexTFInterfaceTypedExpander) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	switch targetType {
	case reflect.TypeFor[testFlexAWSInterfaceInterface]():
		return &testFlexAWSInterfaceInterfaceImpl{
			AWSField: StringValueFromFramework(ctx, t.Field1),
		}, nil
	}

	return nil, nil
}

type testFlexTFInterfaceIncompatibleTypedExpander struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ TypedExpander = testFlexTFInterfaceIncompatibleTypedExpander{}

func (t testFlexTFInterfaceIncompatibleTypedExpander) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	return &testFlexAWSInterfaceIncompatibleImpl{
		AWSField: StringValueFromFramework(ctx, t.Field1),
	}, nil
}

type testFlexAWSExpander struct {
	AWSField string
}

type testFlexAWSExpanderIncompatible struct {
	Incompatible int
}

type testFlexAWSExpanderSingleStruct struct {
	Field1 testFlexAWSExpander
}

type testFlexAWSExpanderSinglePtr struct {
	Field1 *testFlexAWSExpander
}

type testFlexAWSExpanderStructSlice struct {
	Field1 []testFlexAWSExpander
}

type testFlexAWSExpanderPtrSlice struct {
	Field1 []*testFlexAWSExpander
}
