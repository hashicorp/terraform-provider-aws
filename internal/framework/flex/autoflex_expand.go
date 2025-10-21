// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfreflect "github.com/hashicorp/terraform-provider-aws/internal/reflect"
)

// Expand  = TF -->  AWS

// Expander is implemented by types that customize their expansion
type Expander interface {
	Expand(ctx context.Context) (any, diag.Diagnostics)
}

// Expander is implemented by types that customize their expansion and can have multiple target types
type TypedExpander interface {
	ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics)
}

// Expand "expands" a resource's "business logic" data structure,
// implemented using Terraform Plugin Framework data types, into
// an AWS SDK for Go v2 API data structure.
// The resource's data structure is walked and exported fields that
// have a corresponding field in the API data structure (and a suitable
// target data type) are copied.
func Expand(ctx context.Context, tfObject, apiObject any, optFns ...AutoFlexOptionsFunc) diag.Diagnostics {
	var diags diag.Diagnostics
	expander := newAutoExpander(optFns)

	tflog.SubsystemInfo(ctx, subsystemName, "Expanding", map[string]any{
		logAttrKeySourceType: fullTypeName(reflect.TypeOf(tfObject)),
		logAttrKeyTargetType: fullTypeName(reflect.TypeOf(apiObject)),
	})

	diags.Append(autoExpandConvert(ctx, tfObject, apiObject, expander)...)

	return diags
}

type autoExpander struct {
	Options AutoFlexOptions
}

// newAutoExpander initializes an auto-expander with defaults that can be overridden
// via functional options
func newAutoExpander(optFns []AutoFlexOptionsFunc) *autoExpander {
	o := AutoFlexOptions{
		ignoredFieldNames: DefaultIgnoredFieldNames,
	}

	for _, optFn := range optFns {
		optFn(&o)
	}

	return &autoExpander{
		Options: o,
	}
}

func (expander autoExpander) getOptions() AutoFlexOptions {
	return expander.Options
}

