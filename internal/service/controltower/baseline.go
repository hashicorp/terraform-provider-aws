// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package controltower

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/controltower"
	"github.com/aws/aws-sdk-go-v2/service/controltower/document"
	awstypes "github.com/aws/aws-sdk-go-v2/service/controltower/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_controltower_baseline", name="Baseline")
// @Tags(identifierAttribute="arn")
func newResourceBaseline(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceBaseline{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameBaseline = "Baseline"
)

type resourceBaseline struct {
	framework.ResourceWithModel[resourceBaselineData]
	framework.WithTimeouts
}

func (r *resourceBaseline) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"baseline_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"baseline_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operation_identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"target_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrParameters: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[parameter](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrKey: schema.StringAttribute{
							Required: true,
						},
						names.AttrValue: schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceBaseline) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().ControlTowerClient(ctx)

	var plan resourceBaselineData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := controltower.EnableBaselineInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &in)...)
	if response.Diagnostics.HasError() {
		return
	}

	in.Tags = getTagsIn(ctx)

	out, err := conn.EnableBaseline(ctx, &in)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ControlTower, create.ErrActionCreating, ResNameBaseline, plan.BaselineIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil || out.OperationIdentifier == nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ControlTower, create.ErrActionCreating, ResNameBaseline, plan.BaselineIdentifier.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ARN = fwflex.StringToFramework(ctx, out.Arn)
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitBaselineReady(ctx, conn, plan.ARN.ValueString(), createTimeout)
	if err != nil {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrARN), plan.ARN.ValueString())...)
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ControlTower, create.ErrActionWaitingForCreation, ResNameBaseline, plan.BaselineIdentifier.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceBaseline) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().ControlTowerClient(ctx)

	var state resourceBaselineData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findBaselineByID(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ControlTower, create.ErrActionSetting, ResNameBaseline, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, out, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourceBaseline) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().ControlTowerClient(ctx)

	var plan, state resourceBaselineData
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		in := controltower.UpdateEnabledBaselineInput{
			EnabledBaselineIdentifier: plan.ARN.ValueStringPointer(),
		}

		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &in)...)
		if response.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateEnabledBaseline(ctx, &in)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ControlTower, create.ErrActionUpdating, ResNameBaseline, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}

		if out == nil || out.OperationIdentifier == nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ControlTower, create.ErrActionUpdating, ResNameBaseline, plan.ARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		_, err = waitBaselineReady(ctx, conn, plan.ARN.ValueString(), updateTimeout)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ControlTower, create.ErrActionWaitingForUpdate, ResNameBaseline, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(fwflex.Flatten(ctx, out, &plan)...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceBaseline) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().ControlTowerClient(ctx)

	var state resourceBaselineData
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	in := &controltower.DisableBaselineInput{
		EnabledBaselineIdentifier: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DisableBaseline(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ControlTower, create.ErrActionDeleting, ResNameBaseline, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	deleteTimeout := r.UpdateTimeout(ctx, state.Timeouts)
	_, err = waitBaselineDeleted(ctx, conn, state.ARN.ValueString(), deleteTimeout)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ControlTower, create.ErrActionWaitingForDeletion, ResNameBaseline, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceBaseline) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), request, response)
}

func waitBaselineReady(ctx context.Context, conn *controltower.Client, id string, timeout time.Duration) (*awstypes.EnabledBaselineDetails, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EnablementStatusUnderChange),
		Target:                    enum.Slice(awstypes.EnablementStatusSucceeded),
		Refresh:                   statusBaseline(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.EnabledBaselineDetails); ok {
		return out, err
	}

	return nil, err
}

func waitBaselineDeleted(ctx context.Context, conn *controltower.Client, id string, timeout time.Duration) (*awstypes.EnabledBaselineDetails, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EnablementStatusUnderChange),
		Target:                    []string{},
		Refresh:                   statusBaseline(conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.EnabledBaselineDetails); ok {
		return out, err
	}

	return nil, err
}

func statusBaseline(conn *controltower.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findBaselineByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.StatusSummary.Status), nil
	}
}

func findBaselineByID(ctx context.Context, conn *controltower.Client, id string) (*awstypes.EnabledBaselineDetails, error) {
	in := &controltower.GetEnabledBaselineInput{
		EnabledBaselineIdentifier: aws.String(id),
	}

	out, err := conn.GetEnabledBaseline(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil || out.EnabledBaselineDetails == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.EnabledBaselineDetails, nil
}

type resourceBaselineData struct {
	framework.WithRegionModel
	ARN                 types.String                               `tfsdk:"arn"`
	BaselineIdentifier  types.String                               `tfsdk:"baseline_identifier"`
	BaselineVersion     types.String                               `tfsdk:"baseline_version"`
	OperationIdentifier types.String                               `tfsdk:"operation_identifier"`
	Parameters          fwtypes.ListNestedObjectValueOf[parameter] `tfsdk:"parameters"`
	Tags                tftags.Map                                 `tfsdk:"tags"`
	TagsAll             tftags.Map                                 `tfsdk:"tags_all"`
	TargetIdentifier    types.String                               `tfsdk:"target_identifier"`
	Timeouts            timeouts.Value                             `tfsdk:"timeouts"`
}

type parameter struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

func (p *parameter) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics

	switch param := v.(type) {
	case awstypes.EnabledBaselineParameterSummary:
		p.Key = fwflex.StringToFramework(ctx, param.Key)
		if param.Value != nil {
			var value string
			err := param.Value.UnmarshalSmithyDocument(&value)
			if err != nil {
				diags.AddError(
					"Error Reading Control Tower Baseline Parameter",
					"Could not read Control Tower Baseline Parameter: "+err.Error(),
				)
				return diags
			}

			p.Value = fwflex.StringValueToFramework(ctx, value)
		} else {
			p.Value = types.StringNull()
		}
	}
	return diags
}

func (p parameter) ExpandTo(ctx context.Context, targetType reflect.Type) (any, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch targetType {
	case reflect.TypeOf(awstypes.EnabledBaselineParameter{}):
		var r awstypes.EnabledBaselineParameter
		if !p.Key.IsNull() {
			r.Key = fwflex.StringFromFramework(ctx, p.Key)
		}

		if !p.Value.IsNull() {
			r.Value = document.NewLazyDocument(fwflex.StringValueFromFramework(ctx, p.Value))
		}

		return &r, diags
	}

	return nil, diags
}
