// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfreflect "github.com/hashicorp/terraform-provider-aws/internal/reflect"
	"github.com/shopspring/decimal"
)

// Flatten = AWS --> TF

// Flattener is implemented by types that customize their flattening
type Flattener interface {
	Flatten(ctx context.Context, v any) diag.Diagnostics
}

// Flatten "flattens" an AWS SDK for Go v2 API data structure into
// a resource's "business logic" data structure, implemented using
// Terraform Plugin Framework data types.
// The API data structure's fields are walked and exported fields that
// have a corresponding field in the resource's data structure (and a
// suitable target data type) are copied.
func Flatten(ctx context.Context, apiObject, tfObject any, optFns ...AutoFlexOptionsFunc) diag.Diagnostics {
	var diags diag.Diagnostics
	flattener := newAutoFlattener(optFns)

	tflog.SubsystemInfo(ctx, subsystemName, "Flattening", map[string]any{
		logAttrKeySourceType: fullTypeName(reflect.TypeOf(apiObject)),
		logAttrKeyTargetType: fullTypeName(reflect.TypeOf(tfObject)),
	})

	diags.Append(autoFlattenConvert(ctx, apiObject, tfObject, flattener)...)

	return diags
}

type autoFlattener struct {
	Options AutoFlexOptions
}

// newAutoFlattener initializes an auto-flattener with defaults that can be overridden
// via functional options
func newAutoFlattener(optFns []AutoFlexOptionsFunc) *autoFlattener {
	o := AutoFlexOptions{
		ignoredFieldNames: DefaultIgnoredFieldNames,
	}

	for _, optFn := range optFns {
		optFn(&o)
	}

	return &autoFlattener{
		Options: o,
	}
}

func (flattener autoFlattener) getOptions() AutoFlexOptions {
	return flattener.Options
}

// autoFlattenConvert converts `from` to `to` using the specified auto-flexer.
func autoFlattenConvert(ctx context.Context, from, to any, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	sourcePath := path.Empty()
	targetPath := path.Empty()

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(reflect.TypeOf(from)))
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(reflect.TypeOf(to)))

	if _, ok := to.(attr.Value); !ok {
		vf := reflect.ValueOf(from)
		if vf.Kind() == reflect.Invalid || vf.Kind() == reflect.Pointer && vf.IsNil() {
			tflog.SubsystemError(ctx, subsystemName, "Source is nil")
			diags.Append(diagFlatteningSourceIsNil(reflect.TypeOf(from)))
			return diags
		}
	}

	ctx, valFrom, valTo, d := autoFlexValues(ctx, from, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Top-level struct to struct conversion.
	if valFrom.IsValid() && valTo.IsValid() {
		if typFrom, typTo := valFrom.Type(), valTo.Type(); typFrom.Kind() == reflect.Struct && typTo.Kind() == reflect.Struct &&
			!typTo.Implements(reflect.TypeFor[basetypes.ListValuable]()) &&
			!typTo.Implements(reflect.TypeFor[basetypes.SetValuable]()) {
			tflog.SubsystemInfo(ctx, subsystemName, "Converting")
			diags.Append(flattenStruct(ctx, sourcePath, from, targetPath, to, flexer)...)
			return diags
		}
	}

	valFrom = reflect.ValueOf(from)

	// Anything else.
	diags.Append(flexer.convert(ctx, sourcePath, valFrom, targetPath, valTo, fieldOpts{})...)
	return diags
}

