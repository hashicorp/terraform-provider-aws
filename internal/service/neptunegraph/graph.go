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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
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
	framework.ResourceWithConfigure
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
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
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

	var input neptunegraph.CreateGraphInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	// NeptuneGraph Sdk GetGraphOutput param for Name differs from CreateGraphInput param as GraphName
	input.GraphName = aws.String(
		create.NewNameGenerator(
			create.WithConfiguredName(data.Name.ValueString()),
			create.WithConfiguredPrefix(data.NamePrefix.ValueString()),
			create.WithDefaultPrefix("tf-"),
		).Generate(),
	)
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateGraph(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Neptune Graph Graph (%s)", aws.ToString(input.GraphName)), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.Id)

	graph, err := waitGraphCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Neptune Graph Graph (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, graph, &data)...)
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

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Neptune Graph Graph (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
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

	//SkipSnapshot is hardcoded here as this is the same behavior currently supported
	//in AWS CloudFormation.
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

func statusGraph(ctx context.Context, conn *neptunegraph.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
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
		Refresh:                   statusGraph(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetGraphOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

func waitGraphUpdated(ctx context.Context, conn *neptunegraph.Client, id string, timeout time.Duration) (*neptunegraph.GetGraphOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.GraphStatusUpdating),
		Target:                    enum.Slice(awstypes.GraphStatusAvailable),
		Refresh:                   statusGraph(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetGraphOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

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
		Refresh: statusGraph(ctx, conn, id),
		Delay:   delay,
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*neptunegraph.GetGraphOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.StatusReason)))

		return output, err
	}

	return nil, err
}

type graphResourceModel struct {
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
}

type vectorSearchConfigurationModel struct {
	Dimension types.Int32 `tfsdk:"vector_search_dimension"`
}
