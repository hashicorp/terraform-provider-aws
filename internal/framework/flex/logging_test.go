// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"maps"
	"reflect"
	"testing"

	"github.com/hashicorp/go-hclog"
)

func TestFullTypeName_nil(t *testing.T) {
	t.Parallel()

	expected := "<nil>"
	result := fullTypeName(nil)

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_primitive(t *testing.T) {
	t.Parallel()

	expected := "string"
	result := fullTypeName(reflect.TypeFor[string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_type(t *testing.T) {
	t.Parallel()

	expected := "github.com/hashicorp/terraform-provider-aws/internal/framework/flex.emptyStruct"
	result := fullTypeName(reflect.TypeFor[emptyStruct]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_pointerToPrimitive(t *testing.T) {
	t.Parallel()

	expected := "*string"
	result := fullTypeName(reflect.TypeFor[*string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_pointerToType(t *testing.T) {
	t.Parallel()

	expected := "*github.com/hashicorp/terraform-provider-aws/internal/framework/flex.emptyStruct"
	result := fullTypeName(reflect.TypeFor[*emptyStruct]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_sliceOfPrimitive(t *testing.T) {
	t.Parallel()

	expected := "[]string"
	result := fullTypeName(reflect.TypeFor[[]string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_sliceOfType(t *testing.T) {
	t.Parallel()

	expected := "[]github.com/hashicorp/terraform-provider-aws/internal/framework/flex.emptyStruct"
	result := fullTypeName(reflect.TypeFor[[]emptyStruct]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_sliceOfPointerToPrimitive(t *testing.T) {
	t.Parallel()

	expected := "[]*string"
	result := fullTypeName(reflect.TypeFor[[]*string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_sliceOfPointerToType(t *testing.T) {
	t.Parallel()

	expected := "[]*github.com/hashicorp/terraform-provider-aws/internal/framework/flex.emptyStruct"
	result := fullTypeName(reflect.TypeFor[[]*emptyStruct]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_mapPrimitiveKeyPrimitiveValue(t *testing.T) {
	t.Parallel()

	expected := "map[string]string"
	result := fullTypeName(reflect.TypeFor[map[string]string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_mapTypedKeyPrimitiveValue(t *testing.T) {
	t.Parallel()

	expected := "map[github.com/hashicorp/terraform-provider-aws/internal/framework/flex.testEnum]string"
	result := fullTypeName(reflect.TypeFor[map[testEnum]string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_mapPrimitiveKeyTypedValue(t *testing.T) {
	t.Parallel()

	expected := "map[string]github.com/hashicorp/terraform-provider-aws/internal/framework/flex.testEnum"
	result := fullTypeName(reflect.TypeFor[map[string]testEnum]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

const (
	logModule = "provider." + subsystemName
)

func infoExpanding(sourceType, targetType reflect.Type) map[string]any {
	return infoLogLine("Expanding", sourceType, targetType)
}

func infoFlattening(sourceType, targetType reflect.Type) map[string]any {
	return infoLogLine("Flattening", sourceType, targetType)
}

func infoConverting(sourceType, targetType reflect.Type) map[string]any {
	return logInfo("Converting", map[string]any{
		logAttrKeySourcePath: "",
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: "",
		logAttrKeyTargetType: fullTypeName(targetType),
	})
}

func infoConvertingWithPath(sourceFieldPath string, sourceType reflect.Type, targetFieldPath string, targetType reflect.Type) map[string]any {
	return logInfo("Converting", map[string]any{
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourcePath: sourceFieldPath,
		logAttrKeyTargetType: fullTypeName(targetType),
		logAttrKeyTargetPath: targetFieldPath,
	})
}

func traceSkipIgnoredSourceField(sourceType reflect.Type, sourceFieldName string, targetType reflect.Type) map[string]any {
	return traceSkipIgnoredSourceFieldWithPath(
		"", sourceType, sourceFieldName,
		"", targetType,
	)
}

func traceSkipIgnoredSourceFieldWithPath(sourcePath string, sourceType reflect.Type, sourceFieldName string, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping ignored source field",
		logAttrKeySourcePath:      sourcePath,
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: sourceFieldName,
		logAttrKeyTargetPath:      targetPath,
		logAttrKeyTargetType:      fullTypeName(targetType),
	}
}

func traceSkipIgnoredTargetField(sourceType reflect.Type, sourceFieldName string, targetType reflect.Type, targetFieldName string) map[string]any {
	return traceSkipIgnoredTargetFieldWithPath(
		"", sourceType, sourceFieldName,
		"", targetType, targetFieldName,
	)
}

func traceSkipIgnoredTargetFieldWithPath(sourcePath string, sourceType reflect.Type, sourceFieldName string, targetPath string, targetType reflect.Type, targetFieldName string) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping ignored target field",
		logAttrKeySourcePath:      sourcePath,
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: sourceFieldName,
		logAttrKeyTargetPath:      targetPath,
		logAttrKeyTargetType:      fullTypeName(targetType),
		logAttrKeyTargetFieldname: targetFieldName,
	}
}

func traceSkipMapBlockKey(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping map block key",
		logAttrKeySourcePath:      sourcePath,
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: mapBlockKeyFieldName,
		logAttrKeyTargetPath:      targetPath,
		logAttrKeyTargetType:      fullTypeName(targetType),
	}
}

func traceMatchedFields(sourceFieldName string, sourceType reflect.Type, targetFieldName string, targetType reflect.Type) map[string]any {
	return traceMatchedFieldsWithPath(
		"", sourceFieldName, sourceType,
		"", targetFieldName, targetType,
	)
}

func traceMatchedFieldsWithPath(sourcePath, sourceFieldName string, sourceType reflect.Type, targetPath, targetFieldName string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Matched fields",
		logAttrKeySourcePath:      sourcePath,
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: sourceFieldName,
		logAttrKeyTargetPath:      targetPath,
		logAttrKeyTargetType:      fullTypeName(targetType),
		logAttrKeyTargetFieldname: targetFieldName,
	}
}

func debugNoCorrespondingField(sourceType reflect.Type, sourceFieldName string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Debug.String(),
		"@module":                 logModule,
		"@message":                "No corresponding field",
		logAttrKeySourcePath:      "",
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: sourceFieldName,
		logAttrKeyTargetPath:      "",
		logAttrKeyTargetType:      fullTypeName(targetType),
	}
}

func traceExpandingNullValue(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Expanding null value",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceExpandingWithElementsAs(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Expanding with ElementsAs",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourceSize: float64(sourceLen), // numbers are deserialized from JSON as float64
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceExpandingNestedObjectCollection(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Expanding nested object collection",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourceSize: float64(sourceLen), // numbers are deserialized from JSON as float64
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningNullValue(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening null value",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningWithMapNull(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening with MapNull",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningMap(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening map",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourceSize: float64(sourceLen), // numbers are deserialized from JSON as float64
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningWithMapValue(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening with MapValue",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourceSize: float64(sourceLen), // numbers are deserialized from JSON as float64
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningWithSetNull(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening with SetNull",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningWithSetValue(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening with SetValue",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourceSize: float64(sourceLen), // numbers are deserialized from JSON as float64
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningWithListNull(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening with ListNull",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningWithListValue(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening with ListValue",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourceSize: float64(sourceLen), // numbers are deserialized from JSON as float64
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningWithNullValue(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening with NullValue",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningNestedObjectCollection(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening nested object collection",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourceSize: float64(sourceLen), // numbers are deserialized from JSON as float64
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func traceFlatteningWithNewMapValueOf(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Trace.String(),
		"@module":            logModule,
		"@message":           "Flattening with NewMapValueOf",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourceSize: float64(sourceLen), // numbers are deserialized from JSON as float64
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func infoSourceImplementsFlexExpander(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Info.String(),
		"@module":            logModule,
		"@message":           "Source implements flex.Expander",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func infoSourceImplementsFlexTypedExpander(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Info.String(),
		"@module":            logModule,
		"@message":           "Source implements flex.TypedExpander",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func infoTargetImplementsFlexFlattener(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Info.String(),
		"@module":            logModule,
		"@message":           "Target implements flex.Flattener",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func infoSourceImplementsJSONStringer(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Info.String(),
		"@module":            logModule,
		"@message":           "Source implements json.JSONStringer",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorSourceDoesNotImplementAttrValue(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Source does not implement attr.Value",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorSourceIsNil(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Source is nil",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorSourceHasNoMapBlockKey(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Source has no map block key",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorTargetDoesNotImplementAttrValue(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Target does not implement attr.Value",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorTargetIsNil(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Target is nil",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorTargetIsNotPointer(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Target is not a pointer",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorTargetHasNoMapBlockKey(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Target has no map block key",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorMarshallingJSONDocument(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type, err error) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Marshalling JSON document",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
		logAttrKeyError:      err.Error(),
	}
}

func errorExpandingIncompatibleTypes(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Expanding incompatible types",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func errorFlatteningIncompatibleTypes(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Error.String(),
		"@module":            logModule,
		"@message":           "Flattening incompatible types",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func debugUsingLegacyExpander(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Debug.String(),
		"@module":            logModule,
		"@message":           "Using legacy expander",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func debugUsingLegacyFlattener(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Debug.String(),
		"@module":            logModule,
		"@message":           "Using legacy flattener",
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: targetPath,
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}

func infoLogLine(message string, sourceType, targetType reflect.Type) map[string]any {
	return logInfo(message, map[string]any{
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetType: fullTypeName(targetType),
	})
}

func logInfo(message string, attrs map[string]any) map[string]any {
	result := map[string]any{
		"@level":   hclog.Info.String(),
		"@module":  logModule,
		"@message": message,
	}
	maps.Copy(result, attrs)
	return result
}
