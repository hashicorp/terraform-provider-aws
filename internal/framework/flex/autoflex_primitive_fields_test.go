// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

// Tests AutoFlex's struct field Flatten and Expand for all primitive types and all `autoflex` struct tag combinations

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type fieldTestCase struct {
	source   any
	expected any
}

type awsPrimitiveField[T any] struct {
	Field1 T
}

type awsPrimitivePointerField[T any] struct {
	Field1 *T
}

func TestExpandStringField(t *testing.T) { //nolint:paralleltest // false positive
	testExpandField(t, stringValueHandler{})
}

func TestExpandBoolField(t *testing.T) { //nolint:paralleltest // false positive
	testExpandField(t, boolValueHandler{})
}

func TestExpandInt64Field(t *testing.T) { //nolint:paralleltest // false positive
	testExpandField(t, int64ValueHandler{})
}

func TestExpandInt32Field(t *testing.T) { //nolint:paralleltest // false positive
	testExpandField(t, int32ValueHandler{})
}

func TestExpandFloat64Field(t *testing.T) { //nolint:paralleltest // false positive
	testExpandField(t, float64ValueHandler{})
}

func TestExpandFloat32Field(t *testing.T) { //nolint:paralleltest // false positive
	testExpandField(t, float32ValueHandler{})
}

func testExpandField[T any](t *testing.T, valueHandler fieldValueHandler[T]) {
	t.Parallel()

	testcases := map[string]map[string]map[string]fieldTestCase{
		"no options": expandPrimitiveNoOptionsCases(valueHandler),

		"ignore": expandPrimitiveIgnoreFieldCases(valueHandler),

		"legacy": expandPrimitiveLegacyCases(valueHandler),

		"omitempty": expandPrimitiveOmitEmptyCases(valueHandler),

		"noexpand": expandPrimitiveNoExpandCases(valueHandler),

		"noflatten": expandPrimitiveNoFlattenCases(valueHandler),

		"legacy,omitempty": expandPrimitiveLegacyOmitEmptyCases(valueHandler),

		"legacy,noexpand": expandPrimitiveLegacyNoExpandCases(valueHandler),

		"legacy,noflatten": expandPrimitiveLegacyNoFlattenCases(valueHandler),

		"omitempty,noexpand": expandPrimitiveOmitEmptyNoExpandCases(valueHandler),

		"omitempty,noflatten": expandPrimitiveOmitEmptyNoFlattenCases(valueHandler),

		"noexpand,noflatten": expandPrimitiveNoExpandNoFlattenCases(valueHandler),

		"legacy,omitempty,noexpand": expandPrimitiveLegacyOmitEmptyNoExpandCases(valueHandler),

		"legacy,omitempty,noflatten": expandPrimitiveLegacyOmitEmptyNoFlattenCases(valueHandler),

		"legacy,noexpand,noflatten": expandPrimitiveLegacyNoExpandNoFlattenCases(valueHandler),

		"legacy,omitempty,noexpand,noflatten": expandPrimitiveLegacyOmitEmptyNoExpandNoFlattenCases(valueHandler),
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for name, tc := range tc {
				t.Run(name, func(t *testing.T) {
					cases := autoFlexTestCases{}
					for name, tc := range tc {
						cases[name] = autoFlexTestCase{
							Source:     tc.source,
							Target:     reflect.New(reflect.TypeOf(tc.expected).Elem()).Interface(),
							WantTarget: tc.expected,
						}
					}
					runAutoExpandTestCases(t, cases, runChecks{})
				})
			}
		})
	}
}

func expandPrimitiveNoOptionsCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ""
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveRegularCases(tfStructType, valueHandler)
}

func expandPrimitiveIgnoreFieldCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = "-"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveNoExpandGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveLegacyCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveLegacyGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveOmitEmptyCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",omitempty"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveOmitEmptyGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveNoExpandCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveNoExpandGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveRegularCases(tfStructType, valueHandler)
}

func expandPrimitiveLegacyOmitEmptyCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,omitempty"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveLegacyGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveLegacyNoExpandCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveNoExpandGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveLegacyNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveLegacyGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveOmitEmptyNoExpandCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",omitempty,noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveNoExpandGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveOmitEmptyNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",omitempty,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveOmitEmptyGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveNoExpandNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",noexpand,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveNoExpandGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveLegacyOmitEmptyNoExpandCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,omitempty,noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveNoExpandGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveLegacyOmitEmptyNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,omitempty,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveLegacyGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveLegacyNoExpandNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,noexpand,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveNoExpandGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveLegacyOmitEmptyNoExpandNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,omitempty,noexpand,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return expandPrimitiveNoExpandGenericCases(tfStructType, valueHandler)
}

func expandPrimitiveRegularCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value target": {
				source:   valueHandler.tfFieldWithValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithValue(),
			},
		},
		"zero value": {
			"value target": {
				source:   valueHandler.tfFieldWithZeroValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithZeroValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithZeroValue(),
			},
		},
		"null value": {
			"value target": {
				source:   valueHandler.tfFieldWithNullValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithNullValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
		"unknown value": {
			"value target": {
				source:   valueHandler.tfFieldWithUnknownValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithUnknownValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
	}
}

func expandPrimitiveLegacyGenericCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value target": {
				source:   valueHandler.tfFieldWithValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithValue(),
			},
		},
		"zero value": {
			"value target": {
				source:   valueHandler.tfFieldWithZeroValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithZeroValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
		"null value": {
			"value target": {
				source:   valueHandler.tfFieldWithNullValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithNullValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
		"unknown value": {
			"value target": {
				source:   valueHandler.tfFieldWithUnknownValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithUnknownValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
	}
}

func expandPrimitiveOmitEmptyGenericCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value target": {
				source:   valueHandler.tfFieldWithValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithValue(),
			},
		},
		"zero value": {
			"value target": {
				source:   valueHandler.tfFieldWithZeroValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithZeroValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithZeroValue(),
			},
		},
		"null value": {
			"value target": {
				source:   valueHandler.tfFieldWithNullValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithNullValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
		"unknown value": {
			"value target": {
				source:   valueHandler.tfFieldWithUnknownValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithUnknownValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
	}
}

func expandPrimitiveNoExpandGenericCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value target": {
				source:   valueHandler.tfFieldWithValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
		"zero value": {
			"value target": {
				source:   valueHandler.tfFieldWithZeroValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithZeroValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
		"null value": {
			"value target": {
				source:   valueHandler.tfFieldWithNullValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithNullValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
		"unknown value": {
			"value target": {
				source:   valueHandler.tfFieldWithUnknownValue(tfStructType),
				expected: valueHandler.awsScalarFieldWithZeroValue(),
			},
			"pointer target": {
				source:   valueHandler.tfFieldWithUnknownValue(tfStructType),
				expected: valueHandler.awsPointerFieldWithNilValue(),
			},
		},
	}
}

func tfStructSetSingleField(v reflect.Value, fieldName string, fieldValue reflect.Value) {
	field := v.FieldByName(fieldName)
	field.Set(fieldValue)
}

func tfStructSingleFieldType(name string, typ reflect.Type, autoflexTag string) reflect.Type {
	return reflect.StructOf([]reflect.StructField{
		tfStructField(name, typ, autoflexTag),
	})
}

func tfStructField(name string, typ reflect.Type, autoflexTag string) reflect.StructField {
	return reflect.StructField{
		Name: name,
		Type: typ,
		Tag:  reflect.StructTag(`tfsdk:"` + names.ToSnakeCase(name) + `" autoflex:"` + autoflexTag + `"`),
	}
}

type fieldValueHandler[T any] interface {
	nonZeroValue() T
	zeroValue() T
	fieldType() reflect.Type
	tfFieldWithValue(typ reflect.Type) any
	tfFieldWithZeroValue(typ reflect.Type) any
	tfFieldWithNullValue(typ reflect.Type) any
	tfFieldWithUnknownValue(typ reflect.Type) any
	awsScalarFieldWithValue() any
	awsScalarFieldWithZeroValue() any
	awsPointerFieldWithValue() any
	awsPointerFieldWithZeroValue() any
	awsPointerFieldWithNilValue() any
}

type stringValueHandler struct{}

var _ fieldValueHandler[string] = stringValueHandler{}

func (h stringValueHandler) nonZeroValue() string {
	return "test_value"
}

func (h stringValueHandler) zeroValue() string {
	return ""
}

func (h stringValueHandler) fieldType() reflect.Type {
	return reflect.TypeFor[types.String]()
}

func (h stringValueHandler) tfFieldWithValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.StringValue(h.nonZeroValue())))
	return p.Interface()
}

func (h stringValueHandler) tfFieldWithZeroValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.StringValue(h.zeroValue())))
	return p.Interface()
}

