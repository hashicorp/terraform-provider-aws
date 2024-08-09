// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_evaluation_job", name="Evaluation Job")
// @Tags(identifierAttribute="evaluation_job_arn")
func newResourceEvaluationJob(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEvaluationJob{}
	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultReadTimeout(60 * time.Minute)
	return r, nil
}

const (
	ResNameEvaluationJob = "Evaluation Job"
)

type resourceEvaluationJob struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
	framework.WithNoOpDelete
	framework.WithNoOpUpdate[dataEvaluationJob]
}

func (r *resourceEvaluationJob) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrock_evaluation_job"
}

func (r *resourceEvaluationJob) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),

			"client_request_token": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-zA-Z0-9](-*[a-zA-Z0-9])*$"), "client_request_token must conform to ^[a-zA-Z0-9](-*[a-zA-Z0-9])*$"),
				},
			},
			"customer_encryption_key_id": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
					stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:kms:[a-zA-Z0-9-]*:[0-9]{12}:((key/[a-zA-Z0-9-]{36})|(alias/[a-zA-Z0-9-_/]+))$"), "customer_encryption_key_id must conform to ^arn:aws(-[^:]+)?:kms:[a-zA-Z0-9-]*:[0-9]{12}:((key/[a-zA-Z0-9-]{36})|(alias/[a-zA-Z0-9-_/]+))$"),
				},
			},
			names.AttrDescription: schema.StringAttribute{ // job description
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
					stringvalidator.RegexMatches(regexache.MustCompile("^.+$"), "description must conform to ^.+$"),
				},
			},
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-z0-9](-*[a-z0-9]){0,62}$"), "arn must conform to ^[a-z0-9](-*[a-z0-9]){0,62}$"),
				},
			},
			//framework.ARNAttributeComputedOnly()
			names.AttrRoleARN: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048), // change to 0
					stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"), "role_arn must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 2048),
					stringvalidator.RegexMatches(regexache.MustCompile("^[a-z0-9](-*[a-z0-9]){0,62}$"), "name must conform to ^^[a-z0-9](-*[a-z0-9]){0,62}$"),
				},
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluationJobType](),
				Computed:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluationJobStatus](),
				Computed:   true,
			},
			"failure_messages": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"evaluation_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationConfig](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"automated": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[automated](ctx),
							Validators: []validator.List{
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("automated"),
									path.MatchRelative().AtParent().AtName("human"),
								),
							},

							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"dataset_metric_configs": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[DatasetMetricConfigs](ctx),
										Validators: []validator.Set{
											setvalidator.SizeBetween(1, 5),
											setvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metric_names": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Required:   true,
													Validators: []validator.List{
														listvalidator.SizeAtLeast(1),
														listvalidator.ValueStringsAre(stringvalidator.OneOf("Builtin.Accuracy", "Builtin.Robustness", "Builtin.Toxicity")),
													},
												},
												"task_type": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														//stringvalidator.LengthBetween(0, 2048),
														stringvalidator.OneOf("Summarization", "Classification", "QuestionAndAnswer", "Generation", "Custom"),
														//path.Expressions{
														//path.MatchRoot("Summarization"), path.MatchRoot("Classification"), path.MatchRoot("QuestionAndAnswer"), path.MatchRoot("Generation"), path.MatchRoot("Custom"),
														//}...),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"dataset": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[EvaluationDataset](ctx),
													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrName: schema.StringAttribute{
																Optional:   true,
																Validators: []validator.String{
																	//stringvalidator.LengthBetween(0, 63),
																	//stringvalidator.RegexMatches(regexache.MustCompile("^[0-9a-zA-Z-_.]+$"), " must conform to ^[0-9a-zA-Z-_.]+$"),
																	//stringvalidator.OneOf("Builtin.Bold", "Builtin.BoolQ", "Builtin.NaturalQuestions", "Builtin.Gigaword", "Builtin.RealToxicityPrompts", "Builtin.TriviaQa", "Builtin.WomensEcommerceClothingReviews", "Builtin.Wikitext2"),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"dataset_location": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[DatasetLocation](ctx),
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"s3_uri": schema.StringAttribute{
																			Optional: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 1024),
																				stringvalidator.RegexMatches(regexache.MustCompile("^s3://[a-z0-9][\\.\\-a-z0-9]{1,61}[a-z0-9](/.*)?$"), " must conform to ^^s3://[a-z0-9][\\.\\-a-z0-9]{1,61}[a-z0-9](/.*)?$"),
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
						},
						"human": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[human](ctx),
							Validators: []validator.List{
								listvalidator.ExactlyOneOf(
									path.MatchRelative().AtParent().AtName("automated"),
									path.MatchRelative().AtParent().AtName("human"),
								),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"dataset_metric_configs": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[DatasetMetricConfigs](ctx),
										Validators: []validator.Set{
											setvalidator.SizeBetween(1, 5),
											setvalidator.IsRequired(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metric_names": schema.ListAttribute{
													CustomType: fwtypes.ListOfStringType,
													Required:   true,
													Validators: []validator.List{
														listvalidator.SizeAtLeast(1),
														listvalidator.ValueStringsAre(stringvalidator.OneOf("Builtin.Accuracy", "Builtin.Robustness", "Builtin.Toxicity")),
													},
												},
												"task_type": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.OneOf("Summarization", "Classification", "QuestionAndAnswer", "Generation", "Custom"),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"dataset": schema.ListNestedBlock{
													CustomType: fwtypes.NewListNestedObjectTypeOf[EvaluationDataset](ctx),
													Validators: []validator.List{
														listvalidator.IsRequired(),
													},

													NestedObject: schema.NestedBlockObject{
														Attributes: map[string]schema.Attribute{
															names.AttrName: schema.StringAttribute{
																Required:   true,
																Validators: []validator.String{
																	//stringvalidator.LengthBetween(0, 63),
																	//stringvalidator.RegexMatches(regexache.MustCompile("^[0-9a-zA-Z-_.]+$"), " must conform to ^[0-9a-zA-Z-_.]+$"),
																	//stringvalidator.OneOf("Builtin.Bold", "Builtin.BoolQ", "Builtin.NaturalQuestions", "Builtin.Gigaword", "Builtin.RealToxicityPrompts", "Builtin.TriviaQa", "Builtin.WomensEcommerceClothingReviews", "Builtin.Wikitext2"),
																},
															},
														},
														Blocks: map[string]schema.Block{
															"dataset_location": schema.ListNestedBlock{
																CustomType: fwtypes.NewListNestedObjectTypeOf[DatasetLocation](ctx),
																NestedObject: schema.NestedBlockObject{
																	Attributes: map[string]schema.Attribute{
																		"s3_uri": schema.StringAttribute{
																			Optional: true,
																			Validators: []validator.String{
																				stringvalidator.LengthBetween(1, 1024),
																				stringvalidator.RegexMatches(regexache.MustCompile("^s3://[a-z0-9][\\.\\-a-z0-9]{1,61}[a-z0-9](/.*)?$"), " must conform to ^^s3://[a-z0-9][\\.\\-a-z0-9]{1,61}[a-z0-9](/.*)?$"),
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
									"custom_metrics": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[customMetrics](ctx),
										Validators: []validator.Set{
											setvalidator.SizeBetween(1, 10),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrARN: schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 63),
														stringvalidator.RegexMatches(regexache.MustCompile("^[0-9a-zA-Z-_.]+$"), " must conform to ^[0-9a-zA-Z-_.]+$"),
													},
												},
												"rating_method": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.OneOf("ThumbsUpDown", "IndividualLikertScale", "ComparisonLikertScale", "ComparisonChoice", "ComparisonRank"),
													},
												},
												names.AttrDescription: schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 63),
														stringvalidator.RegexMatches(regexache.MustCompile("^.+$"), " must conform to ^.+$"),
													},
												},
											},
										},
									},
									"human_workflow_config": schema.SetNestedBlock{
										CustomType: fwtypes.NewSetNestedObjectTypeOf[humanWorkflowConfig](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"flow_definition_arn": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 1024),
														stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:sagemaker:[a-z0-9-]{1,20}:[0-9]{12}:flow-definition/.*$"), " must conform to ^arn:aws(-[^:]+)?:sagemaker:[a-z0-9-]{1,20}:[0-9]{12}:flow-definition/.*$"),
													},
												},
												names.AttrDescription: schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 5000),
														stringvalidator.RegexMatches(regexache.MustCompile("^[\\S\\s]+$"), " must conform to [\\S\\s]+$"),
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
			"inference_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[evaluationInferenceConfig](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"models": schema.SetNestedBlock{
							CustomType: fwtypes.NewSetNestedObjectTypeOf[Models](ctx),
							Validators: []validator.Set{
								setvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"bedrock_model": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[BedrockModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inference_params": schema.StringAttribute{
													Optional: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 1023),
													},
												},
												"model_identifier": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 2048),
														stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}(([:][a-z0-9-]{1,63}){0,2})?/[a-z0-9]{12})|(:foundation-model/([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})))|(([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2}))|(([0-9a-zA-Z][_-]?)+)$"), "model_identifier must match ^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}(([:][a-z0-9-]{1,63}){0,2})?/[a-z0-9]{12})|(:foundation-model/([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})))|(([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2}))|(([0-9a-zA-Z][_-]?)+)$"),
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
				Read:   true,
			}),
			"output_data_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[OutputDataConfig](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_uri": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1024),
								stringvalidator.RegexMatches(regexache.MustCompile("^s3://[a-z0-9][-.a-z0-9]{1,61}(?:/[-!_*'().a-z0-9A-Z]+(?:/[-!_*'().a-z0-9A-Z]+)*)?/?$"), "must conform to ^s3://[a-z0-9][-.a-z0-9]{1,61}(?:/[-!_*'().a-z0-9A-Z]+(?:/[-!_*'().a-z0-9A-Z]+)*)?/?$"),
							},
						},
					},
				},
			},
		},
	}
}
func (r *resourceEvaluationJob) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)
	var plan dataEvaluationJob
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &bedrock.CreateEvaluationJobInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in.JobName = plan.Name.ValueStringPointer()
	in.JobTags = getTagsIn(ctx)
	outputRaw, err := conn.CreateEvaluationJob(ctx, in)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameEvaluationJob, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if outputRaw == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameEvaluationJob, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	plan.Arn = flex.StringToFramework(ctx, (outputRaw.JobArn))

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	outputRaws, err := tfresource.RetryWhenNotFound(ctx, createTimeout, func() (interface{}, error) {
		return findEvaluationJobByID(ctx, conn, plan.CreationTime.ValueString())
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionCreating, ResNameEvaluationJob, plan.Name.String(), err),
			err.Error(),
		)
	}

	output := outputRaws.(*bedrock.GetEvaluationJobOutput)
	resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}