// autoFlexConvert converts `from` to `to` using the specified auto-flexer.
func autoExpandConvert(ctx context.Context, from, to any, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	sourcePath := path.Empty()
	targetPath := path.Empty()

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx, valFrom, valTo, d := autoFlexValues(ctx, from, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Top-level struct to struct conversion.
	if valFrom.IsValid() && valTo.IsValid() {
		if typFrom, typTo := valFrom.Type(), valTo.Type(); typFrom.Kind() == reflect.Struct && typTo.Kind() == reflect.Struct &&
			!typFrom.Implements(reflect.TypeFor[basetypes.ListValuable]()) &&
			!typFrom.Implements(reflect.TypeFor[basetypes.SetValuable]()) {
			tflog.SubsystemInfo(ctx, subsystemName, "Converting")
			diags.Append(expandStruct(ctx, sourcePath, from, targetPath, to, flexer)...)
			return diags
		}
	}

	// Anything else.
	diags.Append(flexer.convert(ctx, sourcePath, valFrom, targetPath, valTo, fieldOpts{})...)
	return diags
}

// convert converts a single Plugin Framework value to its AWS API equivalent.
func (expander autoExpander) convert(ctx context.Context, sourcePath path.Path, valFrom reflect.Value, targetPath path.Path, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(valueType(valFrom)))
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(valueType(vTo)))

	if valFrom.Kind() == reflect.Invalid {
		tflog.SubsystemError(ctx, subsystemName, "Source is nil")
		diags.Append(diagExpandingSourceIsNil(valueType(valFrom)))
		return diags
	}

	tflog.SubsystemInfo(ctx, subsystemName, "Converting")

	if fromExpander, ok := valFrom.Interface().(Expander); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Source implements flex.Expander")
		diags.Append(expandExpander(ctx, fromExpander, vTo)...)
		return diags
	}

	if fromTypedExpander, ok := valFrom.Interface().(TypedExpander); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Source implements flex.TypedExpander")
		diags.Append(expandTypedExpander(ctx, fromTypedExpander, vTo)...)
		return diags
	}

	vFrom, ok := valFrom.Interface().(attr.Value)
	if !ok {
		tflog.SubsystemError(ctx, subsystemName, "Source does not implement attr.Value")
		diags.Append(diagExpandingSourceDoesNotImplementAttrValue(reflect.TypeOf(valFrom.Interface())))
		return diags
	}

	// No need to set the target value if there's no source value.
	if vFrom.IsNull() {
		tflog.SubsystemTrace(ctx, subsystemName, "Expanding null value")
		return diags
	}
	if vFrom.IsUnknown() {
		tflog.SubsystemTrace(ctx, subsystemName, "Expanding unknown value")
		return diags
	}

	switch vFrom := vFrom.(type) {
	// Primitive types.
	case basetypes.BoolValuable:
		diags.Append(expander.bool(ctx, vFrom, vTo, fieldOpts)...)
		return diags

	case basetypes.Float64Valuable:
		diags.Append(expander.float64(ctx, vFrom, vTo, fieldOpts)...)
		return diags

	case basetypes.Float32Valuable:
		diags.Append(expander.float32(ctx, vFrom, vTo, fieldOpts)...)
		return diags

	case basetypes.Int64Valuable:
		diags.Append(expander.int64(ctx, vFrom, vTo, fieldOpts)...)
		return diags

	case basetypes.Int32Valuable:
		diags.Append(expander.int32(ctx, vFrom, vTo, fieldOpts)...)
		return diags

	case basetypes.StringValuable:
		diags.Append(expander.string(ctx, vFrom, vTo, fieldOpts)...)
		return diags

	// Aggregate types.
	case basetypes.ObjectValuable:
		diags.Append(expander.object(ctx, sourcePath, vFrom, targetPath, vTo)...)
		return diags

	case basetypes.ListValuable:
		diags.Append(expander.list(ctx, sourcePath, vFrom, targetPath, vTo)...)
		return diags

	case basetypes.MapValuable:
		diags.Append(expander.map_(ctx, vFrom, vTo)...)
		return diags

	case basetypes.SetValuable:
		diags.Append(expander.set(ctx, sourcePath, vFrom, targetPath, vTo)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// bool copies a Plugin Framework Bool(ish) value to a compatible AWS API value.
func (expander autoExpander) bool(ctx context.Context, vFrom basetypes.BoolValuable, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToBoolValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Bool:
		//
		// types.Bool -> bool.
		//
		vTo.SetBool(v.ValueBool())
		return diags

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Bool:
			//
			// types.Bool -> *bool.
			//
			if fieldOpts.legacy {
				tflog.SubsystemDebug(ctx, subsystemName, "Using legacy expander")
				if !v.ValueBool() {
					return diags
				}
			}
			vTo.Set(reflect.ValueOf(v.ValueBoolPointer()))
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// float64 copies a Plugin Framework Float64(ish) value to a compatible AWS API value.
func (expander autoExpander) float64(ctx context.Context, vFrom basetypes.Float64Valuable, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToFloat64Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Float32, reflect.Float64:
		//
		// types.Float32/types.Float64 -> float32/float64.
		//
		vTo.SetFloat(v.ValueFloat64())
		return diags

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Float32:
			//
			// types.Float32/types.Float64 -> *float32.
			//
			to := float32(v.ValueFloat64())
			if fieldOpts.legacy {
				tflog.SubsystemDebug(ctx, subsystemName, "Using legacy expander")
				if to == 0 {
					return diags
				}
			}
			vTo.Set(reflect.ValueOf(&to))
			return diags

		case reflect.Float64:
			//
			// types.Float32/types.Float64 -> *float64.
			//
			if fieldOpts.legacy {
				tflog.SubsystemDebug(ctx, subsystemName, "Using legacy expander")
				if v.ValueFloat64() == 0 {
					return diags
				}
			}
			vTo.Set(reflect.ValueOf(v.ValueFloat64Pointer()))
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// float32 copies a Plugin Framework Float32(ish) value to a compatible AWS API value.
func (expander autoExpander) float32(ctx context.Context, vFrom basetypes.Float32Valuable, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToFloat32Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Float32:
		//
		// types.Float32 -> float32.
		//
		vTo.SetFloat(float64(v.ValueFloat32()))
		return diags

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Float32:
			//
			// types.Float32 -> *float32.
			//
			if fieldOpts.legacy {
				tflog.SubsystemDebug(ctx, subsystemName, "Using legacy expander")
				if v.ValueFloat32() == 0 {
					return diags
				}
			}
			vTo.Set(reflect.ValueOf(v.ValueFloat32Pointer()))
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "Expanding incompatible types")
	diags.Append(diagExpandingIncompatibleTypes(reflect.TypeOf(vFrom), vTo.Type()))
	return diags
}

// int64 copies a Plugin Framework Int64(ish) value to a compatible AWS API value.
func (expander autoExpander) int64(ctx context.Context, vFrom basetypes.Int64Valuable, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToInt64Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Int32, reflect.Int64:
		//
		// types.Int32/types.Int64 -> int32/int64.
		//
		vTo.SetInt(v.ValueInt64())
		return diags

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Int32:
			//
			// types.Int32/types.Int64 -> *int32.
			//
			to := int32(v.ValueInt64())
			if fieldOpts.legacy {
				tflog.SubsystemDebug(ctx, subsystemName, "Using legacy expander")
				if to == 0 {
					return diags
				}
			}
			vTo.Set(reflect.ValueOf(&to))
			return diags

		case reflect.Int64:
			//
			// types.Int32/types.Int64 -> *int64.
			//
			if fieldOpts.legacy {
				tflog.SubsystemDebug(ctx, subsystemName, "Using legacy expander")
				if v.ValueInt64() == 0 {
					return diags
				}
			}
			vTo.Set(reflect.ValueOf(v.ValueInt64Pointer()))
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// int32 copies a Plugin Framework Int32(ish) value to a compatible AWS API value.
func (expander autoExpander) int32(ctx context.Context, vFrom basetypes.Int32Valuable, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToInt32Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Int32:
		//
		// types.Int32 -> int32.
		//
		vTo.SetInt(int64(v.ValueInt32()))
		return diags

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Int32:
			//
			// types.Int32 -> *int32.
			//
			if fieldOpts.legacy {
				tflog.SubsystemDebug(ctx, subsystemName, "Using legacy expander")
				if v.ValueInt32() == 0 {
					return diags
				}
			}
			vTo.Set(reflect.ValueOf(v.ValueInt32Pointer()))
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "Expanding incompatible types")
	diags.Append(diagExpandingIncompatibleTypes(reflect.TypeOf(vFrom), vTo.Type()))
	return diags
}

// string copies a Plugin Framework String(ish) value to a compatible AWS API value.
func (expander autoExpander) string(ctx context.Context, vFrom basetypes.StringValuable, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.String:
		//
		// types.String -> string.
		//
		vTo.SetString(v.ValueString())
		return diags

	case reflect.Slice:
		if tTo.Elem().Kind() == reflect.Uint8 {
			//
			// types.String -> []byte (or []uint8).
			//
			vTo.Set(reflect.ValueOf([]byte(v.ValueString())))
			return diags
		}

	case reflect.Struct:
		//
		// timetypes.RFC3339 --> time.Time
		//
		if t, ok := vFrom.(timetypes.RFC3339); ok {
			v, d := t.ValueRFC3339Time()
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(v))
			return diags
		}

	case reflect.Interface:
		if s, ok := vFrom.(fwtypes.SmithyDocumentValue); ok {
			v, d := s.ToSmithyObjectDocument(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(v))
			return diags
		}

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.String:
			//
			// types.String -> *string.
			//
			if fieldOpts.legacy {
				tflog.SubsystemDebug(ctx, subsystemName, "Using legacy expander")
				if len(v.ValueString()) == 0 {
					return diags
				}
			}
			vTo.Set(reflect.ValueOf(v.ValueStringPointer()))
			return diags

		case reflect.Struct:
			//
			// timetypes.RFC3339 --> *time.Time
			//
			if t, ok := vFrom.(timetypes.RFC3339); ok {
				v, d := t.ValueRFC3339Time()
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(&v))
				return diags
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// string copies a Plugin Framework Object(ish) value to a compatible AWS API value.
func (expander autoExpander) object(ctx context.Context, sourcePath path.Path, vFrom basetypes.ObjectValuable, targetPath path.Path, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	_, d := vFrom.ToObjectValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Struct:
		//
		// types.Object -> struct.
		//
		if vFrom, ok := vFrom.(fwtypes.NestedObjectValue); ok {
			diags.Append(expander.nestedObjectToStruct(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
			return diags
		}

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			//
			// types.Object --> *struct
			//
			if vFrom, ok := vFrom.(fwtypes.NestedObjectValue); ok {
				diags.Append(expander.nestedObjectToStruct(ctx, sourcePath, vFrom, targetPath, tElem, vTo)...)
				return diags
			}
		}

	case reflect.Interface:
		//
		// types.Object -> interface.
		//
		if vFrom, ok := vFrom.(fwtypes.NestedObjectValue); ok {
			diags.Append(expander.nestedObjectToStruct(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// list copies a Plugin Framework List(ish) value to a compatible AWS API value.
func (expander autoExpander) list(ctx context.Context, sourcePath path.Path, vFrom basetypes.ListValuable, targetPath path.Path, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToListValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.Int64Typable:
		diags.Append(expander.listOrSetOfInt64(ctx, v, vTo)...)
		return diags

	case basetypes.StringTypable:
		diags.Append(expander.listOrSetOfString(ctx, v, vTo)...)
		return diags

	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.NestedObjectCollectionValue); ok {
			diags.Append(expander.nestedObjectCollection(ctx, sourcePath, vFrom, targetPath, vTo)...)
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from list[%s]": v.ElementType(ctx),
		"to":            vTo.Kind(),
	})

	return diags
}

// listOrSetOfInt64 copies a Plugin Framework ListOfInt64(ish) or SetOfInt64(ish) value to a compatible AWS API value.
func (expander autoExpander) listOrSetOfInt64(ctx context.Context, vFrom valueWithElementsAs, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Struct:
		// Check if target is an XML wrapper struct
		if isXMLWrapperStruct(vTo.Type()) {
			diags.Append(expander.xmlWrapper(ctx, vFrom, vTo, "Items")...)
			return diags
		}

	case reflect.Slice:
		switch tSliceElem := vTo.Type().Elem(); tSliceElem.Kind() {
		case reflect.Int32, reflect.Int64:
			//
			// types.List(OfInt64) -> []int64 or []int32
			//
			tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
				logAttrKeySourceSize: len(vFrom.Elements()),
			})
			var to []int64
			diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
			if diags.HasError() {
				return diags
			}

			vals := reflect.MakeSlice(vTo.Type(), len(to), len(to))
			for i := range to {
				vals.Index(i).SetInt(to[i])
			}
			vTo.Set(vals)
			return diags

		case reflect.Pointer:
			switch tSliceElem.Elem().Kind() {
			case reflect.Int32:
				//
				// types.List(OfInt64) -> []*int32.
				//
				tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
					logAttrKeySourceSize: len(vFrom.Elements()),
				})
				var to []*int32
				diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags

			case reflect.Int64:
				//
				// types.List(OfInt64) -> []*int64.
				//
				tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
					logAttrKeySourceSize: len(vFrom.Elements()),
				})
				var to []*int64
				diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// listOrSetOfString copies a Plugin Framework ListOfString(ish) or SetOfString(ish) value to a compatible AWS API value.
func (expander autoExpander) listOrSetOfString(ctx context.Context, vFrom valueWithElementsAs, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Struct:
		// Check if target is an XML wrapper struct
		if isXMLWrapperStruct(vTo.Type()) {
			diags.Append(expander.xmlWrapper(ctx, vFrom, vTo, "Items")...)
			return diags
		}

	case reflect.Slice:
		switch tSliceElem := vTo.Type().Elem(); tSliceElem.Kind() {
		case reflect.String:
			//
			// types.List(OfString) -> []string.
			//
			tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
				logAttrKeySourceSize: len(vFrom.Elements()),
			})
			var to []string
			diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
			if diags.HasError() {
				return diags
			}

			// Copy elements individually to enable expansion of lists of
			// custom string types (AWS enums)
			vals := reflect.MakeSlice(vTo.Type(), len(to), len(to))
			for i := range to {
				vals.Index(i).SetString(to[i])
			}
			vTo.Set(vals)
			return diags

		case reflect.Pointer:
			switch tSliceElem.Elem().Kind() {
			case reflect.String:
				//
				// types.List(OfString) -> []*string.
				//
				tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
					logAttrKeySourceSize: len(vFrom.Elements()),
				})
				var to []*string
				diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}
		}

	case reflect.Pointer:
		switch tElem := vTo.Type().Elem(); tElem.Kind() {
		case reflect.Struct:
			// Check if target is a pointer to an XML wrapper struct
			if isXMLWrapperStruct(tElem) {
				// Create new instance of the XML wrapper struct
				newStruct := reflect.New(tElem).Elem()
				diags.Append(expander.xmlWrapper(ctx, vFrom, newStruct, "Items")...)
				if !diags.HasError() {
					vTo.Set(newStruct.Addr())
				}
				return diags
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": vFrom.Type(ctx),
		"to":   vTo.Kind(),
	})

	return diags
}

