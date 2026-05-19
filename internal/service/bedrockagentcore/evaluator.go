// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/float64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_evaluator", name="Evaluator")
// @Tags(identifierAttribute="evaluator_arn")
// @Testing(tagsTest=false)
func newEvaluatorResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &evaluatorResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type evaluatorResource struct {
	framework.ResourceWithModel[evaluatorResourceModel]
	framework.WithTimeouts
}

func (r *evaluatorResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"evaluator_arn": framework.ARNAttributeComputedOnly(),
			"evaluator_id":  framework.IDAttribute(),
			"evaluator_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			"level": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluatorLevel](),
				Required:   true,
			},
			names.AttrKMSKeyARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"locked_for_modification": schema.BoolAttribute{
				Computed: true,
			},
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluatorStatus](),
				Computed:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"evaluator_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[evaluatorConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"llm_as_a_judge": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[llmAsAJudgeEvaluatorConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("llm_as_a_judge"),
									path.MatchRelative().AtParent().AtName("code_based"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"instructions": schema.StringAttribute{
										Required:  true,
										Sensitive: true,
									},
								},
								Blocks: map[string]schema.Block{
									"rating_scale": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[ratingScaleModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"numerical": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[numericalScaleDefinitionModel](ctx),
													Validators: []validator.List{
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("numerical"),
															path.MatchRelative().AtParent().AtName("categorical"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"definition": schema.StringAttribute{
																Required: true,
															},
															names.AttrValue: schema.Float64Attribute{
																Required: true,
																Validators: []validator.Float64{
																	float64validator.AtLeast(0),
																},
															},
															"label": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 100),
																},
															},
														},
													},
												},
												"categorical": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[categoricalScaleDefinitionModel](ctx),
													Validators: []validator.List{
														listvalidator.ExactlyOneOf(
															path.MatchRelative().AtParent().AtName("numerical"),
															path.MatchRelative().AtParent().AtName("categorical"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"definition": schema.StringAttribute{
																Required: true,
															},
															"label": schema.StringAttribute{
																Required: true,
																Validators: []validator.String{
																	stringvalidator.LengthBetween(1, 100),
																},
															},
														},
													},
												},
											},
										},
									},
									"model_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[evaluatorModelConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Blocks: map[string]schema.Block{
												"bedrock_evaluator_model_config": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[bedrockEvaluatorModelConfigModel](ctx),
													Validators: []validator.List{
														listvalidator.SizeAtMost(1),
														listvalidator.ExactlyOneOf(
															// If another member is added to the union, this will need to be updated.
															path.MatchRelative().AtParent().AtName("bedrock_evaluator_model_config"),
														),
													},
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															"model_id": schema.StringAttribute{
																Required: true,
															},
															"additional_model_request_fields": schema.StringAttribute{
																CustomType: fwtypes.NewSmithyJSONType(ctx, document.NewLazyDocument),
																Optional:   true,
															},
														},
														Blocks: map[string]schema.Block{
															"inference_config": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[inferenceConfigurationModel](ctx),
																Validators: []validator.List{
																	listvalidator.SizeAtMost(1),
																},
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"max_tokens": schema.Int32Attribute{
																			Optional: true,
																			Validators: []validator.Int32{
																				int32validator.AtLeast(1),
																			},
																		},
																		"temperature": schema.Float64Attribute{
																			Optional: true,
																			Validators: []validator.Float64{
																				float64validator.Between(0, 1),
																			},
																		},
																		"top_p": schema.Float64Attribute{
																			Optional: true,
																			Validators: []validator.Float64{
																				float64validator.Between(0, 1),
																			},
																		},
																		"stop_sequences": schema.ListAttribute{
																			CustomType:  fwtypes.ListOfStringType,
																			Optional:    true,
																			ElementType: types.StringType,
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"code_based": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[codeBasedEvaluatorConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("llm_as_a_judge"),
									path.MatchRelative().AtParent().AtName("code_based"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"lambda_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[lambdaEvaluatorConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
											listvalidator.ExactlyOneOf(
												// If another member is added to the union, this will need to be updated.
												path.MatchRelative().AtParent().AtName("lambda_config"),
											),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"lambda_arn": schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
												},
												"lambda_timeout_in_seconds": schema.Int32Attribute{
													Optional: true,
													Computed: true,
													Validators: []validator.Int32{
														int32validator.Between(1, 300),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *evaluatorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data evaluatorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.CreateEvaluatorInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	var (
		out *bedrockagentcorecontrol.CreateEvaluatorOutput
		err error
	)
	err = tfresource.Retry(ctx, propagationTimeout, func(ctx context.Context) *tfresource.RetryError {
		out, err = conn.CreateEvaluator(ctx, &input)

		// IAM / Lambda propagation.
		if tfawserr.ErrMessageContains(err, errCodeValidationException, "Lambda function") {
			return tfresource.RetryableError(err)
		}

		if err != nil {
			return tfresource.NonRetryableError(err)
		}

		return nil
	})
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.EvaluatorName.String())
		return
	}

	evaluatorID := aws.ToString(out.EvaluatorId)

	if _, err := waitEvaluatorCreated(ctx, conn, evaluatorID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, evaluatorID)
		return
	}

	evaluator, err := findEvaluatorByID(ctx, conn, evaluatorID)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, evaluatorID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, evaluator, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *evaluatorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data evaluatorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	evaluatorID := fwflex.StringValueFromFramework(ctx, data.EvaluatorID)
	out, err := findEvaluatorByID(ctx, conn, evaluatorID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, evaluatorID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *evaluatorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old evaluatorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		evaluatorID := fwflex.StringValueFromFramework(ctx, new.EvaluatorID)
		var input bedrockagentcorecontrol.UpdateEvaluatorInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(create.UniqueId(ctx))
		input.EvaluatorId = aws.String(evaluatorID)

		_, err := conn.UpdateEvaluator(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, evaluatorID)
			return
		}

		evaluator, err := waitEvaluatorUpdated(ctx, conn, evaluatorID, r.UpdateTimeout(ctx, new.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, evaluatorID)
			return
		}

		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, evaluator, &new))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *evaluatorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data evaluatorResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	evaluatorID := fwflex.StringValueFromFramework(ctx, data.EvaluatorID)
	input := bedrockagentcorecontrol.DeleteEvaluatorInput{
		EvaluatorId: aws.String(evaluatorID),
	}

	_, err := conn.DeleteEvaluator(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, evaluatorID)
		return
	}

	if _, err := waitEvaluatorDeleted(ctx, conn, evaluatorID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, evaluatorID)
		return
	}
}

