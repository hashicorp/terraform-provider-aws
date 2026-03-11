// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfreflect "github.com/hashicorp/terraform-provider-aws/internal/reflect"
)

var (
	// xmlWrapperCache caches XML wrapper struct detection results
	xmlWrapperCache      = make(map[reflect.Type]bool)
	xmlWrapperCacheMutex sync.RWMutex
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
	Options    AutoFlexOptions
	fieldCache map[reflect.Type]map[string]reflect.StructField
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
		Options:    o,
		fieldCache: make(map[reflect.Type]map[string]reflect.StructField),
	}
}

func (expander autoExpander) getOptions() AutoFlexOptions {
	return expander.Options
}

// getCachedField returns a cached field lookup or performs and caches the lookup
func (expander *autoExpander) getCachedField(t reflect.Type, fieldName string) (reflect.StructField, bool) {
	if expander.fieldCache == nil {
		expander.fieldCache = make(map[reflect.Type]map[string]reflect.StructField)
	}

	if typeCache, exists := expander.fieldCache[t]; exists {
		if field, found := typeCache[fieldName]; found {
			return field, true
		}
	} else {
		expander.fieldCache[t] = make(map[string]reflect.StructField)
	}

	// Perform lookup and cache result
	if field, found := t.FieldByName(fieldName); found {
		expander.fieldCache[t][fieldName] = field
		return field, true
	}

	return reflect.StructField{}, false
}

// getCachedFieldValue returns a cached field value lookup
func (expander *autoExpander) getCachedFieldValue(v reflect.Value, fieldName string) reflect.Value {
	if field, found := expander.getCachedField(v.Type(), fieldName); found {
		return v.FieldByIndex(field.Index)
	}
	return reflect.Value{}
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
		// Special case: if target is XML wrapper struct and no omitempty, create zero-value
		if vTo.Kind() == reflect.Ptr && !fieldOpts.omitempty {
			targetType := vTo.Type().Elem()
			if targetType.Kind() == reflect.Struct && potentialXMLWrapperStruct(targetType) {
				tflog.SubsystemDebug(ctx, subsystemName, "Source is null but target is XML wrapper without omitempty - creating zero-value struct")
				zeroStruct := reflect.New(targetType)
				// Initialize Items slice and Quantity
				wrapperFieldName := getXMLWrapperSliceFieldName(targetType)
				itemsField := zeroStruct.Elem().FieldByName(wrapperFieldName)
				if itemsField.IsValid() && itemsField.Kind() == reflect.Slice {
					itemsField.Set(reflect.MakeSlice(itemsField.Type(), 0, 0))
				}
				quantityField := zeroStruct.Elem().FieldByName(xmlWrapperFieldQuantity)
				if quantityField.IsValid() && quantityField.Kind() == reflect.Ptr {
					zero := int32(0)
					quantityField.Set(reflect.ValueOf(&zero))
				}
				// Set zero values for other pointer fields
				for i := 0; i < zeroStruct.Elem().NumField(); i++ {
					field := zeroStruct.Elem().Field(i)
					fieldType := targetType.Field(i)
					if fieldType.Name == wrapperFieldName || fieldType.Name == xmlWrapperFieldQuantity {
						continue
					}
					if field.Kind() == reflect.Ptr && field.CanSet() && field.IsNil() {
						switch fieldType.Type.Elem().Kind() {
						case reflect.Bool:
							falseVal := false
							field.Set(reflect.ValueOf(&falseVal))
						case reflect.Int32:
							zeroVal := int32(0)
							field.Set(reflect.ValueOf(&zeroVal))
						}
					}
				}
				vTo.Set(zeroStruct)
				return diags
			}
		}
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
		diags.Append(expander.object(ctx, sourcePath, vFrom, targetPath, vTo, fieldOpts)...)
		return diags

	case basetypes.ListValuable:
		diags.Append(expander.list(ctx, sourcePath, vFrom, targetPath, vTo, fieldOpts)...)
		return diags

	case basetypes.MapValuable:
		diags.Append(expander.map_(ctx, vFrom, vTo, fieldOpts)...)
		return diags

	case basetypes.SetValuable:
		diags.Append(expander.set(ctx, sourcePath, vFrom, targetPath, vTo, fieldOpts)...)
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

	tflog.SubsystemError(ctx, subsystemName, "Expanding incompatible types")
	// TODO: Should continue failing silently for now
	// diags.Append(diagExpandingIncompatibleTypes(reflect.TypeOf(vFrom), vTo.Type()))

	return diags
}

