// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Data Source")
func newDataSourceResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &dataSourceResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDataSource = "Data Source"
)

type dataSourceResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *dataSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrockagent_data_source"
}

func (r *dataSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"created_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"data_source_id": schema.StringAttribute{
				Computed: true,
			},
			"data_deletion_policy": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Optional: true,
			},
			"failure_reasons": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"id": framework.IDAttribute(),
			"knowledge_base_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"server_side_encryption_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[serverSideEncryptionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"kms_key_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
			"data_source_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceConfigurationModel](ctx),
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
						"s3_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3ConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bucket_arn": schema.StringAttribute{
										CustomType: fwtypes.ARNType,
										Required:   true,
									},
									"bucket_owner_account_id": schema.StringAttribute{
										Optional: true,
									},
									"inclusion_prefixes": schema.SetAttribute{
										CustomType:  fwtypes.SetOfStringType,
										ElementType: types.StringType,
										Optional:    true,
									},
								},
							},
						},
					},
				},
			},
			"vector_ingestion_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vectorIngestionConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Blocks: map[string]schema.Block{
						"chunking_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[chunkingConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"chunking_strategy": schema.StringAttribute{
										Required: true,
									},
								},
								Blocks: map[string]schema.Block{
									"fixed_size_chunking_configuration": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[fixedSizeChunkingConfigurationModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"max_tokens": schema.Int64Attribute{
													Required: true,
												},
												"overlap_percentage": schema.Int64Attribute{
													Required: true,
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

func (r *dataSourceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data dataSourceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.CreateDataSourceInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}
	input.ClientToken = aws.String(id.UniqueId())

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateDataSource(ctx, input)
	}, errCodeValidationException, "cannot assume role")

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Data Source", err.Error())

		return
	}

	ds := outputRaw.(*bedrockagent.CreateDataSourceOutput).DataSource
	data.DataSourceID = fwflex.StringToFramework(ctx, ds.DataSourceId)
	data.KnowledgeBaseID = fwflex.StringToFramework(ctx, ds.KnowledgeBaseId)
	data.setID()

	parts, err := flex.ExpandResourceId(data.ID.ValueString(), dataSourceResourceIdPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("Creating Bedrock Agent Data Source", err.Error())

		return
	}

	ds, err = waitDataSourceCreated(ctx, conn, parts[0], parts[1], r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Data Source (%s) create", data.DataSourceID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns after creation is complete.
	data.CreatedAt = fwflex.TimeToFramework(ctx, ds.CreatedAt)
	data.FailureReasons = fwflex.FlattenFrameworkStringValueListOfString(ctx, ds.FailureReasons)
	data.UpdatedAt = fwflex.TimeToFramework(ctx, ds.UpdatedAt)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *dataSourceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data dataSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	parts, err := flex.ExpandResourceId(data.ID.ValueString(), dataSourceResourceIdPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("Reading Bedrock Agent Data Source", err.Error())

		return
	}

	ds, err := findDataSourceByID(ctx, conn, parts[0], parts[1])

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Data Source (%s)", data.DataSourceID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, ds, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *dataSourceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new dataSourceResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !new.Description.Equal(old.Description) ||
		!new.DataDeletionPolicy.Equal(old.DataDeletionPolicy) ||
		!new.Name.Equal(old.Name) ||
		!new.DataSourceConfiguration.Equal(old.DataSourceConfiguration) ||
		!new.DataDeletionPolicy.Equal(old.DataDeletionPolicy) ||
		!new.ServerSideEncryptionConfiguration.Equal(old.ServerSideEncryptionConfiguration) ||
		!new.VectorIngestionConfiguration.Equal(old.VectorIngestionConfiguration) {
		input := &bedrockagent.UpdateDataSourceInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
			return conn.UpdateDataSource(ctx, input)
		}, errCodeValidationException, "cannot assume role")

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Agent Data Source (%s)", new.DataSourceID.ValueString()), err.Error())

			return
		}

		parts, err := flex.ExpandResourceId(new.ID.ValueString(), dataSourceResourceIdPartCount, false)
		if err != nil {
			response.Diagnostics.AddError("Updating Bedrock Agent Data Source", err.Error())

			return
		}

		ds, err := waitDataSourceUpdated(ctx, conn, parts[0], parts[1], r.UpdateTimeout(ctx, new.Timeouts))

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Data Source (%s) update", new.DataSourceID.ValueString()), err.Error())

			return
		}

		new.FailureReasons = fwflex.FlattenFrameworkStringValueListOfString(ctx, ds.FailureReasons)
		new.UpdatedAt = fwflex.TimeToFramework(ctx, ds.UpdatedAt)
	} else {
		new.FailureReasons = old.FailureReasons
		new.UpdatedAt = old.UpdatedAt
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *dataSourceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data dataSourceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	_, err := conn.DeleteDataSource(ctx, &bedrockagent.DeleteDataSourceInput{
		KnowledgeBaseId: aws.String(data.KnowledgeBaseID.ValueString()),
		DataSourceId:    aws.String(data.DataSourceID.ValueString()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Data Source (%s)", data.DataSourceID.ValueString()), err.Error())

		return
	}

	parts, err := flex.ExpandResourceId(data.ID.ValueString(), dataSourceResourceIdPartCount, false)
	if err != nil {
		response.Diagnostics.AddError("Deleting Bedrock Agent Data Source", err.Error())

		return
	}

	_, err = waitDataSourceDeleted(ctx, conn, parts[0], parts[1], r.DeleteTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Bedrock Agent Knowledge Base (%s) delete", data.KnowledgeBaseID.ValueString()), err.Error())

		return
	}
}

func (r *dataSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitDataSourceCreated(ctx context.Context, conn *bedrockagent.Client, id, kbId string, timeout time.Duration) (*awstypes.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.DataSourceStatusAvailable),
		Refresh:                   statusDataSource(ctx, conn, id, kbId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DataSource); ok {
		return out, err
	}

	return nil, err
}

func waitDataSourceUpdated(ctx context.Context, conn *bedrockagent.Client, id, kbId string, timeout time.Duration) (*awstypes.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DataSourceStatusAvailable),
		Target:                    enum.Slice(awstypes.DataSourceStatusAvailable),
		Refresh:                   statusDataSource(ctx, conn, id, kbId),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DataSource); ok {
		return out, err
	}

	return nil, err
}