func (r *evaluatorResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("evaluator_id"), request, response)
}

func waitEvaluatorCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetEvaluatorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EvaluatorStatusCreating),
		Target:                    enum.Slice(awstypes.EvaluatorStatusActive),
		Refresh:                   statusEvaluator(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetEvaluatorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitEvaluatorUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetEvaluatorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EvaluatorStatusUpdating),
		Target:                    enum.Slice(awstypes.EvaluatorStatusActive),
		Refresh:                   statusEvaluator(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetEvaluatorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitEvaluatorDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetEvaluatorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EvaluatorStatusDeleting, awstypes.EvaluatorStatusActive),
		Target:  []string{},
		Refresh: statusEvaluator(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetEvaluatorOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusEvaluator(conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findEvaluatorByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findEvaluatorByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetEvaluatorOutput, error) {
	input := bedrockagentcorecontrol.GetEvaluatorInput{
		EvaluatorId: aws.String(id),
	}

	return findEvaluator(ctx, conn, &input)
}

func findEvaluator(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetEvaluatorInput) (*bedrockagentcorecontrol.GetEvaluatorOutput, error) {
	out, err := conn.GetEvaluator(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type evaluatorResourceModel struct {
	framework.WithRegionModel
	CreatedAt             timetypes.RFC3339                                     `tfsdk:"created_at"`
	Description           types.String                                          `tfsdk:"description"`
	EvaluatorARN          types.String                                          `tfsdk:"evaluator_arn"`
	EvaluatorConfig       fwtypes.ListNestedObjectValueOf[evaluatorConfigModel] `tfsdk:"evaluator_config"`
	EvaluatorID           types.String                                          `tfsdk:"evaluator_id"`
	EvaluatorName         types.String                                          `tfsdk:"evaluator_name"`
	KMSKeyARN             fwtypes.ARN                                           `tfsdk:"kms_key_arn"`
	Level                 fwtypes.StringEnum[awstypes.EvaluatorLevel]           `tfsdk:"level"`
	LockedForModification types.Bool                                            `tfsdk:"locked_for_modification"`
	Status                fwtypes.StringEnum[awstypes.EvaluatorStatus]          `tfsdk:"status"`
	Tags                  tftags.Map                                            `tfsdk:"tags"`
	TagsAll               tftags.Map                                            `tfsdk:"tags_all"`
	Timeouts              timeouts.Value                                        `tfsdk:"timeouts"`
	UpdatedAt             timetypes.RFC3339                                     `tfsdk:"updated_at"`
}

type evaluatorConfigModel struct {
	LlmAsAJudge fwtypes.ListNestedObjectValueOf[llmAsAJudgeEvaluatorConfigModel] `tfsdk:"llm_as_a_judge"`
	CodeBased   fwtypes.ListNestedObjectValueOf[codeBasedEvaluatorConfigModel]   `tfsdk:"code_based"`
}

var (
	_ fwflex.Expander  = evaluatorConfigModel{}
	_ fwflex.Flattener = &evaluatorConfigModel{}
)

func (m *evaluatorConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.EvaluatorConfigMemberLlmAsAJudge:
		var data llmAsAJudgeEvaluatorConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.LlmAsAJudge = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)
	case awstypes.EvaluatorConfigMemberCodeBased:
		var data codeBasedEvaluatorConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.CodeBased = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("evaluator config flatten: %T", v),
		)
	}
	return diags
}

