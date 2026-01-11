// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptunegraph"
	awstypes "github.com/aws/aws-sdk-go-v2/service/neptunegraph/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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

// @FrameworkResource("aws_neptunegraph_graph", name="Graph")
// @Tags(identifierAttribute="arn")
func newGraphResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &graphResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type graphResource struct {
	framework.ResourceWithModel[graphResourceModel]
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *graphResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDeletionProtection: schema.BoolAttribute{
				Description: "A value that indicates whether the graph has deletion protection enabled. The graph can't be deleted when deletion protection is enabled.",
				Computed:    true,
				Optional:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrEndpoint: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"graph_name": schema.StringAttribute{
				Description: `The graph name. For example: my-graph-1.
								The name must contain from 1 to 63 letters, numbers, or hyphens, 
								and its first character must be a letter. It cannot end with a hyphen or contain two consecutive hyphens.
								If you don't specify a graph name, a unique graph name is generated for you using the prefix graph-for, 
								followed by a combination of Stack Name and a UUID.`,
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("graph_name_prefix")),
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(
						regexache.MustCompile("^[a-zA-z][a-zA-Z0-9]*(-[a-zA-Z0-9]+)*$"), ""),
				},
			},
			"graph_name_prefix": schema.StringAttribute{
				Description: "Allows user to specify name prefix and have remainder of name automatically generated.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"kms_key_identifier": schema.StringAttribute{
				Description: "Specifies a KMS key to use to encrypt data in the new graph.  Value must be ARN of KMS Key.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"provisioned_memory": schema.Int32Attribute{
				Description: "The provisioned memory-optimized Neptune Capacity Units (m-NCUs) to use for the graph. Required when max_provisioned_memory and min_provisioned_memory are not specified.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.OneOf(8, 16, 32, 64, 128, 256, 384, 512, 768, 1024, 2048, 3072, 4096),
					int32validator.ConflictsWith(path.MatchRoot("import_task")),
				},
			},
			"public_connectivity": schema.BoolAttribute{
				Description: `Specifies whether or not the graph can be reachable over the internet. 
								All access to graphs is IAM authenticated.
								When the graph is publicly available, its domain name system (DNS) endpoint resolves to 
								the public IP address from the internet. When the graph isn't publicly available, you need 
								to create a PrivateGraphEndpoint in a given VPC to ensure the DNS name resolves to a private 
								IP address that is reachable from the VPC.`,
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"replica_count": schema.Int32Attribute{
				Description: "The number of replicas in other AZs.  Value must be between 0 and 2.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
					int32planmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.Int32{
					int32validator.Between(0, 2),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"import_task": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[importTaskModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "Configuration for importing data into the graph during creation. Forces replacement if changed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrSource: schema.StringAttribute{
							Description: "URL identifying the location of data to import (S3 path, Neptune endpoint, or snapshot).",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						names.AttrRoleARN: schema.StringAttribute{
							Description: "ARN of the IAM role that allows access to the data to be imported.",
							Required:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexache.MustCompile(`^arn:aws[^:]*:iam::\d{12}:(role|role/service-role)(/[\w+=,.@-]+)+$`),
									"must be a valid IAM role ARN",
								),
							},
						},
						names.AttrFormat: schema.StringAttribute{
							Description: `Specifies the format of S3 data to be imported. Valid values are CSV, PARQUET, OPEN_CYPHER, or NTRIPLES.`,
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf("CSV", "PARQUET", "OPEN_CYPHER", "NTRIPLES"),
							},
						},
						"fail_on_error": schema.BoolAttribute{
							Description: "If true, task halts on import error. If false, skips problem data and continues.",
							Optional:    true,
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.RequiresReplace(),
							},
						},
						"max_provisioned_memory": schema.Int32Attribute{
							Description: "Maximum m-NCUs for the import task. Default: 1024.",
							Optional:    true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int32{
								int32validator.OneOf(8, 16, 32, 64, 128, 256, 384, 512, 768, 1024, 2048, 3072, 4096),
							},
						},
						"min_provisioned_memory": schema.Int32Attribute{
							Description: "Minimum m-NCUs for the import task. Default: 16",
							Optional:    true,
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int32{
								int32validator.OneOf(8, 16, 32, 64, 128, 256, 384, 512, 768, 1024, 2048, 3072, 4096),
							},
						},
						"blank_node_handling": schema.StringAttribute{
							Description: `The method to handle blank nodes in the dataset. Currently, only convertToIri is supported.`,
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf("convertToIri"),
							},
						},
						"parquet_type": schema.StringAttribute{
							Description: "Parquet type for import processing.",
							Optional:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf("COLUMNAR"),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"import_options": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[importOptionsModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							Description: "Options for controlling the import process.",
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"neptune": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[neptuneImportOptionsModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										Description: "Options for importing data from a Neptune database.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"s3_export_path": schema.StringAttribute{
													Description: "The path to an S3 bucket from which to import data.",
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 1024),
													},
												},
												"s3_export_kms_key_id": schema.StringAttribute{
													Description: "The KMS key to use to encrypt data in the S3 bucket where the graph data is exported.",
													Optional:    true,
													Validators: []validator.String{
														stringvalidator.LengthBetween(1, 1024),
													},
												},
												"preserve_default_vertex_labels": schema.BoolAttribute{
													Description: "Whether to preserve default vertex labels.",
													Optional:    true,
												},
												"preserve_edge_ids": schema.BoolAttribute{
													Description: "Whether to preserve edge IDs as properties.",
													Optional:    true,
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
			"vector_search_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vectorSearchConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				Description: "Vector search configuration for the Neptune Graph",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"vector_search_dimension": schema.Int32Attribute{
							Optional:    true,
							Description: "Specifies the number of dimensions for vector embeddings.  Value must be between 1 and 65,535.",
							PlanModifiers: []planmodifier.Int32{
								int32planmodifier.RequiresReplace(),
							},
							Validators: []validator.Int32{
								int32validator.Between(1, 65535),
							},
						},
					},
				},
			},
		},
	}
}

