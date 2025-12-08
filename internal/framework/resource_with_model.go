// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ResourceWithModel is a structure to be embedded within a Resource that has a corresponding model.
type ResourceWithModel[T any] struct {
	ResourceWithConfigure
	withNoOpUpdate[T]
	withModel[T]
}

// ValidateModel validates the resource's model against a schema.
func (d *ResourceWithModel[T]) ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics {
	// A resource's model will contain a field like:
	// 	Timeouts timeouts.Value `tfsdk:"timeouts"`
	// Swap a blank timeouts block into the schema, preventing errors like:
	//   Expected framework type from provider logic: timeouts.Type / underlying type: tftypes.Object["create":tftypes.String, "delete":tftypes.String]
	//   Received framework type from provider logic: timeouts.Type / underlying type: tftypes.Object[]
	if schema.Blocks[names.AttrTimeouts] != nil {
		s := *schema
		s.Blocks = maps.Clone(s.Blocks)
		s.Blocks[names.AttrTimeouts] = timeouts.Block(ctx, timeouts.Opts{})
		schema = &s
	}

	var diags diag.Diagnostics
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}

	diags.Append(d.validateModel(ctx, &state)...)

	return diags
}

type ResourceValidateModel interface {
	ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics
}
