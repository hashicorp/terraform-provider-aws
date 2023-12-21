// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource
// @Tags(identifierAttribute="model_arn")
func newResourceCustomModel(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceCustomModel{}
	r.SetDefaultCreateTimeout(120 * time.Minute)
	return r, nil
}

type resourceCustomModel struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceCustomModel) Metadata(_ context.Context, request resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrock_custom_model"
}

// This resource is a composition of the following APIs. These APIs do not have consitently named attributes, so we will normalize them here.
// - CreateModelCustomizationJob
// - GetModelCustomizationJob
// - GetCustomModel
func (r *resourceCustomModel) Schema(ctx context.Context, request resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"base_model_arn": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
				},
			},
			"customization_type": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					enum.FrameworkValidate[awstypes.CustomizationType](),
				},
			},
			"hyper_parameters": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"job_arn": schema.StringAttribute{
				Computed: true,
			},
			"job_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9\+\-\.])*$`),
						"must be up to 63 letters (uppercase and lowercase), numbers, plus sign, dashes, and dots, and must start with an alphanumeric"),
				},
			},
			"job_tags": tftags.TagsAttribute(),
			"job_role_arn": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
				},
			},
			"job_status": schema.StringAttribute{
				Computed: true,
			},
			"kms_key_arn": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 2048),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"job_vpc_config": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[vpcConfig](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"security_group_ids": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"subnet_ids": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
			"output_data_config": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[outputDataConfig](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_uri": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1024),
								stringvalidator.RegexMatches(regexache.MustCompile(`^s3://`), "minimum length of 1. Maximum length of 1024. Must be an S3 URI"),
							},
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
			"training_data_config": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[trainingDataConfig](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_uri": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 1024),
								stringvalidator.RegexMatches(regexache.MustCompile(`^s3://`), "minimum length of 1. Maximum length of 1024. Must be an S3 URI"),
							},
						},
					},
				},
			},
			"training_metrics": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[trainingMetrics](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"training_loss": schema.Float64Attribute{
							Computed: true,
						},
					},
				},
			},
			"validation_data_config": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[validationDataConfig](ctx),
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"validator": schema.ListNestedBlock{
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtMost(10),
							},
							CustomType: fwtypes.NewListNestedObjectTypeOf[validatorConfig](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"s3_uri": schema.StringAttribute{
										Required: true,
										Validators: []validator.String{
											stringvalidator.LengthBetween(1, 1024),
											stringvalidator.RegexMatches(regexache.MustCompile(`^s3://`), "minimum length of 1. Maximum length of 1024. Must be an S3 URI"),
										},
									},
								},
							},
						},
					},
				},
			},
			"validation_metrics": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[validationMetrics](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"validation_loss": schema.Float64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (r *resourceCustomModel) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var data resourceCustomModelData
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var outputData []outputDataConfig
	response.Diagnostics.Append(data.OutputDataConfig.ElementsAs(ctx, &outputData, false)...)
	if response.Diagnostics.HasError() {
		return
	}
	var trainingData []trainingDataConfig
	response.Diagnostics.Append(data.TrainingDataConfig.ElementsAs(ctx, &trainingData, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &bedrock.CreateModelCustomizationJobInput{
		BaseModelIdentifier: data.BaseModelArn.ValueStringPointer(),
		CustomModelName:     data.Name.ValueStringPointer(),
		CustomModelTags:     getTagsIn(ctx),
		HyperParameters:     flex.ExpandFrameworkStringValueMap(ctx, data.HyperParameters),
		JobName:             data.JobName.ValueStringPointer(),
		OutputDataConfig:    expandOutputDataConfig(ctx, outputData),
		TrainingDataConfig:  expandTrainingDataConfig(ctx, trainingData),
		RoleArn:             data.JobRoleArn.ValueStringPointer(),
	}

	if !data.CustomizationType.IsNull() {
		input.CustomModelName = data.CustomizationType.ValueStringPointer()
	}
	if !data.KmsKeyArn.IsNull() {
		input.CustomModelKmsKeyId = data.KmsKeyArn.ValueStringPointer()
	}
	if !data.JobTags.IsNull() {
		input.JobTags = Tags(tftags.New(ctx, data.JobTags))
	}
	if !data.JobVpcConfig.IsNull() {
		var vpcData []vpcConfig
		response.Diagnostics.Append(data.JobVpcConfig.ElementsAs(ctx, &vpcData, false)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.VpcConfig = expandVPCConfig(ctx, vpcData)
	}
	if !data.ValidationDataConfig.IsNull() {
		var validationData []validationDataConfig
		response.Diagnostics.Append(data.ValidationDataConfig.ElementsAs(ctx, &validationData, false)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.ValidationDataConfig = expandValidationDataConfig(ctx, validationData, diag.Diagnostics{})
	}

	job, err := conn.CreateModelCustomizationJob(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, "ModelCustomizationJob", data.JobName.ValueString(), nil),
			err.Error(),
		)
		return
	}

	// Successfully started job. Save the id now
	// data.ID = flex.StringValueToFramework(ctx, "tf-acc-test-1531621220222582981")
	// data.JobArn = flex.StringValueToFramework(ctx, "arn:aws:bedrock:us-east-1:219858395663:model-customization-job/amazon.titan-text-express-v1:0:8k/pc2v9cmxjzlq")

	// Also save job arn into state now incase we need to cancel and destroy.
	data.JobArn = flex.StringToFramework(ctx, job.JobArn)

	response.Diagnostics.Append(data.refresh(ctx, conn)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceCustomModel) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceCustomModelData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	response.Diagnostics.Append(data.refresh(ctx, conn)...)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceCustomModel) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	tflog.Trace(ctx, "Update not supported.")
}

func (r *resourceCustomModel) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	_, err := conn.DeleteModelInvocationLoggingConfiguration(ctx, &bedrock.DeleteModelInvocationLoggingConfigurationInput{})
	if err != nil {
		response.Diagnostics.AddError("failed to delete model invocation logging configuration", err.Error())
		return
	}
}

