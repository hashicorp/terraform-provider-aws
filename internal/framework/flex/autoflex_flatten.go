// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"context"
	"fmt"
	"iter"
	"reflect"
	"strings"
	"time"

	smithydocument "github.com/aws/smithy-go/document"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfreflect "github.com/hashicorp/terraform-provider-aws/internal/reflect"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	"github.com/shopspring/decimal"
)

// Flatten = AWS --> TF

// Flattener is implemented by types that customize their flattening
type Flattener interface {
	Flatten(ctx context.Context, v any) diag.Diagnostics
}

// Flatten "flattens" an AWS SDK for Go v2 API data structure into
// a resource's "business logic" data structure, implemented using
// Terraform Plugin Framework data types.
// The API data structure's fields are walked and exported fields that
// have a corresponding field in the resource's data structure (and a
// suitable target data type) are copied.
func Flatten(ctx context.Context, apiObject, tfObject any, optFns ...AutoFlexOptionsFunc) diag.Diagnostics {
	var diags diag.Diagnostics
	flattener := newAutoFlattener(optFns)

	tflog.SubsystemInfo(ctx, subsystemName, "Flattening", map[string]any{
		logAttrKeySourceType: fullTypeName(reflect.TypeOf(apiObject)),
		logAttrKeyTargetType: fullTypeName(reflect.TypeOf(tfObject)),
	})

	diags.Append(autoFlattenConvert(ctx, apiObject, tfObject, flattener)...)

	return diags
}

type autoFlattener struct {
	Options    AutoFlexOptions
	fieldCache map[reflect.Type]map[string]reflect.StructField
}

// newAutoFlattener initializes an auto-flattener with defaults that can be overridden
// via functional options
func newAutoFlattener(optFns []AutoFlexOptionsFunc) *autoFlattener {
	o := AutoFlexOptions{
		ignoredFieldNames: DefaultIgnoredFieldNames,
	}

	for _, optFn := range optFns {
		optFn(&o)
	}

	return &autoFlattener{
		Options:    o,
		fieldCache: make(map[reflect.Type]map[string]reflect.StructField),
	}
}

func (flattener autoFlattener) getOptions() AutoFlexOptions {
	return flattener.Options
}

// autoFlattenConvert converts `from` to `to` using the specified auto-flexer.
func autoFlattenConvert(ctx context.Context, from, to any, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	sourcePath := path.Empty()
	targetPath := path.Empty()

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(reflect.TypeOf(from)))
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(reflect.TypeOf(to)))

	if _, ok := to.(attr.Value); !ok {
		vf := reflect.ValueOf(from)
		if vf.Kind() == reflect.Invalid || vf.Kind() == reflect.Pointer && vf.IsNil() {
			tflog.SubsystemError(ctx, subsystemName, "Source is nil")
			diags.Append(diagFlatteningSourceIsNil(reflect.TypeOf(from)))
			return diags
		}
	}

	ctx, valFrom, valTo, d := autoFlexValues(ctx, from, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	// Top-level struct to struct conversion.
	if valFrom.IsValid() && valTo.IsValid() {
		if typFrom, typTo := valFrom.Type(), valTo.Type(); typFrom.Kind() == reflect.Struct && typTo.Kind() == reflect.Struct &&
			!typTo.Implements(reflect.TypeFor[basetypes.ListValuable]()) &&
			!typTo.Implements(reflect.TypeFor[basetypes.SetValuable]()) {
			// Special case: Check if source is XML wrapper struct and target has xmlwrapper fields
			if potentialXMLWrapperStruct(typFrom) {
				tflog.SubsystemTrace(ctx, subsystemName, "Source is XML wrapper struct")
				// Check if target has any fields with xmlwrapper tags
				for toField := range tfreflect.ExportedStructFields(typTo) {
					_, toOpts := autoflexTags(toField)
					if sourceFieldName := toOpts.XMLWrapperField(); sourceFieldName != "" {
						// Found xmlwrapper tag, handle direct XML wrapper conversion
						tflog.SubsystemTrace(ctx, subsystemName, "Direct XML wrapper struct conversion", map[string]any{
							logAttrKeySourceFieldname: sourceFieldName,
							logAttrKeyTargetFieldname: toField.Name,
						})
						diags.Append(handleDirectXMLWrapperStruct(ctx, sourcePath, sourceFieldName, valFrom, valTo, typFrom, typTo, targetPath, toField.Name, flexer)...)
						return diags
					}
				}
				tflog.SubsystemTrace(ctx, subsystemName, "No xmlwrapper tags found in target")
			} else {
				tflog.SubsystemTrace(ctx, subsystemName, "Source is not XML wrapper struct")
			}

			tflog.SubsystemInfo(ctx, subsystemName, "Converting")
			diags.Append(flattenStruct(ctx, sourcePath, from, targetPath, to, flexer)...)
			return diags
		}
	}

	valFrom = reflect.ValueOf(from)

	// Anything else.
	diags.Append(flexer.convert(ctx, sourcePath, valFrom, targetPath, valTo, fieldOpts{})...)
	return diags
}

