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

// @FrameworkDataSource("aws_route53profiles_resource_associations", name="Resource Associations")
func newResourceAssociationsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &resourceAssociationsDataSource{}, nil
}

type resourceAssociationsDataSource struct {
	framework.DataSourceWithModel[resourceAssociationsDataSourceModel]
}

func (d *resourceAssociationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"profile_id": schema.StringAttribute{
				Required: true,
			},
			"resource_associations": framework.DataSourceComputedListOfObjectAttribute[resourceAssociationSummaryModel](ctx),
			names.AttrResourceType: schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *resourceAssociationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().Route53ProfilesClient(ctx)

	var data resourceAssociationsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input route53profiles.ListProfileResourceAssociationsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourceAssociations(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.ProfileID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data.ResourceAssociations))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findResourceAssociations(ctx context.Context, conn *route53profiles.Client, input *route53profiles.ListProfileResourceAssociationsInput) ([]awstypes.ProfileResourceAssociation, error) {
	output := []awstypes.ProfileResourceAssociation{}

	pages := route53profiles.NewListProfileResourceAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.ProfileResourceAssociations...)
	}

	return output, nil
}

type resourceAssociationsDataSourceModel struct {
	framework.WithRegionModel
	ProfileID            types.String                                                     `tfsdk:"profile_id"`
	ResourceAssociations fwtypes.ListNestedObjectValueOf[resourceAssociationSummaryModel] `tfsdk:"resource_associations"`
	ResourceType         types.String                                                     `tfsdk:"resource_type"`
}

type resourceAssociationSummaryModel struct {
	ID                 types.String                               `tfsdk:"id"`
	Name               types.String                               `tfsdk:"name"`
	OwnerID            types.String                               `tfsdk:"owner_id"`
	ProfileID          types.String                               `tfsdk:"profile_id"`
	ResourceARN        types.String                               `tfsdk:"resource_arn"`
	ResourceProperties types.String                               `tfsdk:"resource_properties"`
	ResourceType       types.String                               `tfsdk:"resource_type"`
	Status             fwtypes.StringEnum[awstypes.ProfileStatus] `tfsdk:"status"`
	StatusMessage      types.String                               `tfsdk:"status_message"`
}
