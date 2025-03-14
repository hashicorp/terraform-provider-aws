// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"iter"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
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
		if s, ok := vFrom.(fwtypes.SmithyJSON[smithyjson.JSONStringer]); ok {
			v, d := s.ValueInterface()
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
		"from map[string, %s]": v.ElementType(ctx),
		"to":                   vTo.Kind(),
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
		"from map[string, %s]": vFrom.ElementType(ctx),
		"to":                   vTo.Kind(),
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
		sourcePath := sourcePath.AtListIndex(0)
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		diags.Append(expander.nestedObjectToStruct(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
		return diags

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
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

	for fromField := range expandSourceFields(ctx, typeFrom, flexer.getOptions()) {
		fromFieldName := fromField.Name
		_, fromFieldOpts := autoflexTags(fromField)

		toField, ok := findFieldFuzzy(ctx, fromFieldName, typeFrom, typeTo, flexer)
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
