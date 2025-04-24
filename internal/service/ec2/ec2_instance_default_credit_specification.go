// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_instance_default_credit_specification", name="Default Credit Specification")
func newResourceDefaultCreditSpecification(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDefaultCreditSpecification{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameDefaultCreditSpecification = "Default Credit Specification"
)

type resourceDefaultCreditSpecification struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDefaultCreditSpecification) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"cpu_credits": schema.StringAttribute{
				Required: true,
			},
			"instance_family": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UnlimitedSupportedInstanceFamily](),
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (r *resourceDefaultCreditSpecification) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceDefaultCreditSpecificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.ModifyDefaultCreditSpecificationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = plan.InstanceFamily.StringValue

	out, err := createDefaultCreditSpecification(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameDefaultCreditSpecification, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameDefaultCreditSpecification, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	err = waitDefaultCreditSpecificationCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameDefaultCreditSpecification, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDefaultCreditSpecification) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceDefaultCreditSpecificationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDefaultCreditSpecificationByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameDefaultCreditSpecification, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDefaultCreditSpecification) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resourceDefaultCreditSpecificationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input ec2.ModifyDefaultCreditSpecificationInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	out, err := createDefaultCreditSpecification(ctx, conn, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QBusiness, create.ErrActionUpdating, ResNameDefaultCreditSpecification, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	err = waitDefaultCreditSpecificationCreated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForUpdate, ResNameDefaultCreditSpecification, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// No-op Delete
func (r *resourceDefaultCreditSpecification) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *resourceDefaultCreditSpecification) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitDefaultCreditSpecificationCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusDefaultCreditSpecification(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if _, ok := outputRaw.(*awstypes.InstanceFamilyCreditSpecification); ok {
		return err
	}

	return err
}

func statusDefaultCreditSpecification(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findDefaultCreditSpecificationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func findDefaultCreditSpecificationByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.InstanceFamilyCreditSpecification, error) {
	in := ec2.GetDefaultCreditSpecificationInput{
		InstanceFamily: awstypes.UnlimitedSupportedInstanceFamily(id),
	}

	out, err := conn.GetDefaultCreditSpecification(ctx, &in)
	if err != nil {
		return nil, err
	}

	if out == nil || out.InstanceFamilyCreditSpecification == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.InstanceFamilyCreditSpecification, nil
}

func createDefaultCreditSpecification(ctx context.Context, conn *ec2.Client, in *ec2.ModifyDefaultCreditSpecificationInput) (*awstypes.InstanceFamilyCreditSpecification, error) {
	out, err := conn.ModifyDefaultCreditSpecification(ctx, in)
	if err != nil {
		return nil, err
	}

	if out == nil || out.InstanceFamilyCreditSpecification == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.InstanceFamilyCreditSpecification, nil
}

type resourceDefaultCreditSpecificationModel struct {
	ID             types.String                                                  `tfsdk:"id"`
	Timeouts       timeouts.Value                                                `tfsdk:"timeouts"`
	InstanceFamily fwtypes.StringEnum[awstypes.UnlimitedSupportedInstanceFamily] `tfsdk:"instance_family"`
	CpuCredits     types.String                                                  `tfsdk:"cpu_credits"`
}
