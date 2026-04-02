// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arczonalshift

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arczonalshift/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_arczonalshift_autoshift_observer_notification_status", name="Autoshift Observer Notification Status")
func newAutoshiftObserverNotificationStatusResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &autoshiftObserverNotificationStatusResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameAutoshiftObserverNotificationStatus = "Autoshift Observer Notification Status"
)

type autoshiftObserverNotificationStatusResource struct {
	framework.ResourceWithModel[autoshiftObserverNotificationStatusResourceModel]
	framework.WithTimeouts
}

func (r *autoshiftObserverNotificationStatusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"status": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("ENABLED", "DISABLED"),
				},
			},
		},
	}
}

func (r *autoshiftObserverNotificationStatusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var plan autoshiftObserverNotificationStatusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := arczonalshift.UpdateAutoshiftObserverNotificationStatusInput{
		Status: awstypes.AutoshiftObserverNotificationStatus(plan.Status.ValueString()),
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AutoshiftObserverNotificationStatus")))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.UpdateAutoshiftObserverNotificationStatus(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
	if out == nil || out.Status == "" {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"))
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// For this singleton resource, use a constant ID since there's only one per account
	plan.ID = types.StringValue("observer_notification_status")

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *autoshiftObserverNotificationStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var state autoshiftObserverNotificationStatusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAutoshiftObserverNotificationStatus(ctx, conn, "")
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// For this singleton resource, use a constant ID since there's only one per account
	state.ID = types.StringValue("observer_notification_status")

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *autoshiftObserverNotificationStatusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var plan, state autoshiftObserverNotificationStatusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		input := arczonalshift.UpdateAutoshiftObserverNotificationStatusInput{
			Status: awstypes.AutoshiftObserverNotificationStatus(plan.Status.ValueString()),
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input, flex.WithFieldNamePrefix("AutoshiftObserverNotificationStatus")))
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateAutoshiftObserverNotificationStatus(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err)
			return
		}
		if out == nil || out.Status == "" {
			smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"))
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// For this singleton resource, use a constant ID since there's only one per account
	plan.ID = types.StringValue("observer_notification_status")

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *autoshiftObserverNotificationStatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var state autoshiftObserverNotificationStatusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := arczonalshift.UpdateAutoshiftObserverNotificationStatusInput{
		Status: awstypes.AutoshiftObserverNotificationStatusDisabled,
	}

	_, err := conn.UpdateAutoshiftObserverNotificationStatus(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *autoshiftObserverNotificationStatusResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findAutoshiftObserverNotificationStatus(ctx context.Context, conn *arczonalshift.Client, id string) (*arczonalshift.GetAutoshiftObserverNotificationStatusOutput, error) {
	input := arczonalshift.GetAutoshiftObserverNotificationStatusInput{}

	out, err := conn.GetAutoshiftObserverNotificationStatus(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil || out.Status == "" {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type autoshiftObserverNotificationStatusResourceModel struct {
	framework.WithRegionModel
	ID     types.String `tfsdk:"id"`
	Status types.String `tfsdk:"status"`
}
