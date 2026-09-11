// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	tfreflect "github.com/hashicorp/terraform-provider-aws/internal/reflect"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/sync"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// attributeTypesCache memoizes AttributeTypes results, keyed by struct type.
var attributeTypesCache tfsync.Map[reflect.Type, map[string]attr.Type]

// AttributeTypes returns a map of attribute types for the specified type T.
// T must be a struct and reflection is used to find exported fields of T with the `tfsdk` tag.
// The returned map is shared and must not be mutated by callers.
func AttributeTypes[T any](ctx context.Context) (map[string]attr.Type, diag.Diagnostics) {
	var diags diag.Diagnostics

	typ := reflect.TypeFor[T]()
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	if cached, ok := attributeTypesCache.Load(typ); ok {
		return cached, diags
	}

	var t T
	val := reflect.ValueOf(t)

	if val.Kind() == reflect.Pointer && val.Type().Elem().Kind() == reflect.Struct {
		val = reflect.New(val.Type().Elem()).Elem()
	}

	if typ.Kind() != reflect.Struct {
		diags.Append(diag.NewErrorDiagnostic("Invalid type", fmt.Sprintf("%T has unsupported type: %s", t, typ)))
		return nil, diags
	}

	attributeTypes := make(map[string]attr.Type)
	for field := range tfreflect.ExportedStructFields(typ) {
		tag := field.Tag.Get(`tfsdk`)
		if tag == "-" {
			continue // Skip explicitly excluded fields.
		}
		if tag == "" {
			diags.Append(diag.NewErrorDiagnostic("Invalid type", fmt.Sprintf(`%T needs a struct tag for "tfsdk" on %s`, t, field.Name)))
			return nil, diags
		}

		if v, ok := val.FieldByIndex(field.Index).Interface().(attr.Value); ok {
			attributeTypes[tag] = v.Type(ctx)
		}
	}

	// Store the successful result. LoadOrStore returns the canonical map if another
	// goroutine populated the same key concurrently.
	cached, _ := attributeTypesCache.LoadOrStore(typ, attributeTypes)
	return cached, diags
}

// AttributeTypesMust is like AttributeTypes but panics if T is not a valid struct type.
// The returned map is shared and must not be mutated by callers.
func AttributeTypesMust[T any](ctx context.Context) map[string]attr.Type {
	return fwdiag.Must(AttributeTypes[T](ctx))
}

func newAttrTypeOf[T attr.Value](ctx context.Context) attr.Type {
	return inttypes.Zero[T]().Type(ctx)
}
