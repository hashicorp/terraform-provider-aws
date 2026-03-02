// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_agent_runtime_endpoint", name="Agent Runtime Endpoint")
// @Tags(identifierAttribute="agent_runtime_endpoint_arn")
// @Testing(tagsTest=false)
func newAgentRuntimeEndpointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &agentRuntimeEndpointResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type agentRuntimeEndpointResource struct {
	framework.ResourceWithModel[agentRuntimeEndpointResourceModel]
	framework.WithTimeouts
}

func (r *agentRuntimeEndpointResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"agent_runtime_arn":          framework.ARNAttributeComputedOnly(),
			"agent_runtime_endpoint_arn": framework.ARNAttributeComputedOnly(),
			"agent_runtime_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"agent_runtime_version": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 256),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,47}$`), ""),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
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
		},
	}
}

func (r *agentRuntimeEndpointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data agentRuntimeEndpointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	agentRuntimeID, name := fwflex.StringValueFromFramework(ctx, data.AgentRuntimeID), fwflex.StringValueFromFramework(ctx, data.Name)
	var input bedrockagentcorecontrol.CreateAgentRuntimeEndpointInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input, fwflex.WithFieldNamePrefix("AgentRuntimeEndpoint")))
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateAgentRuntimeEndpoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	data.AgentRuntimeARN = fwflex.StringToFramework(ctx, out.AgentRuntimeArn)
	data.AgentRuntimeEndpointARN = fwflex.StringToFramework(ctx, out.AgentRuntimeEndpointArn)
	data.AgentRuntimeVersion = fwflex.StringToFramework(ctx, out.TargetVersion)

	if _, err := waitAgentRuntimeEndpointCreated(ctx, conn, agentRuntimeID, name, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *agentRuntimeEndpointResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data agentRuntimeEndpointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	agentRuntimeID, name := fwflex.StringValueFromFramework(ctx, data.AgentRuntimeID), fwflex.StringValueFromFramework(ctx, data.Name)
	out, err := findAgentRuntimeEndpointByTwoPartKey(ctx, conn, agentRuntimeID, name)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data, fwflex.WithFieldNamePrefix("AgentRuntimeEndpoint")))
	if response.Diagnostics.HasError() {
		return
	}
	data.AgentRuntimeVersion = fwflex.StringToFramework(ctx, out.LiveVersion)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *agentRuntimeEndpointResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old agentRuntimeEndpointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		agentRuntimeID, name := fwflex.StringValueFromFramework(ctx, new.AgentRuntimeID), fwflex.StringValueFromFramework(ctx, new.Name)
		var input bedrockagentcorecontrol.UpdateAgentRuntimeEndpointInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input, fwflex.WithFieldNamePrefix("Endpoint")))
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())

		out, err := conn.UpdateAgentRuntimeEndpoint(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}

		new.AgentRuntimeVersion = fwflex.StringToFramework(ctx, out.TargetVersion)

		if _, err := waitAgentRuntimeEndpointUpdated(ctx, conn, agentRuntimeID, name, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
			return
		}
	} else {
		new.AgentRuntimeVersion = old.AgentRuntimeVersion
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *agentRuntimeEndpointResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data agentRuntimeEndpointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	agentRuntimeID, name := fwflex.StringValueFromFramework(ctx, data.AgentRuntimeID), fwflex.StringValueFromFramework(ctx, data.Name)
	input := bedrockagentcorecontrol.DeleteAgentRuntimeEndpointInput{
		AgentRuntimeId: aws.String(agentRuntimeID),
		ClientToken:    aws.String(sdkid.UniqueId()),
		EndpointName:   aws.String(name),
	}

	_, err := conn.DeleteAgentRuntimeEndpoint(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "was not found") {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}

	if _, err := waitAgentRuntimeEndpointDeleted(ctx, conn, agentRuntimeID, name, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, name)
		return
	}
}

func (r *agentRuntimeEndpointResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, ",")
	if len(parts) != 2 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf(`Unexpected format for import ID (%s), use: "agent_runtime_id,name"`, request.ID))
		return
	}

	agentRuntimeId, endpointName := parts[0], parts[1]

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root("agent_runtime_id"), agentRuntimeId))
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.SetAttribute(ctx, path.Root(names.AttrName), endpointName))
}

func waitAgentRuntimeEndpointCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeID, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentRuntimeEndpointStatusCreating),
		Target:                    enum.Slice(awstypes.AgentRuntimeEndpointStatusReady),
		Refresh:                   statusAgentRuntimeEndpoint(conn, agentRuntimeID, endpointName),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeEndpointUpdated(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeID, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.AgentRuntimeEndpointStatusUpdating),
		Target:                    enum.Slice(awstypes.AgentRuntimeEndpointStatusReady),
		Refresh:                   statusAgentRuntimeEndpoint(conn, agentRuntimeID, endpointName),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAgentRuntimeEndpointDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeID, endpointName string, timeout time.Duration) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AgentRuntimeEndpointStatusDeleting, awstypes.AgentRuntimeEndpointStatusReady),
		Target:  []string{},
		Refresh: statusAgentRuntimeEndpoint(conn, agentRuntimeID, endpointName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(out.FailureReason)))
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAgentRuntimeEndpoint(conn *bedrockagentcorecontrol.Client, agentRuntimeID, endpointName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findAgentRuntimeEndpointByTwoPartKey(ctx, conn, agentRuntimeID, endpointName)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findAgentRuntimeEndpointByTwoPartKey(ctx context.Context, conn *bedrockagentcorecontrol.Client, agentRuntimeID, endpointName string) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	input := bedrockagentcorecontrol.GetAgentRuntimeEndpointInput{
		AgentRuntimeId: aws.String(agentRuntimeID),
		EndpointName:   aws.String(endpointName),
	}

	return findAgentRuntimeEndpoint(ctx, conn, &input)
}

func findAgentRuntimeEndpoint(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetAgentRuntimeEndpointInput) (*bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput, error) {
	out, err := conn.GetAgentRuntimeEndpoint(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || errs.IsAErrorMessageContains[*awstypes.AccessDeniedException](err, "was not found") {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type agentRuntimeEndpointResourceModel struct {
	framework.WithRegionModel
	AgentRuntimeEndpointARN types.String   `tfsdk:"agent_runtime_endpoint_arn"`
	AgentRuntimeARN         types.String   `tfsdk:"agent_runtime_arn"`
	AgentRuntimeID          types.String   `tfsdk:"agent_runtime_id"`
	AgentRuntimeVersion     types.String   `tfsdk:"agent_runtime_version"`
	Description             types.String   `tfsdk:"description"`
	Name                    types.String   `tfsdk:"name"`
	Tags                    tftags.Map     `tfsdk:"tags"`
	TagsAll                 tftags.Map     `tfsdk:"tags_all"`
	Timeouts                timeouts.Value `tfsdk:"timeouts"`
}
