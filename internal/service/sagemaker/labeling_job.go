// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_sagemaker_labeling_job", name="Labeling Job")
// @Tags(identifierAttribute="labeling_job_arn")
// @Testing(tagsTest=false)
func newLabelingJobResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &labelingJobResource{}
	return r, nil
}

type labelingJobResource struct {
	framework.ResourceWithModel[labelingJobResourceModel]
	framework.WithNoUpdate
}

func (r *labelingJobResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"failure_reason": schema.StringAttribute{
				Computed: true,
			},
			"job_reference_code": schema.StringAttribute{
				Computed: true,
			},
			"label_attribute_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,126}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label_category_config_s3_uri": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label_counters":   framework.ResourceComputedListOfObjectsAttribute[labelCountersModel](ctx),
			"labeling_job_arn": framework.ARNAttributeComputedOnly(),
			"labeling_job_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9]){0,62}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"labeling_job_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.LabelingJobStatus](),
				Computed:   true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"stopping_conditions": framework.ResourceOptionalComputedListOfObjectsAttribute[labelingJobStoppingConditionsModel](ctx, 1, nil, listplanmodifier.RequiresReplace()),
			names.AttrTags:        tftags.TagsAttributeForceNew(),
			names.AttrTagsAll:     tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"human_task_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[humanTaskConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_concurrent_task_count": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							Validators: []validator.Int32{
								int32validator.Between(1, 5000),
							},
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
						"number_of_human_workers_per_data_object": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(1, 9),
							},
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
						"pre_human_task_lambda_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"task_availability_lifetime_in_seconds": schema.Int32Attribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
						"task_description": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 255),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"task_keywords": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							ElementType: types.StringType,
							Optional:    true,
							Validators: []validator.Set{
								setvalidator.SizeBetween(1, 5),
								setvalidator.ValueStringsAre(
									stringvalidator.LengthBetween(1, 30),
								),
							},
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						"task_time_limit_in_seconds": schema.Int32Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
						},
						"task_title": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 128),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"workteam_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"annotation_consolidation_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[annotationConsolidationConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"annotation_consolidation_lambda_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"public_workforce_task_price": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[publicWorkforceTaskPriceModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"amount_in_usd": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[usdModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"cents": schema.Int32Attribute{
													Optional: true,
													PlanModifiers: []planmodifier.Int32{
														int32planmodifier.RequiresReplace(),
													},
												},
												"dollars": schema.Int32Attribute{
													Optional: true,
													PlanModifiers: []planmodifier.Int32{
														int32planmodifier.RequiresReplace(),
													},
												},
												"tenth_fractions_of_a_cent": schema.Int32Attribute{
													Optional: true,
													PlanModifiers: []planmodifier.Int32{
														int32planmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
								},
							},
						},
						"ui_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[uiConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"human_task_ui_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Optional:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
									"ui_template_s3_uri": schema.StringAttribute{
										Optional: true,
										Validators: []validator.String{
											stringvalidator.RegexMatches(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
										},
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
					},
				},
			},
			"input_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobInputConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"data_attributes": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobDataAttributesModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"content_classifiers": schema.SetAttribute{
										CustomType: fwtypes.SetOfStringEnumType[awstypes.ContentClassifier](),
										Optional:   true,
										PlanModifiers: []planmodifier.Set{
											setplanmodifier.RequiresReplace(),
										},
									},
								},
							},
						},
						"data_source": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobDataSourceModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"s3_data_source": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobS3DataSourceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"manifest_s3_uri": schema.StringAttribute{
													Required: true,
													Validators: []validator.String{
														stringvalidator.RegexMatches(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
													},
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
									"sns_data_source": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobSNSDataSourceModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrSNSTopicARN: schema.StringAttribute{
													CustomType: fwtypes.ARNType,
													Required:   true,
													PlanModifiers: []planmodifier.String{
														stringplanmodifier.RequiresReplace(),
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
			"labeling_job_algorithms_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobAlgorithmsConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"initial_active_learning_model_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"labeling_job_algorithm_specification_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"labeling_job_resource_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobResourceConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"volume_kms_key_id": schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									names.AttrVPCConfig: schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										PlanModifiers: []planmodifier.List{
											listplanmodifier.RequiresReplace(),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrSecurityGroupIDs: schema.SetAttribute{
													CustomType:  fwtypes.SetOfStringType,
													ElementType: types.StringType,
													Required:    true,
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
													},
												},
												names.AttrSubnets: schema.SetAttribute{
													CustomType:  fwtypes.SetOfStringType,
													ElementType: types.StringType,
													Required:    true,
													PlanModifiers: []planmodifier.Set{
														setplanmodifier.RequiresReplace(),
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
			"output_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[labelingJobOutputConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKMSKeyID: schema.StringAttribute{
							Optional: true,
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"s3_output_path": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(httpsOrS3URIRegexp, "must be HTTPS or Amazon S3 URI"),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrSNSTopicARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *labelingJobResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data labelingJobResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.LabelingJobName)
	var input sagemaker.CreateLabelingJobInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func(ctx context.Context) (any, error) {
		return conn.CreateLabelingJob(ctx, &input)
	}, ErrCodeValidationException)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SageMaker AI Labeling Job (%s)", name), err.Error())
		return
	}

	const (
		timeout = 5 * time.Minute
	)
	output, err := waitLabelingJobInitialized(ctx, conn, name, timeout)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for SageMaker AI Labeling Job (%s) initialize", name), err.Error())
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *labelingJobResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data labelingJobResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.LabelingJobName)
	output, err := findLabelingJobByName(ctx, conn, name)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SageMaker AI Labeling Job (%s)", name), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *labelingJobResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data labelingJobResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SageMakerClient(ctx)

	if status := data.LabelingJobStatus.ValueEnum(); status != awstypes.LabelingJobStatusInProgress && status != awstypes.LabelingJobStatusCompleted {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, data.LabelingJobName)
	input := sagemaker.StopLabelingJobInput{
		LabelingJobName: aws.String(name),
	}
	_, err := conn.StopLabelingJob(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("stopping SageMaker AI Labeling Job (%s)", name), err.Error())
		return
	}
}