// map_ copies a Plugin Framework Map(ish) value to a compatible AWS API value.
func (expander autoExpander) map_(ctx context.Context, vFrom basetypes.MapValuable, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToMapValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.StringTypable:
		diags.Append(expander.mapOfString(ctx, v, vTo)...)
		return diags

	case basetypes.MapTypable:
		data, d := v.ToMapValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		switch tMapElem := vTo.Type().Elem(); tMapElem.Kind() {
		case reflect.Map:
			//
			// types.Map(OfMap) -> map[string]map[string]string.
			//
			switch k := tMapElem.Elem().Kind(); k {
			case reflect.String:
				tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
					logAttrKeySourceSize: len(data.Elements()),
				})
				var out map[string]map[string]string
				diags.Append(data.ElementsAs(ctx, &out, false)...)

				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(out))
				return diags

			case reflect.Pointer:
				switch k := tMapElem.Elem().Elem().Kind(); k {
				case reflect.String:
					//
					// types.Map(OfMap) -> map[string]map[string]*string.
					//
					tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
						logAttrKeySourceSize: len(data.Elements()),
					})
					var to map[string]map[string]*string
					diags.Append(data.ElementsAs(ctx, &to, false)...)

					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": fmt.Sprintf("map[string, %s]", v.ElementType(ctx)),
		"to":   vTo.Kind(),
	})

	return diags
}

// mapOfString copies a Plugin Framework MapOfString(ish) value to a compatible AWS API value.
func (expander autoExpander) mapOfString(ctx context.Context, vFrom basetypes.MapValue, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Map:
		switch tMapKey := vTo.Type().Key(); tMapKey.Kind() {
		case reflect.String: // key
			switch tMapElem := vTo.Type().Elem(); tMapElem.Kind() {
			case reflect.String:
				//
				// types.Map(OfString) -> map[string]string.
				//
				tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
					logAttrKeySourceSize: len(vFrom.Elements()),
				})
				var to map[string]string
				diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags

			case reflect.Pointer:
				switch k := tMapElem.Elem().Kind(); k {
				case reflect.String:
					//
					// types.Map(OfString) -> map[string]*string.
					//
					tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
						logAttrKeySourceSize: len(vFrom.Elements()),
					})
					var to map[string]*string
					diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": fmt.Sprintf("map[string, %s]", vFrom.ElementType(ctx)),
		"to":   vTo.Kind(),
	})

	return diags
}

// set copies a Plugin Framework Set(ish) value to a compatible AWS API value.
func (expander autoExpander) set(ctx context.Context, sourcePath path.Path, vFrom basetypes.SetValuable, targetPath path.Path, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToSetValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.Int64Typable:
		diags.Append(expander.listOrSetOfInt64(ctx, v, vTo)...)
		return diags

	case basetypes.StringTypable:
		diags.Append(expander.listOrSetOfString(ctx, v, vTo)...)
		return diags

	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.NestedObjectCollectionValue); ok {
			diags.Append(expander.nestedObjectCollection(ctx, sourcePath, vFrom, targetPath, vTo)...)
			return diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from set[%s]": v.ElementType(ctx),
		"to":           vTo.Kind(),
	})

	return diags
}

// nestedObjectCollection copies a Plugin Framework NestedObjectCollectionValue value to a compatible AWS API value.
func (expander autoExpander) nestedObjectCollection(ctx context.Context, sourcePath path.Path, vFrom fwtypes.NestedObjectCollectionValue, targetPath path.Path, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Struct:
		// Check if target is an XML wrapper struct before handling as generic struct
		if isXMLWrapperStruct(tTo) {
			// Don't populate XML wrapper struct if source is null or unknown
			if vFrom.IsNull() || vFrom.IsUnknown() {
				// Leave the struct as zero value (all fields nil/empty)
				return diags
			}

			diags.Append(expander.nestedObjectCollectionToXMLWrapper(ctx, sourcePath, vFrom, targetPath, vTo)...)
			return diags
		}

		sourcePath := sourcePath.AtListIndex(0)
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		diags.Append(expander.nestedObjectToStruct(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
		return diags

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			// Check if target is a pointer to XML wrapper struct
			if isXMLWrapperStruct(tElem) {
				// Don't create XML wrapper struct if source is null or unknown
				if vFrom.IsNull() || vFrom.IsUnknown() {
					tflog.SubsystemDebug(ctx, subsystemName, "Skipping XML wrapper creation - source is null/unknown", map[string]any{
						"target_type": tElem.String(),
					})
					// Leave the pointer as nil (zero value)
					return diags
				}

				tflog.SubsystemDebug(ctx, subsystemName, "Creating XML wrapper struct", map[string]any{
					"target_type":    tElem.String(),
					"source_null":    vFrom.IsNull(),
					"source_unknown": vFrom.IsUnknown(),
				})

				// Create new instance of the XML wrapper struct
				newWrapper := reflect.New(tElem)
				diags.Append(expander.nestedObjectCollectionToXMLWrapper(ctx, sourcePath, vFrom, targetPath, newWrapper.Elem())...)
				if !diags.HasError() {
					vTo.Set(newWrapper)
				}
				return diags
			}

			//
			// types.List(OfObject) -> *struct.
			//
			sourcePath := sourcePath.AtListIndex(0)
			ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
			diags.Append(expander.nestedObjectToStruct(ctx, sourcePath, vFrom, targetPath, tElem, vTo)...)
			return diags
		}

	case reflect.Interface:
		//
		// types.List(OfObject) -> interface.
		//
		sourcePath := sourcePath.AtListIndex(0)
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		diags.Append(expander.nestedObjectToStruct(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
		return diags

	case reflect.Map:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			//
			// types.List(OfObject) -> map[string]struct
			//
			diags.Append(expander.nestedKeyObjectToMap(ctx, sourcePath, vFrom, targetPath, tElem, vTo)...)
			return diags

		case reflect.Pointer:
			//
			// types.List(OfObject) -> map[string]*struct
			//
			diags.Append(expander.nestedKeyObjectToMap(ctx, sourcePath, vFrom, targetPath, tElem, vTo)...)
			return diags
		}

	case reflect.Slice:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			//
			// types.List(OfObject) -> []struct
			//
			diags.Append(expander.nestedObjectCollectionToSlice(ctx, sourcePath, vFrom, targetPath, tTo, tElem, vTo)...)
			return diags

		case reflect.Pointer:
			switch tElem := tElem.Elem(); tElem.Kind() {
			case reflect.Struct:
				//
				// types.List(OfObject) -> []*struct.
				//
				diags.Append(expander.nestedObjectCollectionToSlice(ctx, sourcePath, vFrom, targetPath, tTo, tElem, vTo)...)
				return diags
			}

		case reflect.Interface:
			//
			// types.List(OfObject) -> []interface.
			//
			diags.Append(expander.nestedObjectCollectionToSlice(ctx, sourcePath, vFrom, targetPath, tTo, tElem, vTo)...)
			return diags
		}
	}

	diags.AddError("Incompatible types", fmt.Sprintf("nestedObjectCollection[%s] cannot be expanded to %s", vFrom.Type(ctx).(attr.TypeWithElementType).ElementType(), vTo.Kind()))
	return diags
}

