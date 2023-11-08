// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	bedrock_types "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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

func (r *resourceCustomModel) Schema(ctx context.Context, request resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"base_model_id": schema.StringAttribute{
				Required: true,
				// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(arn:aws(-[^:]+)?:bedrock:[a-z0-9-]{1,20}:(([0-9]{12}:custom-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}/[a-z0-9]{12})|(:foundation-model/[a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([a-z0-9-]{1,63}[.]){0,2}[a-z0-9-]{1,63}([:][a-z0-9-]{1,63}){0,2})))|([a-z0-9-]{1,63}[.]{1}[a-z0-9-]{1,63}([.]?[a-z0-9-]{1,63})([:][a-z0-9-]{1,63}){0,2})|(([0-9a-zA-Z][_-]?)+)$`), "minimum length of 1. Maximum length of 2048."),
			},
			"client_request_token": schema.StringAttribute{
				Optional: true,
				// ForceNew:     true,
				// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[a-zA-Z0-9](-*[a-zA-Z0-9])*$`), "minimum length of 1. Maximum length of 256."),
			},
			"custom_model_kms_key_id": schema.StringAttribute{
				Optional: true,
				// ForceNew:     true,
				// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws(-[^:]+)?:kms:[a-zA-Z0-9-]*:[0-9]{12}:((key/[a-zA-Z0-9-]{36})|(alias/[a-zA-Z0-9-_/]+))$`), "minimum length of 1. Maximum length of 2048."),
			},
			"custom_model_name": schema.StringAttribute{
				Required: true,
				// ForceNew:     true,
				// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^([0-9a-zA-Z][_-]?)+$`), "minimum length of 1. Maximum length of 63."),
			},
			"hyper_parameters": schema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				// ForceNew: true,
				// Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"job_name": schema.StringAttribute{
				Required: true,
				// ForceNew: true,
			},
			"job_tags": tftags.TagsAttribute(),
			"output_data_config": schema.StringAttribute{
				Required: true,
				// ForceNew:     true,
				// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$`), "minimum length of 1. Maximum length of 1024."),
			},
			"role_arn": schema.StringAttribute{
				Required: true,
				// ForceNew:     true,
				// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws(-[^:]+)?:iam::([0-9]{12})?:role/.+$`), "minimum length of 1. Maximum length of 2048."),
			},
			"training_data_config": schema.StringAttribute{
				Required: true,
				// ForceNew:     true,
				// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^s3://[a-z0-9][\.\-a-z0-9]{1,61}[a-z0-9](/.*)?$`), "minimum length of 1. Maximum length of 1024."),
			},
			"base_model_arn": schema.StringAttribute{
				Computed: true,
			},
			"creation_time": schema.StringAttribute{
				Computed: true,
			},
			"job_arn": schema.StringAttribute{
				Computed: true,
			},
			"model_arn": schema.StringAttribute{
				Computed: true,
			},
			"model_kms_key_arn": schema.StringAttribute{
				Computed: true,
			},
			"model_name": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"validation_data_config": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeBetween(0, 10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"validators": schema.SetAttribute{
							ElementType: types.StringType,
							Optional:    true,
						},
					},
				},
			},
			"vpc_config": schema.ListNestedBlock{
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"security_group_ids": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
						},
						"subnet_ids": schema.SetAttribute{
							ElementType: types.StringType,
							Required:    true,
							// ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[-0-9a-zA-Z]+$`), "minimum length of 1. Maximum length of 32."),
						},
					},
				},
			},
			"training_metrics": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"training_loss": schema.Float64Attribute{
						Computed: true,
					},
				},
			},
			"validation_metrics": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"validation_loss": schema.Float64Attribute{
							Computed: true,
						},
					},
				},
			},
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *resourceCustomModel) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceCustomModelModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)
	defaultTagsConfig := r.Meta().DefaultTagsConfig
	job_tags := defaultTagsConfig.MergeTags(tftags.New(ctx, data.JobTags))

	input := &bedrock.CreateModelCustomizationJobInput{
		BaseModelIdentifier: data.BaseModelId.ValueStringPointer(),
		CustomModelName:     data.CustomModelName.ValueStringPointer(),
		JobName:             data.JobName.ValueStringPointer(),
		RoleArn:             data.RoleArn.ValueStringPointer(),
		TrainingDataConfig: &bedrock_types.TrainingDataConfig{
			S3Uri: data.TrainingDataConfig.ValueStringPointer(),
		},
		OutputDataConfig: &bedrock_types.OutputDataConfig{
			S3Uri: data.OutputDataConfig.ValueStringPointer(),
		},
		CustomModelTags:      getTagsIn(ctx),
		ValidationDataConfig: expandValidationDataConfig(ctx, data.ValidationDataConfig),
	}

	var hp map[string]string
	resp.Diagnostics.Append(data.HyperParameters.ElementsAs(ctx, &hp, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.HyperParameters = hp

	if !data.ClientRequestToken.IsNull() {
		input.ClientRequestToken = data.ClientRequestToken.ValueStringPointer()
	}
	if !data.CustomModelKmsKeyId.IsNull() {
		input.CustomModelKmsKeyId = data.CustomModelKmsKeyId.ValueStringPointer()
	}
	var vpcConfigs []vpcConfig
	resp.Diagnostics.Append(data.VpcConfig.ElementsAs(ctx, &vpcConfigs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	vpcConfig := expandVpcConfig(ctx, vpcConfigs)
	if len(vpcConfigs) > 0 {
		input.VpcConfig = &vpcConfig[0]
	}

	if len(job_tags) > 0 {
		input.JobTags = Tags(tftags.New(ctx, job_tags.IgnoreAWS()))
	}

	tflog.Info(ctx, "CreateModelCustomizationJobInput:", map[string]any{
		"BaseModelIdentifier":  input.BaseModelIdentifier,
		"ClientRequestToken":   input.ClientRequestToken,
		"CustomModelName":      input.CustomModelName,
		"CustomModelKmsKeyId":  input.CustomModelKmsKeyId,
		"JobName":              input.JobName,
		"RoleArn":              input.RoleArn,
		"OutputDataConfig":     input.OutputDataConfig.S3Uri,
		"TrainingDataConfig":   input.TrainingDataConfig.S3Uri,
		"ValidationDataConfig": input.ValidationDataConfig,
		"VpcConfig":            input.VpcConfig,
	})

	jobStart, err := conn.CreateModelCustomizationJob(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("creating Bedrock Custom Model Customization Job", err.Error())
		return
	}

	// Successfully started job. Save the id now
	data.ID = data.CustomModelName
	// Also save job arn into state now incase we need to cancel and destroy.
	data.JobArn = flex.StringToFramework(ctx, jobStart.JobArn)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	err = waitForModelCustomizationJob(ctx, conn, *jobStart.JobArn, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("failed to complete model customisation job", err.Error())
		return
	}

	resp.Diagnostics.Append(data.refresh(ctx, conn)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceCustomModel) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceCustomModelModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	resp.Diagnostics.Append(data.refresh(ctx, conn)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceCustomModel) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Trace(ctx, "Update not supported.")
}

func (r *resourceCustomModel) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	_, err := conn.DeleteModelInvocationLoggingConfiguration(ctx, &bedrock.DeleteModelInvocationLoggingConfigurationInput{})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete model invocation logging configuration", err.Error())
		return
	}
}

func (r *resourceCustomModel) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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

type resourceCustomModelModel struct {
	ID                   types.String          `tfsdk:"id"`
	BaseModelId          types.String          `tfsdk:"base_model_id"`
	ClientRequestToken   types.String          `tfsdk:"client_request_token"`
	CustomModelKmsKeyId  types.String          `tfsdk:"custom_model_kms_key_id"`
	CustomModelName      types.String          `tfsdk:"custom_model_name"`
	HyperParameters      types.Map             `tfsdk:"hyper_parameters"`
	JobName              types.String          `tfsdk:"job_name"`
	JobTags              types.Map             `tfsdk:"job_tags"`
	OutputDataConfig     types.String          `tfsdk:"output_data_config"`
	RoleArn              types.String          `tfsdk:"role_arn"`
	TrainingDataConfig   types.String          `tfsdk:"training_data_config"`
	BaseModelArn         types.String          `tfsdk:"base_model_arn"`
	CreationTime         types.String          `tfsdk:"creation_time"`
	JobArn               types.String          `tfsdk:"job_arn"`
	ModelArn             types.String          `tfsdk:"model_arn"`
	ModelKmsKeyArn       types.String          `tfsdk:"model_kms_key_arn"`
	ModelName            types.String          `tfsdk:"model_name"`
	ValidationDataConfig *validationDataConfig `tfsdk:"validation_data_config"`
	VpcConfig            types.List            `tfsdk:"vpc_config"`
	TrainingMetrics      *trainingMetrics      `tfsdk:"training_metrics"`
	ValidationMetrics    []validationMetrics   `tfsdk:"validation_metrics"`
	Tags                 types.Map             `tfsdk:"tags"`
	TagsAll              types.Map             `tfsdk:"tags_all"`
	Timeouts             timeouts.Value        `tfsdk:"timeouts"`
}

func (data *resourceCustomModelModel) refresh(ctx context.Context, conn *bedrock.Client) diag.Diagnostics {
	var diags diag.Diagnostics

	modelId := data.ID
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

	return diags
}
