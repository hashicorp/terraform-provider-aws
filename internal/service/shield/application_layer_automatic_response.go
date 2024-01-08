// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Application Layer Automatic Response")
func newResourceApplicationLayerAutomaticResponse(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceApplicationLayerAutomaticResponse{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameApplicationLayerAutomaticResponse = "Application Layer Automatic Response"
)

type resourceApplicationLayerAutomaticResponse struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceApplicationLayerAutomaticResponse) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_shield_application_layer_automatic_response"
}

func (r *resourceApplicationLayerAutomaticResponse) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"resource_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"action": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"BLOCK", "COUNT"}...),
				},
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

func (r *resourceApplicationLayerAutomaticResponse) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var plan resourceApplicationLayerAutomaticResponseData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	action := &shield.ResponseAction{}
	switch plan.Action.ValueString() {
	case "BLOCK":
		action.Block = &shield.BlockAction{}
	case "COUNT":
		action.Count = &shield.CountAction{}
	}

	in := &shield.EnableApplicationLayerAutomaticResponseInput{
		ResourceArn: aws.String(plan.ResourceARN.ValueString()),
		Action:      action,
	}

	_, err := conn.EnableApplicationLayerAutomaticResponseWithContext(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameApplicationLayerAutomaticResponse, plan.ResourceARN.String(), err),
			err.Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitApplicationLayerAutomaticResponseCreated(ctx, conn, plan.ResourceARN.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForCreation, ResNameApplicationLayerAutomaticResponse, plan.ResourceARN.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(plan.ResourceARN.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceApplicationLayerAutomaticResponse) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var state resourceApplicationLayerAutomaticResponseData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.DescribeProtectionInput{
		ResourceArn: aws.String(state.ID.ValueString()),
	}

	out, err := conn.DescribeProtectionWithContext(ctx, in)
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionSetting, ResNameApplicationLayerAutomaticResponse, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out != nil &&
		out.Protection != nil &&
		out.Protection.ApplicationLayerAutomaticResponseConfiguration != nil &&
		out.Protection.ApplicationLayerAutomaticResponseConfiguration.Action != nil {
		if out.Protection.ApplicationLayerAutomaticResponseConfiguration.Action.Block != nil {
			state.Action = types.StringValue("BLOCK")
		}
		if out.Protection.ApplicationLayerAutomaticResponseConfiguration.Action.Count != nil {
			state.Action = types.StringValue("COUNT")
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplicationLayerAutomaticResponse) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var plan, state resourceApplicationLayerAutomaticResponseData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Action.Equal(state.Action) {
		action := &shield.ResponseAction{}
		switch plan.Action.ValueString() {
		case "BLOCK":
			action.Block = &shield.BlockAction{}
		case "COUNT":
			action.Count = &shield.CountAction{}
		}
		in := &shield.UpdateApplicationLayerAutomaticResponseInput{
			ResourceArn: aws.String(plan.ResourceARN.ValueString()),
			Action:      action,
		}

		out, err := conn.UpdateApplicationLayerAutomaticResponseWithContext(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameApplicationLayerAutomaticResponse, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameApplicationLayerAutomaticResponse, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitApplicationLayerAutomaticResponseUpdated(ctx, conn, plan.ResourceARN.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForUpdate, ResNameApplicationLayerAutomaticResponse, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceApplicationLayerAutomaticResponse) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ShieldConn(ctx)

	var state resourceApplicationLayerAutomaticResponseData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.DisableApplicationLayerAutomaticResponseInput{
		ResourceArn: aws.String(state.ResourceARN.ValueString()),
	}

	protectionOutput, err := conn.DescribeProtectionWithContext(ctx, &shield.DescribeProtectionInput{
		ResourceArn: aws.String(state.ResourceARN.ValueString()),
	})
	if err != nil {
		var nfe *shield.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameApplicationLayerAutomaticResponse, state.ID.String(), err),
			err.Error(),
		)
		return
	}
	if protectionOutput != nil {
		if protectionOutput.Protection.ApplicationLayerAutomaticResponseConfiguration != nil && protectionOutput.Protection.ApplicationLayerAutomaticResponseConfiguration.Status != nil {
			if aws.StringValue(protectionOutput.Protection.ApplicationLayerAutomaticResponseConfiguration.Status) == "DISABLED" {
				return
			}
		}
	} else {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameApplicationLayerAutomaticResponse, state.ID.String(), nil),
			errors.New("empty output").Error(),
		)
	}

	_, err = conn.DisableApplicationLayerAutomaticResponseWithContext(ctx, in)
	if err != nil {
		var nfe *shield.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameApplicationLayerAutomaticResponse, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	// TIP: -- 5. Use a waiter to wait for delete to complete
	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitApplicationLayerAutomaticResponseDeleted(ctx, conn, state.ResourceARN.ValueString(), deleteTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForDeletion, ResNameApplicationLayerAutomaticResponse, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceApplicationLayerAutomaticResponse) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitApplicationLayerAutomaticResponseCreated(ctx context.Context, conn *shield.Shield, resourceArn string, timeout time.Duration) (*shield.DescribeProtectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusApplicationLayerAutomaticResponse(ctx, conn, resourceArn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeProtectionOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationLayerAutomaticResponseUpdated(ctx context.Context, conn *shield.Shield, resourceArn string, timeout time.Duration) (*shield.DescribeProtectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusApplicationLayerAutomaticResponse(ctx, conn, resourceArn),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeProtectionOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationLayerAutomaticResponseDeleted(ctx context.Context, conn *shield.Shield, resourceArn string, timeout time.Duration) (*shield.DescribeProtectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusApplicationLayerAutomaticResponseDeleted(ctx, conn, resourceArn),
		Timeout: timeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*shield.DescribeProtectionOutput); ok {
		return out, err
	}

	return nil, err
}

func statusApplicationLayerAutomaticResponse(ctx context.Context, conn *shield.Shield, resourceArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeApplicationLayerAutomaticResponse(ctx, conn, resourceArn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		curStatus := aws.StringValue(out.Protection.ApplicationLayerAutomaticResponseConfiguration.Status)

		if curStatus == "ENABLED" {
			return out, statusNormal, nil
		}
		return nil, statusChangePending, nil
	}
}

func statusApplicationLayerAutomaticResponseDeleted(ctx context.Context, conn *shield.Shield, resourceArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeApplicationLayerAutomaticResponse(ctx, conn, resourceArn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}
		curStatus := aws.StringValue(out.Protection.ApplicationLayerAutomaticResponseConfiguration.Status)

		if curStatus == "DISABLED" {
			return nil, "", nil
		}
		return out, statusDeleting, nil
	}
}

func describeApplicationLayerAutomaticResponse(ctx context.Context, conn *shield.Shield, resourceArn string) (*shield.DescribeProtectionOutput, error) {
	in := &shield.DescribeProtectionInput{
		ResourceArn: aws.String(resourceArn),
	}
	out, err := conn.DescribeProtectionWithContext(ctx, in)
	if err != nil {
		var nfe *shield.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
	}

	if out == nil || out.Protection == nil || out.Protection.ApplicationLayerAutomaticResponseConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceApplicationLayerAutomaticResponseData struct {
	ID          types.String   `tfsdk:"id"`
	ResourceARN types.String   `tfsdk:"resource_arn"`
	Action      types.String   `tfsdk:"action"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}