// nestedObjectToStruct copies a Plugin Framework NestedObjectValue to a compatible AWS API (*)struct value.
func (expander autoExpander) nestedObjectToStruct(ctx context.Context, sourcePath path.Path, vFrom fwtypes.NestedObjectValue, targetPath path.Path, tStruct reflect.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the nested Object as a pointer.
	from, d := vFrom.ToObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Create a new target structure and walk its fields.
	to := reflect.New(tStruct)
	if !reflect.ValueOf(from).IsNil() {
		diags.Append(expandStruct(ctx, sourcePath, from, targetPath, to.Interface(), expander)...)
		if diags.HasError() {
			return diags
		}
	}

	// Set value.
	switch vTo.Type().Kind() {
	case reflect.Struct, reflect.Interface:
		vTo.Set(to.Elem())

	default:
		vTo.Set(to)
	}

	return diags
}

// nestedObjectCollectionToSlice copies a Plugin Framework NestedObjectCollectionValue to a compatible AWS API [](*)struct value.
func (expander autoExpander) nestedObjectCollectionToSlice(ctx context.Context, sourcePath path.Path, vFrom fwtypes.NestedObjectCollectionValue, targetPath path.Path, tSlice, tElem reflect.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the nested Objects as a slice.
	from, d := vFrom.ToObjectSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Create a new target slice and expand each element.
	f := reflect.ValueOf(from)
	n := f.Len()

	tflog.SubsystemTrace(ctx, subsystemName, "Expanding nested object collection", map[string]any{
		logAttrKeySourceSize: n,
	})

	t := reflect.MakeSlice(tSlice, n, n)
	for i := range n {
		sourcePath := sourcePath.AtListIndex(i)
		targetPath := targetPath.AtListIndex(i)
		ctx := tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())
		// Create a new target structure and walk its fields.
		target := reflect.New(tElem)
		diags.Append(expandStruct(ctx, sourcePath, f.Index(i).Interface(), targetPath, target.Interface(), expander)...)
		if diags.HasError() {
			return diags
		}

		// Set value in the target slice.
		switch vTo.Type().Elem().Kind() {
		case reflect.Struct, reflect.Interface:
			t.Index(i).Set(target.Elem())

		default:
			t.Index(i).Set(target)
		}
	}

	vTo.Set(t)

	return diags
}

// nestedKeyObjectToMap copies a Plugin Framework NestedObjectCollectionValue to a compatible AWS API map[string]struct value.
func (expander autoExpander) nestedKeyObjectToMap(ctx context.Context, sourcePath path.Path, vFrom fwtypes.NestedObjectCollectionValue, targetPath path.Path, tElem reflect.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the nested Objects as a slice.
	from, d := vFrom.ToObjectSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if tElem.Kind() == reflect.Pointer {
		tElem = tElem.Elem()
	}

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(tElem))

	// Create a new target slice and expand each element.
	f := reflect.ValueOf(from)
	m := reflect.MakeMap(vTo.Type())
	for i := range f.Len() {
		sourcePath := sourcePath.AtListIndex(i)
		ctx := tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())

		key, d := mapBlockKey(ctx, f.Index(i).Interface())
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		targetPath := targetPath.AtMapKey(key.Interface().(string))
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

		// Create a new target structure and walk its fields.
		target := reflect.New(tElem)
		diags.Append(expandStruct(ctx, sourcePath, f.Index(i).Interface(), targetPath, target.Interface(), expander)...)
		if diags.HasError() {
			return diags
		}

		// Set value (or pointer) in the target map.
		if vTo.Type().Elem().Kind() == reflect.Struct {
			m.SetMapIndex(key, target.Elem())
		} else {
			m.SetMapIndex(key, target)
		}
	}

	vTo.Set(m)

	return diags
}

// expandStruct traverses struct `from`, calling `flexer` for each exported field.
func expandStruct(ctx context.Context, sourcePath path.Path, from any, targetPath path.Path, to any, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx, valFrom, valTo, d := autoFlexValues(ctx, from, to)
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
		tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
			"from": valFrom.Type(),
			"to":   valTo.Kind(),
		})
		return diags
	}

	typeFrom := valFrom.Type()
	typeTo := valTo.Type()

	// Handle XML wrapper collapse patterns where multiple source fields
	// need to be combined into a single complex target field
	processedFields := make(map[string]bool)
	diags.Append(flexer.handleXMLWrapperCollapse(ctx, sourcePath, valFrom, targetPath, valTo, typeFrom, typeTo, processedFields)...)
	if diags.HasError() {
		return diags
	}

	for fromField := range expandSourceFields(ctx, typeFrom, flexer.getOptions()) {
		fromFieldName := fromField.Name
		_, fromFieldOpts := autoflexTags(fromField)

		// Skip fields that were already processed by XML wrapper collapse
		if processedFields[fromFieldName] {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping field already processed by XML wrapper collapse", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
			})
			continue
		}

		toField, ok := (&fuzzyFieldFinder{}).findField(ctx, fromFieldName, typeFrom, typeTo, flexer)
		if !ok {
			// Corresponding field not found in to.
			tflog.SubsystemDebug(ctx, subsystemName, "No corresponding field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
			})
			continue
		}
		toFieldName := toField.Name
		toFieldVal := valTo.FieldByIndex(toField.Index)
		if !toFieldVal.CanSet() {
			// Corresponding field value can't be changed.
			tflog.SubsystemDebug(ctx, subsystemName, "Field cannot be set", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
				logAttrKeyTargetFieldname: toFieldName,
			})
			continue
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Matched fields", map[string]any{
			logAttrKeySourceFieldname: fromFieldName,
			logAttrKeyTargetFieldname: toFieldName,
		})

		opts := fieldOpts{
			legacy: fromFieldOpts.Legacy(),
		}

		diags.Append(flexer.convert(ctx, sourcePath.AtName(fromFieldName), valFrom.FieldByIndex(fromField.Index), targetPath.AtName(toFieldName), toFieldVal, opts)...)
		if diags.HasError() {
			break
		}
	}

	return diags
}

func expandSourceFields(ctx context.Context, typ reflect.Type, opts AutoFlexOptions) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		for field := range tfreflect.ExportedStructFields(typ) {
			fieldName := field.Name
			if opts.isIgnoredField(fieldName) {
				tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored source field", map[string]any{
					logAttrKeySourceFieldname: fieldName,
				})
				continue
			}

			fromNameOverride, _ := autoflexTags(field)
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

			if !yield(field) {
				return
			}
		}
	}
}

