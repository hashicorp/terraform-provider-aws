// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Bedrock Agent Alias")
// @Tags(identifierAttribute="agent_alias_arn")
func newAgentAliasResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &agentAliasResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type agentAliasResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *agentAliasResource) Metadata(_ context.Context, request resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrock_agent_alias"
}

func (r *agentAliasResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_alias_arn": framework.ARNAttributeComputedOnly(),
			"agent_alias_id": schema.StringAttribute{
				Computed: true,
			},
			"agent_alias_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_alias_status": schema.StringAttribute{
				Computed: true,
			},
			"agent_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"description": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"routing_configuration": schema.ListAttribute{
				Optional:   true,
				Computed:   true,
				Validators: []validator.List{listvalidator.SizeAtMost(1)},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				CustomType: fwtypes.NewListNestedObjectTypeOf[roc](ctx),
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"agent_version": types.StringType,
					},
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"agent_alias_history_events": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[history](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"end_date": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"start_date": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"routing_configuration": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[roc](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"agent_version": schema.StringAttribute{
										Computed: true,
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

func (r *agentAliasResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data bedrockAliasResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.CreateAgentAliasInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(id.UniqueId())
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAgentAlias(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Agent Alias", err.Error())

		return
	}

	idParts := []string{
		*output.AgentAlias.AgentAliasId,
		*output.AgentAlias.AgentId,
	}
	id, _ := intflex.FlattenResourceId(idParts, 2, false)
	data.ID = types.StringValue(id)

	alias, err := waitAliasCreated(ctx, conn, id, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Alias Agent (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	_, err = waitAgentVersioned(ctx, conn, *alias.AgentAlias.AgentId, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Alias Agent (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, alias.AgentAlias, &data)...)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentAliasResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data bedrockAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.ID.IsNull() {
		response.Diagnostics.AddError("parsing resource ID", "Agent Alias ID")

		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)
	output, err := findAgentAliasByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Agent Alias (%s)", data.ID.String()), err.Error())

		return
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, output.AgentAlias, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentAliasResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state bedrockAliasResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if bedrockAliasHasChanges(ctx, plan, state) {
		input := &bedrockagent.UpdateAgentAliasInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, input)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.AgentAliasId = fwflex.StringFromFramework(ctx, state.AgentAliasId)
		input.AgentId = fwflex.StringFromFramework(ctx, plan.AgentId)
		input.AgentAliasName = fwflex.StringFromFramework(ctx, plan.AgentAliasName)

		_, err := conn.UpdateAgentAlias(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, "Bedrock Agent", plan.AgentId.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	alias, err := waitAliasUpdated(ctx, conn, plan.ID.ValueString(), r.CreateTimeout(ctx, plan.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Alias Agent (%s) update", plan.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, alias.AgentAlias, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *agentAliasResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data bedrockAliasResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !data.AgentAliasId.IsNull() {
		_, err := conn.DeleteAgentAlias(ctx, &bedrockagent.DeleteAgentAliasInput{
			AgentId:      fwflex.StringFromFramework(ctx, data.AgentId),
			AgentAliasId: fwflex.StringFromFramework(ctx, data.AgentAliasId),
		})

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent (%s)", data.ID.ValueString()), err.Error())

			return
		}
	}
}

func (r *agentAliasResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func (r *agentAliasResource) ImportState(ctx context.Context, request resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, resp)
}

func findAgentAliasByID(ctx context.Context, conn *bedrockagent.Client, id string) (*bedrockagent.GetAgentAliasOutput, error) {
	parts, err := intflex.ExpandResourceId(id, 2, false)
	if err != nil {
		return nil, err
	}
	input := &bedrockagent.GetAgentAliasInput{
		AgentAliasId: aws.String(parts[0]),
		AgentId:      aws.String(parts[1]),
	}

	output, err := conn.GetAgentAlias(ctx, input)

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

func statusAlias(ctx context.Context, conn *bedrockagent.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAgentAliasByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.AgentAlias.AgentAliasStatus), nil
	}
}

func waitAliasCreated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetAgentAliasOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentAliasStatusCreating),
		Target:  enum.Slice(awstypes.AgentAliasStatusPrepared),
		Refresh: statusAlias(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*bedrockagent.GetAgentAliasOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString((*string)(&output.AgentAlias.AgentAliasStatus))))

		return output, err
	}

	return nil, err
}

func waitAliasUpdated(ctx context.Context, conn *bedrockagent.Client, id string, timeout time.Duration) (*bedrockagent.GetAgentAliasOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentAliasStatusUpdating),
		Target:  enum.Slice(awstypes.AgentAliasStatusPrepared),
		Refresh: statusAlias(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*bedrockagent.GetAgentAliasOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString((*string)(&output.AgentAlias.AgentAliasStatus))))

		return output, err
	}

	return nil, err
}

type bedrockAliasResourceModel struct {
	AgentAliasARN        types.String                             `tfsdk:"agent_alias_arn"`
	AgentAliasId         types.String                             `tfsdk:"agent_alias_id"`
	AgentAliasName       types.String                             `tfsdk:"agent_alias_name"`
	AgentAliasStatus     types.String                             `tfsdk:"agent_alias_status"`
	AgentAliasHistory    fwtypes.ListNestedObjectValueOf[history] `tfsdk:"agent_alias_history_events"`
	AgentId              types.String                             `tfsdk:"agent_id"`
	CreatedAt            timetypes.RFC3339                        `tfsdk:"created_at"`
	Description          types.String                             `tfsdk:"description"`
	ID                   types.String                             `tfsdk:"id"`
	RoutingConfiguration fwtypes.ListNestedObjectValueOf[roc]     `tfsdk:"routing_configuration"`
	UpdatedAt            timetypes.RFC3339                        `tfsdk:"updated_at"`
	Tags                 types.Map                                `tfsdk:"tags"`
	TagsAll              types.Map                                `tfsdk:"tags_all"`
	Timeouts             timeouts.Value                           `tfsdk:"timeouts"`
}

type roc struct {
	AgentVersion types.String `tfsdk:"agent_version"`
}

type history struct {
	EndDate              types.String                         `tfsdk:"end_date"`
	StartDate            types.String                         `tfsdk:"start_date"`
	RoutingConfiguration fwtypes.ListNestedObjectValueOf[roc] `tfsdk:"routing_configuration"`
}

func bedrockAliasHasChanges(_ context.Context, plan, state bedrockAliasResourceModel) bool {
	return !plan.AgentAliasName.Equal(state.AgentAliasName) ||
		!plan.Description.Equal(state.Description) ||
		!plan.AgentAliasId.Equal(state.AgentAliasId) ||
		!plan.RoutingConfiguration.Equal(state.RoutingConfiguration)
}
