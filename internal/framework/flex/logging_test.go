// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
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

func flatteningLogLine(sourceType, targetType reflect.Type) map[string]any {
	return infoLogLine("Flattening", sourceType, targetType)
}

func ignoredFieldLogLine(sourceType reflect.Type, sourceFieldName string) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping ignored field",
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: sourceFieldName,
	}
}

func mapBlockKeyFieldLogLine(sourceType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Skipping map block key",
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: MapBlockKey,
	}
}

func matchedFieldsLogLine(sourceType reflect.Type, sourceFieldName string, targetType reflect.Type, targetFieldName string) map[string]any {
	return map[string]any{
		"@level":                  hclog.Trace.String(),
		"@module":                 logModule,
		"@message":                "Matched fields",
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: sourceFieldName,
		logAttrKeyTargetType:      fullTypeName(targetType),
		logAttrKeyTargetFieldname: targetFieldName,
	}
}

func noCorrespondingFieldLogLine(sourceType reflect.Type, sourceFieldName string, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":                  hclog.Debug.String(),
		"@module":                 logModule,
		"@message":                "No corresponding field",
		logAttrKeySourceType:      fullTypeName(sourceType),
		logAttrKeySourceFieldname: sourceFieldName,
		logAttrKeyTargetType:      fullTypeName(targetType),
	}
}

func infoLogLine(message string, sourceType, targetType reflect.Type) map[string]any {
	return map[string]any{
		"@level":             hclog.Info.String(),
		"@module":            logModule,
		"@message":           message,
		logAttrKeySourceType: fullTypeName(sourceType),
		logAttrKeyTargetType: fullTypeName(targetType),
	}
}
