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

// @FrameworkDataSource(name="Profiles")
func newDataSourceProfiles(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceProfiles{}, nil
}

const (
	DSNameProfiles = "Profiles Data Source"
)

type dataSourceProfiles struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceProfiles) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_route53profiles_profiles"
}

func (d *dataSourceProfiles) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"profiles": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[profileSummariesData](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[profileSummariesData](ctx),
				},
			},
		},
	}
}

func (d *dataSourceProfiles) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().Route53ProfilesClient(ctx)

	var data dataSourceProfilesData

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

type dataSourceProfilesData struct {
	ProfileSummaries fwtypes.ListNestedObjectValueOf[profileSummariesData] `tfsdk:"profiles"`
}

type profileSummariesData struct {
	ARN         types.String                             `tfsdk:"arn"`
	ID          types.String                             `tfsdk:"id"`
	Name        types.String                             `tfsdk:"name"`
	ShareStatus fwtypes.StringEnum[awstypes.ShareStatus] `tfsdk:"share_status"`
}
