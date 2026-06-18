// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package uxc

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/uxc"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_uxc_services", name="Services")
func newServicesDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &servicesDataSource{}, nil
}

type servicesDataSource struct {
	framework.DataSourceWithModel[servicesDataSourceModel]
}

func (d *servicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"services": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *servicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().UXCClient(ctx)

	var data servicesDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input uxc.ListServicesInput
	var allServices []string
	paginator := uxc.NewListServicesPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		if page == nil {
			break
		}
		allServices = append(allServices, page.Services...)
	}

	data.Services = fwflex.FlattenFrameworkStringValueListOfString(ctx, allServices)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type servicesDataSourceModel struct {
	Services fwtypes.ListOfString `tfsdk:"services"`
}
