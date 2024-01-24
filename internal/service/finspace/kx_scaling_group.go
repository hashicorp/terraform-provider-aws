// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package finspace

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/finspace"
	awstypes "github.com/aws/aws-sdk-go-v2/service/finspace/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Kx Scaling Group")
// @Tags(identifierAttribute="arn")
func newResourceKxScalingGroup(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceKxScalingGroup{}
	r.SetDefaultCreateTimeout(4 * time.Hour)
	r.SetDefaultReadTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(4 * time.Hour)
	r.SetMigratedFromPluginSDK(true)

	return r, nil
}

type resourceKxScalingGroup struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceKxScalingGroup) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_finspace_kx_scaling_group"
}

func (r *resourceKxScalingGroup) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourceKxScalingGroup) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id":  framework.IDAttribute(),
			"arn": framework.ARNAttributeComputedOnly(),
			"availability_zone_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 63),
				},
			},
			"host_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
				},
			},
			"created_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"last_modified_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"clusters": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"status_reason": schema.StringAttribute{
				Computed: true,
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

func (r *resourceKxScalingGroup) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

const (
	ResNameKxScalingGroup     = "Kx Scaling Group"
	kxScalingGroupIDPartCount = 2
)

func (r *resourceKxScalingGroup) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state resourceKxScalingGroupData
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	idParts := []string{
		state.EnvironmentId.ValueString(),
		state.ScalingGroupName.ValueString(),
	}
	rId, err := flex.FlattenResourceId(idParts, kxScalingGroupIDPartCount, false)
	state.ID = fwflex.StringValueToFramework(ctx, rId)

	createReq := &finspace.CreateKxScalingGroupInput{
		EnvironmentId:      aws.String(state.EnvironmentId.ValueString()),
		ScalingGroupName:   aws.String(state.ScalingGroupName.ValueString()),
		HostType:           aws.String(state.HostType.ValueString()),
		AvailabilityZoneId: aws.String(state.AvailabilityZoneId.ValueString()),
		Tags:               getTagsIn(ctx),
		ClientToken:        aws.String(id.UniqueId()),
	}

	scalingGroup, err := conn.CreateKxScalingGroup(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating Scaling Group", err.Error())
		return
	}
	if scalingGroup == nil || scalingGroup.ScalingGroupName == nil {
		resp.Diagnostics.AddError("Error creating Scaling Group", "empty output")
		return
	}

	createTimeout := r.CreateTimeout(ctx, state.Timeouts)
	if _, err := waitKxScalingGroupCreated(ctx, conn, state.ID.ValueString(), createTimeout); err != nil {
		resp.Diagnostics.AddError("Error waiting for ScalingGroup creation", err.Error())
		return
	}

	getScalingGroup, err := FindKxScalingGroupById(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, state.ID.ValueString(), err))
		return
	}
	state.refreshFromOutput(ctx, getScalingGroup)
	state.EnvironmentId = fwflex.StringToFramework(ctx, scalingGroup.EnvironmentId)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceKxScalingGroup) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceKxScalingGroupData
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scalingGroup, err := FindKxScalingGroupById(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, state.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, state.ID.ValueString(), err))
		return
	}

	state.refreshFromOutput(ctx, scalingGroup)

	parts, err := flex.ExpandResourceId(state.ID.ValueString(), kxScalingGroupIDPartCount, false)
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, state.ID.ValueString(), err))
		return
	}
	state.EnvironmentId = fwflex.StringToFramework(ctx, &parts[0])

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceKxScalingGroup) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Only for Tagging
	var old, new resourceKxScalingGroupData

	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &new)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().FinSpaceClient(ctx)

	if oldTagsAll, newTagsAll := old.TagsAll, new.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, new.ARN.ValueString(), oldTagsAll, newTagsAll); err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("updating tags for scaling group (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	// setting state to read
	scalingGroup, err := FindKxScalingGroupById(ctx, conn, new.ID.ValueString())
	if tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, new.ID.ValueString())
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, new.ID.ValueString(), err))
		return
	}

	new.refreshFromOutput(ctx, scalingGroup)

	parts, err := flex.ExpandResourceId(new.ID.ValueString(), kxScalingGroupIDPartCount, false)
	if err != nil {
		resp.Diagnostics.Append(create.DiagErrorFramework(names.FinSpace, create.ErrActionReading, ResNameKxScalingGroup, new.ID.ValueString(), err))
		return
	}
	new.EnvironmentId = fwflex.StringToFramework(ctx, &parts[0])

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &new)...)
}

