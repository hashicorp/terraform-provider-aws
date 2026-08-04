// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/action/timeouts"
	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ActionWithModel is a structure to be embedded within an Action that has a corresponding model.
type ActionWithModel[T any] struct {
	withModel[T]
	ActionWithConfigure
}

// ValidateModel validates the action's model against a schema.
func (a *ActionWithModel[T]) ValidateModel(ctx context.Context, schema *actionschema.Schema) diag.Diagnostics {
	// An actions's model will contain a field like:
	// 	Timeouts timeouts.Value `tfsdk:"timeouts"`
	// Swap a blank timeouts block into the schema, preventing errors like:
	//   Expected framework type from provider logic: timeouts.Type / underlying type: tftypes.Object["invoke":tftypes.String]
	//   Received framework type from provider logic: timeouts.Type / underlying type: tftypes.Object[]
	if schema.Blocks[names.AttrTimeouts] != nil {
		s := *schema
		s.Blocks = maps.Clone(s.Blocks)
		s.Blocks[names.AttrTimeouts] = actionschema.SingleNestedBlock{
			Attributes: map[string]actionschema.Attribute{},
			CustomType: timeouts.Type{
				ObjectType: types.ObjectType{
					AttrTypes: map[string]attr.Type{},
				},
			},
		}
		schema = &s
	}

	var diags diag.Diagnostics
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}

	diags.Append(a.validateModel(ctx, &state)...)

	return diags
}

type ActionValidateModel interface {
	ValidateModel(ctx context.Context, schema *actionschema.Schema) diag.Diagnostics
}
