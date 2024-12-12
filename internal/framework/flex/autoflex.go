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

type fieldNamePrefixCtxKey string

const (
	fieldNamePrefixRecurse fieldNamePrefixCtxKey = "FIELD_NAME_PREFIX_RECURSE"
	fieldNameSuffixRecurse fieldNamePrefixCtxKey = "FIELD_NAME_SUFFIX_RECURSE"

	mapBlockKeyFieldName = "MapBlockKey"
)

// Expand  = TF -->  AWS
// Flatten = AWS --> TF

// autoFlexer is the interface implemented by an auto-flattener or expander.
type autoFlexer interface {
	convert(context.Context, path.Path, reflect.Value, path.Path, reflect.Value, fieldOpts) diag.Diagnostics
	getOptions() AutoFlexOptions
}

// autoFlexValues returns the underlying `reflect.Value`s of `from` and `to`.
func autoFlexValues(ctx context.Context, from, to any) (context.Context, reflect.Value, reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	valFrom, valTo := reflect.ValueOf(from), reflect.ValueOf(to)
	if kind := valFrom.Kind(); kind == reflect.Pointer {
		valFrom = valFrom.Elem()
	}

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(valueType(valFrom)))
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(valueType(valTo)))

	kind := valTo.Kind()
	switch kind {
	case reflect.Pointer:
		if valTo.IsNil() {
			tflog.SubsystemError(ctx, subsystemName, "Target is nil")
			diags.Append(diagConvertingTargetIsNil(valTo.Type()))
			return ctx, reflect.Value{}, reflect.Value{}, diags
		}
		valTo = valTo.Elem()
		return ctx, valFrom, valTo, diags

	case reflect.Invalid:
		tflog.SubsystemError(ctx, subsystemName, "Target is nil")
		diags.Append(diagConvertingTargetIsNil(nil))
		return ctx, reflect.Value{}, reflect.Value{}, diags

	default:
		tflog.SubsystemError(ctx, subsystemName, "Target is not a pointer")
		diags.Append(diagConvertingTargetIsNotPointer(valTo.Type()))
		return ctx, reflect.Value{}, reflect.Value{}, diags
	}
}

var (
	plural = pluralize.NewClient()
)

// autoFlexConvertStruct traverses struct `from` calling `flexer` for each exported field.
func autoFlexConvertStruct(ctx context.Context, sourcePath path.Path, from any, targetPath path.Path, to any, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx, valFrom, valTo, d := autoFlexValues(ctx, from, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// TODO: this only applies when Expanding
	if fromExpander, ok := valFrom.Interface().(Expander); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Source implements flex.Expander")
		diags.Append(expandExpander(ctx, fromExpander, valTo)...)
		return diags
	}

	// TODO: this only applies when Expanding
	if fromTypedExpander, ok := valFrom.Interface().(TypedExpander); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Source implements flex.TypedExpander")
		diags.Append(expandTypedExpander(ctx, fromTypedExpander, valTo)...)
		return diags
	}

	// TODO: this only applies when Expanding
	if valTo.Kind() == reflect.Interface {
		tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
			"from": valFrom.Type(),
			"to":   valTo.Kind(),
		})
		return diags
	}

	// TODO: this only applies when Flattening
	if toFlattener, ok := to.(Flattener); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Target implements flex.Flattener")
		diags.Append(flattenFlattener(ctx, valFrom, toFlattener)...)
		return diags
	}

	typeFrom := valFrom.Type()
	typeTo := valTo.Type()

	opts := flexer.getOptions()
	for i := 0; i < typeFrom.NumField(); i++ {
		fromField := typeFrom.Field(i)
		if fromField.PkgPath != "" {
			continue // Skip unexported fields.
		}
		fromNameOverride, fromOpts := autoflexTags(fromField)
		fieldName := fromField.Name
		if opts.isIgnoredField(fieldName) {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored source field", map[string]any{
				logAttrKeySourceFieldname: fieldName,
			})
			continue
		}
		// TODO: this only applies when Expanding
		if fromNameOverride == "-" {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored source field", map[string]any{
				logAttrKeySourceFieldname: fieldName,
			})
			continue
		}
		if fieldName == mapBlockKeyFieldName {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping map block key", map[string]any{
				logAttrKeySourceFieldname: mapBlockKeyFieldName,
			})
			continue
		}

		toField, ok := findFieldFuzzy(ctx, fieldName, typeFrom, typeTo, flexer)
		if !ok {
			// Corresponding field not found in to.
			tflog.SubsystemDebug(ctx, subsystemName, "No corresponding field", map[string]any{
				logAttrKeySourceFieldname: fieldName,
			})
			continue
		}
		toFieldName := toField.Name
		// TODO: this only applies when Flattening
		toNameOverride, toOpts := autoflexTags(toField)
		toFieldVal := valTo.FieldByIndex(toField.Index)
		if toNameOverride == "-" {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored target field", map[string]any{
				logAttrKeySourceFieldname: fieldName,
				logAttrKeyTargetFieldname: toFieldName,
			})
			continue
		}
		if toOpts.NoFlatten() {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping noflatten target field", map[string]any{
				logAttrKeySourceFieldname: fieldName,
				logAttrKeyTargetFieldname: toFieldName,
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

		opts := fieldOpts{
			legacy:    fromOpts.Legacy() || toOpts.Legacy(),
			omitempty: toOpts.OmitEmpty(),
		}

		diags.Append(flexer.convert(ctx, sourcePath.AtName(fieldName), valFrom.Field(i), targetPath.AtName(toFieldName), toFieldVal, opts)...)
		if diags.HasError() {
			break
		}
	}

	return diags
}

