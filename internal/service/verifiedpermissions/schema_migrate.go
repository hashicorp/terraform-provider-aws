// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func schemaSchemaV0() schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"namespaces": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"policy_store_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"definition": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				Attributes: map[string]schema.Attribute{
					names.AttrValue: schema.StringAttribute{
						CustomType: jsontypes.NormalizedType{},
						Required:   true,
					},
				},
			},
		},
	}
}

type resourceSchemaDataV0 struct {
	ID            types.String        `tfsdk:"id"`
	Definition    types.Object        `tfsdk:"definition"`
	Namespaces    fwtypes.SetOfString `tfsdk:"namespaces"`
	PolicyStoreID types.String        `tfsdk:"policy_store_id"`
}

func upgradeSchemaStateFromV0(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var schemaDataV0 resourceSchemaDataV0
	response.Diagnostics.Append(request.State.Get(ctx, &schemaDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	schemaDataV1 := schemaResourceModel{
		ID:            schemaDataV0.ID,
		Definition:    upgradeDefinitionStateFromV0(ctx, schemaDataV0.Definition, &response.Diagnostics),
		Namespaces:    schemaDataV0.Namespaces,
		PolicyStoreID: schemaDataV0.PolicyStoreID,
	}

	response.Diagnostics.Append(response.State.Set(ctx, schemaDataV1)...)
}

func upgradeDefinitionStateFromV0(ctx context.Context, old types.Object, diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[definitionData] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[definitionData](ctx)
	}

	var definitionDataV0 definitionData
	diags.Append(old.As(ctx, &definitionDataV0, basetypes.ObjectAsOptions{})...)

	newList := []definitionData{
		{
			Value: definitionDataV0.Value,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newList)
	diags.Append(d...)

	return result
}
