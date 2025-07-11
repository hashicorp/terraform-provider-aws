// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagent

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagent"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagent_agent_prepare", name="Agent Prepare")
func newResourceAgentPrepare(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAgentPrepare{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAgentPrepare = "Agent Prepare"
)

type resourceAgentPrepare struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceAgentPrepare) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[0-9a-zA-Z]{10}$`),
						"must be a 10-character alphanumeric string",
					),
				},
			},
			"prepared_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceAgentPrepare) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var plan resourceAgentPrepareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &bedrockagent.PrepareAgentInput{
		AgentId: plan.ID.ValueStringPointer(),
	}

	out, err := conn.PrepareAgent(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameAgentPrepare, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionCreating, ResNameAgentPrepare, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	// Set the computed values from the response
	if out.PreparedAt != nil {
		plan.PreparedAt = timetypes.NewRFC3339TimeValue(*out.PreparedAt)
	}

	// Wait for the agent to be prepared
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitAgentPrepared(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionWaitingForCreation, ResNameAgentPrepare, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceAgentPrepare) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentClient(ctx)

	var state resourceAgentPrepareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAgentByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.BedrockAgent, create.ErrActionReading, ResNameAgentPrepare, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// Update state with current agent information
	if out.PreparedAt != nil {
		state.PreparedAt = timetypes.NewRFC3339TimeValue(*out.PreparedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAgentPrepare) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *resourceAgentPrepare) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

type resourceAgentPrepareModel struct {
	ID         types.String      `tfsdk:"id"`
	PreparedAt timetypes.RFC3339 `tfsdk:"prepared_at"`
	Timeouts   timeouts.Value    `tfsdk:"timeouts"`
}

// No sweep function needed for agent prepare action
