// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_account_primary_contact", name="Primary Contact")
func newPrimaryContactDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &primaryContactDataSOurce{}, nil
}

type primaryContactDataSOurce struct {
	framework.DataSourceWithModel[primaryContactDataSourceModel]
}

func (d *primaryContactDataSOurce) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"address_line_1": schema.StringAttribute{
				Computed: true,
			},
			"address_line_2": schema.StringAttribute{
				Computed: true,
			},
			"address_line_3": schema.StringAttribute{
				Computed: true,
			},
			"city": schema.StringAttribute{
				Computed: true,
			},
			"company_name": schema.StringAttribute{
				Computed: true,
			},
			"country_code": schema.StringAttribute{
				Computed: true,
			},
			"district_or_county": schema.StringAttribute{
				Computed: true,
			},
			"full_name": schema.StringAttribute{
				Computed: true,
			},
			"phone_number": schema.StringAttribute{
				Computed: true,
			},
			"postal_code": schema.StringAttribute{
				Computed: true,
			},
			"state_or_region": schema.StringAttribute{
				Computed: true,
			},
			"website_url": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *primaryContactDataSOurce) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data primaryContactDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().AccountClient(ctx)

	output, err := findContactInformation(ctx, conn, fwflex.StringValueFromFramework(ctx, data.AccountID))

	if err != nil {
		response.Diagnostics.AddError("reading Account Primary Contact", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type primaryContactDataSourceModel struct {
	AccountID        types.String `tfsdk:"account_id"`
	AddressLine1     types.String `tfsdk:"address_line_1"`
	AddressLine2     types.String `tfsdk:"address_line_2"`
	AddressLine3     types.String `tfsdk:"address_line_3"`
	City             types.String `tfsdk:"city"`
	CompanyName      types.String `tfsdk:"company_name"`
	CountryCode      types.String `tfsdk:"country_code"`
	DistrictOrCounty types.String `tfsdk:"district_or_county"`
	FullName         types.String `tfsdk:"full_name"`
	PhoneNumber      types.String `tfsdk:"phone_number"`
	PostalCode       types.String `tfsdk:"postal_code"`
	StateOrRegion    types.String `tfsdk:"state_or_region"`
	WebsiteURL       types.String `tfsdk:"website_url"`
}
