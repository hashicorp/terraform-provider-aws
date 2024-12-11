// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// AttributeTypes returns a map of attribute types for the specified type T.
// T must be a struct and reflection is used to find exported fields of T with the `tfsdk` tag.
func AttributeTypes[T any](ctx context.Context) (map[string]attr.Type, diag.Diagnostics) {
	var diags diag.Diagnostics
	var t T
	val := reflect.ValueOf(t)
	typ := val.Type()

	if typ.Kind() == reflect.Ptr && typ.Elem().Kind() == reflect.Struct {
		val = reflect.New(typ.Elem()).Elem()
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		diags.Append(diag.NewErrorDiagnostic("Invalid type", fmt.Sprintf("%T has unsupported type: %s", t, typ)))
		return nil, diags
	}

	attributeTypes := make(map[string]attr.Type)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			continue // Skip unexported fields.
		}
		tag := field.Tag.Get(`tfsdk`)
		if tag == "-" {
			continue // Skip explicitly excluded fields.
		}
		if tag == "" {
			diags.Append(diag.NewErrorDiagnostic("Invalid type", fmt.Sprintf(`%T needs a struct tag for "tfsdk" on %s`, t, field.Name)))
			return nil, diags
		}

		if v, ok := val.Field(i).Interface().(attr.Value); ok {
			attributeTypes[tag] = v.Type(ctx)
		}
	}

	return attributeTypes, nil
}

func AttributeTypesMust[T any](ctx context.Context) map[string]attr.Type {
	return fwdiag.Must(AttributeTypes[T](ctx))
}

func newAttrTypeOf[T attr.Value](ctx context.Context) attr.Type {
	var zero T
	return zero.Type(ctx)
}
