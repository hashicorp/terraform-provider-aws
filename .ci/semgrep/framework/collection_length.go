// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// IMPORTANT: The "fixed" file must not be formatted with gofmt.
// Semgrep does not handle formatting of multiline fixes in Go correctly.

package main

import (
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// collection-length-not-null tests

func testNotNull1(x basetypes.ListValue) bool {
	// ruleid: collection-length-not-null
	return !x.IsNull() && len(x.Elements()) > 0
}

func testNotNull2(x basetypes.SetValue) {
	// ruleid: collection-length-not-null
	if !x.IsNull() && len(x.Elements()) > 0 {
		_ = x
	}
}

func testNotNull3(x basetypes.MapValue) {
	// ruleid: collection-length-not-null
	if !x.IsNull() && len(x.Elements()) > 0 {
		_ = x
	}
}

func testNotNullOK1(x basetypes.ListValue) bool {
	// ok: collection-length-not-null
	return x.Length(fwtypes.CollectionLengthUnhandledAsZero) > 0
}

func testNotNullOK2(x basetypes.ListValue) {
	// ok: collection-length-not-null
	if !x.IsNull() && x.IsUnknown() {
		_ = x
	}
}

// collection-length-is-null tests

func testIsNull1(x basetypes.ListValue) bool {
	// ruleid: collection-length-is-null
	return x.IsNull() || len(x.Elements()) == 0
}

func testIsNull2(x basetypes.SetValue) {
	// ruleid: collection-length-is-null
	if x.IsNull() || len(x.Elements()) == 0 {
		_ = x
	}
}

func testIsNull3(x basetypes.MapValue) {
	// ruleid: collection-length-is-null
	if x.IsNull() || len(x.Elements()) == 0 {
		_ = x
	}
}

func testIsNullOK1(x basetypes.ListValue) bool {
	// ok: collection-length-is-null
	return x.Length(fwtypes.CollectionLengthUnhandledAsZero) == 0
}

func testIsNullOK2(x basetypes.ListValue) {
	// ok: collection-length-is-null
	if x.IsNull() || x.IsUnknown() {
		_ = x
	}
}

// collection-length tests

func testLength1(x basetypes.ListValue) bool {
	// ruleid: collection-length
	return len(x.Elements()) > 0
}

func testLength2(x basetypes.SetValue) int {
	// ruleid: collection-length
	return len(x.Elements())
}

func testLength3(x basetypes.MapValue) {
	logAttrs := map[string]any{
		// ruleid: collection-length
		"size": len(x.Elements()),
	}
	_ = logAttrs
}

func testLength4(x basetypes.ListValue) []string {
	// ruleid: collection-length
	out := make([]string, len(x.Elements()))
	return out
}

func testLength5(x basetypes.ListValue) int32 {
	// ruleid: collection-length
	return int32(len(x.Elements()))
}

func testLengthOK1(x basetypes.ListValue) int {
	// ok: collection-length
	return x.Length(fwtypes.CollectionLengthUnhandledAsZero)
}

func testLengthOK2(x []string) int {
	// ok: collection-length
	return len(x)
}
