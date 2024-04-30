// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagent_agent_knowledge_base_association", name="Agent Knowledge Base Association")
func newAgentKnowledgeBaseAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAgentKnowledgeBaseAssociation{}

	return r, nil
}

const (
	ResNameAgentKnowledgeBaseAssociation = "Agent Knowledge Base Association"
)

const (
	agentKnowledegeBaseAssociationResourceIDPartCount = 3
)

type resourceAgentKnowledgeBaseAssociation struct {
	framework.ResourceWithConfigure
}

func (r *resourceAgentKnowledgeBaseAssociation) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrockagent_agent_knowledge_base_association"
}

func (r *resourceAgentKnowledgeBaseAssociation) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z]){10}$`), "valid characters are a-z, A-Z, 0-9. The id can have up to 10 characters"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("DRAFT"),
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^DRAFT$`), "Must be DRAFT."),
				},
			},
			"description": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"knowledge_base_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^([0-9a-zA-Z]){10}$`), "valid characters are a-z, A-Z, 0-9. The id can have up to 10 characters"),
				},
			},
			"knowledge_base_state": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.KnowledgeBaseState](),
			},
		},
	}
}

func (r *resourceAgentKnowledgeBaseAssociation) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data resourceAgentKnowledgeBaseAssociationData
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &bedrockagent.AssociateAgentKnowledgeBaseInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.AssociateAgentKnowledgeBase(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Knowledge Base Association", err.Error())
		return
	}
	if out == nil || out.AgentKnowledgeBase == nil {
		response.Diagnostics.AddError("creating Bedrock Agent Knowledge Base Association empty output", err.Error())
		return
	}
	data.Description = fwflex.StringToFramework(ctx, out.AgentKnowledgeBase.Description)
	data.KnowledgeBaseId = fwflex.StringToFramework(ctx, out.AgentKnowledgeBase.KnowledgeBaseId)
	data.KnowledgeBaseState = fwtypes.StringEnumValue[awstypes.KnowledgeBaseState](out.AgentKnowledgeBase.KnowledgeBaseState)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceAgentKnowledgeBaseAssociation) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data resourceAgentKnowledgeBaseAssociationData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	out, err := findAgentKnowledgeBaseAssociationByThreePartID(ctx, conn, data.AgentId.ValueString(), data.AgentVersion.ValueString(), data.KnowledgeBaseId.ValueString())
	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionSetting, ResNameAgentKnowledgeBaseAssociation, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceAgentKnowledgeBaseAssociation) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var new, old resourceAgentKnowledgeBaseAssociationData
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !new.KnowledgeBaseId.Equal(old.KnowledgeBaseId) ||
		!new.KnowledgeBaseState.Equal(old.KnowledgeBaseState) ||
		!new.Description.Equal(old.Description) {

		input := &bedrockagent.UpdateAgentKnowledgeBaseInput{}

		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateAgentKnowledgeBase(ctx, input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameAgentKnowledgeBaseAssociation, new.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.AgentKnowledgeBase == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionUpdating, ResNameAgentKnowledgeBaseAssociation, new.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
		new.Description = fwflex.StringToFramework(ctx, out.AgentKnowledgeBase.Description)
		new.KnowledgeBaseId = fwflex.StringToFramework(ctx, out.AgentKnowledgeBase.KnowledgeBaseId)
		new.KnowledgeBaseState = fwtypes.StringEnumValue[awstypes.KnowledgeBaseState](out.AgentKnowledgeBase.KnowledgeBaseState)
		new.setID()
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceAgentKnowledgeBaseAssociation) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var data resourceAgentKnowledgeBaseAssociationData
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.DisassociateAgentKnowledgeBase(ctx, &bedrockagent.DisassociateAgentKnowledgeBaseInput{
		AgentId:         fwflex.StringFromFramework(ctx, data.AgentId),
		AgentVersion:    fwflex.StringFromFramework(ctx, data.AgentVersion),
		KnowledgeBaseId: fwflex.StringFromFramework(ctx, data.KnowledgeBaseId),
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionDeleting, ResNameAgentKnowledgeBaseAssociation, data.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceAgentKnowledgeBaseAssociation) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func findAgentKnowledgeBaseAssociationByThreePartID(ctx context.Context, conn *bedrockagent.Client, agentID, agentVersion, knowledgeBaseID string) (*awstypes.AgentKnowledgeBase, error) {
	in := &bedrockagent.GetAgentKnowledgeBaseInput{
		AgentId:         aws.String(agentID),
		AgentVersion:    aws.String(agentVersion),
		KnowledgeBaseId: aws.String(knowledgeBaseID),
	}

	out, err := conn.GetAgentKnowledgeBase(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil || out.AgentKnowledgeBase == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.AgentKnowledgeBase, nil
}

type resourceAgentKnowledgeBaseAssociationData struct {
	AgentId            types.String                                    `tfsdk:"agent_id"`
	AgentVersion       types.String                                    `tfsdk:"agent_version"`
	Description        types.String                                    `tfsdk:"description"`
	ID                 types.String                                    `tfsdk:"id"`
	KnowledgeBaseId    types.String                                    `tfsdk:"knowledge_base_id"`
	KnowledgeBaseState fwtypes.StringEnum[awstypes.KnowledgeBaseState] `tfsdk:"knowledge_base_state"`
}

func (m *resourceAgentKnowledgeBaseAssociationData) InitFromID() error {
	id := m.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, agentKnowledegeBaseAssociationResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.AgentId = types.StringValue(parts[0])
	m.AgentVersion = types.StringValue(parts[1])
	m.KnowledgeBaseId = types.StringValue(parts[2])

	return nil
}

func (m *resourceAgentKnowledgeBaseAssociationData) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AgentId.ValueString(), m.AgentVersion.ValueString(), m.KnowledgeBaseId.ValueString()}, agentKnowledegeBaseAssociationResourceIDPartCount, false)))
}
