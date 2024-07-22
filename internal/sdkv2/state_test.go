// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNormalizeJsonStringSchemaStateFunc(t *testing.T) { // nosemgrep:ci.caps2-in-func-name
	t.Parallel()

	var input interface{} = `{ "key1": "value1", "key2": 42}`
	want := `{"key1":"value1","key2":42}`

	got := NormalizeJsonStringSchemaStateFunc(input)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}

func TestToUpperSchemaStateFunc(t *testing.T) {
	t.Parallel()

	var input interface{} = "in-state"
	want := "IN-STATE"

	got := ToUpperSchemaStateFunc(input)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+want, -got): %s", diff)
	}
}
