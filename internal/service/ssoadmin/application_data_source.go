// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ssoadmin_application", name="Application")
func newApplicationDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &applicationDataSource{}, nil
}

type applicationDataSource struct {
	framework.DataSourceWithModel[applicationDataSourceModel]
}

func (d *applicationDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
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
			"portal_options": framework.DataSourceComputedListOfObjectAttribute[portalOptionsModel](ctx),
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *applicationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data applicationDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SSOAdminClient(ctx)

	output, err := findApplicationByID(ctx, conn, data.ApplicationARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SSO Application (%s)", data.ApplicationARN.ValueString()), err.Error())

		return
	}

	// Skip writing to state if only the visibilty attribute is returned
	// to avoid a nested computed attribute causing a diff.
	if output.PortalOptions != nil && output.PortalOptions.SignInOptions == nil {
		output.PortalOptions = nil
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type applicationDataSourceModel struct {
	framework.WithRegionModel
	ApplicationAccount     types.String                                        `tfsdk:"application_account"`
	ApplicationARN         fwtypes.ARN                                         `tfsdk:"application_arn"`
	ApplicationProviderARN fwtypes.ARN                                         `tfsdk:"application_provider_arn"`
	Description            types.String                                        `tfsdk:"description"`
	ID                     types.String                                        `tfsdk:"id"`
	InstanceARN            fwtypes.ARN                                         `tfsdk:"instance_arn"`
	Name                   types.String                                        `tfsdk:"name"`
	PortalOptions          fwtypes.ListNestedObjectValueOf[portalOptionsModel] `tfsdk:"portal_options"`
	Status                 types.String                                        `tfsdk:"status"`
}
