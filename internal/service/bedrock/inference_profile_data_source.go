// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource(name="Inference Profile")
func newInferenceProfileDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &inferenceProfileDataSource{}, nil
}

type inferenceProfileDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *inferenceProfileDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_inference_profile"
}

func (d *inferenceProfileDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"inference_profile_arn": schema.StringAttribute{
				Computed: true,
			},
			"inference_profile_id": schema.StringAttribute{
				Required: true,
			},
			"inference_profile_name": schema.StringAttribute{
				Computed: true,
			},
			"models": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceProfileModelModel](ctx),
				Computed:   true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *inferenceProfileDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data inferenceProfileDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)
	input := &bedrock.GetInferenceProfileInput{
		InferenceProfileIdentifier: data.InferenceProfileID.ValueStringPointer(),
	}

	output, err := conn.GetInferenceProfile(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("error reading inference profile: %s", data.InferenceProfileID.ValueString()), err.Error())
		return
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(*output.InferenceProfileId)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type inferenceProfileDataSourceModel struct {
	ID                   types.String                                                `tfsdk:"id"`
	InferenceProfileARN  types.String                                                `tfsdk:"inference_profile_arn"`
	InferenceProfileID   types.String                                                `tfsdk:"inference_profile_id"`
	InferenceProfileName types.String                                                `tfsdk:"inference_profile_name"`
	Models               fwtypes.ListNestedObjectValueOf[inferenceProfileModelModel] `tfsdk:"models"`
	Status               types.String                                                `tfsdk:"status"`
	Type                 types.String                                                `tfsdk:"type"`
	CreatedAt            timetypes.RFC3339                                           `tfsdk:"created_at"`
	Description          types.String                                                `tfsdk:"description"`
	UpdatedAt            timetypes.RFC3339                                           `tfsdk:"updated_at"`
}

type inferenceProfileModelModel struct {
	ModelArn types.String `tfsdk:"model_arn"`
}
