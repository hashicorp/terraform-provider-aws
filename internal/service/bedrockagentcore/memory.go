// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"context"
	"errors"
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
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_bedrockagentcore_memory", name="Memory")
func newResourceMemory(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceMemory{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameMemory = "Memory"
)

type resourceMemory struct {
	framework.ResourceWithModel[memoryResourceModel]
	framework.WithTimeouts
}

func (r *resourceMemory) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			"encryption_key_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"event_expiry_duration": schema.Int32Attribute{
				Required: true,
			},
			"memory_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"memory_execution_role_arn": schema.StringAttribute{
				Optional: true,
			},
			names.AttrName: schema.StringAttribute{
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

func (r *resourceMemory) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan memoryResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	var input bedrockagentcorecontrol.CreateMemoryInput
	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateMemory(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.Name.String())
		return
	}
	if out == nil || out.Memory == nil {
		smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.Name.String())
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out.Memory, &plan))
	if response.Diagnostics.HasError() {
		return
	}

	plan.MemoryID = fwflex.StringToFramework(ctx, out.Memory.Id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitMemoryCreated(ctx, conn, plan.MemoryID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.MemoryID.String())
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, plan))
}

func (r *resourceMemory) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state memoryResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findMemoryByID(ctx, conn, state.MemoryID.ValueString())
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.MemoryID.String())
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out.Memory, &state))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &state))
}

func (r *resourceMemory) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var plan, state memoryResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.Plan.Get(ctx, &plan))
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.EnrichAppend(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input bedrockagentcorecontrol.UpdateMemoryInput
		smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if response.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateMemory(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, plan.MemoryID.String())
			return
		}
		if out == nil || out.Memory == nil {
			smerr.AddError(ctx, &response.Diagnostics, errors.New("empty output"), smerr.ID, plan.MemoryID.String())
			return
		}

		smerr.EnrichAppend(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out.Memory, &plan))
		if response.Diagnostics.HasError() {
			return
		}
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, response.State.Set(ctx, &plan))
}

func (r *resourceMemory) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().BedrockAgentCoreClient(ctx)

	var state memoryResourceModel
	smerr.EnrichAppend(ctx, &response.Diagnostics, request.State.Get(ctx, &state))
	if response.Diagnostics.HasError() {
		return
	}

	input := bedrockagentcorecontrol.DeleteMemoryInput{
		MemoryId: state.MemoryID.ValueStringPointer(),
	}

	_, err := conn.DeleteMemory(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.MemoryID.String())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitMemoryDeleted(ctx, conn, state.MemoryID.ValueString(), deleteTimeout)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, state.MemoryID.String())
		return
	}
}

func (w *resourceMemory) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) { // nosemgrep:ci.semgrep.framework.with-import-by-id
	resource.ImportStatePassthroughID(ctx, path.Root("memory_id"), request, response)
}

func waitMemoryCreated(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetMemoryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.MemoryStatusCreating),
		Target:                    enum.Slice(awstypes.MemoryStatusActive),
		Refresh:                   statusMemory(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetMemoryOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitMemoryDeleted(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string, timeout time.Duration) (*bedrockagentcorecontrol.GetMemoryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.MemoryStatusDeleting, awstypes.MemoryStatusActive),
		Target:  []string{},
		Refresh: statusMemory(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*bedrockagentcorecontrol.GetMemoryOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusMemory(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findMemoryByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Memory.Status), nil
	}
}

func findMemoryByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetMemoryOutput, error) {
	input := bedrockagentcorecontrol.GetMemoryInput{
		MemoryId: aws.String(id),
	}

	out, err := conn.GetMemory(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Memory == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(&input))
	}

	return out, nil
}

type memoryResourceModel struct {
	framework.WithRegionModel
	ARN                    fwtypes.ARN    `tfsdk:"arn"`
	ClientToken            types.String   `tfsdk:"client_token"`
	Description            types.String   `tfsdk:"description"`
	EncryptionKeyArn       fwtypes.ARN    `tfsdk:"encryption_key_arn"`
	EventExpiryDuration    types.Int32    `tfsdk:"event_expiry_duration"`
	MemoryID               types.String   `tfsdk:"memory_id"`
	MemoryExecutionRoleArn types.String   `tfsdk:"memory_execution_role_arn"`
	Name                   types.String   `tfsdk:"name"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
}
