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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
				Computed:    true,
				ElementType: types.StringType,
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
	var data rolePoliciesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().IAMClient(ctx)
	roleName := data.RoleName.ValueString()

	out, err := findRolePoliciesByName(ctx, conn, roleName)
	if retry.NotFound(err) {
		response.Diagnostics.AddError("reading IAM Role Policies", "role not found: "+roleName)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("reading IAM Role Policies", err.Error())
		return
	}

	if out == nil {
		out = []string{}
	}

	policyNames, diags := types.SetValueFrom(ctx, types.StringType, out)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	data.PolicyNames = policyNames
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type rolePoliciesDataSourceModel struct {
	RoleName    types.String `tfsdk:"role_name"`
	PolicyNames types.Set    `tfsdk:"policy_names"`
}
