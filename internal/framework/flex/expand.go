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
		v, diags := vFrom.ToFloat64Value(ctx)
		if err := fwdiag.DiagnosticsError(diags); err != nil {
			return err
		}
		switch vFrom := v.ValueFloat64(); kTo {
		case reflect.Float32, reflect.Float64:
			//
			// types.Float32/types.Float64 -> float32/float64.
			//
			valTo.SetFloat(vFrom)
			return nil
		case reflect.Ptr:
			switch valTo.Type().Elem().Kind() {
			case reflect.Float32:
				//
				// types.Float32/types.Float64 -> *float32.
				//
				valTo.Set(reflect.ValueOf(aws.Float32(float32(vFrom))))
				return nil
			case reflect.Float64:
				//
				// types.Float32/types.Float64 -> *float64.
				//
				valTo.Set(reflect.ValueOf(aws.Float64(vFrom)))
				return nil
			}
		}

	case basetypes.Int64Valuable:
		v, diags := vFrom.ToInt64Value(ctx)
		if err := fwdiag.DiagnosticsError(diags); err != nil {
			return err
		}
		switch vFrom := v.ValueInt64(); kTo {
		case reflect.Int32, reflect.Int64:
			//
			// types.Int32/types.Int64 -> int32/int64.
			//
			valTo.SetInt(vFrom)
			return nil
		case reflect.Ptr:
			switch valTo.Type().Elem().Kind() {
			case reflect.Int32:
				//
				// types.Int32/types.Int64 -> *int32.
				//
				valTo.Set(reflect.ValueOf(aws.Int32(int32(vFrom))))
				return nil
			case reflect.Int64:
				//
				// types.Int32/types.Int64 -> *int64.
				//
				valTo.Set(reflect.ValueOf(aws.Int64(vFrom)))
				return nil
			}
		}

	case basetypes.StringValuable:
		v, diags := vFrom.ToStringValue(ctx)
		if err := fwdiag.DiagnosticsError(diags); err != nil {
			return err
		}
		switch vFrom := v.ValueString(); kTo {
		case reflect.String:
			//
			// types.String -> string.
			//
			valTo.SetString(vFrom)
			return nil
		case reflect.Ptr:
			switch valTo.Type().Elem().Kind() {
			case reflect.String:
				//
				// types.String -> *string.
				//
				valTo.Set(reflect.ValueOf(aws.String(vFrom)))
				return nil
			}
		}

		// Aggregate types.
	case basetypes.ListValuable:
		v, diags := vFrom.ToListValue(ctx)
		if err := fwdiag.DiagnosticsError(diags); err != nil {
			return err
		}
		switch tElem := v.ElementType(ctx).(type) {
		case basetypes.StringTypable:
			switch kTo {
			case reflect.Slice:
				switch tSliceElem := valTo.Type().Elem(); tSliceElem.Kind() {
				case reflect.String:
					//
					// types.List(OfString) -> []string.
					//
					valTo.Set(reflect.ValueOf(ExpandFrameworkStringValueList(ctx, v)))
					return nil

				case reflect.Ptr:
					switch tSliceElem.Elem().Kind() {
					case reflect.String:
						//
						// types.List(OfString) -> []*string.
						//
						valTo.Set(reflect.ValueOf(ExpandFrameworkStringList(ctx, v)))
						return nil
					}
				}
			}

		case basetypes.ObjectTypable:
			// TODO...
			switch kTo {
			case reflect.Ptr:
				switch valTo.Type().Elem().Kind() {
				case reflect.Struct:
					if p, ok := tElem.ValueType(ctx).(types.ValueAsPtr); ok {
						if elements := v.Elements(); len(elements) == 1 {
							//
							// types.List(OfObject) -> *struct.
							//
							from, diags := p.ValueAsPtr(ctx)
							if err := fwdiag.DiagnosticsError(diags); err != nil {
								return err
							}
							to := reflect.New(valTo.Type())
							if err := walkStructFields(ctx, from, to, visitor); err != nil {
								return err
							}
							valTo.Set(reflect.ValueOf(to))
							return nil
						}
					}
				}
			}
			//
			// types.List(OfObject) -> ???.
			//
			return nil
		}

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

func (visitor expandVisitor) newIncompatibleTypesError(ctx context.Context, vFrom attr.Value, vTo reflect.Value) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Incompatible types", fmt.Sprintf("%s cannot be expanded to %s", vFrom.Type(ctx), vTo.Kind()))
}
