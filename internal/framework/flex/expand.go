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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// TODO
// TODO Return Diagnostics, not error.
// TODO Add a post-func to tidy up.
// TODO

// Expand "expands" a resource's "business logic" data structure,
// implemented using Terraform Plugin Framework data types, into
// an AWS SDK for Go v2 API data structure.
// The resource's data structure is walked and exported fields that
// have a corresponding field in the API data structure (and a suitable
// target data type) are copied.
func Expand(ctx context.Context, tfObject, apiObject any) error {
	if err := walkStructFields(ctx, tfObject, apiObject, expandVisitor{}); err != nil {
		return fmt.Errorf("Expand[%T, %T]: %w", tfObject, apiObject, err)
	}

	return nil
}

// walkStructFields traverses `from` calling `visitor` for each exported field.
func walkStructFields(ctx context.Context, from any, to any, visitor fieldVisitor) error {
	valFrom, valTo := reflect.ValueOf(from), reflect.ValueOf(to)

	if kind := valFrom.Kind(); kind == reflect.Ptr {
		valFrom = valFrom.Elem()
	}
	if kind := valTo.Kind(); kind != reflect.Ptr {
		return fmt.Errorf("target (%T): %s, want pointer", to, kind)
	}
	valTo = valTo.Elem()

	typFrom, typTo := valFrom.Type(), valTo.Type()

	if typFrom.Kind() != reflect.Struct {
		return fmt.Errorf("source: %s, want struct", typFrom)
	}
	if typTo.Kind() != reflect.Struct {
		return fmt.Errorf("target: %s, want struct", typTo)
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
		if err := visitor.visit(ctx, fieldName, valFrom.Field(i), toFieldVal); err != nil {
			return fmt.Errorf("visit (%s): %w", fieldName, err)
		}
	}

	return nil
}

type fieldVisitor interface {
	visit(context.Context, string, reflect.Value, reflect.Value) error
}

type expandVisitor struct{}

// visit copies a single Plugin Framework structure field value to its AWS API equivalent.
func (visitor expandVisitor) visit(ctx context.Context, fieldName string, valFrom, valTo reflect.Value) error {
	vFrom, ok := valFrom.Interface().(attr.Value)
	if !ok {
		return fmt.Errorf("does not implement attr.Value: %s", valFrom.Kind())
	}

	// No need to set the target value if there's no source value.
	if vFrom.IsNull() || vFrom.IsUnknown() {
		return nil
	}

	tFrom, kTo := vFrom.Type(ctx), valTo.Kind()
	switch vFrom := vFrom.(type) {
	// Simple types.
	case basetypes.BoolValuable:
		diags := visitor.bool(ctx, vFrom, valTo)
		return fwdiag.DiagnosticsError(diags)

	case basetypes.Float64Valuable:
		diags := visitor.float64(ctx, vFrom, valTo)
		return fwdiag.DiagnosticsError(diags)

	case basetypes.Int64Valuable:
		diags := visitor.int64(ctx, vFrom, valTo)
		return fwdiag.DiagnosticsError(diags)

	case basetypes.StringValuable:
		diags := visitor.string(ctx, vFrom, valTo)
		return fwdiag.DiagnosticsError(diags)

		// Aggregate types.
	case basetypes.ListValuable:
		diags := visitor.list(ctx, vFrom, valTo)
		return fwdiag.DiagnosticsError(diags)

	case basetypes.MapValuable:
		v, diags := vFrom.ToMapValue(ctx)
		if err := fwdiag.DiagnosticsError(diags); err != nil {
			return err
		}
		switch v.ElementType(ctx).(type) {
		case basetypes.StringTypable:
			switch kTo {
			case reflect.Map:
				switch tMapKey := valTo.Type().Key(); tMapKey.Kind() {
				case reflect.String:
					switch tMapElem := valTo.Type().Elem(); tMapElem.Kind() {
					case reflect.String:
						//
						// types.Map(OfString) -> map[string]string.
						//
						valTo.Set(reflect.ValueOf(ExpandFrameworkStringValueMap(ctx, v)))
						return nil

					case reflect.Ptr:
						switch tMapElem.Elem().Kind() {
						case reflect.String:
							//
							// types.Map(OfString) -> map[string]*string.
							//
							valTo.Set(reflect.ValueOf(ExpandFrameworkStringMap(ctx, v)))
							return nil
						}
					}
				}
			}
		}

	case basetypes.SetValuable:
		v, diags := vFrom.ToSetValue(ctx)
		if err := fwdiag.DiagnosticsError(diags); err != nil {
			return err
		}
		switch v.ElementType(ctx).(type) {
		case basetypes.StringTypable:
			switch kTo {
			case reflect.Slice:
				switch tSliceElem := valTo.Type().Elem(); tSliceElem.Kind() {
				case reflect.String:
					//
					// types.Set(OfString) -> []string.
					//
					valTo.Set(reflect.ValueOf(ExpandFrameworkStringValueSet(ctx, v)))
					return nil

				case reflect.Ptr:
					switch tSliceElem.Elem().Kind() {
					case reflect.String:
						//
						// types.Set(OfString) -> []*string.
						//
						valTo.Set(reflect.ValueOf(ExpandFrameworkStringSet(ctx, v)))
						return nil
					}
				}
			}

		case basetypes.ObjectTypable:
			//
			// types.Set(OfObject) -> ???.
			//
			return nil
		}
	}

	return fmt.Errorf("incompatible (%s): %s", tFrom, kTo)
}