func (r *resourceEvaluationJob) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dataEvaluationJob
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)
	out, err := waitEvaluationJobRead(ctx, conn, state.Arn.ValueString(), r.ReadTimeout(ctx, state.Timeouts))

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionSetting, ResNameEvaluationJob, state.Arn.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *resourceEvaluationJob) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state dataEvaluationJob

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &bedrock.StopEvaluationJobInput{
		JobIdentifier: state.Arn.ValueStringPointer(),
	}

	_, err := conn.StopEvaluationJob(ctx, in)

	if err != nil && !errs.IsA[*awstypes.ResourceNotFoundException](err) {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.DataZone, create.ErrActionWaitingForDeletion, ResNameEvaluationJob, state.Arn.String(), err),
			err.Error(),
		)
		return
	}
}
func (r *resourceEvaluationJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *resourceEvaluationJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceEvaluationJob) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func waitEvaluationJobRead(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*bedrock.GetEvaluationJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice[awstypes.EvaluationJobStatus](awstypes.EvaluationJobStatusInProgress),
		Target:                    enum.Slice[awstypes.EvaluationJobStatus](awstypes.EvaluationJobStatusCompleted),
		Refresh:                   statusEvaluationJob(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrock.GetEvaluationJobOutput); ok {
		return out, err
	}

	return nil, err
}

