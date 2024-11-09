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

// @FrameworkResource("aws_bedrock_knowledge_base_logging_configuration", name="Knowledge Base Logging Configuration")
func newKnowledgeBaseLoggingConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceKnowledgeBaseLoggingConfiguration{}, nil
}

type resourceKnowledgeBaseLoggingConfiguration struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourceKnowledgeBaseLoggingConfiguration) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrock_knowledge_base_logging_configuration"
}

func (r *resourceKnowledgeBaseLoggingConfiguration) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"knowledge_base_id": schema.StringAttribute{
				Required: true,
			},
			"tags": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
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
				},
				Blocks: map[string]schema.Block{
					"cloudwatch_config": schema.SingleNestedBlock{
						CustomType: fwtypes.NewObjectTypeOf[cloudWatchConfigModel](ctx),
						Attributes: map[string]schema.Attribute{
							"log_group_name": schema.StringAttribute{
								Required: true,
							},
							"role_arn": schema.StringAttribute{
								CustomType: fwtypes.ARNType,
								Required:   true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceKnowledgeBaseLoggingConfiguration) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data knowledgeBaseLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.putKnowledgeBaseLoggingConfiguration(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set the ID to the knowledge base ID
	data.ID = data.KnowledgeBaseID

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceKnowledgeBaseLoggingConfiguration) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data knowledgeBaseLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)
	output, err := findKnowledgeBaseLoggingConfiguration(ctx, conn, data.KnowledgeBaseID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Knowledge Base Logging Configuration (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceKnowledgeBaseLoggingConfiguration) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data knowledgeBaseLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(r.putKnowledgeBaseLoggingConfiguration(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceKnowledgeBaseLoggingConfiguration) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data knowledgeBaseLoggingConfigurationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)
	_, err := conn.DeleteKnowledgeBaseLoggingConfiguration(ctx, &bedrock.DeleteKnowledgeBaseLoggingConfigurationInput{
		KnowledgeBaseId: data.KnowledgeBaseID.ValueString(),
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Knowledge Base Logging Configuration (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

func (r *resourceKnowledgeBaseLoggingConfiguration) putKnowledgeBaseLoggingConfiguration(ctx context.Context, data *knowledgeBaseLoggingConfigurationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := r.Meta().BedrockClient(ctx)

	input := &bedrock.PutKnowledgeBaseLoggingConfigurationInput{
		KnowledgeBaseId: data.KnowledgeBaseID.ValueString(),
		LoggingConfig:   expandLoggingConfig(data.LoggingConfig),
		Tags:            expandTags(data.Tags),
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ValidationException](ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.PutKnowledgeBaseLoggingConfiguration(ctx, input)
		},
		"Failed to validate permissions for log group",
	)

	if err != nil {
		diags.AddError("putting Bedrock Knowledge Base Logging Configuration", err.Error())
		return diags
	}

	return diags
}

func findKnowledgeBaseLoggingConfiguration(ctx context.Context, conn *bedrock.Client, knowledgeBaseID string) (*bedrock.GetKnowledgeBaseLoggingConfigurationOutput, error) {
	input := &bedrock.GetKnowledgeBaseLoggingConfigurationInput{
		KnowledgeBaseId: knowledgeBaseID,
	}

	output, err := conn.GetKnowledgeBaseLoggingConfiguration(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || output.LoggingConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// Resource model for knowledge base logging configuration
type knowledgeBaseLoggingConfigurationResourceModel struct {
	ID              types.String                              `tfsdk:"id"`
	KnowledgeBaseID types.String                              `tfsdk:"knowledge_base_id"`
	LoggingConfig   fwtypes.ObjectValueOf[loggingConfigModel] `tfsdk:"logging_config"`
	Tags            types.Map                                 `tfsdk:"tags"`
}

// Configuration model for logging settings
type loggingConfigModel struct {
	EmbeddingDataDeliveryEnabled types.Bool                                   `tfsdk:"embedding_data_delivery_enabled"`
	CloudWatchConfig             fwtypes.ObjectValueOf[cloudWatchConfigModel] `tfsdk:"cloudwatch_config"`
}

// CloudWatch configuration model
type cloudWatchConfigModel struct {
	LogGroupName types.String `tfsdk:"log_group_name"`
	RoleArn      fwtypes.ARN  `tfsdk:"role_arn"`
}

// Helper functions to expand the logging config and tags
func expandLoggingConfig(logConfig fwtypes.ObjectValueOf[loggingConfigModel]) *awstypes.LoggingConfig {
	var lc loggingConfigModel
	logConfig.As(&lc)

	awslc := &awstypes.LoggingConfig{
		EmbeddingDataDeliveryEnabled: lc.EmbeddingDataDeliveryEnabled.ValueBool(),
	}

	if !lc.CloudWatchConfig.IsNull() {
		var cwc cloudWatchConfigModel
		lc.CloudWatchConfig.As(&cwc)
		awslc.CloudWatchConfig = &awstypes.CloudWatchConfig{
			LogGroupName: cwc.LogGroupName.ValueString(),
			RoleArn:      cwc.RoleArn.ValueString(),
		}
	}

	return awslc
}

func expandTags(tags types.Map) map[string]string {
	if tags.IsNull() {
		return nil
	}
	tagMap := make(map[string]string)
	for k, v := range tags.Elements() {
		if vStr, ok := v.(types.String); ok {
			tagMap[k] = vStr.ValueString()
		}
	}
	return tagMap
}
