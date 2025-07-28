// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// EphemeralResourceWithModel is a structure to be embedded within an EphemeralResource that has a corresponding model.
type EphemeralResourceWithModel[T any] struct {
	withModel[T]
	EphemeralResourceWithConfigure
}

// ValidateModel validates the ephemeral resource's model against a schema.
func (d *EphemeralResourceWithModel[T]) ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics {
	var diags diag.Diagnostics
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}

	diags.Append(d.validateModel(ctx, &state)...)

	return diags
}

type EphemeralResourceValidateModel interface {
	ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics
}