// string copies a Plugin Framework Object(ish) value to a compatible AWS API value.
func (expander autoExpander) object(ctx context.Context, sourcePath path.Path, vFrom basetypes.ObjectValuable, targetPath path.Path, vTo reflect.Value, _ fieldOpts) diag.Diagnostics {
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
func (expander autoExpander) list(ctx context.Context, sourcePath path.Path, vFrom basetypes.ListValuable, targetPath path.Path, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToListValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.Int64Typable:
		diags.Append(expander.listOrSetOfInt64(ctx, v, vTo, fieldOpts)...)
		return diags

	case basetypes.StringTypable:
		diags.Append(expander.listOrSetOfString(ctx, v, vTo, fieldOpts)...)
		return diags

	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.NestedObjectCollectionValue); ok {
			diags.Append(expander.nestedObjectCollection(ctx, sourcePath, vFrom, targetPath, vTo, fieldOpts)...)
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
func (expander autoExpander) listOrSetOfInt64(ctx context.Context, vFrom valueWithElementsAs, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Struct:
		// Check if target is an XML wrapper struct
		if fieldOpts.xmlWrapper {
			diags.Append(expander.xmlWrapper(ctx, vFrom, vTo, fieldOpts.xmlWrapperField)...)
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

	case reflect.Pointer:
		switch tElem := vTo.Type().Elem(); tElem.Kind() {
		case reflect.Struct:
			// Check if target is a pointer to an XML wrapper struct
			if fieldOpts.xmlWrapper {
				// Create new instance of the XML wrapper struct
				newStruct := reflect.New(tElem).Elem()
				diags.Append(expander.xmlWrapper(ctx, vFrom, newStruct, fieldOpts.xmlWrapperField)...)
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

// listOrSetOfString copies a Plugin Framework ListOfString(ish) or SetOfString(ish) value to a compatible AWS API value.
func (expander autoExpander) listOrSetOfString(ctx context.Context, vFrom valueWithElementsAs, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Struct:
		// Check if target is an XML wrapper struct
		if fieldOpts.xmlWrapper {
			diags.Append(expander.xmlWrapper(ctx, vFrom, vTo, fieldOpts.xmlWrapperField)...)
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
			if fieldOpts.xmlWrapper {
				// Create new instance of the XML wrapper struct
				newStruct := reflect.New(tElem).Elem()
				diags.Append(expander.xmlWrapper(ctx, vFrom, newStruct, fieldOpts.xmlWrapperField)...)
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

// listOrSetOfInt32 copies a Plugin Framework ListOfInt32(ish) or SetOfInt32(ish) value to a compatible AWS API value.
func (expander autoExpander) listOrSetOfInt32(ctx context.Context, vFrom valueWithElementsAs, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vTo.Kind() {
	case reflect.Struct:
		// Check if target is an XML wrapper struct
		if fieldOpts.xmlWrapper {
			diags.Append(expander.xmlWrapper(ctx, vFrom, vTo, fieldOpts.xmlWrapperField)...)
			return diags
		}

	case reflect.Slice:
		switch tSliceElem := vTo.Type().Elem(); tSliceElem.Kind() {
		case reflect.Int32:
			//
			// types.Set(OfInt32) -> []int32
			//
			tflog.SubsystemTrace(ctx, subsystemName, "Expanding with ElementsAs", map[string]any{
				logAttrKeySourceSize: len(vFrom.Elements()),
			})
			var to []int32
			diags.Append(vFrom.ElementsAs(ctx, &to, false)...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags

		case reflect.Pointer:
			switch tSliceElem.Elem().Kind() {
			case reflect.Int32:
				//
				// types.Set(OfInt32) -> []*int32
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
			}
		}

	case reflect.Pointer:
		switch tElem := vTo.Type().Elem(); tElem.Kind() {
		case reflect.Struct:
			// Check if target is a pointer to an XML wrapper struct
			if fieldOpts.xmlWrapper {
				// Create new instance of the XML wrapper struct
				newStruct := reflect.New(tElem).Elem()
				diags.Append(expander.xmlWrapper(ctx, vFrom, newStruct, fieldOpts.xmlWrapperField)...)
				if !diags.HasError() {
					vTo.Set(newStruct.Addr())
				}
				return diags
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Expand; incompatible types", map[string]any{
		"from": "Set[Int32]",
		"to":   vTo.Kind(),
	})

	return diags
}

// map_ copies a Plugin Framework Map(ish) value to a compatible AWS API value.
func (expander autoExpander) map_(ctx context.Context, vFrom basetypes.MapValuable, vTo reflect.Value, _ fieldOpts) diag.Diagnostics {
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
func (expander autoExpander) set(ctx context.Context, sourcePath path.Path, vFrom basetypes.SetValuable, targetPath path.Path, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	v, d := vFrom.ToSetValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch v.ElementType(ctx).(type) {
	case basetypes.Int64Typable:
		diags.Append(expander.listOrSetOfInt64(ctx, v, vTo, fieldOpts)...)
		return diags

	case basetypes.Int32Typable:
		diags.Append(expander.listOrSetOfInt32(ctx, v, vTo, fieldOpts)...)
		return diags

	case basetypes.StringTypable:
		diags.Append(expander.listOrSetOfString(ctx, v, vTo, fieldOpts)...)
		return diags

	case basetypes.ObjectTypable:
		if vFrom, ok := vFrom.(fwtypes.NestedObjectCollectionValue); ok {
			diags.Append(expander.nestedObjectCollection(ctx, sourcePath, vFrom, targetPath, vTo, fieldOpts)...)
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
func (expander autoExpander) nestedObjectCollection(ctx context.Context, sourcePath path.Path, vFrom fwtypes.NestedObjectCollectionValue, targetPath path.Path, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	// TRACE: Log entry with field options
	tflog.SubsystemTrace(ctx, subsystemName, "TRACE: nestedObjectCollection entry", map[string]any{
		"xmlWrapper":      fieldOpts.xmlWrapper,
		"xmlWrapperField": fieldOpts.xmlWrapperField,
	})

	switch tTo := vTo.Type(); vTo.Kind() {
	case reflect.Struct:
		// Check if xmlwrapper tag is present
		if fieldOpts.xmlWrapper {
			diags.Append(expander.nestedObjectCollectionToXMLWrapper(ctx, sourcePath, vFrom, targetPath, vTo, fieldOpts.xmlWrapperField)...)
			return diags
		}

		sourcePath := sourcePath.AtListIndex(0)
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		diags.Append(expander.nestedObjectToStruct(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
		return diags

	case reflect.Pointer:
		switch tElem := tTo.Elem(); tElem.Kind() {
		case reflect.Struct:
			// Check if xmlwrapper tag is present
			if fieldOpts.xmlWrapper {
				// Create new instance of the XML wrapper struct
				newWrapper := reflect.New(tElem)
				diags.Append(expander.nestedObjectCollectionToXMLWrapper(ctx, sourcePath, vFrom, targetPath, newWrapper.Elem(), fieldOpts.xmlWrapperField)...)
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

		// Rule 2 XML wrapper: After expanding nested object, check if any field had xmlwrapper tag
		// and set Quantity field if target is XML wrapper struct
		fromVal := reflect.ValueOf(from).Elem()
		for i := 0; i < fromVal.NumField(); i++ {
			field := fromVal.Type().Field(i)
			_, fieldOpts := autoflexTags(field)
			if wrapperField := fieldOpts.XMLWrapperField(); wrapperField != "" {
				// Found xmlwrapper tag - set Quantity in target
				setXMLWrapperQuantity(to, wrapperField)
				break
			}
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
		tflog.SubsystemError(ctx, subsystemName, "Expanding to incompatible interface")
		// TODO: Should continue failing silently for now
		// diags.Append(diagExpandingIncompatibleTypes(reflect.TypeOf(vFrom), vTo.Type()))
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
		if fromFieldOpts.NoExpand() {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping noexpand source field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
			})
			continue
		}

		// TRACE: Log XML wrapper tag detection
		if xmlWrapperField := fromFieldOpts.XMLWrapperField(); xmlWrapperField != "" {
			tflog.SubsystemTrace(ctx, subsystemName, "TRACE: Found xmlwrapper tag", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
				"xmlWrapperField":         xmlWrapperField,
			})
		}

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
			tflog.SubsystemDebug(ctx, subsystemName, "No corresponding target field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
			})
			continue
		}
		toFieldName := toField.Name
		toFieldVal := valTo.FieldByIndex(toField.Index)
		if !toFieldVal.CanSet() {
			// Corresponding field value can't be changed.
			tflog.SubsystemDebug(ctx, subsystemName, "Target field cannot be set", map[string]any{
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
			legacy:          fromFieldOpts.Legacy(),
			omitempty:       fromFieldOpts.OmitEmpty(),
			xmlWrapper:      fromFieldOpts.XMLWrapperField() != "",
			xmlWrapperField: fromFieldOpts.XMLWrapperField(),
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

// XML wrapper status:
// - frankly this is a mess
// - there are patterns and bolt-ons all over the place
// - we need to eventually refactor this into a coherent whole
// - for now, we just try to contain the complexity from here down

const (
	xmlWrapperFieldQuantity = "Quantity"
)

// xmlWrapper handles expansion from TF collection types to AWS XML wrapper structs
// that follow the pattern: {Items: []T, Quantity: *int32}
func (expander *autoExpander) xmlWrapper(ctx context.Context, vFrom valueWithElementsAs, vTo reflect.Value, wrapperField string) diag.Diagnostics {
	var diags diag.Diagnostics

	// Validate target structure
	itemsField, quantityField, d := expander.validateXMLWrapperTarget(vTo, wrapperField)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Convert elements to slice
	itemsSlice, d := expander.convertElementsToXMLWrapperSlice(ctx, vFrom.Elements(), itemsField.Type())
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Set fields
	itemsField.Set(itemsSlice)
	if err := setXMLWrapperQuantityField(quantityField, int32(len(vFrom.Elements()))); err != nil {
		diags.Append(diagExpandingIncompatibleTypes(reflect.TypeOf(vFrom), vTo.Type()))
		return diags
	}

	return diags
}

// validateXMLWrapperTarget validates target structure and returns field references
func (expander *autoExpander) validateXMLWrapperTarget(vTo reflect.Value, wrapperField string) (itemsField, quantityField reflect.Value, diags diag.Diagnostics) {
	itemsField = expander.getCachedFieldValue(vTo, wrapperField)
	quantityField = expander.getCachedFieldValue(vTo, xmlWrapperFieldQuantity)

	if !itemsField.IsValid() || !quantityField.IsValid() {
		diags.Append(diagExpandingIncompatibleTypes(reflect.TypeOf(vTo.Interface()), vTo.Type()))
	}
	return
}

// convertElementsToXMLWrapperSlice converts collection elements to target slice
func (expander *autoExpander) convertElementsToXMLWrapperSlice(ctx context.Context, elements []attr.Value, sliceType reflect.Type) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	slice := reflect.MakeSlice(sliceType, len(elements), len(elements))
	targetElemType := sliceType.Elem()

	for i, elem := range elements {
		itemValue := slice.Index(i)
		if !itemValue.CanSet() {
			continue
		}

		diags.Append(expander.convertElementToXMLWrapperItem(ctx, elem, itemValue, targetElemType)...)
		if diags.HasError() {
			return reflect.Value{}, diags
		}
	}

	return slice, diags
}

// convertElementToXMLWrapperItem converts single element to target item
func (expander *autoExpander) convertElementToXMLWrapperItem(ctx context.Context, elem attr.Value, itemValue reflect.Value, targetElemType reflect.Type) diag.Diagnostics {
	switch elemTyped := elem.(type) {
	case basetypes.StringValuable:
		return convertStringValueableToXMLItem(ctx, elemTyped, itemValue, targetElemType)
	case basetypes.Int64Valuable:
		return convertInt64ValueableToXMLItem(ctx, elemTyped, itemValue, targetElemType)
	case basetypes.Int32Valuable:
		return convertInt32ValueableToXMLItem(ctx, elemTyped, itemValue, targetElemType)
	default:
		return convertComplexValueToXMLItem(elem, itemValue)
	}
}

// convertStringValueableToXMLItem converts string values to XML wrapper items
func convertStringValueableToXMLItem(ctx context.Context, elem basetypes.StringValuable, itemValue reflect.Value, targetElemType reflect.Type) diag.Diagnostics {
	var diags diag.Diagnostics

	strVal, d := elem.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() || strVal.IsNull() {
		return diags
	}

	switch targetElemType.Kind() {
	case reflect.Pointer:
		enumPtr := reflect.New(targetElemType.Elem())
		enumPtr.Elem().SetString(strVal.ValueString())
		itemValue.Set(enumPtr)
	case reflect.String:
		itemValue.SetString(strVal.ValueString())
	}
	return diags
}

// convertInt64ValueableToXMLItem converts int64 values to XML wrapper items
func convertInt64ValueableToXMLItem(ctx context.Context, elem basetypes.Int64Valuable, itemValue reflect.Value, targetElemType reflect.Type) diag.Diagnostics {
	var diags diag.Diagnostics

	int64Val, d := elem.ToInt64Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	switch targetElemType.Kind() {
	case reflect.Int32:
		itemValue.SetInt(int64(int32(int64Val.ValueInt64())))
	case reflect.Int64:
		itemValue.SetInt(int64Val.ValueInt64())
	}
	return diags
}

// convertInt32ValueableToXMLItem converts int32 values to XML wrapper items
func convertInt32ValueableToXMLItem(ctx context.Context, elem basetypes.Int32Valuable, itemValue reflect.Value, targetElemType reflect.Type) diag.Diagnostics {
	var diags diag.Diagnostics

	int32Val, d := elem.ToInt32Value(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if targetElemType.Kind() == reflect.Int32 {
		itemValue.SetInt(int64(int32Val.ValueInt32()))
	}
	return diags
}

// convertComplexValueToXMLItem handles complex type conversions
func convertComplexValueToXMLItem(elem attr.Value, itemValue reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics
	if elem != nil && !elem.IsNull() && !elem.IsUnknown() {
		if itemValue.Type().AssignableTo(reflect.TypeOf(elem)) {
			itemValue.Set(reflect.ValueOf(elem))
		}
	}
	return diags
}

// setXMLWrapperQuantityField sets the Quantity field in XML wrapper struct
func setXMLWrapperQuantityField(quantityField reflect.Value, count int32) error {
	if !quantityField.CanSet() || quantityField.Type().Kind() != reflect.Pointer {
		return fmt.Errorf("quantity field cannot be set")
	}

	quantityPtr := reflect.New(quantityField.Type().Elem())
	quantityPtr.Elem().Set(reflect.ValueOf(count))
	quantityField.Set(quantityPtr)
	return nil
}

// setXMLWrapperQuantity sets the Quantity field based on Items field length
func setXMLWrapperQuantity(vStruct reflect.Value, itemsFieldName string) {
	if vStruct.Kind() == reflect.Pointer {
		if vStruct.IsNil() {
			return
		}
		vStruct = vStruct.Elem()
	}

	itemsField := vStruct.FieldByName(itemsFieldName)
	quantityField := vStruct.FieldByName(xmlWrapperFieldQuantity)

	if itemsField.IsValid() && itemsField.Kind() == reflect.Slice && quantityField.IsValid() {
		count := int32(itemsField.Len())
		_ = setXMLWrapperQuantityField(quantityField, count)
	}
}

// potentialXMLWrapperStruct detects AWS SDK types that follow the XML wrapper pattern
// with Items slice and Quantity pointer fields
func potentialXMLWrapperStruct(t reflect.Type) bool {
	// Check cache first with read lock
	xmlWrapperCacheMutex.RLock()
	result, cached := xmlWrapperCache[t]
	xmlWrapperCacheMutex.RUnlock()

	if cached {
		return result
	}

	result = potentialXMLWrapperStructUncached(t)

	// Cache result with write lock
	xmlWrapperCacheMutex.Lock()
	xmlWrapperCache[t] = result
	xmlWrapperCacheMutex.Unlock()

	return result
}

// potentialXMLWrapperStructUncached performs the actual XML wrapper detection
func potentialXMLWrapperStructUncached(t reflect.Type) bool {
	if t.Kind() != reflect.Struct {
		return false
	}

	hasSliceField := false
	hasQuantityField := false

	// Check if struct has the XML wrapper pattern:
	// - One slice field (Items, Elements, Members, etc.)
	// - One Quantity field (*int32)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := field.Type

		// Check for slice field
		if fieldType.Kind() == reflect.Slice {
			hasSliceField = true
		}

		// Check for Quantity field (pointer to int32)
		if field.Name == xmlWrapperFieldQuantity && fieldType.Kind() == reflect.Pointer && fieldType.Elem().Kind() == reflect.Int32 {
			hasQuantityField = true
		}
	}

	return hasSliceField && hasQuantityField
}

// getXMLWrapperSliceFieldName returns the name of the slice field in an XML wrapper struct
func getXMLWrapperSliceFieldName(t reflect.Type) string {
	if t.Kind() != reflect.Struct {
		return ""
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type.Kind() == reflect.Slice {
			return field.Name
		}
	}

	return ""
}

// nestedObjectCollectionToXMLWrapper converts a NestedObjectCollectionValue to an XML wrapper struct
//
// XML Wrapper Compatibility Rules:
// Rule 1: Items/Quantity only - Direct collection mapping
//
//	AWS: {Items: []T, Quantity: *int32}
//	TF:  Repeatable singular blocks (e.g., lambda_function_association { ... })
//
// Rule 2: Items/Quantity + additional fields - Single plural block
//
//	AWS: {Items: []T, Quantity: *int32, Enabled: *bool, ...}
//	TF:  Single plural block (e.g., trusted_signers { items = [...], enabled = true })
//
// Supports both Rule 1 (Items/Quantity only) and Rule 2 (Items/Quantity + additional fields)
func (expander *autoExpander) nestedObjectCollectionToXMLWrapper(ctx context.Context, _ path.Path, vFrom fwtypes.NestedObjectCollectionValue, _ path.Path, vTo reflect.Value, wrapperField string) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "Expanding NestedObjectCollection to XML wrapper")

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

		// TRACE: Log Rule 2 detection attempt
		tflog.SubsystemTrace(ctx, subsystemName, "TRACE: Checking for Rule 2 pattern", map[string]any{
			"fromSlice_len": fromSlice.Len(),
			"nested_kind":   nestedObj.Kind().String(),
		})

		// Handle pointer to struct (which is what NestedObjectCollection contains)
		if nestedObj.Kind() == reflect.Ptr && !nestedObj.IsNil() {
			structObj := nestedObj.Elem()
			if structObj.Kind() == reflect.Struct {
				// Check if the struct has a wrapper field - indicates Rule 2
				itemsField := structObj.FieldByName(wrapperField)
				if itemsField.IsValid() {
					tflog.SubsystemTrace(ctx, subsystemName, "TRACE: Detected Rule 2 - delegating to expandRule2XMLWrapper", map[string]any{
						"wrapper_field": wrapperField,
					})
					return expander.expandRule2XMLWrapper(ctx, nestedObj, vTo, wrapperField)
				}
			}
		}
	}

	tflog.SubsystemTrace(ctx, subsystemName, "TRACE: Using Rule 1 - direct collection to XML wrapper")

	// Rule 1: Direct collection to XML wrapper (existing logic)
	return expander.expandRule1XMLWrapper(ctx, fromSlice, vTo, wrapperField)
}

// expandRule2XMLWrapper handles Rule 2: single plural block with items + additional fields
func (expander *autoExpander) expandRule2XMLWrapper(ctx context.Context, nestedObjPtr reflect.Value, vTo reflect.Value, wrapperField string) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "Expanding Rule 2 XML wrapper (items + additional fields)")

	// Get target fields
	itemsField := expander.getCachedFieldValue(vTo, wrapperField)
	quantityField := expander.getCachedFieldValue(vTo, xmlWrapperFieldQuantity)

	if !itemsField.IsValid() || !quantityField.IsValid() {
		diags.Append(diagExpandingIncompatibleTypes(nestedObjPtr.Type(), vTo.Type()))
		return diags
	}

	// Get the struct from the pointer
	nestedObj := nestedObjPtr.Elem()

	// Extract wrapper field from nested object
	itemsSourceField := expander.getCachedFieldValue(nestedObj, wrapperField)
	if !itemsSourceField.IsValid() {
		diags.AddError("Missing items field", fmt.Sprintf("Rule 2 XML wrapper requires '%s' field", wrapperField))
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

	// Handle additional fields (e.g., Enabled, CachedMethods)
	for i := 0; i < vTo.NumField(); i++ {
		targetField := vTo.Field(i)
		targetFieldType := vTo.Type().Field(i)
		fieldName := targetFieldType.Name

		// Skip wrapper field and Quantity (already handled)
		if fieldName == wrapperField || fieldName == xmlWrapperFieldQuantity {
			continue
		}

		// Look for matching field in nested object
		sourceField := nestedObj.FieldByName(fieldName)
		if sourceField.IsValid() && targetField.CanAddr() {
			// Check if source field has xmlwrapper tag
			sourceFieldType, ok := nestedObj.Type().FieldByName(fieldName)
			if !ok {
				continue
			}
			_, sourceFieldOpts := autoflexTags(sourceFieldType)

			// Convert TF field to AWS field with fieldOpts
			opts := fieldOpts{
				xmlWrapper:      sourceFieldOpts.XMLWrapperField() != "",
				xmlWrapperField: sourceFieldOpts.XMLWrapperField(),
			}
			diags.Append(expander.convert(ctx, path.Empty(), sourceField, path.Empty(), targetField, opts)...)
			if diags.HasError() {
				return diags
			}
		}
	}

	return diags
}

// expandRule1XMLWrapper handles Rule 1: direct collection to XML wrapper (existing logic)
func (expander *autoExpander) expandRule1XMLWrapper(ctx context.Context, fromSlice reflect.Value, vTo reflect.Value, wrapperField string) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "Expanding Rule 1 XML wrapper (direct collection)")

	// Get the wrapper fields from target struct
	itemsField := expander.getCachedFieldValue(vTo, wrapperField)
	quantityField := expander.getCachedFieldValue(vTo, xmlWrapperFieldQuantity)

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
		} else if targetItemType.Kind() == reflect.Struct {
			// For []struct - need to set the value directly
			newItem := reflect.New(targetItemType)
			diags.Append(autoExpandConvert(ctx, sourceItem.Interface(), newItem.Interface(), expander)...)
			if diags.HasError() {
				return diags
			}
			targetItem.Set(newItem.Elem())
		} else {
			// For primitive types ([]int32, []string, etc.) - direct conversion
			diags.Append(autoExpandConvert(ctx, sourceItem.Interface(), targetItem.Addr().Interface(), expander)...)
			if diags.HasError() {
				return diags
			}
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
	hasSliceField := false
	hasQuantity := false
	hasOtherFields := false

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := field.Name

		if field.Type.Kind() == reflect.Slice {
			hasSliceField = true
		} else if fieldName == xmlWrapperFieldQuantity {
			hasQuantity = true
		} else {
			// Any other field suggests this is a complex collapse target
			hasOtherFields = true
		}
	}

	// Must have slice field + Quantity (XML wrapper pattern) plus other fields
	return hasSliceField && hasQuantity && hasOtherFields
}

// convertCollectionToItemsQuantity converts a source collection to Items slice and Quantity fields
func (expander autoExpander) convertCollectionToItemsQuantity(ctx context.Context, sourceFieldVal reflect.Value, itemsField, quantityField reflect.Value) diag.Diagnostics {
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
			if diags.HasError() {
				return diags
			}
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
			return potentialXMLWrapperStruct(targetType.Elem())
		}
		if targetType.Kind() == reflect.Struct {
			return potentialXMLWrapperStruct(targetType)
		}
	}

	return false
}

// convertToXMLWrapper converts a source collection to an XML wrapper structure
func (expander autoExpander) convertToXMLWrapper(ctx context.Context, sourceFieldVal reflect.Value, targetFieldVal reflect.Value) diag.Diagnostics {
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
			diags.Append(expander.listOrSetOfString(ctx, setValue, targetFieldVal, fieldOpts{})...)
			if diags.HasError() {
				return diags
			}
		}
	case basetypes.ListValuable:
		if listValue, ok := vFrom.(valueWithElementsAs); ok {
			diags.Append(expander.listOrSetOfString(ctx, listValue, targetFieldVal, fieldOpts{})...)
			if diags.HasError() {
				return diags
			}
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

	// Track if we should create the struct (will be set to true if shouldCreateZeroValue)
	var forceCreateStruct bool

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
			// Check if any collection field has omitempty=false (should create zero-value)
			shouldCreateZeroValue := false
			for i := 0; i < typeFrom.NumField(); i++ {
				sourceField := typeFrom.Field(i)
				sourceFieldVal := valFrom.FieldByName(sourceField.Name)
				if sourceValue, ok := sourceFieldVal.Interface().(attr.Value); ok {
					switch sourceValue.(type) {
					case basetypes.SetValuable, basetypes.ListValuable:
						_, sourceFieldOpts := autoflexTags(sourceField)
						if !sourceFieldOpts.OmitEmpty() {
							shouldCreateZeroValue = true
							break
						}
					}
				}
			}

			if !shouldCreateZeroValue {
				// All collection fields are null and have omitempty - set to nil
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

				tflog.SubsystemTrace(ctx, subsystemName, "All source collection fields null with omitempty - setting target to nil")
				return diags
			}
			tflog.SubsystemTrace(ctx, subsystemName, "All source collection fields null but no omitempty - will create zero-value struct")
			// Force creation of zero-value struct even if no source fields are found
			forceCreateStruct = true
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
	wrapperFieldName := getXMLWrapperSliceFieldName(targetStructType)

	if mainSourceFieldName != "" {
		sourceFieldVal := valFrom.FieldByName(mainSourceFieldName)

		// Check if source is a NestedObjectCollectionValue (Rule 2 pattern)
		if sourceValue, ok := sourceFieldVal.Interface().(attr.Value); ok {
			if !sourceValue.IsNull() && !sourceValue.IsUnknown() {
				if nestedObjCollection, ok := sourceValue.(fwtypes.NestedObjectCollectionValue); ok {
					// This is Rule 2: single nested object with items + additional fields
					// Delegate to nestedObjectCollectionToXMLWrapper which handles Rule 2
					tflog.SubsystemTrace(ctx, subsystemName, "Detected Rule 2 pattern - delegating to nestedObjectCollectionToXMLWrapper")
					diags.Append(expander.nestedObjectCollectionToXMLWrapper(ctx, sourcePath.AtName(mainSourceFieldName), nestedObjCollection, targetPath, targetStructVal, wrapperFieldName)...)
					if diags.HasError() {
						return diags
					}
					processedFields[mainSourceFieldName] = true

					// Set the populated struct to the target field
					if isPointer {
						toFieldVal.Set(targetStruct)
					} else {
						toFieldVal.Set(targetStructVal)
					}
					return diags
				}
			} else {
				// Rule 2 field is null - check omitempty
				sourceFieldType, _ := typeFrom.FieldByName(mainSourceFieldName)
				_, sourceFieldOpts := autoflexTags(sourceFieldType)
				if !sourceFieldOpts.OmitEmpty() {
					// Create zero-value struct for round-trip consistency
					tflog.SubsystemDebug(ctx, subsystemName, "Rule 2 field is null but no omitempty - creating zero-value struct", map[string]any{
						"source_field": mainSourceFieldName,
					})
					// Set empty Items and Quantity
					itemsField := targetStructVal.FieldByName(wrapperFieldName)
					quantityField := targetStructVal.FieldByName(xmlWrapperFieldQuantity)
					if itemsField.IsValid() && itemsField.Kind() == reflect.Slice {
						itemsField.Set(reflect.MakeSlice(itemsField.Type(), 0, 0))
					}
					if quantityField.IsValid() && quantityField.Kind() == reflect.Ptr && quantityField.Type().Elem().Kind() == reflect.Int32 {
						zero := int32(0)
						quantityField.Set(reflect.ValueOf(&zero))
					}
					// Set zero values for other fields (e.g., Enabled)
					for i := 0; i < targetStructVal.NumField(); i++ {
						field := targetStructVal.Field(i)
						fieldType := targetStructType.Field(i)
						fieldName := fieldType.Name
						// Skip Items and Quantity (already handled)
						if fieldName == wrapperFieldName || fieldName == xmlWrapperFieldQuantity {
							continue
						}
						// Set zero value for pointer fields
						if field.Kind() == reflect.Ptr && field.CanSet() && field.IsNil() {
							switch fieldType.Type.Elem().Kind() {
							case reflect.Bool:
								falseVal := false
								field.Set(reflect.ValueOf(&falseVal))
							case reflect.Int32:
								zeroVal := int32(0)
								field.Set(reflect.ValueOf(&zeroVal))
							}
						}
					}
					processedFields[mainSourceFieldName] = true

					// Set the struct to the target field
					if isPointer {
						toFieldVal.Set(targetStruct)
					} else {
						toFieldVal.Set(targetStructVal)
					}
					return diags
				}
			}
		}

		// Get the wrapper fields in the target
		itemsField := targetStructVal.FieldByName(wrapperFieldName)
		quantityField := targetStructVal.FieldByName(xmlWrapperFieldQuantity)

		if itemsField.IsValid() && quantityField.IsValid() {
			// Check if the source field is actually usable (not null/unknown)
			if sourceValue, ok := sourceFieldVal.Interface().(attr.Value); ok {
				if !sourceValue.IsNull() && !sourceValue.IsUnknown() {
					// Convert the collection to wrapper slice and Quantity
					diags.Append(expander.convertCollectionToItemsQuantity(ctx, sourceFieldVal, itemsField, quantityField)...)
					if diags.HasError() {
						return diags
					}
					hasMainSourceField = true
				} else {
					// Check if source field has omitempty tag
					sourceFieldType, _ := typeFrom.FieldByName(mainSourceFieldName)
					_, sourceFieldOpts := autoflexTags(sourceFieldType)
					if sourceFieldOpts.OmitEmpty() {
						tflog.SubsystemDebug(ctx, subsystemName, "Main source field is null and has omitempty - skipping XML wrapper creation", map[string]any{
							"source_field": mainSourceFieldName,
						})
					} else {
						// Create zero-value struct for round-trip consistency
						tflog.SubsystemDebug(ctx, subsystemName, "Main source field is null but no omitempty - creating zero-value XML wrapper", map[string]any{
							"source_field": mainSourceFieldName,
						})
						// Set Items to empty slice and Quantity to 0
						if itemsField.Kind() == reflect.Slice {
							itemsField.Set(reflect.MakeSlice(itemsField.Type(), 0, 0))
						}
						if quantityField.Kind() == reflect.Ptr && quantityField.Type().Elem().Kind() == reflect.Int32 {
							zero := int32(0)
							quantityField.Set(reflect.ValueOf(&zero))
						}
						hasMainSourceField = true
					}
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

		// Skip wrapper field and Quantity as they were handled above
		if targetFieldName == wrapperFieldName || targetFieldName == xmlWrapperFieldQuantity {
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
				diags.Append(expander.convertToXMLWrapper(ctx, sourceFieldVal, targetFieldVal)...)
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

	// Only set the constructed struct if we found source fields to populate it OR if we're forcing creation
	if hasAnySourceFields || forceCreateStruct {
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
