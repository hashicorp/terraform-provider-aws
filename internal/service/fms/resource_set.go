// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fms/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Resource Set")
// @Tags(identifierAttribute="arn")
func newResourceResourceSet(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceResourceSet{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameResourceSet = "Resource Set"
)

type resourceResourceSet struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceResourceSet) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_fms_resource_set"
}

func (r *resourceResourceSet) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resourceSetLNB := schema.ListNestedBlock{
		CustomType: fwtypes.NewListNestedObjectTypeOf[resourceSetData](ctx),
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				names.AttrName: schema.StringAttribute{
					Required: true,
				},
				names.AttrDescription: schema.StringAttribute{
					Optional: true,
				},
				names.AttrID: framework.IDAttribute(),
				"last_update_time": schema.StringAttribute{
					Computed:   true,
					CustomType: timetypes.RFC3339Type{},
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.UseStateForUnknown(),
					},
				},
				"update_token": schema.StringAttribute{
					Optional: true,
					Computed: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"resource_set_status": schema.StringAttribute{
					Optional:   true,
					Computed:   true,
					CustomType: fwtypes.StringEnumType[awstypes.ResourceSetStatus](),
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
				},
				"resource_type_list": schema.ListAttribute{
					CustomType: fwtypes.ListOfStringType,
					Optional:   true,
				},
			},
		},
	}

	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:      framework.IDAttribute(),
			names.AttrARN:     framework.ARNAttributeComputedOnly(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"resource_set": resourceSetLNB,
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceResourceSet) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().FMSClient(ctx)

	var plan resourceResourceSetData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &fms.PutResourceSetInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in.TagList = getTagsIn(ctx)

	out, err := conn.PutResourceSet(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionCreating, ResNameResourceSet, plan.ID.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.ResourceSet == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionCreating, ResNameResourceSet, plan.ID.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = flex.StringToFramework(ctx, out.ResourceSetArn)
	plan.ID = flex.StringToFramework(ctx, out.ResourceSet.Id)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	output, err := waitResourceSetCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionWaitingForCreation, ResNameResourceSet, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceResourceSet) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().FMSClient(ctx)

	var state resourceResourceSetData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findResourceSetByID(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionSetting, ResNameResourceSet, state.ID.String(), err),
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

func (r *resourceResourceSet) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().FMSClient(ctx)

	var plan, state resourceResourceSetData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.ResourceSet.Equal(state.ResourceSet) {
		in := &fms.PutResourceSetInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

		out, err := conn.PutResourceSet(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.FMS, create.ErrActionUpdating, ResNameResourceSet, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.ResourceSet == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.FMS, create.ErrActionUpdating, ResNameResourceSet, plan.ID.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	_, err := waitResourceSetUpdated(ctx, conn, plan.ID.ValueString(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionWaitingForUpdate, ResNameResourceSet, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceResourceSet) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().FMSClient(ctx)

	var state resourceResourceSetData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &fms.DeleteResourceSetInput{
		Identifier: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeleteResourceSet(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionDeleting, ResNameResourceSet, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err = waitResourceSetDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.FMS, create.ErrActionWaitingForDeletion, ResNameResourceSet, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceResourceSet) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func (r *resourceResourceSet) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, req, resp)
}

func waitResourceSetCreated(ctx context.Context, conn *fms.Client, id string, timeout time.Duration) (*fms.GetResourceSetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    enum.Slice(awstypes.ResourceSetStatusActive),
		Refresh:                   statusResourceSet(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fms.GetResourceSetOutput); ok {
		return out, err
	}

	return nil, err
}

func waitResourceSetUpdated(ctx context.Context, conn *fms.Client, id string, timeout time.Duration) (*fms.GetResourceSetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ResourceSetStatusOutOfAdminScope),
		Target:                    enum.Slice(awstypes.ResourceSetStatusActive),
		Refresh:                   statusResourceSet(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fms.GetResourceSetOutput); ok {
		return out, err
	}

	return nil, err
}

func waitResourceSetDeleted(ctx context.Context, conn *fms.Client, id string, timeout time.Duration) (*fms.GetResourceSetOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ResourceSetStatusOutOfAdminScope),
		Target:  []string{},
		Refresh: statusResourceSet(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*fms.GetResourceSetOutput); ok {
		return out, err
	}

	return nil, err
}

func statusResourceSet(ctx context.Context, conn *fms.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findResourceSetByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.ToString((*string)(&out.ResourceSet.ResourceSetStatus)), nil
	}
}

func findResourceSetByID(ctx context.Context, conn *fms.Client, id string) (*fms.GetResourceSetOutput, error) {
	in := &fms.GetResourceSetInput{
		Identifier: aws.String(id),
	}

	out, err := conn.GetResourceSet(ctx, in)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.ResourceSet == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceResourceSetData struct {
	ARN         types.String                                     `tfsdk:"arn"`
	ID          types.String                                     `tfsdk:"id"`
	ResourceSet fwtypes.ListNestedObjectValueOf[resourceSetData] `tfsdk:"resource_set"`
	Tags        types.Map                                        `tfsdk:"tags"`
	TagsAll     types.Map                                        `tfsdk:"tags_all"`
	Timeouts    timeouts.Value                                   `tfsdk:"timeouts"`
}

type resourceSetData struct {
	Description       types.String                                   `tfsdk:"description"`
	ID                types.String                                   `tfsdk:"id"`
	LastUpdateTime    timetypes.RFC3339                              `tfsdk:"last_update_time"`
	Name              types.String                                   `tfsdk:"name"`
	ResourceSetStatus fwtypes.StringEnum[awstypes.ResourceSetStatus] `tfsdk:"resource_set_status"`
	ResourceTypeList  fwtypes.ListValueOf[types.String]              `tfsdk:"resource_type_list"`
	UpdateToken       types.String                                   `tfsdk:"update_token"`
}