// convert converts a single AWS API value to its Plugin Framework equivalent.
func (flattener autoFlattener) convert(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(valueType(vFrom)))
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(valueType(vTo)))

	valTo, ok := vTo.Interface().(attr.Value)
	if !ok {
		// Check for `nil` (i.e. Kind == Invalid) here, because we allow primitive types to be `nil`
		if vFrom.Kind() == reflect.Invalid {
			tflog.SubsystemError(ctx, subsystemName, "Source is nil")
			diags.Append(diagFlatteningSourceIsNil(valueType(vFrom)))
			return diags
		}

		tflog.SubsystemError(ctx, subsystemName, "Target does not implement attr.Value")
		diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(vTo.Interface())))
		return diags
	}

	tflog.SubsystemInfo(ctx, subsystemName, "Converting")

	// main control flow
	tTo := valTo.Type(ctx)
	switch k := vFrom.Kind(); k {
	case reflect.Bool:
		diags.Append(flattener.bool(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Float64:
		diags.Append(flattener.float64(ctx, vFrom, vFrom.Type(), false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Float32:
		diags.Append(flattener.float32(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Int64:
		diags.Append(flattener.int64(ctx, vFrom, vFrom.Type(), false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Int32:
		diags.Append(flattener.int32(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.String:
		// []byte (or []uint8) is also handled like string but in flattener.slice()
		diags.Append(flattener.string(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Pointer:
		diags.Append(flattener.pointer(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Slice:
		diags.Append(flattener.slice(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Map:
		diags.Append(flattener.map_(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
		return diags

	case reflect.Struct:
		diags.Append(flattener.struct_(ctx, sourcePath, vFrom, false, targetPath, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Interface:
		diags.Append(flattener.interface_(ctx, vFrom, tTo, vTo)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// bool copies an AWS API bool value to a compatible Plugin Framework value.
func (flattener autoFlattener) bool(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.BoolTypable:
		//
		// bool -> types.Bool.
		//
		var boolValue types.Bool
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				boolValue = types.BoolValue(false)
			} else {
				boolValue = types.BoolValue(vFrom.Bool())
			}
		} else {
			if isNullFrom {
				boolValue = types.BoolNull()
			} else {
				boolValue = types.BoolValue(vFrom.Bool())
			}
		}
		v, d := tTo.ValueFromBool(ctx, boolValue)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// float64 copies an AWS API float64 value to a compatible Plugin Framework value.
func (flattener autoFlattener) float64(ctx context.Context, vFrom reflect.Value, sourceType reflect.Type, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Float64Typable:
		//
		// float64 -> types.Float64.
		//
		var float64Value types.Float64
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				float64Value = types.Float64Value(0)
			} else {
				float64Value = types.Float64Value(vFrom.Float())
			}
		} else {
			if isNullFrom {
				float64Value = types.Float64Null()
			} else {
				float64Value = types.Float64Value(vFrom.Float())
			}
		}
		v, d := tTo.ValueFromFloat64(ctx, float64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case basetypes.Float32Typable:
		// Only returns an error when the target type is Float32Typable to prevent breaking existing resources
		tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")
		diags.Append(DiagFlatteningIncompatibleTypes(sourceType, vTo.Type()))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// float32 copies an AWS API float32 value to a compatible Plugin Framework value.
func (flattener autoFlattener) float32(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Float64Typable:
		//
		// float32 -> types.Float64.
		//
		var float64Value types.Float64
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				float64Value = types.Float64Value(0)
			} else {
				// Avoid loss of equivalence.
				from := vFrom.Interface().(float32)
				float64Value = types.Float64Value(decimal.NewFromFloat32(from).InexactFloat64())
			}
		} else {
			if isNullFrom {
				float64Value = types.Float64Null()
			} else {
				// Avoid loss of equivalence.
				from := vFrom.Interface().(float32)
				float64Value = types.Float64Value(decimal.NewFromFloat32(from).InexactFloat64())
			}
		}
		v, d := tTo.ValueFromFloat64(ctx, float64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case basetypes.Float32Typable:
		//
		// float32/float64 -> types.Float32.
		//
		var float32Value types.Float32
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				float32Value = types.Float32Value(0)
			} else {
				// Avoid loss of equivalence.
				from := vFrom.Interface().(float32)
				float32Value = types.Float32Value(float32(decimal.NewFromFloat32(from).InexactFloat64()))
			}
		} else {
			if isNullFrom {
				float32Value = types.Float32Null()
			} else {
				// Avoid loss of equivalence.
				from := vFrom.Interface().(float32)
				float32Value = types.Float32Value(float32(decimal.NewFromFloat32(from).InexactFloat64()))
			}
		}
		v, d := tTo.ValueFromFloat32(ctx, float32Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// int64 copies an AWS API int64 value to a compatible Plugin Framework value.
func (flattener autoFlattener) int64(ctx context.Context, vFrom reflect.Value, sourceType reflect.Type, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Int64Typable:
		//
		// int32/int64 -> types.Int64.
		//
		var int64Value types.Int64
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				int64Value = types.Int64Value(0)
			} else {
				int64Value = types.Int64Value(vFrom.Int())
			}
		} else {
			if isNullFrom {
				int64Value = types.Int64Null()
			} else {
				int64Value = types.Int64Value(vFrom.Int())
			}
		}
		v, d := tTo.ValueFromInt64(ctx, int64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case basetypes.Int32Typable:
		// Only returns an error when the target type is Int32Typeable to prevent breaking existing resources
		tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")
		diags.Append(DiagFlatteningIncompatibleTypes(sourceType, vTo.Type()))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// int32 copies an AWS API int32 value to a compatible Plugin Framework value.
func (flattener autoFlattener) int32(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Int64Typable:
		//
		// int32/int64 -> types.Int64.
		//
		var int64Value types.Int64
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				int64Value = types.Int64Value(0)
			} else {
				int64Value = types.Int64Value(vFrom.Int())
			}
		} else {
			if isNullFrom {
				int64Value = types.Int64Null()
			} else {
				int64Value = types.Int64Value(vFrom.Int())
			}
		}
		v, d := tTo.ValueFromInt64(ctx, int64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case basetypes.Int32Typable:
		//
		// int32/int64 -> types.Int32.
		//
		var int32Value basetypes.Int32Value
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				int32Value = types.Int32Value(0)
			} else {
				int32Value = types.Int32Value(int32(vFrom.Int()))
			}
		} else {
			if isNullFrom {
				int32Value = types.Int32Null()
			} else {
				int32Value = types.Int32Value(int32(vFrom.Int()))
			}
		}
		v, d := tTo.ValueFromInt32(ctx, int32Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// string copies an AWS API string value to a compatible Plugin Framework value.
func (flattener autoFlattener) string(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.StringTypable:
		stringValue := types.StringNull()
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				stringValue = types.StringValue("")
			} else {
				stringValue = types.StringValue(vFrom.String())
			}
		} else {
			if !isNullFrom {
				if value := vFrom.String(); !(value == "" && (strings.HasPrefix(tTo.String(), "StringEnumType[") || fieldOpts.omitempty)) {
					stringValue = types.StringValue(value)
				}
			}
		}
		v, d := tTo.ValueFromString(ctx, stringValue)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// string -> types.String.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

func (flattener autoFlattener) time(ctx context.Context, vFrom reflect.Value, isNullFrom bool, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if isNullFrom {
		vTo.Set(reflect.ValueOf(timetypes.NewRFC3339Null()))
		return diags
	}

	if !vFrom.CanInterface() {
		diags.AddError("AutoFlEx", fmt.Sprintf("cannot create an interface for: %T", vFrom))
		return diags
	}

	// time.Time --> timetypes.RFC3339
	if from, ok := vFrom.Interface().(time.Time); ok {
		vTo.Set(reflect.ValueOf(timetypes.NewRFC3339TimeValue(from)))
		return diags
	}

	if !vFrom.Elem().CanInterface() {
		diags.AddError("AutoFlEx", fmt.Sprintf("cannot create an interface for: %T", vFrom.Elem()))
		return diags
	}

	// *time.Time --> timetypes.RFC3339
	if from, ok := vFrom.Elem().Interface().(time.Time); ok {
		vTo.Set(reflect.ValueOf(timetypes.NewRFC3339TimeValue(from)))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   vTo,
	})

	return diags
}

// pointer copies an AWS API pointer value to a compatible Plugin Framework value.
func (flattener autoFlattener) pointer(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vElem, isNilFrom := vFrom.Elem(), vFrom.IsNil(); vFrom.Type().Elem().Kind() {
	case reflect.Bool:
		diags.Append(flattener.bool(ctx, vElem, isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Float64:
		diags.Append(flattener.float64(ctx, vElem, vFrom.Type(), isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Float32:
		diags.Append(flattener.float32(ctx, vElem, isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Int64:
		diags.Append(flattener.int64(ctx, vElem, vFrom.Type(), isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Int32:
		diags.Append(flattener.int32(ctx, vElem, isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.String:
		diags.Append(flattener.string(ctx, vElem, isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Struct:
		diags.Append(flattener.struct_(ctx, sourcePath, vElem, isNilFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

func (flattener autoFlattener) interface_(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.StringTypable:
		//
		// JSONStringer -> types.String-ish.
		//
		if vFrom.Type().Implements(reflect.TypeFor[smithyjson.JSONStringer]()) {
			tflog.SubsystemInfo(ctx, subsystemName, "Source implements json.JSONStringer")

			stringValue := types.StringNull()

			if vFrom.IsNil() {
				tflog.SubsystemTrace(ctx, subsystemName, "Flattening null value")
			} else {
				doc := vFrom.Interface().(smithyjson.JSONStringer)
				b, err := doc.MarshalSmithyDocument()
				if err != nil {
					// An error here would be an upstream error in the AWS SDK, because errors in json.Marshal
					// are caused by conditions such as cyclic structures
					// See https://pkg.go.dev/encoding/json#Marshal
					tflog.SubsystemError(ctx, subsystemName, "Marshalling JSON document", map[string]any{
						logAttrKeyError: err.Error(),
					})
					diags.Append(diagFlatteningMarshalSmithyDocument(reflect.TypeOf(doc), err))
					return diags
				}
				stringValue = types.StringValue(string(b))
			}
			v, d := tTo.ValueFromString(ctx, stringValue)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(v))
			return diags
		}

	case fwtypes.NestedObjectType:
		//
		// interface -> types.List(OfObject) or types.Object.
		//
		diags.Append(flattener.interfaceToNestedObject(ctx, vFrom, vFrom.IsNil(), tTo, vTo)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")

	return diags
}

// struct_ copies an AWS API struct value to a compatible Plugin Framework value.
func (flattener autoFlattener) struct_(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, isNilFrom bool, targetPath path.Path, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
		//
		// *struct -> types.List(OfObject) or types.Object.
		//
		diags.Append(flattener.structToNestedObject(ctx, sourcePath, vFrom, isNilFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags
	}

	iTo, ok := vTo.Interface().(attr.Value)
	if !ok {
		diags.AddError("AutoFlEx", fmt.Sprintf("does not implement attr.Value: %s", vTo.Kind()))
		return diags
	}

	switch iTo.(type) {
	case timetypes.RFC3339:
		diags.Append(flattener.time(ctx, vFrom, isNilFrom, vTo)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")

	return diags
}

// slice copies an AWS API slice value to a compatible Plugin Framework value.
func (flattener autoFlattener) slice(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tSliceElem := vFrom.Type().Elem(); tSliceElem.Kind() {
	case reflect.Int32, reflect.Int64:
		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []int32 or []int64 -> types.List(OfInt64).
			//
			diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectValue, fieldOpts)...)
			return diags

		case basetypes.SetTypable:
			//
			// []int32 or []int64 -> types.Set(OfInt64).
			//
			diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectValue, fieldOpts)...)
			return diags
		}

	case reflect.String:
		var elementType attr.Type = types.StringType
		attrValueFromReflectValue := newStringValueFromReflectValue
		if tTo, ok := tTo.(attr.TypeWithElementType); ok {
			if tElem, ok := tTo.ElementType().(basetypes.StringTypable); ok {
				//
				// []stringy -> types.List/Set(OfStringEnum).
				//
				elementType = tElem
				attrValueFromReflectValue = func(val reflect.Value) (attr.Value, diag.Diagnostics) {
					return tElem.ValueFromString(ctx, types.StringValue(val.String()))
				}
			}
		}

		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []string -> types.List(OfString).
			//
			diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, elementType, attrValueFromReflectValue, fieldOpts)...)
			return diags

		case basetypes.SetTypable:
			//
			// []string -> types.Set(OfString).
			//
			diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, elementType, attrValueFromReflectValue, fieldOpts)...)
			return diags
		}

	case reflect.Uint8:
		switch tTo := tTo.(type) {
		case basetypes.StringTypable:
			//
			// []byte (or []uint8) -> types.String.
			//
			vFrom = reflect.ValueOf(string(vFrom.Bytes()))
			diags.Append(flattener.string(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
			return diags
		}

	case reflect.Pointer:
		switch tSliceElem.Elem().Kind() {
		case reflect.Int32:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*int32 -> types.List(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue, fieldOpts)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*int32 -> types.Set(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue, fieldOpts)...)
				return diags
			}

		case reflect.Int64:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*int64 -> types.List(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue, fieldOpts)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*int64 -> types.Set(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue, fieldOpts)...)
				return diags
			}

		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*string -> types.List(OfString).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.StringType, newStringValueFromReflectPointerValue, fieldOpts)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*string -> types.Set(OfString).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.StringType, newStringValueFromReflectPointerValue, fieldOpts)...)
				return diags
			}

		case reflect.Struct:
			if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
				//
				// []*struct -> types.List(OfObject).
				//
				diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
				return diags
			}
		}

	case reflect.Struct:
		if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
			//
			// []struct -> types.List(OfObject).
			//
			diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
			return diags
		}

	case reflect.Interface:
		if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
			//
			// []interface -> types.List(OfObject).
			//
			diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// map_ copies an AWS API map value to a compatible Plugin Framework value.
func (flattener autoFlattener) map_(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tMapKey := vFrom.Type().Key(); tMapKey.Kind() {
	case reflect.String:
		switch tMapElem := vFrom.Type().Elem(); tMapElem.Kind() {
		case reflect.Struct:
			switch tTo := tTo.(type) {
			case basetypes.SetTypable:
				//
				// map[string]struct -> fwtypes.SetNestedObjectOf[Object]
				//
				if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
					diags.Append(flattener.structMapToObjectList(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
					return diags
				}

			case basetypes.ListTypable:
				//
				// map[string]struct -> fwtypes.ListNestedObjectOf[Object]
				//
				if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
					diags.Append(flattener.structMapToObjectList(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
					return diags
				}
			}

		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.MapTypable:
				//
				// map[string]string -> types.Map(OfString).
				//
				if vFrom.IsNil() {
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapNull")
					to, d := tTo.ValueFromMap(ctx, types.MapNull(types.StringType))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}

				from := vFrom.Interface().(map[string]string)
				tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapValue", map[string]any{
					logAttrKeySourceSize: len(from),
				})
				elements := make(map[string]attr.Value, len(from))
				for k, v := range from {
					elements[k] = types.StringValue(v)
				}
				map_, d := types.MapValue(types.StringType, elements)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				to, d := tTo.ValueFromMap(ctx, map_)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}

		case reflect.Map:
			switch tTo := tTo.(type) {
			case basetypes.MapTypable:
				//
				// map[string]map[string]string -> types.Map(OfMap[types.String]).
				//
				if vFrom.IsNil() {
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapNull")
					to, d := tTo.ValueFromMap(ctx, types.MapNull(types.MapType{ElemType: types.StringType}))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}

				switch tMapElem.Elem().Kind() {
				case reflect.String:
					from := vFrom.Interface().(map[string]map[string]string)
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening map", map[string]any{
						logAttrKeySourceSize: len(from),
					})
					elements := make(map[string]attr.Value, len(from))
					for k, v := range from {
						innerElements := make(map[string]attr.Value, len(v))
						for ik, iv := range v {
							innerElements[ik] = types.StringValue(iv)
						}
						tflog.SubsystemTrace(ctx, subsystemName, "Flattening with NewMapValueOf", map[string]any{
							logAttrKeySourcePath: sourcePath.AtMapKey(k).String(),
							logAttrKeySourceType: fullTypeName(reflect.TypeOf(v)),
							logAttrKeySourceSize: len(v),
							logAttrKeyTargetPath: targetPath.AtMapKey(k).String(),
							logAttrKeyTargetType: fullTypeName(reflect.TypeOf(innerElements)),
						})
						innerMap, d := fwtypes.NewMapValueOf[types.String](ctx, innerElements)
						diags.Append(d...)
						if diags.HasError() {
							return diags
						}

						elements[k] = innerMap
					}
					map_, d := fwtypes.NewMapValueOf[fwtypes.MapValueOf[types.String]](ctx, elements)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					to, d := tTo.ValueFromMap(ctx, map_.MapValue)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags

				case reflect.Pointer:
					from := vFrom.Interface().(map[string]map[string]*string)
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening map", map[string]any{
						logAttrKeySourceSize: len(from),
					})
					elements := make(map[string]attr.Value, len(from))
					for k, v := range from {
						innerElements := make(map[string]attr.Value, len(v))
						for ik, iv := range v {
							innerElements[ik] = types.StringValue(*iv)
						}
						tflog.SubsystemTrace(ctx, subsystemName, "Flattening with NewMapValueOf", map[string]any{
							logAttrKeySourcePath: sourcePath.AtMapKey(k).String(),
							logAttrKeySourceType: fullTypeName(reflect.TypeOf(v)),
							logAttrKeySourceSize: len(v),
							logAttrKeyTargetPath: targetPath.AtMapKey(k).String(),
							logAttrKeyTargetType: fullTypeName(reflect.TypeOf(innerElements)),
						})
						innerMap, d := fwtypes.NewMapValueOf[types.String](ctx, innerElements)
						diags.Append(d...)
						if diags.HasError() {
							return diags
						}

						elements[k] = innerMap
					}
					map_, d := fwtypes.NewMapValueOf[fwtypes.MapValueOf[types.String]](ctx, elements)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					to, d := tTo.ValueFromMap(ctx, map_.MapValue)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}
			}

		case reflect.Pointer:
			switch tMapElem.Elem().Kind() {
			case reflect.Struct:
				if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
					diags.Append(flattener.structMapToObjectList(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
					return diags
				}

			case reflect.String:
				switch tTo := tTo.(type) {
				case basetypes.ListTypable:
					//
					// map[string]struct -> fwtypes.ListNestedObjectOf[Object]
					//
					if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
						diags.Append(flattener.structMapToObjectList(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
						return diags
					}

				case basetypes.MapTypable:
					//
					// map[string]*string -> types.Map(OfString).
					//
					if vFrom.IsNil() {
						tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapNull")
						to, d := tTo.ValueFromMap(ctx, types.MapNull(types.StringType))
						diags.Append(d...)
						if diags.HasError() {
							return diags
						}

						vTo.Set(reflect.ValueOf(to))
						return diags
					}

					from := vFrom.Interface().(map[string]*string)
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapValue", map[string]any{
						logAttrKeySourceSize: len(from),
					})
					elements := make(map[string]attr.Value, len(from))
					for k, v := range from {
						elements[k] = types.StringPointerValue(v)
					}
					map_, d := types.MapValue(types.StringType, elements)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					to, d := tTo.ValueFromMap(ctx, map_)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

func (flattener autoFlattener) structMapToObjectList(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if vFrom.IsNil() {
		val, d := tTo.NullValue(ctx)
		tflog.SubsystemTrace(ctx, subsystemName, "Flattening null value")
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))
		return diags
	}

	n := vFrom.Len()
	to, d := tTo.NewObjectSlice(ctx, n, n)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	t := reflect.ValueOf(to)

	i := 0
	for _, key := range vFrom.MapKeys() {
		sourcePath := sourcePath.AtMapKey(key.Interface().(string))
		targetPath := targetPath.AtListIndex(i)
		ctx := tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())
		target, d := tTo.NewObjectPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		fromInterface := vFrom.MapIndex(key).Interface()
		if vFrom.MapIndex(key).Kind() == reflect.Pointer {
			fromInterface = vFrom.MapIndex(key).Elem().Interface()
		}

		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(reflect.TypeOf(fromInterface)))

		diags.Append(setMapBlockKey(ctx, target, key)...)
		if diags.HasError() {
			return diags
		}

		diags.Append(flattenStruct(ctx, sourcePath, fromInterface, targetPath, target, flattener)...)
		if diags.HasError() {
			return diags
		}

		t.Index(i).Set(reflect.ValueOf(target))
		i++
	}

	val, d := tTo.ValueFromObjectSlice(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))

	return diags
}

// structToNestedObject copies an AWS API struct value to a compatible Plugin Framework NestedObjectValue value.
func (flattener autoFlattener) structToNestedObject(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, isNullFrom bool, targetPath path.Path, tTo fwtypes.NestedObjectType, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if fieldOpts.legacy {
		tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
		if isNullFrom {
			to, d := tTo.NewObjectPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			val, d := tTo.ValueFromObjectPtr(ctx, to)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(val))
			return diags
		}
	} else {
		if isNullFrom {
			val, d := tTo.NullValue(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(val))
			return diags
		}
	}

	// Create a new target structure and walk its fields.
	to, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(flattenStruct(ctx, sourcePath, vFrom.Interface(), targetPath, to, flattener)...)
	if diags.HasError() {
		return diags
	}

	// Set the target structure as a mapped Object.
	val, d := tTo.ValueFromObjectPtr(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

// interfaceToNestedObject copies an AWS API interface value to a compatible Plugin Framework NestedObjectValue value.
func (flattener autoFlattener) interfaceToNestedObject(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo fwtypes.NestedObjectType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if isNullFrom {
		val, d := tTo.NullValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))
		return diags
	}

	to, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	toFlattener, ok := to.(Flattener)
	if !ok {
		val, d := tTo.NullValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))

		tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
			"from": vFrom.Kind(),
			"to":   tTo,
		})
		return diags
	}

	tflog.SubsystemInfo(ctx, subsystemName, "Target implements flex.Flattener")

	// Dereference interface
	vFrom = vFrom.Elem()
	// If it's a pointer, dereference again to get the underlying type
	if vFrom.Kind() == reflect.Pointer {
		vFrom = vFrom.Elem()
	}

	diags.Append(flattenFlattener(ctx, vFrom, toFlattener)...)
	if diags.HasError() {
		return diags
	}

	// Set the target structure as a mapped Object.
	val, d := tTo.ValueFromObjectPtr(ctx, toFlattener)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

// sliceOfPrimtiveToList copies an AWS API slice of primitive (or pointer to primitive) value to a compatible Plugin Framework List value.
func (flattener autoFlattener) sliceOfPrimtiveToList(ctx context.Context, vFrom reflect.Value, tTo basetypes.ListTypable, vTo reflect.Value, elementType attr.Type, attrValueFromReflectValue attrValueFromReflectValueFunc, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if fieldOpts.legacy {
		tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
		if vFrom.IsNil() {
			list, d := types.ListValue(elementType, []attr.Value{})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			to, d := tTo.ValueFromList(ctx, list)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags
		}
	} else {
		if vFrom.IsNil() {
			tflog.SubsystemTrace(ctx, subsystemName, "Flattening with ListNull")
			to, d := tTo.ValueFromList(ctx, types.ListNull(elementType))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags
		}
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Flattening with ListValue", map[string]any{
		logAttrKeySourceSize: vFrom.Len(),
	})
	elements := make([]attr.Value, vFrom.Len())
	for i := range vFrom.Len() {
		value, d := attrValueFromReflectValue(vFrom.Index(i))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		elements[i] = value
	}
	list, d := types.ListValue(elementType, elements)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	to, d := tTo.ValueFromList(ctx, list)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(to))
	return diags
}

// sliceOfPrimitiveToSet copies an AWS API slice of primitive (or pointer to primitive) value to a compatible Plugin Framework Set value.
func (flattener autoFlattener) sliceOfPrimitiveToSet(ctx context.Context, vFrom reflect.Value, tTo basetypes.SetTypable, vTo reflect.Value, elementType attr.Type, attrValueFromReflectValue attrValueFromReflectValueFunc, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if fieldOpts.legacy {
		tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
		if vFrom.IsNil() {
			set, d := types.SetValue(elementType, []attr.Value{})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			to, d := tTo.ValueFromSet(ctx, set)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags
		}
	} else {
		if vFrom.IsNil() {
			tflog.SubsystemTrace(ctx, subsystemName, "Flattening with SetNull")
			to, d := tTo.ValueFromSet(ctx, types.SetNull(elementType))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags
		}
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Flattening with SetValue", map[string]any{
		logAttrKeySourceSize: vFrom.Len(),
	})
	elements := make([]attr.Value, vFrom.Len())
	for i := range vFrom.Len() {
		value, d := attrValueFromReflectValue(vFrom.Index(i))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		elements[i] = value
	}
	set, d := types.SetValue(elementType, elements)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	to, d := tTo.ValueFromSet(ctx, set)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(to))
	return diags
}

// sliceOfStructToNestedObjectCollection copies an AWS API []struct value to a compatible Plugin Framework NestedObjectCollectionValue value.
func (flattener autoFlattener) sliceOfStructToNestedObjectCollection(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if fieldOpts.legacy {
		tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
		if vFrom.IsNil() {
			to, d := tTo.NewObjectSlice(ctx, 0, 0)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			val, d := tTo.ValueFromObjectSlice(ctx, to)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(val))
			return diags
		}
	} else {
		if vFrom.IsNil() {
			tflog.SubsystemTrace(ctx, subsystemName, "Flattening with NullValue")
			val, d := tTo.NullValue(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(val))
			return diags
		}
	}

	// Create a new target slice and flatten each element.
	n := vFrom.Len()

	tflog.SubsystemTrace(ctx, subsystemName, "Flattening nested object collection", map[string]any{
		logAttrKeySourceSize: n,
	})

	to, d := tTo.NewObjectSlice(ctx, n, n)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	t := reflect.ValueOf(to)
	for i := range n {
		sourcePath := sourcePath.AtListIndex(i)
		targetPath := targetPath.AtListIndex(i)
		ctx := tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())
		target, d := tTo.NewObjectPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		diags.Append(flattenStruct(ctx, sourcePath, vFrom.Index(i).Interface(), targetPath, target, flattener)...)
		if diags.HasError() {
			return diags
		}

		t.Index(i).Set(reflect.ValueOf(target))
	}

	// Set the target structure as a nested Object.
	val, d := tTo.ValueFromObjectSlice(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

// flattenStruct traverses struct `from`, calling `flexer` for each exported field.
func flattenStruct(ctx context.Context, sourcePath path.Path, from any, targetPath path.Path, to any, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx, valFrom, valTo, d := autoFlexValues(ctx, from, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if toFlattener, ok := to.(Flattener); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Target implements flex.Flattener")
		diags.Append(flattenFlattener(ctx, valFrom, toFlattener)...)
		return diags
	}

	typeFrom := valFrom.Type()
	typeTo := valTo.Type()

	for fromField := range flattenSourceFields(ctx, typeFrom, flexer.getOptions()) {
		fromFieldName := fromField.Name

		toField, ok := findFieldFuzzy(ctx, fromFieldName, typeFrom, typeTo, flexer)
		if !ok {
			// Corresponding field not found in to.
			tflog.SubsystemDebug(ctx, subsystemName, "No corresponding field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
			})
			continue
		}
		toFieldName := toField.Name
		toNameOverride, toOpts := autoflexTags(toField)
		toFieldVal := valTo.FieldByIndex(toField.Index)
		if toNameOverride == "-" {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored target field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
				logAttrKeyTargetFieldname: toFieldName,
			})
			continue
		}
		if toOpts.NoFlatten() {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping noflatten target field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
				logAttrKeyTargetFieldname: toFieldName,
			})
			continue
		}
		if !toFieldVal.CanSet() {
			// Corresponding field value can't be changed.
			tflog.SubsystemDebug(ctx, subsystemName, "Field cannot be set", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
				logAttrKeyTargetFieldname: toFieldName,
			})
			continue
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Matched fields", map[string]any{
			logAttrKeySourceFieldname: fromFieldName,
			logAttrKeyTargetFieldname: toFieldName,
		})

		opts := fieldOpts{
			legacy:    toOpts.Legacy(),
			omitempty: toOpts.OmitEmpty(),
		}

		diags.Append(flexer.convert(ctx, sourcePath.AtName(fromFieldName), valFrom.FieldByIndex(fromField.Index), targetPath.AtName(toFieldName), toFieldVal, opts)...)
		if diags.HasError() {
			break
		}
	}

	return diags
}

