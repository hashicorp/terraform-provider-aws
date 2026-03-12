// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package savingsplans

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/savingsplans"
	awstypes "github.com/aws/aws-sdk-go-v2/service/savingsplans/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_savingsplans_savings_plan", name="Savings Plan")
// @Tags(identifierAttribute="savings_plan_arn")
func newSavingsPlanResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &savingsPlanResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type savingsPlanResource struct {
	framework.ResourceWithModel[savingsPlanResourceModel]
	framework.WithTimeouts
}

func (r *savingsPlanResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"commitment": schema.StringAttribute{
				Required:    true,
				Description: "The hourly commitment, in USD.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"currency": schema.StringAttribute{
				Computed:    true,
				Description: "The currency of the Savings Plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Computed:    true,
				Description: "The description.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ec2_instance_family": schema.StringAttribute{
				Computed:    true,
				Description: "The EC2 instance family for the Savings Plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"end": schema.StringAttribute{
				Computed:    true,
				Description: "The end time of the Savings Plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"offering_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the offering.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"payment_option": schema.StringAttribute{
				Computed:    true,
				Description: "The payment option for the Savings Plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"product_types": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
				Description: "The product types.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"purchase_time": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Optional:    true,
				Description: "The time at which to purchase the Savings Plan, in UTC format (YYYY-MM-DDTHH:MM:SSZ).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"recurring_payment_amount": schema.StringAttribute{
				Computed:    true,
				Description: "The recurring payment amount.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrRegion: schema.StringAttribute{
				Computed:    true,
				Description: "The AWS Region.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"returnable_until": schema.StringAttribute{
				Computed:    true,
				Description: "The recurring payment amount.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"savings_plan_arn": framework.ARNAttributeComputedOnly(),
			"savings_plan_id":  framework.IDAttribute(),
			"savings_plan_offering_id": schema.StringAttribute{
				Required:    true,
				Description: "The unique ID of a Savings Plan offering.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"savings_plan_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of Savings Plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"start": schema.StringAttribute{
				Computed:    true,
				Description: "The start time of the Savings Plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrState: schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.SavingsPlanState](),
				Computed:    true,
				Description: "The current state of the Savings Plan.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"term_duration_in_seconds": schema.Int64Attribute{
				Computed:    true,
				Description: "The duration of the term, in seconds.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"upfront_payment_amount": schema.StringAttribute{
				Optional:    true,
				Description: "The up-front payment amount.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

func (r *savingsPlanResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan savingsPlanResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SavingsPlansClient(ctx)

	savingsPlanOfferingID := fwflex.StringValueFromFramework(ctx, plan.SavingsPlanOfferingID)
	var input savingsplans.CreateSavingsPlanInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateSavingsPlan(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, savingsPlanOfferingID)
		return
	}

	id := aws.ToString(out.SavingsPlanId)
	savingsPlan, err := waitSavingsPlanCreated(ctx, conn, id, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	// Set values for unknowns.
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, savingsPlan, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *savingsPlanResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state savingsPlanResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SavingsPlansClient(ctx)

	id := fwflex.StringValueFromFramework(ctx, state.SavingsPlanID)
	out, err := findSavingsPlanByID(ctx, conn, id)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *savingsPlanResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state savingsPlanResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SavingsPlansClient(ctx)

	// Only queued Savings Plans can be deleted
	if state := state.State.ValueEnum(); state != awstypes.SavingsPlanStateQueued && state != awstypes.SavingsPlanStateQueuedDeleted {
		return
	}

	id := fwflex.StringValueFromFramework(ctx, state.SavingsPlanID)
	input := savingsplans.DeleteQueuedSavingsPlanInput{
		SavingsPlanId: aws.String(id),
	}
	_, err := conn.DeleteQueuedSavingsPlan(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}

	if _, err := waitSavingsPlanDeleted(ctx, conn, id, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, id)
		return
	}
}

func (r *savingsPlanResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("savings_plan_id"), req, resp)
}

func waitSavingsPlanCreated(ctx context.Context, conn *savingsplans.Client, id string, timeout time.Duration) (*awstypes.SavingsPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SavingsPlanStatePaymentPending, awstypes.SavingsPlanStateQueued),
		Target:  enum.Slice(awstypes.SavingsPlanStateActive),
		Refresh: statusSavingsPlan(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SavingsPlan); ok {
		return out, err
	}

	return nil, err
}

func waitSavingsPlanDeleted(ctx context.Context, conn *savingsplans.Client, id string, timeout time.Duration) (*awstypes.SavingsPlan, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SavingsPlanStateQueued, awstypes.SavingsPlanStateQueuedDeleted),
		Target:  []string{},
		Refresh: statusSavingsPlan(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*awstypes.SavingsPlan); ok {
		return out, err
	}

	return nil, err
}

func statusSavingsPlan(conn *savingsplans.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
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

	return findSavingsPlan(ctx, conn, &input)
}

func findSavingsPlan(ctx context.Context, conn *savingsplans.Client, input *savingsplans.DescribeSavingsPlansInput) (*awstypes.SavingsPlan, error) {
	output, err := findSavingsPlans(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findSavingsPlans(ctx context.Context, conn *savingsplans.Client, input *savingsplans.DescribeSavingsPlansInput) ([]awstypes.SavingsPlan, error) { // nosemgrep:ci.savingsplans-in-func-name
	var output []awstypes.SavingsPlan

	err := describeSavingsPlansPages(ctx, conn, input, func(page *savingsplans.DescribeSavingsPlansOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.SavingsPlans...)

		return !lastPage
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

type savingsPlanResourceModel struct {
	Commitment             types.String                                  `tfsdk:"commitment"`
	Currency               types.String                                  `tfsdk:"currency"`
	Description            types.String                                  `tfsdk:"description"`
	EC2InstanceFamily      types.String                                  `tfsdk:"ec2_instance_family"`
	End                    types.String                                  `tfsdk:"end"`
	OfferingID             types.String                                  `tfsdk:"offering_id"`
	PaymentOption          types.String                                  `tfsdk:"payment_option"`
	ProductTypes           fwtypes.ListOfString                          `tfsdk:"product_types"`
	PurchaseTime           timetypes.RFC3339                             `tfsdk:"purchase_time"`
	RecurringPaymentAmount types.String                                  `tfsdk:"recurring_payment_amount"`
	Region                 types.String                                  `tfsdk:"region"`
	ReturnableUntil        types.String                                  `tfsdk:"returnable_until"`
	SavingsPlanARN         types.String                                  `tfsdk:"savings_plan_arn"`
	SavingsPlanID          types.String                                  `tfsdk:"savings_plan_id"`
	SavingsPlanOfferingID  types.String                                  `tfsdk:"savings_plan_offering_id"`
	SavingsPlanType        types.String                                  `tfsdk:"savings_plan_type"`
	Start                  types.String                                  `tfsdk:"start"`
	State                  fwtypes.StringEnum[awstypes.SavingsPlanState] `tfsdk:"state"`
	Tags                   tftags.Map                                    `tfsdk:"tags"`
	TagsAll                tftags.Map                                    `tfsdk:"tags_all"`
	TermDurationInSeconds  types.Int64                                   `tfsdk:"term_duration_in_seconds"`
	Timeouts               timeouts.Value                                `tfsdk:"timeouts"`
	UpfrontPaymentAmount   types.String                                  `tfsdk:"upfront_payment_amount"`
}
