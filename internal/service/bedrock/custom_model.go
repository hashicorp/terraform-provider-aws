// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Custom Model")
// @Tags(identifierAttribute="job_arn")
func newCustomModelResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &customModelResource{}

	r.SetDefaultDeleteTimeout(120 * time.Minute)

	return r, nil
}

type customModelResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *customModelResource) Metadata(_ context.Context, request resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrock_custom_model"
}

func (r *customModelResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	// This resource is a composition of the following APIs. These APIs do not have consitently named attributes, so we will normalize them here.
	// - CreateModelCustomizationJob
	// - GetModelCustomizationJob
	// - GetCustomModel
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_model_identifier": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"custom_model_arn": schema.StringAttribute{
				Computed: true,
			},
			"custom_model_kms_key_id": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"custom_model_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
				},
			},
			"customization_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.CustomizationType](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hyperparameters": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"job_arn":    framework.ARNAttributeComputedOnly(),
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
			"job_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ModelCustomizationJobStatus](),
				Computed:   true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"training_metrics": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelTrainingMetricsModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"training_loss": types.Float64Type,
					},
				},
			},
			"validation_metrics": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelValidationMetricsModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"validation_loss": types.Float64Type,
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"output_data_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelOutputDataConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_uri": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								fwvalidators.S3URI(),
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
			"training_data_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelTrainingDataConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_uri": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								fwvalidators.S3URI(),
							},
						},
					},
				},
			},
			"validation_data_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelValidationDataConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"validator": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[customModelValidatorConfigModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(10),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"s3_uri": schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Validators: []validator.String{
											fwvalidators.S3URI()},
									},
								},
							},
						},
					},
				},
			},
			names.AttrVPCConfig: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelVPCConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSecurityGroupIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						names.AttrSubnetIDs: schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *customModelResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data customModelResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	input := &bedrock.CreateModelCustomizationJobInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientRequestToken = aws.String(id.UniqueId())
	input.CustomModelTags = getTagsIn(ctx)
	input.JobTags = getTagsIn(ctx)

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateModelCustomizationJob(ctx, input)
	}, errCodeValidationException, "Could not assume provided IAM role")

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Custom Model customization job", err.Error())

		return
	}

	jobARN := aws.ToString(outputRaw.(*bedrock.CreateModelCustomizationJobOutput).JobArn)
	job, err := findModelCustomizationJobByID(ctx, conn, jobARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model customization job (%s)", jobARN), err.Error())

		return
	}

	// Set values for unknowns.
	data.CustomizationType = fwtypes.StringEnumValue(job.CustomizationType)
	data.CustomModelARN = fwflex.StringToFramework(ctx, job.OutputModelArn)
	data.JobARN = fwflex.StringToFramework(ctx, job.JobArn)
	data.JobStatus = fwtypes.StringEnumValue(job.Status)
	data.TrainingMetrics = fwtypes.NewListNestedObjectValueOfNull[customModelTrainingMetricsModel](ctx)
	data.ValidationMetrics = fwtypes.NewListNestedObjectValueOfNull[customModelValidationMetricsModel](ctx)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *customModelResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data customModelResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().BedrockClient(ctx)

	jobARN := data.JobARN.ValueString()
	outputGJ, err := findModelCustomizationJobByID(ctx, conn, jobARN)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model customization job (%s)", jobARN), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGJ, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Some fields in GetModelCustomizationJobOutput have different names than in CreateModelCustomizationJobInput.
	data.CustomModelKmsKeyID = fwflex.StringToFrameworkARN(ctx, outputGJ.OutputModelKmsKeyArn)
	data.CustomModelName = fwflex.StringToFramework(ctx, outputGJ.OutputModelName)
	data.JobStatus = fwtypes.StringEnumValue(outputGJ.Status)
	// The base model ARN in GetCustomModelOutput can contain the model version and parameter count.
	baseModelARN := fwflex.StringFromFramework(ctx, data.BaseModelIdentifier)
	data.BaseModelIdentifier = fwflex.StringToFrameworkARN(ctx, outputGJ.BaseModelArn)
	if baseModelARN != nil {
		if old, err := arn.Parse(aws.ToString(baseModelARN)); err == nil {
			if new, err := arn.Parse(aws.ToString(outputGJ.BaseModelArn)); err == nil {
				if len(strings.SplitN(old.Resource, ":", 2)) == 1 {
					// Old ARN doesn't contain the model version and parameter count.
					new.Resource = strings.SplitN(new.Resource, ":", 2)[0]
					data.BaseModelIdentifier = fwtypes.ARNValue(new.String())
				}
			}
		}
	}

	if outputGJ.OutputModelArn != nil {
		customModelARN := aws.ToString(outputGJ.OutputModelArn)
		outputGM, err := findCustomModelByID(ctx, conn, customModelARN)

		if tfresource.NotFound(err) {
			response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
			response.State.RemoveResource(ctx)

			return
		}

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model (%s)", customModelARN), err.Error())

			return
		}

		var dataFromGetCustomModel customModelResourceModel
		response.Diagnostics.Append(fwflex.Flatten(ctx, outputGM, &dataFromGetCustomModel)...)
		if response.Diagnostics.HasError() {
			return
		}

		data.CustomModelARN = fwflex.StringToFramework(ctx, outputGM.ModelArn)
		data.TrainingMetrics = dataFromGetCustomModel.TrainingMetrics
		data.ValidationMetrics = dataFromGetCustomModel.ValidationMetrics
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *customModelResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new customModelResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Update is only called when `tags` are updated.
	// Set unknowns to the old (in state) values.
	new.CustomModelARN = old.CustomModelARN
	new.JobStatus = old.JobStatus
	new.TrainingMetrics = old.TrainingMetrics
	new.ValidationMetrics = old.ValidationMetrics

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *customModelResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data customModelResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	if data.JobStatus.ValueEnum() == awstypes.ModelCustomizationJobStatusInProgress {
		jobARN := data.JobARN.ValueString()
		input := &bedrock.StopModelCustomizationJobInput{
			JobIdentifier: aws.String(jobARN),
		}

		_, err := conn.StopModelCustomizationJob(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("stopping Bedrock Custom Model customization job (%s)", jobARN), err.Error())

			return
		}

		if _, err := waitModelCustomizationJobStopped(ctx, conn, jobARN, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Custom Model customization job (%s) stop", jobARN), err.Error())

			return
		}
	}

	if !data.CustomModelARN.IsNull() {
		_, err := conn.DeleteCustomModel(ctx, &bedrock.DeleteCustomModelInput{
			ModelIdentifier: fwflex.StringFromFramework(ctx, data.CustomModelARN),
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Custom Model (%s)", data.ID.ValueString()), err.Error())

			return
		}
	}
}