// convert converts a single AWS API value to its Plugin Framework equivalent.
func (flattener autoFlattener) convert(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(valueType(vFrom)))
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(valueType(vTo)))

	valTo, ok := vTo.Interface().(attr.Value)
	if !ok {
		// Check for `nil` (i.e. Kind == Invalid) here, because we allow primitive types to be `nil`
		if vFrom.Kind() == reflect.Invalid {
			tflog.SubsystemError(ctx, subsystemName, "Source is nil")
			diags.Append(diagFlatteningSourceIsNil(valueType(vFrom)))
			return diags
		}

		tflog.SubsystemError(ctx, subsystemName, "Target does not implement attr.Value")
		diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(vTo.Interface())))
		return diags
	}

	tflog.SubsystemInfo(ctx, subsystemName, "Converting")

	// main control flow
	tTo := valTo.Type(ctx)
	switch k := vFrom.Kind(); k {
	case reflect.Bool:
		diags.Append(flattener.bool(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Float64:
		diags.Append(flattener.float64(ctx, vFrom, vFrom.Type(), false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Float32:
		diags.Append(flattener.float32(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Int64:
		diags.Append(flattener.int64(ctx, vFrom, vFrom.Type(), false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Int32:
		diags.Append(flattener.int32(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.String:
		// []byte (or []uint8) is also handled like string but in flattener.slice()
		diags.Append(flattener.string(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Pointer:
		diags.Append(flattener.pointer(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Slice:
		diags.Append(flattener.slice(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Map:
		diags.Append(flattener.map_(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Struct:
		diags.Append(flattener.struct_(ctx, sourcePath, vFrom, false, targetPath, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Interface:
		diags.Append(flattener.interface_(ctx, vFrom, tTo, vTo, fieldOpts)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// bool copies an AWS API bool value to a compatible Plugin Framework value.
func (flattener autoFlattener) bool(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.BoolTypable:
		//
		// bool -> types.Bool.
		//
		var boolValue types.Bool
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				boolValue = types.BoolValue(false)
			} else {
				boolValue = types.BoolValue(vFrom.Bool())
			}
		} else {
			if isNullFrom {
				boolValue = types.BoolNull()
			} else {
				boolValue = types.BoolValue(vFrom.Bool())
			}
		}
		v, d := tTo.ValueFromBool(ctx, boolValue)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// float64 copies an AWS API float64 value to a compatible Plugin Framework value.
func (flattener autoFlattener) float64(ctx context.Context, vFrom reflect.Value, sourceType reflect.Type, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Float64Typable:
		//
		// float64 -> types.Float64.
		//
		var float64Value types.Float64
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				float64Value = types.Float64Value(0)
			} else {
				float64Value = types.Float64Value(vFrom.Float())
			}
		} else {
			if isNullFrom {
				float64Value = types.Float64Null()
			} else {
				float64Value = types.Float64Value(vFrom.Float())
			}
		}
		v, d := tTo.ValueFromFloat64(ctx, float64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case basetypes.Float32Typable:
		// Only returns an error when the target type is Float32Typable to prevent breaking existing resources
		tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")
		diags.Append(DiagFlatteningIncompatibleTypes(sourceType, vTo.Type()))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// float32 copies an AWS API float32 value to a compatible Plugin Framework value.
func (flattener autoFlattener) float32(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Float64Typable:
		//
		// float32 -> types.Float64.
		//
		var float64Value types.Float64
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				float64Value = types.Float64Value(0)
			} else {
				// Avoid loss of equivalence.
				from := vFrom.Interface().(float32)
				float64Value = types.Float64Value(decimal.NewFromFloat32(from).InexactFloat64())
			}
		} else {
			if isNullFrom {
				float64Value = types.Float64Null()
			} else {
				// Avoid loss of equivalence.
				from := vFrom.Interface().(float32)
				float64Value = types.Float64Value(decimal.NewFromFloat32(from).InexactFloat64())
			}
		}
		v, d := tTo.ValueFromFloat64(ctx, float64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case basetypes.Float32Typable:
		//
		// float32/float64 -> types.Float32.
		//
		var float32Value types.Float32
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				float32Value = types.Float32Value(0)
			} else {
				// Avoid loss of equivalence.
				from := vFrom.Interface().(float32)
				float32Value = types.Float32Value(float32(decimal.NewFromFloat32(from).InexactFloat64()))
			}
		} else {
			if isNullFrom {
				float32Value = types.Float32Null()
			} else {
				// Avoid loss of equivalence.
				from := vFrom.Interface().(float32)
				float32Value = types.Float32Value(float32(decimal.NewFromFloat32(from).InexactFloat64()))
			}
		}
		v, d := tTo.ValueFromFloat32(ctx, float32Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// int64 copies an AWS API int64 value to a compatible Plugin Framework value.
func (flattener autoFlattener) int64(ctx context.Context, vFrom reflect.Value, sourceType reflect.Type, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Int64Typable:
		//
		// int32/int64 -> types.Int64.
		//
		var int64Value types.Int64
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				int64Value = types.Int64Value(0)
			} else {
				int64Value = types.Int64Value(vFrom.Int())
			}
		} else {
			if isNullFrom {
				int64Value = types.Int64Null()
			} else {
				int64Value = types.Int64Value(vFrom.Int())
			}
		}
		v, d := tTo.ValueFromInt64(ctx, int64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case basetypes.Int32Typable:
		// Only returns an error when the target type is Int32Typeable to prevent breaking existing resources
		tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")
		diags.Append(DiagFlatteningIncompatibleTypes(sourceType, vTo.Type()))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// int32 copies an AWS API int32 value to a compatible Plugin Framework value.
func (flattener autoFlattener) int32(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.Int64Typable:
		//
		// int32/int64 -> types.Int64.
		//
		var int64Value types.Int64
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				int64Value = types.Int64Value(0)
			} else {
				int64Value = types.Int64Value(vFrom.Int())
			}
		} else {
			if isNullFrom {
				int64Value = types.Int64Null()
			} else {
				int64Value = types.Int64Value(vFrom.Int())
			}
		}
		v, d := tTo.ValueFromInt64(ctx, int64Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags

	case basetypes.Int32Typable:
		//
		// int32/int64 -> types.Int32.
		//
		var int32Value basetypes.Int32Value
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				int32Value = types.Int32Value(0)
			} else {
				int32Value = types.Int32Value(int32(vFrom.Int()))
			}
		} else {
			if isNullFrom {
				int32Value = types.Int32Null()
			} else {
				int32Value = types.Int32Value(int32(vFrom.Int()))
			}
		}
		v, d := tTo.ValueFromInt32(ctx, int32Value)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// string copies an AWS API string value to a compatible Plugin Framework value.
func (flattener autoFlattener) string(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.StringTypable:
		stringValue := types.StringNull()
		if fieldOpts.legacy {
			tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
			if isNullFrom {
				stringValue = types.StringValue("")
			} else {
				stringValue = types.StringValue(vFrom.String())
			}
		} else {
			if !isNullFrom {
				if value := vFrom.String(); !(value == "" && (strings.HasPrefix(tTo.String(), "StringEnumType[") || fieldOpts.omitempty)) {
					stringValue = types.StringValue(value)
				}
			}
		}
		v, d := tTo.ValueFromString(ctx, stringValue)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		//
		// string -> types.String.
		//
		vTo.Set(reflect.ValueOf(v))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")
	// TODO: Should continue failing silently for now
	// diags.Append(DiagFlatteningIncompatibleTypes(vFrom.Type(), vTo.Type()))

	return diags
}

func (flattener autoFlattener) time(ctx context.Context, vFrom reflect.Value, isNullFrom bool, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if isNullFrom {
		vTo.Set(reflect.ValueOf(timetypes.NewRFC3339Null()))
		return diags
	}

	if !vFrom.CanInterface() {
		diags.AddError("AutoFlEx", fmt.Sprintf("cannot create an interface for: %T", vFrom))
		return diags
	}

	// time.Time --> timetypes.RFC3339
	if from, ok := vFrom.Interface().(time.Time); ok {
		vTo.Set(reflect.ValueOf(timetypes.NewRFC3339TimeValue(from)))
		return diags
	}

	if !vFrom.Elem().CanInterface() {
		diags.AddError("AutoFlEx", fmt.Sprintf("cannot create an interface for: %T", vFrom.Elem()))
		return diags
	}

	// *time.Time --> timetypes.RFC3339
	if from, ok := vFrom.Elem().Interface().(time.Time); ok {
		vTo.Set(reflect.ValueOf(timetypes.NewRFC3339TimeValue(from)))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   vTo,
	})

	return diags
}

// pointer copies an AWS API pointer value to a compatible Plugin Framework value.
func (flattener autoFlattener) pointer(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch vElem, isNilFrom := vFrom.Elem(), vFrom.IsNil(); vFrom.Type().Elem().Kind() {
	case reflect.Bool:
		diags.Append(flattener.bool(ctx, vElem, isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Float64:
		diags.Append(flattener.float64(ctx, vElem, vFrom.Type(), isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Float32:
		diags.Append(flattener.float32(ctx, vElem, isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Int64:
		diags.Append(flattener.int64(ctx, vElem, vFrom.Type(), isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Int32:
		diags.Append(flattener.int32(ctx, vElem, isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.String:
		diags.Append(flattener.string(ctx, vElem, isNilFrom, tTo, vTo, fieldOpts)...)
		return diags

	case reflect.Struct:
		diags.Append(flattener.struct_(ctx, sourcePath, vElem, isNilFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

func (flattener autoFlattener) interface_(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value, _ fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tTo := tTo.(type) {
	case basetypes.StringTypable:
		//
		// JSONStringer -> types.String-ish.
		//
		if vFrom.Type().Implements(reflect.TypeFor[smithydocument.Marshaler]()) {
			tflog.SubsystemInfo(ctx, subsystemName, "Source implements smithydocument.Marshaler")

			stringValue := types.StringNull()

			if vFrom.IsNil() {
				tflog.SubsystemTrace(ctx, subsystemName, "Flattening null value")
			} else {
				doc := vFrom.Interface().(smithydocument.Marshaler)
				s, err := tfsmithy.DocumentToJSONString(doc)
				if err != nil {
					tflog.SubsystemError(ctx, subsystemName, "Marshalling JSON document", map[string]any{
						logAttrKeyError: err.Error(),
					})
					diags.Append(diagFlatteningMarshalSmithyDocument(reflect.TypeOf(doc), err))
					return diags
				}
				stringValue = types.StringValue(s)
			}
			v, d := tTo.ValueFromString(ctx, stringValue)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(v))
			return diags
		}

	case fwtypes.NestedObjectType:
		//
		// interface -> types.List(OfObject) or types.Object.
		//
		diags.Append(flattener.interfaceToNestedObject(ctx, vFrom, vFrom.IsNil(), tTo, vTo)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")

	return diags
}

// struct_ copies an AWS API struct value to a compatible Plugin Framework value.
func (flattener autoFlattener) struct_(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, isNilFrom bool, targetPath path.Path, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if tTo, ok := tTo.(fwtypes.NestedObjectType); ok {
		//
		// *struct -> types.List(OfObject) or types.Object.
		//
		diags.Append(flattener.structToNestedObject(ctx, sourcePath, vFrom, isNilFrom, targetPath, tTo, vTo, fieldOpts)...)
		return diags
	}

	iTo, ok := vTo.Interface().(attr.Value)
	if !ok {
		diags.AddError("AutoFlEx", fmt.Sprintf("does not implement attr.Value: %s", vTo.Kind()))
		return diags
	}

	switch iTo.(type) {
	case timetypes.RFC3339:
		diags.Append(flattener.time(ctx, vFrom, isNilFrom, vTo)...)
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "Flattening incompatible types")

	return diags
}

// slice copies an AWS API slice value to a compatible Plugin Framework value.
func (flattener autoFlattener) slice(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tSliceElem := vFrom.Type().Elem(); tSliceElem.Kind() {
	case reflect.Int32, reflect.Int64:
		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []int32 or []int64 -> types.List(OfInt64).
			//
			diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectValue, fieldOpts)...)
			return diags

		case basetypes.SetTypable:
			//
			// []int32 or []int64 -> types.Set(OfInt64).
			//
			diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectValue, fieldOpts)...)
			return diags
		}

	case reflect.String:
		var elementType attr.Type = types.StringType
		attrValueFromReflectValue := newStringValueFromReflectValue
		if tTo, ok := tTo.(attr.TypeWithElementType); ok {
			if tElem, ok := tTo.ElementType().(basetypes.StringTypable); ok {
				//
				// []stringy -> types.List/Set(OfStringEnum).
				//
				elementType = tElem
				attrValueFromReflectValue = func(val reflect.Value) (attr.Value, diag.Diagnostics) {
					return tElem.ValueFromString(ctx, types.StringValue(val.String()))
				}
			}
		}

		switch tTo := tTo.(type) {
		case basetypes.ListTypable:
			//
			// []string -> types.List(OfString).
			//
			diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, elementType, attrValueFromReflectValue, fieldOpts)...)
			return diags

		case basetypes.SetTypable:
			//
			// []string -> types.Set(OfString).
			//
			diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, elementType, attrValueFromReflectValue, fieldOpts)...)
			return diags
		}

	case reflect.Uint8:
		switch tTo := tTo.(type) {
		case basetypes.StringTypable:
			//
			// []byte (or []uint8) -> types.String.
			//
			vFrom = reflect.ValueOf(string(vFrom.Bytes()))
			diags.Append(flattener.string(ctx, vFrom, false, tTo, vTo, fieldOpts)...)
			return diags
		}

	case reflect.Pointer:
		switch tSliceElem.Elem().Kind() {
		case reflect.Int32:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*int32 -> types.List(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue, fieldOpts)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*int32 -> types.Set(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue, fieldOpts)...)
				return diags
			}

		case reflect.Int64:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*int64 -> types.List(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue, fieldOpts)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*int64 -> types.Set(OfInt64).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.Int64Type, newInt64ValueFromReflectPointerValue, fieldOpts)...)
				return diags
			}

		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.ListTypable:
				//
				// []*string -> types.List(OfString).
				//
				diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, types.StringType, newStringValueFromReflectPointerValue, fieldOpts)...)
				return diags

			case basetypes.SetTypable:
				//
				// []*string -> types.Set(OfString).
				//
				diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, types.StringType, newStringValueFromReflectPointerValue, fieldOpts)...)
				return diags
			}

		case reflect.Struct:
			if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
				//
				// []*struct -> types.List(OfObject).
				//
				diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
				return diags
			}
		}

	case reflect.Struct:
		if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
			//
			// []struct -> types.List(OfObject).
			//
			diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
			return diags
		}

	case reflect.Interface:
		if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
			//
			// []interface -> types.List(OfObject).
			//
			diags.Append(flattener.sliceOfStructToNestedObjectCollection(ctx, sourcePath, vFrom, targetPath, tTo, vTo, fieldOpts)...)
			return diags
		}

	default:
		// Check for custom types with string underlying type (e.g., enums)
		// For type declarations like "type MyEnum string", Kind() will not be reflect.String,
		// but we can convert values to string using String() method
		if tSliceElem.Kind() != reflect.String {
			// Try to see if this is a string-based custom type by checking if we can call String() on it
			sampleVal := reflect.New(tSliceElem).Elem()
			if sampleVal.CanInterface() {
				// Check if it has an underlying string type
				if tSliceElem.ConvertibleTo(reflect.TypeFor[string]()) {
					var elementType attr.Type = types.StringType
					attrValueFromReflectValue := newStringValueFromReflectValue
					if tTo, ok := tTo.(attr.TypeWithElementType); ok {
						if tElem, ok := tTo.ElementType().(basetypes.StringTypable); ok {
							//
							// []stringy -> types.List/Set(OfStringEnum).
							//
							elementType = tElem
							attrValueFromReflectValue = func(val reflect.Value) (attr.Value, diag.Diagnostics) {
								return tElem.ValueFromString(ctx, types.StringValue(val.String()))
							}
						}
					}

					switch tTo := tTo.(type) {
					case basetypes.ListTypable:
						//
						// []custom_string_type -> types.List(OfString).
						//
						diags.Append(flattener.sliceOfPrimtiveToList(ctx, vFrom, tTo, vTo, elementType, attrValueFromReflectValue, fieldOpts)...)
						return diags

					case basetypes.SetTypable:
						//
						// []custom_string_type -> types.Set(OfString).
						//
						diags.Append(flattener.sliceOfPrimitiveToSet(ctx, vFrom, tTo, vTo, elementType, attrValueFromReflectValue, fieldOpts)...)
						return diags
					}
				}
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// map_ copies an AWS API map value to a compatible Plugin Framework value.
func (flattener autoFlattener) map_(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo attr.Type, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	switch tMapKey := vFrom.Type().Key(); tMapKey.Kind() {
	case reflect.String:
		switch tMapElem := vFrom.Type().Elem(); tMapElem.Kind() {
		case reflect.Struct:
			switch tTo := tTo.(type) {
			case basetypes.SetTypable:
				//
				// map[string]struct -> fwtypes.SetNestedObjectOf[Object]
				//
				if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
					diags.Append(flattener.structMapToObjectList(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
					return diags
				}

			case basetypes.ListTypable:
				//
				// map[string]struct -> fwtypes.ListNestedObjectOf[Object]
				//
				if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
					diags.Append(flattener.structMapToObjectList(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
					return diags
				}
			}

		case reflect.String:
			switch tTo := tTo.(type) {
			case basetypes.MapTypable:
				//
				// map[string]string -> types.Map(OfString).
				//
				if vFrom.IsNil() {
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapNull")
					to, d := tTo.ValueFromMap(ctx, types.MapNull(types.StringType))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}

				from := vFrom.Interface().(map[string]string)
				tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapValue", map[string]any{
					logAttrKeySourceSize: len(from),
				})
				elements := make(map[string]attr.Value, len(from))
				for k, v := range from {
					elements[k] = types.StringValue(v)
				}
				map_, d := types.MapValue(types.StringType, elements)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				to, d := tTo.ValueFromMap(ctx, map_)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}

		case reflect.Map:
			switch tTo := tTo.(type) {
			case basetypes.MapTypable:
				//
				// map[string]map[string]string -> types.Map(OfMap[types.String]).
				//
				if vFrom.IsNil() {
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapNull")
					to, d := tTo.ValueFromMap(ctx, types.MapNull(types.MapType{ElemType: types.StringType}))
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}

				switch tMapElem.Elem().Kind() {
				case reflect.String:
					from := vFrom.Interface().(map[string]map[string]string)
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening map", map[string]any{
						logAttrKeySourceSize: len(from),
					})

					// Check for omitempty: if all inner maps are empty, return null
					if fieldOpts.omitempty {
						allEmpty := true
						for _, innerMap := range from {
							if len(innerMap) > 0 {
								allEmpty = false
								break
							}
						}
						if allEmpty {
							tflog.SubsystemTrace(ctx, subsystemName, "All inner maps empty with omitempty, returning null")
							to, d := tTo.ValueFromMap(ctx, types.MapNull(types.MapType{ElemType: types.StringType}))
							diags.Append(d...)
							if !diags.HasError() {
								vTo.Set(reflect.ValueOf(to))
							}
							return diags
						}
					}

					elements := make(map[string]attr.Value, len(from))
					for k, v := range from {
						innerElements := make(map[string]attr.Value, len(v))
						for ik, iv := range v {
							innerElements[ik] = types.StringValue(iv)
						}
						tflog.SubsystemTrace(ctx, subsystemName, "Flattening with NewMapValueOf", map[string]any{
							logAttrKeySourcePath: sourcePath.AtMapKey(k).String(),
							logAttrKeySourceType: fullTypeName(reflect.TypeOf(v)),
							logAttrKeySourceSize: len(v),
							logAttrKeyTargetPath: targetPath.AtMapKey(k).String(),
							logAttrKeyTargetType: fullTypeName(reflect.TypeFor[map[string]attr.Value]()),
						})
						innerMap, d := fwtypes.NewMapValueOf[types.String](ctx, innerElements)
						diags.Append(d...)
						if diags.HasError() {
							return diags
						}

						elements[k] = innerMap
					}
					map_, d := fwtypes.NewMapValueOf[fwtypes.MapValueOf[types.String]](ctx, elements)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					to, d := tTo.ValueFromMap(ctx, map_.MapValue)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags

				case reflect.Pointer:
					from := vFrom.Interface().(map[string]map[string]*string)
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening map", map[string]any{
						logAttrKeySourceSize: len(from),
					})
					elements := make(map[string]attr.Value, len(from))
					for k, v := range from {
						innerElements := make(map[string]attr.Value, len(v))
						for ik, iv := range v {
							innerElements[ik] = types.StringValue(*iv)
						}
						tflog.SubsystemTrace(ctx, subsystemName, "Flattening with NewMapValueOf", map[string]any{
							logAttrKeySourcePath: sourcePath.AtMapKey(k).String(),
							logAttrKeySourceType: fullTypeName(reflect.TypeOf(v)),
							logAttrKeySourceSize: len(v),
							logAttrKeyTargetPath: targetPath.AtMapKey(k).String(),
							logAttrKeyTargetType: fullTypeName(reflect.TypeFor[map[string]attr.Value]()),
						})
						innerMap, d := fwtypes.NewMapValueOf[types.String](ctx, innerElements)
						diags.Append(d...)
						if diags.HasError() {
							return diags
						}

						elements[k] = innerMap
					}
					map_, d := fwtypes.NewMapValueOf[fwtypes.MapValueOf[types.String]](ctx, elements)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					to, d := tTo.ValueFromMap(ctx, map_.MapValue)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}
			}

		case reflect.Pointer:
			switch tMapElem.Elem().Kind() {
			case reflect.Struct:
				if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
					diags.Append(flattener.structMapToObjectList(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
					return diags
				}

			case reflect.String:
				switch tTo := tTo.(type) {
				case basetypes.ListTypable:
					//
					// map[string]struct -> fwtypes.ListNestedObjectOf[Object]
					//
					if tTo, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
						diags.Append(flattener.structMapToObjectList(ctx, sourcePath, vFrom, targetPath, tTo, vTo)...)
						return diags
					}

				case basetypes.MapTypable:
					//
					// map[string]*string -> types.Map(OfString).
					//
					if vFrom.IsNil() {
						tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapNull")
						to, d := tTo.ValueFromMap(ctx, types.MapNull(types.StringType))
						diags.Append(d...)
						if diags.HasError() {
							return diags
						}

						vTo.Set(reflect.ValueOf(to))
						return diags
					}

					from := vFrom.Interface().(map[string]*string)
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening with MapValue", map[string]any{
						logAttrKeySourceSize: len(from),
					})
					elements := make(map[string]attr.Value, len(from))
					for k, v := range from {
						elements[k] = types.StringPointerValue(v)
					}
					map_, d := types.MapValue(types.StringType, elements)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					to, d := tTo.ValueFromMap(ctx, map_)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}

					vTo.Set(reflect.ValueOf(to))
					return diags
				}
			}
		}
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

func (flattener autoFlattener) structMapToObjectList(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if vFrom.IsNil() {
		val, d := tTo.NullValue(ctx)
		tflog.SubsystemTrace(ctx, subsystemName, "Flattening null value")
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))
		return diags
	}

	n := vFrom.Len()
	to, d := tTo.NewObjectSlice(ctx, n, n)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	t := reflect.ValueOf(to)

	i := 0
	for _, key := range vFrom.MapKeys() {
		sourcePath := sourcePath.AtMapKey(key.Interface().(string))
		targetPath := targetPath.AtListIndex(i)
		ctx := tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())
		target, d := tTo.NewObjectPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		fromInterface := vFrom.MapIndex(key).Interface()
		if vFrom.MapIndex(key).Kind() == reflect.Pointer {
			fromInterface = vFrom.MapIndex(key).Elem().Interface()
		}

		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(reflect.TypeOf(fromInterface)))

		diags.Append(setMapBlockKey(ctx, target, key)...)
		if diags.HasError() {
			return diags
		}

		diags.Append(flattenStruct(ctx, sourcePath, fromInterface, targetPath, target, flattener)...)
		if diags.HasError() {
			return diags
		}

		t.Index(i).Set(reflect.ValueOf(target))
		i++
	}

	val, d := tTo.ValueFromObjectSlice(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))

	return diags
}

// structToNestedObject copies an AWS API struct value to a compatible Plugin Framework NestedObjectValue value.
func (flattener autoFlattener) structToNestedObject(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, isNullFrom bool, targetPath path.Path, tTo fwtypes.NestedObjectType, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if fieldOpts.legacy {
		tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
		if isNullFrom {
			to, d := tTo.NewObjectPtr(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			val, d := tTo.ValueFromObjectPtr(ctx, to)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(val))
			return diags
		}
	} else {
		if isNullFrom {
			val, d := tTo.NullValue(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(val))
			return diags
		}
	}

	// Check if source is a Rule 2 XML wrapper struct with all zero values and omitempty
	if potentialXMLWrapperStruct(vFrom.Type()) && fieldOpts.omitempty {
		wrapperFieldName := getXMLWrapperSliceFieldName(vFrom.Type())
		itemsField := vFrom.FieldByName(wrapperFieldName)

		if itemsField.IsValid() {
			itemsEmpty := itemsField.IsNil() || (itemsField.Kind() == reflect.Slice && itemsField.Len() == 0)

			if itemsEmpty {
				allOtherFieldsZero := true
				for i := 0; i < vFrom.NumField(); i++ {
					sourceField := vFrom.Field(i)
					fieldName := vFrom.Type().Field(i).Name

					if fieldName == wrapperFieldName || fieldName == xmlWrapperFieldQuantity {
						continue
					}

					isFieldZero := sourceField.Kind() == reflect.Pointer && sourceField.IsNil() ||
						sourceField.Kind() == reflect.Pointer && sourceField.Elem().IsZero() ||
						sourceField.Kind() != reflect.Pointer && sourceField.IsZero()

					if !isFieldZero {
						allOtherFieldsZero = false
						break
					}
				}

				if allOtherFieldsZero {
					tflog.SubsystemTrace(ctx, subsystemName, "Nested XML wrapper struct has all zero values with omitempty, returning null")
					val, d := tTo.NullValue(ctx)
					diags.Append(d...)
					if !diags.HasError() {
						vTo.Set(reflect.ValueOf(val))
					}
					return diags
				}
			}
		}
	}

	// Check if source is a regular struct with all zero values and omitempty
	if !potentialXMLWrapperStruct(vFrom.Type()) && fieldOpts.omitempty {
		allFieldsZero := true
		for i := 0; i < vFrom.NumField(); i++ {
			sourceField := vFrom.Field(i)
			isFieldZero := sourceField.Kind() == reflect.Pointer && sourceField.IsNil() ||
				sourceField.Kind() == reflect.Pointer && sourceField.Elem().IsZero() ||
				sourceField.Kind() != reflect.Pointer && sourceField.IsZero()

			if !isFieldZero {
				allFieldsZero = false
				break
			}
		}

		if allFieldsZero {
			tflog.SubsystemTrace(ctx, subsystemName, "Nested struct has all zero values with omitempty, returning null")
			val, d := tTo.NullValue(ctx)
			diags.Append(d...)
			if !diags.HasError() {
				vTo.Set(reflect.ValueOf(val))
			}
			return diags
		}
	}

	// Create a new target structure and walk its fields.
	to, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	diags.Append(flattenStruct(ctx, sourcePath, vFrom.Interface(), targetPath, to, flattener)...)
	if diags.HasError() {
		return diags
	}

	// Set the target structure as a mapped Object.
	val, d := tTo.ValueFromObjectPtr(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

// interfaceToNestedObject copies an AWS API interface value to a compatible Plugin Framework NestedObjectValue value.
func (flattener autoFlattener) interfaceToNestedObject(ctx context.Context, vFrom reflect.Value, isNullFrom bool, tTo fwtypes.NestedObjectType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if isNullFrom {
		val, d := tTo.NullValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))
		return diags
	}

	to, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	toFlattener, ok := to.(Flattener)
	if !ok {
		val, d := tTo.NullValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(val))

		tflog.SubsystemError(ctx, subsystemName, "Source does not implement flex.Flattener")
		// diags.Append(diagFlatteningTargetDoesNotImplementFlexFlattener(reflect.TypeOf(vTo.Interface())))
		return diags
	}

	tflog.SubsystemInfo(ctx, subsystemName, "Target implements flex.Flattener")

	// Dereference interface
	vFrom = vFrom.Elem()
	// If it's a pointer, dereference again to get the underlying type
	if vFrom.Kind() == reflect.Pointer {
		vFrom = vFrom.Elem()
	}

	diags.Append(flattenFlattener(ctx, vFrom, toFlattener)...)
	if diags.HasError() {
		return diags
	}

	// Set the target structure as a mapped Object.
	val, d := tTo.ValueFromObjectPtr(ctx, toFlattener)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

// sliceOfPrimtiveToList copies an AWS API slice of primitive (or pointer to primitive) value to a compatible Plugin Framework List value.
func (flattener autoFlattener) sliceOfPrimtiveToList(ctx context.Context, vFrom reflect.Value, tTo basetypes.ListTypable, vTo reflect.Value, elementType attr.Type, attrValueFromReflectValue attrValueFromReflectValueFunc, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if fieldOpts.legacy {
		tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
		if vFrom.IsNil() {
			list, d := types.ListValue(elementType, []attr.Value{})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			to, d := tTo.ValueFromList(ctx, list)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags
		}
	} else {
		if vFrom.IsNil() {
			if fieldOpts.legacy {
				tflog.SubsystemTrace(ctx, subsystemName, "Flattening with ListValue (empty for nil in legacy mode)")
				list, d := types.ListValue(elementType, []attr.Value{})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				to, d := tTo.ValueFromList(ctx, list)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}

			tflog.SubsystemTrace(ctx, subsystemName, "Flattening with ListNull")
			to, d := tTo.ValueFromList(ctx, types.ListNull(elementType))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags
		}
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Flattening with ListValue", map[string]any{
		logAttrKeySourceSize: vFrom.Len(),
	})
	elements := make([]attr.Value, vFrom.Len())
	for i := range vFrom.Len() {
		value, d := attrValueFromReflectValue(vFrom.Index(i))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		elements[i] = value
	}
	list, d := types.ListValue(elementType, elements)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	to, d := tTo.ValueFromList(ctx, list)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(to))
	return diags
}

// sliceOfPrimitiveToSet copies an AWS API slice of primitive (or pointer to primitive) value to a compatible Plugin Framework Set value.
func (flattener autoFlattener) sliceOfPrimitiveToSet(ctx context.Context, vFrom reflect.Value, tTo basetypes.SetTypable, vTo reflect.Value, elementType attr.Type, attrValueFromReflectValue attrValueFromReflectValueFunc, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if fieldOpts.legacy {
		tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
		if vFrom.IsNil() {
			set, d := types.SetValue(elementType, []attr.Value{})
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			to, d := tTo.ValueFromSet(ctx, set)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags
		}
	} else {
		if vFrom.IsNil() {
			// If omitempty is set, return null
			if fieldOpts.omitempty {
				tflog.SubsystemTrace(ctx, subsystemName, "Flattening with SetNull (omitempty)")
				to, d := tTo.ValueFromSet(ctx, types.SetNull(elementType))
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}

			// If legacy mode, return empty set
			if fieldOpts.legacy {
				tflog.SubsystemTrace(ctx, subsystemName, "Flattening with SetValue (empty for nil in legacy mode)")
				set, d := types.SetValue(elementType, []attr.Value{})
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}
				to, d := tTo.ValueFromSet(ctx, set)
				diags.Append(d...)
				if diags.HasError() {
					return diags
				}

				vTo.Set(reflect.ValueOf(to))
				return diags
			}

			// Default: return null set
			tflog.SubsystemTrace(ctx, subsystemName, "Flattening with SetNull")
			to, d := tTo.ValueFromSet(ctx, types.SetNull(elementType))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(to))
			return diags
		}
	}

	// Check if slice is empty and omitempty is set - return null instead of empty set
	if vFrom.Len() == 0 && fieldOpts.omitempty {
		tflog.SubsystemTrace(ctx, subsystemName, "Flattening with SetValue (null for empty slice with omitempty)")
		to, d := tTo.ValueFromSet(ctx, types.SetNull(elementType))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(to))
		return diags
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Flattening with SetValue", map[string]any{
		logAttrKeySourceSize: vFrom.Len(),
	})
	elements := make([]attr.Value, vFrom.Len())
	for i := range vFrom.Len() {
		value, d := attrValueFromReflectValue(vFrom.Index(i))
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		elements[i] = value
	}
	set, d := types.SetValue(elementType, elements)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	to, d := tTo.ValueFromSet(ctx, set)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(to))
	return diags
}

