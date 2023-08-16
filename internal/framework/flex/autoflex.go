// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

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

	diags.Append(walkStructFields(ctx, tfObject, apiObject, expandVisitor{})...)
	if diags.HasError() {
		diags.AddError("AutoFlEx", fmt.Sprintf("Expand[%T, %T]", tfObject, apiObject))
		return diags
	}

	return diags
}

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

	diags.Append(walkStructFields(ctx, apiObject, tfObject, flattenVisitor{})...)
	if diags.HasError() {
		diags.AddError("AutoFlEx", fmt.Sprintf("Flatten[%T, %T]", apiObject, tfObject))
		return diags
	}

	return diags
}

// autoFlexer is the interface implemented by an auto-flattener or expander.
type autoFlexer interface{}

// AutoFlexOptionsFunc is a type alias for an autoFlexer functional option.
type AutoFlexOptionsFunc func(autoFlexer)

type autoExpander struct{}

type autoFlattener struct{}

// walkStructFields traverses `from` calling `visitor` for each exported field.
func walkStructFields(ctx context.Context, from any, to any, visitor fieldVisitor) diag.Diagnostics {
	var diags diag.Diagnostics

	valFrom, valTo := reflect.ValueOf(from), reflect.ValueOf(to)
	if kind := valFrom.Kind(); kind == reflect.Ptr {
		valFrom = valFrom.Elem()
	}
	if kind := valTo.Kind(); kind != reflect.Ptr {
		diags.AddError("AutoFlEx", fmt.Sprintf("target (%T): %s, want pointer", to, kind))
		return diags
	}
	valTo = valTo.Elem()

	typFrom, typTo := valFrom.Type(), valTo.Type()
	if typFrom.Kind() != reflect.Struct {
		diags.AddError("AutoFlEx", fmt.Sprintf("source: %s, want struct", typFrom))
		return diags
	}
	if typTo.Kind() != reflect.Struct {
		diags.AddError("AutoFlEx", fmt.Sprintf("target: %s, want struct", typTo))
		return diags
	}

	for i := 0; i < typFrom.NumField(); i++ {
		field := typFrom.Field(i)
		if field.PkgPath != "" {
			continue // Skip unexported fields.
		}
		fieldName := field.Name
		if fieldName == "Tags" {
			continue // Resource tags are handled separately.
		}
		toFieldVal := valTo.FieldByName(fieldName)
		if !toFieldVal.IsValid() {
			continue // Corresponding field not found in to.
		}
		if !toFieldVal.CanSet() {
			continue // Corresponding field value can't be changed.
		}
		diags.Append(visitor.visit(ctx, fieldName, valFrom.Field(i), toFieldVal)...)
		if diags.HasError() {
			diags.AddError("AutoFlEx", fmt.Sprintf("visit (%s)", fieldName))
			return diags
		}
	}

	return diags
}

type fieldVisitor interface {
	visit(context.Context, string, reflect.Value, reflect.Value) diag.Diagnostics
}

type expandVisitor struct{}