func (r *customModelResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findCustomModelByID(ctx context.Context, conn *bedrock.Client, id string) (*bedrock.GetCustomModelOutput, error) {
	input := &bedrock.GetCustomModelInput{
		ModelIdentifier: aws.String(id),
	}

	output, err := conn.GetCustomModel(ctx, input)

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

func findModelCustomizationJobByID(ctx context.Context, conn *bedrock.Client, id string) (*bedrock.GetModelCustomizationJobOutput, error) {
	input := &bedrock.GetModelCustomizationJobInput{
		JobIdentifier: aws.String(id),
	}

	output, err := findModelCustomizationJob(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := output.Status; status == awstypes.ModelCustomizationJobStatusStopped {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findModelCustomizationJob(ctx context.Context, conn *bedrock.Client, input *bedrock.GetModelCustomizationJobInput) (*bedrock.GetModelCustomizationJobOutput, error) {
	output, err := conn.GetModelCustomizationJob(ctx, input)

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

func statusModelCustomizationJob(ctx context.Context, conn *bedrock.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &bedrock.GetModelCustomizationJobInput{
			JobIdentifier: aws.String(id),
		}
		output, err := findModelCustomizationJob(ctx, conn, input)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitModelCustomizationJobCompleted(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*bedrock.GetModelCustomizationJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ModelCustomizationJobStatusInProgress),
		Target:  enum.Slice(awstypes.ModelCustomizationJobStatusCompleted),
		Refresh: statusModelCustomizationJob(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*bedrock.GetModelCustomizationJobOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureMessage)))

		return output, err
	}

	return nil, err
}

func waitModelCustomizationJobStopped(ctx context.Context, conn *bedrock.Client, id string, timeout time.Duration) (*bedrock.GetModelCustomizationJobOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ModelCustomizationJobStatusStopping),
		Target:  enum.Slice(awstypes.ModelCustomizationJobStatusStopped),
		Refresh: statusModelCustomizationJob(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*bedrock.GetModelCustomizationJobOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureMessage)))

		return output, err
	}

	return nil, err
}