func (r *labelingJobResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("labeling_job_name"), request, response)
}

func findLabelingJobByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeLabelingJobOutput, error) {
	input := sagemaker.DescribeLabelingJobInput{
		LabelingJobName: aws.String(name),
	}

	output, err := findLabelingJob(ctx, conn, &input)

	if err != nil {
		return nil, err
	}

	if status := output.LabelingJobStatus; status == awstypes.LabelingJobStatusStopping || status == awstypes.LabelingJobStatusStopped {
		return nil, &retry.NotFoundError{
			Message: string(status),
		}
	}

	return output, nil
}

func findLabelingJob(ctx context.Context, conn *sagemaker.Client, input *sagemaker.DescribeLabelingJobInput) (*sagemaker.DescribeLabelingJobOutput, error) {
	output, err := conn.DescribeLabelingJob(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusLabelingJob(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findLabelingJobByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.LabelingJobStatus), nil
	}
}

func waitLabelingJobInitialized(ctx context.Context, conn *sagemaker.Client, name string, timeout time.Duration) (*sagemaker.DescribeLabelingJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LabelingJobStatusInitializing),
		Target:  enum.Slice(awstypes.LabelingJobStatusInProgress, awstypes.LabelingJobStatusCompleted),
		Refresh: statusLabelingJob(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*sagemaker.DescribeLabelingJobOutput); ok {
		return output, err
	}

	return nil, err
}

type labelingJobResourceModel struct {
	framework.WithRegionModel
	FailureReason               types.String                                                        `tfsdk:"failure_reason"`
	HumanTaskConfig             fwtypes.ListNestedObjectValueOf[humanTaskConfigModel]               `tfsdk:"human_task_config"`
	InputConfig                 fwtypes.ListNestedObjectValueOf[labelingJobInputConfigModel]        `tfsdk:"input_config"`
	JobReferenceCode            types.String                                                        `tfsdk:"job_reference_code"`
	LabelAttributeName          types.String                                                        `tfsdk:"label_attribute_name"`
	LabelCategoryConfigS3URI    types.String                                                        `tfsdk:"label_category_config_s3_uri"`
	LabelCounters               fwtypes.ListNestedObjectValueOf[labelCountersModel]                 `tfsdk:"label_counters"`
	LabelingJobAlgorithmsConfig fwtypes.ListNestedObjectValueOf[labelingJobAlgorithmsConfigModel]   `tfsdk:"labeling_job_algorithms_config"`
	LabelingJobARN              types.String                                                        `tfsdk:"labeling_job_arn"`
	LabelingJobName             types.String                                                        `tfsdk:"labeling_job_name"`
	LabelingJobStatus           fwtypes.StringEnum[awstypes.LabelingJobStatus]                      `tfsdk:"labeling_job_status"`
	OutputConfig                fwtypes.ListNestedObjectValueOf[labelingJobOutputConfigModel]       `tfsdk:"output_config"`
	RoleARN                     fwtypes.ARN                                                         `tfsdk:"role_arn"`
	StoppingConditions          fwtypes.ListNestedObjectValueOf[labelingJobStoppingConditionsModel] `tfsdk:"stopping_conditions"`
	Tags                        tftags.Map                                                          `tfsdk:"tags"`
	TagsAll                     tftags.Map                                                          `tfsdk:"tags_all"`
}