// sliceOfStructToNestedObjectCollection copies an AWS API []struct value to a compatible Plugin Framework NestedObjectCollectionValue value.
func (flattener autoFlattener) sliceOfStructToNestedObjectCollection(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value, fieldOpts fieldOpts) diag.Diagnostics {
	var diags diag.Diagnostics

	if fieldOpts.legacy {
		tflog.SubsystemDebug(ctx, subsystemName, "Using legacy flattener")
		if vFrom.IsNil() {
			to, d := tTo.NewObjectSlice(ctx, 0, 0)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			val, d := tTo.ValueFromObjectSlice(ctx, to)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(val))
			return diags
		}
	} else {
		if vFrom.IsNil() {
			tflog.SubsystemTrace(ctx, subsystemName, "Flattening with NullValue (for nested object collection)")
			val, d := tTo.NullValue(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}

			vTo.Set(reflect.ValueOf(val))
			return diags
		}
	}

	// Create a new target slice and flatten each element.
	n := vFrom.Len()

	tflog.SubsystemTrace(ctx, subsystemName, "Flattening nested object collection", map[string]any{
		logAttrKeySourceSize: n,
	})

	// DEBUG: Log the source struct fields for the first element
	if n > 0 {
		firstElem := vFrom.Index(0)
		tflog.SubsystemDebug(ctx, subsystemName, "DEBUG: First element of nested object collection")

		// Log exported fields only
		if firstElem.Kind() == reflect.Struct {
			for i := 0; i < firstElem.NumField(); i++ {
				field := firstElem.Type().Field(i)
				if !field.IsExported() {
					continue
				}
				fieldVal := firstElem.Field(i)
				tflog.SubsystemDebug(ctx, subsystemName, "DEBUG: Struct field", map[string]any{
					"field_name": field.Name,
					"field_type": fieldVal.Type().String(),
					"is_nil":     fieldVal.Kind() == reflect.Pointer && fieldVal.IsNil(),
					"is_zero":    fieldVal.IsZero(),
				})
			}
		}
	}

	to, d := tTo.NewObjectSlice(ctx, n, n)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	t := reflect.ValueOf(to)
	for i := range n {
		sourcePath := sourcePath.AtListIndex(i)
		targetPath := targetPath.AtListIndex(i)
		ctx := tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())
		target, d := tTo.NewObjectPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		diags.Append(flattenStruct(ctx, sourcePath, vFrom.Index(i).Interface(), targetPath, target, flattener)...)
		if diags.HasError() {
			return diags
		}

		t.Index(i).Set(reflect.ValueOf(target))
	}

	// Set the target structure as a nested Object.
	val, d := tTo.ValueFromObjectSlice(ctx, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	vTo.Set(reflect.ValueOf(val))
	return diags
}

