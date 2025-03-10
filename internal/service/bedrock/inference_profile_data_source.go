// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrock_inference_profile", name="Inference Profile")
func newInferenceProfileDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &inferenceProfileDataSource{}, nil
}

type inferenceProfileDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *inferenceProfileDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"inference_profile_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"inference_profile_id": schema.StringAttribute{
				Required: true,
			},
			"inference_profile_name": schema.StringAttribute{
				Computed: true,
			},
			"models": framework.DataSourceComputedListOfObjectAttribute[inferenceProfileModelModel](ctx),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InferenceProfileStatus](),
				Computed:   true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.InferenceProfileType](),
				Computed:   true,
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

	output, err := findInferenceProfileByID(ctx, conn, data.InferenceProfileID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Inference Profile (%s)", data.InferenceProfileID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findInferenceProfileByID(ctx context.Context, conn *bedrock.Client, id string) (*bedrock.GetInferenceProfileOutput, error) {
	input := &bedrock.GetInferenceProfileInput{
		InferenceProfileIdentifier: aws.String(id),
	}

	return findInferenceProfile(ctx, conn, input)
}

func findInferenceProfile(ctx context.Context, conn *bedrock.Client, input *bedrock.GetInferenceProfileInput) (*bedrock.GetInferenceProfileOutput, error) {
	output, err := conn.GetInferenceProfile(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type inferenceProfileDataSourceModel struct {
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
