// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Model Invocation Logging Configuration")
func newModelInvocationLoggingConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceModelInvocationLoggingConfiguration{}, nil
}

type resourceModelInvocationLoggingConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourceModelInvocationLoggingConfiguration) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrock_model_invocation_logging_configuration"
}

func (r *resourceModelInvocationLoggingConfiguration) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"logging_config": schema.SingleNestedBlock{
				CustomType: fwtypes.NewObjectTypeOf[loggingConfigModel](ctx),
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"embedding_data_delivery_enabled": schema.BoolAttribute{
						Required: true,
					},
					"image_data_delivery_enabled": schema.BoolAttribute{
						Required: true,
					},
					"text_data_delivery_enabled": schema.BoolAttribute{
						Required: true,
					},
				},
				Blocks: map[string]schema.Block{
					"cloudwatch_config": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[cloudWatchConfigModel](ctx),
						Attributes: map[string]schema.Attribute{
							names.AttrLogGroupName: schema.StringAttribute{
								// Required: true,
								Optional: true,
							},
							names.AttrRoleARN: schema.StringAttribute{
								CustomType: fwtypes.ARNType,
								Optional:   true,
							},
						},
						Blocks: map[string]schema.Block{
							"large_data_delivery_s3_config": schema.SingleNestedBlock{
								CustomType: fwtypes.NewObjectTypeOf[s3ConfigModel](ctx),
								Attributes: map[string]schema.Attribute{
									names.AttrBucketName: schema.StringAttribute{
										// Required: true,
										Optional: true,
									},
									"key_prefix": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
					"s3_config": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[s3ConfigModel](ctx),
						Attributes: map[string]schema.Attribute{
							names.AttrBucketName: schema.StringAttribute{
								// Required: true,
								Optional: true,
							},
							"key_prefix": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceModelInvocationLoggingConfiguration) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data modelInvocationLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.putModelInvocationLoggingConfiguration(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns.
	data.ID = types.StringValue(r.Meta().Region)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data modelInvocationLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	output, err := findModelInvocationLoggingConfiguration(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Model Invocation Logging Configuration (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data modelInvocationLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.putModelInvocationLoggingConfiguration(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data modelInvocationLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	_, err := conn.DeleteModelInvocationLoggingConfiguration(ctx, &bedrock.DeleteModelInvocationLoggingConfigurationInput{})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Model Invocation Logging Configuration (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *resourceModelInvocationLoggingConfiguration) putModelInvocationLoggingConfiguration(ctx context.Context, data *modelInvocationLoggingConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := r.Meta().BedrockClient(ctx)

	input := &bedrock.PutModelInvocationLoggingConfigurationInput{}
	diags.Append(fwflex.Expand(ctx, data, input)...)
	if diags.HasError() {
		return diags
	}

	// Example:
	//   ValidationException: Failed to validate permissions for log group: <group>, with role: <role>. Verify
	//   the IAM role permissions are correct.
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ValidationException](ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.PutModelInvocationLoggingConfiguration(ctx, input)
		},
		"Failed to validate permissions for log group",
	)

	if err != nil {
		diags.AddError("putting Bedrock Model Invocation Logging Configuration", err.Error())

		return diags
	}

	return diags
}

func findModelInvocationLoggingConfiguration(ctx context.Context, conn *bedrock.Client) (*bedrock.GetModelInvocationLoggingConfigurationOutput, error) {
	input := &bedrock.GetModelInvocationLoggingConfigurationInput{}

	output, err := conn.GetModelInvocationLoggingConfiguration(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.LoggingConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type modelInvocationLoggingConfigurationResourceModel struct {
	ID            types.String                              `tfsdk:"id"`
	LoggingConfig fwtypes.ObjectValueOf[loggingConfigModel] `tfsdk:"logging_config"`
}

type loggingConfigModel struct {
	CloudWatchConfig             fwtypes.ObjectValueOf[cloudWatchConfigModel] `tfsdk:"cloudwatch_config"`
	EmbeddingDataDeliveryEnabled types.Bool                                   `tfsdk:"embedding_data_delivery_enabled"`
	ImageDataDeliveryEnabled     types.Bool                                   `tfsdk:"image_data_delivery_enabled"`
	S3Config                     fwtypes.ObjectValueOf[s3ConfigModel]         `tfsdk:"s3_config"`
	TextDataDeliveryEnabled      types.Bool                                   `tfsdk:"text_data_delivery_enabled"`
}

type cloudWatchConfigModel struct {
	LargeDataDeliveryS3Config fwtypes.ObjectValueOf[s3ConfigModel] `tfsdk:"large_data_delivery_s3_config"`
	LogGroupName              types.String                         `tfsdk:"log_group_name"`
	RoleArn                   fwtypes.ARN                          `tfsdk:"role_arn"`
}

type s3ConfigModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	KeyPrefix  types.String `tfsdk:"key_prefix"`
}
