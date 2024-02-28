// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource(name="DRT Access Role ARN Association")
func newResourceDRTAccessRoleARNAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDRTAccessRoleARNAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDRTAccessRoleARNAssociation = "DRT Access Role ARN Association"
)

const (
	propagationTimeout = 2 * time.Minute
)

type resourceDRTAccessRoleARNAssociation struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDRTAccessRoleARNAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_shield_drt_access_role_arn_association"
}

func (r *resourceDRTAccessRoleARNAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"role_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
				Read:   true,
			}),
		},
	}
}

func (r *resourceDRTAccessRoleARNAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ShieldClient(ctx)

	var plan resourceDRTAccessRoleARNAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.AssociateDRTRoleInput{
		RoleArn: flex.StringFromFramework(ctx, plan.RoleARN),
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AssociateDRTRole(ctx, in)
	}, "InvalidParameterException", "role does not have a valid DRT managed policy")

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameDRTAccessRoleARNAssociation, plan.RoleARN.String(), err),
			err.Error(),
		)
		return
	}

	out := outputRaw.(*shield.AssociateDRTRoleOutput)

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameDRTAccessRoleARNAssociation, plan.RoleARN.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDRTAccessRoleARNAssociationCreated(ctx, conn, plan.RoleARN.ValueString(), createTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForCreation, ResNameDRTAccessRoleARNAssociation, plan.RoleARN.String(), err),
			err.Error(),
		)
		return
	}

	plan.ID = flex.StringValueToFramework(ctx, plan.RoleARN.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDRTAccessRoleARNAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ShieldClient(ctx)

	var state resourceDRTAccessRoleARNAssociationData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &shield.DescribeDRTAccessInput{}
	out, err := conn.DescribeDRTAccess(ctx, in)

	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionSetting, ResNameDRTAccessRoleARNAssociation, state.RoleARN.String(), err),
			err.Error(),
		)
		return
	}
	if state.ID.IsNull() || state.ID.IsUnknown() {
		// Setting ID of state - required by hashicorps terraform plugin testing framework for Import. See issue https://github.com/hashicorp/terraform-plugin-testing/issues/84
		state.ID = flex.StringValueToFramework(ctx, state.RoleARN.ValueString())
	}

	state.RoleARN = flex.StringToFrameworkARN(ctx, out.RoleArn)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDRTAccessRoleARNAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ShieldClient(ctx)

	var plan, state resourceDRTAccessRoleARNAssociationData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.RoleARN.Equal(state.RoleARN) {
		in := &shield.AssociateDRTRoleInput{
			RoleArn: flex.StringFromFramework(ctx, plan.RoleARN),
		}

		out, err := conn.AssociateDRTRole(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameDRTAccessRoleARNAssociation, plan.RoleARN.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameDRTAccessRoleARNAssociation, plan.RoleARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitDRTAccessRoleARNAssociationUpdated(ctx, conn, plan.RoleARN.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForUpdate, ResNameDRTAccessRoleARNAssociation, plan.RoleARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceDRTAccessRoleARNAssociation) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ShieldClient(ctx)

	var state resourceDRTAccessRoleARNAssociationData

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	in := &shield.DisassociateDRTRoleInput{}

	_, err := conn.DisassociateDRTRole(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameDRTAccessRoleARNAssociation, state.RoleARN.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitDRTAccessRoleARNAssociationDeleted(ctx, conn, state.RoleARN.ValueString(), deleteTimeout)

	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionWaitingForDeletion, ResNameDRTAccessRoleARNAssociation, state.RoleARN.String(), err),
			err.Error(),
		)
		return
	}
}

func waitDRTAccessRoleARNAssociationCreated(ctx context.Context, conn *shield.Client, roleARN string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusDRTAccessRoleARNAssociation(ctx, conn, roleARN),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDRTAccessRoleARNAssociationUpdated(ctx context.Context, conn *shield.Client, roleARN string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusDRTAccessRoleARNAssociation(ctx, conn, roleARN),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func waitDRTAccessRoleARNAssociationDeleted(ctx context.Context, conn *shield.Client, roleARN string, timeout time.Duration) (*shield.DescribeDRTAccessOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusDRTAccessRoleARNAssociationDeleted(ctx, conn, roleARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*shield.DescribeDRTAccessOutput); ok {
		return out, err
	}

	return nil, err
}

func statusDRTAccessRoleARNAssociation(ctx context.Context, conn *shield.Client, roleARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeDRTAccessRoleARNAssociation(ctx, conn, roleARN)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func statusDRTAccessRoleARNAssociationDeleted(ctx context.Context, conn *shield.Client, roleARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := describeDRTAccessRoleARNAssociation(ctx, conn, roleARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if out.RoleArn != nil && aws.ToString(out.RoleArn) == roleARN {
			return out, statusDeleting, nil
		}

		return out, statusDeleting, nil
	}
}

func describeDRTAccessRoleARNAssociation(ctx context.Context, conn *shield.Client, roleARN string) (*shield.DescribeDRTAccessOutput, error) {
	in := &shield.DescribeDRTAccessInput{}

	out, err := conn.DescribeDRTAccess(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.RoleArn == nil || aws.ToString(out.RoleArn) != roleARN {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceDRTAccessRoleARNAssociationData struct {
	ID       types.String   `tfsdk:"id"`
	RoleARN  fwtypes.ARN    `tfsdk:"role_arn"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