func (r *resourceCustomModel) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func waitForModelCustomizationJob(ctx context.Context, conn *bedrock.Client, jobArn string, timeout time.Duration) error {
	return retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		jobEnd, err := conn.GetModelCustomizationJob(ctx, &bedrock.GetModelCustomizationJobInput{
			JobIdentifier: &jobArn,
		})
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("getting model customization job: %s", err))
		}

		tflog.Info(ctx, "GetModelCustomizationJobOuput:", map[string]any{
			"Status": jobEnd.Status,
		})

		switch jobEnd.Status {
		case "InProgress":
			return retry.RetryableError(fmt.Errorf("expected instance to be Completed but was in state %s", jobEnd.Status))
		case "Completed":
			return nil
		default:
			return retry.NonRetryableError(fmt.Errorf(*jobEnd.FailureMessage))
		}
	})
}

type resourceCustomModelData struct {
	Arn                  types.String                                          `tfsdk:"arn"`
	BaseModelArn         types.String                                          `tfsdk:"base_model_arn"`
	CustomizationType    types.String                                          `tfsdk:"customization_type"`
	HyperParameters      types.Map                                             `tfsdk:"hyper_parameters"`
	Id                   types.String                                          `tfsdk:"id"`
	JobArn               types.String                                          `tfsdk:"job_arn"`
	JobName              types.String                                          `tfsdk:"job_name"`
	JobTags              types.Map                                             `tfsdk:"job_tags"`
	JobRoleArn           types.String                                          `tfsdk:"job_role_arn"`
	JobStatus            types.String                                          `tfsdk:"job_status"`
	JobVpcConfig         fwtypes.ListNestedObjectValueOf[vpcConfig]            `tfsdk:"job_vpc_config"`
	KmsKeyArn            types.String                                          `tfsdk:"kms_key_arn"`
	Name                 types.String                                          `tfsdk:"name"`
	OutputDataConfig     fwtypes.ListNestedObjectValueOf[outputDataConfig]     `tfsdk:"output_data_config"`
	TrainingDataConfig   fwtypes.ListNestedObjectValueOf[trainingDataConfig]   `tfsdk:"training_data_config"`
	TrainingMetrics      fwtypes.ListNestedObjectValueOf[trainingMetrics]      `tfsdk:"training_metrics"`
	ValidationDataConfig fwtypes.ListNestedObjectValueOf[validationDataConfig] `tfsdk:"validation_data_config"`
	ValidationMetrics    fwtypes.ListNestedObjectValueOf[validationMetrics]    `tfsdk:"validation_metrics"`
	Tags                 types.Map                                             `tfsdk:"tags"`
	TagsAll              types.Map                                             `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                        `tfsdk:"timeouts"`
}

func (data *resourceCustomModelData) refresh(ctx context.Context, conn *bedrock.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	/* 	modelId := data.ID
	   	input := &bedrock.GetCustomModelInput{
	   		ModelIdentifier: modelId.ValueStringPointer(),
	   	}
	   	output, err := conn.GetCustomModel(ctx, input)

	   	if err != nil {
	   		// If we got here, the state has the model name and the job arn.
	   		// Should we check for tainted state instead?
	   		tflog.Info(ctx, "resourceCustomModelRead: Error reading Bedrock Custom Model. Ignoring to allow destroy to attempt to cleanup.")
	   		return diags
	   	}

	   	data.BaseModelArn = flex.StringToFramework(ctx, output.BaseModelArn)
	   	data.CreationTime = flex.StringValueToFramework(ctx, output.CreationTime.Format((time.RFC3339)))
	   	data.HyperParameters = flex.FlattenFrameworkStringValueMap(ctx, output.HyperParameters)
	   	data.JobArn = flex.StringToFramework(ctx, output.JobArn)
	   	// This is nil in the model object - could be a bug
	   	// However this is already in state so we can skip setting this here and avoid a forced update due to value change.
	   	// d.Set("job_name", model.JobName)
	   	data.ModelArn = flex.StringToFramework(ctx, output.ModelArn)
	   	data.ModelKmsKeyArn = flex.StringToFramework(ctx, output.ModelKmsKeyArn)
	   	data.ModelName = flex.StringToFramework(ctx, output.ModelName)
	   	data.OutputDataConfig = flex.StringToFramework(ctx, output.OutputDataConfig.S3Uri)
	   	data.TrainingDataConfig = flex.StringToFramework(ctx, output.TrainingDataConfig.S3Uri)
	   	data.TrainingMetrics = flattenTrainingMetrics(ctx, output.TrainingMetrics)
	   	data.ValidationDataConfig = flattenValidationDataConfig(ctx, output.ValidationDataConfig)
	   	data.ValidationMetrics = flattenValidationMetrics(ctx, output.ValidationMetrics)

	   	jobTags, err := listTags(ctx, conn, *output.JobArn)
	   	if err != nil {
	   		diags.AddError("reading Tags for Job", err.Error())
	   		return diags
	   	}
	   	data.JobTags = flex.FlattenFrameworkStringValueMap(ctx, jobTags.IgnoreAWS().Map())
	*/
	return diags
}
