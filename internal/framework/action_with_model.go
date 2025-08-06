// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"fmt"

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
func (a *ActionWithModel[T]) ValidateModel(ctx context.Context, schema *schema.UnlinkedSchema) diag.Diagnostics {
	fmt.Printf("DEBUG: ActionWithModel.ValidateModel() called with schema having %d attributes\n", len(schema.Attributes))

	var diags diag.Diagnostics
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}

	diags.Append(a.validateModel(ctx, &state)...)

	fmt.Printf("DEBUG: ActionWithModel.ValidateModel() completed, has errors: %t\n", diags.HasError())
	if diags.HasError() {
		for _, d := range diags {
			fmt.Printf("DEBUG: ActionWithModel.ValidateModel() error: %s - %s\n", d.Summary(), d.Detail())
		}
	}

	return diags
}

type ActionValidateModel interface {
	ValidateModel(ctx context.Context, schema *schema.UnlinkedSchema) diag.Diagnostics
}