func (h stringValueHandler) tfFieldWithNullValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.StringNull()))
	return p.Interface()
}

func (h stringValueHandler) tfFieldWithUnknownValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.StringUnknown()))
	return p.Interface()
}

func (h stringValueHandler) awsScalarFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[string]{Field1: h.nonZeroValue()}
}

func (h stringValueHandler) awsScalarFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[string]{Field1: h.zeroValue()}
}

func (h stringValueHandler) awsPointerFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[string]{Field1: aws.String(h.nonZeroValue())}
}

func (h stringValueHandler) awsPointerFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[string]{Field1: aws.String(h.zeroValue())}
}

func (h stringValueHandler) awsPointerFieldWithNilValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[string]{Field1: nil}
}

type boolValueHandler struct{}

var _ fieldValueHandler[bool] = boolValueHandler{}

func (h boolValueHandler) nonZeroValue() bool {
	return true
}

func (h boolValueHandler) zeroValue() bool {
	return false
}

func (h boolValueHandler) fieldType() reflect.Type {
	return reflect.TypeFor[types.Bool]()
}

func (h boolValueHandler) tfFieldWithValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.BoolValue(h.nonZeroValue())))
	return p.Interface()
}

func (h boolValueHandler) tfFieldWithZeroValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.BoolValue(h.zeroValue())))
	return p.Interface()
}

func (h boolValueHandler) tfFieldWithNullValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.BoolNull()))
	return p.Interface()
}

func (h boolValueHandler) tfFieldWithUnknownValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.BoolUnknown()))
	return p.Interface()
}

func (h boolValueHandler) awsScalarFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[bool]{Field1: h.nonZeroValue()}
}

func (h boolValueHandler) awsScalarFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[bool]{Field1: h.zeroValue()}
}

func (h boolValueHandler) awsPointerFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[bool]{Field1: aws.Bool(h.nonZeroValue())}
}

func (h boolValueHandler) awsPointerFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[bool]{Field1: aws.Bool(h.zeroValue())}
}

func (h boolValueHandler) awsPointerFieldWithNilValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[bool]{Field1: nil}
}

type int64ValueHandler struct{}

var _ fieldValueHandler[int64] = int64ValueHandler{}

func (h int64ValueHandler) nonZeroValue() int64 {
	return 1
}

func (h int64ValueHandler) zeroValue() int64 {
	return 0
}

func (h int64ValueHandler) fieldType() reflect.Type {
	return reflect.TypeFor[types.Int64]()
}

func (h int64ValueHandler) tfFieldWithValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Int64Value(h.nonZeroValue())))
	return p.Interface()
}

func (h int64ValueHandler) tfFieldWithZeroValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Int64Value(h.zeroValue())))
	return p.Interface()
}

func (h int64ValueHandler) tfFieldWithNullValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Int64Null()))
	return p.Interface()
}

func (h int64ValueHandler) tfFieldWithUnknownValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Int64Unknown()))
	return p.Interface()
}

func (h int64ValueHandler) awsScalarFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[int64]{Field1: h.nonZeroValue()}
}

func (h int64ValueHandler) awsScalarFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[int64]{Field1: h.zeroValue()}
}

func (h int64ValueHandler) awsPointerFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[int64]{Field1: aws.Int64(h.nonZeroValue())}
}

func (h int64ValueHandler) awsPointerFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[int64]{Field1: aws.Int64(h.zeroValue())}
}

func (h int64ValueHandler) awsPointerFieldWithNilValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[int64]{Field1: nil}
}

type int32ValueHandler struct{}

var _ fieldValueHandler[int32] = int32ValueHandler{}

func (h int32ValueHandler) nonZeroValue() int32 {
	return 1
}

func (h int32ValueHandler) zeroValue() int32 {
	return 0
}

func (h int32ValueHandler) fieldType() reflect.Type {
	return reflect.TypeFor[types.Int32]()
}

func (h int32ValueHandler) tfFieldWithValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Int32Value(h.nonZeroValue())))
	return p.Interface()
}

func (h int32ValueHandler) tfFieldWithZeroValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Int32Value(h.zeroValue())))
	return p.Interface()
}

