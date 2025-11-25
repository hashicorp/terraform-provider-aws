// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// DataSourceWithModel is a structure to be embedded within a DataSource that has a corresponding model.
type DataSourceWithModel[T any] struct {
	withModel[T]
	DataSourceWithConfigure
}

// ValidateModel validates the data source's model against a schema.
func (d *DataSourceWithModel[T]) ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics {
	var diags diag.Diagnostics
	state := tfsdk.State{
		Raw:    tftypes.NewValue(schema.Type().TerraformType(ctx), nil),
		Schema: schema,
	}

	diags.Append(d.validateModel(ctx, &state)...)

	return diags
}

type DataSourceValidateModel interface {
	ValidateModel(ctx context.Context, schema *schema.Schema) diag.Diagnostics
}

// withModel is a structure to be embedded within a DataSource, EphemeralResource, or Resource that has a corresponding model.
type withModel[T any] struct{}

// validateModel validates the data source's model against a schema.
func (d *withModel[T]) validateModel(ctx context.Context, state *tfsdk.State) diag.Diagnostics {
	var diags diag.Diagnostics
	var data T

	diags.Append(fwtypes.NullOutObjectPtrFields(ctx, &data)...)
	if diags.HasError() {
		return diags
	}
	diags.Append(state.Set(ctx, &data)...)

	return diags
}
