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
	expected := "<nil>"
	result := fullTypeName(nil)

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_primitive(t *testing.T) {
	expected := "string"
	result := fullTypeName(reflect.TypeFor[string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_type(t *testing.T) {
	expected := "github.com/hashicorp/terraform-provider-aws/internal/framework/flex.TestFlex00"
	result := fullTypeName(reflect.TypeFor[TestFlex00]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_pointerToPrimitive(t *testing.T) {
	expected := "*string"
	result := fullTypeName(reflect.TypeFor[*string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_pointerToType(t *testing.T) {
	expected := "*github.com/hashicorp/terraform-provider-aws/internal/framework/flex.TestFlex00"
	result := fullTypeName(reflect.TypeFor[*TestFlex00]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_sliceOfPrimitive(t *testing.T) {
	expected := "[]string"
	result := fullTypeName(reflect.TypeFor[[]string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_sliceOfType(t *testing.T) {
	expected := "[]github.com/hashicorp/terraform-provider-aws/internal/framework/flex.TestFlex00"
	result := fullTypeName(reflect.TypeFor[[]TestFlex00]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_sliceOfPointerToPrimitive(t *testing.T) {
	expected := "[]*string"
	result := fullTypeName(reflect.TypeFor[[]*string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_sliceOfPointerToType(t *testing.T) {
	expected := "[]*github.com/hashicorp/terraform-provider-aws/internal/framework/flex.TestFlex00"
	result := fullTypeName(reflect.TypeFor[[]*TestFlex00]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_mapPrimitiveKeyPrimitiveValue(t *testing.T) {
	expected := "map[string]string"
	result := fullTypeName(reflect.TypeFor[map[string]string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_mapTypedKeyPrimitiveValue(t *testing.T) {
	expected := "map[github.com/hashicorp/terraform-provider-aws/internal/framework/flex.TestEnum]string"
	result := fullTypeName(reflect.TypeFor[map[TestEnum]string]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

func TestFullTypeName_mapPrimitiveKeyTypedValue(t *testing.T) {
	expected := "map[string]github.com/hashicorp/terraform-provider-aws/internal/framework/flex.TestEnum"
	result := fullTypeName(reflect.TypeFor[map[string]TestEnum]())

	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

const (
	logModule = "provider." + subsystemName
)

func expandingLogLine(sourceType, targetType reflect.Type) map[string]any {
	return infoLogLine("Expanding", sourceType, targetType)
}

func expandingWithPathLogLine(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return infoWithPathLogLine("Expanding", sourcePath, sourceType, targetPath, targetType)
}

func flatteningLogLine(sourceType, targetType reflect.Type) map[string]any {
	return infoLogLine("Flattening", sourceType, targetType)
}

func flatteningWithPathLogLine(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return infoWithPathLogLine("Flattening", sourcePath, sourceType, targetPath, targetType)
}

func convertingLogLine(sourceType, targetType reflect.Type) map[string]any {
	return logInfo("Converting", map[string]any{
		logAttrKeySourcePath: "",
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetPath: "",
		logAttrKeyTargetType: fullTypeName(targetType),
	})
}

func convertingWithPathLogLine(sourceFieldPath string, sourceType reflect.Type, targetFieldPath string, targetType reflect.Type) map[string]any {
	return logInfo("Converting", map[string]any{
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeySourcePath: sourceFieldPath,
		logAttrKeyTargetType: fullTypeName(targetType),
		logAttrKeyTargetPath: targetFieldPath,
	})
}

func ignoredFieldLogLine(sourceType reflect.Type, sourceFieldName string, targetType reflect.Type) map[string]any {
	return ignoredFieldWithPathLogLine(
		"", sourceType, sourceFieldName,
		"", targetType,
	)
}

func ignoredFieldWithPathLogLine(sourcePath string, sourceType reflect.Type, sourceFieldName string, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping ignored field",
		logAttrKeySourcePath:      sourcePath,
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: sourceFieldName,
		logAttrKeyTargetPath:      targetPath,
		logAttrKeyTargetType:      fullTypeName(targetType),
	}
}

func mapBlockKeyFieldLogLine(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping map block key",
		logAttrKeySourcePath:      sourcePath,
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: MapBlockKey,
		logAttrKeyTargetPath:      targetPath,
		logAttrKeyTargetType:      fullTypeName(targetType),
	}
}

func matchedFieldsLogLine(sourceFieldName string, sourceType reflect.Type, targetFieldName string, targetType reflect.Type) map[string]any {
	return matchedFieldsWithPathLogLine(
		"", sourceFieldName, sourceType,
		"", targetFieldName, targetType,
	)
}

func matchedFieldsWithPathLogLine(sourcePath, sourceFieldName string, sourceType reflect.Type, targetPath, targetFieldName string, targetType reflect.Type) map[string]any {
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

func noCorrespondingFieldLogLine(sourceType reflect.Type, sourceFieldName string, targetType reflect.Type) map[string]any {
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

func expandNullValueLogLine(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
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

func expandElementsAsLogLine(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
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

func flattenNullValueLogLine(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
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

func flattenMapNullLogLine(sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
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

func flattenMapLogLine(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
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

func flattenMapValueLogLine(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
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

func flattenNewMapValueOfLogLine(sourcePath string, sourceType reflect.Type, sourceLen int, targetPath string, targetType reflect.Type) map[string]any {
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

func infoLogLine(message string, sourceType, targetType reflect.Type) map[string]any {
	return logInfo(message, map[string]any{
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetType: fullTypeName(targetType),
	})
}

func infoWithPathLogLine(message string, sourcePath string, sourceType reflect.Type, targetPath string, targetType reflect.Type) map[string]any {
	return logInfo(message, map[string]any{
		logAttrKeySourcePath: sourcePath,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetType: fullTypeName(targetType),
		logAttrKeyTargetPath: targetPath,
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