// visit copies a single Plugin Framework structure field value to its AWS API equivalent.
func (visitor expandVisitor) visit(ctx context.Context, fieldName string, valFrom, vTo reflect.Value) diag.Diagnostics {
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
		diags.Append(visitor.bool(ctx, vFrom, vTo)...)
		return diags

	case basetypes.Float64Valuable:
		diags.Append(visitor.float64(ctx, vFrom, vTo)...)
		return diags

	case basetypes.Int64Valuable:
		diags.Append(visitor.int64(ctx, vFrom, vTo)...)
		return diags

	case basetypes.StringValuable:
		diags.Append(visitor.string(ctx, vFrom, vTo)...)
		return diags

	// Aggregate types.
	case basetypes.ListValuable:
		diags.Append(visitor.list(ctx, vFrom, vTo)...)
		return diags

	case basetypes.MapValuable:
		diags.Append(visitor.map_(ctx, vFrom, vTo)...)
		return diags

	case basetypes.SetValuable:
		diags.Append(visitor.set(ctx, vFrom, vTo)...)
		return diags
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// bool copies a Plugin Framework Bool(ish) value to a compatible AWS API field.
func (visitor expandVisitor) bool(ctx context.Context, vFrom basetypes.BoolValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToBoolValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch vFrom := v.ValueBool(); vTo.Kind() {
	case reflect.Bool:
		//
		// types.Bool -> bool.
		//
		vTo.SetBool(vFrom)
		return diags

	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.Bool:
			//
			// types.Bool -> *bool.
			//
			vTo.Set(reflect.ValueOf(aws.Bool(vFrom)))
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// float64 copies a Plugin Framework Float64(ish) value to a compatible AWS API field.
func (visitor expandVisitor) float64(ctx context.Context, vFrom basetypes.Float64Valuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToFloat64Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch vFrom := v.ValueFloat64(); vTo.Kind() {
	case reflect.Float32, reflect.Float64:
		//
		// types.Float32/types.Float64 -> float32/float64.
		//
		vTo.SetFloat(vFrom)
		return diags

	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.Float32:
			//
			// types.Float32/types.Float64 -> *float32.
			//
			vTo.Set(reflect.ValueOf(aws.Float32(float32(vFrom))))
			return diags

		case reflect.Float64:
			//
			// types.Float32/types.Float64 -> *float64.
			//
			vTo.Set(reflect.ValueOf(aws.Float64(vFrom)))
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// int64 copies a Plugin Framework Int64(ish) value to a compatible AWS API field.
func (visitor expandVisitor) int64(ctx context.Context, vFrom basetypes.Int64Valuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToInt64Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch vFrom := v.ValueInt64(); vTo.Kind() {
	case reflect.Int32, reflect.Int64:
		//
		// types.Int32/types.Int64 -> int32/int64.
		//
		vTo.SetInt(vFrom)
		return diags

	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.Int32:
			//
			// types.Int32/types.Int64 -> *int32.
			//
			vTo.Set(reflect.ValueOf(aws.Int32(int32(vFrom))))
			return diags

		case reflect.Int64:
			//
			// types.Int32/types.Int64 -> *int64.
			//
			vTo.Set(reflect.ValueOf(aws.Int64(vFrom)))
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// string copies a Plugin Framework String(ish) value to a compatible AWS API field.
func (visitor expandVisitor) string(ctx context.Context, vFrom basetypes.StringValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch vFrom := v.ValueString(); vTo.Kind() {
	case reflect.String:
		//
		// types.String -> string.
		//
		vTo.SetString(vFrom)
		return diags

	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.String:
			//
			// types.String -> *string.
			//
			vTo.Set(reflect.ValueOf(aws.String(vFrom)))
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// list copies a Plugin Framework List(ish) value to a compatible AWS API field.
func (visitor expandVisitor) list(ctx context.Context, vFrom basetypes.ListValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToListValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.StringTypable:
		diags.Append(visitor.listOfString(ctx, v, vTo)...)
		return diags

	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.NestedObjectValue); ok {
			diags.Append(visitor.nestedObject(ctx, vFrom, vTo)...)
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from list[%s]": v.ElementType(ctx),
		"to":            vTo.Kind(),
	})

	return diags
}

// listOfString copies a Plugin Framework ListOfString(ish) value to a compatible AWS API field.
func (visitor expandVisitor) listOfString(ctx context.Context, vFrom basetypes.ListValue, vTo reflect.Value) diag.Diagnostics { //nolint:unparam
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

// map_ copies a Plugin Framework Map(ish) value to a compatible AWS API field.
func (visitor expandVisitor) map_(ctx context.Context, vFrom basetypes.MapValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToMapValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.StringTypable:
		diags.Append(visitor.mapOfString(ctx, v, vTo)...)
		return diags
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from map[string, %s]": v.ElementType(ctx),
		"to":                   vTo.Kind(),
	})

	return diags
}

// mapOfString copies a Plugin Framework MapOfString(ish) value to a compatible AWS API field.
func (visitor expandVisitor) mapOfString(ctx context.Context, vFrom basetypes.MapValue, vTo reflect.Value) diag.Diagnostics { //nolint:unparam
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Map:
		switch tMapKey := vTo.Type().Key(); tMapKey.Kind() {
		case reflect.String:
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
				switch tMapElem.Elem().Kind() {
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

// set copies a Plugin Framework Set(ish) value to a compatible AWS API field.
func (visitor expandVisitor) set(ctx context.Context, vFrom basetypes.SetValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToSetValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.StringTypable:
		diags.Append(visitor.setOfString(ctx, v, vTo)...)
		return diags

	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.NestedObjectValue); ok {
			diags.Append(visitor.nestedObject(ctx, vFrom, vTo)...)
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Expand; incompatible types", map[string]interface{}{
		"from set[%s]": v.ElementType(ctx),
		"to":           vTo.Kind(),
	})

	return diags
}

// setOfString copies a Plugin Framework SetOfString(ish) value to a compatible AWS API field.
func (visitor expandVisitor) setOfString(ctx context.Context, vFrom basetypes.SetValue, vTo reflect.Value) diag.Diagnostics { //nolint:unparam
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

// nestedObject copies a Plugin Framework NestedObjectValue value to a compatible AWS API field.
func (visitor expandVisitor) nestedObject(ctx context.Context, vFrom fwtypes.NestedObjectValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Ptr:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			//
			// types.List(OfObject) -> *struct.
			//
			diags.Append(visitor.nestedObjectToStruct(ctx, vFrom, tElem, vTo)...)
			return diags
		}

	case reflect.Slice:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			//
			// types.List(OfObject) -> []struct.
			//
			diags.Append(visitor.nestedObjectToSlice(ctx, vFrom, tTo, tElem, vTo)...)
			return diags

		case reflect.Ptr:
			switch tElem := tElem.Elem(); tElem.Kind() {
			case reflect.Struct:
				//
				// types.List(OfObject) -> []*struct.
				//
				diags.Append(visitor.nestedObjectToSlice(ctx, vFrom, tTo, tElem, vTo)...)
				return diags
			}
		}
	}

	diags.AddError("Incompatible types", fmt.Sprintf("nestedObject[%s] cannot be expanded to %s", vFrom.Type(ctx).(attr.TypeWithElementType).ElementType(), vTo.Kind()))
	return diags
}

// nestedObjectToStruct copies a Plugin Framework NestedObjectValue to a compatible AWS API (*)struct field.
func (visitor expandVisitor) nestedObjectToStruct(ctx context.Context, vFrom fwtypes.NestedObjectValue, tStruct reflect.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the nested Object as a pointer.
	from, d := vFrom.ToObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Create a new target structure and walk its fields.
	to := reflect.New(tStruct)
	diags.Append(walkStructFields(ctx, from, to.Interface(), visitor)...)
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

// nestedObjectToSlice copies a Plugin Framework NestedObjectValue to a compatible AWS API [](*)struct field.
func (visitor expandVisitor) nestedObjectToSlice(ctx context.Context, vFrom fwtypes.NestedObjectValue, tSlice, tElem reflect.Type, vTo reflect.Value) diag.Diagnostics {
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
		diags.Append(walkStructFields(ctx, f.Index(i).Interface(), target.Interface(), visitor)...)
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

type flattenVisitor struct{}

func (visitor flattenVisitor) visit(ctx context.Context, fieldName string, vFrom, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo, ok := vTo.Interface().(attr.Value)
	if !ok {
		diags.AddError("AutoFlEx", fmt.Sprintf("does not implement attr.Value: %s", vTo.Kind()))
		return diags
	}

	tTo := valTo.Type(ctx)
	switch vFrom.Kind() {
	case reflect.Bool:
		diags.Append(visitor.bool(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Float32, reflect.Float64:
		diags.Append(visitor.float(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Int32, reflect.Int64:
		diags.Append(visitor.int(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.String:
		diags.Append(visitor.string(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Ptr:
		diags.Append(visitor.ptr(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Slice:
		diags.Append(visitor.slice(ctx, vFrom, tTo, vTo)...)
		return diags

	case reflect.Map:
		diags.Append(visitor.map_(ctx, vFrom, tTo, vTo)...)
		return diags
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// bool copies an AWS API bool value to a compatible Plugin Framework field.
func (visitor flattenVisitor) bool(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.BoolTypable:
		v, d := tTo.ValueFromBool(ctx, types.BoolValue(vFrom.Bool()))
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

// float copies an AWS API float value to a compatible Plugin Framework field.
func (visitor flattenVisitor) float(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Float64Typable:
		v, d := tTo.ValueFromFloat64(ctx, types.Float64Value(vFrom.Float()))
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

// int copies an AWS API int value to a compatible Plugin Framework field.
func (visitor flattenVisitor) int(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Int64Typable:
		v, d := tTo.ValueFromInt64(ctx, types.Int64Value(vFrom.Int()))
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

// string copies an AWS API string value to a compatible Plugin Framework field.
func (visitor flattenVisitor) string(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.StringTypable:
		v, d := tTo.ValueFromString(ctx, types.StringValue(vFrom.String()))
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

// ptr copies an AWS API pointer value to a compatible Plugin Framework field.
func (visitor flattenVisitor) ptr(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vElem := vFrom.Elem(); vFrom.Type().Elem().Kind() {
	case reflect.Bool:
		if vFrom.IsNil() {
			vTo.Set(reflect.ValueOf(types.BoolNull()))
			return diags
		}

		diags.Append(visitor.bool(ctx, vElem, tTo, vTo)...)
		return diags

	case reflect.Float32, reflect.Float64:
		if vFrom.IsNil() {
			vTo.Set(reflect.ValueOf(types.Float64Null()))
			return diags
		}

		diags.Append(visitor.float(ctx, vElem, tTo, vTo)...)
		return diags

	case reflect.Int32, reflect.Int64:
		if vFrom.IsNil() {
			vTo.Set(reflect.ValueOf(types.Int64Null()))
			return diags
		}

		diags.Append(visitor.int(ctx, vElem, tTo, vTo)...)
		return diags

	case reflect.String:
		if vFrom.IsNil() {
			vTo.Set(reflect.ValueOf(types.StringNull()))
			return diags
		}

		diags.Append(visitor.string(ctx, vElem, tTo, vTo)...)
		return diags

	case reflect.Struct:
		if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
			//
			// *struct -> types.List(OfObject).
			//
			diags.Append(visitor.ptrToStructNestedObject(ctx, vFrom, tTo, vTo)...)
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// slice copies an AWS API slice value to a compatible Plugin Framework field.
func (visitor flattenVisitor) slice(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tSliceElem := vFrom.Type().Elem(); tSliceElem.Kind() {
	case reflect.String:
		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []string -> types.List(OfString).
			//
			if vFrom.IsNil() {
				vTo.Set(reflect.ValueOf(types.ListNull(types.StringType)))
				return diags
			}

			from := vFrom.Interface().([]string)
			elements := make([]attr.Value, len(from))
			for i, v := range from {
				elements[i] = types.StringValue(v)
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
				vTo.Set(reflect.ValueOf(types.SetNull(types.StringType)))
				return diags
			}

			from := vFrom.Interface().([]string)
			elements := make([]attr.Value, len(from))
			for i, v := range from {
				elements[i] = types.StringValue(v)
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
					vTo.Set(reflect.ValueOf(types.ListNull(types.StringType)))
					return diags
				}

				from := vFrom.Interface().([]*string)
				elements := make([]attr.Value, len(from))
				for i, v := range from {
					elements[i] = types.StringValue(aws.ToString(v))
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
					vTo.Set(reflect.ValueOf(types.SetNull(types.StringType)))
					return diags
				}

				from := vFrom.Interface().([]*string)
				elements := make([]attr.Value, len(from))
				for i, v := range from {
					elements[i] = types.StringValue(aws.ToString(v))
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
			if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
				//
				// []*struct -> types.List(OfObject).
				//
				diags.Append(visitor.sliceOfStructNestedObject(ctx, vFrom, tTo, vTo)...)
				return diags
			}
		}

	case reflect.Struct:
		if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
			//
			// []struct -> types.List(OfObject).
			//
			diags.Append(visitor.sliceOfStructNestedObject(ctx, vFrom, tTo, vTo)...)
			return diags
		}
	}

	tflog.Info(ctx, "AutoFlex Flatten; incompatible types", map[string]interface{}{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// map_ copies an AWS API map value to a compatible Plugin Framework field.
func (visitor flattenVisitor) map_(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tMapKey := vFrom.Type().Key(); tMapKey.Kind() {
	case reflect.String:
		switch tMapElem := vFrom.Type().Elem(); tMapElem.Kind() {
		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.MapTypable:
				//
				// map[string]string -> types.Map(OfString).
				//
				if vFrom.IsNil() {
					vTo.Set(reflect.ValueOf(types.MapNull(types.StringType)))
					return diags
				}

				v, d := tTo.ValueFromMap(ctx, FlattenFrameworkStringValueMap(ctx, vFrom.Interface().(map[string]string)))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(v))
				return diags
			}

		case reflect.Ptr:
			switch tMapElem.Elem().Kind() {
			case reflect.String:
				switch tTo := tTo.(type) {
				case basetypes.MapTypable:
					//
					// map[string]*string -> types.Map(OfString).
					//
					if vFrom.IsNil() {
						vTo.Set(reflect.ValueOf(types.MapNull(types.StringType)))
						return diags
					}

					v, d := tTo.ValueFromMap(ctx, FlattenFrameworkStringMap(ctx, vFrom.Interface().(map[string]*string)))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(v))
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

// ptrToStructNestedObject copies an AWS API *struct value to a compatible Plugin Framework NestedObjectValue field.
func (visitor flattenVisitor) ptrToStructNestedObject(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectType, vTo reflect.Value) diag.Diagnostics {
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

	// Create a new target structure and walk its fields.
	to, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(walkStructFields(ctx, vFrom.Elem().Interface(), to, visitor)...)
	if diags.HasError() {
		return diags
	}

	// Set the target structure as a nested Object.
	val, d := tTo.ValueFromObjectPtr(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

// sliceOfStructNestedObject copies an AWS API []struct value to a compatible Plugin Framework NestedObjectValue field.
func (visitor flattenVisitor) sliceOfStructNestedObject(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectType, vTo reflect.Value) diag.Diagnostics {
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

		diags.Append(walkStructFields(ctx, vFrom.Index(i).Interface(), target, visitor)...)
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
