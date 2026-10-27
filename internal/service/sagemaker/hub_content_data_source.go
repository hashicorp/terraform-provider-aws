// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_sagemaker_hub_content", name="Hub Content")
func newHubContentDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &hubContentDataSource{}, nil
}

type hubContentDataSource struct {
	framework.DataSourceWithModel[hubContentDataSourceModel]
}

func (d *hubContentDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"document_schema_version": schema.StringAttribute{
				Computed: true,
			},
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			"hub_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"hub_content_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"hub_content_dependency": framework.DataSourceComputedListOfObjectAttribute[hubContentDependencyModel](ctx),
			"hub_content_description": schema.StringAttribute{
				Computed: true,
			},
			"hub_content_display_name": schema.StringAttribute{
				Computed: true,
			},
			"hub_content_document": schema.StringAttribute{
				Computed: true,
			},
			"hub_content_markdown": schema.StringAttribute{
				Computed: true,
			},
			"hub_content_name": schema.StringAttribute{
				Required: true,
			},
			"hub_content_search_keywords": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"hub_content_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.HubContentStatus](),
				Computed:   true,
			},
			"hub_content_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.HubContentType](),
				Required:   true,
			},
			"hub_content_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"hub_name": schema.StringAttribute{
				Required: true,
			},
			"last_modified_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"reference_min_version": schema.StringAttribute{
				Computed: true,
			},
			"sagemaker_public_hub_content_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"support_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.HubContentSupportStatus](),
				Computed:   true,
			},
			"versioned_hub_content_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
		},
	}
}

func (d *hubContentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SageMakerClient(ctx)

	var data hubContentDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	hubContentName := data.HubContentName.ValueString()

	var input sagemaker.DescribeHubContentInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, data, &input), smerr.ID, hubContentName)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findHubContent(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, hubContentName)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data), smerr.ID, hubContentName)
	if resp.Diagnostics.HasError() {
		return
	}

	// Match aws_sagemaker_hub_content_reference: hub_content_arn and sagemaker_public_hub_content_arn omit the version suffix.
	data.HubContentARN = fwflex.StringToFrameworkARN(ctx, stripARNVersion(out.HubContentArn))
	if out.SageMakerPublicHubContentArn != nil {
		data.SageMakerPublicHubContentARN = fwflex.StringToFrameworkARN(ctx, stripARNVersion(out.SageMakerPublicHubContentArn))
	}
	data.VersionedHubContentARN = fwflex.StringToFrameworkARN(ctx, out.HubContentArn)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, hubContentName)
}

type hubContentDataSourceModel struct {
	framework.WithRegionModel
	CreationTime                 timetypes.RFC3339                                          `tfsdk:"creation_time"`
	DocumentSchemaVersion        types.String                                               `tfsdk:"document_schema_version"`
	FailureReason                types.String                                               `tfsdk:"failure_reason"`
	HubARN                       fwtypes.ARN                                                `tfsdk:"hub_arn"`
	HubContentARN                fwtypes.ARN                                                `tfsdk:"hub_content_arn"`
	HubContentDependencies       fwtypes.ListNestedObjectValueOf[hubContentDependencyModel] `tfsdk:"hub_content_dependency"`
	HubContentDescription        types.String                                               `tfsdk:"hub_content_description"`
	HubContentDisplayName        types.String                                               `tfsdk:"hub_content_display_name"`
	HubContentDocument           types.String                                               `tfsdk:"hub_content_document"`
	HubContentMarkdown           types.String                                               `tfsdk:"hub_content_markdown"`
	HubContentName               types.String                                               `tfsdk:"hub_content_name"`
	HubContentSearchKeywords     fwtypes.ListOfString                                       `tfsdk:"hub_content_search_keywords"`
	HubContentStatus             fwtypes.StringEnum[awstypes.HubContentStatus]              `tfsdk:"hub_content_status"`
	HubContentType               fwtypes.StringEnum[awstypes.HubContentType]                `tfsdk:"hub_content_type"`
	HubContentVersion            types.String                                               `tfsdk:"hub_content_version"`
	HubName                      types.String                                               `tfsdk:"hub_name"`
	LastModifiedTime             timetypes.RFC3339                                          `tfsdk:"last_modified_time"`
	ReferenceMinVersion          types.String                                               `tfsdk:"reference_min_version"`
	SageMakerPublicHubContentARN fwtypes.ARN                                                `tfsdk:"sagemaker_public_hub_content_arn"`
	SupportStatus                fwtypes.StringEnum[awstypes.HubContentSupportStatus]       `tfsdk:"support_status"`
	VersionedHubContentARN       fwtypes.ARN                                                `tfsdk:"versioned_hub_content_arn"`
}

type hubContentDependencyModel struct {
	DependencyCopyPath   types.String `tfsdk:"dependency_copy_path"`
	DependencyOriginPath types.String `tfsdk:"dependency_origin_path"`
}