// xmlWrapperFlatten handles flattening from AWS XML wrapper structs to TF collection types
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
func (flattener *autoFlattener) xmlWrapperFlatten(ctx context.Context, sourcePath path.Path, sourceFieldName string, vFrom reflect.Value, targetPath path.Path, tTo attr.Type, vTo reflect.Value, opts tagOptions) diag.Diagnostics {
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(valueType(vTo)))

	tflog.SubsystemTrace(ctx, subsystemName, "Starting XML wrapper flatten", map[string]any{})

	// Check if target is a NestedObjectCollection (Rule 2 pattern)
	if nestedObjType, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
		tflog.SubsystemTrace(ctx, subsystemName, "Target is NestedObjectCollectionType - checking for Rule 2")

		// Rule 2 detection: check if source AWS struct has more than 2 fields
		// (Items, Quantity, plus additional fields like Enabled)
		sourceStructType := vFrom.Type()
		if sourceStructType.Kind() == reflect.Ptr {
			sourceStructType = sourceStructType.Elem()
		}

		isRule2 := false
		if sourceStructType.Kind() == reflect.Struct {
			// Count fields, excluding noSmithyDocumentSerde
			fieldCount := 0
			for i := 0; i < sourceStructType.NumField(); i++ {
				fieldName := sourceStructType.Field(i).Name
				if fieldName != "noSmithyDocumentSerde" {
					fieldCount++
				}
			}
			isRule2 = fieldCount > 2
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Rule 2 detection result", map[string]any{
			"is_rule2":    isRule2,
			"field_count": sourceStructType.NumField(),
		})

		if isRule2 {
			tflog.SubsystemTrace(ctx, subsystemName, "Using Rule 2 flatten - calling xmlWrapperFlattenRule2")
			return flattener.xmlWrapperFlattenRule2(ctx, vFrom, nestedObjType, vTo, opts)
		}
		tflog.SubsystemTrace(ctx, subsystemName, "NOT Rule 2 - continuing with Rule 1")
	}

	// Rule 1: Flatten Items field directly to collection
	return flattener.xmlWrapperFlattenRule1(ctx, sourcePath, sourceFieldName, vFrom, tTo, vTo)
}

// xmlWrapperFlattenRule1 handles Rule 1: flatten Items field directly to collection
func (flattener *autoFlattener) xmlWrapperFlattenRule1(ctx context.Context, sourcePath path.Path, sourceFieldName string, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	itemsField := vFrom.FieldByName(sourceFieldName)
	if !itemsField.IsValid() {
		tflog.SubsystemError(ctx, subsystemName, "XML wrapper struct missing Items field")
		diags.Append(DiagFlatteningIncompatibleTypes(vFrom.Type(), reflect.TypeOf(vTo.Interface())))
		return diags
	}

	// Determine element type
	var elementType attr.Type = types.StringType // default
	if tToWithElem, ok := tTo.(attr.TypeWithElementType); ok {
		elementType = tToWithElem.ElementType()
		tflog.SubsystemTrace(ctx, subsystemName, "Using target element type", map[string]any{
			"element_type": elementType.String(),
		})
	}

	// Handle different target collection types
	switch tTo := tTo.(type) {
	case basetypes.ListTypable:
		// Items []T -> types.List
		if itemsField.IsNil() {
			tflog.SubsystemTrace(ctx, subsystemName, "Flattening XML wrapper with ListNull")
			to, d := tTo.ValueFromList(ctx, types.ListNull(elementType))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			vTo.Set(reflect.ValueOf(to))
			return diags
		}

		// Convert items slice to list elements
		itemsLen := itemsField.Len()

		// Filter out nil pointers first and collect valid items
		validItems := make([]reflect.Value, 0, itemsLen)
		for i := range itemsLen {
			item := itemsField.Index(i)
			// Skip nil pointers
			if item.Kind() == reflect.Pointer && item.IsNil() {
				continue
			}
			validItems = append(validItems, item)
		}

		elements := make([]attr.Value, len(validItems))

		tflog.SubsystemTrace(ctx, subsystemName, "Converting items to list elements", map[string]any{
			"items_count": itemsLen,
			"valid_count": len(validItems),
		})

		for i, item := range validItems {
			sourcePath := sourcePath.AtListIndex(i)
			ctx := tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
			ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(item.Type()))
			tflog.SubsystemTrace(ctx, subsystemName, "Processing item")

			// Convert each item based on its type
			switch item.Kind() {
			case reflect.Int32:
				// Handle int32 -> Int64 conversion for XML wrapper items
				if val, d := newInt64ValueFromReflectValue(item); d.HasError() {
					diags.Append(d...)
					return diags
				} else {
					elements[i] = val
				}
			case reflect.String:
				// Try to create a value that matches the target element type
				if val, d := flattener.createTargetValue(ctx, types.StringValue(item.String()), elementType); d.HasError() {
					diags.Append(d...)
					return diags
				} else {
					elements[i] = val
				}
			case reflect.Pointer:
				// Handle pointer types like *testEnum (pointer to enum)
				// (nil pointers are already filtered out)
				// Dereference the pointer and get the underlying value
				derefItem := item.Elem()

				// Handle the dereferenced value based on its type
				switch derefItem.Kind() {
				case reflect.String:
					// Handle *string or *testEnum (where testEnum is a string type)
					stringVal := derefItem.String()
					if val, d := flattener.createTargetValue(ctx, types.StringValue(stringVal), elementType); d.HasError() {
						diags.Append(d...)
						return diags
					} else {
						elements[i] = val
					}
				default:
					// Check if the dereferenced type is convertible to string (like custom enums)
					if derefItem.Type().ConvertibleTo(reflect.TypeFor[string]()) {
						stringVal := derefItem.Convert(reflect.TypeFor[string]()).String()
						if val, d := flattener.createTargetValue(ctx, types.StringValue(stringVal), elementType); d.HasError() {
							diags.Append(d...)
							return diags
						} else {
							elements[i] = val
						}
					} else {
						diags.Append(DiagFlatteningIncompatibleTypes(derefItem.Type(), reflect.TypeOf(elementType)))
						return diags
					}
				}
			case reflect.Struct:
				// Handle struct types - convert to ObjectValue
				if objType, ok := elementType.(fwtypes.NestedObjectType); ok {
					objPtr, d := objType.NewObjectPtr(ctx)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}
					diags.Append(autoFlattenConvert(ctx, item.Interface(), objPtr, flattener)...)
					if diags.HasError() {
						return diags
					}
					objVal, d := objType.ValueFromObjectPtr(ctx, objPtr)
					diags.Append(d...)
					if diags.HasError() {
						return diags
					}
					elements[i] = objVal
				} else {
					diags.Append(DiagFlatteningIncompatibleTypes(item.Type(), reflect.TypeOf(elementType)))
					return diags
				}
			default:
				// Check for custom string types (like enums)
				if item.Type().ConvertibleTo(reflect.TypeFor[string]()) {
					// Convert custom string type (like testEnum/Method) to string
					stringVal := item.Convert(reflect.TypeFor[string]()).String()

					// Try to create a value that matches the target element type
					if val, d := flattener.createTargetValue(ctx, types.StringValue(stringVal), elementType); d.HasError() {
						diags.Append(d...)
						return diags
					} else {
						elements[i] = val
					}
				} else {
					diags.Append(DiagFlatteningIncompatibleTypes(item.Type(), reflect.TypeOf(elementType)))
					return diags
				}
			}
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Creating list value", map[string]any{
			"element_count": len(elements),
		})

		list, d := types.ListValue(elementType, elements)
		diags.Append(d...)
		if diags.HasError() {
			tflog.SubsystemError(ctx, subsystemName, "Error creating list value", map[string]any{
				"error": d.Errors(),
			})
			return diags
		}

		to, d := tTo.ValueFromList(ctx, list)
		diags.Append(d...)
		if diags.HasError() {
			tflog.SubsystemError(ctx, subsystemName, "Error converting to target list", map[string]any{
				"error": d.Errors(),
			})
			return diags
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Setting target list value")
		vTo.Set(reflect.ValueOf(to))
		return diags

	case basetypes.SetTypable:
		// Items []T -> types.Set
		if itemsField.IsNil() {
			tflog.SubsystemTrace(ctx, subsystemName, "Flattening XML wrapper with SetNull")
			to, d := tTo.ValueFromSet(ctx, types.SetNull(elementType))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			vTo.Set(reflect.ValueOf(to))
			return diags
		}

		// Convert items slice to set elements
		itemsLen := itemsField.Len()

		// Filter out nil pointers first and collect valid items
		validItems := make([]reflect.Value, 0, itemsLen)
		for i := range itemsLen {
			item := itemsField.Index(i)
			// Skip nil pointers
			if item.Kind() == reflect.Pointer && item.IsNil() {
				continue
			}
			validItems = append(validItems, item)
		}

		elements := make([]attr.Value, len(validItems))

		tflog.SubsystemTrace(ctx, subsystemName, "Converting items to set elements", map[string]any{
			"items_count": itemsLen,
			"valid_count": len(validItems),
		})

		for i, item := range validItems {
			sourcePath := sourcePath.AtListIndex(i)
			ctx := tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
			ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(item.Type()))
			tflog.SubsystemTrace(ctx, subsystemName, "Processing item")

			// Convert each item based on its type
			switch item.Kind() {
			case reflect.Int32:
				// Handle int32 -> Int64 conversion for XML wrapper items
				if val, d := newInt64ValueFromReflectValue(item); d.HasError() {
					diags.Append(d...)
					return diags
				} else {
					elements[i] = val
				}
			case reflect.String:
				// Try to create a value that matches the target element type
				if val, d := flattener.createTargetValue(ctx, types.StringValue(item.String()), elementType); d.HasError() {
					diags.Append(d...)
					return diags
				} else {
					elements[i] = val
				}
			case reflect.Pointer:
				// Handle pointer types like *testEnum (pointer to enum)
				// (nil pointers are already filtered out)
				// Dereference the pointer and get the underlying value
				derefItem := item.Elem()

				// Handle the dereferenced value based on its type
				switch derefItem.Kind() {
				case reflect.String:
					// Handle *string or *testEnum (where testEnum is a string type)
					stringVal := derefItem.String()
					if val, d := flattener.createTargetValue(ctx, types.StringValue(stringVal), elementType); d.HasError() {
						diags.Append(d...)
						return diags
					} else {
						elements[i] = val
					}
				default:
					// Check if the dereferenced type is convertible to string (like custom enums)
					if derefItem.Type().ConvertibleTo(reflect.TypeFor[string]()) {
						stringVal := derefItem.Convert(reflect.TypeFor[string]()).String()
						if val, d := flattener.createTargetValue(ctx, types.StringValue(stringVal), elementType); d.HasError() {
							diags.Append(d...)
							return diags
						} else {
							elements[i] = val
						}
					} else {
						diags.Append(DiagFlatteningIncompatibleTypes(derefItem.Type(), reflect.TypeOf(elementType)))
						return diags
					}
				}
			case reflect.Struct:
				// Handle complex struct types by converting to the target element type
				if elemTyper, ok := tTo.(attr.TypeWithElementType); ok && elemTyper.ElementType() != nil {
					elemType := elemTyper.ElementType()

					// Create a new instance of the target element type
					targetValue := reflect.New(reflect.TypeOf(elemType.ValueType(ctx))).Elem()

					// Use AutoFlex to flatten the struct to the target type
					diags.Append(flattener.convert(ctx, path.Empty(), item, path.Empty(), targetValue, fieldOpts{})...)
					if diags.HasError() {
						return diags
					}

					// The converted value should implement attr.Value
					if attrVal, ok := targetValue.Interface().(attr.Value); ok {
						elements[i] = attrVal
					} else {
						diags.Append(DiagFlatteningIncompatibleTypes(item.Type(), reflect.TypeOf(targetValue.Interface())))
						return diags
					}
				} else {
					diags.Append(DiagFlatteningIncompatibleTypes(item.Type(), reflect.TypeOf(elementType)))
					return diags
				}
			default:
				// Check for custom string types (like enums)
				if item.Type().ConvertibleTo(reflect.TypeFor[string]()) {
					// Convert custom string type (like testEnum/Method) to string
					stringVal := item.Convert(reflect.TypeFor[string]()).String()

					// Try to create a value that matches the target element type
					if val, d := flattener.createTargetValue(ctx, types.StringValue(stringVal), elementType); d.HasError() {
						diags.Append(d...)
						return diags
					} else {
						elements[i] = val
					}
				} else {
					// For other complex types, handle conversion if needed
					diags.Append(DiagFlatteningIncompatibleTypes(item.Type(), reflect.TypeOf(elementType)))
					return diags
				}
			}
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Creating set value", map[string]any{
			"element_count": len(elements),
		})

		set, d := types.SetValue(elementType, elements)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		to, d := tTo.ValueFromSet(ctx, set)
		diags.Append(d...)
		if diags.HasError() {
			return diags
		}

		vTo.Set(reflect.ValueOf(to))
		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "Unsupported target type for XML wrapper flattening", map[string]any{
		"target_type": tTo,
	})
	diags.Append(DiagFlatteningIncompatibleTypes(vFrom.Type(), reflect.TypeOf(vTo.Interface())))
	return diags
}

