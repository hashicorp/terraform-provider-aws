// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestSetZeroValueAttrFieldsToNull(t *testing.T) {
	t.Parallel()

	type nestedModel struct {
		Field1 types.String `tfsdk:"field1"`
		Field2 types.Int32  `tfsdk:"field2"`
	}

	type listModel struct {
		Name   types.String                                 `tfsdk:"name"`
		Tags   fwtypes.MapValueOf[types.String]             `tfsdk:"tags"`
		Config fwtypes.ObjectValueOf[nestedModel]           `tfsdk:"config"`
		Rules  fwtypes.ListNestedObjectValueOf[nestedModel] `tfsdk:"rules"`
		Nested nestedModel                                  `tfsdk:"nested"`
	}

	ctx := context.Background()
	data := &listModel{
		Name: types.StringValue("example"),
	}

	if diags := setZeroValueAttrFieldsToNull(ctx, data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !data.Name.Equal(types.StringValue("example")) {
		t.Fatalf("expected Name to be unchanged")
	}
	if !data.Tags.IsNull() {
		t.Errorf("expected Tags to be null")
	}
	if !data.Config.IsNull() {
		t.Errorf("expected Config to be null")
	}
	if !data.Rules.IsNull() {
		t.Errorf("expected Rules to be null")
	}
	if !data.Nested.Field1.IsNull() {
		t.Errorf("expected Nested.Field1 to be null")
	}
	if !data.Nested.Field2.IsNull() {
		t.Errorf("expected Nested.Field2 to be null")
	}
}

func TestSetZeroValueAttrFieldsToNull_NonPointer(t *testing.T) {
	t.Parallel()

	type listModel struct {
		Name types.String `tfsdk:"name"`
	}

	ctx := context.Background()
	data := listModel{}

	if diags := setZeroValueAttrFieldsToNull(ctx, data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
}

func TestSetZeroValueAttrFieldsToNull_NilPointer(t *testing.T) {
	t.Parallel()

	type listModel struct {
		Name types.String `tfsdk:"name"`
	}

	ctx := context.Background()

	var data *listModel = nil

	if diags := setZeroValueAttrFieldsToNull(ctx, data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
}

func TestSetZeroValueAttrFieldsToNull_UnexportedField(t *testing.T) {
	t.Parallel()

	type modelWithPrivateField struct {
		Public  types.String `tfsdk:"public"`
		private types.String
	}

	ctx := context.Background()
	data := &modelWithPrivateField{}

	if diags := setZeroValueAttrFieldsToNull(ctx, data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !data.Public.IsNull() {
		t.Errorf("expected Public field to be null")
	}
}

func TestSetZeroValueAttrFieldsToNull_AlreadyNull(t *testing.T) {
	t.Parallel()

	type listModel struct {
		Name types.String `tfsdk:"name"`
	}

	ctx := context.Background()
	data := &listModel{
		Name: types.StringNull(),
	}

	if diags := setZeroValueAttrFieldsToNull(ctx, data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !data.Name.IsNull() {
		t.Errorf("expected Name to remain null")
	}
}

func TestSetZeroValueAttrFieldsToNull_UnknownValue(t *testing.T) {
	t.Parallel()

	type listModel struct {
		Name types.String `tfsdk:"name"`
	}

	ctx := context.Background()
	data := &listModel{
		Name: types.StringUnknown(),
	}

	if diags := setZeroValueAttrFieldsToNull(ctx, data); diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if !data.Name.IsUnknown() {
		t.Errorf("expected Name to remain unknown")
	}
}
