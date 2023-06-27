package flex

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// TODO
// TODO Return Diagnostics, not error.
// TODO

// Flatten "flattens" an AWS SDK for Go v2 API data structure into
// a resource's "business logic" data structure, implemented using
// Terraform Plugin Framework data types.
// The API data structure's fields are walked and exported fields that
// have a corresponding field in the resource's data structure (and a
// suitable target data type) are copied.
func Flatten(ctx context.Context, apiObject, tfObject any) error {
	if err := walkStructFields(ctx, apiObject, tfObject, flattenVisitor{}); err != nil {
		return fmt.Errorf("Flatten[%T, %T]: %w", apiObject, tfObject, err)
	}

	return nil
}

type flattenVisitor struct{}

func (v flattenVisitor) visit(ctx context.Context, fieldName string, valFrom, valTo reflect.Value) error {
	vTo, ok := valTo.Interface().(attr.Value)
	if !ok {
		return fmt.Errorf("does not implement attr.Value: %s", valTo.Kind())
	}

	kFrom, tTo := valFrom.Kind(), vTo.Type(ctx)
	switch kFrom {
	// Simple types.
	case reflect.Bool:
		switch tTo := tTo.(type) {
		case basetypes.BoolTypable:
			v, diags := tTo.ValueFromBool(ctx, types.BoolValue(valFrom.Bool()))
			if err := fwdiag.DiagnosticsError(diags); err != nil {
				return err
			}
			valTo.Set(reflect.ValueOf(v))
			return nil
		}

	case reflect.Float32, reflect.Float64:
		switch tTo := tTo.(type) {
		case basetypes.Float64Typable:
			v, diags := tTo.ValueFromFloat64(ctx, types.Float64Value(valFrom.Float()))
			if err := fwdiag.DiagnosticsError(diags); err != nil {
				return err
			}
			valTo.Set(reflect.ValueOf(v))
			return nil
		}

	case reflect.Int32, reflect.Int64:
		switch tTo := tTo.(type) {
		case basetypes.Int64Typable:
			v, diags := tTo.ValueFromInt64(ctx, types.Int64Value(valFrom.Int()))
			if err := fwdiag.DiagnosticsError(diags); err != nil {
				return err
			}
			valTo.Set(reflect.ValueOf(v))
			return nil
		}

	case reflect.String:
		switch tTo := tTo.(type) {
		case basetypes.StringTypable:
			v, diags := tTo.ValueFromString(ctx, types.StringValue(valFrom.String()))
			if err := fwdiag.DiagnosticsError(diags); err != nil {
				return err
			}
			valTo.Set(reflect.ValueOf(v))
			return nil
		}

	// Pointer to simple types.
	case reflect.Ptr:
		valElem := valFrom.Elem()
		switch valFrom.Type().Elem().Kind() {
		case reflect.Bool:
			switch tTo := tTo.(type) {
			case basetypes.BoolTypable:
				if valElem.IsValid() {
					v, diags := tTo.ValueFromBool(ctx, types.BoolValue(valElem.Bool()))
					if err := fwdiag.DiagnosticsError(diags); err != nil {
						return err
					}
					valTo.Set(reflect.ValueOf(v))
				} else {
					valTo.Set(reflect.ValueOf(types.BoolNull()))
				}
				return nil
			}

		case reflect.Float32, reflect.Float64:
			switch tTo := tTo.(type) {
			case basetypes.Float64Typable:
				if valElem.IsValid() {
					v, diags := tTo.ValueFromFloat64(ctx, types.Float64Value(valElem.Float()))
					if err := fwdiag.DiagnosticsError(diags); err != nil {
						return err
					}
					valTo.Set(reflect.ValueOf(v))
				} else {
					valTo.Set(reflect.ValueOf(types.Float64Null()))
				}
				return nil
			}

		case reflect.Int32, reflect.Int64:
			switch tTo := tTo.(type) {
			case basetypes.Int64Typable:
				if valElem.IsValid() {
					v, diags := tTo.ValueFromInt64(ctx, types.Int64Value(valElem.Int()))
					if err := fwdiag.DiagnosticsError(diags); err != nil {
						return err
					}
					valTo.Set(reflect.ValueOf(v))
				} else {
					valTo.Set(reflect.ValueOf(types.Int64Null()))
				}
				return nil
			}

		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.StringTypable:
				if valElem.IsValid() {
					v, diags := tTo.ValueFromString(ctx, types.StringValue(valElem.String()))
					if err := fwdiag.DiagnosticsError(diags); err != nil {
						return err
					}
					valTo.Set(reflect.ValueOf(v))
				} else {
					valTo.Set(reflect.ValueOf(types.StringNull()))
				}
				return nil
			}
		}

	// Slice of simple types or pointer to simple types.
	case reflect.Slice:
		vFrom := valFrom.Interface()
		switch tSliceElem := valFrom.Type().Elem(); tSliceElem.Kind() {
		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				if vFrom != nil {
					v, diags := tTo.ValueFromList(ctx, FlattenFrameworkStringValueList(ctx, vFrom.([]string)))
					if err := fwdiag.DiagnosticsError(diags); err != nil {
						return err
					}
					valTo.Set(reflect.ValueOf(v))
				} else {
					valTo.Set(reflect.ValueOf(types.ListNull(types.StringType)))
				}
				return nil

			case basetypes.SetTypable:
				if vFrom != nil {
					v, diags := tTo.ValueFromSet(ctx, FlattenFrameworkStringValueSet(ctx, vFrom.([]string)))
					if err := fwdiag.DiagnosticsError(diags); err != nil {
						return err
					}
					valTo.Set(reflect.ValueOf(v))
				} else {
					valTo.Set(reflect.ValueOf(types.SetNull(types.StringType)))
				}
				return nil
			}

		case reflect.Ptr:
			switch tSliceElem.Elem().Kind() {
			case reflect.String:
				switch tTo := tTo.(type) {
				case basetypes.ListTypable:
					if vFrom != nil {
						v, diags := tTo.ValueFromList(ctx, FlattenFrameworkStringList(ctx, vFrom.([]*string)))
						if err := fwdiag.DiagnosticsError(diags); err != nil {
							return err
						}
						valTo.Set(reflect.ValueOf(v))
					} else {
						valTo.Set(reflect.ValueOf(types.ListNull(types.StringType)))
					}
					return nil

				case basetypes.SetTypable:
					if vFrom != nil {
						v, diags := tTo.ValueFromSet(ctx, FlattenFrameworkStringSet(ctx, vFrom.([]*string)))
						if err := fwdiag.DiagnosticsError(diags); err != nil {
							return err
						}
						valTo.Set(reflect.ValueOf(v))
					} else {
						valTo.Set(reflect.ValueOf(types.SetNull(types.StringType)))
					}
					return nil
				}
			}
		}
	}

	return fmt.Errorf("incompatible (%s): %s", kFrom, tTo)
}