type humanTaskConfigModel struct {
	AnnotationConsolidationConfig     fwtypes.ListNestedObjectValueOf[annotationConsolidationConfigModel] `tfsdk:"annotation_consolidation_config"`
	MaxConcurrentTaskCount            types.Int32                                                         `tfsdk:"max_concurrent_task_count"`
	NumberOfHumanWorkersPerDataObject types.Int32                                                         `tfsdk:"number_of_human_workers_per_data_object"`
	PreHumanTaskLambdaARN             fwtypes.ARN                                                         `tfsdk:"pre_human_task_lambda_arn"`
	PublicWorkforceTaskPrice          fwtypes.ListNestedObjectValueOf[publicWorkforceTaskPriceModel]      `tfsdk:"public_workforce_task_price"`
	TaskAvailabilityLifetimeInSeconds types.Int32                                                         `tfsdk:"task_availability_lifetime_in_seconds"`
	TaskDescription                   types.String                                                        `tfsdk:"task_description"`
	TaskKeywords                      fwtypes.SetOfString                                                 `tfsdk:"task_keywords"`
	TaskTimeLimitInSeconds            types.Int32                                                         `tfsdk:"task_time_limit_in_seconds"`
	TaskTitle                         types.String                                                        `tfsdk:"task_title"`
	UIConfig                          fwtypes.ListNestedObjectValueOf[uiConfigModel]                      `tfsdk:"ui_config"`
	WorkteamARN                       fwtypes.ARN                                                         `tfsdk:"workteam_arn"`
}

type annotationConsolidationConfigModel struct {
	AnnotationConsolidationLambdaARN fwtypes.ARN `tfsdk:"annotation_consolidation_lambda_arn"`
}

type publicWorkforceTaskPriceModel struct {
	AmountInUSD fwtypes.ListNestedObjectValueOf[usdModel] `tfsdk:"amount_in_usd"`
}

type usdModel struct {
	Cents                 types.Int32 `tfsdk:"cents"`
	Dollars               types.Int32 `tfsdk:"dollars"`
	TenthFractionsOfACent types.Int32 `tfsdk:"tenth_fractions_of_a_cent"`
}

type uiConfigModel struct {
	HumanTaskUIARN  fwtypes.ARN  `tfsdk:"human_task_ui_arn"`
	UITemplateS3URI types.String `tfsdk:"ui_template_s3_uri"`
}

type labelingJobInputConfigModel struct {
	DataAttributes fwtypes.ListNestedObjectValueOf[labelingJobDataAttributesModel] `tfsdk:"data_attributes"`
	DataSource     fwtypes.ListNestedObjectValueOf[labelingJobDataSourceModel]     `tfsdk:"data_source"`
}

type labelingJobDataSourceModel struct {
	S3DataSource  fwtypes.ListNestedObjectValueOf[labelingJobS3DataSourceModel]  `tfsdk:"s3_data_source"`
	SNSDataSource fwtypes.ListNestedObjectValueOf[labelingJobSNSDataSourceModel] `tfsdk:"sns_data_source"`
}

type labelingJobS3DataSourceModel struct {
	ManifestS3URI types.String `tfsdk:"manifest_s3_uri"`
}

type labelingJobSNSDataSourceModel struct {
	SNSTopicARN fwtypes.ARN `tfsdk:"sns_topic_arn"`
}

type labelingJobDataAttributesModel struct {
	ContentClassifiers fwtypes.SetOfStringEnum[awstypes.ContentClassifier] `tfsdk:"content_classifiers"`
}

type labelingJobAlgorithmsConfigModel struct {
	InitialActiveLearningModelARN        fwtypes.ARN                                                     `tfsdk:"initial_active_learning_model_arn"`
	LabelingJobAlgorithmSpecificationARN fwtypes.ARN                                                     `tfsdk:"labeling_job_algorithm_specification_arn"`
	LabelingJobResourceConfig            fwtypes.ListNestedObjectValueOf[labelingJobResourceConfigModel] `tfsdk:"labeling_job_resource_config"`
}

type labelingJobResourceConfigModel struct {
	VolumeKMSKeyID types.String                                    `tfsdk:"volume_kms_key_id"`
	VPCConfig      fwtypes.ListNestedObjectValueOf[vpcConfigModel] `tfsdk:"vpc_config"`
}

type vpcConfigModel struct {
	SecurityGroupIDs fwtypes.SetOfString `tfsdk:"security_group_ids"`
	Subnets          fwtypes.SetOfString `tfsdk:"subnets"`
}

type labelingJobOutputConfigModel struct {
	KMSKeyID     types.String `tfsdk:"kms_key_id"`
	S3OutputPath types.String `tfsdk:"s3_output_path"`
	SNSTopicARN  fwtypes.ARN  `tfsdk:"sns_topic_arn"`
}

type labelingJobStoppingConditionsModel struct {
	MaxHumanLabeledObjectCount         types.Int32 `tfsdk:"max_human_labeled_object_count"`
	MaxPercentageOfInputDatasetLabeled types.Int32 `tfsdk:"max_percentage_of_input_dataset_labeled"`
}

type labelCountersModel struct {
	FailedNonRetryableError types.Int64 `tfsdk:"failed_non_retryable_error"`
	HumanLabeled            types.Int64 `tfsdk:"human_labeled"`
	MachineLabeled          types.Int64 `tfsdk:"machine_labeled"`
	TotalLabeled            types.Int64 `tfsdk:"total_labeled"`
	Unlabeled               types.Int64 `tfsdk:"unlabeled"`
}
