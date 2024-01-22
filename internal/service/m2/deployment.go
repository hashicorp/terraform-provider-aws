// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/m2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="Deployment")
func newResourceDeployment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDeployment{}

	r.SetDefaultCreateTimeout(60 * time.Minute)
	r.SetDefaultUpdateTimeout(60 * time.Minute)
	r.SetDefaultDeleteTimeout(60 * time.Minute)

	return r, nil
}

const (
	ResNameDeployment = "Deployment"
)

type resourceDeployment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDeployment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_m2_deployment"
}

func (r *resourceDeployment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_version": schema.Int64Attribute{
				Required: true,
			},
			"client_token": schema.StringAttribute{
				Optional: true,
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"force_stop": schema.BoolAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"start": schema.BoolAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDeployment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().M2Client(ctx)

	var plan resourceDeploymentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := plan.createDeploymentInput(ctx)

	out, err := conn.CreateDeployment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionCreating, ResNameDeployment, plan.ApplicationId.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.DeploymentId == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionCreating, ResNameDeployment, plan.ApplicationId.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	deploymentId := DeploymentId(plan.ApplicationId.ValueString(), *out.DeploymentId)

	plan.ID = flex.StringValueToFramework(ctx, deploymentId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	deployment, err := waitDeploymentCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForCreation, ResNameDeployment, plan.ApplicationId.String(), err),
			err.Error(),
		)
		return
	}

	if plan.Start.ValueBool() {
		_, err = startApplication(ctx, conn, plan.ApplicationId.ValueString(), createTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForCreation, ResNameDeployment, plan.ApplicationId.String(), err),
				err.Error(),
			)
			return
		}
	}

	plan.refreshFromOutput(ctx, deployment)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDeployment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().M2Client(ctx)

	var state resourceDeploymentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDeploymentByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionSetting, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	state.refreshFromOutput(ctx, out)

	applicationId, _, err := DeploymentParseResourceId(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionSetting, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	app, err := findApplicationByID(ctx, conn, applicationId)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionSetting, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	state.refreshFromApplicationOutput(app)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDeployment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().M2Client(ctx)

	var plan, state resourceDeploymentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringUnknown()

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)

	if !plan.ApplicationVersion.Equal(state.ApplicationVersion) {
		applicationId := flex.StringFromFramework(ctx, plan.ApplicationId)

		// Stop the application if it was running
		if state.Start.ValueBool() {
			err := stopApplicationIfRunning(ctx, conn, *applicationId, plan.ForceStop.ValueBool(), updateTimeout)
			if err != nil {
				resp.Diagnostics.AddError(
					create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameDeployment, plan.ApplicationId.String(), err),
					err.Error(),
				)
				return
			}
		}

		// Create the updated deployment
		in := plan.createDeploymentInput(ctx)

		out, err := conn.CreateDeployment(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameDeployment, plan.ApplicationId.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.DeploymentId == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameDeployment, plan.ApplicationId.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		combinedId := DeploymentId(plan.ApplicationId.ValueString(), *out.DeploymentId)
		plan.ID = flex.StringValueToFramework(ctx, combinedId)

		deployment, err := waitDeploymentUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForUpdate, ResNameDeployment, plan.ApplicationId.String(), err),
				err.Error(),
			)
			return
		}
		plan.refreshFromOutput(ctx, deployment)
	}

	// Start the application if plan says to
	if plan.Start.ValueBool() {
		app, err := startApplication(ctx, conn, plan.ApplicationId.ValueString(), updateTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionUpdating, ResNameDeployment, plan.ApplicationId.String(), err),
				err.Error(),
			)
			return
		}
		plan.refreshFromApplicationOutput(app)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDeployment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().M2Client(ctx)

	var state resourceDeploymentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)

	if state.Start.ValueBool() {
		err := stopApplicationIfRunning(ctx, conn, state.ApplicationId.ValueString(), state.ForceStop.ValueBool(), deleteTimeout)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.M2, create.ErrActionDeleting, ResNameDeployment, state.ApplicationId.String(), err),
				err.Error(),
			)
			return
		}
	}

	in := &m2.DeleteApplicationFromEnvironmentInput{
		ApplicationId: flex.StringFromFramework(ctx, state.ApplicationId),
		EnvironmentId: flex.StringFromFramework(ctx, state.EnvironmentId),
	}

	_, err := conn.DeleteApplicationFromEnvironment(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionDeleting, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitApplicationDeletedFromEnvironment(ctx, conn, state.ApplicationId.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.M2, create.ErrActionWaitingForDeletion, ResNameDeployment, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceDeployment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceDeploymentData) refreshFromOutput(ctx context.Context, out *m2.GetDeploymentOutput) {
	combinedId := DeploymentId(*out.ApplicationId, *out.DeploymentId)
	r.ID = flex.StringValueToFramework(ctx, combinedId)
	r.ApplicationId = flex.StringToFramework(ctx, out.ApplicationId)
	r.ApplicationVersion = flex.Int32ToFramework(ctx, out.ApplicationVersion)
	r.EnvironmentId = flex.StringToFramework(ctx, out.EnvironmentId)
}

func (r *resourceDeploymentData) refreshFromApplicationOutput(app *m2.GetApplicationOutput) {
	r.Start = types.BoolValue(app.Status == awstypes.ApplicationLifecycleRunning)
}

func (r *resourceDeploymentData) createDeploymentInput(ctx context.Context) *m2.CreateDeploymentInput {
	in := &m2.CreateDeploymentInput{
		ApplicationId:      r.ApplicationId.ValueStringPointer(),
		ApplicationVersion: flex.Int32FromFramework(ctx, r.ApplicationVersion),
		EnvironmentId:      r.EnvironmentId.ValueStringPointer(),
	}

	var clientToken string
	if r.ClientToken.IsNull() || r.ClientToken.IsUnknown() {
		clientToken = id.UniqueId()
	} else {
		clientToken = r.ClientToken.ValueString()
	}

	in.ClientToken = aws.String(clientToken)

	return in
}

func waitDeploymentCreated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetDeploymentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DeploymentLifecycleDeploying),
		Target:                    enum.Slice(awstypes.DeploymentLifecycleSucceeded),
		Refresh:                   statusDeployment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetDeploymentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDeploymentUpdated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetDeploymentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.DeploymentLifecycleDeployUpdate),
		Target:                    enum.Slice(awstypes.DeploymentLifecycleSucceeded),
		Refresh:                   statusDeployment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetDeploymentOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDeployment(ctx context.Context, conn *m2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findDeploymentByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func findDeploymentByID(ctx context.Context, conn *m2.Client, id string) (*m2.GetDeploymentOutput, error) {
	applicationId, deploymentId, err := DeploymentParseResourceId(id)
	if err != nil {
		return nil, err
	}

	in := &m2.GetDeploymentInput{
		ApplicationId: aws.String(applicationId),
		DeploymentId:  aws.String(deploymentId),
	}

	out, err := conn.GetDeployment(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceDeploymentData struct {
	ApplicationId      types.String   `tfsdk:"application_id"`
	ApplicationVersion types.Int64    `tfsdk:"application_version"`
	ClientToken        types.String   `tfsdk:"client_token"`
	EnvironmentId      types.String   `tfsdk:"environment_id"`
	ForceStop          types.Bool     `tfsdk:"force_stop"`
	ID                 types.String   `tfsdk:"id"`
	Start              types.Bool     `tfsdk:"start"`
	Timeouts           timeouts.Value `tfsdk:"timeouts"`
}

const deploymentIDSeparator = "_"

func DeploymentId(applicationId, deploymentId string) string {
	parts := []string{applicationId, deploymentId}
	combinedId := strings.Join(parts, deploymentIDSeparator)
	return combinedId
}

func DeploymentParseResourceId(id string) (string, string, error) {
	parts := strings.Split(id, deploymentIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected APPLICATION-ID%[2]sDEPLOYMENT-ID", id, deploymentIDSeparator)
}