// flattenStruct traverses struct `from`, calling `flexer` for each exported field.
// handleXMLWrapperRule1 handles Rule 1: flatten entire source struct to target collection field with xmlwrapper tag
func handleXMLWrapperRule1(ctx context.Context, sourcePath path.Path, valFrom, valTo reflect.Value, typeFrom, typeTo reflect.Type, targetPath path.Path, flexer autoFlexer) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Skip if target is a struct - that's Rule 2 where we map fields individually
	if typeTo.Kind() == reflect.Struct {
		return false, diags
	}

	for toField := range tfreflect.ExportedStructFields(typeTo) {
		toFieldName := toField.Name
		_, toOpts := autoflexTags(toField)
		if wrapperField := toOpts.XMLWrapperField(); wrapperField != "" {
			toFieldVal := valTo.FieldByIndex(toField.Index)
			if !toFieldVal.CanSet() {
				// TODO:
				// // Corresponding field value can't be changed.
				// tflog.SubsystemDebug(ctx, subsystemName, "Target field cannot be set", map[string]any{
				// 	logAttrKeySourceFieldname: sourceFieldName,
				// 	logAttrKeyTargetFieldname: toFieldName,
				// })
				continue
			}

			tflog.SubsystemTrace(ctx, subsystemName, "Converting entire XML wrapper struct to collection field (Rule 1)", map[string]any{
				logAttrKeySourceType:      typeFrom.String(),
				logAttrKeyTargetFieldname: toFieldName,
				"wrapper_field":           wrapperField,
			})

			attrVal, ok := toFieldVal.Interface().(attr.Value)
			if !ok {
				tflog.SubsystemError(ctx, subsystemName, "Target field does not implement attr.Value")
				diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(toFieldVal.Interface())))
				return true, diags
			}

			if f, ok := flexer.(*autoFlattener); ok {
				diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, "XXX", valFrom, targetPath, attrVal.Type(ctx), toFieldVal, toOpts)...)
			} else {
				diags.Append(DiagFlatteningIncompatibleTypes(valFrom.Type(), reflect.TypeOf(toFieldVal.Interface())))
			}
			return true, diags
		}
	}

	return false, diags
}

