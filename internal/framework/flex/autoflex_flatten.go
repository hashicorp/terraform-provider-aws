// Copyright (c) HashiCorp, Inc.
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
	Options AutoFlexOptions
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
		Options: o,
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

	// Special case: XML wrapper struct to NestedObjectCollectionType (Rule 2)
	// Check this BEFORE struct-to-struct conversion
	if valFrom.IsValid() {
		sourceType := valFrom.Type()
		sourceVal := valFrom
		
		// Handle pointer to XML wrapper struct
		if sourceType.Kind() == reflect.Pointer && !valFrom.IsNil() && isXMLWrapperStruct(sourceType.Elem()) {
			sourceType = sourceType.Elem()
			sourceVal = valFrom.Elem()
		}
		
		tflog.SubsystemDebug(ctx, subsystemName, "Checking XML wrapper special case", map[string]any{
			"sourceType": sourceType.String(),
			"sourceKind": sourceType.Kind().String(),
			"isXMLWrapper": isXMLWrapperStruct(sourceType),
			"toType": reflect.TypeOf(to).String(),
		})
		
		if sourceType.Kind() == reflect.Struct && isXMLWrapperStruct(sourceType) {
			tflog.SubsystemDebug(ctx, subsystemName, "Source is XML wrapper, checking target", map[string]any{
				"toImplementsAttrValue": reflect.TypeOf(to).Implements(reflect.TypeOf((*attr.Value)(nil)).Elem()),
			})
			if attrVal, ok := to.(attr.Value); ok {
				// Handle Rule 2: NestedObjectCollectionType target
				if nestedObjType, ok := attrVal.Type(ctx).(fwtypes.NestedObjectCollectionType); ok {
					// Check if wrapper is in empty state (Enabled=false, Items empty/nil)
					// If so, return null instead of creating a block
					enabledField := sourceVal.FieldByName("Enabled")
					itemsField := sourceVal.FieldByName("Items")
					
					if enabledField.IsValid() && itemsField.IsValid() {
						// Check if Enabled is false and Items is empty
						isEnabled := true
						if enabledField.Kind() == reflect.Pointer && !enabledField.IsNil() {
							isEnabled = enabledField.Elem().Bool()
						}
						
						itemsEmpty := itemsField.IsNil() || (itemsField.Kind() == reflect.Slice && itemsField.Len() == 0)
						
						if !isEnabled && itemsEmpty {
							tflog.SubsystemTrace(ctx, subsystemName, "XML wrapper is empty (Enabled=false, Items empty), returning null")
							nullVal, d := nestedObjType.NullValue(ctx)
							diags.Append(d...)
							if !diags.HasError() {
								valTo.Set(reflect.ValueOf(nullVal))
							}
							return diags
						}
					}
					
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening XML wrapper to NestedObjectCollection via autoFlattenConvert")
					
					// Inline Rule 2 logic
					nestedObjPtr, d := nestedObjType.NewObjectPtr(ctx)
					diags.Append(d...)
					if !diags.HasError() {
						nestedObjValue := reflect.ValueOf(nestedObjPtr).Elem()

						// Map all fields except Quantity
						for i := 0; i < sourceVal.NumField(); i++ {
							sourceField := sourceVal.Field(i)
							fieldName := sourceVal.Type().Field(i).Name

							if fieldName == "Quantity" {
								continue
							}

							if targetField := nestedObjValue.FieldByName(fieldName); targetField.IsValid() && targetField.CanAddr() {
								if targetAttr, ok := targetField.Addr().Interface().(attr.Value); ok {
									diags.Append(autoFlattenConvert(ctx, sourceField.Interface(), targetAttr, flexer)...)
								}
							}
						}

						if !diags.HasError() {
							ptrType := reflect.TypeOf(nestedObjPtr)
							objectSlice := reflect.MakeSlice(reflect.SliceOf(ptrType), 1, 1)
							objectSlice.Index(0).Set(reflect.ValueOf(nestedObjPtr))

							targetValue, d := nestedObjType.ValueFromObjectSlice(ctx, objectSlice.Interface())
							diags.Append(d...)
							if !diags.HasError() {
								valTo.Set(reflect.ValueOf(targetValue))
								return diags
							}
						}
					}
					return diags
				}
				
				// Handle Rule 1: Simple Set/List target (extract Items field)
				itemsField := sourceVal.FieldByName("Items")
				if itemsField.IsValid() {
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening XML wrapper Items to simple collection")
					// Call convert on the Items field
					diags.Append(flexer.convert(ctx, sourcePath, itemsField, targetPath, valTo, fieldOpts{})...)
					return diags
				}
			}
		}
	}

	// Top-level struct to struct conversion.
	if valFrom.IsValid() && valTo.IsValid() {
		if typFrom, typTo := valFrom.Type(), valTo.Type(); typFrom.Kind() == reflect.Struct && typTo.Kind() == reflect.Struct &&
			!typTo.Implements(reflect.TypeFor[basetypes.ListValuable]()) &&
			!typTo.Implements(reflect.TypeFor[basetypes.SetValuable]()) {
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

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

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

	default:
		// Check for custom types with string underlying type (e.g., enums)
		// For type declarations like "type MyEnum string", Kind() will not be reflect.String,
		// but we can convert values to string using String() method
		if tSliceElem.Kind() != reflect.String {
			// Try to see if this is a string-based custom type by checking if we can call String() on it
			sampleVal := reflect.New(tSliceElem).Elem()
			if sampleVal.CanInterface() {
				// Check if it has an underlying string type
				if tSliceElem.ConvertibleTo(reflect.TypeOf("")) {
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
	}

	tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
		"from": vFrom.Kind(),
		"to":   tTo,
	})

	return diags
}

// map_ copies an AWS API map value to a compatible Plugin Framework value.
func (flattener autoFlattener) map_(ctx context.Context, sourcePath path.Path, vFrom reflect.Value, targetPath path.Path, tTo attr.Type, vTo reflect.Value, _ fieldOpts) diag.Diagnostics {
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
							logAttrKeyTargetType: fullTypeName(reflect.TypeOf(innerElements)),
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
							logAttrKeyTargetType: fullTypeName(reflect.TypeOf(innerElements)),
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

		tflog.SubsystemError(ctx, subsystemName, "AutoFlex Flatten; incompatible types", map[string]any{
			"from": vFrom.Kind(),
			"to":   tTo,
		})
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
			tflog.SubsystemTrace(ctx, subsystemName, "Flattening with NullValue")
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
func (flattener autoFlattener) xmlWrapperFlatten(ctx context.Context, vFrom reflect.Value, tTo attr.Type, vTo reflect.Value, wrapperField string) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "Starting XML wrapper flatten", map[string]any{
		"source_type":   vFrom.Type().String(),
		"target_type":   tTo.String(),
		"wrapper_field": wrapperField,
	})

	// Check if target is a NestedObjectCollection (Rule 2 pattern)
	// Rule 2: Target model has Items field (represents wrapper structure)
	// Rule 1: Target model represents actual items directly
	if nestedObjType, ok := tTo.(fwtypes.NestedObjectCollectionType); ok {
		tflog.SubsystemTrace(ctx, subsystemName, "Target is NestedObjectCollectionType - checking for Rule 2", map[string]any{
			"target_type": tTo.String(),
		})
		
		// Create a sample object to check its structure
		samplePtr, d := nestedObjType.NewObjectPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			tflog.SubsystemError(ctx, subsystemName, "Failed to create sample object for Rule 2 detection")
			return diags
		}

		// Check if the target model has an Items field (Rule 2 pattern)
		sampleValue := reflect.ValueOf(samplePtr).Elem()
		hasItems := sampleValue.FieldByName("Items").IsValid()
		
		tflog.SubsystemTrace(ctx, subsystemName, "Rule 2 detection result", map[string]any{
			"has_items_field": hasItems,
			"sample_type":     sampleValue.Type().String(),
		})
		
		if hasItems {
			tflog.SubsystemTrace(ctx, subsystemName, "Using Rule 2 flatten")
			return flattener.xmlWrapperFlattenRule2(ctx, vFrom, nestedObjType, vTo)
		}
		// Otherwise continue with Rule 1 logic below
	}

	// Rule 1: Continue with existing logic
	// Verify source is a valid XML wrapper struct
	if !isXMLWrapperStruct(vFrom.Type()) {
		tflog.SubsystemError(ctx, subsystemName, "Source is not a valid XML wrapper struct", map[string]any{
			"source_type": vFrom.Type().String(),
		})
		diags.Append(DiagFlatteningIncompatibleTypes(vFrom.Type(), reflect.TypeOf(vTo.Interface())))
		return diags
	}

	// Get the Items field from the source wrapper struct
	itemsField := vFrom.FieldByName("Items")
	if !itemsField.IsValid() {
		tflog.SubsystemError(ctx, subsystemName, "XML wrapper struct missing Items field")
		diags.Append(DiagFlatteningIncompatibleTypes(vFrom.Type(), reflect.TypeOf(vTo.Interface())))
		return diags
	}

	tflog.SubsystemTrace(ctx, subsystemName, "Found Items field", map[string]any{
		"items_type":   itemsField.Type().String(),
		"items_kind":   itemsField.Kind().String(),
		"items_len":    itemsField.Len(),
		"items_is_nil": itemsField.IsNil(),
	})

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

			tflog.SubsystemTrace(ctx, subsystemName, "Processing item", map[string]any{
				"index":      i,
				"item_kind":  item.Kind().String(),
				"item_value": item.Interface(),
			})

			// Convert each item based on its type
			switch item.Kind() {
			case reflect.Int32:
				elements[i] = types.Int64Value(item.Int())
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
					if derefItem.Type().ConvertibleTo(reflect.TypeOf("")) {
						stringVal := derefItem.Convert(reflect.TypeOf("")).String()
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
			default:
				// Check for custom string types (like enums)
				if item.Type().ConvertibleTo(reflect.TypeOf("")) {
					// Convert custom string type (like testEnum/Method) to string
					stringVal := item.Convert(reflect.TypeOf("")).String()

					// Try to create a value that matches the target element type
					if val, d := flattener.createTargetValue(ctx, types.StringValue(stringVal), elementType); d.HasError() {
						diags.Append(d...)
						return diags
					} else {
						elements[i] = val
					}
				} else if item.Kind() == reflect.Struct {
					// This would need to be handled by a nested object conversion
					// For now, we'll return an error for unsupported types
					diags.Append(DiagFlatteningIncompatibleTypes(item.Type(), reflect.TypeOf(elementType)))
					return diags
				} else {
					diags.Append(DiagFlatteningIncompatibleTypes(item.Type(), reflect.TypeOf(elementType)))
					return diags
				}
			}
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Creating list value", map[string]any{
			"element_count": len(elements),
			"element_type":  elementType.String(),
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

			tflog.SubsystemTrace(ctx, subsystemName, "Processing item", map[string]any{
				"index":      i,
				"item_kind":  item.Kind().String(),
				"item_value": item.Interface(),
			})

			// Convert each item based on its type
			switch item.Kind() {
			case reflect.Int32:
				elements[i] = types.Int64Value(item.Int())
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
					if derefItem.Type().ConvertibleTo(reflect.TypeOf("")) {
						stringVal := derefItem.Convert(reflect.TypeOf("")).String()
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
				if item.Type().ConvertibleTo(reflect.TypeOf("")) {
					// Convert custom string type (like testEnum/Method) to string
					stringVal := item.Convert(reflect.TypeOf("")).String()

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
			"element_type":  elementType.String(),
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

	// Special case: If source is an XML wrapper and target is a NestedObjectCollectionType,
	// handle it directly without needing xmlWrapperFlatten
	if isXMLWrapperStruct(typeFrom) {
		tflog.SubsystemTrace(ctx, subsystemName, "Source is XML wrapper struct", map[string]any{
			logAttrKeySourceType: typeFrom.String(),
			logAttrKeyTargetType: typeTo.String(),
		})
		
		if attrVal, ok := to.(attr.Value); ok {
			tflog.SubsystemTrace(ctx, subsystemName, "Target implements attr.Value", map[string]any{
				"target_type": attrVal.Type(ctx).String(),
			})
			
			if nestedObjType, ok := attrVal.Type(ctx).(fwtypes.NestedObjectCollectionType); ok {
				tflog.SubsystemTrace(ctx, subsystemName, "Target is NestedObjectCollectionType", map[string]any{
					"target_type": nestedObjType.String(),
				})
				
				// Check if wrapper is in empty state (Enabled=false, Items empty/nil)
				// If so, return null instead of creating a block
				enabledField := valFrom.FieldByName("Enabled")
				itemsField := valFrom.FieldByName("Items")
				
				if enabledField.IsValid() && itemsField.IsValid() {
					// Check if Enabled is false and Items is empty
					isEnabled := true
					if enabledField.Kind() == reflect.Pointer && !enabledField.IsNil() {
						isEnabled = enabledField.Elem().Bool()
					}
					
					itemsEmpty := itemsField.IsNil() || (itemsField.Kind() == reflect.Slice && itemsField.Len() == 0)
					
					if !isEnabled && itemsEmpty {
						tflog.SubsystemTrace(ctx, subsystemName, "XML wrapper is empty (Enabled=false, Items empty), returning null for Rule 2")
						nullVal, d := nestedObjType.NullValue(ctx)
						diags.Append(d...)
						if !diags.HasError() {
							valTo.Set(reflect.ValueOf(nullVal))
						}
						return diags
					}
				}
				
				// Check if target model has Items field (Rule 2)
				samplePtr, d := nestedObjType.NewObjectPtr(ctx)
				diags.Append(d...)
				if !diags.HasError() {
					sampleValue := reflect.ValueOf(samplePtr).Elem()
					if sampleValue.FieldByName("Items").IsValid() {
						// Rule 2: Create single-element collection with nested object containing Items + other fields
						tflog.SubsystemTrace(ctx, subsystemName, "Flattening XML wrapper to NestedObjectCollection (Rule 2)", map[string]any{
							logAttrKeySourceType: typeFrom.String(),
							logAttrKeyTargetType: typeTo.String(),
						})
						
						// Map source fields to target nested object fields
						for i := 0; i < typeFrom.NumField(); i++ {
							sourceField := typeFrom.Field(i)
							sourceFieldName := sourceField.Name
							sourceFieldVal := valFrom.Field(i)
							
							if sourceFieldName == "Quantity" {
								continue // Skip Quantity
							}
							
							targetField := sampleValue.FieldByName(sourceFieldName)
							if targetField.IsValid() && targetField.CanAddr() {
								if targetAttr, ok := targetField.Addr().Interface().(attr.Value); ok {
									diags.Append(autoFlattenConvert(ctx, sourceFieldVal.Interface(), targetAttr, flexer)...)
								}
							}
						}
						
						if !diags.HasError() {
							// Create single-element collection
							ptrType := reflect.TypeOf(samplePtr)
							objectSlice := reflect.MakeSlice(reflect.SliceOf(ptrType), 1, 1)
							objectSlice.Index(0).Set(reflect.ValueOf(samplePtr))
							
							targetValue, d := nestedObjType.ValueFromObjectSlice(ctx, objectSlice.Interface())
							diags.Append(d...)
							if !diags.HasError() {
								valTo.Set(reflect.ValueOf(targetValue))
								return diags
							}
						}
					}
				}
			}
			
			// Try other collection types if we have autoFlattener
			if f, ok := flexer.(*autoFlattener); ok {
				switch attrVal.Type(ctx).(type) {
				case basetypes.SetTypable, basetypes.ListTypable:
					tflog.SubsystemTrace(ctx, subsystemName, "Flattening XML wrapper directly to collection", map[string]any{
						logAttrKeySourceType: typeFrom.String(),
						logAttrKeyTargetType: typeTo.String(),
					})
					return f.xmlWrapperFlatten(ctx, valFrom, attrVal.Type(ctx), valTo, "Items")
				}
			}
		}
	}

	// Special handling: Check if the entire source struct is an XML wrapper
	// and should be flattened to a target field with wrapper tag (Rule 1 only)
	// Skip this if target is a struct - that's Rule 2 where we map fields individually
	if isXMLWrapperStruct(typeFrom) && typeTo.Kind() != reflect.Struct {
		for toField := range tfreflect.ExportedStructFields(typeTo) {
			toFieldName := toField.Name
			_, toOpts := autoflexTags(toField)
			if wrapperField := toOpts.XMLWrapperField(); wrapperField != "" {
				toFieldVal := valTo.FieldByIndex(toField.Index)
				if !toFieldVal.CanSet() {
					continue
				}

				tflog.SubsystemTrace(ctx, subsystemName, "Converting entire XML wrapper struct to collection field (Rule 1)", map[string]any{
					logAttrKeySourceType:      typeFrom.String(),
					logAttrKeyTargetFieldname: toFieldName,
					"wrapper_field":           wrapperField,
				})

				valTo, ok := toFieldVal.Interface().(attr.Value)
				if !ok {
					tflog.SubsystemError(ctx, subsystemName, "Target field does not implement attr.Value")
					diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(toFieldVal.Interface())))
					return diags
				}

				if f, ok := flexer.(*autoFlattener); ok {
					diags.Append(f.xmlWrapperFlatten(ctx, valFrom, valTo.Type(ctx), toFieldVal, wrapperField)...)
				} else {
					diags.Append(DiagFlatteningIncompatibleTypes(valFrom.Type(), reflect.TypeOf(toFieldVal.Interface())))
				}
				if diags.HasError() {
					return diags
				}
				// Successfully handled as XML wrapper, don't process individual fields
				return diags
			}
		}
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
			tflog.SubsystemDebug(ctx, subsystemName, "No corresponding field", map[string]any{
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

		// Check if target has wrapper tag and source is an XML wrapper struct
		if wrapperField := toFieldOpts.XMLWrapperField(); wrapperField != "" {
			fromFieldVal := valFrom.FieldByIndex(fromField.Index)

			// Handle direct XML wrapper struct
			if isXMLWrapperStruct(fromFieldVal.Type()) {
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
					diags.Append(f.xmlWrapperFlatten(ctx, fromFieldVal, valTo.Type(ctx), toFieldVal, wrapperField)...)
				} else {
					diags.Append(DiagFlatteningIncompatibleTypes(fromFieldVal.Type(), reflect.TypeOf(toFieldVal.Interface())))
				}
				if diags.HasError() {
					break
				}
				continue
			}

			// Handle pointer to XML wrapper struct
			if fromFieldVal.Kind() == reflect.Pointer && !fromFieldVal.IsNil() && isXMLWrapperStruct(fromFieldVal.Type().Elem()) {
				tflog.SubsystemTrace(ctx, subsystemName, "Converting pointer to XML wrapper struct to collection via wrapper tag", map[string]any{
					logAttrKeySourceFieldname: fromFieldName,
					logAttrKeyTargetFieldname: toFieldName,
					"wrapper_field":           wrapperField,
					"flexer_type":             fmt.Sprintf("%T", flexer),
				})

				valTo, ok := toFieldVal.Interface().(attr.Value)
				if !ok {
					tflog.SubsystemError(ctx, subsystemName, "Target field does not implement attr.Value")
					diags.Append(diagFlatteningTargetDoesNotImplementAttrValue(reflect.TypeOf(toFieldVal.Interface())))
					break
				}

				// Try both value and pointer type assertions
				if f, ok := flexer.(autoFlattener); ok {
					diags.Append(f.xmlWrapperFlatten(ctx, fromFieldVal.Elem(), valTo.Type(ctx), toFieldVal, wrapperField)...)
				} else if f, ok := flexer.(*autoFlattener); ok {
					diags.Append(f.xmlWrapperFlatten(ctx, fromFieldVal.Elem(), valTo.Type(ctx), toFieldVal, wrapperField)...)
				} else {
					tflog.SubsystemError(ctx, subsystemName, "Type assertion to autoFlattener failed for wrapper tag", map[string]any{
						"flexer_type": fmt.Sprintf("%T", flexer),
					})
					diags.Append(DiagFlatteningIncompatibleTypes(fromFieldVal.Type(), reflect.TypeOf(toFieldVal.Interface())))
				}
				if diags.HasError() {
					break
				}
				continue
			}

			// Handle nil pointer to XML wrapper struct
			if fromFieldVal.Kind() == reflect.Pointer && fromFieldVal.IsNil() && isXMLWrapperStruct(fromFieldVal.Type().Elem()) {
				tflog.SubsystemTrace(ctx, subsystemName, "Converting nil pointer to XML wrapper struct to null collection", map[string]any{
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
			if toOpts.XMLWrapperField() != "" && !fromFieldVal.IsNil() && isXMLWrapperStruct(fromFieldVal.Type().Elem()) {
				// Check if target is a collection type (Set, List, or NestedObjectCollection)
				if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
					targetType := valTo.Type(ctx)
					switch tt := targetType.(type) {
					case basetypes.SetTypable, basetypes.ListTypable, fwtypes.NestedObjectCollectionType:
						// Check if wrapper is in empty state (Enabled=false, Items empty/nil)
						// If so, set null and skip processing
						wrapperVal := fromFieldVal.Elem()
						enabledField := wrapperVal.FieldByName("Enabled")
						itemsField := wrapperVal.FieldByName("Items")
						
						if enabledField.IsValid() && itemsField.IsValid() {
							isEnabled := true
							if enabledField.Kind() == reflect.Pointer && !enabledField.IsNil() {
								isEnabled = enabledField.Elem().Bool()
							}
							
							itemsEmpty := itemsField.IsNil() || (itemsField.Kind() == reflect.Slice && itemsField.Len() == 0)
							
							if !isEnabled && itemsEmpty {
								tflog.SubsystemTrace(ctx, subsystemName, "XML wrapper is empty (Enabled=false, Items empty), setting null", map[string]any{
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
						
						// Determine wrapper field based on target type
						// For NestedObjectCollection, the wrapper field doesn't matter as Rule 2 will be used
						wrapperField := "Items"
						
						tflog.SubsystemTrace(ctx, subsystemName, "Auto-converting pointer to XML wrapper struct to collection", map[string]any{
							logAttrKeySourceFieldname: fromFieldName,
							logAttrKeyTargetFieldname: toFieldName,
							"wrapper_field":           wrapperField,
						})

						// Try both value and pointer type assertions
						if f, ok := flexer.(autoFlattener); ok {
							diags.Append(f.xmlWrapperFlatten(ctx, fromFieldVal.Elem(), targetType, toFieldVal, wrapperField)...)
							if diags.HasError() {
								break
							}
							continue // Successfully handled, skip normal processing
						} else if f, ok := flexer.(*autoFlattener); ok {
							diags.Append(f.xmlWrapperFlatten(ctx, fromFieldVal.Elem(), targetType, toFieldVal, wrapperField)...)
							if diags.HasError() {
								break
							}
							continue // Successfully handled, skip normal processing
						}
						// If flexer is not autoFlattener, fall through to normal field matching
					}
				}
			} else if toOpts.XMLWrapperField() != "" && fromFieldVal.IsNil() && isXMLWrapperStruct(fromFieldVal.Type().Elem()) {
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
		} else if toOpts.XMLWrapperField() != "" && isXMLWrapperStruct(fromFieldVal.Type()) {
			// Handle direct XML wrapper struct (non-pointer)
			if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
				switch valTo.Type(ctx).(type) {
				case basetypes.SetTypable, basetypes.ListTypable:
					tflog.SubsystemTrace(ctx, subsystemName, "Auto-converting XML wrapper struct to collection", map[string]any{
						logAttrKeySourceFieldname: fromFieldName,
						logAttrKeyTargetFieldname: toFieldName,
						"wrapper_field":           "Items",
					})

					if f, ok := flexer.(*autoFlattener); ok {
						diags.Append(f.xmlWrapperFlatten(ctx, fromFieldVal, valTo.Type(ctx), toFieldVal, "Items")...)
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

		opts := fieldOpts{
			legacy:     toFieldOpts.Legacy(),
			omitempty:  toFieldOpts.OmitEmpty(),
			xmlWrapper: toFieldOpts.XMLWrapperField() != "",
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

// handleXMLWrapperCollapse handles the reverse of XML wrapper collapse - i.e., XML wrapper split.
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
		if targetField, ok := (&fuzzyFieldFinder{}).findField(ctx, fromFieldName, typeFrom, typeTo, flattener); ok {
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
			}
		}

		tflog.SubsystemTrace(ctx, subsystemName, "Found XML wrapper split source", map[string]any{
			logAttrKeySourceFieldname: fromFieldName,
			logAttrKeySourceType:      sourceStructType.String(),
			"is_nil":                  isNil,
		})

		// Handle the XML wrapper split
		diags.Append(flattener.handleXMLWrapperSplit(ctx, sourcePath.AtName(fromFieldName), sourceStructVal, targetPath, valTo, sourceStructType, typeTo, isNil, processedFields)...)
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
		case "Items":
			// Items must be a slice to be a valid XML wrapper
			if fieldType.Kind() == reflect.Slice {
				hasValidItems = true
			}
		case "Quantity":
			// Quantity must be *int32 to be a valid XML wrapper
			if fieldType == reflect.TypeOf((*int32)(nil)) {
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
func (flattener autoFlattener) handleXMLWrapperSplit(ctx context.Context, sourcePath path.Path, sourceStructVal reflect.Value, targetPath path.Path, valTo reflect.Value, sourceStructType, typeTo reflect.Type, isNil bool, processedFields map[string]bool) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog.SubsystemTrace(ctx, subsystemName, "Handling XML wrapper split", map[string]any{
		logAttrKeySourceType: sourceStructType.String(),
		logAttrKeyTargetType: typeTo.String(),
	})

	// If source is nil, find and set the matching target field to null
	if isNil {
		// Extract source field name from path
		sourceFieldName := ""
		sourcePathStr := sourcePath.String()
		if lastDot := strings.LastIndex(sourcePathStr, "."); lastDot >= 0 {
			sourceFieldName = sourcePathStr[lastDot+1:]
		}

		// Find matching target field
		if sourceFieldName != "" {
			if targetField, ok := (&fuzzyFieldFinder{}).findField(ctx, sourceFieldName, reflect.StructOf([]reflect.StructField{{Name: sourceFieldName, Type: reflect.TypeOf(""), PkgPath: ""}}), typeTo, flattener); ok {
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
				}
			}
		}
		return diags
	}

	// Map each field in the source struct to corresponding target fields
	for i := 0; i < sourceStructType.NumField(); i++ {
		sourceField := sourceStructType.Field(i)
		sourceFieldName := sourceField.Name
		sourceFieldVal := sourceStructVal.Field(i)

		tflog.SubsystemTrace(ctx, subsystemName, "Processing source field for split", map[string]any{
			"source_field": sourceFieldName,
			"source_type":  sourceField.Type.String(),
		})

		// Map the source field to target field(s)
		if sourceFieldName == "Items" || sourceFieldName == "Quantity" {
			// Items and Quantity should map to the main collection field
			// Find a target field that matches the parent source field name
			mainTargetFieldName := flattener.findMainTargetFieldForSplit(sourcePath, typeTo)
			if mainTargetFieldName != "" {
				if _, found := typeTo.FieldByName(mainTargetFieldName); found {
					toFieldVal := valTo.FieldByName(mainTargetFieldName)
					if toFieldVal.CanSet() {
						tflog.SubsystemTrace(ctx, subsystemName, "Mapping Items/Quantity to main target field", map[string]any{
							"source_field": sourceFieldName,
							"target_field": mainTargetFieldName,
						})

						// Only process this for the Items field, skip Quantity
						if sourceFieldName == "Items" {
							diags.Append(flattener.convertXMLWrapperFieldToCollection(ctx, sourcePath, sourceStructVal, targetPath.AtName(mainTargetFieldName), toFieldVal)...)
							if diags.HasError() {
								return diags
							}
						}
					}
				}
			}
		} else {
			// Other fields should map directly by name
			if _, found := typeTo.FieldByName(sourceFieldName); found {
				toFieldVal := valTo.FieldByName(sourceFieldName)
				if toFieldVal.CanSet() {
					tflog.SubsystemTrace(ctx, subsystemName, "Mapping additional field by name", map[string]any{
						"source_field": sourceFieldName,
						"target_field": sourceFieldName,
					})

					// Convert the source field to target field
					diags.Append(flattener.convertXMLWrapperFieldToCollection(ctx, sourcePath.AtName(sourceFieldName), sourceFieldVal, targetPath.AtName(sourceFieldName), toFieldVal)...)
					if diags.HasError() {
						return diags
					}
				}
			}
		}
	}

	return diags
}

// findMainTargetFieldForSplit determines which target field should receive the Items/Quantity from the source
func (flattener autoFlattener) findMainTargetFieldForSplit(sourcePath path.Path, typeTo reflect.Type) string {
	sourcePathStr := sourcePath.String()
	if lastDot := strings.LastIndex(sourcePathStr, "."); lastDot >= 0 {
		sourceFieldName := sourcePathStr[lastDot+1:]

		// Use fuzzy field finder for proper singular/plural and case matching
		dummySourceType := reflect.StructOf([]reflect.StructField{{Name: sourceFieldName, Type: reflect.TypeOf(""), PkgPath: ""}})
		if targetField, ok := (&fuzzyFieldFinder{}).findField(context.Background(), sourceFieldName, dummySourceType, typeTo, flattener); ok {
			return targetField.Name
		}
	}

	return ""
}

// convertXMLWrapperFieldToCollection converts a source field (either XML wrapper or simple field) to a target collection
func (flattener autoFlattener) convertXMLWrapperFieldToCollection(ctx context.Context, sourcePath path.Path, sourceFieldVal reflect.Value, targetPath path.Path, toFieldVal reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	// Check if source is an XML wrapper struct (has Items/Quantity)
	if sourceFieldVal.Kind() == reflect.Struct && isXMLWrapperStruct(sourceFieldVal.Type()) {
		tflog.SubsystemTrace(ctx, subsystemName, "Converting XML wrapper struct to collection", map[string]any{
			"source_type": sourceFieldVal.Type().String(),
		})

		// Use existing XML wrapper flatten logic
		if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
			diags.Append(flattener.xmlWrapperFlatten(ctx, sourceFieldVal, valTo.Type(ctx), toFieldVal, "Items")...)
		}
	} else if sourceFieldVal.Kind() == reflect.Pointer && !sourceFieldVal.IsNil() && isXMLWrapperStruct(sourceFieldVal.Type().Elem()) {
		tflog.SubsystemTrace(ctx, subsystemName, "Converting pointer to XML wrapper struct to collection", map[string]any{
			"source_type": sourceFieldVal.Type().String(),
		})

		// Use existing XML wrapper flatten logic
		if valTo, ok := toFieldVal.Interface().(attr.Value); ok {
			diags.Append(flattener.xmlWrapperFlatten(ctx, sourceFieldVal.Elem(), valTo.Type(ctx), toFieldVal, "Items")...)
		}
	} else {
		tflog.SubsystemTrace(ctx, subsystemName, "Converting non-XML wrapper field", map[string]any{
			"source_type": sourceFieldVal.Type().String(),
		})

		// For non-XML wrapper fields, use regular conversion
		opts := fieldOpts{}
		diags.Append(flattener.convert(ctx, sourcePath, sourceFieldVal, targetPath, toFieldVal, opts)...)
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
func (flattener autoFlattener) xmlWrapperFlattenRule2(ctx context.Context, vFrom reflect.Value, tTo fwtypes.NestedObjectCollectionType, vTo reflect.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	if !isXMLWrapperStruct(vFrom.Type()) {
		diags.Append(DiagFlatteningIncompatibleTypes(vFrom.Type(), reflect.TypeOf(vTo.Interface())))
		return diags
	}

	// Check if wrapper is in empty state (Enabled=false, Items empty/nil)
	// If so, return null instead of creating a block
	enabledField := vFrom.FieldByName("Enabled")
	itemsField := vFrom.FieldByName("Items")
	
	if enabledField.IsValid() && itemsField.IsValid() {
		// Check if Enabled is false and Items is empty
		isEnabled := true
		if enabledField.Kind() == reflect.Pointer && !enabledField.IsNil() {
			isEnabled = enabledField.Elem().Bool()
		}
		
		itemsEmpty := itemsField.IsNil() || (itemsField.Kind() == reflect.Slice && itemsField.Len() == 0)
		
		if !isEnabled && itemsEmpty {
			tflog.SubsystemTrace(ctx, subsystemName, "XML wrapper is empty (Enabled=false, Items empty), returning null for Rule 2")
			nullVal, d := tTo.NullValue(ctx)
			diags.Append(d...)
			if !diags.HasError() {
				vTo.Set(reflect.ValueOf(nullVal))
			}
			return diags
		}
	}

	nestedObjPtr, d := tTo.NewObjectPtr(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return diags
	}

	nestedObjValue := reflect.ValueOf(nestedObjPtr).Elem()

	// Map Items field
	if itemsField := vFrom.FieldByName("Items"); itemsField.IsValid() {
		if targetField := nestedObjValue.FieldByName("Items"); targetField.IsValid() && targetField.CanAddr() {
			if targetAttr, ok := targetField.Addr().Interface().(attr.Value); ok {
				diags.Append(autoFlattenConvert(ctx, itemsField.Interface(), targetAttr, flattener)...)
			}
		}
	}

	// Map all other fields (skip Items and Quantity)
	for i := 0; i < vFrom.NumField(); i++ {
		sourceField := vFrom.Field(i)
		fieldName := vFrom.Type().Field(i).Name

		if fieldName == "Items" || fieldName == "Quantity" {
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
