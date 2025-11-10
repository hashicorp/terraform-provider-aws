// Copyright (c) HashiCorp, Inc.
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
)

// @FrameworkDataSource("aws_route53profiles_profiles", name="Profiles")
func newProfilesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &profilesDataSource{}, nil
}

const (
	DSNameProfiles = "Profiles Data Source"
)

type profilesDataSource struct {
	framework.DataSourceWithModel[profilesDataSourceModel]
}

func (d *profilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"profiles": framework.DataSourceComputedListOfObjectAttribute[profileSummaryModel](ctx),
		},
	}
}

func (d *profilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().Route53ProfilesClient(ctx)

	var data profilesDataSourceModel

	input := &route53profiles.ListProfilesInput{}
	var output *route53profiles.ListProfilesOutput
	pages := route53profiles.NewListProfilesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError("listing Route53 Profiles", err.Error())

			return
		}

		if output == nil {
			output = page
		} else {
			output.ProfileSummaries = append(output.ProfileSummaries, page.ProfileSummaries...)
		}
	}

	resp.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type profilesDataSourceModel struct {
	framework.WithRegionModel
	ProfileSummaries fwtypes.ListNestedObjectValueOf[profileSummaryModel] `tfsdk:"profiles"`
}

type profileSummaryModel struct {
	ARN         types.String                             `tfsdk:"arn"`
	ID          types.String                             `tfsdk:"id"`
	Name        types.String                             `tfsdk:"name"`
	ShareStatus fwtypes.StringEnum[awstypes.ShareStatus] `tfsdk:"share_status"`
}
