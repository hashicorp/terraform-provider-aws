package flex

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
	case reflect.Int32, reflect.Int64:
		vFrom := valFrom.Int()
		switch {
		case tTo.Equal(types.Int64Type):
			valTo.Set(reflect.ValueOf(types.Int64Value(vFrom)))
			return nil
		}

	case reflect.String:
		vFrom := valFrom.String()
		switch {
		case tTo.Equal(types.StringType):
			valTo.Set(reflect.ValueOf(types.StringValue(vFrom)))
			return nil
		}

	case reflect.Ptr:
		vFrom := valFrom.Elem()
		switch valFrom.Type().Elem().Kind() {
		case reflect.Int32, reflect.Int64:
			if vFrom.IsValid() {
				valTo.Set(reflect.ValueOf(types.Int64Value(vFrom.Int())))
			} else {
				valTo.Set(reflect.ValueOf(types.Int64Null()))
			}
			return nil

		case reflect.String:
			if vFrom.IsValid() {
				valTo.Set(reflect.ValueOf(types.StringValue(vFrom.String())))
			} else {
				valTo.Set(reflect.ValueOf(types.StringNull()))
			}
			return nil
		}
	}

	return fmt.Errorf("incompatible (%s): %s", kFrom, tTo)
}
