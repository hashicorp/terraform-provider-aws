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
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_uxc_services", name="Services")
func newDataSourceServices(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServices{}, nil
}

const (
	DSNameServices = "Services"
)

type dataSourceServices struct {
	framework.DataSourceWithModel[dataSourceServicesModel]
}

func (d *dataSourceServices) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"services": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *dataSourceServices) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().UXCClient(ctx)

	var data dataSourceServicesModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var allServices []string
	paginator := uxc.NewListServicesPaginator(conn, &uxc.ListServicesInput{})
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

	data.ID = flex.StringValueToFramework(ctx, d.Meta().AccountID(ctx))

	servicesList, listDiags := types.ListValueFrom(ctx, types.StringType, allServices)
	smerr.AddEnrich(ctx, &resp.Diagnostics, listDiags)
	data.Services = fwtypes.ListOfString{ListValue: servicesList}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type dataSourceServicesModel struct {
	ID       types.String         `tfsdk:"id"`
	Services fwtypes.ListOfString `tfsdk:"services"`
}
