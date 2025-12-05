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

// @FrameworkDataSource("aws_bedrock_prompt_router", name="Prompt Router")
func newPromptRouterDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &promptRouterDataSource{}, nil
}

type promptRouterDataSource struct {
	framework.DataSourceWithModel[promptRouterDataSourceModel]
}

func (d *promptRouterDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"fallback_model": framework.DataSourceComputedListOfObjectAttribute[promptRouterTargetModelModel](ctx),
			"models":         framework.DataSourceComputedListOfObjectAttribute[promptRouterTargetModelModel](ctx),
			"prompt_router_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"prompt_router_name": schema.StringAttribute{
				Computed: true,
			},
			"routing_criteria": framework.DataSourceComputedListOfObjectAttribute[routingCriteriaModel](ctx),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PromptRouterStatus](),
				Computed:   true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PromptRouterType](),
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *promptRouterDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data promptRouterDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	output, err := findPromptRouterByARN(ctx, conn, data.PromptRouterARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Prompt Router (%s)", data.PromptRouterARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findPromptRouterByARN(ctx context.Context, conn *bedrock.Client, arn string) (*bedrock.GetPromptRouterOutput, error) {
	input := &bedrock.GetPromptRouterInput{
		PromptRouterArn: aws.String(arn),
	}

	return findPromptRouter(ctx, conn, input)
}

func findPromptRouter(ctx context.Context, conn *bedrock.Client, input *bedrock.GetPromptRouterInput) (*bedrock.GetPromptRouterOutput, error) {
	output, err := conn.GetPromptRouter(ctx, input)

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

type promptRouterDataSourceModel struct {
	framework.WithRegionModel
	CreatedAt        timetypes.RFC3339                                             `tfsdk:"created_at"`
	Description      types.String                                                  `tfsdk:"description"`
	FallbackModel    fwtypes.ListNestedObjectValueOf[promptRouterTargetModelModel] `tfsdk:"fallback_model"`
	Models           fwtypes.ListNestedObjectValueOf[promptRouterTargetModelModel] `tfsdk:"models"`
	PromptRouterARN  fwtypes.ARN                                                   `tfsdk:"prompt_router_arn"`
	PromptRouterName types.String                                                  `tfsdk:"prompt_router_name"`
	RoutingCriteria  fwtypes.ListNestedObjectValueOf[routingCriteriaModel]         `tfsdk:"routing_criteria"`
	Status           fwtypes.StringEnum[awstypes.PromptRouterStatus]               `tfsdk:"status"`
	Type             fwtypes.StringEnum[awstypes.PromptRouterType]                 `tfsdk:"type"`
	UpdatedAt        timetypes.RFC3339                                             `tfsdk:"updated_at"`
}

type promptRouterTargetModelModel struct {
	ModelARN types.String `tfsdk:"model_arn"`
}

type routingCriteriaModel struct {
	ResponseQualityDifference types.Float64 `tfsdk:"response_quality_difference"`
}
