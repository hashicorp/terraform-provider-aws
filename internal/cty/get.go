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
	tSrc := source.Type()
	if !tSrc.IsObjectType() {
		return fmt.Errorf("source must be an object, got %s", tSrc.FriendlyName())
	}

	vTarget := reflect.ValueOf(target)
	if kind := vTarget.Kind(); kind != reflect.Ptr {
		return fmt.Errorf("target must be a pointer, got %T (Kind: %s)", target, kind)
	}
	vTarget = vTarget.Elem()
	if kind := vTarget.Kind(); kind != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct, got %T (Kind: %s)", target, kind)
	}

	for key, vSource := range ValueElements(source) {
		key := key.AsString()
		field, ok := tfreflect.FieldByTag(target, "tfsdk", key)
		if !ok {
			continue
		}
		vField, err := vTarget.FieldByIndexErr(field.Index)
		if err != nil {
			return err
		}

		attrValue, err := attrValueOf(ctx, vSource, vField.Interface())
		if err != nil {
			return err
		}

		vField.Set(reflect.ValueOf(attrValue))
	}

	return nil
}

func attrValueOf(ctx context.Context, source cty.Value, target any) (attr.Value, error) {
	tfType, err := ToTfValue(source)
	if err != nil {
		return nil, err
	}

	if attrValue, ok := target.(attr.Value); ok {
		return attrValue.Type(ctx).ValueFromTerraform(ctx, *tfType)
	}

	return nil, fmt.Errorf("attrValueOf unsupported type: %T", target)
}
