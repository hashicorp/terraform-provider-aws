package flex

import (
	"context"
	"fmt"
	"reflect"
)

func expand(ctx context.Context, tfObject any, apiObject any) error {
	valFrom, valTo := reflect.ValueOf(tfObject), reflect.ValueOf(apiObject)

	if kind := valFrom.Kind(); kind == reflect.Ptr {
		valFrom = valFrom.Elem()
	}
	if kind := valTo.Kind(); kind != reflect.Ptr {
		return fmt.Errorf("target (%T): %s, want pointer", apiObject, kind)
	}
	valTo = valTo.Elem()

	typFrom, typTo := valFrom.Type(), valTo.Type()

	if typFrom.Kind() != reflect.Struct {
		return fmt.Errorf("source: %s, want struct", typFrom)
	}
	if typTo.Kind() != reflect.Struct {
		return fmt.Errorf("target: %s, want struct", typFrom)
	}

	return nil
}