type customModelResourceModel struct {
	BaseModelIdentifier  fwtypes.ARN                                                           `tfsdk:"base_model_identifier"`
	CustomModelARN       types.String                                                          `tfsdk:"custom_model_arn"`
	CustomModelKmsKeyID  fwtypes.ARN                                                           `tfsdk:"custom_model_kms_key_id"`
	CustomModelName      types.String                                                          `tfsdk:"custom_model_name"`
	CustomizationType    fwtypes.StringEnum[awstypes.CustomizationType]                        `tfsdk:"customization_type"`
	HyperParameters      fwtypes.MapValueOf[types.String]                                      `tfsdk:"hyperparameters"`
	ID                   types.String                                                          `tfsdk:"id"`
	JobARN               types.String                                                          `tfsdk:"job_arn"`
	JobName              types.String                                                          `tfsdk:"job_name"`
	JobStatus            fwtypes.StringEnum[awstypes.ModelCustomizationJobStatus]              `tfsdk:"job_status"`
	OutputDataConfig     fwtypes.ListNestedObjectValueOf[customModelOutputDataConfigModel]     `tfsdk:"output_data_config"`
	RoleARN              fwtypes.ARN                                                           `tfsdk:"role_arn"`
	Tags                 types.Map                                                             `tfsdk:"tags"`
	TagsAll              types.Map                                                             `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                                        `tfsdk:"timeouts"`
	TrainingDataConfig   fwtypes.ListNestedObjectValueOf[customModelTrainingDataConfigModel]   `tfsdk:"training_data_config"`
	TrainingMetrics      fwtypes.ListNestedObjectValueOf[customModelTrainingMetricsModel]      `tfsdk:"training_metrics"`
	ValidationDataConfig fwtypes.ListNestedObjectValueOf[customModelValidationDataConfigModel] `tfsdk:"validation_data_config"`
	ValidationMetrics    fwtypes.ListNestedObjectValueOf[customModelValidationMetricsModel]    `tfsdk:"validation_metrics"`
	VPCConfig            fwtypes.ListNestedObjectValueOf[customModelVPCConfigModel]            `tfsdk:"vpc_config"`
}

func (data *customModelResourceModel) InitFromID() error {
	data.JobARN = data.ID

	return nil
}

func (data *customModelResourceModel) setID() {
	data.ID = data.JobARN
}

type customModelOutputDataConfigModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}

type customModelTrainingDataConfigModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}

type customModelTrainingMetricsModel struct {
	TrainingLoss types.Float64 `tfsdk:"training_loss"`
}

type customModelValidationDataConfigModel struct {
	Validators fwtypes.ListNestedObjectValueOf[customModelValidatorConfigModel] `tfsdk:"validator"`
}

type customModelValidationMetricsModel struct {
	ValidationLoss types.Float64 `tfsdk:"validation_loss"`
}

type customModelValidatorConfigModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}

type customModelVPCConfigModel struct {
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
}
