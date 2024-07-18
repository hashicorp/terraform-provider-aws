// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/bedrock/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// awstypes.<Type Name>.
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
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_evaluation_job", name="Evaluation Job")
func newResourceEvaluationJob(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEvaluationJob{}
	r.SetDefaultReadTimeout(30 * time.Minute)
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
			"role_arn": schema.StringAttribute{
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
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.EvaluationJobType](),
				Computed:   true,
			},

			//names.AttrTags:    tftags.TagsAttribute(),
			//names.AttrTagsAll: tftags.TagsAttribute(), // not too sure how to do these
		},
		Blocks: map[string]schema.Block{
			"evaluation_config": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[dEvaluationConfig](ctx),
				Validators: []validator.Object{
					/*
						objectvalidator.ConflictsWith(path.Expressions{
							path.MatchRoot("human_config"),
						}...),
					*/
				},
				Blocks: map[string]schema.Block{
					"automated": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[dAutomatedEvaluationConfig](ctx),
						Blocks: map[string]schema.Block{
							"dataset_metric_configs": schema.ListNestedBlock{
								CustomType: fwtypes.NewListNestedObjectTypeOf[dEvaluationDatasetMetricConfig](ctx),
								Validators: []validator.List{
									listvalidator.SizeBetween(1, 5),
									listvalidator.IsRequired(),
								},
								NestedObject: schema.NestedBlockObject{
									Attributes: map[string]schema.Attribute{
										"metric_names": schema.ListAttribute{
											ElementType: types.StringType,
											Required:    true,
											Validators: []validator.List{
												listvalidator.SizeBetween(1, 10),
												listvalidator.ValueStringsAre(stringvalidator.LengthBetween(1, 63)),
												listvalidator.ValueStringsAre(stringvalidator.RegexMatches(regexache.MustCompile(`^[0-9a-zA-Z-_.]+$`), "metric_names must conform to: ^[0-9a-zA-Z-_.]+$ ")),
											},
										},
										"task_type": schema.StringAttribute{
											Required: true,
											Validators: []validator.String{
												stringvalidator.LengthBetween(0, 2048),
												//stringvalidator.RegexMatches(regexache.MustCompile("^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"), "role_arn must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
												stringvalidator.AtLeastOneOf(path.Expressions{
													path.MatchRoot("Summarization"), path.MatchRoot("Classification"), path.MatchRoot("QuestionAndAnswer"), path.MatchRoot("Generation"), path.MatchRoot("Custom"),
												}...),
											},
										},
									},
									Blocks: map[string]schema.Block{
										"data_set": schema.SingleNestedBlock{
											CustomType: fwtypes.NewObjectTypeOf[dEvaluationDataset](ctx),
											Validators: []validator.Object{
												objectvalidator.IsRequired(),
											},
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(0, 63),
														stringvalidator.RegexMatches(regexache.MustCompile("^[0-9a-zA-Z-_.]+$"), " must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
														stringvalidator.AtLeastOneOf(path.Expressions{
															path.MatchRoot("Builtin.Bold"), path.MatchRoot("Builtin.BoolQ"), path.MatchRoot("Builtin.NaturalQuestions"), path.MatchRoot("Builtin.Gigaword"), path.MatchRoot("Builtin.RealToxicityPrompts"), path.MatchRoot("Builtin.TriviaQa"), path.MatchRoot("Builtin.WomensEcommerceClothingReviews"), path.MatchRoot("Builtin.Wikitext2"),
														}...),
													},
												},
											},
											Blocks: map[string]schema.Block{
												"evaluation_dataset_location": schema.SingleNestedBlock{
													CustomType: fwtypes.NewObjectTypeOf[awstypes.EvaluationDatasetLocation](ctx),
													Attributes: map[string]schema.Attribute{
														"s3_uri": schema.StringAttribute{
															Optional: true,
															Validators: []validator.String{
																stringvalidator.LengthBetween(1, 1024),
																stringvalidator.RegexMatches(regexache.MustCompile("^s3://[a-z0-9][\\.\\-a-z0-9]{1,61}[a-z0-9](/.*)?$"), "role_arn must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
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
			"inference_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dEvaluationInferenceConfig](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"models": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[dEvaluationModelConfig](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"bedrock_model": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[dEvaluationBedrockModel](ctx),
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"inference_params": schema.StringAttribute{
													Required: true,
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
				Read: true,
			}),
			"output_data_config": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[dOutputDataConfig](ctx),
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"s3_uri": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.LengthBetween(1, 1024),
							stringvalidator.RegexMatches(regexache.MustCompile("^s3://[a-z0-9][\\.\\-a-z0-9]{1,61}[a-z0-9](/.*)?$"), "role_arn must conform to ^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$"),
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

	out, err := conn.CreateEvaluationJob(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameEvaluationJob, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.JobArn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResNameEvaluationJob, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.Arn = flex.StringToFramework(ctx, out.JobArn)
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

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
	}

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
}
func (r *resourceEvaluationJob) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *resourceEvaluationJob) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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

	ClientRequestToken      types.String                                      `tfsdk:"client_request_token"`
	CustomerEncryptionKeyId types.String                                      `tfsdk:"customer_encryption_key_id"`
	EvaluationConfig        fwtypes.ObjectValueOf[dEvaluationConfig]          `tfsdk:"evaluation_config"`
	InferenceConfig         fwtypes.ObjectValueOf[dEvaluationInferenceConfig] `tfsdk:"inference_config"`
	Description             types.String                                      `tfsdk:"description"`
	Name                    types.String                                      `tfsdk:"name"`
	//Tags                    types.Map                                         `tfsdk:"tags"`     // check these
	//TagsAll                 types.Map                                         `tfsdk:"tags_all"` // check these
	OutputDataConfig fwtypes.ObjectValueOf[dOutputDataConfig] `tfsdk:"output_data_config"`
	RoleArn          types.String                             `tfsdk:"role_arn"`
	Arn              types.String                             `tfsdk:"arn"`
	Timeouts         timeouts.Value                           `tfsdk:"timeouts"`
}

type dataGetJob struct {
	CreationTime            timetypes.RFC3339                                 `tfsdk:"creation_time"`
	CustomerEncryptionKeyId types.String                                      `tfsdk:"customer_encryption_key_id"`
	EvaluationConfig        fwtypes.ObjectValueOf[dAutomatedEvaluationConfig] `tfsdk:"evaluation_config"`
	InferenceConfig         fwtypes.ObjectValueOf[dEvaluationInferenceConfig] `tfsdk:"inference_config"`
	Arn                     types.String                                      `tfsdk:"arn"`
	Name                    types.String                                      `tfsdk:"name"`
	Type                    fwtypes.StringEnum[awstypes.EvaluationJobType]    `tfsdk:"type"`
	//LastModifiedTime        timetypes.RFC3339                                           `tfsdk:"last_modified_time"`
	OutOutputDataConfig fwtypes.ListNestedObjectValueOf[dEvaluationDatasetLocation] `tfsdk:"output_data_config"`
	RoleArn             types.String                                                `tfsdk:"role_arn"`
	Status              fwtypes.StringEnum[awstypes.EvaluationJobStatus]            `tfsdk:"status"`
}

// start of evaluation_config
type dEvaluationConfig struct {
	Automated fwtypes.ObjectValueOf[dAutomatedEvaluationConfig] `tfsdk:"automated"`
	// Human             fwtypes.ListNestedObjectValueOf[dsHumanEvaluationConfig]   `tfsdk:"human"`
}
type dAutomatedEvaluationConfig struct {
	EvaluationDataSet fwtypes.ListNestedObjectValueOf[dEvaluationDatasetMetricConfig] `tfsdk:"dataset_metric_configs"` // this is a list make sure its a list
	/*
		Array Members: Minimum number of 1 item. Maximum number of 5 items.
	*/
}
type dEvaluationDatasetMetricConfig struct {
	Dataset     fwtypes.ObjectValueOf[dEvaluationDataset] `tfsdk:"dataset"`
	MetricNames types.List                                `tfsdk:"metric_names"` // array members; min 1 max 10 items ; min length of items 1 max 63 ; pattern ^[0-9a-zA-Z-_.]+$ ; required yes
	TaskType    types.String                              `tfsdk:"task_type"`    // min length of 1 max 63 ; pattern ^[A-Za-z0-9]+$ ; Valid values Summarization | Classification | QuestionAndAnswer | Generation | Custom ; required yes
}
type dEvaluationDataset struct {
	name types.String `tfsdk:"name"`
	/*
		Used to specify supported built-in prompt datasets. Valid values are Builtin.Bold, Builtin.BoolQ, Builtin.NaturalQuestions, Builtin.Gigaword, Builtin.RealToxicityPrompts, Builtin.TriviaQa, Builtin.T-Rex, Builtin.WomensEcommerceClothingReviews and Builtin.Wikitext2.

		Type: String

		Length Constraints: Minimum length of 1. Maximum length of 63.

		Pattern: ^[0-9a-zA-Z-_.]+$

		Required: Yes
	*/
	EvaluationDatasetLocation fwtypes.ObjectValueOf[dEvaluationDatasetLocation] `tfsdk:"evaluation_dataset_location"`
	/*
		Type: EvaluationDatasetLocation object

		Note: This object is a Union. Only one member of this object can be specified or returned.

		Required: No
	*/
}
type dEvaluationDatasetLocation struct {
	S3Uri types.String `tfsdk:"s3_uri"`
	/*
		Length Constraints: Minimum length of 1. Maximum length of 1024.

		Pattern: ^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$

		Required: No

	*/
}

// end of evaluation_config

// start of inference_config
type dEvaluationInferenceConfig struct {
	Models fwtypes.ListNestedObjectValueOf[dEvaluationModelConfig] `tfsdk:"models"`
	/*
		Array Members: Minimum number of 1 item. Maximum number of 2 items.

		Required: No
	*/
}

type dEvaluationModelConfig struct {
	EvaluationModelConfig fwtypes.ObjectValueOf[dEvaluationBedrockModel] `tfsdk:"evaluation_model_config"`
}

type dEvaluationBedrockModel struct {
	InferenceParams types.String `tfsdk:"inference_params"`
	/*
		Each Amazon Bedrock support different inference parameters that change how the model behaves during inference.

		Type: String

		Length Constraints: Minimum length of 1. Maximum length of 1023.

		Required: Yes
	*/
	ModelIdentifiers types.String `tfsdk:"model_identifier"`
	/*

		The ARN of the Amazon Bedrock model specified.

		Type: String

		Length Constraints: Minimum length of 1. Maximum length of 2048.

		Pattern: ^arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}(([:][a-z0-9-]{1,63}){0,2})?/[a-z0-9]{12})|(:foundation-model/([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})))|(([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2}))|(([0-9a-zA-Z][_-]?)+)$

		Required: Yes

	*/
}

// end of evaluation_config

type dOutputDataConfig struct {
	S3Uri types.String `tfsdk:"s3_uri"`
}