func waitDataSourceDeleted(ctx context.Context, conn *bedrockagent.Client, id, kbId string, timeout time.Duration) (*awstypes.DataSource, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DataSourceStatusDeleting),
		Target:  []string{},
		Refresh: statusDataSource(ctx, conn, id, kbId),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DataSource); ok {
		return out, err
	}

	return nil, err
}

func statusDataSource(ctx context.Context, conn *bedrockagent.Client, id, kbId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDataSourceByID(ctx, conn, id, kbId)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findDataSourceByID(ctx context.Context, conn *bedrockagent.Client, id, kbId string) (*awstypes.DataSource, error) {
	input := &bedrockagent.GetDataSourceInput{
		DataSourceId:    aws.String(id),
		KnowledgeBaseId: aws.String(kbId),
	}

	out, err := conn.GetDataSource(ctx, input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if out == nil || out.DataSource == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out.DataSource, nil
}

type dataSourceResourceModel struct {
	CreatedAt                         timetypes.RFC3339                                                       `tfsdk:"created_at"`
	DataDeletionPolicy                types.String                                                            `tfsdk:"data_deletion_policy"`
	DataSourceConfiguration           fwtypes.ListNestedObjectValueOf[dataSourceConfigurationModel]           `tfsdk:"data_source_configuration"`
	DataSourceID                      types.String                                                            `tfsdk:"data_source_id"`
	ID                                types.String                                                            `tfsdk:"id"`
	Description                       types.String                                                            `tfsdk:"description"`
	FailureReasons                    fwtypes.ListValueOf[types.String]                                       `tfsdk:"failure_reasons"`
	KnowledgeBaseID                   types.String                                                            `tfsdk:"knowledge_base_id"`
	Name                              types.String                                                            `tfsdk:"name"`
	ServerSideEncryptionConfiguration fwtypes.ListNestedObjectValueOf[serverSideEncryptionConfigurationModel] `tfsdk:"server_side_encryption_configuration"`
	VectorIngestionConfiguration      fwtypes.ListNestedObjectValueOf[vectorIngestionConfigurationModel]      `tfsdk:"vector_ingestion_configuration"`
	Timeouts                          timeouts.Value                                                          `tfsdk:"timeouts"`
	UpdatedAt                         timetypes.RFC3339                                                       `tfsdk:"updated_at"`
}

const (
	dataSourceResourceIdPartCount = 2
)

func (data *dataSourceResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.DataSourceID.ValueString(), data.KnowledgeBaseID.ValueString()}, dataSourceResourceIdPartCount, false)))
}

type dataSourceConfigurationModel struct {
	Type            types.String                                          `tfsdk:"type"`
	S3Configuration fwtypes.ListNestedObjectValueOf[s3ConfigurationModel] `tfsdk:"s3_configuration"`
}

type s3ConfigurationModel struct {
	BucketARN            fwtypes.ARN                      `tfsdk:"bucket_arn"`
	BucketOwnerAccountId types.String                     `tfsdk:"bucket_owner_account_id"`
	InclusionPrefixes    fwtypes.SetValueOf[types.String] `tfsdk:"inclusion_prefixes"`
}

type serverSideEncryptionConfigurationModel struct {
	KmsKeyArn types.String `tfsdk:"kms_key_arn"`
}

type vectorIngestionConfigurationModel struct {
	ChunkingConfiguration fwtypes.ListNestedObjectValueOf[chunkingConfigurationModel] `tfsdk:"chunking_configuration"`
}

type chunkingConfigurationModel struct {
	ChunkingStrategy               types.String                                                         `tfsdk:"chunking_strategy"`
	FixedSizeChunkingConfiguration fwtypes.ListNestedObjectValueOf[fixedSizeChunkingConfigurationModel] `tfsdk:"fixed_size_chunking_configuration"`
}

type fixedSizeChunkingConfigurationModel struct {
	MaxTokens         types.Int64 `tfsdk:"max_tokens"`
	OverlapPercentage types.Int64 `tfsdk:"overlap_percentage"`
}
