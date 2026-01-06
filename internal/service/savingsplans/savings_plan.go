// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package savingsplans

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/savingsplans"
	awstypes "github.com/aws/aws-sdk-go-v2/service/savingsplans/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_savingsplans_plan", name="Savings Plan")
// @Tags(identifierAttribute="arn")
func newResourceSavingsPlan(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceSavingsPlan{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

const (
	ResNameSavingsPlan = "Savings Plan"
)

type resourceSavingsPlan struct {
	framework.ResourceWithModel[resourceSavingsPlanModel]
	framework.WithTimeouts
}

func (r *resourceSavingsPlan) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			"savings_plan_offering_id": schema.StringAttribute{
				Required:    true,
				Description: "The unique ID of a Savings Plan offering.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"commitment": schema.StringAttribute{
				Required:    true,
				Description: "The hourly commitment, in USD.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"purchase_time": schema.StringAttribute{
				Optional:    true,
				Description: "The time at which to purchase the Savings Plan, in UTC format (YYYY-MM-DDTHH:MM:SSZ).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"client_token": schema.StringAttribute{
				Optional:    true,
				Description: "Unique, case-sensitive identifier to ensure the idempotency of the request.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrState: schema.StringAttribute{
				Computed:    true,
				Description: "The current state of the Savings Plan.",
			},
			"start": schema.StringAttribute{
				Computed:    true,
				Description: "The start time of the Savings Plan.",
			},
			"end": schema.StringAttribute{
				Computed:    true,
				Description: "The end time of the Savings Plan.",
			},
			"savings_plan_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of Savings Plan.",
			},
			"payment_option": schema.StringAttribute{
				Computed:    true,
				Description: "The payment option for the Savings Plan.",
			},
			"currency": schema.StringAttribute{
				Computed:    true,
				Description: "The currency of the Savings Plan.",
			},
			"upfront_payment_amount": schema.StringAttribute{
				Computed:    true,
				Description: "The up-front payment amount.",
			},
			"recurring_payment_amount": schema.StringAttribute{
				Computed:    true,
				Description: "The recurring payment amount.",
			},
			"term_duration_in_seconds": schema.Int64Attribute{
				Computed:    true,
				Description: "The duration of the term, in seconds.",
			},
			"ec2_instance_family": schema.StringAttribute{
				Computed:    true,
				Description: "The EC2 instance family for the Savings Plan.",
			},
			names.AttrRegion: schema.StringAttribute{
				Computed:    true,
				Description: "The AWS Region.",
			},
			"offering_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the offering.",
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceSavingsPlan) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().SavingsPlansClient(ctx)

	var plan resourceSavingsPlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	input := savingsplans.CreateSavingsPlanInput{
		SavingsPlanOfferingId: plan.SavingsPlanOfferingID.ValueStringPointer(),
		Commitment:            plan.Commitment.ValueStringPointer(),
		Tags:                  getTagsIn(ctx),
	}

	if !plan.PurchaseTime.IsNull() && !plan.PurchaseTime.IsUnknown() {
		purchaseTime, err := time.Parse(time.RFC3339, plan.PurchaseTime.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, "parse purchase_time")
			return
		}
		input.PurchaseTime = aws.Time(purchaseTime)
	}

	if !plan.ClientToken.IsNull() && !plan.ClientToken.IsUnknown() {
		input.ClientToken = plan.ClientToken.ValueStringPointer()
	}

	out, err := conn.CreateSavingsPlan(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.SavingsPlanOfferingID.String())
		return
	}
	if out == nil || out.SavingsPlanId == nil {
		smerr.AddError(ctx, &resp.Diagnostics, errors.New("empty output"), smerr.ID, plan.SavingsPlanOfferingID.String())
		return
	}

	plan.ID = types.StringPointerValue(out.SavingsPlanId)

	// Read the full details of the created savings plan
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	savingsPlan, err := waitSavingsPlanCreated(ctx, conn, plan.ID.ValueString(), createTimeout)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.ID.String())
		return
	}

	// Flatten the response into the plan
	// Flatten the response into the plan
	flattenSavingsPlan(ctx, savingsPlan, &plan)

	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
}

func (r *resourceSavingsPlan) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().SavingsPlansClient(ctx)

	var state resourceSavingsPlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findSavingsPlanByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	flattenSavingsPlan(ctx, out, &state)
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceSavingsPlan) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Savings Plans cannot be updated - all attributes require replacement
	// Tags are handled separately by the framework
	var plan resourceSavingsPlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	// The framework handles tags, so we only need to set the state
	diags := resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *resourceSavingsPlan) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().SavingsPlansClient(ctx)

	var state resourceSavingsPlanModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	// Only queued Savings Plans can be deleted
	// Check current state first
	sp, err := findSavingsPlanByID(ctx, conn, state.ID.ValueString())
	if retry.NotFound(err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
		return
	}

	// Only attempt to delete if in queued state
	if sp.State == awstypes.SavingsPlanStateQueued || sp.State == awstypes.SavingsPlanStateQueuedDeleted {
		input := savingsplans.DeleteQueuedSavingsPlanInput{
			SavingsPlanId: state.ID.ValueStringPointer(),
		}

		_, err = conn.DeleteQueuedSavingsPlan(ctx, &input)
		if err != nil {
			if errs.IsA[*awstypes.ResourceNotFoundException](err) {
				return
			}

			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}

		deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
		_, err = waitSavingsPlanDeleted(ctx, conn, state.ID.ValueString(), deleteTimeout)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.String())
			return
		}
	}
	// For active plans, we simply remove from state but cannot actually delete them
}

