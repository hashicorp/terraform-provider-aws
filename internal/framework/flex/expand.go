// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

func (v expandVisitor) visit(ctx context.Context, fieldName string, valFrom, valTo reflect.Value) error {
	vFrom, ok := valFrom.Interface().(attr.Value)
	if !ok {
		return fmt.Errorf("does not implement attr.Value: %s", valFrom.Kind())
	}

	// No need to set the target value if there's no source value.
	if vFrom.IsNull() || vFrom.IsUnknown() {
		return nil
	}

	tFrom, kTo := vFrom.Type(ctx), valTo.Kind()
	switch {
	// Simple types.
	case tFrom.Equal(types.BoolType):
		vFrom := vFrom.(types.Bool).ValueBool()
		switch kTo {
		case reflect.Bool:
			valTo.SetBool(vFrom)
			return nil
		case reflect.Ptr:
			switch valTo.Type().Elem().Kind() {
			case reflect.Bool:
				valTo.Set(reflect.ValueOf(aws.Bool(vFrom)))
				return nil
			}
		}

	case tFrom.Equal(types.Float64Type):
		vFrom := vFrom.(types.Float64).ValueFloat64()
		switch kTo {
		case reflect.Float32, reflect.Float64:
			valTo.SetFloat(vFrom)
			return nil
		case reflect.Ptr:
			switch valTo.Type().Elem().Kind() {
			case reflect.Float32:
				valTo.Set(reflect.ValueOf(aws.Float32(float32(vFrom))))
				return nil
			case reflect.Float64:
				valTo.Set(reflect.ValueOf(aws.Float64(vFrom)))
				return nil
			}
		}

	case tFrom.Equal(types.Int64Type):
		vFrom := vFrom.(types.Int64).ValueInt64()
		switch kTo {
		case reflect.Int32, reflect.Int64:
			valTo.SetInt(vFrom)
			return nil
		case reflect.Ptr:
			switch valTo.Type().Elem().Kind() {
			case reflect.Int32:
				valTo.Set(reflect.ValueOf(aws.Int32(int32(vFrom))))
				return nil
			case reflect.Int64:
				valTo.Set(reflect.ValueOf(aws.Int64(vFrom)))
				return nil
			}
		}

	case tFrom.Equal(types.StringType):
		vFrom := vFrom.(types.String).ValueString()
		switch kTo {
		case reflect.String:
			valTo.SetString(vFrom)
			return nil
		case reflect.Ptr:
			switch valTo.Type().Elem().Kind() {
			case reflect.String:
				valTo.Set(reflect.ValueOf(aws.String(vFrom)))
				return nil
			}
		}

		// Aggregate types.
	case tFrom.Equal(types.ListType{ElemType: types.StringType}):
		vFrom := vFrom.(types.List)
		switch kTo {
		case reflect.Slice:
			switch tSliceElem := valTo.Type().Elem(); tSliceElem.Kind() {
			case reflect.String:
				valTo.Set(reflect.ValueOf(ExpandFrameworkStringValueList(ctx, vFrom)))
				return nil

			case reflect.Ptr:
				switch tSliceElem.Elem().Kind() {
				case reflect.String:
					valTo.Set(reflect.ValueOf(ExpandFrameworkStringList(ctx, vFrom)))
					return nil
				}
			}
		}

	case tFrom.Equal(types.SetType{ElemType: types.StringType}):
		vFrom := vFrom.(types.Set)
		switch kTo {
		case reflect.Slice:
			switch tSliceElem := valTo.Type().Elem(); tSliceElem.Kind() {
			case reflect.String:
				valTo.Set(reflect.ValueOf(ExpandFrameworkStringValueSet(ctx, vFrom)))
				return nil

			case reflect.Ptr:
				switch tSliceElem.Elem().Kind() {
				case reflect.String:
					valTo.Set(reflect.ValueOf(ExpandFrameworkStringSet(ctx, vFrom)))
					return nil
				}
			}
		}
	}

	return fmt.Errorf("incompatible (%s): %s", tFrom, kTo)
}