func (m evaluatorConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.LlmAsAJudge.IsNull():
		data, d := m.LlmAsAJudge.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluatorConfigMemberLlmAsAJudge
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.CodeBased.IsNull():
		data, d := m.CodeBased.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluatorConfigMemberCodeBased
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type llmAsAJudgeEvaluatorConfigModel struct {
	Instructions types.String                                               `tfsdk:"instructions"`
	RatingScale  fwtypes.ListNestedObjectValueOf[ratingScaleModel]          `tfsdk:"rating_scale"`
	ModelConfig  fwtypes.ListNestedObjectValueOf[evaluatorModelConfigModel] `tfsdk:"model_config"`
}

type codeBasedEvaluatorConfigModel struct {
	LambdaConfig fwtypes.ListNestedObjectValueOf[lambdaEvaluatorConfigModel] `tfsdk:"lambda_config"`
}

var (
	_ fwflex.Expander  = codeBasedEvaluatorConfigModel{}
	_ fwflex.Flattener = &codeBasedEvaluatorConfigModel{}
)

func (m *codeBasedEvaluatorConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.CodeBasedEvaluatorConfigMemberLambdaConfig:
		var data lambdaEvaluatorConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.LambdaConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("code based evaluator config flatten: %T", v),
		)
	}
	return diags
}

func (m codeBasedEvaluatorConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.LambdaConfig.IsNull():
		data, d := m.LambdaConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.CodeBasedEvaluatorConfigMemberLambdaConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type lambdaEvaluatorConfigModel struct {
	LambdaARN              fwtypes.ARN `tfsdk:"lambda_arn"`
	LambdaTimeoutInSeconds types.Int32 `tfsdk:"lambda_timeout_in_seconds"`
}

type ratingScaleModel struct {
	Numerical   fwtypes.ListNestedObjectValueOf[numericalScaleDefinitionModel]   `tfsdk:"numerical"`
	Categorical fwtypes.ListNestedObjectValueOf[categoricalScaleDefinitionModel] `tfsdk:"categorical"`
}

var (
	_ fwflex.Expander  = ratingScaleModel{}
	_ fwflex.Flattener = &ratingScaleModel{}
)

func (m *ratingScaleModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.RatingScaleMemberNumerical:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &m.Numerical))
		if diags.HasError() {
			return diags
		}
	case awstypes.RatingScaleMemberCategorical:
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &m.Categorical))
		if diags.HasError() {
			return diags
		}

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("rating scale flatten: %T", v),
		)
	}
	return diags
}

func (m ratingScaleModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.Numerical.IsNull():
		var r awstypes.RatingScaleMemberNumerical
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.Numerical, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	case !m.Categorical.IsNull():
		var r awstypes.RatingScaleMemberCategorical
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, m.Categorical, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type numericalScaleDefinitionModel struct {
	Definition types.String  `tfsdk:"definition"`
	Value      types.Float64 `tfsdk:"value"`
	Label      types.String  `tfsdk:"label"`
}

type categoricalScaleDefinitionModel struct {
	Definition types.String `tfsdk:"definition"`
	Label      types.String `tfsdk:"label"`
}

type evaluatorModelConfigModel struct {
	BedrockEvaluatorModelConfig fwtypes.ListNestedObjectValueOf[bedrockEvaluatorModelConfigModel] `tfsdk:"bedrock_evaluator_model_config"`
}

var (
	_ fwflex.Expander  = evaluatorModelConfigModel{}
	_ fwflex.Flattener = &evaluatorModelConfigModel{}
)

func (m *evaluatorModelConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.EvaluatorModelConfigMemberBedrockEvaluatorModelConfig:
		var data bedrockEvaluatorModelConfigModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, t.Value, &data))
		if diags.HasError() {
			return diags
		}
		m.BedrockEvaluatorModelConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &data)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("evaluator model config flatten: %T", v),
		)
	}
	return diags
}

