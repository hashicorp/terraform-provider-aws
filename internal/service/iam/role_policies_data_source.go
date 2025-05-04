// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Role Policies")
func newDataSourceRolePolicies(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceRolePolicies{}, nil
}

const (
	DSNameRolePolicies = "Role Policies Data Source"
)

type dataSourceRolePolicies struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceRolePolicies) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_iam_role_policies"
}

func (d *dataSourceRolePolicies) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"policy_names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
			"role_name": schema.StringAttribute{
				Description: "Name of the role where we should get the inline policies.",
				Required:    true,
			},
		},
	}
}

func (d *dataSourceRolePolicies) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	conn := d.Meta().IAMClient(ctx)

	var data dataSourceRolePoliciesData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var nextMarker *string
	var policyNames []string
	for {
		input := &iam.ListRolePoliciesInput{RoleName: aws.String(data.RoleName.ValueString())}
		if nextMarker != nil {
			input.Marker = nextMarker
		}

		output, err := conn.ListRolePolicies(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IAM, create.ErrActionReading, DSNameRolePolicies, data.RoleName.String(), err),
				err.Error(),
			)
			return
		}
		policyNames = append(policyNames, output.PolicyNames...)

		if !output.IsTruncated {
			break
		}
		// Specify the next marker to retrieve the next page of results
		nextMarker = output.Marker
	}

	data.PolicyNames = flex.FlattenFrameworkStringValueList(ctx, policyNames)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceRolePoliciesData struct {
	PolicyNames types.List   `tfsdk:"policy_names"`
	RoleName    types.String `tfsdk:"role_name"`
}