func flattenStruct(ctx context.Context, sourcePath path.Path, from any, targetPath path.Path, to any, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	ctx, valFrom, valTo, d := autoFlexValues(ctx, from, to)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	if toFlattener, ok := to.(Flattener); ok {
		tflog.SubsystemInfo(ctx, subsystemName, "Target implements flex.Flattener")
		diags.Append(flattenFlattener(ctx, valFrom, toFlattener)...)
		return diags
	}

	typeFrom := valFrom.Type()
	typeTo := valTo.Type()

	// Special handling: Check if target has xmlwrapper tag (Rule 1)
	if handled, d := handleXMLWrapperRule1(ctx, sourcePath, valFrom, valTo, typeFrom, typeTo, targetPath, flexer); handled {
		diags.Append(d...)
		return diags
	}

	// Handle XML wrapper split patterns where complex source fields
	// need to be split into multiple target collection fields
	processedFields := make(map[string]bool)
	diags.Append(flexer.handleXMLWrapperCollapse(ctx, sourcePath, valFrom, targetPath, valTo, typeFrom, typeTo, processedFields)...)
	if diags.HasError() {
		return diags
	}

	for fromField := range flattenSourceFields(ctx, typeFrom, flexer.getOptions()) {
		fromFieldName := fromField.Name

		// Skip fields that were already processed by XML wrapper split
		if processedFields[fromFieldName] {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping field already processed by XML wrapper split", map[string]any{
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
		toNameOverride, toFieldOpts := autoflexTags(toField)
		toFieldVal := valTo.FieldByIndex(toField.Index)
		if toNameOverride == "-" {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored target field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
				logAttrKeyTargetFieldname: toFieldName,
			})
			continue
		}
		if toFieldOpts.NoFlatten() {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping noflatten target field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
				logAttrKeyTargetFieldname: toFieldName,
			})
			continue
		}
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

		// Check if target has wrapper tag and source is an XML wrapper struct
		if wrapperField := toFieldOpts.XMLWrapperField(); wrapperField != "" {
			fromFieldVal := valFrom.FieldByIndex(fromField.Index)

			// Handle direct XML wrapper struct
			if potentialXMLWrapperStruct(fromFieldVal.Type()) {
				tflog.SubsystemTrace(ctx, subsystemName, "Converting XML wrapper struct to collection", map[string]any{
					logAttrKeySourceFieldname: fromFieldName,
					logAttrKeyTargetFieldname: toFieldName,
					"wrapper_field":           wrapperField,
				})

				valTo, ok := toFieldVal.Interface().(attr.Value)
				if !ok {
					tflog.SubsystemError(ctx, subsystemName, "Target field does not implement attr.Value")
					diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(toFieldVal.Interface())))
					break
				}

				if f, ok := flexer.(*autoFlattener); ok {
					diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, wrapperField, fromFieldVal, targetPath, valTo.Type(ctx), toFieldVal, toFieldOpts)...)
				} else {
					diags.Append(DiagFlatteningIncompatibleTypes(fromFieldVal.Type(), reflect.TypeOf(toFieldVal.Interface())))
				}
				if diags.HasError() {
					break
				}
				continue
			}

			// Handle pointer to XML wrapper struct
			if fromFieldVal.Kind() == reflect.Pointer && !fromFieldVal.IsNil() && potentialXMLWrapperStruct(fromFieldVal.Type().Elem()) {
				tflog.SubsystemTrace(ctx, subsystemName, "Converting pointer to XML wrapper struct to collection via wrapper tag", map[string]any{
					logAttrKeySourceFieldname: fromFieldName,
					logAttrKeyTargetFieldname: toFieldName,
					"wrapper_field":           wrapperField,
				})

				valTo, ok := toFieldVal.Interface().(attr.Value)
				if !ok {
					tflog.SubsystemError(ctx, subsystemName, "Target field does not implement attr.Value")
					diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(toFieldVal.Interface())))
					break
				}

				// Try both value and pointer type assertions
				if f, ok := flexer.(autoFlattener); ok {
					diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, wrapperField, fromFieldVal.Elem(), targetPath, valTo.Type(ctx), toFieldVal, toFieldOpts)...)
				} else if f, ok := flexer.(*autoFlattener); ok {
					diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, wrapperField, fromFieldVal.Elem(), targetPath, valTo.Type(ctx), toFieldVal, toFieldOpts)...)
				} else {
					tflog.SubsystemError(ctx, subsystemName, "Type assertion to autoFlattener failed for wrapper tag")
					diags.Append(DiagFlatteningIncompatibleTypes(fromFieldVal.Type(), reflect.TypeOf(toFieldVal.Interface())))
				}
				if diags.HasError() {
					break
				}
				continue
			}

			// Handle nil pointer to XML wrapper struct
			if fromFieldVal.Kind() == reflect.Pointer && fromFieldVal.IsNil() && potentialXMLWrapperStruct(fromFieldVal.Type().Elem()) {
				tflog.SubsystemTrace(ctx, subsystemName, "Converting nil pointer to XML wrapper struct to null collection", map[string]any{
					logAttrKeySourceFieldname: fromFieldName,
					logAttrKeyTargetFieldname: toFieldName,
					"wrapper_field":           wrapperField, // TODO: rename to wrapperFieldName
				})

				valTo, ok := toFieldVal.Interface().(attr.Value)
				if !ok {
					tflog.SubsystemError(ctx, subsystemName, "Target field does not implement attr.Value")
					diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(toFieldVal.Interface())))
					break
				}

				// For nil XML wrapper pointers, set target to null collection
				switch tTo := valTo.Type(ctx).(type) {
				case basetypes.ListTypable:
					var elementType attr.Type = types.StringType // default
					if tToWithElem, ok := tTo.(attr.TypeWithElementType); ok {
						elementType = tToWithElem.ElementType()
					}
					nullList, d := tTo.ValueFromList(ctx, types.ListNull(elementType))
					diags.Append(d...)
					if !diags.HasError() {
						toFieldVal.Set(reflect.ValueOf(nullList))
					}
				case basetypes.SetTypable:
					var elementType attr.Type = types.StringType // default
					if tToWithElem, ok := tTo.(attr.TypeWithElementType); ok {
						elementType = tToWithElem.ElementType()
					}
					nullSet, d := tTo.ValueFromSet(ctx, types.SetNull(elementType))
					diags.Append(d...)
					if !diags.HasError() {
						toFieldVal.Set(reflect.ValueOf(nullSet))
					}
				default:
					diags.Append(DiagFlatteningIncompatibleTypes(fromFieldVal.Type(), reflect.TypeOf(toFieldVal.Interface())))
				}
				if diags.HasError() {
					break
				}
				continue
			}
		}

		// Automatic XML wrapper detection (without explicit wrapper tags)
		fromFieldVal := valFrom.FieldByIndex(fromField.Index)
		_, toOpts := autoflexTags(toField)

		// Handle pointer to XML wrapper struct
		// Only auto-detect if target has explicit wrapper tag
		if fromFieldVal.Kind() == reflect.Pointer {
			if toOpts.XMLWrapperField() != "" && !fromFieldVal.IsNil() && potentialXMLWrapperStruct(fromFieldVal.Type().Elem()) {
				// Check if target is a collection type (Set, List, or NestedObjectCollection)
				if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
					targetType := valTo.Type(ctx)
					switch tt := targetType.(type) {
					case basetypes.SetTypable, basetypes.ListTypable, fwtypes.NestedObjectCollectionType:
						// Check if wrapper Items field is empty/nil
						// If so, set null and skip processing
						wrapperVal := fromFieldVal.Elem()
						wrapperFieldName := getXMLWrapperSliceFieldName(wrapperVal.Type())
						itemsField := wrapperVal.FieldByName(wrapperFieldName)

						if itemsField.IsValid() {
							itemsEmpty := itemsField.IsNil() || (itemsField.Kind() == reflect.Slice && itemsField.Len() == 0)

							if itemsEmpty {
								tflog.SubsystemTrace(ctx, subsystemName, "XML wrapper Items is empty, setting null", map[string]any{
									logAttrKeySourceFieldname: fromFieldName,
									logAttrKeyTargetFieldname: toFieldName,
								})

								// Call NullValue on NestedObjectCollectionType
								if nestedObjType, ok := tt.(fwtypes.NestedObjectCollectionType); ok {
									nullVal, d := nestedObjType.NullValue(ctx)
									diags.Append(d...)
									if !diags.HasError() {
										toFieldVal.Set(reflect.ValueOf(nullVal))
									}
									continue
								}
							}
						}

						tflog.SubsystemTrace(ctx, subsystemName, "Auto-converting pointer to XML wrapper struct to collection", map[string]any{
							logAttrKeySourceFieldname: fromFieldName,
							logAttrKeyTargetFieldname: toFieldName,
							"wrapper_field":           wrapperFieldName,
						})

						// Try both value and pointer type assertions
						if f, ok := flexer.(autoFlattener); ok {
							diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, wrapperFieldName, fromFieldVal.Elem(), targetPath, targetType, toFieldVal, toFieldOpts)...)
							if diags.HasError() {
								break
							}
							continue // Successfully handled, skip normal processing
						} else if f, ok := flexer.(*autoFlattener); ok {
							diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, wrapperFieldName, fromFieldVal.Elem(), targetPath, targetType, toFieldVal, toFieldOpts)...)
							if diags.HasError() {
								break
							}
							continue // Successfully handled, skip normal processing
						}
						// If flexer is not autoFlattener, fall through to normal field matching
					}
				}
			} else if toOpts.XMLWrapperField() != "" && fromFieldVal.IsNil() && potentialXMLWrapperStruct(fromFieldVal.Type().Elem()) {
				// Handle nil pointer to XML wrapper struct - should result in null collection
				if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
					switch tTo := valTo.Type(ctx).(type) {
					case basetypes.SetTypable:
						tflog.SubsystemTrace(ctx, subsystemName, "Auto-converting nil pointer to XML wrapper struct to null set", map[string]any{
							logAttrKeySourceFieldname: fromFieldName,
							logAttrKeyTargetFieldname: toFieldName,
						})
						var elemType attr.Type = types.StringType
						if tToWithElem, ok := tTo.(attr.TypeWithElementType); ok {
							elemType = tToWithElem.ElementType()
						}
						nullSet, d := tTo.ValueFromSet(ctx, types.SetNull(elemType))
						diags.Append(d...)
						if !diags.HasError() {
							toFieldVal.Set(reflect.ValueOf(nullSet))
						}
						if diags.HasError() {
							break
						}
						continue
					case basetypes.ListTypable:
						tflog.SubsystemTrace(ctx, subsystemName, "Auto-converting nil pointer to XML wrapper struct to null list", map[string]any{
							logAttrKeySourceFieldname: fromFieldName,
							logAttrKeyTargetFieldname: toFieldName,
						})
						var elemType attr.Type = types.StringType
						if tToWithElem, ok := tTo.(attr.TypeWithElementType); ok {
							elemType = tToWithElem.ElementType()
						}
						nullList, d := tTo.ValueFromList(ctx, types.ListNull(elemType))
						diags.Append(d...)
						if !diags.HasError() {
							toFieldVal.Set(reflect.ValueOf(nullList))
						}
						if diags.HasError() {
							break
						}
						continue
					}
				}
			}
		} else if toOpts.XMLWrapperField() != "" && potentialXMLWrapperStruct(fromFieldVal.Type()) {
			// Handle direct XML wrapper struct (non-pointer)
			if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
				switch valTo.Type(ctx).(type) {
				case basetypes.SetTypable, basetypes.ListTypable:
					wrapperField := toOpts.XMLWrapperField()
					tflog.SubsystemTrace(ctx, subsystemName, "Auto-converting XML wrapper struct to collection", map[string]any{
						logAttrKeySourceFieldname: fromFieldName,
						logAttrKeyTargetFieldname: toFieldName,
						"wrapper_field":           wrapperField,
					})

					if f, ok := flexer.(*autoFlattener); ok {
						diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, wrapperField, fromFieldVal, targetPath, valTo.Type(ctx), toFieldVal, toFieldOpts)...)
					} else {
						diags.Append(DiagFlatteningIncompatibleTypes(fromFieldVal.Type(), reflect.TypeOf(toFieldVal.Interface())))
					}
					if diags.HasError() {
						break
					}
					continue
				}
			}
		}

		// Check for Rule 2: Source is XML wrapper, target is NestedObjectCollection
		// Only apply if target field has xmlwrapper tag
		fromFieldVal2 := valFrom.FieldByIndex(fromField.Index)
		if toFieldOpts.XMLWrapperField() != "" && fromFieldVal2.Kind() == reflect.Pointer && !fromFieldVal2.IsNil() && potentialXMLWrapperStruct(fromFieldVal2.Type().Elem()) {
			if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
				if nestedObjType, ok := valTo.Type(ctx).(fwtypes.NestedObjectCollectionType); ok {
					sourceStructType := fromFieldVal2.Type().Elem()

					// Check if source is Rule 2 (more than 2 fields: Items, Quantity, + additional fields)
					isRule2 := sourceStructType.NumField() > 2

					if isRule2 {
						tflog.SubsystemTrace(ctx, subsystemName, "Detected nested Rule 2 XML wrapper to NestedObjectCollection", map[string]any{
							logAttrKeySourceFieldname: fromFieldName,
							logAttrKeyTargetFieldname: toFieldName,
							"field_count":             sourceStructType.NumField(),
						})

						if f, ok := flexer.(*autoFlattener); ok {
							diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, fromFieldName, fromFieldVal2.Elem(), targetPath, valTo.Type(ctx), toFieldVal, toFieldOpts)...)
							if diags.HasError() {
								break
							}
							continue
						}
					}

					// Also check if nested model has xmlwrapper field (original logic)
					samplePtr, d := nestedObjType.NewObjectPtr(ctx)
					diags.Append(d...)
					if !diags.HasError() {
						sampleValue := reflect.ValueOf(samplePtr).Elem()
						hasXMLWrapperTag := false
						for i := 0; i < sampleValue.NumField(); i++ {
							field := sampleValue.Type().Field(i)
							_, fieldOpts := autoflexTags(field)
							if fieldOpts.XMLWrapperField() != "" {
								hasXMLWrapperTag = true
								break
							}
						}

						if hasXMLWrapperTag {
							tflog.SubsystemTrace(ctx, subsystemName, "Detected Rule 2: XML wrapper to NestedObjectCollection with nested xmlwrapper field", map[string]any{
								logAttrKeySourceFieldname: fromFieldName,
								logAttrKeyTargetFieldname: toFieldName,
							})

							_ = getXMLWrapperSliceFieldName(fromFieldVal2.Type().Elem())
							if f, ok := flexer.(*autoFlattener); ok {
								diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, fromFieldName, fromFieldVal2.Elem(), targetPath, valTo.Type(ctx), toFieldVal, toFieldOpts)...)
								if diags.HasError() {
									break
								}
								continue
							}
						}
					}
				}
			}
		}

		opts := fieldOpts{
			legacy:          toFieldOpts.Legacy(),
			omitempty:       toFieldOpts.OmitEmpty(),
			xmlWrapper:      toFieldOpts.XMLWrapperField() != "",
			xmlWrapperField: toFieldOpts.XMLWrapperField(),
		}

		diags.Append(flexer.convert(ctx, sourcePath.AtName(fromFieldName), valFrom.FieldByIndex(fromField.Index), targetPath.AtName(toFieldName), toFieldVal, opts)...)
		if diags.HasError() {
			break
		}
	}

	return diags
}

func flattenSourceFields(ctx context.Context, typ reflect.Type, opts AutoFlexOptions) iter.Seq[reflect.StructField] {
	return func(yield func(reflect.StructField) bool) {
		for field := range tfreflect.ExportedStructFields(typ) {
			fieldName := field.Name
			if opts.isIgnoredField(fieldName) {
				tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored source field", map[string]any{
					logAttrKeySourceFieldname: fieldName,
				})
				continue
			}

			if !yield(field) {
				return
			}
		}
	}
}

// setMapBlockKey takes a struct and assigns the value of the `key`
func setMapBlockKey(ctx context.Context, to any, key reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo := reflect.ValueOf(to)
	if kind := valTo.Kind(); kind == reflect.Pointer {
		valTo = valTo.Elem()
	}

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(valueType(valTo)))

	if valTo.Kind() != reflect.Struct {
		diags.AddError("AutoFlEx", fmt.Sprintf("wrong type (%T), expected struct", valTo))
		return diags
	}

	for field := range tfreflect.ExportedStructFields(valTo.Type()) {
		if field.Name != mapBlockKeyFieldName {
			continue
		}

		fieldVal := valTo.FieldByIndex(field.Index)           // vTo
		fieldAttrVal, ok := fieldVal.Interface().(attr.Value) // valTo
		if !ok {
			tflog.SubsystemError(ctx, subsystemName, "Target does not implement attr.Value")
			diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(fieldVal.Interface())))
			return diags
		}
		tTo := fieldAttrVal.Type(ctx)

		switch tTo := tTo.(type) {
		case basetypes.StringTypable:
			v, d := tTo.ValueFromString(ctx, types.StringValue(key.String()))
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			fieldVal.Set(reflect.ValueOf(v))
		}

		return diags
	}

	tflog.SubsystemError(ctx, subsystemName, "Target has no map block key")
	diags.Append(diagFlatteningNoMapBlockKey(valTo.Type()))

	return diags
}

type attrValueFromReflectValueFunc func(reflect.Value) (attr.Value, diag.Diagnostics)

func newInt64ValueFromReflectValue(v reflect.Value) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	return types.Int64Value(v.Int()), diags
}

func newInt64ValueFromReflectPointerValue(v reflect.Value) (attr.Value, diag.Diagnostics) {
	if v.IsNil() {
		var diags diag.Diagnostics
		return types.Int64Null(), diags
	}

	return newInt64ValueFromReflectValue(v.Elem())
}

func newStringValueFromReflectValue(v reflect.Value) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	return types.StringValue(v.String()), diags
}

func newStringValueFromReflectPointerValue(v reflect.Value) (attr.Value, diag.Diagnostics) {
	if v.IsNil() {
		var diags diag.Diagnostics
		return types.StringNull(), diags
	}

	return newStringValueFromReflectValue(v.Elem())
}

func flattenFlattener(ctx context.Context, fromVal reflect.Value, toFlattener Flattener) diag.Diagnostics {
	var diags diag.Diagnostics

	valTo := reflect.ValueOf(toFlattener)

	diags.Append(flattenPrePopulate(ctx, valTo)...)
	if diags.HasError() {
		return diags
	}

	from := fromVal.Interface()

	diags.Append(toFlattener.Flatten(ctx, from)...)

	return diags
}

func flattenPrePopulate(ctx context.Context, toVal reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if toVal.Kind() == reflect.Pointer {
		toVal = toVal.Elem()
	}

	for field := range tfreflect.ExportedStructFields(toVal.Type()) {
		fieldVal := toVal.FieldByIndex(field.Index)
		if !fieldVal.CanSet() {
			diags.AddError(
				"Incompatible Types",
				"An unexpected error occurred while flattening configuration. "+
					"This is always an error in the provider. "+
					"Please report the following to the provider developer:\n\n"+
					fmt.Sprintf("Field %q in %q is not settable.", field.Name, fullTypeName(fieldVal.Type())),
			)
		}

		fieldTo, ok := fieldVal.Interface().(attr.Value)
		if !ok {
			continue // Skip non-attr.Type fields.
		}
		tTo := fieldTo.Type(ctx)
		switch tTo := tTo.(type) {
		// The zero values of primitive types are Null.

		// Aggregate types.
		case fwtypes.NestedObjectType:
			v, d := tTo.NullValue(ctx)
			diags.Append(d...)
			if diags.HasError() {
				return diags
			}
			fieldVal.Set(reflect.ValueOf(v))
		}
	}

	return diags
}

func diagFlatteningSourceIsNil(sourceType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Source of type %q is nil.", fullTypeName(sourceType)),
	)
}

func diagFlatteningTargetDoesNotImplementAttrValue(targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Target type %q does not implement attr.Value", fullTypeName(targetType)),
	)
}

