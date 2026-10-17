// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package interconnect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/interconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/interconnect/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_interconnect_attach_points", name="Attach Points")
func newAttachPointsDataSource(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &attachPointsDataSource{}, nil
}

type attachPointsDataSource struct {
	framework.DataSourceWithModel[attachPointsDataSourceModel]
}

func (d *attachPointsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"environment_id": schema.StringAttribute{
				Required: true,
			},
			"attach_points": framework.DataSourceComputedListOfObjectAttribute[attachPointDescriptorModel](ctx),
		},
	}
}

func (d *attachPointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().InterconnectClient(ctx)

	var data attachPointsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	input := interconnect.ListAttachPointsInput{
		EnvironmentId: data.EnvironmentID.ValueStringPointer(),
	}

	var attachPoints []awstypes.AttachPointDescriptor
	paginator := interconnect.NewListAttachPointsPaginator(conn, &input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		attachPoints = append(attachPoints, page.AttachPoints...)
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, attachPoints, &data.AttachPoints))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type attachPointsDataSourceModel struct {
	framework.WithRegionModel
	AttachPoints  fwtypes.ListNestedObjectValueOf[attachPointDescriptorModel] `tfsdk:"attach_points"`
	EnvironmentID types.String                                                `tfsdk:"environment_id"`
}

type attachPointDescriptorModel struct {
	Identifier types.String                                 `tfsdk:"identifier"`
	Name       types.String                                 `tfsdk:"name"`
	Type       fwtypes.StringEnum[awstypes.AttachPointType] `tfsdk:"type"`
}
