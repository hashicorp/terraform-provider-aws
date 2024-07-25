// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
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

	diags.Append(autoFlexConvert(ctx, apiObject, tfObject, flattener)...)
	if diags.HasError() {
		diags.AddError("AutoFlEx", fmt.Sprintf("Flatten[%T, %T]", apiObject, tfObject))
		return diags
	}

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

// convert converts a single AWS API value to its Plugin Framework equivalent.
func (flattener autoFlattener) convert(ctx context.Context, vFrom, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo, ok := vTo.Interface().(attr.Value)
	if !ok {
		// Check for `nil` (i.e. Kind == Invalid) here, because primitive types can be `nil`
		if vFrom.Kind() == reflect.Invalid {
			diags.AddError("AutoFlEx", "Cannot flatten nil source")
			return diags
		}

		diags.AddError("AutoFlEx", fmt.Sprintf("does not implement attr.Value: %s", vTo.Kind()))
		return diags
	}

	tTo := valTo.Type(ctx)
	switch k := vFrom.Kind(); k {
	case reflect.Bool:
		diags.Append(flattener.bool(ctx, vFrom, false, tTo, vTo)...)
		return diags

	case reflect.Float32, reflect.Float64:
		diags.Append(flattener.float(ctx, vFrom, false, tTo, vTo)...)
		return diags

	case reflect.Int32, reflect.Int64:
		diags.Append(flattener.int(ctx, vFrom, false, tTo, vTo)...)
		return diags

	case reflect.String:
		diags.Append(flattener.string(ctx, vFrom, false, tTo, vTo)...)
		return diags

	case reflect.Ptr:
		diags.Append(flattener.ptr(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Slice:
		diags.Append(flattener.slice(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Map:
		diags.Append(flattener.map_(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Struct:
		diags.Append(flattener.struct_(ctx, vFrom, false, tTo, vTo)...)
		return diags

	case reflect.Interface:
		diags.Append(flattener.interface_(ctx, vFrom, false, tTo, vTo)...)
		return diags
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// bool copies an AWS API bool value to a compatible Plugin Framework value.
func (flattener autoFlattener) bool(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.BoolTypable:
		boolValue := types.BoolNull()
		if !isNullFrom {
			boolValue = types.BoolValue(vFrom.Bool())
		}
		v, d := tTo.ValueFromBool(ctx, boolValue)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// bool -> types.Bool.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// float copies an AWS API float value to a compatible Plugin Framework value.
func (flattener autoFlattener) float(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Float64Typable:
		float64Value := types.Float64Null()
		if !isNullFrom {
			switch from := vFrom.Interface().(type) {
			// Avoid loss of equivalence.
			case float32:
				float64Value = types.Float64Value(decimal.NewFromFloat32(from).InexactFloat64())
			default:
				float64Value = types.Float64Value(vFrom.Float())
			}
		}
		v, d := tTo.ValueFromFloat64(ctx, float64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// float32/float64 -> types.Float64.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// int copies an AWS API int value to a compatible Plugin Framework value.
func (flattener autoFlattener) int(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Int64Typable:
		int64Value := types.Int64Null()
		if !isNullFrom {
			int64Value = types.Int64Value(vFrom.Int())
		}
		v, d := tTo.ValueFromInt64(ctx, int64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// int32/int64 -> types.Int64.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// string copies an AWS API string value to a compatible Plugin Framework value.
func (flattener autoFlattener) string(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.StringTypable:
		stringValue := types.StringNull()
		if !isNullFrom {
			// If the target is a StringEnumType, an empty string value is converted to a null String.
			if value := vFrom.String(); !strings.HasPrefix(tTo.String(), "StringEnumType[") || value != "" {
				stringValue = types.StringValue(value)
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

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
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

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   vTo,
	})

	return diags
}

// ptr copies an AWS API pointer value to a compatible Plugin Framework value.
func (flattener autoFlattener) ptr(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vElem, isNilFrom := vFrom.Elem(), vFrom.IsNil(); vFrom.Type().Elem().Kind() {
	case reflect.Bool:
		diags.Append(flattener.bool(ctx, vElem, isNilFrom, tTo, vTo)...)
		return diags

	case reflect.Float32, reflect.Float64:
		diags.Append(flattener.float(ctx, vElem, isNilFrom, tTo, vTo)...)
		return diags

	case reflect.Int32, reflect.Int64:
		diags.Append(flattener.int(ctx, vElem, isNilFrom, tTo, vTo)...)
		return diags

	case reflect.String:
		diags.Append(flattener.string(ctx, vElem, isNilFrom, tTo, vTo)...)
		return diags

	case reflect.Struct:
		diags.Append(flattener.struct_(ctx, vElem, isNilFrom, tTo, vTo)...)
		return diags
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

func (flattener autoFlattener) interface_(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.StringTypable:
		stringValue := types.StringNull()
		if !isNullFrom {
			//
			// JSONStringer -> types.String-ish.
			//
			if vFrom.Type().Implements(reflect.TypeOf((*smithyjson.JSONStringer)(nil)).Elem()) {
				doc := vFrom.Interface().(smithyjson.JSONStringer)
				b, err := doc.MarshalSmithyDocument()
				if err != nil {
					diags.AddError("AutoFlEx", err.Error())
					return diags
				}
				stringValue = types.StringValue(string(b))
			}
		}
		v, d := tTo.ValueFromString(ctx, stringValue)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case fwtypes.NestedObjectType:
		//
		// interface -> types.List(OfObject) or types.Object.
		//
		diags.Append(flattener.interfaceToNestedObject(ctx, vFrom, vFrom.IsNil(), tTo, vTo)...)
		return diags
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// struct_ copies an AWS API struct value to a compatible Plugin Framework value.
func (flattener autoFlattener) struct_(ctx context.Context, vFrom reflect.Value, isNilFrom bool, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
		//
		// *struct -> types.List(OfObject) or types.Object.
		//
		diags.Append(flattener.structToNestedObject(ctx, vFrom, isNilFrom, tTo, vTo)...)
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

	return diags
}

// slice copies an AWS API slice value to a compatible Plugin Framework value.
func (flattener autoFlattener) slice(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tSliceElem := vFrom.Type().Elem(); tSliceElem.Kind() {
	case reflect.Int32, reflect.Int64:
		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []int32 or []int64 -> types.List(OfInt64).
			//
			diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectValue)...)
			return diags

		case basetypes.SetTypable:
			//
			// []int32 or []int64 -> types.Set(OfInt64).
			//
			diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectValue)...)
			return diags
		}

	case reflect.String:
		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []string -> types.List(OfString).
			//
			diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.StringType, newStringValueFromReflectValue)...)
			return diags

		case basetypes.SetTypable:
			//
			// []string -> types.Set(OfString).
			//
			diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.StringType, newStringValueFromReflectValue)...)
			return diags
		}

	case reflect.Ptr:
		switch tSliceElem.Elem().Kind() {
		case reflect.Int32:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*int32 -> types.List(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*int32 -> types.Set(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue)...)
				return diags
			}

		case reflect.Int64:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*int64 -> types.List(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*int64 -> types.Set(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue)...)
				return diags
			}

		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*string -> types.List(OfString).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.StringType, newStringValueFromReflectPointerValue)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*string -> types.Set(OfString).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.StringType, newStringValueFromReflectPointerValue)...)
				return diags
			}

		case reflect.Struct:
			if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
				//
				// []*struct -> types.List(OfObject).
				//
				diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, vFrom, tTo, vTo)...)
				return diags
			}
		}

	case reflect.Struct:
		if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
			//
			// []struct -> types.List(OfObject).
			//
			diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, vFrom, tTo, vTo)...)
			return diags
		}

	case reflect.Interface:
		if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
			//
			// []interface -> types.List(OfObject).
			//
			diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, vFrom, tTo, vTo)...)
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// map_ copies an AWS API map value to a compatible Plugin Framework value.
func (flattener autoFlattener) map_(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
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
					diags.Append(flattener.structMapToObjectList(ctx, vFrom, tTo, vTo)...)
					return diags
				}

			case basetypes.ListTypable:
				//
				// map[string]struct -> fwtypes.ListNestedObjectOf[Object]
				//
				if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
					diags.Append(flattener.structMapToObjectList(ctx, vFrom, tTo, vTo)...)
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
					to, d := tTo.ValueFromMap(ctx, types.MapNull(types.StringType))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}

				from := vFrom.Interface().(map[string]string)
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
					elements := make(map[string]attr.Value, len(from))
					for k, v := range from {
						innerElements := make(map[string]attr.Value, len(v))
						for ik, iv := range v {
							innerElements[ik] = types.StringValue(iv)
						}
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

				case reflect.Ptr:
					from := vFrom.Interface().(map[string]map[string]*string)
					elements := make(map[string]attr.Value, len(from))
					for k, v := range from {
						innerElements := make(map[string]attr.Value, len(v))
						for ik, iv := range v {
							innerElements[ik] = types.StringValue(*iv)
						}
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

		case reflect.Ptr:
			switch tMapElem.Elem().Kind() {
			case reflect.Struct:
				if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
					diags.Append(flattener.structMapToObjectList(ctx, vFrom, tTo, vTo)...)
					return diags
				}

			case reflect.String:
				switch tTo := tTo.(type) {
				case basetypes.ListTypable:
					//
					// map[string]struct -> fwtypes.ListNestedObjectOf[Object]
					//
					if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
						diags.Append(flattener.structMapToObjectList(ctx, vFrom, tTo, vTo)...)
						return diags
					}

				case basetypes.MapTypable:
					//
					// map[string]*string -> types.Map(OfString).
					//
					if vFrom.IsNil() {
						to, d := tTo.ValueFromMap(ctx, types.MapNull(types.StringType))
						diags.Append(d...)
						if diags.HasError() {
							return diags
						}

						vTo.Set(reflect.ValueOf(to))
						return diags
					}

					from := vFrom.Interface().(map[string]*string)
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

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

func (flattener autoFlattener) structMapToObjectList(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if vFrom.IsNil() {
		val, d := tTo.NullValue(ctx)
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
		target, d := tTo.NewObjectPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		fromInterface := vFrom.MapIndex(key).Interface()
		if vFrom.MapIndex(key).Kind() == reflect.Ptr {
			fromInterface = vFrom.MapIndex(key).Elem().Interface()
		}

		diags.Append(autoFlexConvertStruct(ctx, fromInterface, target, flattener)...)
		if diags.HasError() {
			return diags
		}

		d = blockKeyMapSet(target, key)
		diags.Append(d...)

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
func (flattener autoFlattener) structToNestedObject(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo fwtypes.NestedObjectType, vTo reflect.Value) diag.Diagnostics {
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

	// Create a new target structure and walk its fields.
	to, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(autoFlexConvertStruct(ctx, vFrom.Interface(), to, flattener)...)
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

		tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]any{
			"from": vFrom.Kind(),
			"to":   tTo,
		})
		return diags
	}

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
func (flattener autoFlattener) sliceOfPrimtiveToList(ctx context.Context, vFrom reflect.Value, tTo basetypes.ListTypable, vTo reflect.Value, elementType attr.Type, f attrValueFromReflectValueFunc) diag.Diagnostics {
	var diags diag.Diagnostics

	if vFrom.IsNil() {
		to, d := tTo.ValueFromList(ctx, types.ListNull(elementType))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(to))
		return diags
	}

	elements := make([]attr.Value, vFrom.Len())
	for i := 0; i < vFrom.Len(); i++ {
		elements[i] = f(vFrom.Index(i))
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
func (flattener autoFlattener) sliceOfPrimitiveToSet(ctx context.Context, vFrom reflect.Value, tTo basetypes.SetTypable, vTo reflect.Value, elementType attr.Type, f attrValueFromReflectValueFunc) diag.Diagnostics {
	var diags diag.Diagnostics

	if vFrom.IsNil() {
		to, d := tTo.ValueFromSet(ctx, types.SetNull(elementType))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(to))
		return diags
	}

	elements := make([]attr.Value, vFrom.Len())
	for i := 0; i < vFrom.Len(); i++ {
		elements[i] = f(vFrom.Index(i))
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
func (flattener autoFlattener) sliceOfStructToNestedObjectCollection(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if vFrom.IsNil() {
		val, d := tTo.NullValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))
		return diags
	}

	// Create a new target slice and flatten each element.
	n := vFrom.Len()
	to, d := tTo.NewObjectSlice(ctx, n, n)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	t := reflect.ValueOf(to)
	for i := 0; i < n; i++ {
		target, d := tTo.NewObjectPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		diags.Append(autoFlexConvertStruct(ctx, vFrom.Index(i).Interface(), target, flattener)...)
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

// blockKeyMapSet takes a struct and assigns the value of the `key`
func blockKeyMapSet(to any, key reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo := reflect.ValueOf(to)
	if kind := valTo.Kind(); kind == reflect.Ptr {
		valTo = valTo.Elem()
	}

	if valTo.Kind() != reflect.Struct {
		diags.AddError("AutoFlEx", fmt.Sprintf("wrong type (%T), expected struct", valTo))
		return diags
	}

	for i, typTo := 0, valTo.Type(); i < typTo.NumField(); i++ {
		field := typTo.Field(i)
		if field.PkgPath != "" {
			continue // Skip unexported fields.
		}

		if field.Name != MapBlockKey {
			continue
		}

		if _, ok := valTo.Field(i).Interface().(basetypes.StringValue); ok {
			valTo.Field(i).Set(reflect.ValueOf(basetypes.NewStringValue(key.String())))
			return diags
		}

		fieldType := valTo.Field(i).Type()

		method, found := fieldType.MethodByName("StringEnumValue")
		if found {
			result := fieldType.Method(method.Index).Func.Call([]reflect.Value{valTo.Field(i), key})
			if len(result) > 0 {
				valTo.Field(i).Set(result[0])
			}
		}

		return diags
	}

	diags.AddError("AutoFlEx", fmt.Sprintf("unable to find map block key (%s)", MapBlockKey))

	return diags
}

type attrValueFromReflectValueFunc func(reflect.Value) attr.Value

func newInt64ValueFromReflectValue(v reflect.Value) attr.Value {
	return types.Int64Value(v.Int())
}

func newInt64ValueFromReflectPointerValue(v reflect.Value) attr.Value {
	if v.IsNil() {
		return types.Int64Null()
	}

	return newInt64ValueFromReflectValue(v.Elem())
}

func newStringValueFromReflectValue(v reflect.Value) attr.Value {
	return types.StringValue(v.String())
}

func newStringValueFromReflectPointerValue(v reflect.Value) attr.Value {
	if v.IsNil() {
		return types.StringNull()
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

	if toVal.Kind() == reflect.Ptr {
		toVal = toVal.Elem()
	}

	typeTo := toVal.Type()
	for i := 0; i < typeTo.NumField(); i++ {
		field := typeTo.Field(i)
		if !field.IsExported() {
			continue // Skip unexported fields.
		}

		fieldVal := toVal.Field(i)
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