func (h int32ValueHandler) tfFieldWithNullValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Int32Null()))
	return p.Interface()
}

func (h int32ValueHandler) tfFieldWithUnknownValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Int32Unknown()))
	return p.Interface()
}

func (h int32ValueHandler) awsScalarFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[int32]{Field1: h.nonZeroValue()}
}

func (h int32ValueHandler) awsScalarFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[int32]{Field1: h.zeroValue()}
}

func (h int32ValueHandler) awsPointerFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[int32]{Field1: aws.Int32(h.nonZeroValue())}
}

func (h int32ValueHandler) awsPointerFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[int32]{Field1: aws.Int32(h.zeroValue())}
}

func (h int32ValueHandler) awsPointerFieldWithNilValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[int32]{Field1: nil}
}

type float64ValueHandler struct{}

var _ fieldValueHandler[float64] = float64ValueHandler{}

func (h float64ValueHandler) nonZeroValue() float64 {
	return 1
}

func (h float64ValueHandler) zeroValue() float64 {
	return 0
}

func (h float64ValueHandler) fieldType() reflect.Type {
	return reflect.TypeFor[types.Float64]()
}

func (h float64ValueHandler) tfFieldWithValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Float64Value(h.nonZeroValue())))
	return p.Interface()
}

func (h float64ValueHandler) tfFieldWithZeroValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Float64Value(h.zeroValue())))
	return p.Interface()
}

func (h float64ValueHandler) tfFieldWithNullValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Float64Null()))
	return p.Interface()
}

func (h float64ValueHandler) tfFieldWithUnknownValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Float64Unknown()))
	return p.Interface()
}

func (h float64ValueHandler) awsScalarFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[float64]{Field1: h.nonZeroValue()}
}

func (h float64ValueHandler) awsScalarFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[float64]{Field1: h.zeroValue()}
}

func (h float64ValueHandler) awsPointerFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[float64]{Field1: aws.Float64(h.nonZeroValue())}
}

func (h float64ValueHandler) awsPointerFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[float64]{Field1: aws.Float64(h.zeroValue())}
}

func (h float64ValueHandler) awsPointerFieldWithNilValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[float64]{Field1: nil}
}

type float32ValueHandler struct{}

var _ fieldValueHandler[float32] = float32ValueHandler{}

func (h float32ValueHandler) nonZeroValue() float32 {
	return 1
}

func (h float32ValueHandler) zeroValue() float32 {
	return 0
}

func (h float32ValueHandler) fieldType() reflect.Type {
	return reflect.TypeFor[types.Float32]()
}

func (h float32ValueHandler) tfFieldWithValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Float32Value(h.nonZeroValue())))
	return p.Interface()
}

func (h float32ValueHandler) tfFieldWithZeroValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Float32Value(h.zeroValue())))
	return p.Interface()
}

func (h float32ValueHandler) tfFieldWithNullValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Float32Null()))
	return p.Interface()
}

func (h float32ValueHandler) tfFieldWithUnknownValue(typ reflect.Type) any {
	p := reflect.New(typ)
	v := p.Elem()
	tfStructSetSingleField(v, "Field1", reflect.ValueOf(types.Float32Unknown()))
	return p.Interface()
}

func (h float32ValueHandler) awsScalarFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[float32]{Field1: h.nonZeroValue()}
}

func (h float32ValueHandler) awsScalarFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitiveField[float32]{Field1: h.zeroValue()}
}

func (h float32ValueHandler) awsPointerFieldWithValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[float32]{Field1: aws.Float32(h.nonZeroValue())}
}

func (h float32ValueHandler) awsPointerFieldWithZeroValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[float32]{Field1: aws.Float32(h.zeroValue())}
}

func (h float32ValueHandler) awsPointerFieldWithNilValue() any { // nosemgrep:ci.aws-in-func-name
	return &awsPrimitivePointerField[float32]{Field1: nil}
}