func statusEvaluationJob(ctx context.Context, conn *bedrock.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findEvaluationJobByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.Status)), nil
	}
}

func findEvaluationJobByID(ctx context.Context, conn *bedrock.Client, id string) (*bedrock.GetEvaluationJobOutput, error) {
	in := &bedrock.GetEvaluationJobInput{
		JobIdentifier: aws.String(id),
	}

	out, err := conn.GetEvaluationJob(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.JobArn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type dataEvaluationJob struct {
	CreationTime timetypes.RFC3339                                `tfsdk:"creation_time"`
	Type         fwtypes.StringEnum[awstypes.EvaluationJobType]   `tfsdk:"type"`
	Status       fwtypes.StringEnum[awstypes.EvaluationJobStatus] `tfsdk:"status"`

	ClientRequestToken      types.String                                               `tfsdk:"client_request_token"`
	CustomerEncryptionKeyId types.String                                               `tfsdk:"customer_encryption_key_id"`
	EvaluationConfig        fwtypes.ListNestedObjectValueOf[evaluationConfig]          `tfsdk:"evaluation_config"`
	InferenceConfig         fwtypes.ListNestedObjectValueOf[evaluationInferenceConfig] `tfsdk:"inference_config"`
	Description             types.String                                               `tfsdk:"description"`
	Name                    types.String                                               `tfsdk:"name"`
	Tags                    types.Map                                                  `tfsdk:"tags"`
	TagsAll                 types.Map                                                  `tfsdk:"tags_all"`
	OutputDataConfig        fwtypes.ListNestedObjectValueOf[OutputDataConfig]          `tfsdk:"output_data_config"`
	RoleArn                 types.String                                               `tfsdk:"role_arn"`
	Arn                     types.String                                               `tfsdk:"arn"`
	Timeouts                timeouts.Value                                             `tfsdk:"timeouts"`
	FailureMessages         fwtypes.ListValueOf[types.String]                          `tfsdk:"failure_messages"`
}

type evaluationConfig struct {
	Automated fwtypes.ListNestedObjectValueOf[automated] `tfsdk:"automated"`
	Human     fwtypes.ListNestedObjectValueOf[human]     `tfsdk:"human"`
}

type automated struct {
	DatasetMetricConfigs fwtypes.SetNestedObjectValueOf[DatasetMetricConfigs] `tfsdk:"dataset_metric_configs"`
}
type DatasetMetricConfigs struct {
	Dataset     fwtypes.ListNestedObjectValueOf[EvaluationDataset] `tfsdk:"dataset"`
	MetricNames fwtypes.ListValueOf[types.String]                  `tfsdk:"metric_names"`
	TaskType    types.String                                       `tfsdk:"task_type"`
}
type EvaluationDataset struct {
	Name            types.String                                     `tfsdk:"name"`
	DatasetLocation fwtypes.ListNestedObjectValueOf[DatasetLocation] `tfsdk:"dataset_location"`
}
type DatasetLocation struct {
	S3Uri types.String `tfsdk:"s3_uri"`
}

type evaluationInferenceConfig struct {
	Models fwtypes.SetNestedObjectValueOf[Models] `tfsdk:"models"`
}

type Models struct {
	BedrockModel fwtypes.ListNestedObjectValueOf[BedrockModel] `tfsdk:"bedrock_model"`
}

type BedrockModel struct {
	InferenceParams types.String `tfsdk:"inference_params"`

	ModelIdentifiers types.String `tfsdk:"model_identifier"`
}

type OutputDataConfig struct {
	S3Uri types.String `tfsdk:"s3_uri"`
}

type human struct {
	DatasetMetricConfigs fwtypes.SetNestedObjectValueOf[DatasetMetricConfigs] `tfsdk:"dataset_metric_configs"`
	CustomMetrics        fwtypes.SetNestedObjectValueOf[customMetrics]        `tfsdk:"custom_metrics"`
	HumanWorkflowConfig  fwtypes.ListNestedObjectValueOf[humanWorkflowConfig] `tfsdk:"human_workflow_config"`
}

type customMetrics struct {
	Name         types.String `tfsdk:"name"`
	RatingMethod types.String `tfsdk:"rating_method"`
	Description  types.String `tfsdk:"description"`
}

type humanWorkflowConfig struct {
	FlowDefinitionArn types.String `tfsdk:"flow_definition_arn"`
	Instructions      types.String `tfsdk:"instructions"`
}

func (e evaluationConfig) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !e.Automated.IsNull():
		automatedData, d := e.Automated.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.EvaluationConfigMemberAutomated
		diags.Append(flex.Expand(ctx, automatedData, &r.Value)...)
		return &r, diags

	case !e.Human.IsNull():
		humanData, d := e.Human.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var r awstypes.EvaluationConfigMemberHuman

		diags.Append(flex.Expand(ctx, humanData, &r.Value)...)
		return &r, diags
	}
	return nil, diags
}

