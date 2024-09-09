// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diff

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

type Results struct {
	hasChanges            bool
	ignoredFieldNames     []string
	flexIgnoredFieldNames []fwflex.AutoFlexOptionsFunc
}

// HasChanges returns whether there are changes between the plan and state values
func (r *Results) HasChanges() bool {
	return r.hasChanges
}

// FlexIgnoredFieldNames returns the list of ignored field names as AutoFlexOptionsFunc
func (r *Results) FlexIgnoredFieldNames() []fwflex.AutoFlexOptionsFunc {
	for _, v := range r.ignoredFieldNames {
		r.flexIgnoredFieldNames = append(r.flexIgnoredFieldNames, fwflex.WithIgnoredFieldNamesAppend(v))
	}

	return r.flexIgnoredFieldNames
}

func (r *Results) IgnoredFieldNames() []string {
	return r.ignoredFieldNames
}

// Calculate compares the plan and state values and returns whether there are changes
func Calculate(ctx context.Context, plan, state any, options ...ChangeOptionsFunc) (*Results, diag.Diagnostics) {
	var diags diag.Diagnostics
	opts := initChangeOptions(options)

	p, s := setValue(reflect.ValueOf(plan)), setValue(reflect.ValueOf(state))
	typeOfP, typesOfS := p.Type(), s.Type()
	var ignoredFields []string
	result := Results{}

	if typeOfP != typesOfS {
		diags.AddError(
			"Type mismatch between plan and state",
			fmt.Sprintf("plan type: %s, state type: %s", typeOfP.String(), typesOfS.String()),
		)
		return &result, diags
	}

	var hasChanges bool
	// check every field on the plan struct
	for i := 0; i < p.NumField(); i++ {
		name := typeOfP.Field(i).Name

		// skip fields that are handled outside of regular update flow
		if slices.Contains(skippedFields(), name) {
			continue
		}

		if slices.Contains(opts.ignoredFieldNames, name) {
			ignoredFields = append(ignoredFields, name)
			continue
		}

		// if the field is not present in the state, skip it
		_, sHasField := typesOfS.FieldByName(name)
		if !sHasField {
			continue
		}

		// if the fields do not implement the correct interfaces, skip it
		fieldType := reflect.TypeFor[attr.Value]()
		if !p.FieldByName(name).Type().Implements(fieldType) || !s.FieldByName(name).Type().Implements(fieldType) {
			continue
		}

		pValue := p.FieldByName(name).Interface().(attr.Value)
		sValue := s.FieldByName(name).Interface().(attr.Value)

		// check that the types are the same so that they can be compared
		if !pValue.Type(ctx).Equal(sValue.Type(ctx)) {
			continue
		}

		if ok := !pValue.Equal(sValue); ok {
			hasChanges = ok
		} else {
			ignoredFields = append(ignoredFields, name)
		}
	}

	result.hasChanges = hasChanges
	result.ignoredFieldNames = ignoredFields

	return &result, diags
}

func setValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Ptr {
		return value.Elem()
	}
	return value
}

func skippedFields() []string {
	return []string{
		"Tags",
		"TagsAll",
		"Timeouts",
	}
}
