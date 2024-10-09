// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Inference Profiles")
func newInferenceProfilesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &inferenceProfilesDataSource{}, nil
}

type inferenceProfilesDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *inferenceProfilesDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_inference_profiles"
}

func (d *inferenceProfilesDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"inference_profile_summaries": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceProfileSummaryModel](ctx),
				Computed:   true,
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
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.ListInferenceProfiles(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("listing Bedrock Inference Profiles", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(d.Meta().Region)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type inferenceProfilesDataSourceModel struct {
	ID                        types.String                                                  `tfsdk:"id"`
	InferenceProfileSummaries fwtypes.ListNestedObjectValueOf[inferenceProfileSummaryModel] `tfsdk:"inference_profile_summaries"`
}

type inferenceProfileSummaryModel struct {
	CreatedAt            timetypes.RFC3339                                            `tfsdk:"created_at"`
	Description          types.String                                                 `tfsdk:"description"`
	InferenceProfileArn  types.String                                                 `tfsdk:"inference_profile_arn"`
	InferenceProfileId   types.String                                                 `tfsdk:"inference_profile_id"`
	InferenceProfileName types.String                                                 `tfsdk:"inference_profile_name"`
	Models               fwtypes.ListNestedObjectValueOf[inferenceProfilesModelModel] `tfsdk:"models"`
	Status               types.String                                                 `tfsdk:"status"`
	Type                 types.String                                                 `tfsdk:"type"`
	UpdatedAt            timetypes.RFC3339                                            `tfsdk:"updated_at"`
}

type inferenceProfilesModelModel struct {
	ModelArn types.String `tfsdk:"model_arn"`
}