func flattenSourceFields(ctx context.Context, typ reflect.Type, opts AutoFlexOptions) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		for field := range tfreflect.ExportedStructFields(typ) {
			fieldName := field.Name
			if opts.isIgnoredField(fieldName) {
				tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored source field", map[string]any{
					logAttrKeySourceFieldname: fieldName,
				})
				continue
			}

			if !yield(field) {
				return
			}
		}
	}
}

// setMapBlockKey takes a struct and assigns the value of the `key`
func setMapBlockKey(ctx context.Context, to any, key reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo := reflect.ValueOf(to)
	if kind := valTo.Kind(); kind == reflect.Pointer {
		valTo = valTo.Elem()
	}

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(valueType(valTo)))

	if valTo.Kind() != reflect.Struct {
		diags.AddError("AutoFlEx", fmt.Sprintf("wrong type (%T), expected struct", valTo))
		return diags
	}

	for field := range tfreflect.ExportedStructFields(valTo.Type()) {
		if field.Name != mapBlockKeyFieldName {
			continue
		}

		fieldVal := valTo.FieldByIndex(field.Index)           // vTo
		fieldAttrVal, ok := fieldVal.Interface().(attr.Value) // valTo
		if !ok {
			tflog.SubsystemError(ctx, subsystemName, "Target does not implement attr.Value")
			diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(fieldVal.Interface())))
			return diags
		}
		tTo := fieldAttrVal.Type(ctx)

		switch tTo := tTo.(type) {
		case basetypes.StringTypable:
			v, d := tTo.ValueFromString(ctx, types.StringValue(key.String()))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			fieldVal.Set(reflect.ValueOf(v))
		}

		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "Target has no map block key")
	diags.Append(diagFlatteningNoMapBlockKey(valTo.Type()))

	return diags
}

