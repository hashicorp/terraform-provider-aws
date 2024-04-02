// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Knowledge Base")
// @Tags(identifierAttribute="id")
func newKnowledgeBaseResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &knowledgeBaseResource{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameKnowledgeBase = "Knowledge Base"
)

type knowledgeBaseResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *knowledgeBaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrockagent_knowledge_base"
}

func (r *knowledgeBaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": framework.ARNAttributeComputedOnly(),
			"description": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role_name_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"knowledge_base_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[knowledgeBaseConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"vector_knowledge_base_configuration": schema.ListAttribute{
							CustomType: fwtypes.NewListNestedObjectTypeOf[vectorKnowledgeBaseConfigurationModel](ctx),
							Optional:   true,
							Computed:   true,
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							ElementType: types.ObjectType{
								AttrTypes: map[string]attr.Type{
									"embedding_model_arn": fwtypes.ARNType,
								},
							},
						},
						"type": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			"storage_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[storageConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
						},
					},
					Blocks: map[string]schema.Block{
						"opensearch_serverless_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[opensearchServerlessConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"collection_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[opensearchServerlessFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Optional: true,
												},
												"text_field": schema.StringAttribute{
													Optional: true,
												},
												"vector_field": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"pinecone_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[pineconeConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"connection_string": schema.StringAttribute{
										Required: true,
									},
									"credentials_secret_arn": schema.StringAttribute{
										Required:   true,
										CustomType: fwtypes.ARNType,
									},
									"namespace": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[pineconeFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Optional: true,
												},
												"text_field": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"rds_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[rdsConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_secret_arn": schema.StringAttribute{
										Required: true,
									},
									"database_name": schema.StringAttribute{
										Required: true,
									},
									"resource_arn": schema.StringAttribute{
										Required: true,
									},
									"table_name": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[rdsFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Optional: true,
												},
												"text_field": schema.StringAttribute{
													Optional: true,
												},
												"vector_field": schema.StringAttribute{
													Optional: true,
												},
												"primary_key_field": schema.StringAttribute{
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"redis_enterprise_cloud_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[redisEnterpriseCloudConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"credentials_secret_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"endpoint": schema.StringAttribute{
										Required: true,
									},
									"vector_index_name": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"field_mapping": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[redisEnterpriseCloudFieldMappingModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"metadata_field": schema.StringAttribute{
													Optional: true,
												},
												"text_field": schema.StringAttribute{
													Optional: true,
												},
												"vector_field": schema.StringAttribute{
													Optional: true,
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
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *knowledgeBaseResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data knowledgeBaseResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &bedrockagent.CreateKnowledgeBaseInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.CreateKnowledgeBase(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameKnowledgeBase, data.Name.String(), err),
			err.Error(),
		)
		return
	}
	if output == nil || output.KnowledgeBase == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameKnowledgeBase, data.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	knowledgebase := output.KnowledgeBase
	data.KnowledgeBaseARN = flex.StringToFramework(ctx, output.KnowledgeBase.KnowledgeBaseArn)
	data.KnowledgeBaseId = flex.StringToFramework(ctx, output.KnowledgeBase.KnowledgeBaseId)

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	knowledgebase, err = waitKnowledgeBaseCreated(ctx, conn, data.KnowledgeBaseId.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForCreation, ResNameKnowledgeBase, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	var knowledgeBaseConfiguration knowledgeBaseConfigurationModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, knowledgebase.KnowledgeBaseConfiguration, &knowledgeBaseConfiguration)...)
	if response.Diagnostics.HasError() {
		return
	}

	var storageConfiguration storageConfigurationModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, knowledgebase.StorageConfiguration, &storageConfiguration)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Set values for unknowns after creation is complete.
	data.KnowledgeBaseConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &knowledgeBaseConfiguration)
	data.StorageConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &storageConfiguration)
	data.Name = fwflex.StringToFramework(ctx, knowledgebase.Name)
	data.Description = fwflex.StringToFramework(ctx, knowledgebase.Description)
	data.RoleARN = fwflex.StringToFramework(ctx, knowledgebase.RoleArn)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *knowledgeBaseResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data knowledgeBaseResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findKnowledgeBaseByID(ctx, conn, data.KnowledgeBaseId.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionSetting, ResNameKnowledgeBase, data.KnowledgeBaseId.String(), err),
			err.Error(),
		)
		return
	}

	var knowledgeBaseConfiguration knowledgeBaseConfigurationModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.KnowledgeBaseConfiguration, &knowledgeBaseConfiguration)...)
	if response.Diagnostics.HasError() {
		return
	}

	var storageConfiguration storageConfigurationModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.StorageConfiguration, &storageConfiguration)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.KnowledgeBaseConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &knowledgeBaseConfiguration)
	data.StorageConfiguration = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &storageConfiguration)
	data.Name = fwflex.StringToFramework(ctx, output.Name)
	data.Description = fwflex.StringToFramework(ctx, output.Description)
	data.RoleARN = fwflex.StringToFramework(ctx, output.RoleArn)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *knowledgeBaseResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var old, new knowledgeBaseResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !new.Name.Equal(old.Name) ||
		!new.Description.Equal(old.Description) ||
		!new.KnowledgeBaseConfiguration.Equal(old.KnowledgeBaseConfiguration) ||
		!new.StorageConfiguration.Equal(old.StorageConfiguration) {

		input := &bedrockagent.UpdateKnowledgeBaseInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateKnowledgeBase(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameKnowledgeBase, new.KnowledgeBaseId.String(), err),
				err.Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, new.Timeouts)
	_, err := waitKnowledgeBaseUpdated(ctx, conn, new.KnowledgeBaseId.ValueString(), updateTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForUpdate, ResNameKnowledgeBase, new.KnowledgeBaseId.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *knowledgeBaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data knowledgeBaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &bedrockagent.DeleteKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(data.KnowledgeBaseId.ValueString()),
	}

	_, err := conn.DeleteKnowledgeBase(ctx, in)

	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionDeleting, ResNameKnowledgeBase, data.KnowledgeBaseId.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, data.Timeouts)
	_, err = waitKnowledgeBaseDeleted(ctx, conn, data.KnowledgeBaseId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForDeletion, ResNameKnowledgeBase, data.KnowledgeBaseId.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *knowledgeBaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitKnowledgeBaseCreated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.KnowledgeBaseStatusCreating),
		Target:                    enum.Slice(awstypes.KnowledgeBaseStatusActive),
		Refresh:                   statusKnowledgeBase(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		return out, err
	}

	return nil, err
}

