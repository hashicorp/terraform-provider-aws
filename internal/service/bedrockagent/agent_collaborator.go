// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"fmt"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagent_agent_collaborator", name="Agent Collaborator")
func newAgentCollaboratorResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &agentCollaboratorResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)

	return r, nil
}

type agentCollaboratorResource struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *agentCollaboratorResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			names.AttrID: framework.IDAttribute(),
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"agent_descriptor": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[agentDescriptorModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
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

func (r *agentCollaboratorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data agentCollaboratorResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	var input bedrockagent.AssociateAgentCollaboratorInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.AssociateAgentCollaborator(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Collaborator", err.Error())

		return
	}

	data.CollaboratorID = fwflex.StringToFramework(ctx, output.AgentCollaborator.CollaboratorId)

	id, err := data.setID()
	if err != nil {
		response.Diagnostics.AddError("setting ID", err.Error())

		return
	}
	data.ID = types.StringValue(id)

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.PrepareAgent.ValueBool() {
		if _, err := prepareAgent(ctx, conn, data.AgentID.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
			response.Diagnostics.AddError("preparing Agent", err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *agentCollaboratorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data agentCollaboratorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())
		return
	}

	out, err := findAgentCollaboratorByThreePartKey(ctx, conn, data.AgentID.ValueString(), data.AgentVersion.ValueString(), data.CollaboratorID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Collaborator (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentCollaboratorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new agentCollaboratorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	if !new.AgentDescriptor.Equal(old.AgentDescriptor) ||
		!new.CollaborationInstruction.Equal(old.CollaborationInstruction) ||
		!new.CollaboratorName.Equal(old.CollaboratorName) ||
		!new.RelayConversationHistory.Equal(old.RelayConversationHistory) {
		var input bedrockagent.UpdateAgentCollaboratorInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdateAgentCollaborator(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Agent Collaborator (%s)", new.ID.ValueString()), err.Error())

			return
		}

		if new.PrepareAgent.ValueBool() {
			if _, err := prepareAgent(ctx, conn, new.AgentID.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
				response.Diagnostics.AddError("preparing Agent", err.Error())
				return
			}
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *agentCollaboratorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data agentCollaboratorResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := bedrockagent.DisassociateAgentCollaboratorInput{
		AgentId:        data.AgentID.ValueStringPointer(),
		AgentVersion:   data.AgentVersion.ValueStringPointer(),
		CollaboratorId: data.CollaboratorID.ValueStringPointer(),
	}
	_, err := conn.DisassociateAgentCollaborator(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Collaborator (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if data.PrepareAgent.ValueBool() {
		response.Diagnostics.Append(prepareSupervisorToReleaseCollaborator(ctx, conn, data.AgentID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts))...)
		if response.Diagnostics.HasError() {
			return
		}
	}
}

func (r *agentCollaboratorResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
	// Set prepare_agent to default value on import.
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("prepare_agent"), true)...)
}

func findAgentCollaboratorByThreePartKey(ctx context.Context, conn *bedrockagent.Client, agentID, agentVersion, collaboratorID string) (*awstypes.AgentCollaborator, error) {
	input := bedrockagent.GetAgentCollaboratorInput{
		AgentId:        aws.String(agentID),
		AgentVersion:   aws.String(agentVersion),
		CollaboratorId: aws.String(collaboratorID),
	}

	output, err := conn.GetAgentCollaborator(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AgentCollaborator == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AgentCollaborator, nil
}

type agentCollaboratorResourceModel struct {
	AgentID                  types.String                                          `tfsdk:"agent_id"`
	AgentVersion             types.String                                          `tfsdk:"agent_version"`
	AgentDescriptor          fwtypes.ListNestedObjectValueOf[agentDescriptorModel] `tfsdk:"agent_descriptor"`
	CollaboratorID           types.String                                          `tfsdk:"collaborator_id"`
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
	parts, err := flex.ExpandResourceId(id, agentCollaboratorResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.AgentID = types.StringValue(parts[0])
	m.AgentVersion = types.StringValue(parts[1])
	m.CollaboratorID = types.StringValue(parts[2])

	return nil
}

func (m *agentCollaboratorResourceModel) setID() (string, error) {
	parts := []string{
		m.AgentID.ValueString(),
		m.AgentVersion.ValueString(),
		m.CollaboratorID.ValueString(),
	}

	return flex.FlattenResourceId(parts, agentCollaboratorResourceIDPartCount, false)
}

type agentDescriptorModel struct {
	AliasARN fwtypes.ARN `tfsdk:"alias_arn"`
}
