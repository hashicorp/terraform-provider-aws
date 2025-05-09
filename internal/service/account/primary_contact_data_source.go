// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package account

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_account_primary_contact", name="Primary Contact")
func newDataSourcePrimaryContact(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourcePrimaryContact{}, nil
}

const (
	DSNamePrimaryContact = "Primary Contact Data Source"
)

type dataSourcePrimaryContact struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourcePrimaryContact) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
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

func (d *dataSourcePrimaryContact) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AccountClient(ctx)

	var data dataSourcePrimaryContactModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AccountId.IsNull() {
		data.AccountId = types.StringValue("")
	}

	output, err := findContactInformation(ctx, conn, data.AccountId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Account, create.ErrActionReading, DSNamePrimaryContact, data.AccountId.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourcePrimaryContactModel struct {
	AccountId        types.String `tfsdk:"account_id"`
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
	WebsiteUrl       types.String `tfsdk:"website_url"`
}