// mapBlockKey takes a struct and extracts the value of the `key`
func mapBlockKey(ctx context.Context, from any) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	valFrom := reflect.ValueOf(from)
	if kind := valFrom.Kind(); kind == reflect.Pointer {
		valFrom = valFrom.Elem()
	}

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(valueType(valFrom)))

	for field := range tfreflect.ExportedStructFields(valFrom.Type()) {
		// go from StringValue to string
		if field.Name == mapBlockKeyFieldName {
			fieldVal := valFrom.FieldByIndex(field.Index)

			if v, ok := fieldVal.Interface().(basetypes.StringValuable); ok {
				v, d := v.ToStringValue(ctx)
				diags.Append(d...)
				if d.HasError() {
					return reflect.Zero(reflect.TypeFor[string]()), diags
				}
				return reflect.ValueOf(v.ValueString()), diags
			}

			// this is not ideal but perhaps better than a panic?
			// return reflect.ValueOf(fmt.Sprintf("%s", valFrom.Field(i))), diags

			return fieldVal, diags
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "Source has no map block key")
	diags.Append(diagExpandingNoMapBlockKey(valFrom.Type()))

	return reflect.Zero(reflect.TypeFor[string]()), diags
}

func expandExpander(ctx context.Context, fromExpander Expander, toVal reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	expanded, d := fromExpander.Expand(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if expanded == nil {
		diags.Append(diagExpandsToNil(reflect.TypeOf(fromExpander)))
		return diags
	}

	expandedVal := reflect.ValueOf(expanded)

	targetType := toVal.Type()
	if targetType.Kind() == reflect.Interface {
		expandedType := reflect.TypeOf(expanded)
		if !expandedType.Implements(targetType) {
			diags.Append(diagExpandedTypeDoesNotImplement(expandedType, targetType))
			return diags
		}

		toVal.Set(expandedVal)

		return diags
	}

	if targetType.Kind() == reflect.Struct {
		expandedVal = expandedVal.Elem()
	}
	expandedType := expandedVal.Type()

	if !expandedType.AssignableTo(targetType) {
		diags.Append(diagCannotBeAssigned(expandedType, targetType))
		return diags
	}

	toVal.Set(expandedVal)

	return diags
}

func expandTypedExpander(ctx context.Context, fromTypedExpander TypedExpander, toVal reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	expanded, d := fromTypedExpander.ExpandTo(ctx, toVal.Type())
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if expanded == nil {
		diags.Append(diagExpandsToNil(reflect.TypeOf(fromTypedExpander)))
		return diags
	}

	expandedVal := reflect.ValueOf(expanded)

	targetType := toVal.Type()
	if targetType.Kind() == reflect.Interface {
		expandedType := reflect.TypeOf(expanded)
		if !expandedType.Implements(targetType) {
			diags.Append(diagExpandedTypeDoesNotImplement(expandedType, targetType))
			return diags
		}

		toVal.Set(expandedVal)

		return diags
	}

	if targetType.Kind() == reflect.Struct {
		expandedVal = expandedVal.Elem()
	}
	expandedType := expandedVal.Type()

	if !expandedType.AssignableTo(targetType) {
		diags.Append(diagCannotBeAssigned(expandedType, targetType))
		return diags
	}

	toVal.Set(expandedVal)

	return diags
}

func diagExpandingSourceIsNil(sourceType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while expanding configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Source of type %q is nil.", fullTypeName(sourceType)),
	)
}

func diagExpandsToNil(expanderType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while expanding configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Expanding %q returned nil.", fullTypeName(expanderType)),
	)
}

func diagExpandedTypeDoesNotImplement(expandedType, targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while expanding configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Type %q does not implement %q.", fullTypeName(expandedType), fullTypeName(targetType)),
	)
}

func diagCannotBeAssigned(expandedType, targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while expanding configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Type %q cannot be assigned to %q.", fullTypeName(expandedType), fullTypeName(targetType)),
	)
}

func diagExpandingSourceDoesNotImplementAttrValue(sourceType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while expanding configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Source type %q does not implement attr.Value", fullTypeName(sourceType)),
	)
}

func diagExpandingNoMapBlockKey(sourceType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while expanding configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Source type %q does not contain field %q", fullTypeName(sourceType), mapBlockKeyFieldName),
	)
}

func diagExpandingIncompatibleTypes(sourceType, targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while expanding configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Source type %q cannot be expanded to target type %q.", fullTypeName(sourceType), fullTypeName(targetType)),
	)
}

// xmlWrapper handles expansion from TF collection types to AWS XML wrapper structs
// that follow the pattern: {Items: []T, Quantity: *int32}
func (expander autoExpander) xmlWrapper(ctx context.Context, vFrom valueWithElementsAs, vTo reflect.Value, wrapperField string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Verify target is a struct with Items and Quantity fields
	if !isXMLWrapperStruct(vTo.Type()) {
		tflog.SubsystemError(ctx, subsystemName, "Target is not a valid XML wrapper struct", map[string]any{
			"target_type": vTo.Type().String(),
		})
		diags.Append(diagExpandingIncompatibleTypes(reflect.TypeOf(vFrom), vTo.Type()))
		return diags
	}

	// Get the Items and Quantity fields
	itemsField := vTo.FieldByName("Items")
	quantityField := vTo.FieldByName("Quantity")

	if !itemsField.IsValid() || !quantityField.IsValid() {
		tflog.SubsystemError(ctx, subsystemName, "XML wrapper struct missing required fields")
		diags.Append(diagExpandingIncompatibleTypes(reflect.TypeOf(vFrom), vTo.Type()))
		return diags
	}

	// Convert the collection elements to a slice
	elements := vFrom.Elements()
	itemsSliceType := itemsField.Type()
	itemsSlice := reflect.MakeSlice(itemsSliceType, len(elements), len(elements))

	// Convert each element
	for i, elem := range elements {
		itemValue := itemsSlice.Index(i)
		if !itemValue.CanSet() {
			continue
		}

		// Handle different element types
		switch elemTyped := elem.(type) {
		case basetypes.StringValuable:
			// Check if target is a pointer to the enum type and source is StringEnum
			if itemsSliceType.Elem().Kind() == reflect.Pointer {
				// Try to extract StringEnum value for pointer slice conversion
				strVal, d := elemTyped.ToStringValue(ctx)
				diags.Append(d...)
				if !diags.HasError() && !strVal.IsNull() {
					enumValue := strVal.ValueString()

					// Create a pointer to the enum type
					enumPtr := reflect.New(itemsSliceType.Elem().Elem())
					enumPtr.Elem().SetString(enumValue)
					itemValue.Set(enumPtr)
				}
			} else if itemsSliceType.Elem().Kind() == reflect.String {
				strVal, d := elemTyped.ToStringValue(ctx)
				diags.Append(d...)
				if !diags.HasError() {
					itemValue.SetString(strVal.ValueString())
				}
			}
		case basetypes.Int64Valuable:
			if elemKind := itemsSliceType.Elem().Kind(); elemKind == reflect.Int32 {
				int64Val, d := elemTyped.ToInt64Value(ctx)
				diags.Append(d...)
				if !diags.HasError() {
					itemValue.SetInt(int64(int32(int64Val.ValueInt64())))
				}
			} else if elemKind == reflect.Int64 {
				int64Val, d := elemTyped.ToInt64Value(ctx)
				diags.Append(d...)
				if !diags.HasError() {
					itemValue.SetInt(int64Val.ValueInt64())
				}
			}
		case basetypes.Int32Valuable:
			if itemsSliceType.Elem().Kind() == reflect.Int32 {
				int32Val, d := elemTyped.ToInt32Value(ctx)
				diags.Append(d...)
				if !diags.HasError() {
					itemValue.SetInt(int64(int32Val.ValueInt32()))
				}
			}
		default:
			// For complex types, try direct assignment if types are compatible
			if elem != nil && !elem.IsNull() && !elem.IsUnknown() {
				if itemValue.Type().AssignableTo(reflect.TypeOf(elem)) {
					itemValue.Set(reflect.ValueOf(elem))
				}
			}
		}
	}

	// Set the Items field
	if itemsField.CanSet() {
		itemsField.Set(itemsSlice)
	}

	// Set the Quantity field
	if quantityField.CanSet() && quantityField.Type().Kind() == reflect.Pointer {
		quantity := int32(len(elements))
		quantityPtr := reflect.New(quantityField.Type().Elem())
		quantityPtr.Elem().Set(reflect.ValueOf(quantity))
		quantityField.Set(quantityPtr)
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Successfully expanded to XML wrapper", map[string]any{
		"source_type":   reflect.TypeOf(vFrom).String(),
		"target_type":   vTo.Type().String(),
		"items_count":   len(elements),
		"wrapper_field": wrapperField,
	})

	return diags
}