func findFieldFuzzy(ctx context.Context, fieldNameFrom string, typeFrom reflect.Type, typeTo reflect.Type, flexer autoFlexer) (reflect.StructField, bool) {
	// first precedence is exact match (case sensitive)
	if fieldTo, ok := typeTo.FieldByName(fieldNameFrom); ok {
		return fieldTo, true
	}

	// If a "from" field fuzzy matches a "to" field, we are certain the fuzzy match
	// is NOT correct if "from" also contains a field by the fuzzy matched name.
	// For example, if "from" has "Value" and "Values", "Values" should *never*
	// fuzzy match "Value" in "to" since "from" also has "Value". We check "from"
	// to make sure fuzzy matches are not in "from".

	// second precedence is exact match (case insensitive)
	opts := flexer.getOptions()
	for i := 0; i < typeTo.NumField(); i++ {
		field := typeTo.Field(i)
		if field.PkgPath != "" {
			continue // Skip unexported fields.
		}
		fieldNameTo := field.Name
		if opts.isIgnoredField(fieldNameTo) {
			continue
		}
		if fieldTo, ok := typeTo.FieldByName(fieldNameTo); ok && strings.EqualFold(fieldNameFrom, fieldNameTo) && !fieldExistsInStruct(fieldNameTo, typeFrom) {
			// probably could assume validity here since reflect gave the field name
			return fieldTo, true
		}
	}

	// third precedence is singular/plural
	fieldNameTo := plural.Plural(fieldNameFrom)
	if plural.IsSingular(fieldNameFrom) && !fieldExistsInStruct(fieldNameTo, typeFrom) {
		if fieldTo, ok := typeTo.FieldByName(fieldNameTo); ok {
			return fieldTo, true
		}
	}

	fieldNameTo = plural.Singular(fieldNameFrom)
	if plural.IsPlural(fieldNameFrom) && !fieldExistsInStruct(fieldNameTo, typeFrom) {
		if fieldTo, ok := typeTo.FieldByName(fieldNameTo); ok {
			return fieldTo, true
		}
	}

	// fourth precedence is using field name prefix
	if v := opts.fieldNamePrefix; v != "" {
		v = strings.ReplaceAll(v, " ", "")
		if ctx.Value(fieldNamePrefixRecurse) == nil {
			// so it will only recurse once
			ctx = context.WithValue(ctx, fieldNamePrefixRecurse, true)
			if strings.HasPrefix(fieldNameFrom, v) {
				return findFieldFuzzy(ctx, strings.TrimPrefix(fieldNameFrom, v), typeFrom, typeTo, flexer)
			}
			return findFieldFuzzy(ctx, v+fieldNameFrom, typeFrom, typeTo, flexer)
		}
	}

	// fifth precedence is using field name suffix
	if v := opts.fieldNameSuffix; v != "" {
		v = strings.ReplaceAll(v, " ", "")
		if ctx.Value(fieldNameSuffixRecurse) == nil {
			// so it will only recurse once
			ctx = context.WithValue(ctx, fieldNameSuffixRecurse, true)
			if strings.HasSuffix(fieldNameFrom, v) {
				return findFieldFuzzy(ctx, strings.TrimSuffix(fieldNameFrom, v), typeFrom, typeTo, flexer)
			}
			return findFieldFuzzy(ctx, fieldNameFrom+v, typeFrom, typeTo, flexer)
		}
	}

	// no finds, fuzzy or otherwise - return zero value
	return reflect.StructField{}, false
}

func fieldExistsInStruct(field string, structType reflect.Type) bool {
	_, ok := structType.FieldByName(field)
	return ok
}

func autoflexTags(field reflect.StructField) (string, tagOptions) {
	return parseTag(field.Tag.Get("autoflex"))
}

type fieldOpts struct {
	legacy    bool
	omitempty bool
}

// valueWithElementsAs extends the Value interface for values that have an ElementsAs method.
type valueWithElementsAs interface {
	attr.Value

	Elements() []attr.Value
	ElementsAs(context.Context, any, bool) diag.Diagnostics
}

func diagConvertingTargetIsNil(targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while converting configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Target of type %q is nil", fullTypeName(targetType)),
	)
}

func diagConvertingTargetIsNotPointer(targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while converting configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Target type %q is not a pointer", fullTypeName(targetType)),
	)
}

func valueType(v reflect.Value) reflect.Type {
	if v.Kind() == reflect.Invalid {
		return nil
	}
	return v.Type()
}
