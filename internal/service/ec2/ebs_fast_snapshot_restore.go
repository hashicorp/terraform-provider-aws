// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="EBS Fast Snapshot Restore")
func newResourceEBSFastSnapshotRestore(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEBSFastSnapshotRestore{}
	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameEBSFastSnapshotRestore = "EBS Fast Snapshot Restore"

	ebsFastSnapshotRestoreIDPartCount = 2
)

type resourceEBSFastSnapshotRestore struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceEBSFastSnapshotRestore) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ebs_fast_snapshot_restore"
}

func (r *resourceEBSFastSnapshotRestore) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"availability_zone": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": framework.IDAttribute(),
			"snapshot_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceEBSFastSnapshotRestore) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourceEBSFastSnapshotRestoreData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	availabilityZone := plan.AvailabilityZone.ValueString()
	snapshotID := plan.SnapshotID.ValueString()

	idParts := []string{
		availabilityZone,
		snapshotID,
	}
	id, err := intflex.FlattenResourceId(idParts, ebsFastSnapshotRestoreIDPartCount, false)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameEBSFastSnapshotRestore, plan.SnapshotID.String(), err),
			err.Error(),
		)
		return
	}
	plan.ID = types.StringValue(id)

	in := &ec2.EnableFastSnapshotRestoresInput{
		AvailabilityZones: []string{availabilityZone},
		SourceSnapshotIds: []string{snapshotID},
	}

	out, err := conn.EnableFastSnapshotRestores(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameEBSFastSnapshotRestore, plan.SnapshotID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameEBSFastSnapshotRestore, plan.SnapshotID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}
	if len(out.Unsuccessful) > 0 || len(out.Successful) != 1 {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNameEBSFastSnapshotRestore, plan.SnapshotID.String(), nil),
			errors.New("enable fast snapshot restore was unsuccessful").Error(),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	waitOut, err := waitEBSFastSnapshotRestoreCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForCreation, ResNameEBSFastSnapshotRestore, plan.AvailabilityZone.String(), err),
			err.Error(),
		)
		return
	}

	plan.State = flex.StringValueToFramework(ctx, waitOut.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEBSFastSnapshotRestore) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceEBSFastSnapshotRestoreData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEBSFastSnapshotRestoreByID(ctx, conn, state.ID.ValueString())
	if errors.Is(err, tfresource.ErrEmptyResult) || out.State == awstypes.FastSnapshotRestoreStateCodeDisabled {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNameEBSFastSnapshotRestore, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.AvailabilityZone = flex.StringToFramework(ctx, out.AvailabilityZone)
	state.SnapshotID = flex.StringToFramework(ctx, out.SnapshotId)
	state.State = flex.StringValueToFramework(ctx, out.State)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceEBSFastSnapshotRestore) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is a no-op
}

func (r *resourceEBSFastSnapshotRestore) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourceEBSFastSnapshotRestoreData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ec2.DisableFastSnapshotRestoresInput{
		AvailabilityZones: []string{state.AvailabilityZone.ValueString()},
		SourceSnapshotIds: []string{state.SnapshotID.ValueString()},
	}

	_, err := conn.DisableFastSnapshotRestores(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNameEBSFastSnapshotRestore, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitEBSFastSnapshotRestoreDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionWaitingForDeletion, ResNameEBSFastSnapshotRestore, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEBSFastSnapshotRestore) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitEBSFastSnapshotRestoreCreated(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FastSnapshotRestoreStateCodeEnabling, awstypes.FastSnapshotRestoreStateCodeOptimizing),
		Target:  enum.Slice(awstypes.FastSnapshotRestoreStateCodeEnabled),
		Refresh: statusEBSFastSnapshotRestore(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DescribeFastSnapshotRestoreSuccessItem); ok {
		return out, err
	}

	return nil, err
}

func waitEBSFastSnapshotRestoreDeleted(ctx context.Context, conn *ec2.Client, id string, timeout time.Duration) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FastSnapshotRestoreStateCodeDisabling, awstypes.FastSnapshotRestoreStateCodeOptimizing, awstypes.FastSnapshotRestoreStateCodeEnabled),
		Target:  enum.Slice(awstypes.FastSnapshotRestoreStateCodeDisabled),
		Refresh: statusEBSFastSnapshotRestore(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.DescribeFastSnapshotRestoreSuccessItem); ok {
		return out, err
	}

	return nil, err
}

func statusEBSFastSnapshotRestore(ctx context.Context, conn *ec2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findEBSFastSnapshotRestoreByID(ctx, conn, id)
		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findEBSFastSnapshotRestoreByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.DescribeFastSnapshotRestoreSuccessItem, error) {
	parts, err := intflex.ExpandResourceId(id, ebsFastSnapshotRestoreIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &ec2.DescribeFastSnapshotRestoresInput{
		Filters: []awstypes.Filter{
			{
				Name:   aws.String("availability-zone"),
				Values: []string{parts[0]},
			},
			{
				Name:   aws.String("snapshot-id"),
				Values: []string{parts[1]},
			},
		},
	}

	out, err := conn.DescribeFastSnapshotRestores(ctx, in)
	if err != nil {
		return nil, err
	}

	if out == nil || len(out.FastSnapshotRestores) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}
	if len(out.FastSnapshotRestores) != 1 {
		return nil, tfresource.NewTooManyResultsError(len(out.FastSnapshotRestores), in)
	}

	return &out.FastSnapshotRestores[0], nil
}

type resourceEBSFastSnapshotRestoreData struct {
	AvailabilityZone types.String   `tfsdk:"availability_zone"`
	ID               types.String   `tfsdk:"id"`
	SnapshotID       types.String   `tfsdk:"snapshot_id"`
	State            types.String   `tfsdk:"state"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}
