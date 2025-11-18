// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	responseQualityDifferenceMin float64 = 0
	responseQualityDifferenceMax float64 = 100
)

// @FrameworkResource("aws_bedrock_prompt_router", name="Prompt Router")
// @Tags(identifierAttribute="prompt_router_arn")
// @Testing(tagsTest=false)
func newPromptRouterResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &promptRouterResource{}

	return r, nil
}

type promptRouterResource struct {
	framework.ResourceWithModel[promptRouterResourceModel]
}

func (r *promptRouterResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z:.][ _-]?)+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"prompt_router_arn": framework.ARNAttributeComputedOnly(),
			"prompt_router_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][ _-]?)+$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"fallback_model": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[promptRouterTargetModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"model_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"models": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[promptRouterTargetModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"model_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
			"routing_criteria": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[routingCriteriaModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
					listvalidator.SizeAtLeast(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"response_quality_difference": schema.Float64Attribute{
							Required: true,
							Validators: []validator.Float64{
								float64validator.Between(responseQualityDifferenceMin, responseQualityDifferenceMax),
							},
							PlanModifiers: []planmodifier.Float64{
								float64planmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *promptRouterResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data promptRouterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	var input bedrock.CreatePromptRouterInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientRequestToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreatePromptRouter(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.PromptRouterName.String())
		return
	}

	promptRouterARN := aws.ToString(out.PromptRouterArn)

	router, err := findPromptRouterByARN(ctx, conn, promptRouterARN)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, promptRouterARN)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, router, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *promptRouterResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data promptRouterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	promptRouterARN := fwflex.StringValueFromFramework(ctx, data.PromptRouterARN)
	out, err := findPromptRouterByARN(ctx, conn, promptRouterARN)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, promptRouterARN)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *promptRouterResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data promptRouterResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	promptRouterARN := fwflex.StringValueFromFramework(ctx, data.PromptRouterARN)
	input := bedrock.DeletePromptRouterInput{
		PromptRouterArn: aws.String(promptRouterARN),
	}

	_, err := conn.DeletePromptRouter(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, promptRouterARN)
		return
	}
}

func (r *promptRouterResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("prompt_router_arn"), request, response)
}

func findPromptRouterByARN(ctx context.Context, conn *bedrock.Client, arn string) (*bedrock.GetPromptRouterOutput, error) {
	input := bedrock.GetPromptRouterInput{
		PromptRouterArn: aws.String(arn),
	}

	return findPromptRouter(ctx, conn, &input)
}

func findPromptRouter(ctx context.Context, conn *bedrock.Client, input *bedrock.GetPromptRouterInput) (*bedrock.GetPromptRouterOutput, error) {
	out, err := conn.GetPromptRouter(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&sdkretry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(input))
	}

	return out, nil
}

type promptRouterResourceModel struct {
	framework.WithRegionModel
	Description      types.String                                             `tfsdk:"description"`
	FallbackModel    fwtypes.ListNestedObjectValueOf[promptRouterTargetModel] `tfsdk:"fallback_model"`
	Models           fwtypes.ListNestedObjectValueOf[promptRouterTargetModel] `tfsdk:"models"`
	PromptRouterARN  types.String                                             `tfsdk:"prompt_router_arn"`
	PromptRouterName types.String                                             `tfsdk:"prompt_router_name"`
	RoutingCriteria  fwtypes.ListNestedObjectValueOf[routingCriteriaModel]    `tfsdk:"routing_criteria"`
	Tags             tftags.Map                                               `tfsdk:"tags"`
	TagsAll          tftags.Map                                               `tfsdk:"tags_all"`
}

type promptRouterTargetModel struct {
	ModelARN fwtypes.ARN `tfsdk:"model_arn"`
}

type routingCriteriaModel struct {
	ResponseQualityDifference types.Float64 `tfsdk:"response_quality_difference"`
}