// isXMLWrapperStruct detects AWS SDK types that follow the XML wrapper pattern
// with Items slice and Quantity pointer fields
func isXMLWrapperStruct(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	// Check for Items field (slice)
	itemsField, hasItems := t.FieldByName("Items")
	if !hasItems || itemsField.Type.Kind() != reflect.Slice {
		return false
	}

	// Check for Quantity field (pointer to int32)
	quantityField, hasQuantity := t.FieldByName("Quantity")
	if !hasQuantity || quantityField.Type.Kind() != reflect.Pointer {
		return false
	}

	// Quantity should be *int32
	if quantityField.Type.Elem().Kind() != reflect.Int32 {
		return false
	}

	return true
}

// nestedObjectCollectionToXMLWrapper converts a NestedObjectCollectionValue to an XML wrapper struct
// Supports both Rule 1 (Items/Quantity only) and Rule 2 (Items/Quantity + additional fields)
func (expander autoExpander) nestedObjectCollectionToXMLWrapper(ctx context.Context, _ path.Path, vFrom fwtypes.NestedObjectCollectionValue, _ path.Path, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "Expanding NestedObjectCollection to XML wrapper", map[string]any{
		"source_type": vFrom.Type(ctx).String(),
		"target_type": vTo.Type().String(),
	})

	tflog.SubsystemTrace(ctx, subsystemName, "Expanding NestedObjectCollection to XML wrapper", map[string]any{
		"source_type": vFrom.Type(ctx).String(),
		"target_type": vTo.Type().String(),
	})

	// Get the nested Objects as a slice
	from, d := vFrom.ToObjectSlice(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Get reflect value of the slice
	fromSlice := reflect.ValueOf(from)
	if fromSlice.Kind() != reflect.Slice {
		diags.AddError("Invalid source", "ToObjectSlice did not return a slice")
		return diags
	}

	// Check if this is Rule 2 pattern (single nested object with items + additional fields)
	if fromSlice.Len() == 1 {
		nestedObj := fromSlice.Index(0)
		
		// Handle pointer to struct (which is what NestedObjectCollection contains)
		if nestedObj.Kind() == reflect.Ptr && !nestedObj.IsNil() {
			structObj := nestedObj.Elem()
			if structObj.Kind() == reflect.Struct {
				// Check if the struct has an "Items" field - indicates Rule 2
				itemsField := structObj.FieldByName("Items")
				if itemsField.IsValid() {
					return expander.expandRule2XMLWrapper(ctx, nestedObj, vTo)
				}
			}
		}
	}

	// Rule 1: Direct collection to XML wrapper (existing logic)
	return expander.expandRule1XMLWrapper(ctx, fromSlice, vTo)
}

// expandRule2XMLWrapper handles Rule 2: single plural block with items + additional fields
func (expander autoExpander) expandRule2XMLWrapper(ctx context.Context, nestedObjPtr reflect.Value, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "Expanding Rule 2 XML wrapper (items + additional fields)")

	// Get target fields
	itemsField := vTo.FieldByName("Items")
	quantityField := vTo.FieldByName("Quantity")
	
	if !itemsField.IsValid() || !quantityField.IsValid() {
		diags.Append(diagExpandingIncompatibleTypes(nestedObjPtr.Type(), vTo.Type()))
		return diags
	}

	// Get the struct from the pointer
	nestedObj := nestedObjPtr.Elem()
	
	// Extract "Items" field from nested object
	itemsSourceField := nestedObj.FieldByName("Items")
	if !itemsSourceField.IsValid() {
		diags.AddError("Missing items field", "Rule 2 XML wrapper requires 'Items' field")
		return diags
	}

	// Convert items collection to Items slice using existing logic
	if itemsAttr, ok := itemsSourceField.Interface().(attr.Value); ok {
		if collectionValue, ok := itemsAttr.(valueWithElementsAs); ok {
			diags.Append(expander.convertCollectionToXMLWrapperFields(ctx, collectionValue, itemsField, quantityField)...)
			if diags.HasError() {
				return diags
			}
		}
	}

	// Handle additional fields (e.g., Enabled)
	for i := 0; i < vTo.NumField(); i++ {
		targetField := vTo.Field(i)
		targetFieldType := vTo.Type().Field(i)
		fieldName := targetFieldType.Name
		
		// Skip Items and Quantity (already handled)
		if fieldName == "Items" || fieldName == "Quantity" {
			continue
		}
		
		// Look for matching field in nested object
		sourceField := nestedObj.FieldByName(fieldName)
		if sourceField.IsValid() {
			// Convert TF field to AWS field
			if tfAttr, ok := sourceField.Interface().(attr.Value); ok {
				diags.Append(autoExpandConvert(ctx, tfAttr, targetField.Addr().Interface(), expander)...)
			}
		}
	}

	return diags
}

// expandRule1XMLWrapper handles Rule 1: direct collection to XML wrapper (existing logic)
func (expander autoExpander) expandRule1XMLWrapper(ctx context.Context, fromSlice reflect.Value, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "Expanding Rule 1 XML wrapper (direct collection)")

	// Get the Items and Quantity fields from target struct
	itemsField := vTo.FieldByName("Items")
	quantityField := vTo.FieldByName("Quantity")

	if !itemsField.IsValid() || !quantityField.IsValid() {
		diags.Append(diagExpandingIncompatibleTypes(fromSlice.Type(), vTo.Type()))
		return diags
	}

	// Create the Items slice
	itemsSliceType := itemsField.Type()
	itemsCount := fromSlice.Len()
	itemsSlice := reflect.MakeSlice(itemsSliceType, itemsCount, itemsCount)

	tflog.SubsystemTrace(ctx, subsystemName, "Converting nested objects to items", map[string]any{
		"items_count": itemsCount,
		"items_type":  itemsSliceType.String(),
	})

	// Convert each nested object
	for i := range itemsCount {
		sourceItem := fromSlice.Index(i)
		targetItem := itemsSlice.Index(i)

		// Create new instance for the target item
		targetItemType := itemsSliceType.Elem()
		if targetItemType.Kind() == reflect.Pointer {
			// For []*struct
			newItem := reflect.New(targetItemType.Elem())
			diags.Append(autoExpandConvert(ctx, sourceItem.Interface(), newItem.Interface(), expander)...)
			if diags.HasError() {
				return diags
			}
			targetItem.Set(newItem)
		} else {
			// For []struct - need to set the value directly
			newItem := reflect.New(targetItemType)
			diags.Append(autoExpandConvert(ctx, sourceItem.Interface(), newItem.Interface(), expander)...)
			if diags.HasError() {
				return diags
			}
			targetItem.Set(newItem.Elem())
		}
	}

	// Set the Items field
	if itemsField.CanSet() {
		itemsField.Set(itemsSlice)
	}

	// Set the Quantity field
	if quantityField.CanSet() && quantityField.Type().Kind() == reflect.Pointer {
		quantity := int32(itemsCount)
		quantityPtr := reflect.New(quantityField.Type().Elem())
		quantityPtr.Elem().Set(reflect.ValueOf(quantity))
		quantityField.Set(quantityPtr)
	}

	return diags
}

