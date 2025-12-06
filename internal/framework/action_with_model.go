// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ActionWithModel is a structure to be embedded within an Action that has a corresponding model.
type ActionWithModel[T any] struct {
	withModel[T]
	ActionWithConfigure
}

// ValidateModel validates the action's model against a schema.
func (a *ActionWithModel[T]) ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics {
	var diags diag.Diagnostics
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}

	diags.Append(a.validateModel(ctx, &state)...)

	return diags
}

type ActionValidateModel interface {
	ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics
}
