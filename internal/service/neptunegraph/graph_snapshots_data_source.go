// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptunegraph/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_neptunegraph_graph_snapshots", name="Graph Snapshots")
func newGraphSnapshotsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &graphSnapshotsDataSource{}, nil
}

type graphSnapshotsDataSource struct {
	framework.DataSourceWithModel[graphSnapshotsDataSourceModel]
}

func (d *graphSnapshotsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"graph_identifier": schema.StringAttribute{
				Description: "The unique identifier of the Neptune Analytics graph to scope the listing to. If omitted, snapshots for all graphs are returned.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^g-[a-z0-9]{10}$`), ""),
				},
			},
			"graph_snapshots": framework.DataSourceComputedListOfObjectAttribute[graphSnapshotModel](ctx),
		},
	}
}

func (d *graphSnapshotsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data graphSnapshotsDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().NeptuneGraphClient(ctx)

	var input neptunegraph.ListGraphSnapshotsInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findGraphSnapshots(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output, &data.GraphSnapshots))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func findGraphSnapshots(ctx context.Context, conn *neptunegraph.Client, input *neptunegraph.ListGraphSnapshotsInput) ([]awstypes.GraphSnapshotSummary, error) {
	var output []awstypes.GraphSnapshotSummary

	pages := neptunegraph.NewListGraphSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		output = append(output, page.GraphSnapshots...)
	}

	return output, nil
}

type graphSnapshotsDataSourceModel struct {
	framework.WithRegionModel
	GraphIdentifier types.String                                        `tfsdk:"graph_identifier"`
	GraphSnapshots  fwtypes.ListNestedObjectValueOf[graphSnapshotModel] `tfsdk:"graph_snapshots"`
}

type graphSnapshotModel struct {
	ARN                types.String      `tfsdk:"arn"`
	ID                 types.String      `tfsdk:"id"`
	KMSKeyIdentifier   types.String      `tfsdk:"kms_key_identifier"`
	Name               types.String      `tfsdk:"name"`
	SnapshotCreateTime timetypes.RFC3339 `tfsdk:"snapshot_create_time"`
	SourceGraphID      types.String      `tfsdk:"source_graph_id"`
	Status             types.String      `tfsdk:"status"`
}
