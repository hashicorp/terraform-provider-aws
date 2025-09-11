// Copyright (c) HashiCorp, Inc.
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

// @FrameworkResource("aws_prometheus_workspace_resource_policy", name="Workspace Resource Policy")
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
	framework.WithImportByID
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
				Required: false,
				Computed: true,
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

	input := amp.PutResourcePolicyInput{
		ClientToken:    aws.String(sdkid.UniqueId()),
		PolicyDocument: fwflex.StringFromFramework(ctx, data.PolicyDocument),
		WorkspaceId:    fwflex.StringFromFramework(ctx, data.WorkspaceId),
	}

	if !data.RevisionId.IsNull() {
		input.RevisionId = fwflex.StringFromFramework(ctx, data.RevisionId)
	}

	output, err := conn.PutResourcePolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Prometheus Workspace Resource Policy", err.Error())
		return
	}

	// Set values for unknowns.
	// data.PolicyStatus = fwflex.StringValueToFramework(ctx, string(output.PolicyStatus))
	data.RevisionId = fwflex.StringToFramework(ctx, output.RevisionId)

	// policy, err :=

	if waitResourcePolicyCreated(ctx, conn, data.WorkspaceId.ValueString(), r.CreateTimeout(ctx, data.Timeouts)); err != nil {
		// response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID) // Set 'id' so as to taint the resource.
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Workspace Resource Policy (%s) create", data.WorkspaceId.ValueString()), err.Error())
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

	policy, err := findResourcePolicyByWorkspaceID(ctx, conn, data.WorkspaceId.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Prometheus Workspace Resource Policy (%s)", data.WorkspaceId.ValueString()), err.Error())
		return
	}

	// Set attributes for import.
	data.PolicyDocument = fwflex.StringToFramework(ctx, policy.PolicyDocument)
	// data.PolicyStatus = fwflex.StringValueToFramework(ctx, string(policy.PolicyStatus))
	data.RevisionId = fwflex.StringToFramework(ctx, policy.RevisionId)

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

	if !new.PolicyDocument.Equal(old.PolicyDocument) {
		input := amp.PutResourcePolicyInput{
			ClientToken:    aws.String(sdkid.UniqueId()),
			PolicyDocument: fwflex.StringFromFramework(ctx, new.PolicyDocument),
			WorkspaceId:    fwflex.StringFromFramework(ctx, new.WorkspaceId),
		}

		if !new.RevisionId.IsNull() {
			input.RevisionId = fwflex.StringFromFramework(ctx, new.RevisionId)
		}

		output, err := conn.PutResourcePolicy(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Prometheus Workspace Resource Policy (%s)", new.WorkspaceId.ValueString()), err.Error())
			return
		}

		// new.PolicyStatus = fwflex.StringValueToFramework(ctx, string(output.PolicyStatus))
		new.RevisionId = fwflex.StringToFramework(ctx, output.RevisionId)

		if _, err := waitResourcePolicyUpdated(ctx, conn, new.WorkspaceId.ValueString(), r.UpdateTimeout(ctx, new.Timeouts)); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Workspace Resource Policy (%s) update", new.WorkspaceId.ValueString()), err.Error())
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

	input := amp.DeleteResourcePolicyInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		WorkspaceId: fwflex.StringFromFramework(ctx, data.WorkspaceId),
	}

	if !data.RevisionId.IsNull() {
		input.RevisionId = fwflex.StringFromFramework(ctx, data.RevisionId)
	}

	_, err := conn.DeleteResourcePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Prometheus Workspace Resource Policy (%s)", data.WorkspaceId.ValueString()), err.Error())
		return
	}

	if _, err := waitResourcePolicyDeleted(ctx, conn, data.WorkspaceId.ValueString(), r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Prometheus Workspace Resource Policy (%s) delete", data.WorkspaceId.ValueString()), err.Error())
		return
	}
}

type resourcePolicyResourceModel struct {
	framework.WithRegionModel
	// ID             types.String   `tfsdk:"id"`
	PolicyDocument types.String `tfsdk:"policy_document"`
	// PolicyStatus   types.String   `tfsdk:"policy_status"`
	RevisionId  types.String   `tfsdk:"revision_id"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
	WorkspaceId types.String   `tfsdk:"workspace_id"`
}

func findResourcePolicyByWorkspaceID(ctx context.Context, conn *amp.Client, workspaceID string) (*amp.DescribeResourcePolicyOutput, error) {
	input := amp.DescribeResourcePolicyInput{
		WorkspaceId: aws.String(workspaceID),
	}

	output, err := conn.DescribeResourcePolicy(ctx, &input)

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
		Pending: enum.Slice("CREATING"),
		Target:  enum.Slice("ACTIVE"),
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
		Pending: enum.Slice("UPDATING"),
		Target:  enum.Slice("ACTIVE"),
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
		Pending: enum.Slice("ACTIVE", "DELETING"),
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
