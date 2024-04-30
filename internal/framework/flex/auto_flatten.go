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
)

// Flatten = AWS --> TF

// Flatten "flattens" an AWS SDK for Go v2 API data structure into
// a resource's "business logic" data structure, implemented using
// Terraform Plugin Framework data types.
// The API data structure's fields are walked and exported fields that
// have a corresponding field in the resource's data structure (and a
// suitable target data type) are copied.
func Flatten(ctx context.Context, apiObject, tfObject any, optFns ...AutoFlexOptionsFunc) diag.Diagnostics {
	var diags diag.Diagnostics
	flattener := &autoFlattener{}

	for _, optFn := range optFns {
		optFn(flattener)
	}

	diags.Append(autoFlexConvert(ctx, apiObject, tfObject, flattener)...)
	if diags.HasError() {
		diags.AddError("AutoFlEx", fmt.Sprintf("Flatten[%T, %T]", apiObject, tfObject))
		return diags
	}

	return diags
}

type autoFlattener struct{}

// convert converts a single AWS API value to its Plugin Framework equivalent.
func (flattener autoFlattener) convert(ctx context.Context, vFrom, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo, ok := vTo.Interface().(attr.Value)
	if !ok {
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
		// Smithy union type handling not yet implemented. Silently skip.
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
			float64Value = types.Float64Value(vFrom.Float())
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
	case reflect.String:
		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []string -> types.List(OfString).
			//
			if vFrom.IsNil() {
				to, d := tTo.ValueFromList(ctx, types.ListNull(types.StringType))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}

			elements := make([]attr.Value, vFrom.Len())
			for i := 0; i < vFrom.Len(); i++ {
				elements[i] = types.StringValue(vFrom.Index(i).String())
			}
			list, d := types.ListValue(types.StringType, elements)
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

		case basetypes.SetTypable:
			//
			// []string -> types.Set(OfString).
			//
			if vFrom.IsNil() {
				to, d := tTo.ValueFromSet(ctx, types.SetNull(types.StringType))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}

			elements := make([]attr.Value, vFrom.Len())
			for i := 0; i < vFrom.Len(); i++ {
				elements[i] = types.StringValue(vFrom.Index(i).String())
			}
			set, d := types.SetValue(types.StringType, elements)
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

	case reflect.Ptr:
		switch tSliceElem.Elem().Kind() {
		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*string -> types.List(OfString).
				//
				if vFrom.IsNil() {
					to, d := tTo.ValueFromList(ctx, types.ListNull(types.StringType))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}

				from := vFrom.Interface().([]*string)
				elements := make([]attr.Value, len(from))
				for i, v := range from {
					elements[i] = types.StringPointerValue(v)
				}
				list, d := types.ListValue(types.StringType, elements)
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

			case basetypes.SetTypable:
				//
				// []string -> types.Set(OfString).
				//
				if vFrom.IsNil() {
					to, d := tTo.ValueFromSet(ctx, types.SetNull(types.StringType))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}

				from := vFrom.Interface().([]*string)
				elements := make([]attr.Value, len(from))
				for i, v := range from {
					elements[i] = types.StringPointerValue(v)
				}
				set, d := types.SetValue(types.StringType, elements)
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

		case reflect.Struct:
			if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
				//
				// []*struct -> types.List(OfObject).
				//
				diags.Append(flattener.sliceOfStructNestedObjectCollection(ctx, vFrom, tTo, vTo)...)
				return diags
			}
		}

	case reflect.Struct:
		if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
			//
			// []struct -> types.List(OfObject).
			//
			diags.Append(flattener.sliceOfStructNestedObjectCollection(ctx, vFrom, tTo, vTo)...)
			return diags
		}

	case reflect.Interface:
		// Smithy union type handling not yet implemented. Silently skip.
		return diags
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

// sliceOfStructNestedObjectCollection copies an AWS API []struct value to a compatible Plugin Framework NestedObjectCollectionValue value.
func (flattener autoFlattener) sliceOfStructNestedObjectCollection(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value) diag.Diagnostics {
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