// Future error diagnostic for interfaceToNestedObject
// Currently just logging an error because some existing resource types may have this error and
// returning an error diagnostic would be a breaking change.
// func diagFlatteningTargetDoesNotImplementFlexFlattener(targetType reflect.Type) diag.ErrorDiagnostic {
// 	return diag.NewErrorDiagnostic(
// 		"Incompatible Types",
// 		"An unexpected error occurred while flattening configuration. "+
// 			"This is always an error in the provider. "+
// 			"Please report the following to the provider developer:\n\n"+
// 			fmt.Sprintf("Target type %q does not implement flex.Flattener", fullTypeName(targetType)),
// 	)
// }

func diagFlatteningNoMapBlockKey(sourceType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Target type %q does not contain field %q", fullTypeName(sourceType), mapBlockKeyFieldName),
	)
}

func diagFlatteningMarshalSmithyDocument(sourceType reflect.Type, err error) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Marshalling JSON document of type %q failed: %s", fullTypeName(sourceType), err.Error()),
	)
}

func DiagFlatteningIncompatibleTypes(sourceType, targetType reflect.Type) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic(
		"Incompatible Types",
		"An unexpected error occurred while flattening configuration. "+
			"This is always an error in the provider. "+
			"Please report the following to the provider developer:\n\n"+
			fmt.Sprintf("Source type %q cannot be flattened to target type %q.", fullTypeName(sourceType), fullTypeName(targetType)),
	)
}

// handleDirectXMLWrapperStruct handles direct XML wrapper struct to target with xmlwrapper tags
func handleDirectXMLWrapperStruct(ctx context.Context, sourcePath path.Path, sourceFieldName string, valFrom, valTo reflect.Value, typeFrom, typeTo reflect.Type, targetPath path.Path, targetFieldName string, flexer autoFlexer) diag.Diagnostics {
	var diags diag.Diagnostics

	sourceItemsField := valFrom.FieldByName(sourceFieldName)
	if !sourceItemsField.IsValid() {
		tflog.SubsystemError(ctx, subsystemName, "Source Items field not valid", map[string]any{
			logAttrKeySourceFieldname: sourceFieldName,
		})
		return diags
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Processing direct XML wrapper", map[string]any{
		logAttrKeySourceFieldname: sourceFieldName,
		logAttrKeyTargetFieldname: targetFieldName,
	})

	// Find target fields with matching xmlwrapper tags and map the source Items field to them
	for toField := range tfreflect.ExportedStructFields(typeTo) {
		toFieldName := toField.Name
		_, toOpts := autoflexTags(toField)

		tflog.SubsystemTrace(ctx, subsystemName, "Checking target field", map[string]any{
			logAttrKeySourceFieldname: toOpts.XMLWrapperField(),
			logAttrKeyTargetFieldname: toFieldName,
		})

		// Check if this target field expects the wrapper field from source
		if toOpts.XMLWrapperField() == sourceFieldName {
			toFieldVal := valTo.FieldByName(toFieldName)
			if !toFieldVal.IsValid() || !toFieldVal.CanSet() {
				tflog.SubsystemError(ctx, subsystemName, "Target field not valid or settable", map[string]any{
					logAttrKeyTargetFieldname: toFieldName,
				})
				continue
			}

			sourcePath = sourcePath.AtName(sourceFieldName)

			tflog.SubsystemTrace(ctx, subsystemName, "Found matching xmlwrapper field", map[string]any{
				logAttrKeySourceFieldname: toOpts.XMLWrapperField(),
				logAttrKeyTargetFieldname: toFieldName,
			})

			// Get the target field as attr.Value for XML wrapper flattening
			if toAttr, ok := toFieldVal.Interface().(attr.Value); ok {
				if f, ok := flexer.(*autoFlattener); ok {
					tflog.SubsystemTrace(ctx, subsystemName, "Calling xmlWrapperFlatten", map[string]any{
						logAttrKeySourceFieldname: toOpts.XMLWrapperField(),
						logAttrKeyTargetFieldname: toFieldName,
					})

					targetPath = targetPath.AtName(toFieldName)
					typ := sourceItemsField.Type()
					ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(typ))
					// Use XML wrapper flattening to convert the source Items field to the target collection
					diags.Append(f.xmlWrapperFlatten(ctx, sourcePath, sourceFieldName, valFrom, targetPath, toAttr.Type(ctx), toFieldVal, toOpts)...)
				} else {
					tflog.SubsystemError(ctx, subsystemName, "Flexer is not autoFlattener")
					diags.Append(DiagFlatteningIncompatibleTypes(typeFrom, toField.Type))
				}
			} else {
				tflog.SubsystemError(ctx, subsystemName, "Target field does not implement attr.Value")
			}
		}
	}

	return diags
}

// This takes complex AWS structures with XML wrapper patterns and splits them into multiple TF fields.
func (flattener autoFlattener) handleXMLWrapperCollapse(ctx context.Context, sourcePath path.Path, valFrom reflect.Value, targetPath path.Path, valTo reflect.Value, typeFrom, typeTo reflect.Type, processedFields map[string]bool) diag.Diagnostics {
	var diags diag.Diagnostics

	// Look for source fields that are complex XML wrapper structures that should be split
	for i := 0; i < typeFrom.NumField(); i++ {
		fromField := typeFrom.Field(i)
		fromFieldName := fromField.Name
		fromFieldType := fromField.Type
		fromFieldVal := valFrom.Field(i)

		// Skip already processed fields
		if processedFields[fromFieldName] {
			continue
		}

		// Check if this is a pointer to a struct or direct struct that could be split
		var sourceStructType reflect.Type
		var sourceStructVal reflect.Value
		isNil := false

		if fromFieldType.Kind() == reflect.Pointer && fromFieldType.Elem().Kind() == reflect.Struct {
			if fromFieldVal.IsNil() {
				isNil = true
				sourceStructType = fromFieldType.Elem()
				sourceStructVal = reflect.Zero(sourceStructType)
			} else {
				sourceStructType = fromFieldType.Elem()
				sourceStructVal = fromFieldVal.Elem()
			}
		} else if fromFieldType.Kind() == reflect.Struct {
			sourceStructType = fromFieldType
			sourceStructVal = fromFieldVal
		} else {
			continue
		}

		// Check if this source struct should be split into multiple target fields
		if !flattener.isXMLWrapperSplitSource(sourceStructType) {
			continue
		}

		// Before splitting, check if there's a direct field match in the target
		// If the target has a field with the same name that can accept this XML wrapper,
		// skip the split and let normal field matching handle it
		targetField, ok := (&fuzzyFieldFinder{}).findField(ctx, fromFieldName, typeFrom, typeTo, flattener)
		if !ok {
			// Corresponding field not found in target.
			tflog.SubsystemDebug(ctx, subsystemName, "No corresponding target field", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
			})

			// Mark the source field as processed
			processedFields[fromFieldName] = true
			continue
		}

		if targetFieldVal := valTo.FieldByIndex(targetField.Index); targetFieldVal.CanSet() {
			// Check if target field is a NestedObjectCollection that can accept this wrapper
			if targetAttr, ok := targetFieldVal.Interface().(attr.Value); ok {
				if _, isNestedObjCollection := targetAttr.Type(ctx).(fwtypes.NestedObjectCollectionType); isNestedObjCollection {
					// Skip split - let normal field matching handle this as a direct wrapper-to-wrapper mapping
					tflog.SubsystemTrace(ctx, subsystemName, "Skipping XML wrapper split - direct field match found", map[string]any{
						logAttrKeySourceFieldname: fromFieldName,
						logAttrKeyTargetFieldname: targetField.Name,
					})
					continue
				}
			}
		} else {
			// Corresponding field value can't be changed.
			tflog.SubsystemDebug(ctx, subsystemName, "Target field cannot be set", map[string]any{
				logAttrKeySourceFieldname: fromFieldName,
				logAttrKeyTargetFieldname: targetField.Name,
			})
		}

		toFieldName := targetField.Name

		tflog.SubsystemTrace(ctx, subsystemName, "Found XML wrapper split source", map[string]any{
			logAttrKeySourceFieldname: fromFieldName,
			logAttrKeyTargetFieldname: toFieldName,
			"is_nil":                  isNil,
		})

		// Handle the XML wrapper split
		diags.Append(flattener.handleXMLWrapperSplit(ctx, sourcePath, sourceStructVal, fromFieldName, targetPath, valTo, sourceStructType, typeTo, toFieldName, isNil)...)
		if diags.HasError() {
			return diags
		}

		// Mark the source field as processed
		processedFields[fromFieldName] = true
	}

	return diags
}

// isXMLWrapperSplitSource checks if a struct type represents a source that should be
// split into multiple target fields (complex XML wrapper with additional fields beyond Items/Quantity)
func (flattener autoFlattener) isXMLWrapperSplitSource(structType reflect.Type) bool {
	hasValidItems := false
	hasValidQuantity := false
	hasOtherFields := false

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := field.Name
		fieldType := field.Type

		switch fieldName {
		case getXMLWrapperSliceFieldName(structType):
			// Items must be a slice to be a valid XML wrapper
			if fieldType.Kind() == reflect.Slice {
				hasValidItems = true
			}
		case xmlWrapperFieldQuantity:
			// Quantity must be *int32 to be a valid XML wrapper
			if fieldType == reflect.TypeFor[*int32]() {
				hasValidQuantity = true
			}
		default:
			// Any other field suggests this is a complex structure that should be split
			hasOtherFields = true
		}
	}

	// Must have properly typed Items/Quantity (XML wrapper pattern) plus other fields to be a split candidate
	return hasValidItems && hasValidQuantity && hasOtherFields
}

// handleXMLWrapperSplit splits a complex AWS XML wrapper structure into multiple Terraform fields
func (flattener autoFlattener) handleXMLWrapperSplit(ctx context.Context, sourcePath path.Path, sourceStructVal reflect.Value, sourceFieldName string, targetPath path.Path, valTo reflect.Value, sourceStructType, typeTo reflect.Type, targetFieldName string, isNil bool) diag.Diagnostics {
	var diags diag.Diagnostics

	sourcePath = sourcePath.AtName(sourceFieldName)
	targetPath = targetPath.AtName(targetFieldName)

	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(sourceStructType))
	ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetPath, targetPath.String())

	targetField, found := (&fuzzyFieldFinder{}).findField(ctx, sourceFieldName, reflect.StructOf([]reflect.StructField{{Name: sourceFieldName, Type: reflect.TypeFor[string](), PkgPath: ""}}), typeTo, flattener)
	if found { // Redundant, this was already checked in `handleXMLWrapperCollapse`
		ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeyTargetType, fullTypeName(targetField.Type))
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Handling XML wrapper split")

	// If source is nil, find and set the matching target field to null
	if isNil {
		toFieldVal := valTo.FieldByIndex(targetField.Index)
		if toFieldVal.CanSet() {
			if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
				switch tTo := valTo.Type(ctx).(type) {
				case basetypes.SetTypable, basetypes.ListTypable:
					tflog.SubsystemTrace(ctx, subsystemName, "Setting target collection field to null", map[string]any{
						logAttrKeyTargetFieldname: targetField.Name,
					})

					var elemType attr.Type = types.StringType
					if tToWithElem, ok := tTo.(attr.TypeWithElementType); ok {
						elemType = tToWithElem.ElementType()
					}

					var nullVal attr.Value
					var d diag.Diagnostics
					if setType, ok := tTo.(basetypes.SetTypable); ok {
						nullVal, d = setType.ValueFromSet(ctx, types.SetNull(elemType))
					} else if listType, ok := tTo.(basetypes.ListTypable); ok {
						nullVal, d = listType.ValueFromList(ctx, types.ListNull(elemType))
					}
					diags.Append(d...)
					if !diags.HasError() {
						toFieldVal.Set(reflect.ValueOf(nullVal))
					}
				}
			}
		} else {
			// Corresponding field value can't be changed.
			tflog.SubsystemDebug(ctx, subsystemName, "Target field cannot be set", map[string]any{
				logAttrKeySourceFieldname: sourceFieldName,
				logAttrKeyTargetFieldname: targetFieldName,
			})
		}
		return diags
	}

	// Map each field in the source struct to corresponding target fields
	for sourceField := range tfreflect.ExportedStructFields(sourceStructType) {
		sourceFieldName := sourceField.Name
		sourceFieldVal := sourceStructVal.FieldByName(sourceFieldName)

		if sourceFieldName == xmlWrapperFieldQuantity {
			tflog.SubsystemTrace(ctx, subsystemName, "Skipping ignored source field")
			continue
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Processing source field for split", map[string]any{
			logAttrKeySourceFieldname: sourceFieldName,
			"source_type":             sourceField.Type.String(),
		})

		// Map the source field to target field(s)
		wrapperFieldName := getXMLWrapperSliceFieldName(sourceStructType)
		if sourceFieldName == wrapperFieldName {
			// Items and Quantity should map to the main collection field
			// Find a target field that matches the parent source field name
			mainTargetFieldName := flattener.findMainTargetFieldForSplit(ctx, sourcePath, typeTo)
			if mainTargetFieldName != "" {
				if _, found := typeTo.FieldByName(mainTargetFieldName); found {
					toFieldVal := valTo.FieldByName(mainTargetFieldName)
					if toFieldVal.CanSet() {
						// TODO: should be able to modify sourcePath here, but can't yet.
						// `convertXMLWrapperFieldToCollection` calls on to `convert` with sourcePath, and there's no test to confirm that it works
						// sourcePath = sourcePath.AtName(sourceFieldName)
						// ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourcePath, sourcePath.String())
						ctx = tflog.SubsystemSetField(ctx, subsystemName, logAttrKeySourceType, fullTypeName(sourceField.Type))

						tflog.SubsystemTrace(ctx, subsystemName, "Mapping Items to main target field", map[string]any{
							logAttrKeySourcePath: sourcePath.AtName(sourceFieldName).String(),
						})

						// Extract tagOptions from target field
						opts := tagOptions("")
						if targetField, ok := typeTo.FieldByName(mainTargetFieldName); ok {
							if tag := targetField.Tag.Get("autoflex"); tag != "" {
								_, opts = parseTag(tag)
							}
						}
						diags.Append(flattener.convertXMLWrapperFieldToCollection(ctx, sourcePath, sourceFieldName, sourceStructVal, targetPath, toFieldVal, opts)...)
						if diags.HasError() {
							return diags
						}
					} else {
						// Corresponding field value can't be changed.
						tflog.SubsystemDebug(ctx, subsystemName, "Target field cannot be set", map[string]any{
							logAttrKeySourceFieldname: sourceFieldName,
							logAttrKeyTargetFieldname: targetFieldName,
						})
					}
				}
			}
		} else {
			// Other fields should map directly by name
			if targetField, found := typeTo.FieldByName(sourceFieldName); found {
				toFieldVal := valTo.FieldByName(sourceFieldName)
				if toFieldVal.CanSet() {
					tflog.SubsystemTrace(ctx, subsystemName, "Mapping additional field by name", map[string]any{
						"source_field": sourceFieldName,
						"target_field": sourceFieldName,
					})

					// Extract tagOptions from target field
					opts := tagOptions("")
					if tag := targetField.Tag.Get("autoflex"); tag != "" {
						_, opts = parseTag(tag)
					}

					// Convert the source field to target field
					diags.Append(flattener.convertXMLWrapperFieldToCollection(ctx, sourcePath.AtName(sourceFieldName), "AAA", sourceFieldVal, targetPath, toFieldVal, opts)...)
					if diags.HasError() {
						return diags
					}
				} else {
					// Corresponding field value can't be changed.
					tflog.SubsystemDebug(ctx, subsystemName, "Target field cannot be set", map[string]any{
						logAttrKeySourceFieldname: sourceFieldName,
						logAttrKeyTargetFieldname: targetFieldName,
					})
				}
			}
		}
	}

	return diags
}

