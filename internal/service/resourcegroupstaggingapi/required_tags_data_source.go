// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package resourcegroupstaggingapi

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_resourcegroupstaggingapi_required_tags", name="Required Tags")
func newRequiredTagsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &requiredTagsDataSource{}, nil
}

type requiredTagsDataSource struct {
	framework.DataSourceWithModel[requiredTagsDataSourceModel]
}

func (d *requiredTagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"required_tags": framework.DataSourceComputedListOfObjectAttribute[requiredTagModel](ctx),
		},
	}
}
func (d *requiredTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ResourceGroupsTaggingAPIClient(ctx)

	var data requiredTagsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRequiredTags(ctx, conn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.RequiredTags))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findRequiredTags(ctx context.Context, conn *resourcegroupstaggingapi.Client) ([]awstypes.RequiredTag, error) {
	input := resourcegroupstaggingapi.ListRequiredTagsInput{}
	paginator := resourcegroupstaggingapi.NewListRequiredTagsPaginator(conn, &input)

	var output []awstypes.RequiredTag
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.RequiredTags...)
	}

	return output, nil
}

type requiredTagsDataSourceModel struct {
	framework.WithRegionModel
	RequiredTags fwtypes.ListNestedObjectValueOf[requiredTagModel] `tfsdk:"required_tags"`
}

type requiredTagModel struct {
	CloudFormationResourceTypes fwtypes.ListValueOf[types.String] `tfsdk:"cloud_formation_resource_types"`
	ReportingTagKeys            fwtypes.ListValueOf[types.String] `tfsdk:"reporting_tag_keys"`
	ResourceType                types.String                      `tfsdk:"resource_type"`
}
