// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53profiles

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_route53profiles_associations", name="Associations")
func newAssociationsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &associationsDataSource{}, nil
}

type associationsDataSource struct {
	framework.DataSourceWithModel[associationsDataSourceModel]
}

func (d *associationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"associations": framework.DataSourceComputedListOfObjectAttribute[associationSummaryModel](ctx),
			"profile_id": schema.StringAttribute{
				Optional: true,
			},
			names.AttrResourceID: schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *associationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().Route53ProfilesClient(ctx)

	var data associationsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input route53profiles.ListProfileAssociationsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAssociations(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data.Associations))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findAssociations(ctx context.Context, conn *route53profiles.Client, input *route53profiles.ListProfileAssociationsInput) ([]awstypes.ProfileAssociation, error) {
	output := []awstypes.ProfileAssociation{}

	pages := route53profiles.NewListProfileAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.ProfileAssociations...)
	}

	return output, nil
}

type associationsDataSourceModel struct {
	framework.WithRegionModel
	Associations fwtypes.ListNestedObjectValueOf[associationSummaryModel] `tfsdk:"associations"`
	ProfileID    types.String                                             `tfsdk:"profile_id"`
	ResourceID   types.String                                             `tfsdk:"resource_id"`
}

type associationSummaryModel struct {
	ID            types.String                               `tfsdk:"id"`
	Name          types.String                               `tfsdk:"name"`
	OwnerID       types.String                               `tfsdk:"owner_id"`
	Priority      types.Int32                                `tfsdk:"priority"`
	ProfileID     types.String                               `tfsdk:"profile_id"`
	ResourceID    types.String                               `tfsdk:"resource_id"`
	Status        fwtypes.StringEnum[awstypes.ProfileStatus] `tfsdk:"status"`
	StatusMessage types.String                               `tfsdk:"status_message"`
}
