// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package route53profiles

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53profiles"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53profiles/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_route53profiles_profile", name="Profile")
// @Tags(identifierAttribute="arn")
func newProfileDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &profileDataSource{}, nil
}

const (
	DSNameProfile = "Profile Data Source"
)

type profileDataSource struct {
	framework.DataSourceWithModel[profileDataSourceModel]
}

func (d *profileDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
			},
			"share_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ShareStatus](),
				Computed:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ProfileStatus](),
				Computed:   true,
			},
			names.AttrStatusMessage: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *profileDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot(names.AttrName),
			path.MatchRoot(names.AttrID),
		),
	}
}

func (d *profileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().Route53ProfilesClient(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)

	var data profileDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueString()
	if id == "" {
		name := data.Name.ValueString()
		summary, err := findProfileByName(ctx, conn, name)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
			return
		}

		id = aws.ToString(summary.Id)
	}

	out, err := findProfileByID(ctx, conn, id)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	tags, err := listTags(ctx, conn, aws.ToString(out.Arn))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	data.Tags = tftags.FlattenStringValueMap(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findProfileByName(ctx context.Context, conn *route53profiles.Client, name string) (*awstypes.ProfileSummary, error) {
	var input route53profiles.ListProfilesInput

	var results []awstypes.ProfileSummary
	pages := route53profiles.NewListProfilesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, summary := range page.ProfileSummaries {
			if aws.ToString(summary.Name) == name {
				results = append(results, summary)
			}
		}
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(results))
}

type profileDataSourceModel struct {
	framework.WithRegionModel
	ARN           types.String                               `tfsdk:"arn"`
	ID            types.String                               `tfsdk:"id"`
	Name          types.String                               `tfsdk:"name"`
	OwnerId       types.String                               `tfsdk:"owner_id"`
	ShareStatus   fwtypes.StringEnum[awstypes.ShareStatus]   `tfsdk:"share_status"`
	Status        fwtypes.StringEnum[awstypes.ProfileStatus] `tfsdk:"status"`
	StatusMessage types.String                               `tfsdk:"status_message"`
	Tags          tftags.Map                                 `tfsdk:"tags"`
}
