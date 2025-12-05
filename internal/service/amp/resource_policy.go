// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	awstypes "github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_prometheus_resource_policy", name="Resource Policy")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp;amp.DescribeResourcePolicyOutput")
func newResourcePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyResource{}

	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultDeleteTimeout(5 * time.Minute)
	r.SetDefaultUpdateTimeout(5 * time.Minute)

	return r, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithModel[resourcePolicyResourceModel]
	framework.WithTimeouts
}

func (r *resourcePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"policy_document": schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			"revision_id": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"workspace_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Update: true,
			}),
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourcePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	workspaceID := fwflex.StringValueFromFramework(ctx, data.WorkspaceID)
	input := amp.PutResourcePolicyInput{
		ClientToken:    aws.String(sdkid.UniqueId()),
		PolicyDocument: fwflex.StringFromFramework(ctx, data.PolicyDocument),
		WorkspaceId:    aws.String(workspaceID),
	}

	if !data.RevisionID.IsNull() {
		input.RevisionId = fwflex.StringFromFramework(ctx, data.RevisionID)
	}

	output, err := conn.PutResourcePolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Prometheus Workspace (%s) Resource Policy", workspaceID), err.Error())
		return
	}

	// Set values for unknowns.
	data.RevisionID = fwflex.StringToFramework(ctx, output.RevisionId)

	if _, err := waitResourcePolicyCreated(ctx, conn, workspaceID, r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Workspace (%s) Resource Policy create", workspaceID), err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	workspaceID := fwflex.StringValueFromFramework(ctx, data.WorkspaceID)
	output, err := findResourcePolicyByWorkspaceID(ctx, conn, workspaceID)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Prometheus Workspace (%s) Resource Policy", workspaceID), err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old resourcePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	if !new.PolicyDocument.Equal(old.PolicyDocument) || !new.RevisionID.Equal(old.RevisionID) {
		workspaceID := fwflex.StringValueFromFramework(ctx, new.WorkspaceID)
		input := amp.PutResourcePolicyInput{
			ClientToken:    aws.String(sdkid.UniqueId()),
			PolicyDocument: fwflex.StringFromFramework(ctx, new.PolicyDocument),
			WorkspaceId:    aws.String(workspaceID),
		}

		if !new.RevisionID.IsNull() {
			input.RevisionId = fwflex.StringFromFramework(ctx, new.RevisionID)
		}

		output, err := conn.PutResourcePolicy(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Prometheus Workspace (%s) Resource Policy", workspaceID), err.Error())
			return
		}

		new.RevisionID = fwflex.StringToFramework(ctx, output.RevisionId)

		if _, err := waitResourcePolicyUpdated(ctx, conn, workspaceID, r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Workspace (%s) Resource Policy update", workspaceID), err.Error())
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourcePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().AMPClient(ctx)

	workspaceID := fwflex.StringValueFromFramework(ctx, data.WorkspaceID)
	input := amp.DeleteResourcePolicyInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		WorkspaceId: aws.String(workspaceID),
	}

	if !data.RevisionID.IsNull() {
		input.RevisionId = fwflex.StringFromFramework(ctx, data.RevisionID)
	}

	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Prometheus Workspace (%s) Resource Policy", workspaceID), err.Error())
		return
	}

	if _, err := waitResourcePolicyDeleted(ctx, conn, workspaceID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Workspace (%s) Resource Policy delete", workspaceID), err.Error())
		return
	}
}

func (r *resourcePolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("workspace_id"), request, response)
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	PolicyDocument fwtypes.IAMPolicy `tfsdk:"policy_document"`
	RevisionID     types.String      `tfsdk:"revision_id"`
	Timeouts       timeouts.Value    `tfsdk:"timeouts"`
	WorkspaceID    types.String      `tfsdk:"workspace_id"`
}

func findResourcePolicyByWorkspaceID(ctx context.Context, conn *amp.Client, workspaceID string) (*amp.DescribeResourcePolicyOutput, error) {
	input := amp.DescribeResourcePolicyInput{
		WorkspaceId: aws.String(workspaceID),
	}

	return findResourcePolicy(ctx, conn, &input)
}

func findResourcePolicy(ctx context.Context, conn *amp.Client, input *amp.DescribeResourcePolicyInput) (*amp.DescribeResourcePolicyOutput, error) {
	output, err := conn.DescribeResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PolicyDocument == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusResourcePolicy(ctx context.Context, conn *amp.Client, workspaceID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findResourcePolicyByWorkspaceID(ctx, conn, workspaceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.PolicyStatus), nil
	}
}

func waitResourcePolicyCreated(ctx context.Context, conn *amp.Client, workspaceID string, timeout time.Duration) (*amp.DescribeResourcePolicyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspacePolicyStatusCodeCreating),
		Target:  enum.Slice(awstypes.WorkspacePolicyStatusCodeActive),
		Refresh: statusResourcePolicy(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*amp.DescribeResourcePolicyOutput); ok {
		return output, err
	}

	return nil, err
}

func waitResourcePolicyUpdated(ctx context.Context, conn *amp.Client, workspaceID string, timeout time.Duration) (*amp.DescribeResourcePolicyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspacePolicyStatusCodeUpdating),
		Target:  enum.Slice(awstypes.WorkspacePolicyStatusCodeActive),
		Refresh: statusResourcePolicy(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*amp.DescribeResourcePolicyOutput); ok {
		return output, err
	}

	return nil, err
}

func waitResourcePolicyDeleted(ctx context.Context, conn *amp.Client, workspaceID string, timeout time.Duration) (*amp.DescribeResourcePolicyOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspacePolicyStatusCodeDeleting, awstypes.WorkspacePolicyStatusCodeActive),
		Target:  []string{},
		Refresh: statusResourcePolicy(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*amp.DescribeResourcePolicyOutput); ok {
		return output, err
	}

	return nil, err
}