// findMainTargetFieldForSplit determines which target field should receive the Items/Quantity from the source
func (flattener autoFlattener) findMainTargetFieldForSplit(ctx context.Context, sourcePath path.Path, typeTo reflect.Type) string {
	sourcePathStr := sourcePath.String()
	if lastDot := strings.LastIndex(sourcePathStr, "."); lastDot >= 0 {
		sourceFieldName := sourcePathStr[lastDot+1:]

		// Use fuzzy field finder for proper singular/plural and case matching
		dummySourceType := reflect.StructOf([]reflect.StructField{{Name: sourceFieldName, Type: reflect.TypeFor[string](), PkgPath: ""}})
		if targetField, ok := (&fuzzyFieldFinder{}).findField(ctx, sourceFieldName, dummySourceType, typeTo, flattener); ok {
			return targetField.Name
		}
	}

	return ""
}

// convertXMLWrapperFieldToCollection converts a source field (either XML wrapper or simple field) to a target collection
func (flattener autoFlattener) convertXMLWrapperFieldToCollection(ctx context.Context, sourcePath path.Path, sourceFieldName string, sourceFieldVal reflect.Value, targetPath path.Path, toFieldVal reflect.Value, opts tagOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	// Check if source is an XML wrapper struct (has slice field + Quantity)
	if sourceFieldVal.Kind() == reflect.Struct && potentialXMLWrapperStruct(sourceFieldVal.Type()) {
		sourcePath = sourcePath.AtName(sourceFieldName)
		tflog.SubsystemTrace(ctx, subsystemName, "Converting XML wrapper struct to collection", map[string]any{
			logAttrKeySourcePath: sourcePath.String(),
		})

		// Use existing XML wrapper flatten logic
		if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
			diags.Append(flattener.xmlWrapperFlatten(ctx, sourcePath, sourceFieldName, sourceFieldVal, targetPath, valTo.Type(ctx), toFieldVal, opts)...)
			if diags.HasError() {
				return diags
			}
		}
	} else if sourceFieldVal.Kind() == reflect.Pointer && !sourceFieldVal.IsNil() && potentialXMLWrapperStruct(sourceFieldVal.Type().Elem()) {
		sourcePath = sourcePath.AtName(sourceFieldName)
		tflog.SubsystemTrace(ctx, subsystemName, "Converting pointer to XML wrapper struct to collection", map[string]any{
			logAttrKeySourcePath: sourcePath.String(),
		})

		// Use existing XML wrapper flatten logic
		if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
			diags.Append(flattener.xmlWrapperFlatten(ctx, sourcePath, sourceFieldName, sourceFieldVal.Elem(), targetPath, valTo.Type(ctx), toFieldVal, opts)...)
			if diags.HasError() {
				return diags
			}
		}
	} else {
		tflog.SubsystemTrace(ctx, subsystemName, "Converting non-XML wrapper field", map[string]any{
			"source_type": sourceFieldVal.Type().String(),
		})

		// For non-XML wrapper fields, use regular conversion
		fieldOpts := fieldOpts{}
		diags.Append(flattener.convert(ctx, sourcePath, sourceFieldVal, targetPath, toFieldVal, fieldOpts)...)
		if diags.HasError() {
			return diags
		}
	}

	return diags
}

// createTargetValue creates a value of the target type from a source string value
func (flattener autoFlattener) createTargetValue(ctx context.Context, sourceValue types.String, targetType attr.Type) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Check what type of target we have
	switch targetType := targetType.(type) {
	case basetypes.StringTypable:
		// Regular string type
		val, d := targetType.ValueFromString(ctx, sourceValue)
		diags.Append(d...)
		return val, diags
	default:
		// For StringEnum types or other complex types, try to use regular conversion
		// This is a bit of a hack, but StringEnum types should accept string values
		targetTypeStr := targetType.String()
		if strings.Contains(targetTypeStr, "StringEnum") {
			// Create a zero value of the target type and try to set it
			targetVal := reflect.New(reflect.TypeOf(targetType.ValueType(ctx))).Elem()

			// Try to convert using the AutoFlex conversion logic
			diags.Append(flattener.convert(ctx, path.Empty(), reflect.ValueOf(sourceValue), path.Empty(), targetVal, fieldOpts{})...)
			if diags.HasError() {
				return nil, diags
			}

			if attrVal, ok := targetVal.Interface().(attr.Value); ok {
				return attrVal, diags
			}
		}

		// Fallback to string value
		return sourceValue, diags
	}
}

// xmlWrapperFlattenRule2 handles Rule 2: XML wrapper to single plural block with items + additional fields
func (flattener *autoFlattener) xmlWrapperFlattenRule2(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value, opts tagOptions) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "xmlWrapperFlattenRule2 ENTRY", map[string]any{
		"source_type": vFrom.Type().String(),
	})

	// Check if wrapper Items field is empty/nil
	wrapperFieldName := getXMLWrapperSliceFieldName(vFrom.Type())
	itemsField := vFrom.FieldByName(wrapperFieldName)

	itemsLen := 0
	if itemsField.IsValid() && !itemsField.IsNil() {
		itemsLen = itemsField.Len()
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Checking Items field", map[string]any{
		"wrapper_field_name": wrapperFieldName,
		"items_valid":        itemsField.IsValid(),
		"items_len":          itemsLen,
	})

	if itemsField.IsValid() {
		itemsEmpty := itemsField.IsNil() || (itemsField.Kind() == reflect.Slice && itemsField.Len() == 0)

		tflog.SubsystemTrace(ctx, subsystemName, "Items empty check", map[string]any{
			"items_empty": itemsEmpty,
		})

		// If Items is empty AND all other fields (except Quantity) are nil or zero,
		// return null instead of creating a block with all default values
		if itemsEmpty {
			allOtherFieldsZero := true
			for i := 0; i < vFrom.NumField(); i++ {
				sourceField := vFrom.Field(i)
				fieldName := vFrom.Type().Field(i).Name

				// Skip Items and Quantity fields
				if fieldName == wrapperFieldName || fieldName == xmlWrapperFieldQuantity {
					continue
				}

				// For pointers: nil is zero, non-nil pointer to zero value is also considered zero
				// For non-pointers: use IsZero()
				isFieldZero := false
				if sourceField.Kind() == reflect.Pointer {
					if sourceField.IsNil() {
						isFieldZero = true
					} else if sourceField.Elem().IsZero() {
						isFieldZero = true
					}
				} else {
					isFieldZero = sourceField.IsZero()
				}

				tflog.SubsystemTrace(ctx, subsystemName, "Checking field zero status", map[string]any{
					"field_name":    fieldName,
					"is_field_zero": isFieldZero,
					"field_kind":    sourceField.Kind().String(),
					"field_is_nil":  sourceField.Kind() == reflect.Pointer && sourceField.IsNil(),
				})

				if !isFieldZero {
					allOtherFieldsZero = false
					break
				}
			}

			tflog.SubsystemTrace(ctx, subsystemName, "Zero value check result", map[string]any{
				"all_other_fields_zero": allOtherFieldsZero,
			})

			if allOtherFieldsZero {
				// Check if target field has omitempty - only return null if it does
				if opts.OmitEmpty() {
					tflog.SubsystemTrace(ctx, subsystemName, "XML wrapper Items is empty and all other fields are zero, returning null for Rule 2 (omitempty)")
					nullVal, d := tTo.NullValue(ctx)
					diags.Append(d...)
					if !diags.HasError() {
						vTo.Set(reflect.ValueOf(nullVal))
					}
					return diags
				}
				tflog.SubsystemTrace(ctx, subsystemName, "XML wrapper Items is empty and all other fields are zero, but no omitempty - creating zero-value struct for Rule 2")
			}
		}
	}

	nestedObjPtr, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	nestedObjValue := reflect.ValueOf(nestedObjPtr).Elem()

	// Map Items field - find target field by xmlwrapper tag
	if itemsField := vFrom.FieldByName(wrapperFieldName); itemsField.IsValid() {
		for i := 0; i < nestedObjValue.NumField(); i++ {
			targetField := nestedObjValue.Field(i)
			targetFieldType := nestedObjValue.Type().Field(i)
			_, fieldOpts := autoflexTags(targetFieldType)
			if fieldOpts.XMLWrapperField() == wrapperFieldName {
				if targetField.CanAddr() {
					if targetAttr, ok := targetField.Addr().Interface().(attr.Value); ok {
						diags.Append(autoFlattenConvert(ctx, itemsField.Interface(), targetAttr, flattener)...)
					}
				}
				break
			}
		}
	}

	// Map all other fields (skip Items and Quantity)
	for i := 0; i < vFrom.NumField(); i++ {
		sourceField := vFrom.Field(i)
		fieldName := vFrom.Type().Field(i).Name

		if fieldName == wrapperFieldName || fieldName == xmlWrapperFieldQuantity {
			continue
		}

		if targetField := nestedObjValue.FieldByName(fieldName); targetField.IsValid() && targetField.CanAddr() {
			if targetAttr, ok := targetField.Addr().Interface().(attr.Value); ok {
				diags.Append(autoFlattenConvert(ctx, sourceField.Interface(), targetAttr, flattener)...)
			}
		}
	}

	ptrType := reflect.TypeOf(nestedObjPtr)
	objectSlice := reflect.MakeSlice(reflect.SliceOf(ptrType), 1, 1)
	objectSlice.Index(0).Set(reflect.ValueOf(nestedObjPtr))

	targetValue, d := tTo.ValueFromObjectSlice(ctx, objectSlice.Interface())
	diags.Append(d...)
	if !diags.HasError() {
		vTo.Set(reflect.ValueOf(targetValue))
	}

	return diags
}
