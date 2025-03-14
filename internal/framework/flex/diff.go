// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	tfreflect "github.com/hashicorp/terraform-provider-aws/internal/reflect"
)

type Results struct {
	hasChanges            bool
	ignoredFieldNames     []string
	flexIgnoredFieldNames []AutoFlexOptionsFunc
}

// HasChanges returns whether there are changes between the plan and state values
func (r *Results) HasChanges() bool {
	return r.hasChanges
}

// IgnoredFieldNamesOpts returns the list of ignored field names as AutoFlexOptionsFunc
func (r *Results) IgnoredFieldNamesOpts() []AutoFlexOptionsFunc {
	for _, v := range r.ignoredFieldNames {
		r.flexIgnoredFieldNames = append(r.flexIgnoredFieldNames, WithIgnoredFieldNamesAppend(v))
	}
	return r.flexIgnoredFieldNames
}

// IgnoredFieldNames returns the list of ignored field names
func (r *Results) IgnoredFieldNames() []string {
	return r.ignoredFieldNames
}

// Diff compares the plan and state values and returns whether there are changes
func Diff(ctx context.Context, plan, state any, options ...ChangeOption) (*Results, diag.Diagnostics) {
	var diags diag.Diagnostics
	opts := NewChangeOptions(options...)

	planValue, stateValue := dereferencePointer(reflect.ValueOf(plan)), dereferencePointer(reflect.ValueOf(state))
	planType, stateType := planValue.Type(), stateValue.Type()
	var ignoredFields []string
	result := Results{}

	if planType != stateType {
		diags.AddError(
			"Type mismatch between plan and state",
			fmt.Sprintf("plan type: %s, state type: %s", planType.String(), stateType.String()),
		)
		return &result, diags
	}

	var hasChanges bool
	for field := range tfreflect.ExportedStructFields(planValue.Type()) {
		fieldName := field.Name

		if shouldSkipField(fieldName, opts.IgnoredFields) {
			ignoredFields = append(ignoredFields, fieldName)
			continue
		}

		if !implementsAttrValue(planValue.FieldByIndex(field.Index)) || !implementsAttrValue(stateValue.FieldByIndex(field.Index)) {
			continue
		}

		planFieldValue := planValue.FieldByIndex(field.Index).Interface().(attr.Value)
		stateFieldValue := stateValue.FieldByIndex(field.Index).Interface().(attr.Value)

		if !planFieldValue.Type(ctx).Equal(stateFieldValue.Type(ctx)) {
			continue
		}

		if !planFieldValue.Equal(stateFieldValue) {
			hasChanges = true
		} else {
			ignoredFields = append(ignoredFields, fieldName)
		}
	}

	result.hasChanges = hasChanges
	result.ignoredFieldNames = ignoredFields

	return &result, diags
}

func dereferencePointer(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return value.Elem()
	}
	return value
}

func shouldSkipField(fieldName string, ignoredFieldNames []string) bool {
	return slices.Contains(skippedFields(), fieldName) || slices.Contains(ignoredFieldNames, fieldName)
}

func implementsAttrValue(field reflect.Value) bool {
	return field.Type().Implements(reflect.TypeFor[attr.Value]())
}

func skippedFields() []string {
	return []string{
		"Tags",
		"TagsAll",
		"Timeouts",
	}
}