func waitKnowledgeBaseUpdated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.KnowledgeBaseStatusActive, awstypes.KnowledgeBaseStatusUpdating),
		Target:                    enum.Slice(awstypes.KnowledgeBaseStatusActive),
		Refresh:                   statusKnowledgeBase(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		return out, err
	}

	return nil, err
}

func waitKnowledgeBaseDeleted(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*awstypes.KnowledgeBase, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KnowledgeBaseStatusActive, awstypes.KnowledgeBaseStatusDeleting),
		Target:  []string{},
		Refresh: statusKnowledgeBase(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.KnowledgeBase); ok {
		return out, err
	}

	return nil, err
}

func statusKnowledgeBase(ctx context.Context, conn *bedrockagent.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findKnowledgeBaseByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findKnowledgeBaseByID(ctx context.Context, conn *bedrockagent.Client, id string) (*awstypes.KnowledgeBase, error) {
	in := &bedrockagent.GetKnowledgeBaseInput{
		KnowledgeBaseId: aws.String(id),
	}

	out, err := conn.GetKnowledgeBase(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.KnowledgeBase == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.KnowledgeBase, nil
}

type knowledgeBaseResourceModel struct {
	Description                types.String                                                     `tfsdk:"description"`
	KnowledgeBaseConfiguration fwtypes.ListNestedObjectValueOf[knowledgeBaseConfigurationModel] `tfsdk:"knowledge_base_configuration"`
	Name                       types.String                                                     `tfsdk:"name"`
	RoleARN                    types.String                                                     `tfsdk:"role_arn"`
	StorageConfiguration       fwtypes.ListNestedObjectValueOf[storageConfigurationModel]       `tfsdk:"storage_configuration"`
	KnowledgeBaseARN           types.String                                                     `tfsdk:"arn"`
	KnowledgeBaseId            types.String                                                     `tfsdk:"id"`
	RoleNameArn                types.String                                                     `tfsdk:"role_name_arn"`
	Timeouts                   timeouts.Value                                                   `tfsdk:"timeouts"`
}

type knowledgeBaseConfigurationModel struct {
	Type                             types.String                                                           `tfsdk:"type"`
	VectorKnowledgeBaseConfiguration fwtypes.ListNestedObjectValueOf[vectorKnowledgeBaseConfigurationModel] `tfsdk:"vector_knowledge_base_configuration"`
}

type vectorKnowledgeBaseConfigurationModel struct {
	EmbeddingModelARN types.String `tfsdk:"embedding_model_arn"`
}

type storageConfigurationModel struct {
	OpensearchServerlessConfiguration fwtypes.ListNestedObjectValueOf[opensearchServerlessConfigurationModel] `tfsdk:"opensearch_serverless_configuration"`
	PineconeConfiguration             fwtypes.ListNestedObjectValueOf[pineconeConfigurationModel]             `tfsdk:"pinecone_configuration"`
	RdsConfiguration                  fwtypes.ListNestedObjectValueOf[rdsConfigurationModel]                  `tfsdk:"rds_configuration"`
	RedisEnterpriseCloudConfiguration fwtypes.ListNestedObjectValueOf[redisEnterpriseCloudConfigurationModel] `tfsdk:"redis_enterprise_cloud_configuration"`
	Type                              types.String                                                            `tfsdk:"type"`
}

type opensearchServerlessConfigurationModel struct {
	CollectionArn   types.String                                                           `tfsdk:"collection_arn"`
	FieldMapping    fwtypes.ListNestedObjectValueOf[opensearchServerlessFieldMappingModel] `tfsdk:"field_mapping"`
	VectorIndexName types.String                                                           `tfsdk:"vector_index_name"`
}

type opensearchServerlessFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}

type pineconeConfigurationModel struct {
	ConnectionString     types.String                                               `tfsdk:"connection_string"`
	CredentialsSecretARN types.String                                               `tfsdk:"credentials_secret_arn"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[pineconeFieldMappingModel] `tfsdk:"field_mapping"`
	Namespace            types.String                                               `tfsdk:"namespace"`
}

type pineconeFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
}

type rdsConfigurationModel struct {
	CredentialsSecretArn types.String                                          `tfsdk:"credentials_secret_arn"`
	DatabaseName         types.String                                          `tfsdk:"database_name"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[rdsFieldMappingModel] `tfsdk:"field_mapping"`
	ResourceArn          types.String                                          `tfsdk:"resource_arn"`
	TableName            types.String                                          `tfsdk:"table_name"`
}

type rdsFieldMappingModel struct {
	MetadataField   types.String `tfsdk:"metadata_field"`
	TextField       types.String `tfsdk:"text_field"`
	VectorField     types.String `tfsdk:"vector_field"`
	PrimaryKeyField types.String `tfsdk:"primary_key_field"`
}

type redisEnterpriseCloudConfigurationModel struct {
	CredentialsSecretArn types.String                                                           `tfsdk:"credentials_secret_arn"`
	Endpoint             types.String                                                           `tfsdk:"endpoint"`
	FieldMapping         fwtypes.ListNestedObjectValueOf[redisEnterpriseCloudFieldMappingModel] `tfsdk:"field_mapping"`
	VectorIndexName      types.String                                                           `tfsdk:"vector_index_name"`
}

type redisEnterpriseCloudFieldMappingModel struct {
	MetadataField types.String `tfsdk:"metadata_field"`
	TextField     types.String `tfsdk:"text_field"`
	VectorField   types.String `tfsdk:"vector_field"`
}