// bool copies a Plugin Framework Bool(ish) value to a compatible AWS API field.
func (visitor expandVisitor) bool(ctx context.Context, vFrom basetypes.BoolValuable, vTo reflect.Value) diag.Diagnostics {
	v, diags := vFrom.ToBoolValue(ctx)
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
	v, diags := vFrom.ToFloat64Value(ctx)
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
	v, diags := vFrom.ToInt64Value(ctx)
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
	v, diags := vFrom.ToStringValue(ctx)
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
	v, diags := vFrom.ToListValue(ctx)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.StringTypable:
		return visitor.listOfString(ctx, v, vTo)

	case basetypes.ObjectTypable:
		return visitor.listOfObject(ctx, v, vTo)
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

// listOfSObject copies a Plugin Framework ListOfObject(ish) value to a compatible AWS API field.
func (visitor expandVisitor) listOfObject(ctx context.Context, vFrom basetypes.ListValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Ptr:
		switch vTo.Type().Elem().Kind() {
		case reflect.Struct:
			if p, ok := vFrom.ElementType(ctx).ValueType(ctx).(types.ValueAsPtr); ok {
				if elements := vFrom.Elements(); len(elements) == 1 {
					//
					// types.List(OfObject) -> *struct.
					//
					from, d := p.ValueAsPtr(ctx)
					if d.HasError() {
						diags.Append(d...)
						return diags
					}

					to := reflect.New(vTo.Type())
					if err := walkStructFields(ctx, from, to, visitor); err != nil {
						diags.Append(diag.NewErrorDiagnostic("", err.Error()))
						return diags
					}
				}
			}
		}
	}

	diags.Append(visitor.newIncompatibleListTypesError(ctx, vFrom, vTo))

	return diags
}

func (visitor expandVisitor) newIncompatibleTypesError(ctx context.Context, vFrom attr.Value, vTo reflect.Value) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Incompatible types", fmt.Sprintf("%s cannot be expanded to %s", vFrom.Type(ctx), vTo.Kind()))
}

func (visitor expandVisitor) newIncompatibleListTypesError(ctx context.Context, vFrom basetypes.ListValue, vTo reflect.Value) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Incompatible types", fmt.Sprintf("list[%s] cannot be expanded to %s", vFrom.ElementType(ctx), vTo.Kind()))
}
