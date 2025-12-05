// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"reflect"
	"testing"
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
