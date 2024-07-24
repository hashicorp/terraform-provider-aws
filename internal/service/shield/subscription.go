// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
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

// @FrameworkResource("aws_shield_subscription", name="Subscription")
func newResourceSubscription(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceSubscription{}, nil
}

const (
	ResNameSubscription = "Subscription"
)

type resourceSubscription struct {
	framework.ResourceWithConfigure
}

func (r *resourceSubscription) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_shield_subscription"
}

func (r *resourceSubscription) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"auto_renew": schema.StringAttribute{
				Description: "Whether to automatically renew the subscription when it expires.",
				Optional:    true,
				Computed:    true,
				CustomType:  fwtypes.StringEnumType[awstypes.AutoRenew](),
				Default:     stringdefault.StaticString(string(awstypes.AutoRenewEnabled)),
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrSkipDestroy: schema.BoolAttribute{
				Optional: true,
			},
		},
	}
}

func (r *resourceSubscription) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().ShieldClient(ctx)

	var plan resourceSubscriptionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(r.Meta().AccountID)

	if plan.AutoRenew.Equal(types.StringValue(string(awstypes.AutoRenewDisabled))) {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameSubscription, plan.ID.String(), nil),
			errors.New("subscription auto_renew flag cannot be changed earlier than 30 days before subscription end and later than 1 day before subscription end").Error(),
		)
		return
	}

	in := &shield.CreateSubscriptionInput{}
	_, err := conn.CreateSubscription(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionCreating, ResNameSubscription, plan.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSubscription) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().ShieldClient(ctx)

	var state resourceSubscriptionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSubscriptionByID(ctx, conn)
	if errs.IsA[*retry.NotFoundError](err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionReading, ResNameSubscription, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSubscription) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().ShieldClient(ctx)

	var plan, state resourceSubscriptionData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AutoRenew.Equal(state.AutoRenew) {
		in := &shield.UpdateSubscriptionInput{
			AutoRenew: awstypes.AutoRenew(plan.AutoRenew.ValueString()),
		}

		_, err := conn.UpdateSubscription(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Shield, create.ErrActionUpdating, ResNameSubscription, plan.ID.String(), err),
				err.Error(),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceSubscription) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().ShieldClient(ctx)

	var state resourceSubscriptionData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.SkipDestroy.ValueBool() {
		return
	}

	in := &shield.UpdateSubscriptionInput{
		AutoRenew: awstypes.AutoRenewDisabled,
	}

	_, err := conn.UpdateSubscription(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionDeleting, ResNameSubscription, state.ID.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceSubscription) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

func findSubscriptionByID(ctx context.Context, conn *shield.Client) (*awstypes.Subscription, error) {
	in := &shield.DescribeSubscriptionInput{}

	out, err := conn.DescribeSubscription(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}
		return nil, err
	}

	if out == nil || out.Subscription == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Subscription, nil
}

type resourceSubscriptionData struct {
	AutoRenew   fwtypes.StringEnum[awstypes.AutoRenew] `tfsdk:"auto_renew"`
	ID          types.String                           `tfsdk:"id"`
	SkipDestroy types.Bool                             `tfsdk:"skip_destroy"`
}