func (r *graphResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data graphResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NeptuneGraphClient(ctx)

	// Generate graph name
	graphName := create.NewNameGenerator(
		create.WithConfiguredName(data.Name.ValueString()),
		create.WithConfiguredPrefix(data.NamePrefix.ValueString()),
		create.WithDefaultPrefix("tf-"),
	).Generate()

	// Determine whether to use CreateGraphUsingImportTask or CreateGraph API
	if !data.ImportTask.IsNull() && !data.ImportTask.IsUnknown() {
		var importInput neptunegraph.CreateGraphUsingImportTaskInput

		response.Diagnostics.Append(fwflex.Expand(ctx, data, &importInput)...)
		if response.Diagnostics.HasError() {
			return
		}

		var importTaskData []importTaskModel
		response.Diagnostics.Append(data.ImportTask.ElementsAs(ctx, &importTaskData, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		if len(importTaskData) > 0 {
			task := importTaskData[0]

			taskImportOptions := task.ImportOptions
			task.ImportOptions = fwtypes.NewListNestedObjectValueOfNull[importOptionsModel](ctx)

			response.Diagnostics.Append(fwflex.Expand(ctx, task, &importInput)...)
			if response.Diagnostics.HasError() {
				return
			}

			if !taskImportOptions.IsNull() && !taskImportOptions.IsUnknown() {
				var importOptionsDataSlice []importOptionsModel
				response.Diagnostics.Append(taskImportOptions.ElementsAs(ctx, &importOptionsDataSlice, false)...)
				if response.Diagnostics.HasError() {
					return
				}

				if len(importOptionsDataSlice) > 0 {
					importOptionsData := importOptionsDataSlice[0]

					if !importOptionsData.Neptune.IsNull() {
						var neptuneOptionsDataSlice []neptuneImportOptionsModel
						response.Diagnostics.Append(importOptionsData.Neptune.ElementsAs(ctx, &neptuneOptionsDataSlice, false)...)
						if response.Diagnostics.HasError() {
							return
						}
						if len(neptuneOptionsDataSlice) > 0 {
							neptuneOptionsData := neptuneOptionsDataSlice[0]

							neptuneOpts := &awstypes.NeptuneImportOptions{}
							response.Diagnostics.Append(fwflex.Expand(ctx, neptuneOptionsData, neptuneOpts)...)
							if response.Diagnostics.HasError() {
								return
							}
							importInput.ImportOptions = &awstypes.ImportOptionsMemberNeptune{
								Value: *neptuneOpts,
							}
						}
					}
				}
			}
		}

		importInput.GraphName = aws.String(graphName)
		importInput.Tags = getTagsIn(ctx)

		output, err := conn.CreateGraphUsingImportTask(ctx, &importInput)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("creating Neptune Graph with import task (%s)", graphName), err.Error())
			return
		}

		data.ID = fwflex.StringToFramework(ctx, output.GraphId)

		_, err = waitGraphImportTaskCompleted(ctx, conn, aws.ToString(output.TaskId), r.CreateTimeout(ctx, data.Timeouts))
		if err != nil {
			response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Neptune Graph import task (%s) completion", data.ID.ValueString()), err.Error())
			return
		}

		graph, err := findGraphByID(ctx, conn, data.ID.ValueString())
		if err != nil {
			response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
			response.Diagnostics.AddError(fmt.Sprintf("reading Neptune Graph (%s) after import", data.ID.ValueString()), err.Error())
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, graph, &data, fwflex.WithIgnoredFieldNames([]string{"ImportTask"}))...)
	} else {
		// Use existing CreateGraph API
		var input neptunegraph.CreateGraphInput
		response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		input.GraphName = aws.String(graphName)
		input.Tags = getTagsIn(ctx)

		output, err := conn.CreateGraph(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("creating Neptune Graph (%s)", graphName), err.Error())
			return
		}

		data.ID = fwflex.StringToFramework(ctx, output.Id)

		graph, err := waitGraphCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
		if err != nil {
			response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Neptune Graph (%s) create", data.ID.ValueString()), err.Error())
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, graph, &data)...)
	}

	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *graphResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data graphResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NeptuneGraphClient(ctx)

	output, err := findGraphByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Neptune Graph Graph (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *graphResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new graphResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NeptuneGraphClient(ctx)

	if !new.Name.Equal(old.Name) ||
		!new.NamePrefix.Equal(old.NamePrefix) ||
		!new.DeletionProtection.Equal(old.DeletionProtection) ||
		!new.ProvisionedMemory.Equal(old.ProvisionedMemory) ||
		!new.PublicConnectivity.Equal(old.PublicConnectivity) ||
		!new.ReplicaCount.Equal(old.ReplicaCount) {
		input := neptunegraph.UpdateGraphInput{
			GraphIdentifier: new.ID.ValueStringPointer(),
		}

		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateGraph(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Neptune Graph Graph (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if _, err := waitGraphUpdated(ctx, conn, old.ID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Neptune Graph Graph (%s) update", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *graphResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data graphResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().NeptuneGraphClient(ctx)

	input := neptunegraph.DeleteGraphInput{
		GraphIdentifier: data.ID.ValueStringPointer(),
		SkipSnapshot:    aws.Bool(true),
	}

	_, err := conn.DeleteGraph(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Neptune Graph Graph (%s)", data.ID.ValueString()), err.Error())
		return
	}

	if _, err := waitGraphDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Neptune Graph Graph (%s) delete", data.ID.ValueString()), err.Error())
		return
	}
}

func findGraphByID(ctx context.Context, conn *neptunegraph.Client, id string) (*neptunegraph.GetGraphOutput, error) {
	input := neptunegraph.GetGraphInput{
		GraphIdentifier: aws.String(id),
	}

	return findGraph(ctx, conn, &input)
}

func findGraph(ctx context.Context, conn *neptunegraph.Client, input *neptunegraph.GetGraphInput) (*neptunegraph.GetGraphOutput, error) {
	output, err := conn.GetGraph(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
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

func statusGraph(conn *neptunegraph.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGraphByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitGraphCreated(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.GetGraphOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GraphStatusCreating),
		Target:                    enum.Slice(awstypes.GraphStatusAvailable),
		Refresh:                   statusGraph(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetGraphOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		return output, err
	}

	return nil, err
}

func findImportTaskByID(ctx context.Context, conn *neptunegraph.Client, taskID string) (*neptunegraph.GetImportTaskOutput, error) {
	input := neptunegraph.GetImportTaskInput{
		TaskIdentifier: aws.String(taskID),
	}

	output, err := conn.GetImportTask(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
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

func statusImportTask(conn *neptunegraph.Client, taskID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findImportTaskByID(ctx, conn, taskID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitGraphImportTaskCompleted(ctx context.Context, conn *neptunegraph.Client, taskID string, timeout time.Duration) (*neptunegraph.GetImportTaskOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImportTaskStatusInitializing, awstypes.ImportTaskStatusImporting, awstypes.ImportTaskStatusExporting,
			awstypes.ImportTaskStatusAnalyzingData, awstypes.ImportTaskStatusReprovisioning),
		Target:  enum.Slice(awstypes.ImportTaskStatusSucceeded),
		Refresh: statusImportTask(conn, taskID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetImportTaskOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))
		return output, err
	}

	return nil, err
}

func waitGraphUpdated(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.GetGraphOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GraphStatusUpdating),
		Target:                    enum.Slice(awstypes.GraphStatusAvailable),
		Refresh:                   statusGraph(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetGraphOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitGraphDeleted(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.GetGraphOutput, error) {
	const (
		delay = 10 * time.Second
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.GraphStatusDeleting),
		Target:  []string{},
		Refresh: statusGraph(conn, id),
		Delay:   delay,
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetGraphOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

type graphResourceModel struct {
	framework.WithRegionModel
	ARN                       types.String                                                    `tfsdk:"arn"`
	DeletionProtection        types.Bool                                                      `tfsdk:"deletion_protection"`
	Endpoint                  types.String                                                    `tfsdk:"endpoint"`
	ID                        types.String                                                    `tfsdk:"id"`
	Name                      types.String                                                    `tfsdk:"graph_name"`
	NamePrefix                types.String                                                    `tfsdk:"graph_name_prefix"`
	KMSKeyIdentifier          types.String                                                    `tfsdk:"kms_key_identifier"`
	ProvisionedMemory         types.Int32                                                     `tfsdk:"provisioned_memory"`
	PublicConnectivity        types.Bool                                                      `tfsdk:"public_connectivity"`
	ReplicaCount              types.Int32                                                     `tfsdk:"replica_count"`
	Tags                      tftags.Map                                                      `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                      `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                                  `tfsdk:"timeouts"`
	VectorSearchConfiguration fwtypes.ListNestedObjectValueOf[vectorSearchConfigurationModel] `tfsdk:"vector_search_configuration"`
	ImportTask                fwtypes.ListNestedObjectValueOf[importTaskModel]                `tfsdk:"import_task"`
}

type importTaskModel struct {
	Source               types.String                                        `tfsdk:"source"`
	RoleArn              types.String                                        `tfsdk:"role_arn"`
	Format               types.String                                        `tfsdk:"format"`
	FailOnError          types.Bool                                          `tfsdk:"fail_on_error"`
	MaxProvisionedMemory types.Int32                                         `tfsdk:"max_provisioned_memory"`
	MinProvisionedMemory types.Int32                                         `tfsdk:"min_provisioned_memory"`
	BlankNodeHandling    types.String                                        `tfsdk:"blank_node_handling"`
	ParquetType          types.String                                        `tfsdk:"parquet_type"`
	ImportOptions        fwtypes.ListNestedObjectValueOf[importOptionsModel] `tfsdk:"import_options"`
}

type importOptionsModel struct {
	Neptune fwtypes.ListNestedObjectValueOf[neptuneImportOptionsModel] `tfsdk:"neptune"`
}

type neptuneImportOptionsModel struct {
	S3ExportPath                types.String `tfsdk:"s3_export_path"`
	S3ExportKmsKeyId            types.String `tfsdk:"s3_export_kms_key_id"`
	PreserveDefaultVertexLabels types.Bool   `tfsdk:"preserve_default_vertex_labels"`
	PreserveEdgeIds             types.Bool   `tfsdk:"preserve_edge_ids"`
}

type vectorSearchConfigurationModel struct {
	Dimension types.Int32 `tfsdk:"vector_search_dimension"`
}
