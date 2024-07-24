// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"encoding/json"
	"time"

	smithydocument "github.com/aws/smithy-go/document"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

type TestFlex00 struct{}

type TestFlexTF01 struct {
	Field1 types.String `tfsdk:"field1"`
}

type TestFlexTF02 struct {
	Field1 types.Int64 `tfsdk:"field1"`
}

// All primitive types.
type TestFlexTF03 struct {
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

// List/Set/Map of primitive types.
type TestFlexTF04 struct {
	Field1 types.List `tfsdk:"field1"`
	Field2 types.List `tfsdk:"field2"`
	Field3 types.Set  `tfsdk:"field3"`
	Field4 types.Set  `tfsdk:"field4"`
	Field5 types.Map  `tfsdk:"field5"`
	Field6 types.Map  `tfsdk:"field6"`
}

type TestFlexTF05 struct {
	Field1 fwtypes.ListNestedObjectValueOf[TestFlexTF01] `tfsdk:"field1"`
}

type TestFlexTF06 struct {
	Field1 fwtypes.SetNestedObjectValueOf[TestFlexTF01] `tfsdk:"field1"`
}

type TestFlexTF07 struct {
	Field1 types.String                                  `tfsdk:"field1"`
	Field2 fwtypes.ListNestedObjectValueOf[TestFlexTF05] `tfsdk:"field2"`
	Field3 types.Map                                     `tfsdk:"field3"`
	Field4 fwtypes.SetNestedObjectValueOf[TestFlexTF02]  `tfsdk:"field4"`
}

// TestFlexTF08 testing for idiomatic singular on TF side but plural on AWS side
type TestFlexTF08 struct {
	Field fwtypes.ListNestedObjectValueOf[TestFlexTF01] `tfsdk:"field"`
}

type TestFlexTF09 struct {
	City      types.List `tfsdk:"city"`
	Coach     types.List `tfsdk:"coach"`
	Tomato    types.List `tfsdk:"tomato"`
	Vertex    types.List `tfsdk:"vertex"`
	Criterion types.List `tfsdk:"criterion"`
	Datum     types.List `tfsdk:"datum"`
	Hive      types.List `tfsdk:"hive"`
}

// TestFlexTF10 testing for fields that only differ by capitalization
type TestFlexTF10 struct {
	FieldURL types.String `tfsdk:"field_url"`
}

type TestFlexAWS01 struct {
	Field1 string
}

type TestFlexAWS02 struct {
	Field1 *string
}

type TestFlexAWS03 struct {
	Field1 int64
}

type TestFlexAWS04 struct {
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

type TestFlexAWS05 struct {
	Field1 []string
	Field2 []*string
	Field3 []string
	Field4 []*string
	Field5 map[string]string
	Field6 map[string]*string
}

type TestFlexAWS06 struct {
	Field1 *TestFlexAWS01
}

type TestFlexAWS07 struct {
	Field1 []*TestFlexAWS01
}

type TestFlexAWS08 struct {
	Field1 []TestFlexAWS01
}

type TestFlexAWS09 struct {
	Field1 string
	Field2 *TestFlexAWS06
	Field3 map[string]*string
	Field4 []TestFlexAWS03
}

type TestFlexAWS10 struct {
	Fields []TestFlexAWS01
}

type TestFlexAWS11 struct {
	Cities   []*string
	Coaches  []*string
	Tomatoes []*string
	Vertices []*string
	Criteria []*string
	Data     []*string
	Hives    []*string
}

type TestFlexAWS12 struct {
	FieldUrl *string
}

type TestFlexTF16 struct {
	Name types.String `tfsdk:"name"`
}

type TestFlexAWS18 struct {
	IntentName *string
}

type TestFlexTimeTF01 struct {
	CreationDateTime timetypes.RFC3339 `tfsdk:"creation_date_time"`
}
type TestFlexTimeAWS01 struct {
	CreationDateTime *time.Time
}
type TestFlexTimeAWS02 struct {
	CreationDateTime time.Time
}

type TestFlexTF11 struct {
	FieldInner fwtypes.MapValueOf[basetypes.StringValue] `tfsdk:"field_inner"`
}

type TestFlexTF14 struct {
	FieldOuter fwtypes.ListNestedObjectValueOf[TestFlexTF11] `tfsdk:"field_outer"`
}

type TestFlexAWS13 struct {
	FieldInner map[string]string
}

type TestFlexAWS14 struct {
	FieldInner map[string]TestFlexAWS01
}

type TestFlexAWS15 struct {
	FieldInner map[string]*TestFlexAWS01
}

type TestFlexAWS16 struct {
	FieldOuter TestFlexAWS13
}

type TestFlexAWS17 struct {
	FieldOuter TestFlexAWS14
}

type TestEnum string

// Enum values for SlotShape
const (
	TestEnumScalar TestEnum = "Scalar"
	TestEnumList   TestEnum = "List"
)

func (TestEnum) Values() []TestEnum {
	return []TestEnum{
		"Scalar",
		"List",
	}
}

type TestFlexComplexNestAWS01 struct { // ie, DialogState
	DialogAction      *TestFlexComplexNestAWS02
	Intent            *TestFlexComplexNestAWS03
	SessionAttributes map[string]string
}

type TestFlexComplexNestTF02 struct { // ie, DialogAction
	Type                fwtypes.StringEnum[TestEnum] `tfsdk:"type"`
	SlotToElicit        types.String                 `tfsdk:"slot_to_elicit"`
	SuppressNextMessage types.Bool                   `tfsdk:"suppress_next_message"`
}
type TestFlexComplexNestAWS02 struct { // ie, DialogAction
	Type                TestEnum
	SlotToElicit        *string
	SuppressNextMessage *bool
}

type TestFlexComplexNestAWS03 struct { // ie, IntentOverride
	Name  *string
	Slots map[string]TestFlexComplexNestAWS04
}

type TestFlexComplexNestTF04 struct { // ie, TestFlexComplexNestAWS04
	Shape fwtypes.StringEnum[TestEnum]                             `tfsdk:"shape"`
	Value fwtypes.ListNestedObjectValueOf[TestFlexComplexNestTF05] `tfsdk:"value"`
}
type TestFlexComplexNestAWS04 struct { // ie, SlotValueOverride
	Shape  TestEnum
	Value  *TestFlexComplexNestAWS05
	Values []TestFlexComplexNestAWS04 // recursive type
}

type TestFlexComplexNestTF05 struct { // ie, SlotValue
	InterpretedValue types.String `tfsdk:"interpreted_value"`
}
type TestFlexComplexNestAWS05 struct { // ie, SlotValue
	InterpretedValue *string
}

type TestFlexPluralityTF01 struct {
	Value types.String `tfsdk:"Value"`
}
type TestFlexPluralityAWS01 struct {
	Value  string
	Values string
}

type TestFlexTF17 struct {
	Field1 fwtypes.ARN `tfsdk:"field1"`
}

// List/Set/Map of string types.
type TestFlexTF18 struct {
	Field1 fwtypes.ListValueOf[types.String] `tfsdk:"field1"`
	Field2 fwtypes.ListValueOf[types.String] `tfsdk:"field2"`
	Field3 fwtypes.SetValueOf[types.String]  `tfsdk:"field3"`
	Field4 fwtypes.SetValueOf[types.String]  `tfsdk:"field4"`
	Field5 fwtypes.MapValueOf[types.String]  `tfsdk:"field5"`
	Field6 fwtypes.MapValueOf[types.String]  `tfsdk:"field6"`
}

type TestFlexMapBlockKeyTF01 struct {
	MapBlock fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF02] `tfsdk:"map_block"`
}
type TestFlexMapBlockKeyAWS01 struct {
	MapBlock map[string]TestFlexMapBlockKeyAWS02
}

type TestFlexMapBlockKeyTF02 struct {
	MapBlockKey types.String `tfsdk:"map_block_key"`
	Attr1       types.String `tfsdk:"attr1"`
	Attr2       types.String `tfsdk:"attr2"`
}
type TestFlexMapBlockKeyAWS02 struct {
	Attr1 string
	Attr2 string
}

type TestFlexMapBlockKeyTF03 struct {
	MapBlock fwtypes.SetNestedObjectValueOf[TestFlexMapBlockKeyTF02] `tfsdk:"map_block"`
}

type TestFlexMapBlockKeyAWS03 struct {
	MapBlock map[string]*TestFlexMapBlockKeyAWS02
}

type TestFlexMapBlockKeyTF04 struct {
	MapBlock fwtypes.ListNestedObjectValueOf[TestFlexMapBlockKeyTF05] `tfsdk:"map_block"`
}
type TestFlexMapBlockKeyTF05 struct {
	MapBlockKey fwtypes.StringEnum[TestEnum] `tfsdk:"map_block_key"`
	Attr1       types.String                 `tfsdk:"attr1"`
	Attr2       types.String                 `tfsdk:"attr2"`
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

type TestFlexAWS19 struct {
	Field1 smithyjson.JSONStringer `json:"field1"`
}

type TestFlexTF19 struct {
	Field1 types.String `tfsdk:"field1"`
}

type TestFlexTF20 struct {
	Field1 fwtypes.SmithyJSON[smithyjson.JSONStringer] `tfsdk:"field1"`
}

type TestFlexTF21 struct {
	Field1 fwtypes.MapValueOf[fwtypes.MapValueOf[types.String]] `tfsdk:"field1"`
}

type TestFlexAWS21 struct {
	Field1 map[string]map[string]string
}

type TestFlexAWS22 struct {
	Field1 map[string]map[string]*string
}

type testFlexTFInterfaceListNestedObject struct {
	Field1 fwtypes.ListNestedObjectValueOf[testFlexTFInterfaceFlexer] `tfsdk:"field1"`
}

type testFlexTFInterfaceListNestedObjectNonFlexer struct {
	Field1 fwtypes.ListNestedObjectValueOf[TestFlexTF01] `tfsdk:"field1"`
}

type testFlexTFInterfaceSetNestedObject struct {
	Field1 fwtypes.SetNestedObjectValueOf[testFlexTFInterfaceFlexer] `tfsdk:"field1"`
}

type testFlexTFInterfaceObjectValue struct {
	Field1 fwtypes.ObjectValueOf[testFlexTFInterfaceFlexer] `tfsdk:"field1"`
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
		AWSField: t.Field1.ValueString(),
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
		AWSField: t.Field1.ValueString(),
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
		AWSField: t.Field1.ValueString(),
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

type testFlexTFExpanderListNestedObject struct {
	Field1 fwtypes.ListNestedObjectValueOf[testFlexTFFlexer] `tfsdk:"field1"`
}

type testFlexTFExpanderSetNestedObject struct {
	Field1 fwtypes.SetNestedObjectValueOf[testFlexTFFlexer] `tfsdk:"field1"`
}

type testFlexTFExpanderObjectValue struct {
	Field1 fwtypes.ObjectValueOf[testFlexTFFlexer] `tfsdk:"field1"`
}

type testFlexTFExpanderToString struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ Expander = testFlexTFExpanderToString{}

func (t testFlexTFExpanderToString) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return t.Field1.ValueString(), nil
}

type testFlexTFExpanderToNil struct {
	Field1 types.String `tfsdk:"field1"`
}

var _ Expander = testFlexTFExpanderToNil{}

func (t testFlexTFExpanderToNil) Expand(ctx context.Context) (any, diag.Diagnostics) {
	return nil, nil
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
