// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Inference Profiles")
func newInferenceProfilesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &inferenceProfilesDataSource{}, nil
}

type inferenceProfilesDataSource struct {
	framework.DataSourceWithConfigure
}

func (*inferenceProfilesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_inference_profiles"
}

func (d *inferenceProfilesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARNs: schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
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
	inferenceProfiles, err := findInferencesProfiles(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError("listing Bedrock Inference Profiles", err.Error())
		return
	}

	arns := tfslices.ApplyToAll(inferenceProfiles, func(v awstypes.InferenceProfileSummary) string {
		return aws.ToString(v.InferenceProfileArn)
	})

	data.ARNs = fwflex.FlattenFrameworkStringValueList(ctx, arns)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findInferencesProfiles(ctx context.Context, conn *bedrock.Client, input *bedrock.ListInferenceProfilesInput) ([]awstypes.InferenceProfileSummary, error) {
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
	ARNs types.List `tfsdk:"arns"`
}
