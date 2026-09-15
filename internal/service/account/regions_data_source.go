// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package account

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/account"
	awstypes "github.com/aws/aws-sdk-go-v2/service/account/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_account_regions", name="Regions")
func newDataSourceRegions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceRegions{}, nil
}

const (
	DSNameRegions = "Regions Data Source"
)

type dataSourceRegions struct {
	framework.DataSourceWithModel[dataSourceRegionsModel]
}

func (d *dataSourceRegions) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"region_opt_status_contains": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				CustomType:  fwtypes.ListOfStringEnumType[awstypes.RegionOptStatus](),
				ElementType: types.StringType,
			},
			"regions": framework.DataSourceComputedListOfObjectAttribute[regionsModel](ctx),
		},
	}
}

func (d *dataSourceRegions) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().AccountClient(ctx)

	var data dataSourceRegionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input account.ListRegionsInput
	resp.Diagnostics.Append(flex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var output account.ListRegionsOutput
	paginator := account.NewListRegionsPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Account, create.ErrActionReading, DSNameRegions, data.AccountID.String(), err),
				err.Error(),
			)
			return
		}
		if page == nil {
			break
		}
		if len(page.Regions) > 0 {
			output.Regions = append(output.Regions, page.Regions...)
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceRegionsModel struct {
	AccountID               types.String                                       `tfsdk:"account_id"`
	RegionOptStatusContains fwtypes.ListOfStringEnum[awstypes.RegionOptStatus] `tfsdk:"region_opt_status_contains"`
	Regions                 fwtypes.ListNestedObjectValueOf[regionsModel]      `tfsdk:"regions"`
}

type regionsModel struct {
	RegionName      types.String                                 `tfsdk:"region_name"`
	RegionOptStatus fwtypes.StringEnum[awstypes.RegionOptStatus] `tfsdk:"region_opt_status"`
}