type attrValueFromReflectValueFunc func(reflect.Value) (attr.Value, diag.Diagnostics)

func newInt64ValueFromReflectValue(v reflect.Value) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	return types.Int64Value(v.Int()), diags
}

func newInt64ValueFromReflectPointerValue(v reflect.Value) (attr.Value, diag.Diagnostics) {
	if v.IsNil() {
		var diags diag.Diagnostics
		return types.Int64Null(), diags
	}

	return newInt64ValueFromReflectValue(v.Elem())
}

func newStringValueFromReflectValue(v reflect.Value) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	return types.StringValue(v.String()), diags
}

func newStringValueFromReflectPointerValue(v reflect.Value) (attr.Value, diag.Diagnostics) {
	if v.IsNil() {
		var diags diag.Diagnostics
		return types.StringNull(), diags
	}

	return newStringValueFromReflectValue(v.Elem())
}

func flattenFlattener(ctx context.Context, fromVal reflect.Value, toFlattener Flattener) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo := reflect.ValueOf(toFlattener)

	diags.Append(flattenPrePopulate(ctx, valTo)...)
	if diags.HasError() {
		return diags
	}

	from := fromVal.Interface()

	diags.Append(toFlattener.Flatten(ctx, from)...)

	return diags
}

func flattenPrePopulate(ctx context.Context, toVal reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if toVal.Kind() == reflect.Pointer {
		toVal = toVal.Elem()
	}

	for field := range tfreflect.ExportedStructFields(toVal.Type()) {
		fieldVal := toVal.FieldByIndex(field.Index)
		if !fieldVal.CanSet() {
			diags.AddError(
				"Incompatible Types",
				"An unexpected error occurred while flattening configuration. "+
					"This is always an error in the provider. "+
					"Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Field %q in %q is not settable.", field.Name, fullTypeName(fieldVal.Type())),
			)
		}

		fieldTo, ok := fieldVal.Interface().(attr.Value)
		if !ok {
			continue // Skip non-attr.Type fields.
		}
		tTo := fieldTo.Type(ctx)
		switch tTo := tTo.(type) {
		// The zero values of primitive types are Null.

		// Aggregate types.
		case fwtypes.NestedObjectType:
			v, d := tTo.NullValue(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			fieldVal.Set(reflect.ValueOf(v))
		}
	}

	return diags
}

func diagFlatteningSourceIsNil(sourceType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Source of type %q is nil.", fullTypeName(sourceType)),
	)
}

func diagFlatteningTargetDoesNotImplementAttrValue(targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Target type %q does not implement attr.Value", fullTypeName(targetType)),
	)
}

func diagFlatteningNoMapBlockKey(sourceType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Target type %q does not contain field %q", fullTypeName(sourceType), mapBlockKeyFieldName),
	)
}

func diagFlatteningMarshalSmithyDocument(sourceType reflect.Type, err error) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Marshalling JSON document of type %q failed: %s", fullTypeName(sourceType), err.Error()),
	)
}

func DiagFlatteningIncompatibleTypes(sourceType, targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Source type %q cannot be flattened to target type %q.", fullTypeName(sourceType), fullTypeName(targetType)),
	)
}
