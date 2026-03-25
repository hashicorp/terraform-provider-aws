// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/go-cty/cty"
	tfreflect "github.com/hashicorp/terraform-provider-aws/internal/reflect"
)

// Get populates the struct passed as `target` with the entire cty object passed as `source`.
func Get(ctx context.Context, source cty.Value, target any) error {
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
		return fmt.Errorf("target must be a pointer to struct, got %T", target)
	}

	for key, _ := range ValueElements(source) {
		key := key.AsString()
		_, ok := tfreflect.FieldByTag(target, "tfsdk", key)
		if !ok {
			continue
		}
	}

	return nil
}
