// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_opensearchserverless_access_policy", name="Access Policy")
func newAccessPolicyDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &accessPolicyDataSource{}, nil
}

const (
	DSNameAccessPolicy = "Access Policy Data Source"
)

type accessPolicyDataSource struct {
	framework.DataSourceWithModel[accessPolicyDataSourceModel]
}

func (d *accessPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Description: "Description of the policy. Typically used to store information about the permissions defined in the policy.",
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Description: "Name of the policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				Description: "JSON policy document to use as the content for the new policy.",
				Computed:    true,
			},
			"policy_version": schema.StringAttribute{
				Description: "Version of the policy.",
				Computed:    true,
			},
			names.AttrType: schema.StringAttribute{
				Description: "Type of access policy. Must be `data`.",
				Required:    true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.AccessPolicyType](),
				},
			},
		},
	}
}
func (d *accessPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data accessPolicyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAccessPolicyByNameAndType(ctx, conn, data.Name.ValueString(), data.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.OpenSearchServerless, create.ErrActionReading, DSNameAccessPolicy, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = flex.StringToFramework(ctx, out.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type accessPolicyDataSourceModel struct {
	framework.WithRegionModel
	Description   types.String `tfsdk:"description"`
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Policy        types.String `tfsdk:"policy"`
	PolicyVersion types.String `tfsdk:"policy_version"`
	Type          types.String `tfsdk:"type"`
}
