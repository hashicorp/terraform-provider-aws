// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagent_agent_collaborator", name="Agent Collaborator")
func newResourceAgentCollaborator(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAgentCollaborator{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

const (
	ResNameAgentCollaborator = "Agent Collaborator"
)

type resourceAgentCollaborator struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceAgentCollaborator) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrockagent_agent_collaborator"
}

func (r *resourceAgentCollaborator) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"agent_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("DRAFT"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collaborator_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"collaboration_instruction": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 4000),
				},
			},
			"collaborator_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z][_-]?){1,100}$`), "valid characters are a-z, A-Z, 0-9, _ (underscore) and - (hyphen). The name can have up to 100 characters"),
				},
			},
			"prepare_agent": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"relay_conversation_history": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RelayConversationHistory](),
				Optional:   true,
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"agent_descriptor": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[agentDescriptorModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alias_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceAgentCollaborator) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data agentCollaboratorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var input bedrockagent.AssociateAgentCollaboratorInput

	response.Diagnostics.Append(flex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.AssociateAgentCollaborator(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameAgentCollaborator, data.CollaboratorName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.AgentCollaborator == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameAgentCollaborator, data.CollaboratorName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	data.CollaboratorId = flex.StringToFramework(ctx, out.AgentCollaborator.CollaboratorId)
	id, err := data.setID()
	if err != nil {
		response.Diagnostics.AddError("flattening resource ID Bedrock Agent Action Group", err.Error())
		return
	}
	data.ID = types.StringValue(id)

	response.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.PrepareAgent.ValueBool() {
		if _, err := prepareAgent(ctx, conn, data.AgentId.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError("preparing Agent", err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceAgentCollaborator) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data agentCollaboratorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	out, err := findAgentCollaboratorByThreePartKey(ctx, conn, data.AgentId.ValueString(), data.AgentVersion.ValueString(), data.CollaboratorId.ValueString())
	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionSetting, ResNameAgentCollaborator, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceAgentCollaborator) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var plan, state agentCollaboratorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.AgentDescriptor.Equal(state.AgentDescriptor) ||
		!plan.CollaborationInstruction.Equal(state.CollaborationInstruction) ||
		!plan.CollaboratorName.Equal(state.CollaboratorName) ||
		!plan.RelayConversationHistory.Equal(state.RelayConversationHistory) {
		var input bedrockagent.UpdateAgentCollaboratorInput
		response.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateAgentCollaborator(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameAgentCollaborator, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.AgentCollaborator == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameAgentCollaborator, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		response.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if response.Diagnostics.HasError() {
			return
		}

		if plan.PrepareAgent.ValueBool() {
			if _, err := prepareAgent(ctx, conn, plan.AgentId.ValueString(), r.UpdateTimeout(ctx, plan.Timeouts)); err != nil {
				response.Diagnostics.AddError("preparing Agent", err.Error())
				return
			}
		}
	}
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceAgentCollaborator) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var state agentCollaboratorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := bedrockagent.DisassociateAgentCollaboratorInput{
		AgentId:        state.AgentId.ValueStringPointer(),
		AgentVersion:   state.AgentVersion.ValueStringPointer(),
		CollaboratorId: state.CollaboratorId.ValueStringPointer(),
	}

	_, err := conn.DisassociateAgentCollaborator(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionDeleting, ResNameAgentCollaborator, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	if state.PrepareAgent.ValueBool() || state.PrepareAgent.IsNull() {
		if _, err := prepareSupervisorToReleaseCollaborator(ctx, conn, state.AgentId.ValueString(), r.UpdateTimeout(ctx, state.Timeouts)); err != nil {
			response.Diagnostics.AddError("preparing Agent", err.Error())
			return
		}
	}
}

func (r *resourceAgentCollaborator) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
	// Set prepare_agent to default value on import
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("prepare_agent"), true)...)
}

func findAgentCollaboratorByThreePartKey(ctx context.Context, conn *bedrockagent.Client, agentId string, agentVersion string, collaboratorId string) (*awstypes.AgentCollaborator, error) {
	in := &bedrockagent.GetAgentCollaboratorInput{
		AgentId:        aws.String(agentId),
		AgentVersion:   aws.String(agentVersion),
		CollaboratorId: aws.String(collaboratorId),
	}

	out, err := conn.GetAgentCollaborator(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.AgentCollaborator == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AgentCollaborator, nil
}

type agentCollaboratorResourceModel struct {
	AgentId                  types.String                                          `tfsdk:"agent_id"`
	AgentVersion             types.String                                          `tfsdk:"agent_version"`
	AgentDescriptor          fwtypes.ListNestedObjectValueOf[agentDescriptorModel] `tfsdk:"agent_descriptor"`
	CollaboratorId           types.String                                          `tfsdk:"collaborator_id"`
	CollaborationInstruction types.String                                          `tfsdk:"collaboration_instruction"`
	CollaboratorName         types.String                                          `tfsdk:"collaborator_name"`
	ID                       types.String                                          `tfsdk:"id"`
	PrepareAgent             types.Bool                                            `tfsdk:"prepare_agent"`
	RelayConversationHistory fwtypes.StringEnum[awstypes.RelayConversationHistory] `tfsdk:"relay_conversation_history"`
	Timeouts                 timeouts.Value                                        `tfsdk:"timeouts"`
}

const (
	agentCollaboratorResourceIDPartCount = 3
)

func (m *agentCollaboratorResourceModel) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := fwflex.ExpandResourceId(id, agentCollaboratorResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.AgentId = types.StringValue(parts[0])
	m.AgentVersion = types.StringValue(parts[1])
	m.CollaboratorId = types.StringValue(parts[2])

	return nil
}

func (m *agentCollaboratorResourceModel) setID() (string, error) {
	parts := []string{
		m.AgentId.ValueString(),
		m.AgentVersion.ValueString(),
		m.CollaboratorId.ValueString(),
	}

	return fwflex.FlattenResourceId(parts, agentCollaboratorResourceIDPartCount, false)
}

type agentDescriptorModel struct {
	AliasArn fwtypes.ARN `tfsdk:"alias_arn"`
}
