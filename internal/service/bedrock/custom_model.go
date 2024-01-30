// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
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
// @Tags(identifierAttribute="arn")
func newCustomModelResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &customModelResource{}

	r.SetDefaultCreateTimeout(120 * time.Minute)

	return r, nil
}

type customModelResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate
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
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"base_model_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
			"hyper_parameters": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
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
			"job_tags": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"job_role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"kms_key_arn": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.ARNType,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
			"job_vpc_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelVPCConfigModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"security_group_ids": schema.SetAttribute{
							CustomType:  fwtypes.SetOfStringType,
							Required:    true,
							ElementType: types.StringType,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
						},
						"subnet_ids": schema.SetAttribute{
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
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
		},
	}
}

func (r *customModelResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceCustomModelData
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

	// Some fields from CreateModelCustomizationJobInput don't match those from GetCustomModelOutput, so we need to set them manually.
	input.BaseModelIdentifier = fwflex.StringFromFramework(ctx, data.BaseModelARN)
	input.ClientRequestToken = aws.String(id.UniqueId())
	input.CustomModelKmsKeyId = fwflex.StringFromFramework(ctx, data.ModelKmsKeyARN)
	input.CustomModelName = fwflex.StringFromFramework(ctx, data.ModelName)
	input.CustomModelTags = getTagsIn(ctx)
	input.JobTags = getTagsIn(ctx)

	outputCJ, err := conn.CreateModelCustomizationJob(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Custom Model customization job", err.Error())

		return
	}

	jobARN := aws.ToString(outputCJ.JobArn)
	outputGJ, err := waitModelCustomizationJobCompleted(ctx, conn, jobARN, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Custom Model customization job (%s) complete", jobARN), err.Error())

		return
	}

	data.ModelARN = fwflex.StringToFramework(ctx, outputGJ.OutputModelArn)
	data.setID()

	// Set values for unknowns.
	// We need to read the model as not all fields are returned by GetModelCustomizationJob.
	outputGM, err := findCustomModelByID(ctx, conn, data.ModelARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGM, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *customModelResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceCustomModelData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().BedrockClient(ctx)

	output, err := findCustomModelByID(ctx, conn, data.ModelARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	jobARN := aws.ToString(output.JobArn)
	jobTags, err := listTags(ctx, conn, jobARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model Job (%s) tags", jobARN), err.Error())

		return
	}

	data.JobTags = fwflex.FlattenFrameworkStringValueMap(ctx, jobTags.IgnoreAWS().Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}
func (r *customModelResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceCustomModelData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	_, err := conn.DeleteCustomModel(ctx, &bedrock.DeleteCustomModelInput{
		ModelIdentifier: fwflex.StringFromFramework(ctx, data.ModelARN),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Custom Model (%s)", data.ID.ValueString()), err.Error())

		return
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
		output, err := findModelCustomizationJobByID(ctx, conn, id)

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

type resourceCustomModelData struct {
	BaseModelARN         fwtypes.ARN                                                           `tfsdk:"base_model_arn"`
	CustomizationType    fwtypes.StringEnum[awstypes.CustomizationType]                        `tfsdk:"customization_type"`
	HyperParameters      fwtypes.MapValueOf[types.String]                                      `tfsdk:"hyper_parameters"`
	ID                   types.String                                                          `tfsdk:"id"`
	JobARN               types.String                                                          `tfsdk:"job_arn"`
	JobName              types.String                                                          `tfsdk:"job_name"`
	JobTags              types.Map                                                             `tfsdk:"job_tags"`
	JobVPCConfig         fwtypes.ListNestedObjectValueOf[customModelVPCConfigModel]            `tfsdk:"job_vpc_config"`
	ModelARN             types.String                                                          `tfsdk:"arn"`
	ModelKmsKeyARN       fwtypes.ARN                                                           `tfsdk:"kms_key_arn"`
	ModelName            types.String                                                          `tfsdk:"name"`
	OutputDataConfig     fwtypes.ListNestedObjectValueOf[customModelOutputDataConfigModel]     `tfsdk:"output_data_config"`
	RoleARN              fwtypes.ARN                                                           `tfsdk:"job_role_arn"`
	Tags                 types.Map                                                             `tfsdk:"tags"`
	TagsAll              types.Map                                                             `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                                                        `tfsdk:"timeouts"`
	TrainingDataConfig   fwtypes.ListNestedObjectValueOf[customModelTrainingDataConfigModel]   `tfsdk:"training_data_config"`
	TrainingMetrics      fwtypes.ListNestedObjectValueOf[customModelTrainingMetricsModel]      `tfsdk:"training_metrics"`
	ValidationDataConfig fwtypes.ListNestedObjectValueOf[customModelValidationDataConfigModel] `tfsdk:"validation_data_config"`
	ValidationMetrics    fwtypes.ListNestedObjectValueOf[customModelValidationMetricsModel]    `tfsdk:"validation_metrics"`
}

func (data *resourceCustomModelData) InitFromID() error {
	data.ModelARN = data.ID

	return nil
}

func (data *resourceCustomModelData) setID() {
	data.ID = data.ModelARN
}
