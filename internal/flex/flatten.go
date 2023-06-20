package flex

import (
	"context"
	"fmt"
	"reflect"
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
	return nil
}
