// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Application")
func newDataSourceApplication(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceApplication{}, nil
}

const (
	DSNameApplication = "Application Data Source"
)

type dataSourceApplication struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceApplication) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_ssoadmin_application"
}

func (d *dataSourceApplication) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_account": schema.StringAttribute{
				Computed: true,
			},
			"application_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"application_provider_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			"instance_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"portal_options": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"visibility": schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"sign_in_options": schema.ListNestedBlock{
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"origin": schema.StringAttribute{
										Computed: true,
									},
									"application_url": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceApplication) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SSOAdminClient(ctx)

	var data dataSourceApplicationData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findApplicationByID(ctx, conn, data.ApplicationARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSOAdmin, create.ErrActionReading, DSNameApplication, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	data.ApplicationAccount = flex.StringToFramework(ctx, out.ApplicationAccount)
	data.ApplicationARN = flex.StringToFrameworkARN(ctx, out.ApplicationArn)
	data.ApplicationProviderARN = flex.StringToFrameworkARN(ctx, out.ApplicationProviderArn)
	data.Description = flex.StringToFramework(ctx, out.Description)
	data.ID = flex.StringToFramework(ctx, out.ApplicationArn)
	data.InstanceARN = flex.StringToFrameworkARN(ctx, out.InstanceArn)
	data.Name = flex.StringToFramework(ctx, out.Name)
	data.Status = flex.StringValueToFramework(ctx, out.Status)

	portalOptions, diags := flattenPortalOptions(ctx, out.PortalOptions)
	resp.Diagnostics.Append(diags...)
	data.PortalOptions = portalOptions

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceApplicationData struct {
	ApplicationAccount     types.String `tfsdk:"application_account"`
	ApplicationARN         fwtypes.ARN  `tfsdk:"application_arn"`
	ApplicationProviderARN fwtypes.ARN  `tfsdk:"application_provider_arn"`
	Description            types.String `tfsdk:"description"`
	ID                     types.String `tfsdk:"id"`
	InstanceARN            fwtypes.ARN  `tfsdk:"instance_arn"`
	Name                   types.String `tfsdk:"name"`
	PortalOptions          types.List   `tfsdk:"portal_options"`
	Status                 types.String `tfsdk:"status"`
}
