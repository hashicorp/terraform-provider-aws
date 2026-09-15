// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arczonalshift

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arczonalshift/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_arczonalshift_autoshift_observer_notification_status", name="Autoshift Observer Notification Status")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(serialize=true)
// @Testing(generator=false)
func newAutoshiftObserverNotificationStatusResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &autoshiftObserverNotificationStatusResource{}

	return r, nil
}

const (
	ResNameAutoshiftObserverNotificationStatus = "Autoshift Observer Notification Status"
)

type autoshiftObserverNotificationStatusResource struct {
	framework.ResourceWithModel[autoshiftObserverNotificationStatusResourceModel]
	framework.WithImportByIdentity
}

func (r *autoshiftObserverNotificationStatusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AutoshiftObserverNotificationStatus](),
				Required:   true,
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
		Status: plan.Status.ValueEnum(),
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

	plan.ID = types.StringValue(r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *autoshiftObserverNotificationStatusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	var state autoshiftObserverNotificationStatusResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findAutoshiftObserverNotificationStatus(ctx, conn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	state.Status = fwtypes.StringEnumValue(out.Status)

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

	if !plan.Status.Equal(state.Status) {
		input := arczonalshift.UpdateAutoshiftObserverNotificationStatusInput{
			Status: plan.Status.ValueEnum(),
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
	}

	plan.ID = types.StringValue(r.Meta().Region(ctx))

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *autoshiftObserverNotificationStatusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ARCZonalShiftClient(ctx)

	input := arczonalshift.UpdateAutoshiftObserverNotificationStatusInput{
		Status: awstypes.AutoshiftObserverNotificationStatusDisabled,
	}

	_, err := conn.UpdateAutoshiftObserverNotificationStatus(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}
}

func findAutoshiftObserverNotificationStatus(ctx context.Context, conn *arczonalshift.Client) (*arczonalshift.GetAutoshiftObserverNotificationStatusOutput, error) {
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
	ID     types.String                                                     `tfsdk:"id"`
	Status fwtypes.StringEnum[awstypes.AutoshiftObserverNotificationStatus] `tfsdk:"status"`
}
