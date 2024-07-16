// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ebs_fast_snapshot_restore", name="EBS Fast Snapshot Restore")
func newEBSFastSnapshotRestoreResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &ebsFastSnapshotRestoreResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type ebsFastSnapshotRestoreResource struct {
	framework.ResourceWithConfigure
	framework.WithNoUpdate
	framework.WithImportByID
	framework.WithTimeouts
}

func (*ebsFastSnapshotRestoreResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_ebs_fast_snapshot_restore"
}

func (r *ebsFastSnapshotRestoreResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAvailabilityZone: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrSnapshotID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *ebsFastSnapshotRestoreResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ebsFastSnapshotRestoreResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	availabilityZone := data.AvailabilityZone.ValueString()
	snapshotID := data.SnapshotID.ValueString()
	input := &ec2.EnableFastSnapshotRestoresInput{
		AvailabilityZones: []string{availabilityZone},
		SourceSnapshotIds: []string{snapshotID},
	}

	output, err := conn.EnableFastSnapshotRestores(ctx, input)

	if err == nil && output != nil {
		err = enableFastSnapshotRestoreItemsError(output.Unsuccessful)
	}

	if err != nil {
		response.Diagnostics.AddError("creating EC2 EBS Fast Snapshot Restore", err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	v, err := waitFastSnapshotRestoreCreated(ctx, conn, availabilityZone, snapshotID, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 EBS Fast Snapshot Restore (%s) create", data.ID.ValueString()), err.Error())

		return
	}

	data.State = fwflex.StringValueToFramework(ctx, v.State)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *ebsFastSnapshotRestoreResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ebsFastSnapshotRestoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().EC2Client(ctx)

	v, err := findFastSnapshotRestoreByTwoPartKey(ctx, conn, data.AvailabilityZone.ValueString(), data.SnapshotID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading EC2 EBS Fast Snapshot Restore (%s)", data.ID.ValueString()), err.Error())

		return
	}

	data.State = fwflex.StringValueToFramework(ctx, v.State)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *ebsFastSnapshotRestoreResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ebsFastSnapshotRestoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	availabilityZone := data.AvailabilityZone.ValueString()
	snapshotID := data.SnapshotID.ValueString()
	_, err := conn.DisableFastSnapshotRestores(ctx, &ec2.DisableFastSnapshotRestoresInput{
		AvailabilityZones: []string{availabilityZone},
		SourceSnapshotIds: []string{snapshotID},
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting EC2 EBS Fast Snapshot Restore (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if _, err := waitFastSnapshotRestoreDeleted(ctx, conn, availabilityZone, snapshotID, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for EC2 EBS Fast Snapshot Restore (%s) delete", data.ID.ValueString()), err.Error())

		return
	}
}

type ebsFastSnapshotRestoreResourceModel struct {
	AvailabilityZone types.String   `tfsdk:"availability_zone"`
	ID               types.String   `tfsdk:"id"`
	SnapshotID       types.String   `tfsdk:"snapshot_id"`
	State            types.String   `tfsdk:"state"`
	Timeouts         timeouts.Value `tfsdk:"timeouts"`
}

const (
	ebsFastSnapshotRestoreIDPartCount = 2
)

func (data *ebsFastSnapshotRestoreResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, ebsFastSnapshotRestoreIDPartCount, false)

	if err != nil {
		return err
	}

	data.AvailabilityZone = types.StringValue(parts[0])
	data.SnapshotID = types.StringValue(parts[1])

	return nil
}

func (data *ebsFastSnapshotRestoreResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.AvailabilityZone.ValueString(), data.SnapshotID.ValueString()}, ebsFastSnapshotRestoreIDPartCount, false)))
}