func (i evaluationInferenceConfig) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !i.Models.IsNull():
		_, d := i.Models.ToSlice(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationInferenceConfigMemberModels

		diags.Append(flex.Expand(ctx, i.Models, &r.Value)...)
		return &r, diags
	}
	return nil, diags

}

func (m Models) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !m.BedrockModel.IsNull():
		bedrockData, d := m.BedrockModel.ToPtr(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationModelConfigMemberBedrockModel

		diags.Append(flex.Expand(ctx, bedrockData, &r.Value)...)
		return &r, diags

	}
	return nil, diags
}

func (l DatasetLocation) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	switch {
	case !l.S3Uri.IsNull():
		locationData, d := l.S3Uri.ToStringValue(ctx)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		var r awstypes.EvaluationDatasetLocationMemberS3Uri
		r.Value = locationData.ValueString()
		return &r, diags
	}
	return nil, diags
}

/*
func flattenAutomated(ctx context.Context, cfg awstypes.EvaluationConfig) (*evaluationConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var evalCfg evaluationConfig
	if cfg != nil {
		switch v := cfg.(type) {
		case *awstypes.EvaluationConfigMemberAutomated:
			var autoData automated
			d := flex.Flatten(ctx, v.Value, &autoData)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			evalCfg = evaluationConfig{
				Automated: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &autoData),
				Human:     fwtypes.NewListNestedObjectValueOfNull[human](ctx),
			}

		case *awstypes.EvaluationConfigMemberHuman:
			var humanData human
			d := flex.Flatten(ctx, v.Value, &humanData)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
			evalCfg = evaluationConfig{
				Automated: fwtypes.NewListNestedObjectValueOfNull[automated](ctx),
				Human:     fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &humanData),
			}

		}
		return &evalCfg, diags
	}
	return nil, diags

}
func flattenInference(ctx context.Context, cfg awstypes.EvaluationInferenceConfig) (*evaluationInferenceConfig, diag.Diagnostics) {
	var diags diag.Diagnostics
	var infCfg evaluationInferenceConfig
	if cfg != nil {
		switch v := cfg.(type) {
		case *awstypes.EvaluationInferenceConfigMemberModels:
			var infData Models
			d := flex.Flatten(ctx, v.Value, &infData)

			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}

			infCfg = evaluationInferenceConfig{
				Models: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &infData),
			}
		}
	}
	return &infCfg, diags
}
*/
