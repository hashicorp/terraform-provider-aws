// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

func TestAttributeTypes(t *testing.T) {
	t.Parallel()

	type struct1 struct{}
	type struct2 struct {
		ARN             types.String `tfsdk:"arn"`
		ID              types.Int64  `tfsdk:"id"`
		IncludeProperty types.Bool   `tfsdk:"include_property"`
	}

	ctx := context.Background()
	got := flex.AttributeTypesMust[struct1](ctx)
	wanted := map[string]attr.Type{}

	if diff := cmp.Diff(got, wanted); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}

	_, err := flex.AttributeTypes[int](ctx)

	if err == nil {
		t.Fatalf("expected error")
	}

	got, err = flex.AttributeTypes[struct2](ctx)

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
}
