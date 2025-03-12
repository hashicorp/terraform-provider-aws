// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource("aws_bedrock_inference_profiles", name="Inference Profiles")
func newInferenceProfilesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &inferenceProfilesDataSource{}, nil
}

type inferenceProfilesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *inferenceProfilesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"inference_profile_summaries": framework.DataSourceComputedListOfObjectAttribute[inferenceProfileSummaryModel](ctx),
		},
	}
}

func (d *inferenceProfilesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data inferenceProfilesDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	input := &bedrock.ListInferenceProfilesInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	inferenceProfiles, err := findInferenceProfiles(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError("listing Bedrock Inference Profiles", err.Error())
		return
	}

	output := &bedrock.ListInferenceProfilesOutput{
		InferenceProfileSummaries: inferenceProfiles,
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findInferenceProfiles(ctx context.Context, conn *bedrock.Client, input *bedrock.ListInferenceProfilesInput) ([]awstypes.InferenceProfileSummary, error) {
	var output []awstypes.InferenceProfileSummary

	pages := bedrock.NewListInferenceProfilesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.InferenceProfileSummaries...)
	}

	return output, nil
}

type inferenceProfilesDataSourceModel struct {
	InferenceProfileSummaries fwtypes.ListNestedObjectValueOf[inferenceProfileSummaryModel] `tfsdk:"inference_profile_summaries"`
}

type inferenceProfileSummaryModel struct {
	CreatedAt            timetypes.RFC3339                                           `tfsdk:"created_at"`
	Description          types.String                                                `tfsdk:"description"`
	InferenceProfileARN  fwtypes.ARN                                                 `tfsdk:"inference_profile_arn"`
	InferenceProfileID   types.String                                                `tfsdk:"inference_profile_id"`
	InferenceProfileName types.String                                                `tfsdk:"inference_profile_name"`
	Models               fwtypes.ListNestedObjectValueOf[inferenceProfileModelModel] `tfsdk:"models"`
	Status               fwtypes.StringEnum[awstypes.InferenceProfileStatus]         `tfsdk:"status"`
	Type                 fwtypes.StringEnum[awstypes.InferenceProfileType]           `tfsdk:"type"`
	UpdatedAt            timetypes.RFC3339                                           `tfsdk:"updated_at"`
}

type inferenceProfileModelModel struct {
	ModelARN fwtypes.ARN `tfsdk:"model_arn"`
}