func (m evaluatorModelConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch {
	case !m.BedrockEvaluatorModelConfig.IsNull():
		data, d := m.BedrockEvaluatorModelConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluatorModelConfigMemberBedrockEvaluatorModelConfig
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, data, &r.Value))
		if diags.HasError() {
			return nil, diags
		}
		return &r, diags
	}
	return nil, diags
}

type bedrockEvaluatorModelConfigModel struct {
	AdditionalModelRequestFields fwtypes.SmithyJSON[document.Interface]                       `tfsdk:"additional_model_request_fields" autoflex:"-"`
	ModelID                      types.String                                                 `tfsdk:"model_id"`
	InferenceConfig              fwtypes.ListNestedObjectValueOf[inferenceConfigurationModel] `tfsdk:"inference_config"`
}

var (
	_ fwflex.Expander  = bedrockEvaluatorModelConfigModel{}
	_ fwflex.Flattener = &bedrockEvaluatorModelConfigModel{}
)

func (m bedrockEvaluatorModelConfigModel) Expand(ctx context.Context) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	var out awstypes.BedrockEvaluatorModelConfig
	out.ModelId = fwflex.StringFromFramework(ctx, m.ModelID)

	if !m.InferenceConfig.IsNull() {
		ic, d := m.InferenceConfig.ToPtr(ctx)
		smerr.AddEnrich(ctx, &diags, d)
		if diags.HasError() {
			return nil, diags
		}
		var icOut awstypes.InferenceConfiguration
		smerr.AddEnrich(ctx, &diags, fwflex.Expand(ctx, ic, &icOut))
		if diags.HasError() {
			return nil, diags
		}
		out.InferenceConfig = &icOut
	}

	if !m.AdditionalModelRequestFields.IsNull() && !m.AdditionalModelRequestFields.IsUnknown() {
		doc, err := tfsmithy.DocumentFromJSONString(m.AdditionalModelRequestFields.ValueString(), document.NewLazyDocument)
		if err != nil {
			diags.AddError("creating Smithy document", err.Error())
			return nil, diags
		}
		out.AdditionalModelRequestFields = doc
	}

	return &out, diags
}

func (m *bedrockEvaluatorModelConfigModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	in, ok := v.(awstypes.BedrockEvaluatorModelConfig)
	if !ok {
		if p, pok := v.(*awstypes.BedrockEvaluatorModelConfig); pok && p != nil {
			in = *p
		} else {
			diags.AddError("Unsupported Type", fmt.Sprintf("bedrock evaluator model config flatten: %T", v))
			return diags
		}
	}

	m.ModelID = fwflex.StringToFramework(ctx, in.ModelId)

	if in.InferenceConfig != nil {
		var ic inferenceConfigurationModel
		smerr.AddEnrich(ctx, &diags, fwflex.Flatten(ctx, in.InferenceConfig, &ic))
		if diags.HasError() {
			return diags
		}
		m.InferenceConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ic)
	} else {
		m.InferenceConfig = fwtypes.NewListNestedObjectValueOfNull[inferenceConfigurationModel](ctx)
	}

	if in.AdditionalModelRequestFields != nil {
		s, err := tfsmithy.DocumentToJSONString(in.AdditionalModelRequestFields)
		if err != nil {
			diags.AddError("reading Smithy document", err.Error())
			return diags
		}
		m.AdditionalModelRequestFields = fwtypes.NewSmithyJSONValue(s, document.NewLazyDocument)
	} else {
		m.AdditionalModelRequestFields = fwtypes.NewSmithyJSONNull[document.Interface]()
	}

	return diags
}

type inferenceConfigurationModel struct {
	MaxTokens     types.Int32          `tfsdk:"max_tokens"`
	Temperature   types.Float64        `tfsdk:"temperature"`
	TopP          types.Float64        `tfsdk:"top_p"`
	StopSequences fwtypes.ListOfString `tfsdk:"stop_sequences"`
}
