// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_iam_role_policies", name="Role Policies")
func newRolePoliciesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &rolePoliciesDataSource{}, nil
}

type rolePoliciesDataSource struct {
	framework.DataSourceWithModel[rolePoliciesDataSourceModel]
}

func (d *rolePoliciesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"policy_names": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"role_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, roleNameMaxLen),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[\w+=,.@-]*$`), "must match [\\w+=,.@-]"),
				},
			},
		},
	}
}

func (d *rolePoliciesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().IAMClient(ctx)
	var data rolePoliciesDataSourceModel

	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	roleName := data.RoleName.ValueString()
	out, err := findRolePoliciesByName(ctx, conn, roleName)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, "role_name", roleName)
		return
	}

	data.PolicyNames = flex.FlattenFrameworkStringValueSetOfString(ctx, out)
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data), "role_name", roleName)
}

type rolePoliciesDataSourceModel struct {
	PolicyNames fwtypes.SetOfString `tfsdk:"policy_names"`
	RoleName    types.String        `tfsdk:"role_name"`
}