// handleXMLWrapperCollapse handles the special case where multiple source fields
// need to be combined into a single complex target field containing XML wrapper structures.
// This handles patterns like:
//   - Source: separate XMLWrappedEnumSlice and Other fields
//   - Target: single XMLWrappedEnumSlice field containing XMLWrappedEnumSliceOther struct
//     with Items/Quantity from main field and Other nested XML wrapper from other field
func (expander autoExpander) handleXMLWrapperCollapse(ctx context.Context, sourcePath path.Path, valFrom reflect.Value, targetPath path.Path, valTo reflect.Value, typeFrom, typeTo reflect.Type, processedFields map[string]bool) diag.Diagnostics {
	var diags diag.Diagnostics

	// Look for target fields that are complex XML wrapper structures
	for i := 0; i < typeTo.NumField(); i++ {
		toField := typeTo.Field(i)
		toFieldName := toField.Name
		toFieldType := toField.Type
		toFieldVal := valTo.Field(i)

		// Check if this is a pointer to a struct or direct struct
		var targetStructType reflect.Type
		if toFieldType.Kind() == reflect.Pointer && toFieldType.Elem().Kind() == reflect.Struct {
			targetStructType = toFieldType.Elem()
		} else if toFieldType.Kind() == reflect.Struct {
			targetStructType = toFieldType
		} else {
			continue
		}

		// Check if this target struct has the XML wrapper collapse pattern:
		// - Contains Items/Quantity fields (making it an XML wrapper)
		// - Contains additional fields that should come from other source fields
		if !expander.isXMLWrapperCollapseTarget(targetStructType) {
			continue
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Found XML wrapper collapse target", map[string]any{
			logAttrKeyTargetFieldname: toFieldName,
			logAttrKeyTargetType:      targetStructType.String(),
		})

		// Handle XML wrapper collapse patterns generically
		diags.Append(expander.buildGenericXMLWrapperCollapse(ctx, sourcePath, valFrom, targetPath.AtName(toFieldName), toFieldVal, typeFrom, targetStructType, toFieldType.Kind() == reflect.Pointer, processedFields)...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

// isXMLWrapperCollapseTarget checks if a struct type represents a target that should be
// populated via XML wrapper collapse (multiple source fields -> single complex target)
func (expander autoExpander) isXMLWrapperCollapseTarget(structType reflect.Type) bool {
	hasItems := false
	hasQuantity := false
	hasOtherFields := false

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := field.Name

		switch fieldName {
		case "Items":
			hasItems = true
		case "Quantity":
			hasQuantity = true
		default:
			// Any other field suggests this is a complex collapse target
			hasOtherFields = true
		}
	}

	// Must have Items/Quantity (XML wrapper pattern) plus other fields
	return hasItems && hasQuantity && hasOtherFields
}

// convertCollectionToItemsQuantity converts a source collection to Items slice and Quantity fields
func (expander autoExpander) convertCollectionToItemsQuantity(ctx context.Context, sourcePath path.Path, sourceFieldVal reflect.Value, targetPath path.Path, itemsField, quantityField reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	sourceValue, ok := sourceFieldVal.Interface().(attr.Value)
	if !ok {
		tflog.SubsystemError(ctx, subsystemName, "Source field is not an attr.Value")
		return diags
	}

	// Convert based on source type
	switch vFrom := sourceValue.(type) {
	case basetypes.SetValuable, basetypes.ListValuable:
		if setValue, ok := vFrom.(valueWithElementsAs); ok {
			// Use existing logic to convert to Items/Quantity, but target specific fields
			diags.Append(expander.convertCollectionToXMLWrapperFields(ctx, setValue, itemsField, quantityField)...)
		}
	default:
		tflog.SubsystemError(ctx, subsystemName, "Unsupported source type for Items/Quantity conversion", map[string]any{
			"source_type": fmt.Sprintf("%T", vFrom),
		})
	}

	return diags
}

// convertCollectionToXMLWrapperFields converts a collection directly to Items and Quantity fields
func (expander autoExpander) convertCollectionToXMLWrapperFields(ctx context.Context, vFrom valueWithElementsAs, itemsField, quantityField reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Get the source elements
	elements := vFrom.Elements()
	itemsCount := len(elements)

	// Create the Items slice
	if itemsField.CanSet() {
		itemsType := itemsField.Type()
		itemsSlice := reflect.MakeSlice(itemsType, itemsCount, itemsCount)

		// Convert each element - use the same logic as xmlWrapper
		for i, elem := range elements {
			itemValue := itemsSlice.Index(i)
			if !itemValue.CanSet() {
				continue
			}

			// Handle different element types
			switch elemTyped := elem.(type) {
			case basetypes.StringValuable:
				// Check if target is a pointer to the enum type and source is StringEnum
				if itemsType.Elem().Kind() == reflect.Pointer {
					// Try to extract StringEnum value for pointer slice conversion
					strVal, d := elemTyped.ToStringValue(ctx)
					diags.Append(d...)
					if !diags.HasError() && !strVal.IsNull() {
						enumValue := strVal.ValueString()

						// Create a pointer to the enum type
						enumPtr := reflect.New(itemsType.Elem().Elem())
						enumPtr.Elem().SetString(enumValue)
						itemValue.Set(enumPtr)
					}
				} else if itemsType.Elem().Kind() == reflect.String {
					strVal, d := elemTyped.ToStringValue(ctx)
					diags.Append(d...)
					if !diags.HasError() {
						itemValue.SetString(strVal.ValueString())
					}
				}
			case basetypes.Int64Valuable:
				if elemKind := itemsType.Elem().Kind(); elemKind == reflect.Int32 {
					int64Val, d := elemTyped.ToInt64Value(ctx)
					diags.Append(d...)
					if !diags.HasError() {
						itemValue.SetInt(int64(int32(int64Val.ValueInt64())))
					}
				} else if elemKind == reflect.Int64 {
					int64Val, d := elemTyped.ToInt64Value(ctx)
					diags.Append(d...)
					if !diags.HasError() {
						itemValue.SetInt(int64Val.ValueInt64())
					}
				}
			case basetypes.Int32Valuable:
				if itemsType.Elem().Kind() == reflect.Int32 {
					int32Val, d := elemTyped.ToInt32Value(ctx)
					diags.Append(d...)
					if !diags.HasError() {
						itemValue.SetInt(int64(int32Val.ValueInt32()))
					}
				}
			default:
				// For complex types, try direct assignment if types are compatible
				if elem != nil && !elem.IsNull() && !elem.IsUnknown() {
					if itemValue.Type().AssignableTo(reflect.TypeOf(elem)) {
						itemValue.Set(reflect.ValueOf(elem))
					}
				}
			}
		}

		itemsField.Set(itemsSlice)
	}

	// Set the Quantity field
	if quantityField.CanSet() && quantityField.Type().Kind() == reflect.Pointer {
		quantity := int32(itemsCount)
		quantityPtr := reflect.New(quantityField.Type().Elem())
		quantityPtr.Elem().Set(reflect.ValueOf(quantity))
		quantityField.Set(quantityPtr)
	}

	return diags
}

// shouldConvertToXMLWrapper determines if a source field should be converted to XML wrapper format
func (expander autoExpander) shouldConvertToXMLWrapper(sourceFieldVal, targetFieldVal reflect.Value) bool {
	// Check if source is a collection (Set/List) and target is XML wrapper struct
	sourceValue, ok := sourceFieldVal.Interface().(attr.Value)
	if !ok {
		return false
	}

	// Source should be a collection
	switch sourceValue.(type) {
	case basetypes.SetValuable, basetypes.ListValuable:
		// Target should be a pointer to struct or struct with XML wrapper pattern
		targetType := targetFieldVal.Type()
		if targetType.Kind() == reflect.Pointer && targetType.Elem().Kind() == reflect.Struct {
			return isXMLWrapperStruct(targetType.Elem())
		}
		if targetType.Kind() == reflect.Struct {
			return isXMLWrapperStruct(targetType)
		}
	}

	return false
}

// convertToXMLWrapper converts a source collection to an XML wrapper structure
func (expander autoExpander) convertToXMLWrapper(ctx context.Context, sourcePath path.Path, sourceFieldVal reflect.Value, targetPath path.Path, targetFieldVal reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	sourceValue, ok := sourceFieldVal.Interface().(attr.Value)
	if !ok {
		tflog.SubsystemError(ctx, subsystemName, "Source field is not an attr.Value", map[string]any{
			"source_type": sourceFieldVal.Type().String(),
		})
		return diags
	}

	// Handle conversion based on source type
	switch vFrom := sourceValue.(type) {
	case basetypes.SetValuable:
		if setValue, ok := vFrom.(valueWithElementsAs); ok {
			diags.Append(expander.listOrSetOfString(ctx, setValue, targetFieldVal)...)
		}
	case basetypes.ListValuable:
		if listValue, ok := vFrom.(valueWithElementsAs); ok {
			diags.Append(expander.listOrSetOfString(ctx, listValue, targetFieldVal)...)
		}
	default:
		tflog.SubsystemError(ctx, subsystemName, "Unsupported source type for XML wrapper conversion", map[string]any{
			"source_type": fmt.Sprintf("%T", vFrom),
		})
	}

	return diags
}

// buildGenericXMLWrapperCollapse handles any XML wrapper collapse pattern generically
func (expander autoExpander) buildGenericXMLWrapperCollapse(ctx context.Context, sourcePath path.Path, valFrom reflect.Value, targetPath path.Path, toFieldVal reflect.Value, typeFrom, targetStructType reflect.Type, isPointer bool, processedFields map[string]bool) diag.Diagnostics {
	var diags diag.Diagnostics

	// Create the target struct
	targetStruct := reflect.New(targetStructType)
	targetStructVal := targetStruct.Elem()

	tflog.SubsystemTrace(ctx, subsystemName, "Building generic XML wrapper collapse struct", map[string]any{
		"target_type": targetStructType.String(),
	})

	// Check for null handling - if all source collection fields are null and target is pointer, set to nil
	if isPointer {
		allFieldsNull := true
		hasCollectionFields := false

		for i := 0; i < typeFrom.NumField(); i++ {
			sourceField := typeFrom.Field(i)
			sourceFieldVal := valFrom.FieldByName(sourceField.Name)

			if sourceValue, ok := sourceFieldVal.Interface().(attr.Value); ok {
				switch sourceValue.(type) {
				case basetypes.SetValuable, basetypes.ListValuable:
					hasCollectionFields = true
					if !sourceValue.IsNull() {
						allFieldsNull = false
						break
					}
				}
			}
		}

		if hasCollectionFields && allFieldsNull {
			// All collection fields are null and target is pointer - set to nil
			toFieldVal.SetZero()

			// Mark all collection fields as processed
			for i := 0; i < typeFrom.NumField(); i++ {
				sourceField := typeFrom.Field(i)
				sourceFieldVal := valFrom.FieldByName(sourceField.Name)
				if sourceValue, ok := sourceFieldVal.Interface().(attr.Value); ok {
					switch sourceValue.(type) {
					case basetypes.SetValuable, basetypes.ListValuable:
						processedFields[sourceField.Name] = true
					}
				}
			}

			tflog.SubsystemTrace(ctx, subsystemName, "All source collection fields null - setting target to nil")
			return diags
		}
	}

	// First, identify which source field should populate Items/Quantity
	// This is typically the field that matches the target field name or the "main" collection
	var mainSourceFieldName string

	// Try to find a source field that matches the target field name
	targetFieldName := targetPath.String()
	if lastDot := strings.LastIndex(targetFieldName, "."); lastDot >= 0 {
		targetFieldName = targetFieldName[lastDot+1:]
	}

	if _, found := typeFrom.FieldByName(targetFieldName); found {
		mainSourceFieldName = targetFieldName
	}
	// Remove the fallback logic that incorrectly picks any collection field
	// This was causing the wrong field data to be used for XML wrappers

	tflog.SubsystemTrace(ctx, subsystemName, "Identified main source field", map[string]any{
		"main_source_field": mainSourceFieldName,
	})

	// Process Items and Quantity from the main source field
	hasMainSourceField := false
	if mainSourceFieldName != "" {
		sourceFieldVal := valFrom.FieldByName(mainSourceFieldName)

		// Get the Items and Quantity fields in the target
		itemsField := targetStructVal.FieldByName("Items")
		quantityField := targetStructVal.FieldByName("Quantity")

		if itemsField.IsValid() && quantityField.IsValid() {
			// Check if the source field is actually usable (not null/unknown)
			if sourceValue, ok := sourceFieldVal.Interface().(attr.Value); ok {
				if !sourceValue.IsNull() && !sourceValue.IsUnknown() {
					// Convert the collection to Items slice and Quantity
					diags.Append(expander.convertCollectionToItemsQuantity(ctx, sourcePath.AtName(mainSourceFieldName), sourceFieldVal, targetPath.AtName("Items"), itemsField, quantityField)...)
					if diags.HasError() {
						return diags
					}
					hasMainSourceField = true
				} else {
					tflog.SubsystemDebug(ctx, subsystemName, "Main source field is null or unknown - skipping XML wrapper creation", map[string]any{
						"source_field": mainSourceFieldName,
						"is_null":      sourceValue.IsNull(),
						"is_unknown":   sourceValue.IsUnknown(),
					})
				}
			}
		}

		// Mark main source field as processed
		processedFields[mainSourceFieldName] = true
	}

	// Track if we found any fields to populate
	hasAnySourceFields := hasMainSourceField

	// Now process each remaining field in the target struct
	for i := 0; i < targetStructType.NumField(); i++ {
		targetField := targetStructType.Field(i)
		targetFieldName := targetField.Name
		targetFieldVal := targetStructVal.Field(i)

		// Skip Items and Quantity as they were handled above
		if targetFieldName == "Items" || targetFieldName == "Quantity" {
			continue
		}

		if !targetFieldVal.CanSet() {
			continue
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Processing additional target field", map[string]any{
			"target_field": targetFieldName,
			"target_type":  targetField.Type.String(),
		})

		// Look for a source field with the same name
		if _, found := typeFrom.FieldByName(targetFieldName); found {
			sourceFieldVal := valFrom.FieldByName(targetFieldName)

			tflog.SubsystemTrace(ctx, subsystemName, "Found matching source field", map[string]any{
				"source_field": targetFieldName,
				"target_field": targetFieldName,
			})

			// Check if we need special XML wrapper conversion
			if expander.shouldConvertToXMLWrapper(sourceFieldVal, targetFieldVal) {
				// Convert collection to XML wrapper structure
				diags.Append(expander.convertToXMLWrapper(ctx, sourcePath.AtName(targetFieldName), sourceFieldVal, targetPath.AtName(targetFieldName), targetFieldVal)...)
				if diags.HasError() {
					return diags
				}
			} else {
				// Regular field conversion
				opts := fieldOpts{}
				diags.Append(expander.convert(ctx, sourcePath.AtName(targetFieldName), sourceFieldVal, targetPath.AtName(targetFieldName), targetFieldVal, opts)...)
				if diags.HasError() {
					return diags
				}
			}

			// Mark source field as processed and track that we found fields
			processedFields[targetFieldName] = true
			hasAnySourceFields = true
		} else {
			tflog.SubsystemDebug(ctx, subsystemName, "No source field found for target field", map[string]any{
				"target_field": targetFieldName,
			})
		}
	}

	// Only set the constructed struct if we found source fields to populate it
	if hasAnySourceFields {
		// Set the constructed struct into the target field
		if isPointer {
			toFieldVal.Set(targetStruct)
		} else {
			toFieldVal.Set(targetStruct.Elem())
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Successfully built generic XML wrapper collapse struct")
	} else {
		tflog.SubsystemDebug(ctx, subsystemName, "No source fields found for XML wrapper collapse target - leaving as nil/zero value")
		// Leave the field as nil/zero value (don't set anything)
	}

	return diags
}