func (r *resourceSavingsPlan) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), req, resp)
}

// Waiters
func waitSavingsPlanCreated(ctx context.Context, conn *savingsplans.Client, id string, timeout time.Duration) (*awstypes.SavingsPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{string(awstypes.SavingsPlanStatePaymentPending), string(awstypes.SavingsPlanStateQueued)},
		Target:                    []string{string(awstypes.SavingsPlanStateActive), string(awstypes.SavingsPlanStateQueued)},
		Refresh:                   statusSavingsPlan(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SavingsPlan); ok {
		return out, err
	}

	return nil, err
}

func waitSavingsPlanDeleted(ctx context.Context, conn *savingsplans.Client, id string, timeout time.Duration) (*awstypes.SavingsPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{string(awstypes.SavingsPlanStateQueued), string(awstypes.SavingsPlanStateQueuedDeleted)},
		Target:  []string{},
		Refresh: statusSavingsPlan(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SavingsPlan); ok {
		return out, err
	}

	return nil, err
}

func statusSavingsPlan(_ context.Context, conn *savingsplans.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (interface{}, string, error) {
		out, err := findSavingsPlanByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func findSavingsPlanByID(ctx context.Context, conn *savingsplans.Client, id string) (*awstypes.SavingsPlan, error) {
	input := savingsplans.DescribeSavingsPlansInput{
		SavingsPlanIds: []string{id},
	}

	out, err := conn.DescribeSavingsPlans(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil || len(out.SavingsPlans) == 0 {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return &out.SavingsPlans[0], nil
}

func flattenSavingsPlan(ctx context.Context, sp *awstypes.SavingsPlan, model *resourceSavingsPlanModel) {
	model.ARN = flex.StringToFramework(ctx, sp.SavingsPlanArn)
	model.ID = flex.StringToFramework(ctx, sp.SavingsPlanId)
	model.State = types.StringValue(string(sp.State))
	model.SavingsPlanType = types.StringValue(string(sp.SavingsPlanType))
	model.PaymentOption = types.StringValue(string(sp.PaymentOption))
	model.Currency = types.StringValue(string(sp.Currency))
	model.UpfrontPaymentAmount = flex.StringToFramework(ctx, sp.UpfrontPaymentAmount)
	model.RecurringPaymentAmount = flex.StringToFramework(ctx, sp.RecurringPaymentAmount)
	model.TermDurationInSeconds = types.Int64Value(sp.TermDurationInSeconds)
	model.EC2InstanceFamily = flex.StringToFramework(ctx, sp.Ec2InstanceFamily)
	model.Region = flex.StringToFramework(ctx, sp.Region)
	model.OfferingID = flex.StringToFramework(ctx, sp.OfferingId)
	model.Commitment = flex.StringToFramework(ctx, sp.Commitment)

	if sp.Start != nil {
		model.Start = types.StringValue(*sp.Start)
	}
	if sp.End != nil {
		model.End = types.StringValue(*sp.End)
	}
}

type resourceSavingsPlanModel struct {
	ARN                    types.String   `tfsdk:"arn"`
	ClientToken            types.String   `tfsdk:"client_token"`
	Commitment             types.String   `tfsdk:"commitment"`
	Currency               types.String   `tfsdk:"currency"`
	EC2InstanceFamily      types.String   `tfsdk:"ec2_instance_family"`
	End                    types.String   `tfsdk:"end"`
	ID                     types.String   `tfsdk:"id"`
	OfferingID             types.String   `tfsdk:"offering_id"`
	PaymentOption          types.String   `tfsdk:"payment_option"`
	PurchaseTime           types.String   `tfsdk:"purchase_time"`
	RecurringPaymentAmount types.String   `tfsdk:"recurring_payment_amount"`
	Region                 types.String   `tfsdk:"region"`
	SavingsPlanOfferingID  types.String   `tfsdk:"savings_plan_offering_id"`
	SavingsPlanType        types.String   `tfsdk:"savings_plan_type"`
	Start                  types.String   `tfsdk:"start"`
	State                  types.String   `tfsdk:"state"`
	Tags                   tftags.Map     `tfsdk:"tags"`
	TagsAll                tftags.Map     `tfsdk:"tags_all"`
	TermDurationInSeconds  types.Int64    `tfsdk:"term_duration_in_seconds"`
	Timeouts               timeouts.Value `tfsdk:"timeouts"`
	UpfrontPaymentAmount   types.String   `tfsdk:"upfront_payment_amount"`
}
