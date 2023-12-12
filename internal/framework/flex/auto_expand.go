// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Expand  = TF -->  AWS

// Expand "expands" a resource's "business logic" data structure,
// implemented using Terraform Plugin Framework data types, into
// an AWS SDK for Go v2 API data structure.
// The resource's data structure is walked and exported fields that
// have a corresponding field in the API data structure (and a suitable
// target data type) are copied.
func Expand(ctx context.Context, tfObject, apiObject any, optFns ...AutoFlexOptionsFunc) diag.Diagnostics {
	var diags diag.Diagnostics
	expander := &autoExpander{}

	for _, optFn := range optFns {
		optFn(expander)
	}

	diags.Append(autoFlexConvert(ctx, tfObject, apiObject, expander)...)
	if diags.HasError() {
		diags.AddError("AutoFlEx", fmt.Sprintf("Expand[%T, %T]", tfObject, apiObject))
		return diags
	}

	return diags
}

type autoExpander struct{}

// convert converts a single Plugin Framework value to its AWS API equivalent.
func (expander autoExpander) convert(ctx context.Context, valFrom, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	vFrom, ok := valFrom.Interface().(attr.Value)
	if !ok {
		diags.AddError("AutoFlEx", fmt.Sprintf("does not implement attr.Value: %s", valFrom.Kind()))
		return diags
	}

	// No need to set the target value if there's no source value.
	if vFrom.IsNull() || vFrom.IsUnknown() {
		return diags
	}

	switch vFrom := vFrom.(type) {
	// Primitive types.
	case basetypes.BoolValuable:
		diags.Append(expander.bool(ctx, vFrom, vTo)...)
		return diags

	case basetypes.Float64Valuable:
		diags.Append(expander.float64(ctx, vFrom, vTo)...)
		return diags

	case basetypes.Int64Valuable:
		diags.Append(expander.int64(ctx, vFrom, vTo)...)
		return diags

	case basetypes.StringValuable:
		diags.Append(expander.string(ctx, vFrom, vTo)...)
		return diags

	// Aggregate types.
	case basetypes.ListValuable:
		diags.Append(expander.list(ctx, vFrom, vTo)...)
		return diags

	case basetypes.MapValuable:
		diags.Append(expander.map_(ctx, vFrom, vTo)...)
		return diags

	case basetypes.SetValuable:
		diags.Append(expander.set(ctx, vFrom, vTo)...)
		return diags
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// bool copies a Plugin Framework Bool(ish) value to a compatible AWS API value.
func (expander autoExpander) bool(ctx context.Context, vFrom basetypes.BoolValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToBoolValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch vTo.Kind() {
	case reflect.Bool:
		//
		// types.Bool -> bool.
		//
		vTo.SetBool(v.ValueBool())
		return diags

	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.Bool:
			//
			// types.Bool -> *bool.
			//
			vTo.Set(reflect.ValueOf(v.ValueBoolPointer()))
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// float64 copies a Plugin Framework Float64(ish) value to a compatible AWS API value.
func (expander autoExpander) float64(ctx context.Context, vFrom basetypes.Float64Valuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToFloat64Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch vTo.Kind() {
	case reflect.Float32, reflect.Float64:
		//
		// types.Float32/types.Float64 -> float32/float64.
		//
		vTo.SetFloat(v.ValueFloat64())
		return diags

	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.Float32:
			//
			// types.Float32/types.Float64 -> *float32.
			//
			to := float32(v.ValueFloat64())
			vTo.Set(reflect.ValueOf(&to))
			return diags

		case reflect.Float64:
			//
			// types.Float32/types.Float64 -> *float64.
			//
			vTo.Set(reflect.ValueOf(v.ValueFloat64Pointer()))
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// int64 copies a Plugin Framework Int64(ish) value to a compatible AWS API value.
func (expander autoExpander) int64(ctx context.Context, vFrom basetypes.Int64Valuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToInt64Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch vTo.Kind() {
	case reflect.Int32, reflect.Int64:
		//
		// types.Int32/types.Int64 -> int32/int64.
		//
		vTo.SetInt(v.ValueInt64())
		return diags

	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.Int32:
			//
			// types.Int32/types.Int64 -> *int32.
			//
			to := int32(v.ValueInt64())
			vTo.Set(reflect.ValueOf(&to))
			return diags

		case reflect.Int64:
			//
			// types.Int32/types.Int64 -> *int64.
			//
			vTo.Set(reflect.ValueOf(v.ValueInt64Pointer()))
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// string copies a Plugin Framework String(ish) value to a compatible AWS API value.
func (expander autoExpander) string(ctx context.Context, vFrom basetypes.StringValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch vTo.Kind() {
	case reflect.String:
		//
		// types.String -> string.
		//
		vTo.SetString(v.ValueString())
		return diags
	case reflect.Struct:
		//
		// fwtypes.Timestamp --> time.Time
		//
		if t, ok := vFrom.(fwtypes.Timestamp); ok {
			vTo.Set(reflect.ValueOf(t.ValueTimestamp()))
			return diags
		}
	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.String:
			//
			// types.String -> *string.
			//
			vTo.Set(reflect.ValueOf(v.ValueStringPointer()))
			return diags
		case reflect.Struct:
			//
			// fwtypes.Timestamp --> *time.Time
			//
			if t, ok := vFrom.(fwtypes.Timestamp); ok {
				vTo.Set(reflect.ValueOf(t.ValueTimestampPointer()))
				return diags
			}
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// list copies a Plugin Framework List(ish) value to a compatible AWS API value.
func (expander autoExpander) list(ctx context.Context, vFrom basetypes.ListValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToListValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.StringTypable:
		diags.Append(expander.listOfString(ctx, v, vTo)...)
		return diags

	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.NestedObjectValue); ok {
			diags.Append(expander.nestedObject(ctx, vFrom, vTo)...)
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from list[%s]": v.ElementType(ctx),
		"to":            vTo.Kind(),
	})

	return diags
}

// listOfString copies a Plugin Framework ListOfString(ish) value to a compatible AWS API value.
func (expander autoExpander) listOfString(ctx context.Context, vFrom basetypes.ListValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Slice:
		switch tSliceElem := vTo.Type().Elem(); tSliceElem.Kind() {
		case reflect.String:
			//
			// types.List(OfString) -> []string.
			//
			var to []string
			diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags

		case reflect.Ptr:
			switch tSliceElem.Elem().Kind() {
			case reflect.String:
				//
				// types.List(OfString) -> []*string.
				//
				var to []*string
				diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from list[%s]": vFrom.ElementType(ctx),
		"to":            vTo.Kind(),
	})

	return diags
}

// map_ copies a Plugin Framework Map(ish) value to a compatible AWS API value.
func (expander autoExpander) map_(ctx context.Context, vFrom basetypes.MapValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToMapValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.StringTypable:
		diags.Append(expander.mapOfString(ctx, v, vTo)...)
		return diags
	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.ObjectMapValue); ok {
			diags.Append(expander.objectMap(ctx, vFrom, vTo)...)
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from map[string, %s]": v.ElementType(ctx),
		"to":                   vTo.Kind(),
	})

	return diags
}

// mapOfString copies a Plugin Framework MapOfString(ish) value to a compatible AWS API value.
func (expander autoExpander) mapOfString(ctx context.Context, vFrom basetypes.MapValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Map:
		switch tMapKey := vTo.Type().Key(); tMapKey.Kind() {
		case reflect.String: // key
			switch tMapElem := vTo.Type().Elem(); tMapElem.Kind() {
			case reflect.String:
				//
				// types.Map(OfString) -> map[string]string.
				//
				var to map[string]string
				diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags

			case reflect.Ptr:
				switch k := tMapElem.Elem().Kind(); k {
				case reflect.String:
					//
					// types.Map(OfString) -> map[string]*string.
					//
					var to map[string]*string
					diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}
			}
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from map[string, %s]": vFrom.ElementType(ctx),
		"to":                   vTo.Kind(),
	})

	return diags
}

// set copies a Plugin Framework Set(ish) value to a compatible AWS API value.
func (expander autoExpander) set(ctx context.Context, vFrom basetypes.SetValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToSetValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.StringTypable:
		diags.Append(expander.setOfString(ctx, v, vTo)...)
		return diags

	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.NestedObjectValue); ok {
			diags.Append(expander.nestedObject(ctx, vFrom, vTo)...)
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from set[%s]": v.ElementType(ctx),
		"to":           vTo.Kind(),
	})

	return diags
}

// setOfString copies a Plugin Framework SetOfString(ish) value to a compatible AWS API value.
func (expander autoExpander) setOfString(ctx context.Context, vFrom basetypes.SetValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Slice:
		switch tSliceElem := vTo.Type().Elem(); tSliceElem.Kind() {
		case reflect.String:
			//
			// types.Set(OfString) -> []string.
			//
			var to []string
			diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags

		case reflect.Ptr:
			switch tSliceElem.Elem().Kind() {
			case reflect.String:
				//
				// types.Set(OfString) -> []*string.
				//
				var to []*string
				diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from set[%s]": vFrom.ElementType(ctx),
		"to":           vTo.Kind(),
	})

	return diags
}

// nestedObject copies a Plugin Framework NestedObjectValue value to a compatible AWS API value.
func (expander autoExpander) nestedObject(ctx context.Context, vFrom fwtypes.NestedObjectValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Struct:
		diags.Append(expander.nestedObjectToStruct(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Ptr:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			//
			// types.List(OfObject) -> *struct.
			//
			diags.Append(expander.nestedObjectToStruct(ctx, vFrom, tElem, vTo)...)
			return diags
		}

	case reflect.Slice:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			//
			// types.List(OfObject) -> []struct.
			//
			diags.Append(expander.nestedObjectToSlice(ctx, vFrom, tTo, tElem, vTo)...)
			return diags

		case reflect.Ptr:
			switch tElem := tElem.Elem(); tElem.Kind() {
			case reflect.Struct:
				//
				// types.List(OfObject) -> []*struct.
				//
				diags.Append(expander.nestedObjectToSlice(ctx, vFrom, tTo, tElem, vTo)...)
				return diags
			}
		}
	}

	diags.AddError("Incompatible types", fmt.Sprintf("nestedObject[%s] cannot be expanded to %s", vFrom.Type(ctx).(attr.TypeWithElementType).ElementType(), vTo.Kind()))
	return diags
}

// nestedObjectToStruct copies a Plugin Framework NestedObjectValue to a compatible AWS API (*)struct value.
func (expander autoExpander) nestedObjectToStruct(ctx context.Context, vFrom fwtypes.NestedObjectValue, tStruct reflect.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the nested Object as a pointer.
	from, d := vFrom.ToObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Create a new target structure and walk its fields.
	to := reflect.New(tStruct)
	diags.Append(autoFlexConvertStruct(ctx, from, to.Interface(), expander)...)
	if diags.HasError() {
		return diags
	}

	// Set value (or pointer).
	if vTo.Type().Kind() == reflect.Struct {
		vTo.Set(to.Elem())
	} else {
		vTo.Set(to)
	}

	return diags
}

// nestedObjectToSlice copies a Plugin Framework NestedObjectValue to a compatible AWS API [](*)struct value.
func (expander autoExpander) nestedObjectToSlice(ctx context.Context, vFrom fwtypes.NestedObjectValue, tSlice, tElem reflect.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the nested Objects as a slice.
	from, d := vFrom.ToObjectSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Create a new target slice and expand each element.
	f := reflect.ValueOf(from)
	n := f.Len()
	t := reflect.MakeSlice(tSlice, n, n)
	for i := 0; i < n; i++ {
		// Create a new target structure and walk its fields.
		target := reflect.New(tElem)
		diags.Append(autoFlexConvertStruct(ctx, f.Index(i).Interface(), target.Interface(), expander)...)
		if diags.HasError() {
			return diags
		}

		// Set value (or pointer) in the target slice.
		if vTo.Type().Elem().Kind() == reflect.Struct {
			t.Index(i).Set(target.Elem())
		} else {
			t.Index(i).Set(target)
		}
	}

	vTo.Set(t)

	return diags
}

// objectMap copies a Plugin Framework ObjectMapValue value to a compatible AWS API value.
func (expander autoExpander) objectMap(ctx context.Context, vFrom fwtypes.ObjectMapValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Map:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Ptr:
			diags.Append(expander.mappedObjectToStruct(ctx, vFrom, tElem, vTo)...)
			return diags
		case reflect.Struct:
			diags.Append(expander.mappedObjectToStruct(ctx, vFrom, tElem, vTo)...)
			return diags
		}
	case reflect.Ptr:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			diags.Append(expander.mappedObjectToStruct(ctx, vFrom, tElem, vTo)...)
			return diags
		}
	}

	diags.AddError("Incompatible types", fmt.Sprintf("objectMap[%s] cannot be expanded to %s", vFrom.Type(ctx).(attr.TypeWithElementType).ElementType(), vTo.Kind()))
	return diags
}

// mappedObjectToStruct copies a Plugin Framework ObjectMapValue to a compatible AWS API (*)struct value.
func (expander autoExpander) mappedObjectToStruct(ctx context.Context, vFrom fwtypes.ObjectMapValue, tStruct reflect.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the mapped Objects as a map
	from, d := vFrom.ToObjectMap(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if tStruct.Kind() == reflect.Ptr {
		tStruct = tStruct.Elem()
	}

	f := reflect.ValueOf(from)
	m := reflect.MakeMap(vTo.Type())

	for _, key := range f.MapKeys() {
		// Create a new target map and populate it
		target := reflect.New(tStruct)

		fromInterface := f.MapIndex(key).Interface()
		if f.MapIndex(key).Kind() == reflect.Ptr {
			fromInterface = f.MapIndex(key).Elem().Interface()
		}

		diags.Append(autoFlexConvertStruct(ctx, fromInterface, target.Interface(), expander)...)
		if diags.HasError() {
			return diags
		}

		// Set value (or pointer) in the target slice.
		if vTo.Type().Elem().Kind() == reflect.Struct {
			m.SetMapIndex(key, target.Elem())
		} else {
			m.SetMapIndex(key, target)
		}
	}

	vTo.Set(m)

	return diags
}
