// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagent/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// @FrameworkResource(name="Agent Knowledge Base Association")
func newAgentKnowledgeBaseAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &agentKnowledgeBaseAssociationResource{}

	return r, nil
}

type agentKnowledgeBaseAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (*agentKnowledgeBaseAssociationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_bedrockagent_agent_knowledge_base_association"
}

func (r *agentKnowledgeBaseAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
				Validators: []validator.String{
					stringvalidator.OneOf("DRAFT"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 200),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"knowledge_base_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"knowledge_base_state": schema.StringAttribute{
				Required:   true,
				CustomType: fwtypes.StringEnumType[awstypes.KnowledgeBaseState](),
			},
		},
	}
}

func (r *agentKnowledgeBaseAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data agentKnowledgeBaseAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.AssociateAgentKnowledgeBaseInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.AssociateAgentKnowledgeBase(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("creating Bedrock Agent Knowledge Base Association", err.Error())
		return
	}

	// Set values for unknowns.
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *agentKnowledgeBaseAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data agentKnowledgeBaseAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	output, err := findAgentKnowledgeBaseAssociationByThreePartKey(ctx, conn, data.AgentID.ValueString(), data.AgentVersion.ValueString(), data.KnowledgeBaseID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Agent Knowledge Base Association (%s)", data.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *agentKnowledgeBaseAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new agentKnowledgeBaseAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	input := &bedrockagent.UpdateAgentKnowledgeBaseInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateAgentKnowledgeBase(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Bedrock Agent Knowledge Base Association (%s)", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *agentKnowledgeBaseAssociationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data agentKnowledgeBaseAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentClient(ctx)

	_, err := conn.DisassociateAgentKnowledgeBase(ctx, &bedrockagent.DisassociateAgentKnowledgeBaseInput{
		AgentId:         fwflex.StringFromFramework(ctx, data.AgentID),
		AgentVersion:    fwflex.StringFromFramework(ctx, data.AgentVersion),
		KnowledgeBaseId: fwflex.StringFromFramework(ctx, data.KnowledgeBaseID),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Bedrock Agent Knowledge Base Association (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findAgentKnowledgeBaseAssociationByThreePartKey(ctx context.Context, conn *bedrockagent.Client, agentID, agentVersion, knowledgeBaseID string) (*awstypes.AgentKnowledgeBase, error) {
	input := &bedrockagent.GetAgentKnowledgeBaseInput{
		AgentId:         aws.String(agentID),
		AgentVersion:    aws.String(agentVersion),
		KnowledgeBaseId: aws.String(knowledgeBaseID),
	}

	output, err := conn.GetAgentKnowledgeBase(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AgentKnowledgeBase == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.AgentKnowledgeBase, nil
}

type agentKnowledgeBaseAssociationResourceModel struct {
	AgentID            types.String                                    `tfsdk:"agent_id"`
	AgentVersion       types.String                                    `tfsdk:"agent_version"`
	Description        types.String                                    `tfsdk:"description"`
	ID                 types.String                                    `tfsdk:"id"`
	KnowledgeBaseID    types.String                                    `tfsdk:"knowledge_base_id"`
	KnowledgeBaseState fwtypes.StringEnum[awstypes.KnowledgeBaseState] `tfsdk:"knowledge_base_state"`
}

const (
	agentKnowledgeBaseAssociationResourceIDPartCount = 3
)

func (m *agentKnowledgeBaseAssociationResourceModel) InitFromID() error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), agentKnowledgeBaseAssociationResourceIDPartCount, false)

	if err != nil {
		return err
	}

	m.AgentID = types.StringValue(parts[0])
	m.AgentVersion = types.StringValue(parts[1])
	m.KnowledgeBaseID = types.StringValue(parts[2])

	return nil
}

func (m *agentKnowledgeBaseAssociationResourceModel) setID() {
	m.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{m.AgentID.ValueString(), m.AgentVersion.ValueString(), m.KnowledgeBaseID.ValueString()}, agentKnowledgeBaseAssociationResourceIDPartCount, false)))
}
