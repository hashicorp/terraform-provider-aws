// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package pricingplanmanager

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pricingplanmanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/pricingplanmanager/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_pricingplanmanager_subscription", name="Subscription")
// @ArnIdentity
// @Testing(hasNoPreExistingResource=true)
// @Testing(preCheck="testAccPreCheck")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/pricingplanmanager;pricingplanmanager.GetSubscriptionOutput")
func newSubscriptionResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &subscriptionResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)

	return r, nil
}

type subscriptionResource struct {
	framework.ResourceWithModel[subscriptionResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *subscriptionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"approval_mode": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ApprovalMode](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"etag": schema.StringAttribute{
				Computed: true,
			},
			"plan_family": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan_tier": schema.StringAttribute{
				Required: true,
			},
			"resource_arns": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Required:    true,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 10),
				},
			},
			"scheduled_change": schema.ListAttribute{
				CustomType:  fwtypes.NewListNestedObjectTypeOf[scheduledChangeModel](ctx),
				Computed:    true,
				ElementType: fwtypes.NewObjectTypeOf[scheduledChangeModel](ctx),
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatusReason: schema.StringAttribute{
				Computed: true,
			},
			"usage_level": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *subscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().PricingPlanManagerClient(ctx)

	var plan subscriptionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input pricingplanmanager.CreateSubscriptionInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.CreateSubscription(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.PlanFamily.String())
		return
	}
	if out == nil || out.Subscription == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.PlanFamily.String())
		return
	}

	arn := aws.ToString(out.Subscription.Arn)
	output, err := waitSubscriptionSynced(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Subscription, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ETag = flex.StringToFramework(ctx, output.ETag)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *subscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().PricingPlanManagerClient(ctx)

	var state subscriptionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findSubscriptionByARN(ctx, conn, state.ARN.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ARN.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Subscription, &state))
	if resp.Diagnostics.HasError() {
		return
	}
	state.ETag = flex.StringToFramework(ctx, output.ETag)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *subscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().PricingPlanManagerClient(ctx)

	var plan, state subscriptionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := state.ARN.ValueString()
	updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
	// Every mutation invalidates the subscription's entity tag, so each
	// successive call must use the value returned by the previous one.
	etag := state.ETag.ValueStringPointer()

	if !plan.PlanTier.Equal(state.PlanTier) || !plan.UsageLevel.Equal(state.UsageLevel) {
		input := pricingplanmanager.UpdateSubscriptionInput{
			Arn:      aws.String(arn),
			IfMatch:  etag,
			PlanTier: plan.PlanTier.ValueStringPointer(),
			// Omitting usageLevel resets it to the plan tier's default.
			UsageLevel: plan.UsageLevel.ValueStringPointer(),
		}

		out, err := conn.UpdateSubscription(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		etag = out.ETag

		if _, err := waitSubscriptionSynced(ctx, conn, arn, updateTimeout); err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}
	}

	if !plan.ResourceARNs.Equal(state.ResourceARNs) {
		os := flex.ExpandFrameworkStringValueSet(ctx, state.ResourceARNs)
		ns := flex.ExpandFrameworkStringValueSet(ctx, plan.ResourceARNs)

		if add := ns.Difference(os); len(add) > 0 {
			input := pricingplanmanager.AssociateResourcesToSubscriptionInput{
				Arn:          aws.String(arn),
				IfMatch:      etag,
				ResourceArns: add,
			}

			out, err := conn.AssociateResourcesToSubscription(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
				return
			}

			etag = out.ETag

			if _, err := waitSubscriptionSynced(ctx, conn, arn, updateTimeout); err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
				return
			}
		}

		if del := os.Difference(ns); len(del) > 0 {
			input := pricingplanmanager.DisassociateResourcesFromSubscriptionInput{
				Arn:          aws.String(arn),
				IfMatch:      etag,
				ResourceArns: del,
			}

			_, err := conn.DisassociateResourcesFromSubscription(ctx, &input)
			if err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
				return
			}

			if _, err := waitSubscriptionSynced(ctx, conn, arn, updateTimeout); err != nil {
				smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
				return
			}
		}
	}

	output, err := findSubscriptionByARN(ctx, conn, arn)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, output.Subscription, &plan))
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ETag = flex.StringToFramework(ctx, output.ETag)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *subscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().PricingPlanManagerClient(ctx)

	var state subscriptionResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := state.ARN.ValueString()

	// Re-read to get the current entity tag; the one in state may be stale.
	output, err := findSubscriptionByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	// Cancellation of an active subscription is scheduled by the service to
	// take effect at the end of the current billing period; until then the
	// subscription remains visible with a scheduled CANCELLATION change.
	// Subscriptions pending approval are deleted immediately.
	input := pricingplanmanager.CancelSubscriptionInput{
		Arn:     aws.String(arn),
		IfMatch: output.ETag,
	}

	if _, err := conn.CancelSubscription(ctx, &input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func findSubscriptionByARN(ctx context.Context, conn *pricingplanmanager.Client, arn string) (*pricingplanmanager.GetSubscriptionOutput, error) {
	input := pricingplanmanager.GetSubscriptionInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetSubscription(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if output == nil || output.Subscription == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func statusSubscription(conn *pricingplanmanager.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSubscriptionByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return output, string(output.Subscription.Status), nil
	}
}

// waitSubscriptionSynced waits for an in-flight change to settle. A paid
// subscription created with MANUAL approval mode parks in PENDING_APPROVAL
// until ApprovePaidSubscription is called, so that status is also terminal.
func waitSubscriptionSynced(ctx context.Context, conn *pricingplanmanager.Client, arn string, timeout time.Duration) (*pricingplanmanager.GetSubscriptionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusSyncInProgress),
		Target:  enum.Slice(awstypes.StatusActive, awstypes.StatusPendingApproval),
		Refresh: statusSubscription(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*pricingplanmanager.GetSubscriptionOutput); ok {
		if subscription := output.Subscription; subscription.Status == awstypes.StatusFailed {
			retry.SetLastError(err, errors.New(aws.ToString(subscription.StatusReason)))
		}

		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

type subscriptionResourceModel struct {
	ApprovalMode    fwtypes.StringEnum[awstypes.ApprovalMode]             `tfsdk:"approval_mode"`
	ARN             types.String                                          `tfsdk:"arn"`
	ETag            types.String                                          `tfsdk:"etag"`
	PlanFamily      types.String                                          `tfsdk:"plan_family"`
	PlanTier        types.String                                          `tfsdk:"plan_tier"`
	ResourceARNs    fwtypes.SetOfString                                   `tfsdk:"resource_arns"`
	ScheduledChange fwtypes.ListNestedObjectValueOf[scheduledChangeModel] `tfsdk:"scheduled_change"`
	Status          types.String                                          `tfsdk:"status"`
	StatusReason    types.String                                          `tfsdk:"status_reason"`
	Timeouts        timeouts.Value                                        `tfsdk:"timeouts"`
	UsageLevel      types.String                                          `tfsdk:"usage_level"`
}

type scheduledChangeModel struct {
	ChangeType    fwtypes.StringEnum[awstypes.ScheduledChangeType] `tfsdk:"change_type"`
	EffectiveDate timetypes.RFC3339                                `tfsdk:"effective_date"`
	PlanTier      types.String                                     `tfsdk:"plan_tier"`
	UsageLevel    types.String                                     `tfsdk:"usage_level"`
}
