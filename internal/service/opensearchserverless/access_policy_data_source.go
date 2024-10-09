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

// @FrameworkDataSource(name="Access Policy")
func newDataSourceAccessPolicy(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAccessPolicy{}, nil
}

const (
	DSNameAccessPolicy = "Access Policy Data Source"
)

type dataSourceAccessPolicy struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceAccessPolicy) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_opensearchserverless_access_policy"
}

func (d *dataSourceAccessPolicy) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				Computed: true,
			},
			"policy_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.AccessPolicyType](),
				},
			},
		},
	}
}
func (d *dataSourceAccessPolicy) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().OpenSearchServerlessClient(ctx)

	var data dataSourceAccessPolicyData
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

type dataSourceAccessPolicyData struct {
	Description   types.String `tfsdk:"description"`
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Policy        types.String `tfsdk:"policy"`
	PolicyVersion types.String `tfsdk:"policy_version"`
	Type          types.String `tfsdk:"type"`
}
