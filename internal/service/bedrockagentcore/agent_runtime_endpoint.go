// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_agent_runtime_endpoint", name="Agent Runtime Endpoint")
func newResourceAgentRuntimeEndpoint(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAgentRuntimeEndpoint{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAgentRuntimeEndpoint = "Agent Runtime Endpoint"
)

type resourceAgentRuntimeEndpoint struct {
	framework.ResourceWithModel[resourceAgentRuntimeEndpointModel]
	framework.WithTimeouts
}

func (r *resourceAgentRuntimeEndpoint) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agent_runtime_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"agent_runtime_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_token": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_runtime_version": schema.StringAttribute{
				Required: true,
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

func (r *resourceAgentRuntimeEndpoint) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan resourceAgentRuntimeEndpointModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateAgentRuntimeEndpointInput
	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AgentRuntimeEndpoint")))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateAgentRuntimeEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("AgentRuntimeEndpoint")))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.AgentRuntimeVersion = flex.StringToFramework(ctx, out.TargetVersion)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitAgentRuntimeEndpointCreated(ctx, conn, plan.AgentRuntimeId.ValueString(), plan.Name.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceAgentRuntimeEndpoint) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceAgentRuntimeEndpointModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAgentRuntimeEndpointByRuntimeIDAndName(ctx, conn, state.AgentRuntimeId.ValueString(), state.Name.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state, flex.WithFieldNamePrefix("AgentRuntimeEndpoint")))
	if resp.Diagnostics.HasError() {
		return
	}
	state.AgentRuntimeVersion = flex.StringToFramework(ctx, out.LiveVersion)

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *resourceAgentRuntimeEndpoint) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state resourceAgentRuntimeEndpointModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateAgentRuntimeEndpointInput
		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("Endpoint")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateAgentRuntimeEndpoint(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
			return
		}
		if out == nil {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
			return
		}

		smerr.EnrichAppend(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan, flex.WithFieldNamePrefix("AgentRuntimeEndpoint")))
		if resp.Diagnostics.HasError() {
			return
		}
		plan.AgentRuntimeVersion = flex.StringToFramework(ctx, out.TargetVersion)
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitAgentRuntimeEndpointUpdated(ctx, conn, plan.AgentRuntimeId.ValueString(), plan.Name.ValueString(), updateTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *resourceAgentRuntimeEndpoint) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state resourceAgentRuntimeEndpointModel
	smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteAgentRuntimeEndpointInput{
		AgentRuntimeId: state.AgentRuntimeId.ValueStringPointer(),
		EndpointName:   state.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteAgentRuntimeEndpoint(ctx, &input)
	if err != nil {
		switch {
		case errs.IsA[*awstypes.ResourceNotFoundException](err):
			return

		case errs.IsA[*awstypes.AccessDeniedException](err):
			msg := err.Error()
			if strings.Contains(msg, "was not found") {
				return
			}
		default:
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
			return
		}
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitAgentRuntimeEndpointDeleted(ctx, conn, state.AgentRuntimeId.ValueString(), state.Name.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.Name.String())
		return
	}
}

func (r *resourceAgentRuntimeEndpoint) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Resource Import Invalid ID", fmt.Sprintf(`Unexpected format for import ID (%s), use: "agent_runtime_id,endpoint_name"`, req.ID))
		return
	}

	agentRuntimeId, endpointName := parts[0], parts[1]

	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root("agent_runtime_id"), agentRuntimeId))
	smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.SetAttribute(ctx, path.Root(names.AttrName), endpointName))
}

func waitAgentRuntimeEndpointCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeId, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentEndpointStatusCreating),
		Target:                    enum.Slice(awstypes.AgentEndpointStatusReady),
		Refresh:                   statusAgentRuntimeEndpoint(ctx, conn, agentRuntimeId, endpointName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeEndpointUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeId, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentEndpointStatusUpdating),
		Target:                    enum.Slice(awstypes.AgentEndpointStatusReady),
		Refresh:                   statusAgentRuntimeEndpoint(ctx, conn, agentRuntimeId, endpointName),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeEndpointDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeId, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentEndpointStatusDeleting, awstypes.AgentEndpointStatusReady),
		Target:  []string{},
		Refresh: statusAgentRuntimeEndpoint(ctx, conn, agentRuntimeId, endpointName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAgentRuntimeEndpoint(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeId, endpointName string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findAgentRuntimeEndpointByRuntimeIDAndName(ctx, conn, agentRuntimeId, endpointName)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findAgentRuntimeEndpointByRuntimeIDAndName(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeID, endpointName string) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	input := bedrockagentcorecontrol.GetAgentRuntimeEndpointInput{
		AgentRuntimeId: aws.String(agentRuntimeID),
		EndpointName:   aws.String(endpointName),
	}

	out, err := conn.GetAgentRuntimeEndpoint(ctx, &input)
	if err != nil {
		switch {
		case errs.IsA[*awstypes.ResourceNotFoundException](err):
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})

		case errs.IsA[*awstypes.AccessDeniedException](err):
			msg := err.Error()
			if strings.Contains(msg, "was not found") {
				return nil, smarterr.NewError(&retry.NotFoundError{
					LastError:   err,
					LastRequest: &input,
				})
			}
		default:
			return nil, smarterr.NewError(err)
		}
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type resourceAgentRuntimeEndpointModel struct {
	framework.WithRegionModel
	ARN                 fwtypes.ARN    `tfsdk:"arn"`
	AgentRuntimeARN     fwtypes.ARN    `tfsdk:"agent_runtime_arn"`
	AgentRuntimeId      types.String   `tfsdk:"agent_runtime_id"`
	ClientToken         types.String   `tfsdk:"client_token"`
	Description         types.String   `tfsdk:"description"`
	Name                types.String   `tfsdk:"name"`
	AgentRuntimeVersion types.String   `tfsdk:"agent_runtime_version"`
	Timeouts            timeouts.Value `tfsdk:"timeouts"`
}
