// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type AutoFlexCtxKey string

const (
	FieldNamePrefixRecurse AutoFlexCtxKey = "FIELD_NAME_PREFIX_RECURSE"

	MapBlockKey = "MapBlockKey"
)

// Expand  = TF -->  AWS
// Flatten = AWS --> TF

// autoFlexer is the interface implemented by an auto-flattener or expander.
type autoFlexer interface {
	convert(context.Context, path.Path, reflect.Value, path.Path, reflect.Value) diag.Diagnostics
	getOptions() AutoFlexOptions
}

// autoFlexValues returns the underlying `reflect.Value`s of `from` and `to`.
func autoFlexValues(_ context.Context, from, to any) (reflect.Value, reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	valFrom, valTo := reflect.ValueOf(from), reflect.ValueOf(to)
	if kind := valFrom.Kind(); kind == reflect.Ptr {
		valFrom = valFrom.Elem()
	}

	kind := valTo.Kind()
	switch kind {
	case reflect.Ptr:
		if valTo.IsNil() {
			diags.AddError("AutoFlEx", "Target cannot be nil")
			return reflect.Value{}, reflect.Value{}, diags
		}
		valTo = valTo.Elem()
		return valFrom, valTo, diags

	case reflect.Invalid:
		diags.AddError("AutoFlEx", "Target cannot be nil")
		return reflect.Value{}, reflect.Value{}, diags

	default:
		diags.AddError("AutoFlEx", fmt.Sprintf("target (%T): %s, want pointer", to, kind))
		return reflect.Value{}, reflect.Value{}, diags
	}
}

var (
	plural = pluralize.NewClient()
)

// autoFlexConvertStruct traverses struct `from` calling `flexer` for each exported field.
func autoFlexConvertStruct(ctx context.Context, sourcePath path.Path, from any, targetPath path.Path, to any, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(reflect.TypeOf(from)))
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(reflect.TypeOf(to)))

	valFrom, valTo, d := autoFlexValues(ctx, from, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if fromExpander, ok := valFrom.Interface().(Expander); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Source implements flex.Expander")
		diags.Append(expandExpander(ctx, fromExpander, valTo)...)
		return diags
	}

	if fromTypedExpander, ok := valFrom.Interface().(TypedExpander); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Source implements flex.TypedExpander")
		diags.Append(expandTypedExpander(ctx, fromTypedExpander, valTo)...)
		return diags
	}

	if valTo.Kind() == reflect.Interface {
		tflog.SubsystemInfo(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
			"from": valFrom.Type(),
			"to":   valTo.Kind(),
		})
		return diags
	}

	if toFlattener, ok := to.(Flattener); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Source implements flex.Flattener")
		diags.Append(flattenFlattener(ctx, valFrom, toFlattener)...)
		return diags
	}

	opts := flexer.getOptions()
	for i, typFrom := 0, valFrom.Type(); i < typFrom.NumField(); i++ {
		field := typFrom.Field(i)
		if field.PkgPath != "" {
			continue // Skip unexported fields.
		}
		fieldName := field.Name
		if opts.isIgnoredField(fieldName) {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored field", map[string]any{
				logAttrKeySourceFieldname: fieldName,
			})
			continue
		}
		if fieldName == MapBlockKey {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping map block key", map[string]any{
				logAttrKeySourceFieldname: MapBlockKey,
			})
			continue
		}

		toFieldVal, toFieldName := findFieldFuzzy(ctx, fieldName, valTo, valFrom, flexer)
		if !toFieldVal.IsValid() {
			// Corresponding field not found in to.
			tflog.SubsystemDebug(ctx, subsystemName, "No corresponding field", map[string]any{
				logAttrKeySourceFieldname: fieldName,
			})
			continue
		}
		if !toFieldVal.CanSet() {
			// Corresponding field value can't be changed.
			tflog.SubsystemDebug(ctx, subsystemName, "Field cannot be set", map[string]any{
				logAttrKeySourceFieldname: fieldName,
				logAttrKeyTargetFieldname: toFieldName,
			})
			continue
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Matched fields", map[string]any{
			logAttrKeySourceFieldname: fieldName,
			logAttrKeyTargetFieldname: toFieldName,
		})

		diags.Append(flexer.convert(ctx, sourcePath.AtName(fieldName), valFrom.Field(i), targetPath.AtName(toFieldName), toFieldVal)...)
		if diags.HasError() {
			diags.AddError("AutoFlEx", fmt.Sprintf("convert (%s)", fieldName))
			return diags
		}
	}

	return diags
}

func findFieldFuzzy(ctx context.Context, fieldNameFrom string, valTo, valFrom reflect.Value, flexer autoFlexer) (reflect.Value, string) {
	// first precedence is exact match (case sensitive)
	if v := valTo.FieldByName(fieldNameFrom); v.IsValid() {
		return v, fieldNameFrom
	}

	// If a "from" field fuzzy matches a "to" field, we are certain the fuzzy match
	// is NOT correct if "from" also contains a field by the fuzzy matched name.
	// For example, if "from" has "Value" and "Values", "Values" should *never*
	// fuzzy match "Value" in "to" since "from" also has "Value". We check "from"
	// to make sure fuzzy matches are not in "from".

	// second precedence is exact match (case insensitive)
	opts := flexer.getOptions()
	for i, typTo := 0, valTo.Type(); i < typTo.NumField(); i++ {
		field := typTo.Field(i)
		if field.PkgPath != "" {
			continue // Skip unexported fields.
		}
		fieldNameTo := field.Name
		if opts.isIgnoredField(fieldNameTo) {
			continue
		}
		if v := valTo.FieldByName(fieldNameTo); v.IsValid() && strings.EqualFold(fieldNameFrom, fieldNameTo) && !fieldExistsInStruct(fieldNameTo, valFrom) {
			// probably could assume validity here since reflect gave the field name
			return v, fieldNameTo
		}
	}

	// third precedence is singular/plural
	fieldNameTo := plural.Plural(fieldNameFrom)
	if plural.IsSingular(fieldNameFrom) && !fieldExistsInStruct(fieldNameTo, valFrom) {
		if v := valTo.FieldByName(fieldNameTo); v.IsValid() {
			return v, fieldNameTo
		}
	}

	fieldNameTo = plural.Singular(fieldNameFrom)
	if plural.IsPlural(fieldNameFrom) && !fieldExistsInStruct(fieldNameTo, valFrom) {
		if v := valTo.FieldByName(fieldNameTo); v.IsValid() {
			return v, fieldNameTo
		}
	}

	// fourth precedence is using resource prefix
	if v := opts.fieldNamePrefix; v != "" {
		v = strings.ReplaceAll(v, " ", "")
		if ctx.Value(FieldNamePrefixRecurse) == nil {
			// so it will only recurse once
			ctx = context.WithValue(ctx, FieldNamePrefixRecurse, true)
			if strings.HasPrefix(fieldNameFrom, v) {
				return findFieldFuzzy(ctx, strings.TrimPrefix(fieldNameFrom, v), valTo, valFrom, flexer)
			}
			return findFieldFuzzy(ctx, v+fieldNameFrom, valTo, valFrom, flexer)
		}
	}

	// no finds, fuzzy or otherwise - return zero value
	return reflect.Value{}, ""
}

func fieldExistsInStruct(field string, structVal reflect.Value) bool {
	v := structVal.FieldByName(field)
	return v.IsValid()
}

// valueWithElementsAs extends the Value interface for values that have an ElementsAs method.
type valueWithElementsAs interface {
	attr.Value

	Elements() []attr.Value
	ElementsAs(context.Context, any, bool) diag.Diagnostics
}
