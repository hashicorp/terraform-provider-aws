// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourceexplorer2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
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
			names.AttrID: framework.IDAttribute(),
			"query_string": schema.StringAttribute{
				Required: true,
			},
			"resource_count": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[countData](ctx),
				Computed:   true,
			},
			names.AttrResources: schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[resourcesData](ctx),
				ElementType: fwtypes.NewObjectTypeOf[resourcesData](ctx),
				Computed:    true,
			},
			"view_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 1011),
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

	if data.ViewArn.IsNull() {
		data.ID = types.StringValue(fmt.Sprintf(",%s", data.QueryString.ValueString()))
	} else {
		data.ID = types.StringValue(fmt.Sprintf("%s,%s", data.ViewArn.ValueString(), data.QueryString.ValueString()))
	}

	input := &resourceexplorer2.SearchInput{
		QueryString: aws.String(data.QueryString.ValueString()),
	}
	if !data.ViewArn.IsNull() {
		input.ViewArn = aws.String(data.ViewArn.ValueString())
	}

	paginator := resourceexplorer2.NewSearchPaginator(conn, input)

	var out resourceexplorer2.SearchOutput
	commonFieldsSet := false
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
			if !commonFieldsSet {
				out.Count = page.Count
				out.ViewArn = page.ViewArn
				commonFieldsSet = true
			}
			out.Resources = append(out.Resources, page.Resources...)
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceSearchData struct {
	Count       fwtypes.ListNestedObjectValueOf[countData]     `tfsdk:"resource_count"`
	ID          types.String                                   `tfsdk:"id"`
	QueryString types.String                                   `tfsdk:"query_string"`
	Resources   fwtypes.ListNestedObjectValueOf[resourcesData] `tfsdk:"resources"`
	ViewArn     fwtypes.ARN                                    `tfsdk:"view_arn"`
}

type countData struct {
	Complete       types.Bool  `tfsdk:"complete"`
	TotalResources types.Int64 `tfsdk:"total_resources"`
}

type resourcesData struct {
	ARN             fwtypes.ARN                                     `tfsdk:"arn"`
	LastReportedAt  timetypes.RFC3339                               `tfsdk:"last_reported_at"`
	OwningAccountID types.String                                    `tfsdk:"owning_account_id"`
	Properties      fwtypes.ListNestedObjectValueOf[propertiesData] `tfsdk:"properties"`
	Region          types.String                                    `tfsdk:"region"`
	ResourceType    types.String                                    `tfsdk:"resource_type"`
	Service         types.String                                    `tfsdk:"service"`
}

type propertiesData struct {
	Data           jsontypes.Normalized `tfsdk:"data"`
	LastReportedAt timetypes.RFC3339    `tfsdk:"last_reported_at"`
	Name           types.String         `tfsdk:"name"`
}
