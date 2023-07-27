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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Expand "expands" a resource's "business logic" data structure,
// implemented using Terraform Plugin Framework data types, into
// an AWS SDK for Go v2 API data structure.
// The resource's data structure is walked and exported fields that
// have a corresponding field in the API data structure (and a suitable
// target data type) are copied.
func Expand(ctx context.Context, tfObject, apiObject any) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(walkStructFields(ctx, tfObject, apiObject, expandVisitor{})...)
	if diags.HasError() {
		diags.AddError("AutoFlEx", fmt.Sprintf("Expand[%T, %T]", tfObject, apiObject))
		return diags
	}

	return diags
}

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

	diags.Append(visitor.newIncompatibleTypesError(ctx, vFrom, vTo))
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

	diags.Append(visitor.newIncompatibleTypesError(ctx, vFrom, vTo))
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

	diags.Append(visitor.newIncompatibleTypesError(ctx, vFrom, vTo))
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

	diags.Append(visitor.newIncompatibleTypesError(ctx, vFrom, vTo))
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

	diags.Append(visitor.newIncompatibleTypesError(ctx, vFrom, vTo))
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
		diags.Append(visitor.listOfObject(ctx, vFrom, vTo)...)
		return diags
	}

	diags.Append(visitor.newIncompatibleListTypesError(ctx, v, vTo))
	return diags
}

// listOfString copies a Plugin Framework ListOfString(ish) value to a compatible AWS API field.
func (visitor expandVisitor) listOfString(ctx context.Context, vFrom basetypes.ListValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Slice:
		switch tSliceElem := vTo.Type().Elem(); tSliceElem.Kind() {
		case reflect.String:
			//
			// types.List(OfString) -> []string.
			//
			vTo.Set(reflect.ValueOf(ExpandFrameworkStringValueList(ctx, vFrom)))
			return diags

		case reflect.Ptr:
			switch tSliceElem.Elem().Kind() {
			case reflect.String:
				//
				// types.List(OfString) -> []*string.
				//
				vTo.Set(reflect.ValueOf(ExpandFrameworkStringList(ctx, vFrom)))
				return diags
			}
		}
	}

	diags.Append(visitor.newIncompatibleListTypesError(ctx, vFrom, vTo))
	return diags
}

// listOfObject copies a Plugin Framework ListOfObject(ish) value to a compatible AWS API field.
func (visitor expandVisitor) listOfObject(ctx context.Context, vFrom attr.Value, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if vNestedObject, ok := vFrom.(types.NestedObjectValue); ok {
		switch tTo := vTo.Type(); vTo.Kind() {
		case reflect.Ptr:
			switch tElem := tTo.Elem(); tElem.Kind() {
			case reflect.Struct:
				//
				// types.List(OfObject) -> *struct.
				//

				// Get the nested Object as a pointer.
				from, d := vNestedObject.ToObjectPtr(ctx)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				// Create a new target structure and walk its fields.
				to := reflect.New(tElem)
				diags.Append(walkStructFields(ctx, from, to.Interface(), visitor)...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(to)
				return diags
			}

		case reflect.Slice:
			switch tElem := tTo.Elem(); tElem.Kind() {
			case reflect.Struct:
				//
				// types.List(OfObject) -> []struct.
				//

				// Get the nested Objects as a slice.
				from, d := vNestedObject.ToObjectSlice(ctx)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				// Create a new target slice and expand each element.
				f := reflect.ValueOf(from)
				n := f.Len()
				t := reflect.MakeSlice(tTo, n, n)
				for i := 0; i < n; i++ {
					// Create a new target structure and walk its fields.
					target := reflect.New(tElem)
					diags.Append(walkStructFields(ctx, f.Index(i).Interface(), target.Interface(), visitor)...)
					if diags.HasError() {
						return diags
					}

					// Set value in the target slice.
					t.Index(i).Set(target.Elem())
				}

				vTo.Set(t)
				return diags
			}
		}
	}

	diags.AddError("Incompatible types", fmt.Sprintf("listOfObject[%s] cannot be expanded to %s", vFrom.Type(ctx).(attr.TypeWithElementType).ElementType(), vTo.Kind()))
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

	diags.Append(visitor.newIncompatibleMapTypesError(ctx, v, vTo))
	return diags
}

// mapOfString copies a Plugin Framework MapOfString(ish) value to a compatible AWS API field.
func (visitor expandVisitor) mapOfString(ctx context.Context, vFrom basetypes.MapValue, vTo reflect.Value) diag.Diagnostics {
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
				vTo.Set(reflect.ValueOf(ExpandFrameworkStringValueMap(ctx, vFrom)))
				return diags

			case reflect.Ptr:
				switch tMapElem.Elem().Kind() {
				case reflect.String:
					//
					// types.Map(OfString) -> map[string]*string.
					//
					vTo.Set(reflect.ValueOf(ExpandFrameworkStringMap(ctx, vFrom)))
					return diags
				}
			}
		}
	}

	diags.Append(visitor.newIncompatibleMapTypesError(ctx, vFrom, vTo))
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
	}

	diags.Append(visitor.newIncompatibleSetTypesError(ctx, v, vTo))
	return diags
}

// setOfString copies a Plugin Framework SetOfString(ish) value to a compatible AWS API field.
func (visitor expandVisitor) setOfString(ctx context.Context, vFrom basetypes.SetValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Slice:
		switch tSliceElem := vTo.Type().Elem(); tSliceElem.Kind() {
		case reflect.String:
			//
			// types.Set(OfString) -> []string.
			//
			vTo.Set(reflect.ValueOf(ExpandFrameworkStringValueSet(ctx, vFrom)))
			return diags

		case reflect.Ptr:
			switch tSliceElem.Elem().Kind() {
			case reflect.String:
				//
				// types.Set(OfString) -> []*string.
				//
				vTo.Set(reflect.ValueOf(ExpandFrameworkStringSet(ctx, vFrom)))
				return diags
			}
		}
	}

	diags.Append(visitor.newIncompatibleSetTypesError(ctx, vFrom, vTo))
	return diags
}

func (visitor expandVisitor) newIncompatibleTypesError(ctx context.Context, vFrom attr.Value, vTo reflect.Value) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Incompatible types", fmt.Sprintf("%s cannot be expanded to %s", vFrom.Type(ctx), vTo.Kind()))
}

func (visitor expandVisitor) newIncompatibleListTypesError(ctx context.Context, vFrom basetypes.ListValue, vTo reflect.Value) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Incompatible types", fmt.Sprintf("list[%s] cannot be expanded to %s", vFrom.ElementType(ctx), vTo.Kind()))
}

func (visitor expandVisitor) newIncompatibleMapTypesError(ctx context.Context, vFrom basetypes.MapValue, vTo reflect.Value) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Incompatible types", fmt.Sprintf("map[string, %s] cannot be expanded to %s", vFrom.ElementType(ctx), vTo.Kind()))
}

func (visitor expandVisitor) newIncompatibleSetTypesError(ctx context.Context, vFrom basetypes.SetValue, vTo reflect.Value) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Incompatible types", fmt.Sprintf("set[%s] cannot be expanded to %s", vFrom.ElementType(ctx), vTo.Kind()))
}
