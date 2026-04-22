// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrock_model_invocation_logging_configuration", name="Model Invocation Logging Configuration")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(preIdentityVersion="v5.100.0")
func newModelInvocationLoggingConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &modelInvocationLoggingConfigurationResource{}, nil
}

type modelInvocationLoggingConfigurationResource struct {
	framework.ResourceWithModel[modelInvocationLoggingConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *modelInvocationLoggingConfigurationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
		},
		Blocks: map[string]schema.Block{
			"logging_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[loggingConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"embedding_data_delivery_enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						"image_data_delivery_enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						"text_data_delivery_enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
						"video_data_delivery_enabled": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(true),
						},
					},
					Blocks: map[string]schema.Block{
						"cloudwatch_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[cloudWatchConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrLogGroupName: schema.StringAttribute{
										Required: true,
									},
									names.AttrRoleARN: schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
								},
								Blocks: map[string]schema.Block{
									"large_data_delivery_s3_config": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												names.AttrBucketName: schema.StringAttribute{
													Required: true,
												},
												"key_prefix": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"s3_config": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrBucketName: schema.StringAttribute{
										Required: true,
									},
									"key_prefix": schema.StringAttribute{
										Optional: true,
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

func (r *modelInvocationLoggingConfigurationResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := modelInvocationLoggingConfigurationSchemaV0(ctx)

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeModelInvocationLoggingConfigurationFromV0,
		},
	}
}

func (r *modelInvocationLoggingConfigurationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
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
	data.ID = types.StringValue(r.Meta().Region(ctx))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *modelInvocationLoggingConfigurationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data modelInvocationLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	output, err := findModelInvocationLoggingConfiguration(ctx, conn)

	if retry.NotFound(err) {
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

func (r *modelInvocationLoggingConfigurationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
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

func (r *modelInvocationLoggingConfigurationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data modelInvocationLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)

	input := bedrock.DeleteModelInvocationLoggingConfigurationInput{}
	_, err := conn.DeleteModelInvocationLoggingConfiguration(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Model Invocation Logging Configuration (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

func (r *modelInvocationLoggingConfigurationResource) putModelInvocationLoggingConfiguration(ctx context.Context, data *modelInvocationLoggingConfigurationResourceModel) diag.Diagnostics {
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
	_, err := tfresource.RetryWhenIsAErrorMessageContains[any, *awstypes.ValidationException](ctx, propagationTimeout,
		func(ctx context.Context) (any, error) {
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
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type modelInvocationLoggingConfigurationResourceModel struct {
	framework.WithRegionModel
	ID            types.String                                        `tfsdk:"id"`
	LoggingConfig fwtypes.ListNestedObjectValueOf[loggingConfigModel] `tfsdk:"logging_config"`
}

type loggingConfigModel struct {
	CloudWatchConfig             fwtypes.ListNestedObjectValueOf[cloudWatchConfigModel] `tfsdk:"cloudwatch_config"`
	EmbeddingDataDeliveryEnabled types.Bool                                             `tfsdk:"embedding_data_delivery_enabled"`
	ImageDataDeliveryEnabled     types.Bool                                             `tfsdk:"image_data_delivery_enabled"`
	S3Config                     fwtypes.ListNestedObjectValueOf[s3ConfigModel]         `tfsdk:"s3_config"`
	TextDataDeliveryEnabled      types.Bool                                             `tfsdk:"text_data_delivery_enabled"`
	VideoDataDeliveryEnabled     types.Bool                                             `tfsdk:"video_data_delivery_enabled"`
}

type cloudWatchConfigModel struct {
	LargeDataDeliveryS3Config fwtypes.ListNestedObjectValueOf[s3ConfigModel] `tfsdk:"large_data_delivery_s3_config"`
	LogGroupName              types.String                                   `tfsdk:"log_group_name"`
	RoleArn                   fwtypes.ARN                                    `tfsdk:"role_arn"`
}

type s3ConfigModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	KeyPrefix  types.String `tfsdk:"key_prefix"`
}
