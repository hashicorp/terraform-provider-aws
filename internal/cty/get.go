// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	tfreflect "github.com/hashicorp/terraform-provider-aws/internal/reflect"
)

// GetFramework populates the struct passed as `target` with the entire cty object passed as `source`.
// The target's fields must be Plugin Framework types.
func GetFramework(ctx context.Context, source cty.Value, target any) error {
	if !source.IsKnown() {
		return fmt.Errorf("source must not be unknown")
	}
	if source.IsNull() {
		return fmt.Errorf("source must not be nul")
	}
	if typ := source.Type(); !typ.IsObjectType() {
		return fmt.Errorf("source must be an object, got %s", typ.FriendlyName())
	}

	vTarget := reflect.ValueOf(target)
	if kind := vTarget.Kind(); kind != reflect.Ptr {
		return fmt.Errorf("target must be a pointer, got %T (Kind: %s)", target, kind)
	}
	vTarget = vTarget.Elem()
	if kind := vTarget.Kind(); kind != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct, got %T (Kind: %s)", target, kind)
	}

	for key, sourceFieldCtyValue := range ValueElements(source) {
		// Map source field to target field via `tfsdk:"<key>"` struct tag.
		key := key.AsString()
		targetField, ok := tfreflect.FieldByTag(target, "tfsdk", key)
		if !ok {
			continue
		}
		vTargetField := vTarget.FieldByIndex(targetField.Index)

		sourceFieldTfValue, err := ToTfValue(sourceFieldCtyValue)
		if err != nil {
			return fmt.Errorf("source field %s: %w", key, err)
		}

		targetFieldRaw := vTargetField.Interface()
		if attrValue, ok := targetFieldRaw.(attr.Value); ok {
			newAttrValue, err := attrValue.Type(ctx).ValueFromTerraform(ctx, *sourceFieldTfValue)
			if err != nil {
				return fmt.Errorf("source field %s: %w", key, err)
			}

			// Set the target field's value to that of the source field.
			vTargetField.Set(reflect.ValueOf(newAttrValue))
		} else {
			return fmt.Errorf("target field %s: unsupported type: %T", key, targetFieldRaw)
		}
	}

	return nil
}
