// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/notifications"
	awstypes "github.com/aws/aws-sdk-go-v2/service/notifications/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/maps"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @FrameworkResource("aws_notifications_event_rule", name="Event Rule")
func newResourceEventRule(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceEventRule{}

	return r, nil
}

const (
	ResNameEventRule = "Event Rule"
)

type resourceEventRule struct {
	framework.ResourceWithConfigure
}

func (r *resourceEventRule) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			"event_pattern": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(0, 4096),
				},
			},
			"event_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(regexache.MustCompile(`([a-zA-Z0-9 \-\(\)])+`), ""),
				},
			},
			"notification_configuration_arn": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"regions": schema.SetAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(2, 25),
						stringvalidator.RegexMatches(regexache.MustCompile(`([a-z]{1,2})-([a-z]{1,15}-)+([0-9])`), ""),
					),
				},
			},
			names.AttrSource: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 36),
					stringvalidator.RegexMatches(regexache.MustCompile(`aws.([a-z0-9\-])+`), ""),
				},
			},
		},
	}
}

func (r *resourceEventRule) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var plan resourceEventRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input notifications.CreateEventRuleInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	out, err := conn.CreateEventRule(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionCreating, ResNameEventRule, plan.EventPattern.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil || out.Arn == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionCreating, ResNameEventRule, plan.EventPattern.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err = waitEventRuleCreated(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionWaitingForCreation, ResNameEventRule, plan.EventType.String(), err),
			err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceEventRule) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var state resourceEventRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEventRuleByARN(ctx, conn, state.ARN.ValueString())
	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionReading, ResNameEventRule, state.ARN.String(), err),
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

func (r *resourceEventRule) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var plan, state resourceEventRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		var input notifications.UpdateEventRuleInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		out, err := conn.UpdateEventRule(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Notifications, create.ErrActionUpdating, ResNameEventRule, plan.ARN.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil || out.Arn == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Notifications, create.ErrActionUpdating, ResNameEventRule, plan.ARN.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	_, err := waitEventRuleUpdated(ctx, conn, plan.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionWaitingForUpdate, ResNameEventRule, plan.ARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceEventRule) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NotificationsClient(ctx)

	var state resourceEventRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := notifications.DeleteEventRuleInput{
		Arn: state.ARN.ValueStringPointer(),
	}

	_, err := conn.DeleteEventRule(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}

		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionDeleting, ResNameEventRule, state.ARN.String(), err),
			err.Error(),
		)
		return
	}

	_, err = waitEventRuleDeleted(ctx, conn, state.ARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Notifications, create.ErrActionWaitingForDeletion, ResNameEventRule, state.ARN.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceEventRule) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrARN), req, resp)
}

func waitEventRuleCreated(ctx context.Context, conn *notifications.Client, id string) (*notifications.GetEventRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.EventRuleStatusCreating),
		Target:                    enum.Slice(awstypes.EventRuleStatusActive, awstypes.EventRuleStatusInactive),
		Refresh:                   statusEventRule(ctx, conn, id),
		Timeout:                   10 * time.Minute,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*notifications.GetEventRuleOutput); ok {
		return out, err
	}

	return nil, err
}

func waitEventRuleUpdated(ctx context.Context, conn *notifications.Client, id string) (*notifications.GetEventRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		// If regions were added/removed then rule status across regions can be a mix of "CREATING", "DELETING", "UPDATING"
		Pending:                   enum.Slice(awstypes.EventRuleStatusCreating, awstypes.EventRuleStatusUpdating, awstypes.EventRuleStatusDeleting),
		Target:                    enum.Slice(awstypes.EventRuleStatusActive, awstypes.EventRuleStatusInactive),
		Refresh:                   statusEventRule(ctx, conn, id),
		Timeout:                   10 * time.Minute,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*notifications.GetEventRuleOutput); ok {
		return out, err
	}

	return nil, err
}

func waitEventRuleDeleted(ctx context.Context, conn *notifications.Client, id string) (*notifications.GetEventRuleOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.EventRuleStatusDeleting),
		Target:  []string{},
		Refresh: statusEventRule(ctx, conn, id),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*notifications.GetEventRuleOutput); ok {
		return out, err
	}

	return nil, err
}

func statusEventRule(ctx context.Context, conn *notifications.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findEventRuleByARN(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		allActive := true
		allInactive := true

		for _, status := range maps.Values(out.StatusSummaryByRegion) {
			switch status.Status {
			// If regions were added/deleted then rule status across regions can be a mix of "CREATING", "DELETING", "UPDATING"
			// Does not matter which is returned as any of these is valid for waitEventRuleUpdated implementation
			case awstypes.EventRuleStatusCreating,
				awstypes.EventRuleStatusUpdating,
				awstypes.EventRuleStatusDeleting:
				return out, string(status.Status), nil
			case awstypes.EventRuleStatusActive:
				allInactive = false
			case awstypes.EventRuleStatusInactive:
				allActive = false
			}
		}

		if allActive {
			return out, string(awstypes.EventRuleStatusActive), nil
		}
		if allInactive {
			return out, string(awstypes.EventRuleStatusInactive), nil
		}

		return out, "", nil
	}
}

func findEventRuleByARN(ctx context.Context, conn *notifications.Client, arn string) (*notifications.GetEventRuleOutput, error) {
	input := notifications.GetEventRuleInput{
		Arn: aws.String(arn),
	}

	out, err := conn.GetEventRule(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: &input,
			}
		}

		return nil, err
	}

	if out == nil || out.Arn == nil {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return out, nil
}

type resourceEventRuleModel struct {
	ARN                          types.String `tfsdk:"arn"`
	EventPattern                 types.String `tfsdk:"event_pattern"`
	EventType                    types.String `tfsdk:"event_type"`
	NotificationConfigurationARN types.String `tfsdk:"notification_configuration_arn"`
	Regions                      types.Set    `tfsdk:"regions"`
	Source                       types.String `tfsdk:"source"`
}

func sweepEventRules(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := notifications.ListEventRulesInput{}
	conn := client.NotificationsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := notifications.NewListEventRulesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.EventRules {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceEventRule, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.Arn))),
			)
		}
	}

	return sweepResources, nil
}
