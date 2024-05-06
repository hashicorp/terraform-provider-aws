// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Search")
func newDataSourceSearch(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceSearch{}, nil
}

const (
	DSNameSearch = "Search Data Source"
)

type dataSourceSearch struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceSearch) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_resourceexplorer2_search"
}

func (d *dataSourceSearch) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"query_string": schema.StringAttribute{
				Required: true,
			},
			names.AttrID: framework.IDAttribute(),
			"view_arn": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1011),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"resource_count": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[countData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"complete": schema.BoolAttribute{
							Computed: true,
						},
						"total_resources": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
			"resources": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[resourcesData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrARN: schema.StringAttribute{
							Computed: true,
						},
						"last_reported_at": schema.StringAttribute{
							Computed: true,
						},
						"owning_account_id": schema.StringAttribute{
							Computed: true,
						},
						"region": schema.StringAttribute{
							Computed: true,
						},
						"resource_type": schema.StringAttribute{
							Computed: true,
						},
						"service": schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"resource_property": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[resourcePropertyData](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"data": schema.StringAttribute{
										Computed: true,
									},
									"last_reported_at": schema.StringAttribute{
										Computed: true,
									},
									names.AttrName: schema.StringAttribute{
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

func (d *dataSourceSearch) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ResourceExplorer2Client(ctx)

	var data dataSourceSearchData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(data.QueryString.ValueString())

	input := &resourceexplorer2.SearchInput{
		QueryString: aws.String(data.QueryString.ValueString()),
	}

	paginator := resourceexplorer2.NewSearchPaginator(conn, input)

	var out resourceexplorer2.SearchOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ResourceExplorer2, create.ErrActionReading, DSNameSearch, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		if page != nil && len(page.Resources) > 0 {
			out.Resources = append(out.Resources, page.Resources...)
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type dataSourceSearchData struct {
	ResourceCount fwtypes.ListNestedObjectValueOf[countData]     `tfsdk:"resource_count"`
	Resources     fwtypes.ListNestedObjectValueOf[resourcesData] `tfsdk:"resources"`
	ID            types.String                                   `tfsdk:"id"`
	ViewArn       types.String                                   `tfsdk:"view_arn"`
	QueryString   types.String                                   `tfsdk:"query_string"`
}

type countData struct {
	Completed      types.Bool  `tfsdk:"completed"`
	TotalResources types.Int64 `tfsdk:"total_resources"`
}

type resourcesData struct {
	ARN              types.String                                          `tfsdk:"arn"`
	LastReportedAt   types.String                                          `tfsdk:"last_reported_at"`
	OwningAccountID  types.String                                          `tfsdk:"owning_account_id"`
	Region           types.String                                          `tfsdk:"region"`
	ResourceProperty fwtypes.ListNestedObjectValueOf[resourcePropertyData] `tfsdk:"resource_property"`
	ResourceType     types.String                                          `tfsdk:"resource_type"`
	Service          types.String                                          `tfsdk:"service"`
}

type resourcePropertyData struct {
	Data           types.String `tfsdk:"data"`
	LastReportedAt types.String `tfsdk:"last_reported_at"`
	Name           types.String `tfsdk:"name"`
}