func TestFlattenStringField(t *testing.T) {
	t.Parallel()

	valueHandler := stringValueHandler{}

	testcases := map[string]map[string]map[string]fieldTestCase{
		"no options": flattenPrimitiveNoOptionsCases(valueHandler),

		"ignore": flattenPrimitiveIgnoreFieldCases(valueHandler),

		"legacy": flattenPrimitiveLegacyCases(valueHandler),

		"omitempty": flattenStringOmitEmptyCases(valueHandler),

		"noexpand": flattenPrimitiveNoExpandCases(valueHandler),

		"noflatten": flattenPrimitiveNoFlattenCases(valueHandler),

		"legacy,omitempty": flattenPrimitiveLegacyOmitEmptyCases(valueHandler),

		"legacy,noexpand": flattenPrimitiveLegacyNoExpandCases(valueHandler),

		"legacy,noflatten": flattenPrimitiveLegacyNoFlattenCases(valueHandler),

		"omitempty,noexpand": flattenStringOmitEmptyNoExpandCases(valueHandler),

		"omitempty,noflatten": flattenPrimitiveOmitEmptyNoFlattenCases(valueHandler),

		"noexpand,noflatten": flattenPrimitiveNoExpandNoFlattenCases(valueHandler),

		"legacy,omitempty,noexpand": flattenPrimitiveLegacyOmitEmptyNoExpandCases(valueHandler),

		"legacy,omitempty,noflatten": flattenPrimitiveLegacyOmitEmptyNoFlattenCases(valueHandler),

		"legacy,noexpand,noflatten": flattenPrimitiveLegacyNoExpandNoFlattenCases(valueHandler),

		"legacy,omitempty,noexpand,noflatten": flattenPrimitiveLegacyOmitEmptyNoExpandNoFlattenCases(valueHandler),
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for name, tc := range tc {
				t.Run(name, func(t *testing.T) {
					cases := autoFlexTestCases{}
					for name, tc := range tc {
						cases[name] = autoFlexTestCase{
							Source:     tc.source,
							Target:     reflect.New(reflect.TypeOf(tc.expected).Elem()).Interface(),
							WantTarget: tc.expected,
						}
					}
					runAutoFlattenTestCases(t, cases, runChecks{})
				})
			}
		})
	}
}

func TestFlattenBoolField(t *testing.T) { //nolint:paralleltest // false positive
	testFlattenField(t, boolValueHandler{})
}

func TestFlattenInt64Field(t *testing.T) { //nolint:paralleltest // false positive
	testFlattenField(t, int64ValueHandler{})
}

func TestFlattenInt32Field(t *testing.T) { //nolint:paralleltest // false positive
	testFlattenField(t, int32ValueHandler{})
}

func TestFlattenFloat64Field(t *testing.T) { //nolint:paralleltest // false positive
	testFlattenField(t, float64ValueHandler{})
}

func TestFlattenFloat32Field(t *testing.T) { //nolint:paralleltest // false positive
	testFlattenField(t, float32ValueHandler{})
}

func testFlattenField[T any](t *testing.T, valueHandler fieldValueHandler[T]) {
	t.Parallel()

	testcases := map[string]map[string]map[string]fieldTestCase{
		"no options": flattenPrimitiveNoOptionsCases(valueHandler),

		"ignore": flattenPrimitiveIgnoreFieldCases(valueHandler),

		"legacy": flattenPrimitiveLegacyCases(valueHandler),

		"omitempty": flattenPrimitiveOmitEmptyCases(valueHandler),

		"noexpand": flattenPrimitiveNoExpandCases(valueHandler),

		"noflatten": flattenPrimitiveNoFlattenCases(valueHandler),

		"legacy,omitempty": flattenPrimitiveLegacyOmitEmptyCases(valueHandler),

		"legacy,noexpand": flattenPrimitiveLegacyNoExpandCases(valueHandler),

		"legacy,noflatten": flattenPrimitiveLegacyNoFlattenCases(valueHandler),

		"omitempty,noexpand": flattenPrimitiveOmitEmptyNoExpandCases(valueHandler),

		"omitempty,noflatten": flattenPrimitiveOmitEmptyNoFlattenCases(valueHandler),

		"noexpand,noflatten": flattenPrimitiveNoExpandNoFlattenCases(valueHandler),

		"legacy,omitempty,noexpand": flattenPrimitiveLegacyOmitEmptyNoExpandCases(valueHandler),

		"legacy,omitempty,noflatten": flattenPrimitiveLegacyOmitEmptyNoFlattenCases(valueHandler),

		"legacy,noexpand,noflatten": flattenPrimitiveLegacyNoExpandNoFlattenCases(valueHandler),

		"legacy,omitempty,noexpand,noflatten": flattenPrimitiveLegacyOmitEmptyNoExpandNoFlattenCases(valueHandler),
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for name, tc := range tc {
				t.Run(name, func(t *testing.T) {
					cases := autoFlexTestCases{}
					for name, tc := range tc {
						cases[name] = autoFlexTestCase{
							Source:     tc.source,
							Target:     reflect.New(reflect.TypeOf(tc.expected).Elem()).Interface(),
							WantTarget: tc.expected,
						}
					}
					runAutoFlattenTestCases(t, cases, runChecks{})
				})
			}
		})
	}
}

func flattenPrimitiveNoOptionsCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ""
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveRegularCases(tfStructType, valueHandler)
}

func flattenPrimitiveIgnoreFieldCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = "-"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericNoFlattenCases(tfStructType, valueHandler)
}

func flattenPrimitiveLegacyCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericLegacyCases(tfStructType, valueHandler)
}

func flattenStringOmitEmptyCases(valueHandler stringValueHandler) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",omitempty"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenStringGenericOmitEmptyCases(tfStructType, valueHandler)
}

func flattenPrimitiveOmitEmptyCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",omitempty"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericOmitEmptyCases(tfStructType, valueHandler)
}

func flattenPrimitiveNoExpandCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveRegularCases(tfStructType, valueHandler)
}

func flattenPrimitiveNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericNoFlattenCases(tfStructType, valueHandler)
}

func flattenPrimitiveLegacyOmitEmptyCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,omitempty"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericLegacyCases(tfStructType, valueHandler)
}

func flattenPrimitiveLegacyNoExpandCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericLegacyCases(tfStructType, valueHandler)
}

func flattenPrimitiveLegacyNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericNoFlattenCases(tfStructType, valueHandler)
}

func flattenStringOmitEmptyNoExpandCases(valueHandler stringValueHandler) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",omitempty,noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenStringGenericOmitEmptyCases(tfStructType, valueHandler)
}

func flattenPrimitiveOmitEmptyNoExpandCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",omitempty,noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericOmitEmptyCases(tfStructType, valueHandler)
}

func flattenPrimitiveOmitEmptyNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",omitempty,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericNoFlattenCases(tfStructType, valueHandler)
}

func flattenPrimitiveNoExpandNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",noexpand,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericNoFlattenCases(tfStructType, valueHandler)
}

func flattenPrimitiveLegacyOmitEmptyNoExpandCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,omitempty,noexpand"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericLegacyCases(tfStructType, valueHandler)
}

func flattenPrimitiveLegacyOmitEmptyNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,omitempty,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericNoFlattenCases(tfStructType, valueHandler)
}

func flattenPrimitiveLegacyNoExpandNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,noexpand,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericNoFlattenCases(tfStructType, valueHandler)
}

func flattenPrimitiveLegacyOmitEmptyNoExpandNoFlattenCases[T any](valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	const fieldName = "Field1"
	const autoflexTag = ",legacy,omitempty,noexpand,noflatten"
	tfStructType := tfStructSingleFieldType(fieldName, valueHandler.fieldType(), autoflexTag)

	return flattenPrimitiveGenericNoFlattenCases(tfStructType, valueHandler)
}

func flattenPrimitiveRegularCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithValue(),
				expected: valueHandler.tfFieldWithValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithValue(),
				expected: valueHandler.tfFieldWithValue(tfStructType),
			},
		},
		"zero value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithZeroValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithZeroValue(tfStructType),
			},
		},
		"null value": {
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithNilValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
		},
	}
}

func flattenPrimitiveGenericLegacyCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithValue(),
				expected: valueHandler.tfFieldWithValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithValue(),
				expected: valueHandler.tfFieldWithValue(tfStructType),
			},
		},
		"zero value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithZeroValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithZeroValue(tfStructType),
			},
		},
		"null value": {
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithNilValue(),
				expected: valueHandler.tfFieldWithZeroValue(tfStructType),
			},
		},
	}
}

func flattenStringGenericOmitEmptyCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithValue(),
				expected: valueHandler.tfFieldWithValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithValue(),
				expected: valueHandler.tfFieldWithValue(tfStructType),
			},
		},
		"zero value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
		},
		"null value": {
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithNilValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
		},
	}
}

func flattenPrimitiveGenericOmitEmptyCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithValue(),
				expected: valueHandler.tfFieldWithValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithValue(),
				expected: valueHandler.tfFieldWithValue(tfStructType),
			},
		},
		"zero value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithZeroValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithZeroValue(tfStructType),
			},
		},
		"null value": {
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithNilValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
		},
	}
}

func flattenPrimitiveGenericNoFlattenCases[T any](tfStructType reflect.Type, valueHandler fieldValueHandler[T]) map[string]map[string]fieldTestCase {
	return map[string]map[string]fieldTestCase{
		"with value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
		},
		"zero value": {
			"value source": {
				source:   valueHandler.awsScalarFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithZeroValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
		},
		"null value": {
			"pointer source": {
				source:   valueHandler.awsPointerFieldWithNilValue(),
				expected: valueHandler.tfFieldWithNullValue(tfStructType),
			},
		},
	}
}

type tfBoolField struct {
	Field1 types.Bool `tfsdk:"field1"`
}

type tfFloat64Field struct {
	Field1 types.Float64 `tfsdk:"field1"`
}

type tfFloat32Field struct {
	Field1 types.Float32 `tfsdk:"field1"`
}

type tfInt64Field struct {
	Field1 types.Int64 `tfsdk:"field1"`
}

type tfInt32Field struct {
	Field1 types.Int32 `tfsdk:"field1"`
}

type awsSingleStructField struct {
	Field1 struct{ X int }
}

func TestExpandPrimitiveIncompatibleTypes(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"bool to incompatible": {
			Source:     &tfBoolField{Field1: types.BoolValue(true)},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
		"float64 to incompatible": {
			Source:     &tfFloat64Field{Field1: types.Float64Value(1.1)},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
		"float32 to incompatible": {
			Source:        &tfFloat32Field{Field1: types.Float32Value(1.1)},
			Target:        &awsSingleStringValue{},
			WantTarget:    &awsSingleStringValue{},
			ExpectedDiags: diag.Diagnostics{diagExpandingIncompatibleTypes(reflect.TypeFor[types.Float32](), reflect.TypeFor[string]())},
		},
		"int64 to incompatible": {
			Source:     &tfInt64Field{Field1: types.Int64Value(1)},
			Target:     &awsSingleStringValue{},
			WantTarget: &awsSingleStringValue{},
		},
		"int32 to incompatible": {
			Source:        &tfInt32Field{Field1: types.Int32Value(1)},
			Target:        &awsSingleStringValue{},
			WantTarget:    &awsSingleStringValue{},
			ExpectedDiags: diag.Diagnostics{diagExpandingIncompatibleTypes(reflect.TypeFor[types.Int32](), reflect.TypeFor[string]())},
		},
		"string to incompatible": {
			Source:     &tfSingleStringField{Field1: types.StringValue("hello")},
			Target:     &awsSingleStructField{},
			WantTarget: &awsSingleStructField{},
		},
	}

	runAutoExpandTestCases(t, testCases, runChecks{})
}

type awsSingleBoolField struct {
	Field1 bool
}

type awsSingleFloat64Field struct {
	Field1 float64
}

type awsSingleFloat32Field struct {
	Field1 float32
}

type awsSingleInt64Field struct {
	Field1 int64
}

type awsSingleInt32Field struct {
	Field1 int32
}

func TestFlattenPrimitiveIncompatibleTypes(t *testing.T) {
	t.Parallel()

	testCases := autoFlexTestCases{
		"bool to string": {
			Source:     &awsSingleBoolField{Field1: true},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{},
			WantDiff:   true,
		},
		"float64 to string": {
			Source:     &awsSingleFloat64Field{Field1: 1.1},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{},
			WantDiff:   true,
		},
		"float32 to string": {
			Source:     &awsSingleFloat32Field{Field1: 1.1},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{},
			WantDiff:   true,
		},
		"int64 to string": {
			Source:     &awsSingleInt64Field{Field1: 1},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{},
			WantDiff:   true,
		},
		"int32 to string": {
			Source:     &awsSingleInt32Field{Field1: 1},
			Target:     &tfSingleStringField{},
			WantTarget: &tfSingleStringField{},
			WantDiff:   true,
		},
		"string to bool": {
			Source:     &awsSingleStringValue{Field1: "hello"},
			Target:     &tfBoolField{},
			WantTarget: &tfBoolField{},
			WantDiff:   true,
		},
	}

	runAutoFlattenTestCases(t, testCases, runChecks{})
}