func (r *resourceKxScalingGroup) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceKxScalingGroupData
	conn := r.Meta().FinSpaceClient(ctx)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteKxScalingGroup(ctx, &finspace.DeleteKxScalingGroupInput{
		EnvironmentId:    aws.String(state.EnvironmentId.ValueString()),
		ScalingGroupName: aws.String(state.ScalingGroupName.ValueString()),
		ClientToken:      aws.String(id.UniqueId()),
	})

	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		resp.Diagnostics.AddError("Error deleting scaling group", err.Error())
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	if _, err := waitKxScalingGroupDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout); err != nil && !tfresource.NotFound(err) {
		resp.Diagnostics.AddError("Error waiting for scaling group deletion", err.Error())
		return
	}
}

func FindKxScalingGroupById(ctx context.Context, conn *finspace.Client, id string) (*finspace.GetKxScalingGroupOutput, error) {
	idParts, err := flex.ExpandResourceId(id, kxScalingGroupIDPartCount, false)
	if err != nil {
		return nil, err
	}

	in := &finspace.GetKxScalingGroupInput{
		EnvironmentId:    aws.String(idParts[0]),
		ScalingGroupName: aws.String(idParts[1]),
	}

	out, err := conn.GetKxScalingGroup(ctx, in)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException

		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}

	if out == nil || out.ScalingGroupName == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}
	return out, nil
}

func waitKxScalingGroupCreated(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxScalingGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.KxScalingGroupStatusCreating),
		Target:                    enum.Slice(awstypes.KxScalingGroupStatusActive),
		Refresh:                   statusKxScalingGroup(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxScalingGroupOutput); ok {
		return out, err
	}
	return nil, err
}

func waitKxScalingGroupDeleted(ctx context.Context, conn *finspace.Client, id string, timeout time.Duration) (*finspace.GetKxScalingGroupOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.KxScalingGroupStatusDeleting),
		Target:  enum.Slice(awstypes.KxScalingGroupStatusDeleted),
		Refresh: statusKxScalingGroup(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*finspace.GetKxScalingGroupOutput); ok {
		return out, err
	}
	return nil, err
}

func statusKxScalingGroup(ctx context.Context, conn *finspace.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindKxScalingGroupById(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}
		return out, string(out.Status), nil
	}
}

type resourceKxScalingGroupData struct {
	ID                    types.String   `tfsdk:"id"`
	ARN                   types.String   `tfsdk:"arn"`
	AvailabilityZoneId    types.String   `tfsdk:"availability_zone_id"`
	EnvironmentId         types.String   `tfsdk:"environment_id"`
	ScalingGroupName      types.String   `tfsdk:"name"`
	HostType              types.String   `tfsdk:"host_type"`
	CreatedTimestamp      types.String   `tfsdk:"created_timestamp"`
	LastModifiedTimestamp types.String   `tfsdk:"last_modified_timestamp"`
	Status                types.String   `tfsdk:"status"`
	StatusReason          types.String   `tfsdk:"status_reason"`
	Clusters              types.List     `tfsdk:"clusters"`
	Timeouts              timeouts.Value `tfsdk:"timeouts"`
	Tags                  types.Map      `tfsdk:"tags"`
	TagsAll               types.Map      `tfsdk:"tags_all"`
}

func (rd *resourceKxScalingGroupData) refreshFromOutput(ctx context.Context, out *finspace.GetKxScalingGroupOutput) {
	if out == nil {
		return
	}
	rd.ScalingGroupName = fwflex.StringToFramework(ctx, out.ScalingGroupName)
	rd.AvailabilityZoneId = fwflex.StringToFramework(ctx, out.AvailabilityZoneId)
	rd.Status = fwflex.StringValueToFramework(ctx, out.Status)
	rd.CreatedTimestamp = fwflex.StringValueToFramework(ctx, out.CreatedTimestamp.String())
	rd.LastModifiedTimestamp = fwflex.StringValueToFramework(ctx, out.LastModifiedTimestamp.String())
	rd.HostType = fwflex.StringToFramework(ctx, out.HostType)
	rd.ARN = fwflex.StringToFramework(ctx, out.ScalingGroupArn)
	rd.Clusters = fwflex.FlattenFrameworkStringValueList(ctx, out.Clusters)
	rd.StatusReason = fwflex.StringToFramework(ctx, out.StatusReason)

}
