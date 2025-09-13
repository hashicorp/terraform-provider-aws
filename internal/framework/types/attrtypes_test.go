// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestAttributeTypes(t *testing.T) {
	t.Parallel()

	type struct1 struct{}

	type struct2 struct {
		ARN             types.String `tfsdk:"arn"`
		ID              types.Int64  `tfsdk:"id"`
		IncludeProperty types.Bool   `tfsdk:"include_property"`
	}

	type struct3 struct {
		F1 types.String `tfsdk:"f1"`
		struct2
		F2 types.Int32 `tfsdk:"f2"`
	}

	ctx := context.Background()
	got := fwtypes.AttributeTypesMust[struct1](ctx)
	wanted := map[string]attr.Type{}

	if diff := cmp.Diff(got, wanted); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}

	_, err := fwtypes.AttributeTypes[int](ctx)

	if err == nil {
		t.Fatalf("expected error")
	}

	got, err = fwtypes.AttributeTypes[struct2](ctx)

	if err != nil {
		t.Fatalf("unexpected error")
	}

	wanted = map[string]attr.Type{
		"arn":              types.StringType,
		"id":               types.Int64Type,
		"include_property": types.BoolType,
	}

	if diff := cmp.Diff(got, wanted); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}

	got, err = fwtypes.AttributeTypes[struct3](ctx)
	if err != nil {
		t.Fatalf("unexpected error")
	}

	wanted = map[string]attr.Type{
		"f1":               types.StringType,
		"arn":              types.StringType,
		"id":               types.Int64Type,
		"include_property": types.BoolType,
		"f2":               types.Int32Type,
	}

	if diff := cmp.Diff(got, wanted); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}
