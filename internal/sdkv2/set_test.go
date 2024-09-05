// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"testing"
)

func TestStringCaseInsensitiveSetFunc(t *testing.T) {
	t.Parallel()

	v1 := "testInG"
	v2 := "TestiNG"
	v3 := "test1ng"
	f := StringCaseInsensitiveSetFunc

	if f(v1) != f(v2) {
		t.Errorf("expected equal")
	}
	if f(v1) == f(v3) {
		t.Errorf("expected not equal")
	}
}

func TestSimpleSchemaSetFuncNil(t *testing.T) {
	t.Parallel()

	var v interface{}
	f := SimpleSchemaSetFunc("key1", "key3", "key4")

	if got, want := f(v), 0; got != want {
		t.Errorf("SimpleSchemaSetFunc(%q) err %q, want %q", v, got, want)
	}
}

func TestSimpleSchemaSetFunc(t *testing.T) {
	t.Parallel()

	v1 := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": 3,
		"key4": true,
	}
	v2 := map[string]interface{}{
		"key1": "value1",
		"key2": "value2-new",
		"key3": 3,
		"key4": true,
	}
	v3 := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": 4,
		"key4": true,
	}
	f := SimpleSchemaSetFunc("key1", "key3", "key4")

	if f(v1) != f(v2) {
		t.Errorf("expected equal")
	}
	if f(v1) == f(v3) {
		t.Errorf("expected not equal")
	}
}
