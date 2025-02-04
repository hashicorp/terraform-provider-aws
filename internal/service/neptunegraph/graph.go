// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptunegraph

import (
	"context"
	"errors"
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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
	// Imports needed if protocol version 6 becomes supported - ability to move
	// vector_search_configuration into attributes vs. using blocks:
	//
	// "github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
)

// @FrameworkResource("aws_neptunegraph_graph", name="Graph")
// @Tags(identifierAttribute="arn")
func newResourceGraph(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceGraph{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameGraph = "Graph"
)

type resourceGraph struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceGraph) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_neptunegraph_graph"
}

func (r *resourceGraph) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
			names.AttrID: schema.StringAttribute{
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
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"kms_key_identifier": schema.StringAttribute{
				Description: "Specifies a KMS key to use to encrypt data in the new graph.  Value must be ARN of KMS Key.",
				Optional:    true,
				Computed:    true, //value will default to AWS_OWNED_KEY if no KMS key ARN is specified
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"provisioned_memory": schema.Int32Attribute{
				Description: "The provisioned memory-optimized Neptune Capacity Units (m-NCUs) to use for the graph.",
				Required:    true,
				Validators: []validator.Int32{
					int32validator.OneOf(8, 16, 32, 64, 128, 256, 384, 512, 768, 1024, 2048, 3072, 4096),
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
			// Preferred configuration for vector_search_configuration, though currently
			// not supported in protocol version 5.  Leaving here for when (if ever?)
			// protocol version 6 becomes supported:
			//
			// "vector_search_configuration": schema.SingleNestedAttribute{
			// 	Attributes: map[string]schema.Attribute{
			// 		"vector_search_dimension": schema.Int32Attribute{
			// 			Description: "Specifies the number of dimensions for vector embeddings.  Value must be between 1 and 65,535.",
			// 			Optional:    true,
			// 			Computed:    true,
			// 			Validators: []validator.Int32{
			// 				int32validator.Between(1, 65535),
			// 			},
			// 			PlanModifiers: []planmodifier.Int32{
			// 				int32planmodifier.UseStateForUnknown(),
			// 			},
			// 		},
			// 	},
			// 	Description: "Vector search configuration for the Neptune Graph",
			// 	Optional:    true,
			// 	Computed:    true,
			// 	PlanModifiers: []planmodifier.Object{ /*START PLAN MODIFIERS*/
			// 		objectplanmodifier.UseStateForUnknown(),
			// 		objectplanmodifier.RequiresReplaceIfConfigured(),
			// 	},
			// },
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"vector_search_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[vectorSearchConfiguration](ctx),
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

func (r *resourceGraph) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {

	conn := r.Meta().NeptuneGraphClient(ctx)

	var plan resourceGraphModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input neptunegraph.CreateGraphInput
	input.Tags = getTagsIn(ctx)

	// NeptuneGraph Sdk GetGraphOutput param for Name differs from CreateGraphInput param as GraphName
	input.GraphName = aws.String(
		create.NewNameGenerator(
			create.WithConfiguredName(plan.Name.ValueString()),
			create.WithConfiguredPrefix(plan.NamePrefix.ValueString()),
			create.WithDefaultPrefix("tf-"),
		).Generate(),
	)

	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateGraph(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionCreating, ResNameGraph, plan.Name.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Id == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionCreating, ResNameGraph, plan.Name.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitGraphCreated(ctx, conn, *out.Id, createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionWaitingForCreation, ResNameGraph, plan.Name.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceGraph) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

	conn := r.Meta().NeptuneGraphClient(ctx)

	var state resourceGraphModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findGraphByID(ctx, conn, state.Id.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionSetting, ResNameGraph, state.Id.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceGraph) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	conn := r.Meta().NeptuneGraphClient(ctx)

	var plan, state resourceGraphModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) ||
		!plan.NamePrefix.Equal(state.NamePrefix) ||
		!plan.DeletionProtection.Equal(state.DeletionProtection) ||
		!plan.ProvisionedMemory.Equal(state.ProvisionedMemory) ||
		!plan.PublicConnectivity.Equal(state.PublicConnectivity) ||
		!plan.ReplicaCount.Equal(state.ReplicaCount) {

		input := neptunegraph.UpdateGraphInput{
			GraphIdentifier: state.Id.ValueStringPointer(),
		}

		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateGraph(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionUpdating, ResNameGraph, plan.Id.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Id == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionUpdating, ResNameGraph, plan.Id.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		output, err := waitGraphUpdated(ctx, conn, state.Id.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionWaitingForUpdate, ResNameGraph, plan.Id.String(), err),
				err.Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}

	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceGraph) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	conn := r.Meta().NeptuneGraphClient(ctx)

	var state resourceGraphModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	//SkipSnapshot is hardcoded here as this is the same behavior currently supported
	//in AWS CloudFormation.
	input := neptunegraph.DeleteGraphInput{
		GraphIdentifier: state.Id.ValueStringPointer(),
		SkipSnapshot:    aws.Bool(true),
	}

	_, err := conn.DeleteGraph(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionDeleting, ResNameGraph, state.Id.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitGraphDeleted(ctx, conn, state.Id.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.NeptuneGraph, create.ErrActionWaitingForDeletion, ResNameGraph, state.Id.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceGraph) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceGraph) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

const (
	statusChangePending = "UPDATING"
	statusDeleting      = "DELETING"
	statusAvailable     = "AVAILABLE"
)

func waitGraphCreated(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.CreateGraphOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusAvailable},
		Refresh:                   statusGraph(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*neptunegraph.CreateGraphOutput); ok {
		return out, err
	}

	return nil, err
}

func waitGraphUpdated(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.GetGraphOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusAvailable},
		Refresh:                   statusGraph(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*neptunegraph.GetGraphOutput); ok {
		return out, err
	}

	return nil, err
}

func waitGraphDeleted(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.DeleteGraphOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusAvailable},
		Target:  []string{},
		Refresh: statusGraph(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*neptunegraph.DeleteGraphOutput); ok {
		return out, err
	}

	return nil, err
}

func statusGraph(ctx context.Context, conn *neptunegraph.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findGraphByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findGraphByID(ctx context.Context, conn *neptunegraph.Client, id string) (*neptunegraph.GetGraphOutput, error) {
	in := &neptunegraph.GetGraphInput{
		GraphIdentifier: aws.String(id),
	}

	out, err := conn.GetGraph(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.Id == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceGraphModel struct {
	Arn                       types.String                                               `tfsdk:"arn"`
	DeletionProtection        types.Bool                                                 `tfsdk:"deletion_protection"`
	Endpoint                  types.String                                               `tfsdk:"endpoint"`
	Id                        types.String                                               `tfsdk:"id"`
	Name                      types.String                                               `tfsdk:"graph_name"`
	NamePrefix                types.String                                               `tfsdk:"graph_name_prefix"`
	KmsKeyIdentifier          types.String                                               `tfsdk:"kms_key_identifier"`
	ProvisionedMemory         types.Int32                                                `tfsdk:"provisioned_memory"`
	PublicConnectivity        types.Bool                                                 `tfsdk:"public_connectivity"`
	ReplicaCount              types.Int32                                                `tfsdk:"replica_count"`
	Tags                      tftags.Map                                                 `tfsdk:"tags"`
	TagsAll                   tftags.Map                                                 `tfsdk:"tags_all"`
	Timeouts                  timeouts.Value                                             `tfsdk:"timeouts"`
	VectorSearchConfiguration fwtypes.ListNestedObjectValueOf[vectorSearchConfiguration] `tfsdk:"vector_search_configuration"`

	// Change required if protocol version 6 becomes supported.  Allows
	// vector_search_configuration to be a nested attribute vs. a block.
	//
	//VectorSearchConfiguration *vectorSearchConfiguration `tfsdk:"vector_search_configuration"`
}

type vectorSearchConfiguration struct {
	Dimension types.Int32 `tfsdk:"vector_search_dimension"`
}
